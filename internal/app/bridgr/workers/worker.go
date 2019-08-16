package workers

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/assets"
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"

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

func pullImage(cli client.ImageAPIClient, image string) error {
	output, err := cli.ImagePull(context.Background(), image, types.ImagePullOptions{})
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
	// log.Printf("%+v", cli)
	_ = cleanContainer(cli, name)
	_ = pullImage(cli, "docker.io/"+containerConfig.Image)

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
