package bridgr_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/google/go-cmp/cmp"
)

func TestHelmImage(t *testing.T) {
	helm := bridgr.Helm{}
	if helm.Image() != nil {
		t.Errorf("expected nil, but got %+v", helm.Image())
	}
}

func TestHelmName(t *testing.T) {
	expected := "helm"
	helm := bridgr.Helm{}
	if !cmp.Equal(expected, helm.Name()) {
		t.Error(cmp.Diff(expected, helm.Name()))
	}
}

func TestHelmHook(t *testing.T) {
	helm := bridgr.Helm{}
	result := reflect.TypeOf(helm.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(result.Name(), reflect.Func))
	}
}
