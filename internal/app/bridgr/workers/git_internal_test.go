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

type MockGitCredentailRW struct {
	isError bool
	user    string
	pass    string
	bytes.Buffer
	CredentialReaderWriter
}

func (mgcrw *MockGitCredentailRW) Write(c Credential) error {
	// up := append([]byte(c.Username + c.Password))
	mgcrw.Buffer.Write([]byte(c.Username + c.Password))
	return nil
}

func (mgcrw *MockGitCredentailRW) Read(url *url.URL) (Credential, bool) {
	if mgcrw.isError {
		return Credential{}, false
	}
	return Credential{Username: mgcrw.user, Password: mgcrw.pass}, true
}

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

func TestGitAuth(t *testing.T) {
	dummyURL, _ := url.Parse("nothing")
	tests := []struct {
		name    string
		envUser string
		envPass string
		found   bool
		expect  string
	}{
		{"user + pass found", "buster", "monster!", true, "bustermonster!"},
		{"user + pass not found", "buster", "monster!", false, ""},
		{"only token", "", "her?", true, "her?"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			creds := MockGitCredentailRW{isError: !test.found, user: test.envUser, pass: test.envPass}
			gitAuth(dummyURL, &creds)
			if !cmp.Equal(creds.Buffer.String(), test.expect) {
				t.Error(cmp.Diff(creds.Buffer.String(), test.expect))
			}
		})
	}
}

func TestGitCredentialWrite(t *testing.T) {
	tests := []struct {
		name  string
		creds Credential
	}{
		{"user and password", Credential{Username: "buster", Password: "monster!"}},
		{"empty", Credential{}},
		{"only password", Credential{Password: "monster?"}},
		{"only user", Credential{Username: "buster"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gitCreds := gitCredentials{}
			gitCreds.Write(test.creds)
			expectedUser := test.creds.Username
			if expectedUser == "" {
				expectedUser = "git"
			}
			if !cmp.Equal(gitCreds.Username, expectedUser) || !cmp.Equal(gitCreds.Password, test.creds.Password) {
				t.Error("Mismatch credential username/password")
				t.Errorf("gitCredential: %v   credential: %v", gitCreds, test.creds)
			}
		})
	}
}
