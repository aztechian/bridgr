package config

import (
	"path"

	"github.com/docker/distribution/reference"
)

const defaultYumImage = "centos:7"

// Yum is the normalized structure for workers to get YUM information from the config file
type Yum struct {
	Repos []string
	Items []string
	Image reference.Named
}

// BaseDir is the top-level directory name for all objects written out under the Yum worker
func (y *Yum) BaseDir() string {
	return path.Join(BaseDir(), "yum")
}

// func parseYum(config tempConfig) Yum {
// 	namedImg, _ := reference.ParseNormalizedNamed(defaultYumImage)
// 	yum := Yum{
// 		Image: namedImg,
// 	}
// 	switch c := config.Yum.(type) {
// 	case []interface{}:
// 		_ = yum.parsePackages(c)
// 	case map[interface{}]interface{}:
// 		packages := c["packages"]
// 		repos := c["repos"]
// 		if repos == nil {
// 			repos = []interface{}{}
// 		}
// 		if _, present := c["image"]; present {
// 			customImg, err := reference.ParseNormalizedNamed(c["image"].(string))
// 			if err != nil {
// 				yum.Image = namedImg
// 				log.Printf("Invalid image given for YUM: %s, defaulting to %s", err, namedImg.Name())
// 			} else {
// 				yum.Image = customImg
// 			}
// 		}
// 		_ = yum.parseRepos(repos.([]interface{}))
// 		_ = yum.parsePackages(packages.([]interface{}))
// 	case nil:

// 	default:
// 		fmt.Printf("DEBUG: Unknown configuration section for Yum: %+s", c)
// 	}
// 	return yum
// }

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
