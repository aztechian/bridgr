package bridgr

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"

	"github.com/distribution/reference"
	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	imagespecs "github.com/opencontainers/image-spec/specs-go/v1"
	log "unknwon.dev/clog/v2"
)

var baseImage = map[string]string{
	"yum":    "centos",
	"ruby":   "ruby",
	"python": "python",
}

// DefaultContainerPlatform is the default platform used for container creation.
var DefaultContainerPlatform = &imagespecs.Platform{
	Architecture: runtime.GOARCH,
	OS:           "linux",
}

// Batch is a struct that implements basic features common to all workers that are "batch" style.
// This means workers that create their repositories by setting up config files and using the execution of a docker container
// to download and/or create the repository content.
type batch struct {
	Mounts          []mount.Mount
	ContainerConfig *containertypes.Config
	Client          containerImagerClient
}

type containerImagerClient interface {
	ImagePuller
	ContainerAttach(ctx context.Context, container string, options containertypes.AttachOptions) (types.HijackedResponse, error)
	ContainerCreate(ctx context.Context, config *containertypes.Config, hostConfig *containertypes.HostConfig, networkingConfig *networktypes.NetworkingConfig, platform *imagespecs.Platform, containerName string) (containertypes.CreateResponse, error)
	ContainerLogs(ctx context.Context, container string, options containertypes.LogsOptions) (io.ReadCloser, error)
	ContainerStart(ctx context.Context, container string, options containertypes.StartOptions) error
	ContainerRemove(ctx context.Context, container string, options containertypes.RemoveOptions) error
}

func newBatch(image, pkgSource, repoSource, repoTarget string) batch {
	client, _ := client.NewClientWithOpts(client.WithAPIVersionNegotiation(), client.FromEnv)
	return batch{
		Client: client,
		Mounts: []mount.Mount{
			{Type: mount.TypeBind, Source: pkgSource, Target: "/packages"}, // package mount
			{Type: mount.TypeBind, Source: repoSource, Target: repoTarget}, // mount for repository config
		},
		ContainerConfig: &containertypes.Config{
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

func (b *batch) cleanContainer(name string) {
	if err := b.Client.ContainerRemove(context.Background(), name, containertypes.RemoveOptions{Force: true}); err != nil {
		log.Warn("Error while cleaning batch container %s: %s", name, err)
	}
}

// RunContainer is a function available to BatchWorkers for running their "batch" operations
func (b *batch) runContainer(name, script string) error {
	ctx := context.Background()
	img, _ := reference.ParseNormalizedNamed(b.ContainerConfig.Image)
	name = fmt.Sprintf("%s_%d", name, os.Getpid()) // suffix the PID to the container name to not conflict with concurrent runs
	defer b.cleanContainer(name)
	_ = PullImage(b.Client, img)

	resp, err := b.Client.ContainerCreate(ctx, b.ContainerConfig, &containertypes.HostConfig{Mounts: b.Mounts}, nil, DefaultContainerPlatform, name)
	if err != nil {
		return err
	}

	hijack, err := b.Client.ContainerAttach(ctx, resp.ID, containertypes.AttachOptions{
		Stream: true,
		Stdin:  true,
	})
	if err != nil {
		return err
	}
	_, _ = io.Copy(hijack.Conn, bytes.NewBufferString(script))
	hijack.Conn.Close()

	if err := b.Client.ContainerStart(ctx, resp.ID, containertypes.StartOptions{}); err != nil {
		return err
	}

	out, err := b.Client.ContainerLogs(ctx, resp.ID, containertypes.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return err
	}
	defer out.Close()
	writer := bytes.Buffer{}
	// create an in-memory buffer for container output (both stdout and stderr)
	// use StdCopy to de-multiplex the stream from docker
	// send it to our logger
	_, err = stdcopy.StdCopy(&writer, &writer, out)
	scanner := bufio.NewScanner(&writer)
	for scanner.Scan() {
		log.Trace("%s", scanner.Text())
	}
	return err
}
