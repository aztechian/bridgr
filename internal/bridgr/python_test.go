package bridgr_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

var namedComparer = cmp.Comparer(func(got, want reference.Named) bool {
	return got.String() == want.String()
})

func TestPythonImage(t *testing.T) {
	dflt, _ := reference.ParseNormalizedNamed("python:3.7")
	python := bridgr.Python{}
	if !cmp.Equal(python.Image(), dflt, namedComparer) {
		t.Error(cmp.Diff(python.Image(), dflt))
	}

	py := bridgr.Python{Version: dflt}
	if !cmp.Equal(py.Image(), dflt, namedComparer) {
		t.Error(cmp.Diff(py.Image(), dflt))
	}
}

func TestPythonName(t *testing.T) {
	expected := "python"
	python := bridgr.Python{}
	if !cmp.Equal(expected, python.Name()) {
		t.Errorf(cmp.Diff(expected, python.Name()))
	}
}

func TestPythonHook(t *testing.T) {
	python := bridgr.Python{}
	result := reflect.TypeOf(python.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(result.Name(), reflect.Func))
	}
}
