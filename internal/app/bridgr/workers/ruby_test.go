package workers_test

import (
	"bridgr/internal/app/bridgr/config"
	"bridgr/internal/app/bridgr/workers"
	"strings"
	"testing"

	"github.com/docker/distribution/reference"
)

var rbImage, _ = reference.ParseNormalizedNamed("ruby:2")
var rbStub = workers.Ruby{
	Config:     &config.Ruby{Items: []config.RubyItem{{Package: "simple-package"}}, Image: rbImage},
	ReqtWriter: &reqtBuffer,
}

func TestRubySetup(t *testing.T) {
	err := rbStub.Setup()
	if err != nil {
		t.Errorf("Error during Ruby.Setup(): %s", err)
	}
	if reqtBuffer.Len() <= 0 {
		t.Error("Expected content in Gemfile, but got size 0")
	}
	if !strings.Contains(reqtBuffer.String(), "simple-package") {
		t.Error("Gemfile does not contain simple-package")
	}
}

func TestRubyRun(t *testing.T) {
}

func TestRubyName(t *testing.T) {
	rb := workers.Ruby{}
	if rb.Name() != "Ruby" {
		t.Errorf("Ruby worker does not provide the correct Name() response (%s)", rb.Name())
	}
}
