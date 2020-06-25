package bridgr

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
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

type gitCredentials struct {
	http.BasicAuth
	WorkerCredentialReader
}

func (gi GitItem) String() string {
	return gi.URL.String()
}

// Image implements the Imager interface
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

// dir is the top-level directory name for all objects written out under the Git worker
func (g Git) dir() string {
	return BaseDir(g.Name())
}

// Name returns the name of this Configuration
func (g Git) Name() string {
	return "git"
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

// Hook implements the Parser interface, returns a function for use by mapstructure when parsing config files
func (g *Git) Hook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		stringToGitItem,
		mapToGitItem,
	)
}

// Setup does any initial setup for the Git worker
func (g *Git) Setup() error {
	return os.MkdirAll(g.dir(), os.ModePerm)
}

// Run executes the Git worker to fetch artifacts
func (g *Git) Run() error {
	err := g.Setup()
	if err != nil {
		return err
	}
	for _, item := range *g {
		dir := g.prepDir(item.URL)
		repo, err := item.clone(dir)
		if err != nil {
			Printf("Error cloning Git repository '%s': %s", item.URL.String(), err)
		}
		if item.Bare {
			_ = os.MkdirAll(path.Join(dir, "info"), os.ModePerm)
			_ = os.MkdirAll(path.Join(dir, "objects", "info"), os.ModePerm)
			infoRefs, _ := os.Create(path.Join(dir, "info", "refs"))
			generateRefInfo(repo, infoRefs)
			infoRefs.Close()

			infoPack, _ := os.Create(path.Join(dir, "objects", "info", "packs"))
			generatePackInfo(repo, infoPack)
			infoPack.Close()
		}
	}

	return nil
}

func (g *Git) prepDir(url *url.URL) string {
	dir := path.Base(url.Path)
	dir = strings.TrimSuffix(dir, git.GitDirName)
	dir = path.Join(g.dir(), dir)

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		Debugf("%s exists, removing to allow new clone", dir)
		os.RemoveAll(dir)
	}
	return dir
}

func (gi GitItem) clone(dir string) (*git.Repository, error) {
	Debugf("About to clone %s into %s", gi.URL.String(), dir)
	creds := gitCredentials{}
	gitAuth(gi.URL, &creds)
	opts := git.CloneOptions{URL: gi.URL.String(), SingleBranch: false, Auth: &creds.BasicAuth}
	if gi.Tag != "" {
		Debugf("Getting specific tag %s", gi.Tag.String())
		opts.ReferenceName = gi.Tag
		opts.SingleBranch = true
	}
	if gi.Branch != "" {
		Debugf("Getting specific branch %s", gi.Branch.String())
		opts.ReferenceName = gi.Branch
		opts.SingleBranch = true
	}
	// TODO: PlainClone() is nice and simple, but we need to be able to pass in a filesystem for testing (ie, memory)
	return git.PlainClone(dir, gi.Bare, &opts)
}

func gitAuth(url *url.URL, rw CredentialReaderWriter) {
	if creds, ok := rw.Read(url); ok {
		Debugf("Git: Found credentials for %s", url.String())
		_ = rw.Write(creds)
	}
}

func generateRefInfo(repo *git.Repository, file io.Writer) {
	refs, _ := repo.References()
	_ = refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Type() == plumbing.HashReference {
			line := fmt.Sprintf("%s\t%s\n", ref.Strings()[1], ref.Strings()[0])
			_, _ = file.Write([]byte(line))
		}
		return nil
	})
}

func generatePackInfo(repo *git.Repository, file io.Writer) {
	pos, ok := repo.Storer.(storer.PackedObjectStorer)
	if !ok {
		// whatever format this repo is in doesn't support Packed objects
		return
	}
	// Get the existing object packs.
	hs, err := pos.ObjectPacks()
	if err != nil {
		return
	}
	// write out the info content to file
	for _, pack := range hs {
		line := fmt.Sprintf("P pack-%s.pack\n", pack.String())
		_, _ = file.Write([]byte(line))
	}
}

func (c *gitCredentials) Write(creds Credential) error {
	c.Username = creds.Username
	if c.Username == "" {
		c.Username = "git" // for token auth, the username must be anything _but_ blank
	}
	c.Password = creds.Password
	return nil
}
