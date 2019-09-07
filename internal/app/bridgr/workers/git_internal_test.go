package workers

import (
	"bridgr/internal/app/bridgr/config"
	"bytes"
	"net/url"
	"path"
	"strings"
	"testing"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"

	"github.com/google/go-cmp/cmp"
)

func TestGitPrepDir(t *testing.T) {
	conf := config.Git{}
	baseDir := conf.BaseDir()
	tests := []struct {
		name   string
		in     string
		expect string
	}{
		{"dotgit", "https://corp.server/repo.git", path.Join(baseDir, "repo")},
		{"plain github", "https://github.com/aztechian/bridgr", path.Join(baseDir, "bridgr")},
		{"local file", "/path/to/something.git", path.Join(baseDir, "something")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			git := Git{}
			url, _ := url.Parse(test.in)
			dir := git.prepDir(url)
			if !cmp.Equal(dir, test.expect) {
				t.Errorf("Expected result from prepDir() to be %s, but got %s", dir, test.expect)
			}
		})
	}
}

func TestGitGenerateRefInfo(t *testing.T) {
	mem := memory.NewStorage()
	repo, _ := git.Init(mem, nil)
	head, _ := repo.Reference(plumbing.HEAD, false)
	ref := plumbing.NewHashReference("refs/heads/testing", head.Hash())
	_ = repo.Storer.SetReference(ref)
	buff := bytes.Buffer{}
	generateRefInfo(repo, &buff)
	if !strings.Contains(buff.String(), "testing") {
		t.Errorf("Expected refs file to contain 'testing' but got %s", buff.String())
	}
}

func TestGitGeneratePackInfo(t *testing.T) {
	mem := memory.NewStorage()
	repo, _ := git.Init(mem, nil)
	head, _ := repo.Reference(plumbing.HEAD, false)
	ref := plumbing.NewHashReference("refs/heads/testing", head.Hash())
	_ = repo.Storer.SetReference(ref)
	buff := bytes.Buffer{}
	t.Log(repo.References())
	generatePackInfo(repo, &buff)
	// I don't know how to create a packfile for this, so... we don't care what our function returns
}
