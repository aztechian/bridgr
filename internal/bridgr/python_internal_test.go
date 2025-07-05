package bridgr

import (
	"reflect"
	"testing"

	"github.com/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

func TestPythonDir(t *testing.T) {
	expected := BaseDir("python")
	result := Python{}.dir()
	if !cmp.Equal(expected, result) {
		t.Error(cmp.Diff(expected, result))
	}
}

func TestVersionToPythonImage(t *testing.T) {
	img, _ := reference.ParseNormalizedNamed("python:looseseal")
	img2, _ := reference.ParseNormalizedNamed("python:2.0")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid image", reflect.TypeOf(""), "", false, ""},
		{"valid", reflect.TypeOf((*pythonVersion)(&img)).Elem(), "looseseal", false, img},
		{"valid float", reflect.TypeOf((*pythonVersion)(&img2)).Elem(), 2.0134, false, img2},
		{"invalid type", reflect.TypeOf((*pythonVersion)(&img2)).Elem(), 12, true, img2},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := versionToPythonImage(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(test.expect, result, namedComparer) && !test.isError {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestArrayToPython(t *testing.T) {
	packages := []interface{}{"buster", "awpoorbuster", "milfordman"}
	src := []string{"https://pypi.org"}
	version, _ := reference.ParseNormalizedNamed("python:3.7")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid type", reflect.TypeOf(""), "", false, ""},
		{"valid", reflect.TypeOf(Python{}), packages, false, Python{Version: version, Sources: src, Packages: []pythonPackage{{Package: "buster"}, {Package: "awpoorbuster"}, {Package: "milfordman"}}}},
		// {"parse error", reflect.TypeOf(Python{}), packages, true, img},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := arrayToPython(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(test.expect, result, namedComparer) && !test.isError {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}
