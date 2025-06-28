package bridgr

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	imagespecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
}

func (m *mockClient) ContainerAttach(ctx context.Context, container string, options container.AttachOptions) (types.HijackedResponse, error) {
	args := m.Called(ctx, container, options)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}
func (m *mockClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *networktypes.NetworkingConfig, platform *imagespecs.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}
func (m *mockClient) ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, container, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}
func (m *mockClient) ContainerStart(ctx context.Context, container string, options container.StartOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}
func (m *mockClient) ContainerRemove(ctx context.Context, container string, options container.RemoveOptions) error {
	args := m.Called(ctx, container, options)
	return args.Error(0)
}
func (m *mockClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

type mockConn struct {
	net.Conn
}

func (mc mockConn) Write(data []byte) (int, error) {
	io.Copy(ioutil.Discard, bytes.NewReader(data))
	return len(data), nil
}

func TestNewBatch(t *testing.T) {
	pkgMount := mount.Mount{
		Type:   "bind",
		Source: "michael",
		Target: "/packages",
	}
	repoMount := mount.Mount{
		Type:   "bind",
		Source: "george michael",
		Target: "bluth",
	}
	cfg := container.Config{
		Image:        "george",
		Cmd:          []string{"/bin/sh", "-"},
		Tty:          false,
		OpenStdin:    true,
		AttachStdout: true,
		AttachStderr: true,
		StdinOnce:    true,
	}

	expect := batch{
		Mounts:          []mount.Mount{pkgMount, repoMount},
		ContainerConfig: &cfg,
	}

	result := newBatch("george", "michael", "george michael", "bluth")

	if !cmp.Equal(expect, result, cmpopts.IgnoreFields(result, "Client")) {
		t.Error(cmp.Diff(expect, result))
	}
}

func TestCleanContainer(t *testing.T) {
	b := newBatch("ann", "gob", "G.O.B.'s wife", "marta")
	cli := mockClient{}
	cli.On("ContainerRemove", context.Background(), "gene", container.RemoveOptions{Force: true}).Return(nil).Return(errors.New("gene!!!"))
	b.Client = &cli
	b.cleanContainer("gene")
	b.cleanContainer("gene")
	cli.AssertNumberOfCalls(t, "ContainerRemove", 2)
}

func TestRunContainer(t *testing.T) {
	b := newBatch("rebel", "maeby", "spain", "maharis")
	namePid := fmt.Sprintf("rebel_%d", os.Getpid())
	connWriter, _ := net.Pipe()
	connWriter = mockConn{Conn: connWriter}
	cli := mockClient{}
	cli.On("ContainerCreate", context.Background(), b.ContainerConfig, mock.Anything, mock.Anything, mock.Anything, namePid).Return(container.CreateResponse{ID: "something"}, nil)
	cli.On("ImagePull", context.Background(), "docker.io/library/rebel", mock.Anything).Return(ioutil.NopCloser(bytes.NewReader([]byte("do not eat"))), nil)
	cli.On("ContainerRemove", context.Background(), namePid, container.RemoveOptions{Force: true}).Return(nil)
	cli.On("ContainerAttach", context.Background(), "something", mock.Anything).Return(types.HijackedResponse{Conn: connWriter, Reader: bufio.NewReader(strings.NewReader("annyong"))}, nil)
	cli.On("ContainerStart", context.Background(), "something", mock.Anything).Return(nil)
	cli.On("ContainerLogs", context.Background(), "something", mock.Anything).Return(ioutil.NopCloser(bytes.NewReader([]byte("do not eat"))), nil)
	b.Client = &cli
	err := b.runContainer("rebel", "fakeblock")
	if err != nil && !strings.Contains(err.Error(), "Unrecognized input header") { // we use StdCopy which is the duplexed output streaming format for docker. I'm not doing that.
		t.Error(err)
	}
}
