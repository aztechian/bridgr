package workers

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/assets"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
	imgDomain := "https://" + reference.Domain(image) // by putting scheme in front, it forces url.Parse to correctly identify the host portion
	bridgr.Debugf("Got image domain of %s", imgDomain)
	url, err := url.Parse(imgDomain)
	bridgr.Debugf("Parsed URL: %s", url)
	encodedAuth := ""
	if err == nil {
		username, password := credentials(url)
		if username != "" && password != "" {
			imgAuth := types.AuthConfig{
				Username: username,
				Password: password,
			}
			bridgr.Debugf("Docker: Found credentials for %s", url.Hostname())
			jsonAuth, _ := json.Marshal(imgAuth)
			encodedAuth = base64.URLEncoding.EncodeToString(jsonAuth)
		}
	}
	output, err := cli.ImagePull(context.Background(), image.String(), types.ImagePullOptions{RegistryAuth: encodedAuth})
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

func credentials(url *url.URL) (string, string) {
	basename := "BRIDGR_" + strings.ToUpper(strings.ReplaceAll(url.Hostname(), ".", "_"))
	uservar := basename + "_USER"
	passwdvar := basename + "_PASS"
	bridgr.Debugf("Looking for env var: %s", uservar)
	if value, ok := os.LookupEnv(uservar); ok {
		return value, os.Getenv(passwdvar)
	}
	bridgr.Debugf("Env Var %s was not found :(", uservar)
	return "", ""
}

func credentialsConjoined(url *url.URL) string {
	u, p := credentials(url)
	return u + ":" + p
}

func credentialsBase64(url *url.URL) string {
	v := credentialsConjoined(url)
	return base64.StdEncoding.EncodeToString([]byte(v))
}
