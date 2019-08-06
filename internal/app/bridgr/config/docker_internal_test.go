package config

import (
	"strings"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

func TestDockerParseItems(t *testing.T) {
	// there are other unexported fields in Named interface implementations. We don't really care
	// this comparer says we just care about the final string outputted by Named
	opt := cmp.Comparer(func(got, want reference.Named) bool {
		return got.String() == want.String()
	})
	simpleRef, _ := reference.ParseNormalizedNamed("myimage:1.2")
	longRef, _ := reference.ParseNormalizedNamed("myrepo.io/project/image:4.3-alpine")

	tests := []struct {
		name   string
		item   []interface{}
		expect Docker
	}{
		{"simple string", []interface{}{simpleRef.String()}, Docker{Items: []reference.Named{simpleRef}}},
		{"string with server", []interface{}{longRef.String()}, Docker{Items: []reference.Named{longRef}}},
		{"complex obj", []interface{}{map[interface{}]interface{}{"image": "myimage", "version": "1.2"}}, Docker{Items: []reference.Named{simpleRef}}},
		{"multiple complex obj", []interface{}{map[interface{}]interface{}{"image": "myimage", "version": "1.2"}, map[interface{}]interface{}{"host": "myrepo.io", "image": "project/image", "version": "4.3-alpine"}}, Docker{Items: []reference.Named{simpleRef, longRef}}},
		{"error - no image name", []interface{}{map[interface{}]interface{}{"host": "myhost", "version": "1.3"}}, Docker{}},
		{"error - bad spec complex", []interface{}{map[interface{}]interface{}{"image": "mypath/"}}, Docker{}},
		{"error - bad spec simple", []interface{}{"mypath/"}, Docker{}},
		{"error - unknown interface", []interface{}{map[string]int{"host": 42, "version": 13}}, Docker{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := Docker{}
			_ = d.parseItems(test.item)
			if !cmp.Equal(d, test.expect, opt) {
				t.Errorf("Docker config not parsed correctly. Expected %+v but got %+v", test.expect, d)
			}
		})
	}
}

func TestDockerParseComplex(t *testing.T) {
	// there are other unexported fields in Named interface implementations. We don't really care
	// this comparer says we just care about the final string outputted by Named
	opt := cmp.Comparer(func(got, want reference.Named) bool {
		return got.String() == want.String()
	})

	simpleRef, _ := reference.ParseNormalizedNamed("myimage:1.3")
	hostRef, _ := reference.ParseNormalizedNamed("myrepo.io/awesome")
	intVersionRef, _ := reference.ParseNormalizedNamed("myimage:4")

	tests := []struct {
		name   string
		item   map[interface{}]interface{}
		expect Docker
	}{
		{"image and verison", map[interface{}]interface{}{"image": "myimage", "version": "1.3"}, Docker{Items: []reference.Named{simpleRef}}},
		{"host no version", map[interface{}]interface{}{"host": "myrepo.io", "image": "awesome"}, Docker{Items: []reference.Named{hostRef}}},
		{"error - missing image", map[interface{}]interface{}{"version": "x"}, Docker{}},
		{"error - unsupported version", map[interface{}]interface{}{"image": "myimage", "version": 1.3}, Docker{}},
		{"version with int", map[interface{}]interface{}{"image": "myimage", "version": 4}, Docker{Items: []reference.Named{intVersionRef}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := Docker{}
			err := d.parseComplex(test.item)
			if strings.Contains(test.name, "error -") && err == nil {
				t.Errorf("Expected a test failure for %s, but none was recieved", test.name)
			}
			if !cmp.Equal(d, test.expect, opt) {
				t.Errorf("Docker config not parsed correctly. Expected %+v but got %+v", test.expect, d)
			}
		})
	}
}

func TestParseDocker(t *testing.T) {
	// there are other unexported fields in Named interface implementations. We don't really care
	// this comparer says we just care about the final string outputted by Named
	opt := cmp.Comparer(func(got, want reference.Named) bool {
		return got.String() == want.String()
	})

	simpleRef, _ := reference.ParseNormalizedNamed("myimage:1.3")
	hostRef, _ := reference.ParseNormalizedNamed("myrepo.io/awesome")

	tests := []struct {
		name   string
		in     tempConfig
		expect Docker
	}{
		{"only string array", tempConfig{Docker: []interface{}{"myimage:1.3"}}, Docker{Items: []reference.Named{simpleRef}}},
		{"only complex array", tempConfig{Docker: []interface{}{map[interface{}]interface{}{"image": "myimage:1.3"}, map[interface{}]interface{}{"image": "awesome", "host": "myrepo.io"}}}, Docker{Items: []reference.Named{simpleRef, hostRef}}},
		{"nil array", tempConfig{Docker: nil}, Docker{}},
		{"error - bad section", tempConfig{Docker: []int{34}}, Docker{}},
		{"map", tempConfig{Docker: map[interface{}]interface{}{"repository": "somewhere.beer", "images": []interface{}{"myimage:1.3"}}}, Docker{Destination: "somewhere.beer", Items: []reference.Named{simpleRef}}},
		{"nil", tempConfig{Docker: nil}, Docker{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseDocker(test.in)
			if !cmp.Equal(result, test.expect, opt) {
				t.Errorf("Expected %v in Docker struct, got %v", test.expect.Items, result.Items)
			}
		})
	}
}

func TestDockerParseSimple(t *testing.T) {
	d := Docker{}
	err := d.parseSimple("mypath/")
	if err == nil {
		t.Errorf("Expected an error from ParseNormalizedName(), but didn't get one.")
	}
}
