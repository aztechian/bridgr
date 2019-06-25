package config_test

import (
	"bridgr/internal/app/bridgr/config"
	"os"
	"strings"
	"testing"
)

func TestYumBaseDir(t *testing.T) {
	yum := config.Yum{}
	v := yum.BaseDir()
	expect, _ := os.Getwd()
	if len(v) == 0 {
		t.Error("BaseDir() has 0 length string")
	}
	if !strings.HasPrefix(v, expect) {
		t.Errorf("Expected BaseDir prefix of %s, but got %s", expect, v)
	}
	if !strings.HasSuffix(v, "/yum") {
		t.Errorf("Expected BaseDir to end with \"/yum\", but got %s", v)
	}
}
