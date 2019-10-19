package workers

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/config"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"strings"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/storer"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

// Git is a struct that implements a Worker interface for fetching Git artifacts
type Git struct {
	Config *config.Git
}

type gitCredentials struct {
	*http.BasicAuth
	workerCredentialReader
}

// NewGit creates a new Git worker struct
func NewGit(conf *config.BridgrConf) Worker {
	return &Git{Config: &conf.Git}
}

// Name returns the friendly name of the Git struct
func (g *Git) Name() string {
	return "Git"
}

// Setup does any initial setup for the Git worker
func (g *Git) Setup() error {
	return nil
}

// Run executes the Git worker to fetch artifacts
func (g *Git) Run() error {
	err := g.Setup()
	if err != nil {
		return err
	}
	for _, item := range g.Config.Items {
		dir := g.prepDir(item.URL)
		repo, err := gitClone(item, dir)
		if err != nil {
			bridgr.Printf("Error cloning Git repository '%s': %s", item.URL.String(), err)
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
	dir = path.Join(g.Config.BaseDir(), dir)

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		bridgr.Debugf("%s exists, removing to allow new clone", dir)
		os.RemoveAll(dir)
	}
	return dir
}

func gitClone(item config.GitItem, dir string) (*git.Repository, error) {
	bridgr.Debugf("About to clone %s into %s", item.URL.String(), dir)
	creds := &gitCredentials{}
	gitAuth(item.URL, creds)
	opts := git.CloneOptions{URL: item.URL.String(), SingleBranch: false, Auth: creds}
	if item.Tag != "" {
		bridgr.Debugf("Getting specific tag %s", item.Tag.String())
		opts.ReferenceName = item.Tag
		opts.SingleBranch = true
	}
	if item.Branch != "" {
		bridgr.Debugf("Getting specific branch %s", item.Branch.String())
		opts.ReferenceName = item.Branch
		opts.SingleBranch = true
	}
	// TODO: PlainClone() is nice and simple, but we need to be able to pass in a filesystem for testing (ie, memory)
	return git.PlainClone(dir, item.Bare, &opts)
}

func gitAuth(url *url.URL, rw CredentialReaderWriter) {
	if creds, ok := rw.Read(url); ok {
		bridgr.Debugf("Git: Found credentials for %s", url.String())
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
