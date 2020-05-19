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
	if !cmp.Equal(yum.Image(), dflt, namedComparer) {
		t.Error(cmp.Diff(yum.Image(), dflt))
	}

	yum2 := bridgr.Yum{Version: dflt}
	if !cmp.Equal(yum2.Image(), dflt, namedComparer) {
		t.Error(cmp.Diff(yum2.Image(), dflt))
	}
}

func TestYumName(t *testing.T) {
	yum := bridgr.Yum{}
	if !cmp.Equal(yum.Name(), "yum") {
		t.Errorf(cmp.Diff(yum.Name(), "yum"))
	}
}

func TestYumHook(t *testing.T) {
	yum := bridgr.Yum{}
	result := reflect.TypeOf(yum.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(result.Name(), reflect.Func))
	}
}
