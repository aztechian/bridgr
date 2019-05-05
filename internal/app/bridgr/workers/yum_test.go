package workers_test

import (
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"bytes"
	"testing"
)

var confStruct = config.BridgrConf{
	Yum: config.Yum{
		Repos: []string{"http://repo1.test"},
		Items: []string{"mypackage"},
	},
}

var memBuffer = bytes.Buffer{}

var yumStub = workers.Yum{
	Config:     confStruct,
	RepoWriter: &memBuffer,
}

func TestRun(t *testing.T) {

}

func TestSetup(t *testing.T) {
	err := yumStub.Setup()
	if err != nil {
		t.Errorf("uh oh, %s", err)
	}
	// t.Logf("Got repo file of %+s", memBuffer.String())
}
