package config

import (
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

var rbTestAltImg, _ = reference.ParseNormalizedNamed("ruby:2.6.4")
var rbTestImg, _ = reference.ParseNormalizedNamed("ruby:2-alpine")

func TestParseRuby(t *testing.T) {
	tests := []struct {
		name   string
		data   tempConfig
		expect Ruby
	}{
		{"array of packages", tempConfig{Ruby: []interface{}{"package1", "package2"}}, Ruby{Items: []RubyItem{{Package: "package1"}, {Package: "package2"}}, Image: rbTestImg}},
		{"map of pkg versions", tempConfig{Ruby: map[interface{}]interface{}{"gems": []interface{}{map[interface{}]interface{}{"package": "pkg", "version": "~3.3"}}}}, Ruby{Items: []RubyItem{{Package: "pkg", Version: "~3.3"}}, Image: rbTestImg}},
		{"map with version", tempConfig{Ruby: map[interface{}]interface{}{"version": "2.6.4", "gems": []interface{}{"pkg"}}}, Ruby{Items: []RubyItem{{Package: "pkg"}}, Image: rbTestAltImg}},
		{"map with source", tempConfig{Ruby: map[interface{}]interface{}{"sources": []interface{}{"gemsource"}, "gems": []interface{}{"pkg"}}}, Ruby{Items: []RubyItem{{Package: "pkg"}}, Sources: []string{"gemsource"}, Image: rbTestImg}},
		{"error - bad image version", tempConfig{Ruby: map[interface{}]interface{}{"version": "9#ks", "gems": []interface{}{"pkg"}}}, Ruby{Items: []RubyItem{{Package: "pkg"}}, Image: rbTestImg}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rb := parseRuby(test.data)
			if !cmp.Equal(rb, test.expect, opt) {
				t.Errorf("Ruby config not parsed correctly. Expected %+v but got %+v", test.expect, rb)
			}
		})
	}
}

func TestRubyParsePackages(t *testing.T) {
	tests := []struct {
		name string
		in   []interface{}
		want Ruby
	}{
		{"simple string", []interface{}{"mypackage"}, Ruby{Items: []RubyItem{{Package: "mypackage"}}, Image: rbTestImg}},
		{"map", []interface{}{map[interface{}]interface{}{"package": "testing", "version": ">1.2.3"}}, Ruby{Items: []RubyItem{{Package: "testing", Version: ">1.2.3"}}, Image: rbTestImg}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rb := Ruby{Image: rbTestImg}
			err := rb.parsePackages(test.in)
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(rb, test.want, opt) {
				t.Errorf("Ruby config not parsed correctly. Expected %+v but got %+v", test.want, rb)
			}
		})
	}
}

func TestRubyAddSources(t *testing.T) {
	tests := []struct {
		name string
		in   []interface{}
		want Ruby
	}{
		{"single source", []interface{}{"http://testserver.net"}, Ruby{Sources: []string{"http://testserver.net"}, Image: rbTestImg}},
		{"multiple sources", []interface{}{"http://testserver.net", "https://corpserver.com"}, Ruby{Sources: []string{"http://testserver.net", "https://corpserver.com"}, Image: rbTestImg}},
		{"no sources", []interface{}{}, Ruby{Image: rbTestImg}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rb := Ruby{Image: rbTestImg}
			err := rb.addSources(test.in)
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(rb, test.want, opt) {
				t.Errorf("Ruby config not parsed correctly. Expected %+v but got %+v", test.want, rb)
			}
		})
	}
}
