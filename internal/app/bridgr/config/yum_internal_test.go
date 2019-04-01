package config

import (
	"testing"
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
	c := tempConfig{
		Yum: []interface{}{"repo1"},
	}
	y := parseYum(c)
	if y.Items[0] != "repo1" {
		t.Errorf("YUM config is incorrect %+v", y)
	}
}
