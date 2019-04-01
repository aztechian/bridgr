package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"testing"
)

func TestBaseDir(t *testing.T) {
	expected := "files"
	files := config.Files{}
	tested := files.BaseDir()
	if tested != expected {
		t.Errorf("Expected %s but got %s", expected, tested)
	}
}
