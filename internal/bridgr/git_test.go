package bridgr_test

import (
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/google/go-cmp/cmp"
)

func TestGitImage(t *testing.T) {
	git := bridgr.Git{}
	if git.Image() != nil {
		t.Errorf("expected nil, but got %+v", git.Image())
	}
}

func TestGitName(t *testing.T) {
	expected := "git"
	git := bridgr.Git{}
	if !cmp.Equal(expected, git.Name()) {
		t.Error(cmp.Diff(expected, git.Name()))
	}
}

func TestGitItemString(t *testing.T) {
	expect := "git://github.com/repo.git"
	src, _ := url.Parse(expect)
	git := bridgr.GitItem{URL: src}
	if !cmp.Equal(git.String(), expect) {
		t.Error(cmp.Diff(git.String(), expect))
	}
}

func TestGitHook(t *testing.T) {
	git := bridgr.Git{}
	result := reflect.TypeOf(git.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(reflect.Func, result.Name()))
	}
}

func TestGetNew(t *testing.T) {
	src, _ := url.Parse("https://github.com/repo.git")

	g := bridgr.NewGitItem(src.String())
	expect := bridgr.GitItem{URL: src, Bare: true}
	if !cmp.Equal(expect, g) {
		t.Error(cmp.Diff(expect, g))
	}

	g2 := bridgr.NewGitItem("")
	expect2 := bridgr.GitItem{URL: nil, Bare: true}
	if !cmp.Equal(expect2, g2) {
		t.Error(cmp.Diff(expect2, g2))
	}
}
