package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"bytes"
	"os"
	"path"
	"strings"
	"testing"
)

type MemReadCloser struct {
	bytes.Buffer
}

func (mwc *MemReadCloser) Close() error {
	return nil
}

var validConfig = `---
yum:
  - package1
`

func TestNew(t *testing.T) {
	content := bytes.NewBufferString("helloworld")
	configFile := MemReadCloser{*content}
	c, err := config.New(&configFile)
	if err != nil {
		t.Error("Error creating new Config")
	}
	if len(c.Yum.Items) > 0 {
		t.Error("Yum config is magically populated")
	}
}

func TestBaseDir(t *testing.T) {
	v := config.BaseDir()
	expect, _ := os.Getwd()
	if len(v) == 0 {
		t.Error("BaseDir() has 0 length string")
	}
	if !strings.HasPrefix(v, expect) {
		t.Errorf("Expected BaseDir prefix of %s, but got %s", expect, v)
	}
	if v != path.Join(expect, "packages") {
		t.Errorf("Expected BaseDir to be %s, but got %s", path.Join(expect, "packages"), v)
	}
}
