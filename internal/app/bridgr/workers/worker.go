package workers

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/assets"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Worker is the interface for how to talk to all instances of worker structs
type Worker interface {
	Setup() error
	Run() error
	Name() string
}

type workerCredentialReader struct{}

func loadTemplate(name string) (string, error) {
	f, err := assets.Templates.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func cleanContainer(cli client.ContainerAPIClient, name string) error {
	return cli.ContainerRemove(context.Background(), name, types.ContainerRemoveOptions{Force: true})
}

func pullImage(cli client.ImageAPIClient, image reference.Named) error {
	creds := dockerCredential{}
	getDockerAuth(image, &creds)
	output, err := cli.ImagePull(context.Background(), image.String(), types.ImagePullOptions{RegistryAuth: creds.String()})
	writer := ioutil.Discard
	if err != nil {
		return err
	}
	defer output.Close()
	if bridgr.Verbose {
		writer = os.Stderr
	}
	_, _ = io.Copy(writer, output) // must wait for output before returning
	return nil
}

func getDockerAuth(image reference.Named, rw CredentialReaderWriter) {
	imgDomain := "https://" + reference.Domain(image) // by putting scheme in front, it forces url.Parse to correctly identify the host portion
	url, _ := url.Parse(imgDomain)

	if creds, ok := rw.Read(url); ok {
		bridgr.Debugf("Docker: Found credentials for %s", url.Hostname())
		_ = rw.Write(creds)
	}
}

func runContainer(name string, containerConfig *container.Config, hostConfig *container.HostConfig, script string) error {
	ctx := context.Background()
	cli, _ := client.NewClientWithOpts(client.FromEnv)
	img, _ := reference.ParseNormalizedNamed(containerConfig.Image)
	// log.Printf("%+v", cli)
	_ = cleanContainer(cli, name)
	_ = pullImage(cli, img)

	resp, err := cli.ContainerCreate(ctx, containerConfig, hostConfig, nil, name)
	if err != nil {
		return err
	}

	hijack, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
	})
	if err != nil {
		return err
	}
	_, _ = io.Copy(hijack.Conn, bytes.NewBufferString(script))
	hijack.Conn.Close()

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return err
	}
	defer out.Close()
	_, _ = stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	return nil
}

func (w *workerCredentialReader) Read(url *url.URL) (Credential, bool) {
	basename := "BRIDGR_" + strings.ToUpper(strings.ReplaceAll(url.Hostname(), ".", "_"))
	bridgr.Debugf("Looking up credentials for: %s", basename)
	found := false
	userVal, ok := os.LookupEnv(basename + "_USER")
	found = found || ok
	passwdVal := ""
	if pw, ok := os.LookupEnv(basename + "_PASS"); ok {
		passwdVal = pw
		found = found || ok
	} else {
		token, tok := os.LookupEnv(basename + "_TOKEN")
		passwdVal = token
		found = found || tok
	}
	return Credential{
		userVal,
		passwdVal,
	}, found
}
