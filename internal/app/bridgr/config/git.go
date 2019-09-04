package config

import (
	"bridgr/internal/app/bridgr"
	"net/url"
	"path"
)

// Git is the struct for holding a Git configuration in Bridgr
type Git struct {
	Items []GitItem
}

// GitItem is the sub-struct of items in a Git struct
type GitItem struct {
	URL  *url.URL
	Bare bool
	Tree string
}

// BaseDir is the top-level directory name for all objects written out under the Python worker
func (g *Git) BaseDir() string {
	return path.Join(BaseDir(), "git")
}

func parseGit(config tempConfig) Git {
	g := Git{}
	for _, entry := range config.Git {
		var err error
		switch repoObj := entry.(type) {
		case string: //simple string entry
			err = g.parseSimple(repoObj)
		case map[interface{}]interface{}: // complex type
			err = g.parseComplex(repoObj)
		default:
			bridgr.Debugf("Unknown configuration section for Git: %+s", repoObj)
		}
		if err != nil {
			bridgr.Debug(err)
		}
	}

	bridgr.Debugf("Final Git configuration %+v", g)
	return g
}

func (g *Git) parseComplex(pkg map[interface{}]interface{}) error {
	url, err := url.Parse(pkg["repo"].(string))
	if err != nil {
		return err
	}
	item := GitItem{URL: url, Bare: true, Tree: "master"}
	if bare, present := pkg["bare"]; present {
		item.Bare = bare.(bool)
	}
	if tree, present := pkg["tree"]; present {
		item.Tree = tree.(string)
	}
	g.Items = append(g.Items, item)
	return nil
}

func (g *Git) parseSimple(pkg string) error {
	url, err := url.Parse(pkg)
	if err != nil {
		return err
	}
	g.Items = append(g.Items, GitItem{URL: url, Bare: true, Tree: "master"})
	return nil
}
