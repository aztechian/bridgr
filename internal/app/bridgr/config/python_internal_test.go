package config

import (
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

var pyTestAltImg, _ = reference.ParseNormalizedNamed("python:3.7.5-alpine")
var pyTestImg, _ = reference.ParseNormalizedNamed("python:2")

// there are other unexported fields in Named interface implementations. We don't really care
// this comparer says we just care about the final string outputted by Named
var opt = cmp.Comparer(func(got, want Python) bool {
	return got.Image.String() == want.Image.String() && cmp.Equal(got.Items, want.Items)
})

func TestParsePython(t *testing.T) {
	tests := []struct {
		name   string
		data   tempConfig
		expect Python
	}{
		{"array of packages", tempConfig{Python: []interface{}{"package1", "package2"}}, Python{Items: []string{"package1", "package2"}, Image: pyTestImg}},
		{"map of pkg versions", tempConfig{Python: map[interface{}]interface{}{"packages": []interface{}{map[interface{}]interface{}{"package": "pkg", "version": ">1.0"}}}}, Python{Items: []string{"pkg>1.0"}, Image: pyTestImg}},
		{"map with version", tempConfig{Python: map[interface{}]interface{}{"version": "3.7.5-alpine", "packages": []interface{}{"pkg"}}}, Python{Items: []string{"pkg"}, Image: pyTestAltImg}},
		{"error - bad image version", tempConfig{Python: map[interface{}]interface{}{"version": "9#ks", "packages": []interface{}{"pkg"}}}, Python{Items: []string{"pkg"}, Image: pyTestImg}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			y := parsePython(test.data)
			if !cmp.Equal(y, test.expect, opt) {
				t.Errorf("Python config not parsed correctly. Expected %+v but got %+v", test.expect, y)
			}
		})
	}
}

func TestPythonParsePackages(t *testing.T) {
	tests := []struct {
		name string
		in   []interface{}
		want Python
	}{
		{"simple string", []interface{}{"mypackage"}, Python{Items: []string{"mypackage"}, Image: pyTestImg}},
		{"map", []interface{}{map[interface{}]interface{}{"package": "testing", "version": ">1.2.3"}}, Python{Items: []string{"testing>1.2.3"}, Image: pyTestImg}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			py := Python{Image: pyTestImg}
			err := py.parsePackages(test.in)
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(py, test.want, opt) {
				t.Errorf("Python config not parsed correctly. Expected %+v but got %+v", test.want, py)
			}
		})
	}
}
