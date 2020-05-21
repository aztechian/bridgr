package cmd

import (
	"net/url"
	"reflect"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

var namedComparer = cmp.Comparer(func(got, want reference.Named) bool {
	return got.String() == want.String()
})

func TestStringToImage(t *testing.T) {
	img, _ := reference.ParseNormalizedNamed("tobias:nevernude")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid type", reflect.TypeOf(39), 39, false, 39},
		{"valid", reflect.TypeOf((*reference.Reference)(nil)).Elem(), "tobias:nevernude", false, img},
		{"invalid image", reflect.TypeOf((*reference.Reference)(nil)).Elem(), "", true, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := stringToImage(reflect.TypeOf(test.input), test.target, test.input)
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

func TestStringToURL(t *testing.T) {
	expectedURL, _ := url.Parse("tobias.com/analrapist.html")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid image", reflect.TypeOf(4.302), 4.302, false, 4.302},
		{"valid", reflect.TypeOf(&url.URL{}), expectedURL.String(), false, expectedURL},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := stringToURL(reflect.TypeOf(test.input), test.target, test.input)
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

func TestDebugHook(t *testing.T) {
	result, err := debugHook(reflect.TypeOf(false), reflect.TypeOf(42), false)
	if !cmp.Equal(false, result) {
		t.Error(cmp.Diff(false, result))
	}
	if err != nil {
		t.Error(err)
	}
}
