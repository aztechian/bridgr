package workers

import (
	"bridgr/internal/app/bridgr/config"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path"
)

// Yum is the worker implementation for Yum repositories
type Yum struct {
	Config     config.BridgrConf
	RepoWriter io.Writer
}

// NewYum creates a worker.Yum struct
func NewYum(conf config.BridgrConf) (*Yum, error) {
	y := Yum{Config: conf}
	os.Mkdir(conf.Yum.BaseDir(), os.ModePerm)
	repo, err := os.Create(path.Join(conf.Yum.BaseDir(), "bridgr.repo"))
	if err != nil {
		return &y, fmt.Errorf("Error creating repo file bridgr.repo: %s", err)
	}
	y.RepoWriter = repo // TODO: where to close repo file?
	return &y, nil
}

// Run sets up, creates and fetches a YUM repository based on the settings from the config file
func (y *Yum) Run() error {
	y.Setup()
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
	tmpl := template.New("yumrepo")
	tmpl.Parse(yumTmpl)
	return tmpl.Execute(y.RepoWriter, y.Config.Yum.Repos)
}
