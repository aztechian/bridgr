package bridgr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

var baseImage = map[string]string{
	"yum":    "centos",
	"ruby":   "ruby",
	"python": "python",
}

// Batch is a struct that implements basic features common to all workers that are "batch" style.
// This means workers that create their repositories by setting up config files and using the execution of a docker container
// to download and/or create the repository content.
type batch struct {
	Mounts          []mount.Mount
	ContainerConfig *container.Config
}

func newBatch(image, pkgSource, repoSource, repoTarget string) batch {
	return batch{
		Mounts: []mount.Mount{
			{Type: mount.TypeBind, Source: pkgSource, Target: "/packages"}, // package mount
			{Type: mount.TypeBind, Source: repoSource, Target: repoTarget},  // mount for repository config
		},
		ContainerConfig: &container.Config{
			Image:        image,
			Cmd:          []string{"/bin/sh", "-"},
			Tty:          false,
			OpenStdin:    true,
			AttachStdout: true,
			AttachStderr: true,
			StdinOnce:    true,
		},
	}
}

func (b *batch) cleanContainer(cli client.ContainerAPIClient, name string) {
	if err := cli.ContainerRemove(context.Background(), name, types.ContainerRemoveOptions{Force: true}); err != nil {
		Printf("Error while cleaning batch container %s: %s", name, err)
	}
}

// RunContainer is a function available to BatchWorkers for running their "batch" operations
func (b *batch) runContainer(name, script string) error {
	ctx := context.Background()
	cli, _ := client.NewClientWithOpts(client.FromEnv)
	img, _ := reference.ParseNormalizedNamed(b.ContainerConfig.Image)
	name = fmt.Sprintf("%s_%d", name, os.Getpid()) // suffix the PID to the container name to not conflict with concurrent runs
	defer b.cleanContainer(cli, name)
	_ = PullImage(cli, img)

	resp, err := cli.ContainerCreate(ctx, b.ContainerConfig, &container.HostConfig{Mounts: b.Mounts}, nil, name)
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
