package workers

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/config"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/distribution/reference"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

var defaultImg, _ = reference.ParseNormalizedNamed("myimage:3.2.1")
var defaultDockerConf = config.Docker{
	Destination: "corp.repo",
	Items:       []reference.Named{defaultImg},
}

type stubCli struct {
	client.ImageAPIClient
	isError bool
	imgPush int
	imgTag  int
	imgSave int
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

func (c *stubCli) ImageSave(ctx context.Context, images []string) (io.ReadCloser, error) {
	c.imgSave = c.imgSave + 1
	if c.isError {
		return nil, errors.New("Error saving image")
	}
	return ioutil.NopCloser(strings.NewReader("OK")), nil
}

func TestDockerWriteRemote(t *testing.T) {
	tests := []struct {
		name    string
		cli     stubCli
		verbose bool
	}{
		{"succcess quiet", stubCli{}, false},
		{"success verbose", stubCli{}, true},
		{"error quiet", stubCli{isError: true}, false},
		{"error verbose", stubCli{isError: true}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bridgr.Verbose = test.verbose
			d := Docker{Cli: &test.cli}
			err := d.writeRemote("nothing", defaultImg)
			if err == nil && test.cli.isError {
				t.Error(err)
			}
			if test.cli.imgPush != 1 {
				t.Errorf("Expected ImagePush to be called once, got %d calls", test.cli.imgPush)
			}
		})
	}
}

func TestDockerTagForRemote(t *testing.T) {
	customImg, _ := reference.ParseNormalizedNamed("myproj/myimage:1.0")
	tests := []struct {
		name  string
		image reference.Named
		cli   stubCli
		want  string
	}{
		{"succcess", defaultImg, stubCli{}, "corp.repo/library/myimage:3.2.1"},
		{"full path", customImg, stubCli{}, "corp.repo/myproj/myimage:1.0"},
		{"error", defaultImg, stubCli{isError: true}, "corp.repo/library/myimage:3.2.1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := Docker{Cli: &test.cli, Config: &defaultDockerConf}
			tag := d.tagForRemote(test.image)
			if test.cli.imgTag != 1 {
				t.Errorf("Expected ImageTag to be called once, got %d calls", test.cli.imgTag)
			}
			if tag != test.want {
				t.Errorf("Expected %s from tagForRemote(), but got %s", test.want, tag)
			}
		})
	}
}

func TestDockerWriteLocal(t *testing.T) {
	tests := []struct {
		name   string
		cli    stubCli
		writer fakeWriteCloser
		error  bool
	}{
		{"success", stubCli{}, fakeWriteCloser{}, false},
		{"docker error", stubCli{isError: true}, fakeWriteCloser{}, true},
		// {"writer error", stubCli{}, fakeWriteCloser{isError: true}, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := Docker{Cli: &test.cli}
			err := d.writeLocal(&test.writer, defaultImg)
			if err == nil && test.error {
				t.Error("Expected an error condition from writeLocal() but got none")
			}
			if test.cli.imgSave != 1 {
				t.Errorf("Expected ImageSave to be called once, got %d calls", test.cli.imgSave)
			}
		})
	}
}
