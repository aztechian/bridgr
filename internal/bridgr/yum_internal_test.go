package bridgr

import (
	"reflect"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

func TestYumDir(t *testing.T) {
	expected := BaseDir("yum")
	result := Yum{}.dir()
	if !cmp.Equal(expected, result) {
		t.Error(cmp.Diff(expected, result))
	}
}

func TestVersionToYumImage(t *testing.T) {
	img, _ := reference.ParseNormalizedNamed("centos:looseseal")
	img2, _ := reference.ParseNormalizedNamed("centos:2.0")
	img3, _ := reference.ParseNormalizedNamed("lucile:2")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid image", reflect.TypeOf(""), "", false, ""},
		{"valid", reflect.TypeOf((*yumVersion)(&img)).Elem(), "looseseal", false, img},
		{"invalid type", reflect.TypeOf((*yumVersion)(&img2)).Elem(), 12, true, img2},
		{"tagged image", reflect.TypeOf((*yumVersion)(&img2)).Elem(), "lucile:2", false, img3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := versionToYumImage(reflect.TypeOf(test.input), test.target, test.input)
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

func TestArrayToYum(t *testing.T) {
	packages := []interface{}{"buster", "awpoorbuster", "milfordman"}
	version, _ := reference.ParseNormalizedNamed("centos:7")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid target", reflect.TypeOf(4.23), "monster", false, "monster"},
		{"invalid input", reflect.TypeOf(Yum{}), 33, false, 33},
		{"valid", reflect.TypeOf(Yum{}), packages, false, Yum{Version: version, Packages: []string{"buster", "awpoorbuster", "milfordman"}}},
		{"invalid array", reflect.TypeOf(Yum{}), []interface{}{83, 9.4822}, false, Yum{Version: version}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := arrayToYum(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(test.expect, result, namedComparer) && !test.isError {
				t.Error(cmp.Diff(test.expect, result, namedComparer))
			}
		})
	}
}
