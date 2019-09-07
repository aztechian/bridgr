package workers_test

import (
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGitName(t *testing.T) {
	g := workers.Git{}
	if g.Name() != "Git" {
		t.Errorf("Expected Name() to be Git but got %s", g.Name())
	}
}

func TestNewGit(t *testing.T) {
	c := &config.BridgrConf{Git: config.Git{}}
	worker := workers.NewGit(c)
	git, _ := worker.(*workers.Git)
	if !cmp.Equal(&c.Git, git.Config) {
		t.Errorf("Unexpected Config object: %s", cmp.Diff(&c.Git, git.Config))
	}
}

func TestGitSetup(t *testing.T) {
	g := workers.Git{}
	err := g.Setup()
	if err != nil {
		t.Error(err)
	}
}

func TestGitRun(t *testing.T) {
	c := &config.BridgrConf{Git: config.Git{}}
	g := workers.NewGit(c)
	err := g.Run()
	if err != nil {
		t.Error(err)
	}
}
