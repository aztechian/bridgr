package bridgr_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

func TestYumImage(t *testing.T) {
	dflt, _ := reference.ParseNormalizedNamed("centos:7")
	yum := bridgr.Yum{}
	if !cmp.Equal(dflt, yum.Image(), namedComparer) {
		t.Error(cmp.Diff(dflt, yum.Image()))
	}

	yum2 := bridgr.Yum{Version: dflt}
	if !cmp.Equal(dflt, yum2.Image(), namedComparer) {
		t.Error(cmp.Diff(dflt, yum2.Image()))
	}
}

func TestYumName(t *testing.T) {
	expected := "yum"
	yum := bridgr.Yum{}
	if !cmp.Equal(expected, yum.Name()) {
		t.Errorf(cmp.Diff(expected, yum.Name()))
	}
}

func TestYumHook(t *testing.T) {
	yum := bridgr.Yum{}
	result := reflect.TypeOf(yum.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(result.Name(), reflect.Func))
	}
}
