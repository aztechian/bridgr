package config

import (
	"fmt"
	"path"
	"reflect"

	"github.com/docker/distribution/reference"
)

var defaultRbImg reference.Named

const defaultRbSrc = "https://rubygems.org"
const baseRbImage = "ruby"

func init() {
	defaultRbImg, _ = reference.ParseNormalizedNamed(baseRbImage + ":2-alpine")
}

// Ruby struct is the configuration object specifically for the Ruby section of the config file
type Ruby struct {
	Gems    []rubyItem
	Version rubyVersion
	Sources []string
}

type rubyVersion reference.Named

func NewRuby() *Ruby {
	return &Ruby{
		Version: defaultRbImg,
		Sources: []string{defaultRbSrc},
	}
}

// RubyItem is a struct to hold a ruby gem specification
type rubyItem struct {
	Package string
	Version string
}

func (ri rubyItem) String() string {
	if ri.Version != "" {
		return fmt.Sprintf("%s, %s", ri.Package, ri.Version)
	}
	return ri.Package
}

// BaseDir is the top-level directory name for all objects written out under the Python worker
func (r *Ruby) BaseDir() string {
	return path.Join(BaseDir(), "ruby")
}

func (r *Ruby) Count() int {
	return len(r.Gems)
}

func (r *Ruby) Image() reference.Named {
	if r.Version == nil {
		return defaultRbImg
	}
	return r.Version
}

func stringToRuby(t reflect.Type, f reflect.Type, data interface{}) (interface{}, error) {
	if f == reflect.TypeOf(rubyItem{}) && t.Kind() == reflect.String {
		return rubyItem{
			Package: data.(string),
		}, nil
	}
	return data, nil
}

func versionToRubyImage(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf((*rubyVersion)(nil)).Elem() {
		return data, nil
	}
	return reference.ParseAnyReference(baseRbImage + ":" + data.(string))
}

func arrayToRuby(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || t != reflect.TypeOf(Ruby{}) {
		return data, nil
	}
	var gemList []rubyItem
	for _, g := range data.([]interface{}) {
		gemList = append(gemList, rubyItem{Package: g.(string)})
	}

	return &Ruby{
		Version: defaultRbImg,
		Gems:    gemList,
		Sources: []string{defaultRbSrc},
	}, nil
}
