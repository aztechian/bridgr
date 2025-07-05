package bridgr_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

func TestRubyImage(t *testing.T) {
	dflt, _ := reference.ParseNormalizedNamed("ruby:2-alpine")
	ruby := bridgr.Ruby{}
	if !cmp.Equal(dflt, ruby.Image(), namedComparer) {
		t.Error(cmp.Diff(dflt, ruby.Image()))
	}

	rb := bridgr.Ruby{Version: dflt}
	if !cmp.Equal(dflt, rb.Image(), namedComparer) {
		t.Error(cmp.Diff(dflt, rb.Image()))
	}
}

func TestRubyName(t *testing.T) {
	expected := "ruby"
	ruby := bridgr.Ruby{}
	if !cmp.Equal(expected, ruby.Name()) {
		t.Error(cmp.Diff(expected, ruby.Name()))
	}
}

func TestRubyHook(t *testing.T) {
	ruby := bridgr.Ruby{}
	result := reflect.TypeOf(ruby.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(result.Name(), reflect.Func))
	}
}
