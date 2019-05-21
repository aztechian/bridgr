package config

import (
	"fmt"
	"path"
)

const defaultYumImage = "library/centos:7" // TODO move this to a function that returns the canonical form of an image

// Yum is the normalized structure for workers to get YUM information from the config file
type Yum struct {
	Repos []string
	Items []string
	Image string
}

// BaseDir is the top-level directory name for all objects written out under the Yum worker
func (y *Yum) BaseDir() string {
	return path.Join(BaseDir(), "yum")
}

func parseYum(config tempConfig) Yum {
	yum := Yum{
		Image: defaultYumImage,
	}
	switch c := config.Yum.(type) {
	case []interface{}:
		yum.parsePackages(c)
	case map[interface{}]interface{}:
		repos := c["repos"]
		packages := c["packages"]
		yum.Image = defaultYumImage
		if _, present := c["image"]; present {
			yum.Image = c["image"].(string)
		}
		yum.parseRepos(repos.([]interface{}))
		yum.parsePackages(packages.([]interface{}))
	default:
		fmt.Printf("DEBUG: Unknown configuration section for Yum: %+s", c)
	}
	return yum
}

func (y *Yum) parseRepos(repolist []interface{}) error {
	for _, repo := range repolist {
		s, ok := repo.(string)
		if ok {
			y.Repos = append(y.Repos, s)
		}
	}
	return nil
}

func (y *Yum) parsePackages(packagelist []interface{}) error {
	for _, pkg := range packagelist {
		s, ok := pkg.(string)
		if ok {
			y.Items = append(y.Items, s)
		}
	}
	return nil
}
