package workers

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/config"
	"context"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

// Docker is the worker struct for fetching Docker images
type Docker struct {
	Config *config.BridgrConf
	Cli    client.ImageAPIClient
}

// NewDocker creates a Docker worker from the configuration object
func NewDocker(conf *config.BridgrConf) Worker {
	_ = os.MkdirAll(conf.Docker.BaseDir(), os.ModePerm)
	cli, _ := client.NewClientWithOpts(client.FromEnv)
	return &Docker{Config: conf, Cli: cli}
}

// Name returns the string name of the Docker struct
func (d *Docker) Name() string {
	return "Docker"
}

// Run executes the Docker worker to fetch artifacts
func (d *Docker) Run() error {
	for _, img := range d.Config.Docker.Items {
		if d.Config.Docker.Destination != "" {
			dest := d.tagForRemote(img)
			err := d.writeRemote(dest, img)
			if err != nil {
				log.Println(err)
			}
		} else {
			re := regexp.MustCompile(`[:/]`)
			outFile := re.ReplaceAllString(reference.Path(img), "_") + ".tar"
			out, err := os.Create(path.Join(d.Config.Docker.BaseDir(), outFile))
			if err != nil {
				log.Println(err)
				continue
			}
			err = d.writeLocal(out, img)
			if err != nil {
				log.Println(err)
				os.Remove(out.Name())
				continue
			}
			bridgr.Debugf("saved Docker image %s to %s", img.String(), out.Name())
		}
	}
	return nil
}

// Setup gets the environment ready to run the Docker worker
func (d *Docker) Setup() error {
	for _, img := range d.Config.Docker.Items {
		bridgr.Debugf("pulling image %s", img.String())
		err := pullImage(d.Cli, img.String())
		if err != nil {
			bridgr.Printf("Error fetching Docker image `%s`: %s", img.String(), err)
		}
	}
	return nil
}

func (d *Docker) writeLocal(out io.WriteCloser, in reference.Named) error {
	ctx := context.Background()
	reader, err := d.Cli.ImageSave(ctx, []string{in.String()})
	if err != nil {
		return err
	}
	defer reader.Close()
	defer out.Close()
	_, err = io.Copy(out, reader)
	if err != nil {
		return err
	}
	return nil
}

func (d *Docker) writeRemote(remote string, in reference.Named) error {
	writer := ioutil.Discard
	if bridgr.Verbose {
		writer = os.Stderr
	}
	output, err := d.Cli.ImagePush(context.Background(), remote, types.ImagePushOptions{})
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(writer, output)
	return err
}

func (d *Docker) tagForRemote(local reference.Named) string {
	destReg := d.Config.Docker.Destination
	remoteTag := strings.Replace(local.String(), reference.Domain(local), destReg, -1)

	_ = d.Cli.ImageTag(context.Background(), local.Name(), remoteTag)
	return remoteTag
}
