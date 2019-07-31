package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"path"
	"testing"
)

func TestPythonBaseDir(t *testing.T) {
	expected := path.Join(config.BaseDir(), "python")
	python := config.Python{}
	tested := python.BaseDir()
	if tested != expected {
		t.Errorf("Expected %s but got %s", expected, tested)
	}
}
