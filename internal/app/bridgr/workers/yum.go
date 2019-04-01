package workers

import (
	"bridgr/internal/app/bridgr/config"
	"fmt"
	"html/template"
	"log"
	"os"
)

// Yum is the worker implementation for Yum repositories
type Yum struct{}

// Run sets up, creates and fetches a YUM repository based on the settings from the config file
func (y *Yum) Run(conf config.BridgrConf) error {
	y.Setup(conf)
	return nil
}

// Setup only does the setup step of the YUM worker
func (y *Yum) Setup(conf config.BridgrConf) error {
	log.Println("Called Yum.setup()")
	os.Mkdir(conf.Yum.BaseDir(), os.ModePerm)
	err := y.writeRepos("bridgr.repo", conf.Yum.Repos)
	if err != nil {
		return err
	}

	return nil
}

func (y *Yum) writeRepos(repofile string, repolist []string) error {
	repoTemplate, err := template.ParseFiles("internal/app/bridgr/templates/yum.repo")
	if err != nil {
		return fmt.Errorf("Error parsing YUM repo template from templates/yum/repo")
	}
	repo, err := os.Create(repofile)
	if err != nil {
		return fmt.Errorf("Error creating repo file %s: %s", repofile, err)
	}
	defer repo.Close()

	err = repoTemplate.Execute(repo, repolist)
	return err
}
