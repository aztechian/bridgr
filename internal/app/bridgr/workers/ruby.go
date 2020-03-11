package workers

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/config"
	"io"
	"os"
	"path"
	"text/template"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
)

// Ruby is the struct defining a Worker for Ruby
type Ruby struct {
	Config          *config.Ruby
	ReqtWriter      io.WriteCloser
	ShellWriter     io.WriteCloser
	RepoTemplate    string
	PackageMount    mount.Mount
	RepoMount       mount.Mount
	ContainerConfig *container.Config
}

// NewRuby is the constructor to create a Ruby Worker
func NewRuby(conf *config.BridgrConf) Worker {
	_ = os.MkdirAll(conf.Ruby.BaseDir(), os.ModePerm)
	reqt, err := os.Create(path.Join(config.BaseDir(), "Gemfile"))
	if err != nil {
		bridgr.Printf("Unable to create Ruby Gemfile: %s", err)
		return nil
	}
	bridgr.Debugf("Created %s for writing Gemfile template", reqt.Name())
	return &Ruby{
		Config:     conf.Ruby,
		ReqtWriter: reqt,
		PackageMount: mount.Mount{
			Type:   mount.TypeBind,
			Source: conf.Ruby.BaseDir(),
			Target: "/packages",
		},
		RepoMount: mount.Mount{
			Type:   mount.TypeBind,
			Source: reqt.Name(),
			Target: "/Gemfile",
		},
		ContainerConfig: &container.Config{
			Image:        conf.Ruby.Image().String(),
			Cmd:          []string{"/bin/sh", "-"},
			Tty:          false,
			OpenStdin:    true,
			AttachStdout: true,
			AttachStderr: true,
			StdinOnce:    true,
		},
	}
}

// Name returns the name of this Python worker
func (r *Ruby) Name() string {
	return "Ruby"
}

// Setup creates the items that are needed to fetch artifacts for the Python worker. It does not actually fetch artifacts.
func (r *Ruby) Setup() error {
	err := r.writeGemfile()
	if err != nil {
		return err
	}
	return nil
}

// Run fetches all artifacts for the Python configuration
func (r *Ruby) Run() error {
	bridgr.Debug("Called Ruby.Setup()")
	err := r.Setup()
	if err != nil {
		return err
	}
	shell, err := r.script()
	if err != nil {
		return err
	}
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			r.PackageMount,
			r.RepoMount,
		},
	}
	return runContainer("bridgr_ruby", r.ContainerConfig, &hostConfig, shell)
}

func (r *Ruby) writeGemfile() error {
	gemTmpl, err := loadTemplate("Gemfile")
	if err != nil {
		return err
	}
	defer r.ReqtWriter.Close()
	tmpl, _ := template.New("rubygems").Parse(gemTmpl)
	return tmpl.Execute(r.ReqtWriter, r.Config)
}

func (r *Ruby) script() (string, error) {
	rbSh, err := loadTemplate("ruby.sh")
	if err != nil {
		return "", err
	}
	return rbSh, nil
}
