package bridgr_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

func TestRubyImage(t *testing.T) {
	dflt, _ := reference.ParseNormalizedNamed("ruby:2-alpine")
	ruby := bridgr.Ruby{}
	if !cmp.Equal(ruby.Image(), dflt, namedComparer) {
		t.Error(cmp.Diff(ruby.Image(), dflt))
	}

	rb := bridgr.Ruby{Version: dflt}
	if !cmp.Equal(rb.Image(), dflt, namedComparer) {
		t.Error(cmp.Diff(rb.Image(), dflt))
	}
}

func TestRubyName(t *testing.T) {
	ruby := bridgr.Ruby{}
	if !cmp.Equal(ruby.Name(), "ruby") {
		t.Errorf(cmp.Diff(ruby.Name(), "ruby"))
	}
}

func TestRubyHook(t *testing.T) {
	ruby := bridgr.Ruby{}
	result := reflect.TypeOf(ruby.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(result.Name(), reflect.Func))
	}
}
