package bridgr

import (
	"bytes"
	"net/url"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/storage/memory"
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
	baseDir := BaseDir("git")
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
				t.Error(cmp.Diff(dir, test.expect))
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

func TestGitClone(t *testing.T) {
	src, _ := url.Parse("https://git.bluth/michael.git")
	tests := []struct {
		name string
		item GitItem
	}{
		{"basic", GitItem{URL: src, Bare: true}},
		{"tagged", GitItem{URL: src, Bare: true, Tag: "v1"}},
		{"branch", GitItem{URL: src, Bare: true, Branch: "president"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.item.clone("test")
		})
	}
}

func TestMapToGitItem(t *testing.T) {
	source, _ := url.Parse("https://motherboy.com/results.git")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid type", reflect.TypeOf(""), "", false, ""},
		{"valid", reflect.TypeOf(&GitItem{}).Elem(), map[interface{}]interface{}{"repo": source.String(), "bare": false}, false, GitItem{URL: source, Bare: false}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := mapToGitItem(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(result, test.expect, namedComparer) && !test.isError {
				t.Error(cmp.Diff(result, test.expect))
			}
		})
	}
}

func TestStringToGitItem(t *testing.T) {
	source, _ := url.Parse("https://motherboy.com/results.git")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid type", reflect.TypeOf(42), 42, false, 42},
		{"valid", reflect.TypeOf(&GitItem{}).Elem(), source.String(), false, GitItem{URL: source, Bare: true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := stringToGitItem(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(result, test.expect, namedComparer) && !test.isError {
				t.Error(cmp.Diff(result, test.expect))
			}
		})
	}
}

func TestGitParseComplex(t *testing.T) {
	src, _ := url.Parse("https://motherboy.com/results.git")
	tests := []struct {
		name   string
		input  map[interface{}]interface{}
		expect *GitItem
	}{
		{"only repo", map[interface{}]interface{}{"repo": src.String()}, &GitItem{URL: src}},
		{"missing repo", map[interface{}]interface{}{"repo": "\007forget-me-now"}, nil},
		{"has bare", map[interface{}]interface{}{"repo": src.String(), "bare": false}, &GitItem{URL: src, Bare: false}},
		{"has tag", map[interface{}]interface{}{"repo": src.String(), "tag": "cornballer"}, &GitItem{URL: src, Tag: "refs/tags/cornballer"}},
		{"has branch", map[interface{}]interface{}{"repo": src.String(), "branch": "thething"}, &GitItem{URL: src, Branch: "refs/heads/thething"}},
		{"has tag and branch", map[interface{}]interface{}{"repo": src.String(), "tag": "cornballer", "branch": "thething"}, &GitItem{URL: src, Branch: "refs/heads/thething"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := GitItem{}
			err := result.parseComplex(test.input)
			if test.expect == nil && err == nil {
				t.Errorf("expected error but got %s", err)
			}
			if test.expect == nil {
				return
			}
			if !cmp.Equal(*test.expect, result) {
				t.Error(cmp.Diff(*test.expect, result))
			}
		})
	}
}
