package workers

import (
	"bridgr/internal/app/bridgr/config"
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path"
	"strings"

	"github.com/docker/docker/api/types/mount"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// Yum is the worker implementation for Yum repositories
type Yum struct {
	Config       config.BridgrConf
	RepoWriter   io.WriteCloser
	RepoTemplate string
	bindMount    mount.Mount
}

// NewYum creates a worker.Yum struct
func NewYum(conf config.BridgrConf) (*Yum, error) {
	y := Yum{Config: conf}
	os.Mkdir(conf.Yum.BaseDir(), os.ModePerm)
	repo, err := os.Create(path.Join(conf.Yum.BaseDir(), "bridgr.repo"))
	if err != nil {
		return &y, fmt.Errorf("Error creating repo file bridgr.repo: %s", err)
	}
	y.RepoWriter = repo
	y.bindMount = mount.Mount{
		Type:   mount.TypeBind,
		Source: "/Users/ian.martin/Documents/code/bridgr/yum", // TODO remove this hardcoding
		Target: "/packages",
	}
	return &y, nil
}

// Run sets up, creates and fetches a YUM repository based on the settings from the config file
func (y *Yum) Run() error {
	y.Setup()
	ctx := context.Background()
	cli, _ := client.NewEnvClient()
	cli.ContainerRemove(ctx, "bridgr_yum", types.ContainerRemoveOptions{Force: true})
	// if err != nil {
	// 	fmt.Printf("Error while deleting container (%s): %s", "bridgr_yum", err)
	// }
	reader, err := cli.ImagePull(ctx, "docker.io/"+y.Config.Yum.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer reader.Close()
	io.Copy(os.Stdout, reader)

	script, _ := y.script()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:     y.Config.Yum.Image,
		Cmd:       []string{"/bin/bash"},
		Tty:       false,
		OpenStdin: true,
		StdinOnce: true,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			y.bindMount,
		},
	}, nil, "bridgr_yum")
	if err != nil {
		panic(err)
	}

	hijack, err := cli.ContainerAttach(ctx, resp.ID, types.ContainerAttachOptions{
		Stream: true,
		Stdin:  true,
	})
	if err != nil {
		return err
	}
	io.Copy(hijack.Conn, bytes.NewBufferString(script))
	hijack.Conn.Close()

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	defer out.Close()
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, out)
	return nil
}

// Setup only does the setup step of the YUM worker
func (y *Yum) Setup() error {
	log.Println("Called Yum.setup()")

	err := y.writeRepos()
	if err != nil {
		return err
	}
	return nil
}

func (y *Yum) writeRepos() error {
	yumTmpl, err := loadTemplate("yum.repo")
	if err != nil {
		log.Printf("Error loading yum.repo template: %s", err)
	}
	defer y.RepoWriter.Close()
	tmpl := template.New("yumrepo")
	tmpl.Parse(yumTmpl)
	return tmpl.Execute(y.RepoWriter, y.Config.Yum.Repos)
}

func (y *Yum) script() (string, error) {
	docker, err := loadTemplate("yum.sh")
	if err != nil {
		log.Printf("Error loading yum.sh template: %s", err)
	}
	tmpl := template.New("yumscript")
	tmpl = tmpl.Funcs(template.FuncMap{"Join": strings.Join})
	tmpl.Parse(docker)
	final := bytes.Buffer{}
	if err := tmpl.Execute(&final, y.Config.Yum.Items); err != nil {
		return "", err
	}
	return final.String(), nil
}
