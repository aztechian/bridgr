package bridgr_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/google/go-cmp/cmp"
)

func TestDockerImage(t *testing.T) {
	docker := bridgr.Docker{}
	if docker.Image() != nil {
		t.Errorf("expected nil, but got %+v", docker.Image())
	}
}

func TestDockerName(t *testing.T) {
	docker := bridgr.Docker{}
	if !cmp.Equal(docker.Name(), "docker") {
		t.Errorf(cmp.Diff(docker.Name(), "docker"))
	}
}

func TestDockerHook(t *testing.T) {
	docker := bridgr.Docker{}
	result := reflect.TypeOf(docker.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(result.Name(), reflect.Func))
	}
}
