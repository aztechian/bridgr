package workers_test

import (
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/distribution/reference"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

type stubCli struct {
	client.ImageAPIClient
	isError bool
	imgPull int
	imgPush int
	imgTag  int
}

func (c *stubCli) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	c.imgPull = c.imgPull + 1
	if c.isError {
		return nil, fmt.Errorf("Error pulling image")
	}
	return ioutil.NopCloser(strings.NewReader("OK")), nil
}

func (c *stubCli) ImagePush(ctx context.Context, ref string, options types.ImagePushOptions) (io.ReadCloser, error) {
	c.imgPush = c.imgPush + 1
	if c.isError {
		return nil, fmt.Errorf("Error pushing image")
	}
	return ioutil.NopCloser(strings.NewReader("OK")), nil
}

func (c *stubCli) ImageTag(ctx context.Context, image string, ref string) error {
	c.imgTag = c.imgTag + 1
	return nil
}

var defaultNamedImg, _ = reference.ParseNormalizedNamed("mytest:3.2")
var dockerConf = config.Docker{Items: []reference.Named{defaultNamedImg}}

func TestDockerRun(t *testing.T) {
	// TODO: unfortunately, os.Create() makes this test difficult. Refactor Run() to take an abstracted filesystem
	tests := []struct {
		name string
		cli  stubCli
	}{
		{"basic", stubCli{}},
		{"docker error", stubCli{isError: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf := config.Docker{Items: []reference.Named{defaultNamedImg}, Destination: "my.repo"}
			d := workers.Docker{Cli: &test.cli, Config: &conf}
			err := d.Run()
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestDockerSetup(t *testing.T) {
	tests := []struct {
		cli *stubCli
	}{
		{cli: &stubCli{imgPull: 0}},
		{cli: &stubCli{imgPull: 0, isError: true}},
	}

	for _, test := range tests {
		d := workers.Docker{Cli: test.cli, Config: &dockerConf}
		_ = d.Setup()
		if test.cli.imgPull != 1 {
			t.Errorf("Expected ImagePull() to have been called once, but it was called %d times", test.cli.imgPull)
		}
	}
}

func TestDockerName(t *testing.T) {
	d := workers.Docker{}
	want := "Docker"
	got := d.Name()
	if got != want {
		t.Errorf("Expected %s from Docker.Name(), but got %s", want, got)
	}
}
