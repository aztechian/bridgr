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

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Yum is the worker implementation for Yum repositories
type Yum struct {
	Config       config.BridgrConf
	RepoWriter   io.WriteCloser
	RepoTemplate string
	packageMount mount.Mount
	repoMount    mount.Mount
}

// NewYum creates a worker.Yum struct
func NewYum(conf config.BridgrConf) (*Yum, error) {
	y := Yum{Config: conf}
	os.MkdirAll(conf.Yum.BaseDir(), os.ModePerm)
	repo, err := os.Create(path.Join(config.BaseDir(), "bridgr.repo"))
	if err != nil {
		return &y, fmt.Errorf("Error creating repo file bridgr.repo: %s", err)
	}
	// log.Printf("Created %s for writing repo template\n", repo.Name())
	y.RepoWriter = repo
	y.packageMount = mount.Mount{
		Type:   mount.TypeBind,
		Source: conf.Yum.BaseDir(),
		Target: "/packages",
	}
	y.repoMount = mount.Mount{
		Type:   mount.TypeBind,
		Source: repo.Name(),
		Target: "/etc/yum.repos.d/bridgr.repo",
	}
	return &y, nil
}

// Run sets up, creates and fetches a YUM repository based on the settings from the config file
func (y *Yum) Run() error {
	y.Setup()
	ctx := context.Background()
	cli, _ := client.NewEnvClient()
	// log.Printf("%+v", cli)
	cli.ContainerRemove(ctx, "bridgr_yum", types.ContainerRemoveOptions{Force: true})

	_, err := cli.ImagePull(ctx, "docker.io/"+y.Config.Yum.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	log.Println("Setting up container to populate YUM repository...")

	script, _ := y.script()
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        y.Config.Yum.Image,
		Cmd:          []string{"/bin/bash", "-"},
		Tty:          false,
		OpenStdin:    true,
		AttachStdout: true,
		AttachStderr: true,
		StdinOnce:    true,
	}, &container.HostConfig{
		Mounts: []mount.Mount{
			y.packageMount,
			y.repoMount,
		},
	}, nil, "bridgr_yum")
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
	io.Copy(hijack.Conn, bytes.NewBufferString(script))
	hijack.Conn.Close()

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		panic(err)
	}
	defer out.Close()
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

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
