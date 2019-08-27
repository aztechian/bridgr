package workers_test

import (
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"bytes"
	"strings"
	"testing"

	"github.com/docker/distribution/reference"
)

var reqtBuffer = MemWriteCloser{bytes.Buffer{}}
var defaultImg, _ = reference.ParseNormalizedNamed("python:2")
var pyStub = workers.Python{
	Config:     &config.Python{Items: []string{"simple-package"}, Image: defaultImg},
	ReqtWriter: &reqtBuffer,
}

func TestPythonSetup(t *testing.T) {
	err := pyStub.Setup()
	if err != nil {
		t.Errorf("Error during Python.Setup(): %s", err)
	}
	if reqtBuffer.Len() <= 0 {
		t.Error("Expected content in requirements.txt, but got size 0")
	}
	if !strings.Contains(reqtBuffer.String(), "simple-package") {
		t.Error("requirements.txt does not contain simple-package")
	}
}

func TestPythonRun(t *testing.T) {
}

func TestPythonName(t *testing.T) {
	py := workers.Python{}
	if py.Name() != "Python" {
		t.Errorf("Yum worker does not provide the correct Name() response (%s)", py.Name())
	}
}
