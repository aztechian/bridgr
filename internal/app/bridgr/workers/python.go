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

// Python is the struct defining a Worker for Python
type Python struct {
	Config          *config.Python
	ReqtWriter      io.WriteCloser
	ShellWriter     io.WriteCloser
	RepoTemplate    string
	PackageMount    mount.Mount
	RepoMount       mount.Mount
	ContainerConfig *container.Config
}

// NewPython is the constructor to create a Python Worker
func NewPython(conf *config.BridgrConf) Worker {
	_ = os.MkdirAll(conf.Python.BaseDir(), os.ModePerm)
	reqt, err := os.Create(path.Join(config.BaseDir(), "requirements.txt"))
	if err != nil {
		bridgr.Printf("Unable to create Python requirements file: %s", err)
		return nil
	}
	bridgr.Debugf("Created %s for writing repo template", reqt.Name())
	return &Python{
		Config:     conf.Python,
		ReqtWriter: reqt,
		PackageMount: mount.Mount{
			Type:   mount.TypeBind,
			Source: conf.Python.BaseDir(),
			Target: "/packages",
		},
		RepoMount: mount.Mount{
			Type:   mount.TypeBind,
			Source: reqt.Name(),
			Target: "/requirements.txt",
		},
		ContainerConfig: &container.Config{
			Image:        conf.Python.Image().String(),
			Cmd:          []string{"/bin/bash", "-"},
			Tty:          false,
			OpenStdin:    true,
			AttachStdout: true,
			AttachStderr: true,
			StdinOnce:    true,
		},
	}
}

// Name returns the name of this Python worker
func (p *Python) Name() string {
	return "Python"
}

// Setup creates the items that are needed to fetch artifacts for the Python worker. It does not actually fetch artifacts.
func (p *Python) Setup() error {
	err := p.writeRequirements()
	if err != nil {
		return err
	}
	return nil
}

// Run fetches all artifacts for the Python configuration
func (p *Python) Run() error {
	err := p.Setup()
	if err != nil {
		return err
	}
	shell, err := p.script()
	if err != nil {
		return err
	}
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			p.PackageMount,
			p.RepoMount,
		},
	}
	return runContainer("bridgr_python", p.ContainerConfig, &hostConfig, shell)
}

func (p *Python) writeRequirements() error {
	reqtTmpl, err := loadTemplate("requirements.txt")
	if err != nil {
		return err
	}
	defer p.ReqtWriter.Close()
	tmpl, _ := template.New("pythonreqt").Parse(reqtTmpl)
	return tmpl.Execute(p.ReqtWriter, p.Config.Packages)
}

func (p *Python) script() (string, error) {
	pySh, err := loadTemplate("python.sh")
	if err != nil {
		return "", err
	}
	return pySh, nil
}
