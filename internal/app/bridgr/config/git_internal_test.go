package config

import (
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestGitParseSimple(t *testing.T) {
	defaultURL, _ := url.Parse("http://banana.stand/repo")
	fileURL, _ := url.Parse("/dead/dove/donoteat")
	gitURL, _ := url.Parse("http://banana.stand/repo.git")
	tests := []struct {
		name string
		in   string
		item GitItem
	}{
		{"simple repo", defaultURL.String(), GitItem{URL: defaultURL, Bare: true}},
		{"file url", fileURL.String(), GitItem{URL: fileURL, Bare: true}},
		{"with .git", gitURL.String(), GitItem{URL: gitURL, Bare: true}},
		{"error - bad url", "\x7f", GitItem{}},
	}

	for _, test := range tests {
		expect := Git{Items: nil}
		if !strings.HasPrefix(test.name, "error -") {
			expect.Items = []GitItem{test.item}
		}
		t.Run(test.name, func(t *testing.T) {
			git := Git{}
			_ = git.parseSimple(test.in)
			if !cmp.Equal(expect, git) {
				t.Errorf("Unexpected result: %s", cmp.Diff(expect, git))
			}
		})
	}
}

func TestGitParseComplex(t *testing.T) {
	defaultBranch := plumbing.NewBranchReferenceName("dont-call-it-that")
	defaultTag := plumbing.NewTagReferenceName("no-touching")
	defaultURL, _ := url.Parse("http://banana.stand/repo.git")
	tests := []struct {
		name string
		in   map[interface{}]interface{}
		item GitItem
	}{
		{"repo only", map[interface{}]interface{}{"repo": "http://banana.stand/repo.git"}, GitItem{URL: defaultURL, Bare: true}},
		{"with bare", map[interface{}]interface{}{"repo": "http://banana.stand/repo.git", "bare": false}, GitItem{URL: defaultURL, Bare: false}},
		{"with tag", map[interface{}]interface{}{"repo": "http://banana.stand/repo.git", "tag": defaultTag.Short()}, GitItem{URL: defaultURL, Bare: true, Tag: defaultTag}},
		{"with branch", map[interface{}]interface{}{"repo": "http://banana.stand/repo.git", "branch": defaultBranch.Short()}, GitItem{URL: defaultURL, Bare: true, Branch: defaultBranch}},
		{"tag and branch", map[interface{}]interface{}{"repo": "http://banana.stand/repo.git", "tag": defaultTag.Short(), "branch": defaultBranch.Short()}, GitItem{URL: defaultURL, Bare: true, Branch: defaultBranch}},
		{"error - bad url", map[interface{}]interface{}{"repo": "\x7f"}, GitItem{}},
	}
	for _, test := range tests {
		expect := Git{Items: nil}
		if !strings.HasPrefix(test.name, "error -") {
			expect.Items = []GitItem{test.item}
		}
		t.Run(test.name, func(t *testing.T) {
			git := Git{}
			_ = git.parseComplex(test.in)
			if !cmp.Equal(expect, git) {
				t.Errorf("Unexpected result: %s", cmp.Diff(expect, git))
			}
		})
	}
}

func TestParseGit(t *testing.T) {
	defaultURL, _ := url.Parse("testrepo")

	tests := []struct {
		name   string
		in     tempConfig
		expect Git
	}{
		{"string", tempConfig{Git: []interface{}{defaultURL.String()}}, Git{Items: []GitItem{{URL: defaultURL, Bare: true}}}},
		{"map", tempConfig{Git: []interface{}{map[interface{}]interface{}{"repo": defaultURL.String(), "bare": false}}}, Git{Items: []GitItem{{URL: defaultURL, Bare: false}}}},
		{"error - bad config", tempConfig{Git: []interface{}{42}}, Git{Items: nil}},
		{"error - parsing error", tempConfig{Git: []interface{}{"\x7f"}}, Git{Items: nil}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			git := parseGit(test.in)
			if !cmp.Equal(test.expect, git) {
				t.Errorf("Unexpected result: %s", cmp.Diff(test.expect, git))
			}
		})
	}
}
