package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"path"
	"testing"
)

func TestFilesBaseDir(t *testing.T) {
	expected := path.Join(config.BaseDir(), "files")
	files := config.Files{}
	tested := files.BaseDir()
	if tested != expected {
		t.Errorf("Expected %s but got %s", expected, tested)
	}
}
