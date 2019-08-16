package workers_test

import (
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"bytes"
	"testing"

	"github.com/docker/docker/api/types/mount"
)

type MemWriteCloser struct {
	bytes.Buffer
}

func (mwc *MemWriteCloser) Close() error {
	return nil
}

var confStruct = config.BridgrConf{
	Yum: config.Yum{
		Repos: []string{"http://repo1.test"},
		Items: []string{"mypackage"},
	},
}

var memBuffer = MemWriteCloser{bytes.Buffer{}}

var yumStub = workers.Yum{
	Config:     &confStruct,
	RepoWriter: &memBuffer,
	PackageMount: mount.Mount{
		Type:   mount.TypeBind,
		Source: "/dev/null",
		Target: "/packages",
	},
	RepoMount: mount.Mount{
		Type:   mount.TypeBind,
		Source: "/dev/zero",
		Target: "/etc/yum.repos.d/bridgr.repo",
	},
}

func TestYumSetup(t *testing.T) {
	err := yumStub.Setup()
	if err != nil {
		t.Errorf("Error during Yum.Setup(): %s", err)
	}
	if memBuffer.Len() <= 0 {
		t.Error("Expected content in the yum.repo file, but got size 0")
	}
}

func TestYumName(t *testing.T) {
	y := workers.Yum{}
	if y.Name() != "Yum" {
		t.Errorf("Yum worker does not provide the correct Name() response (%s)", y.Name())
	}
}
