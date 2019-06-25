package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"bytes"
	"errors"
	"os"
	"path"
	"strings"
	"testing"
)

type MemReadCloser struct {
	bytes.Buffer
}

type ErrMemReadCloser struct {
	bytes.Buffer
}

func (mwc *MemReadCloser) Close() error {
	return nil
}

func (emwc *ErrMemReadCloser) Close() error {
	return nil
}
func (ErrMemReadCloser) Read(p []byte) (n int, err error) {
	return 0, errors.New("test error")
}

var validConfig = `---
yum:
  - package1
`

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		expected int
	}{
		{"invalid config", "helloworld", 0},
		{"yaml config", validConfig, 1},
		{"read error", "", 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			content := bytes.NewBufferString(test.data)
			if test.data == "" {
				configFile := ErrMemReadCloser{*content}
				_, err := config.New(&configFile)
				if err == nil {
					t.Error("Read error from config.New should have raised error")
				}
			} else {
				configFile := MemReadCloser{*content}
				c, err := config.New(&configFile)
				if err != nil {
					t.Error("Error creating new Config")
				}
				if len(c.Yum.Items) != test.expected {
					t.Errorf("Yum config has %d items, expected %d", len(c.Yum.Items), test.expected)
				}
			}
		})
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
