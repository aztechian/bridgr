package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseRepos(t *testing.T) {
	tests := []struct {
		name   string
		data   []interface{}
		expect int
	}{
		{"single item", []interface{}{"package1"}, 1},
		{"multiple items", []interface{}{"package1", "package2"}, 2},
		{"non-string value", []interface{}{"package4", 4}, 1},
		{"nil", nil, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			yum := Yum{}
			yum.parseRepos(test.data)
			if len(yum.Repos) != test.expect {
				t.Errorf("Expected %d repos in File struct, got %d", test.expect, len(yum.Repos))
			}
		})
	}
}

func TestParsePackages(t *testing.T) {
	tests := []struct {
		name   string
		data   []interface{}
		expect int
	}{
		{"single item", []interface{}{"package1"}, 1},
		{"multiple item", []interface{}{"package1", "package2", "package3"}, 3},
		{"mixed items", []interface{}{"package1", 4.32, nil}, 1},
		{"nil", nil, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			yum := Yum{}
			yum.parsePackages(test.data)
			if len(yum.Items) != test.expect {
				t.Errorf("Expected %d items in File struct, got %d", test.expect, len(yum.Items))
			}
		})
	}
}

func TestParseYum(t *testing.T) {
	tests := []struct {
		name   string
		data   tempConfig
		expect Yum
	}{
		{"array of packages", tempConfig{Yum: []interface{}{"package1", "package2"}}, Yum{Repos: nil, Items: []string{"package1", "package2"}, Image: "library/centos:7"}},
		{"map with repos", tempConfig{Yum: map[interface{}]interface{}{"repos": []interface{}{"testrepo"}, "packages": []interface{}{"pkg"}}}, Yum{Repos: []string{"testrepo"}, Items: []string{"pkg"}, Image: "library/centos:7"}},
		{"map with image", tempConfig{Yum: map[interface{}]interface{}{"image": "my/centos:1.0", "repos": []interface{}{"testrepo"}, "packages": []interface{}{"pkg"}}}, Yum{Repos: []string{"testrepo"}, Items: []string{"pkg"}, Image: "my/centos:1.0"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			y := parseYum(test.data)
			if !cmp.Equal(y, test.expect) {
				t.Errorf("YUM config not parsed correctly. Expected %+v but got %+v", test.expect, y)
			}
		})
	}
}
