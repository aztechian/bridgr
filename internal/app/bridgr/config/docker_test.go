package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"path"
	"testing"
)

func TestDockerBaseDir(t *testing.T) {
	expected := path.Join(config.BaseDir(), "docker")
	docker := config.Docker{}
	tested := docker.BaseDir()
	if tested != expected {
		t.Errorf("Expected %s but got %s", expected, tested)
	}
}
