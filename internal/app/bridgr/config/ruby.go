package config

import (
	"bridgr/internal/app/bridgr"
	"path"

	"github.com/docker/distribution/reference"
)

// Ruby struct is the configuration object specifically for the Ruby section of the config file
type Ruby struct {
	Items   []RubyItem
	Sources []string
	Image   reference.Named
}

// RubyItem is a struct to hold a ruby gem specification
type RubyItem struct {
	Package string
	Version string
}

var defaultRbImg, _ = reference.ParseNormalizedNamed("ruby:2-alpine")
var defaultRbSrc = "https://rubygems.org"

// BaseDir is the top-level directory name for all objects written out under the Python worker
func (r *Ruby) BaseDir() string {
	return path.Join(BaseDir(), "ruby")
}

func parseRuby(config tempConfig) Ruby {
	rb := Ruby{Image: defaultRbImg}
	switch c := config.Ruby.(type) {
	case []interface{}:
		_ = rb.parsePackages(c)
	case map[interface{}]interface{}:
		if _, present := c["version"]; present {
			var err error
			rb.Image, err = reference.ParseNormalizedNamed("ruby:" + c["version"].(string))
			if err != nil {
				bridgr.Debugf("Error using Ruby image of 'ruby:%s', falling back to %s", c["version"].(string), defaultRbImg.String())
				rb.Image = defaultRbImg
			}
		}
		if sources, present := c["sources"]; present {
			rb.addSources(sources.([]interface{}))
		}
		pkgList := c["gems"].([]interface{})
		_ = rb.parsePackages(pkgList)
	default:
		bridgr.Debugf("Unknown configuration section for Ruby: %+s", c)
	}
	bridgr.Debugf("Final Ruby configuration %+v", rb)
	return rb
}

func (r *Ruby) parsePackages(pkgList []interface{}) error {
	for _, pkg := range pkgList {
		switch pkgObj := pkg.(type) {
		case string:
			r.Items = append(r.Items, RubyItem{Package: pkgObj})
		case map[interface{}]interface{}:
			item := RubyItem{
				Package: pkgObj["package"].(string),
				Version: pkgObj["version"].(string),
			}
			r.Items = append(r.Items, item)
		}
	}
	return nil
}

func (r *Ruby) addSources(srcList []interface{}) error {
	for _, src := range srcList {
		r.Sources = append(r.Sources, src.(string))
	}
	return nil
}
