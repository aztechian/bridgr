package config

import (
	"fmt"
)

// Yum is the normalized structure for workers to get YUM information from the config file
type Yum struct {
	Repos []string
	Items []string
}

// BaseDir is the top-level directory name for all objects written out under the Yum worker
func (y *Yum) BaseDir() string {
	return "yum"
}

func parseYum(config tempConfig) Yum {
	yum := Yum{}
	switch c := config.Yum.(type) {
	case []interface{}:
		yum.parsePackages(c)
	case map[interface{}]interface{}:
		repos := c["repos"]
		packages := c["packages"]
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
