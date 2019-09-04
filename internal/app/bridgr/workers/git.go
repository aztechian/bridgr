package workers

import (
	"bridgr/internal/app/bridgr/config"
	"os"
)

// Git is a struct that implements a Worker interface for fetching Git artifacts
type Git struct {
	Config config.Git
}

// NewGit creates a new Git worker struct
func NewGit(conf *config.BridgrConf) Worker {
	_ = os.MkdirAll(conf.Git.BaseDir(), os.ModePerm)
	return &Git{Config: conf.Git}
}

// Name returns the friendly name of the Git struct
func (g *Git) Name() string {
	return "Git"
}

// Setup does any initial setup for the Git worker
func (g *Git) Setup() error {
	return nil
}

// Run executes the Git worker to fetch artifacts
func (g *Git) Run() error {
	err := g.Setup()
	if err != nil {
		return err
	}
	return nil
}
