package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"path"
	"testing"
)

func TestRubyBaseDir(t *testing.T) {
	expected := path.Join(config.BaseDir(), "ruby")
	ruby := config.Ruby{}
	tested := ruby.BaseDir()
	if tested != expected {
		t.Errorf("Expected %s but got %s", expected, tested)
	}
}
