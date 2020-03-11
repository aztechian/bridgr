package config

import (
	"net/url"
	"path"
	"reflect"

	"github.com/docker/distribution/reference"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

// Git is the struct for holding a Git configuration in Bridgr
type Git []GitItem

// GitItem is the sub-struct of items in a Git struct
type GitItem struct {
	URL    *url.URL
	Bare   bool
	Branch plumbing.ReferenceName
	Tag    plumbing.ReferenceName
}

func (gi GitItem) String() string {
	return gi.URL.String()
}

func (g Git) Count() int {
	return len(g)
}

func (g Git) Image() reference.Named {
	return nil
}

// NewGitItem creates a new, default GitItem struct
func NewGitItem(repo string) GitItem {
	var u *url.URL = nil
	if len(repo) > 0 {
		u, _ = url.Parse(repo)
	}
	return GitItem{URL: u, Bare: true}
}

// BaseDir is the top-level directory name for all objects written out under the Python worker
func (g *Git) BaseDir() string {
	return path.Join(BaseDir(), "git")
}

func (gi *GitItem) parseComplex(pkg map[string]interface{}) error {
	url, err := url.Parse(pkg["repo"].(string))
	if err != nil {
		return err
	}

	gi.URL = url
	if bare, present := pkg["bare"]; present {
		gi.Bare = bare.(bool)
	}
	if branch, present := pkg["branch"]; present {
		gi.Branch = plumbing.NewBranchReferenceName(branch.(string))
	}
	// branch name wins if both are present in the config
	if tag, present := pkg["tag"]; present && gi.Branch == "" {
		gi.Tag = plumbing.NewTagReferenceName(tag.(string))
	}
	return nil
}

func stringToGitItem(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() == reflect.String && t == reflect.TypeOf(GitItem{}) {
		return NewGitItem(data.(string)), nil
	}
	return data, nil
}

func mapToGitItem(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() == reflect.Map && t == reflect.TypeOf(GitItem{}) {
		item := NewGitItem("")
		_ = item.parseComplex(data.(map[string]interface{}))
		return item, nil
	}
	return data, nil
}
