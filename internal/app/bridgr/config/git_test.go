package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"path"
	"testing"
)

func TestGitBaseDir(t *testing.T) {
	expected := path.Join(config.BaseDir(), "git")
	git := config.Git{}
	tested := git.BaseDir()
	if tested != expected {
		t.Errorf("Expected %s but got %s", expected, tested)
	}
}
