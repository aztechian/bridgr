package bridgr

import (
	"reflect"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

func TestRubyDir(t *testing.T) {
	expected := BaseDir("ruby")
	result := Ruby{}.dir()
	if !cmp.Equal(result, expected) {
		t.Error(cmp.Diff(result, expected))
	}
}

func TestVersionToRubyImage(t *testing.T) {
	img, _ := reference.ParseNormalizedNamed("ruby:looseseal")
	img2, _ := reference.ParseNormalizedNamed("ruby:2.0")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid image", reflect.TypeOf(""), "", false, ""},
		{"valid", reflect.TypeOf((*rubyVersion)(&img)).Elem(), "looseseal", false, img},
		{"invalid type", reflect.TypeOf((*rubyVersion)(&img2)).Elem(), 12, true, img2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := versionToRubyImage(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(result, test.expect, namedComparer) && !test.isError {
				t.Error(cmp.Diff(result, test.expect))
			}
		})
	}
}

func TestStringToRuby(t *testing.T) {
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid image", reflect.TypeOf(42), 42, false, 42},
		{"valid", reflect.TypeOf(rubyItem{}), "looseseal", false, rubyItem{Package: "looseseal"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := stringToRuby(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(result, test.expect, namedComparer) && !test.isError {
				t.Error(cmp.Diff(result, test.expect))
			}
		})
	}
}

func TestArrayToRuby(t *testing.T) {
	packages := []interface{}{"buster", "awpoorbuster", "milfordman"}
	src := []string{"https://rubygems.org"}
	version, _ := reference.ParseNormalizedNamed("ruby:2-alpine")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid type", reflect.TypeOf(""), "", false, ""},
		{"valid", reflect.TypeOf(Ruby{}), packages, false, Ruby{Version: version, Sources: src, Gems: []rubyItem{{Package: "buster"}, {Package: "awpoorbuster"}, {Package: "milfordman"}}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := arrayToRuby(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(result, test.expect, namedComparer) && !test.isError {
				t.Error(cmp.Diff(result, test.expect))
			}
		})
	}
}

func TestRubyItemString(t *testing.T) {
	tests := []struct {
		name   string
		input  rubyItem
		expect string
	}{
		{"only package", rubyItem{Package: "anustart"}, "anustart"},
		{"package and version", rubyItem{Package: "anew", Version: "start"}, "anew, start"},
		{"only version", rubyItem{Version: "15"}, ", 15"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.input.String()
			if !cmp.Equal(result, test.expect) {
				t.Error(cmp.Diff(result, test.expect))
			}
		})
	}
}
