package bridgr

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"text/template"

	"github.com/aztechian/bridgr/internal/bridgr/asset"
	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
)

var (
	rbImage reference.Named
	rbGems  *template.Template
)

const defaultRbSource = "https://rubygems.org"

func init() {
	rbImage, _ = reference.ParseNormalizedNamed(baseImage["ruby"] + ":2-alpine")
	rbGems = asset.Template("Gemfile")
}

// Ruby struct is the configuration object specifically for the Ruby section of the config file
type Ruby struct {
	Gems    []rubyItem
	Version rubyVersion
	Sources []string
}

// RubyItem is a struct to hold a ruby gem specification
type rubyItem struct {
	Package string
	Version string
}

type rubyVersion reference.Named

func (ri rubyItem) String() string {
	if ri.Version != "" {
		return fmt.Sprintf("%s, %s", ri.Package, ri.Version)
	}
	return ri.Package
}

// BaseDir is the top-level directory name for all objects written out under the Python worker
func (r Ruby) dir() string {
	return BaseDir(r.Name())
}

// Image implements the Imager interface
func (r *Ruby) Image() reference.Named {
	if r.Version == nil {
		return rbImage
	}
	return r.Version
}

// Name returns the name of this Configuration
func (r Ruby) Name() string {
	return "ruby"
}

func stringToRuby(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t == reflect.TypeOf(rubyItem{}) && f.Kind() == reflect.String {
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
	return reference.ParseAnyReference(baseImage["ruby"] + ":" + data.(string))
}

func arrayToRuby(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || t != reflect.TypeOf(Ruby{}) {
		return data, nil
	}
	var gemList []rubyItem
	for _, g := range data.([]interface{}) {
		if pkg, ok := g.(string); ok {
			gemList = append(gemList, rubyItem{Package: pkg})
		}
	}

	return Ruby{
		Version: rbImage,
		Sources: []string{defaultRbSource},
		Gems:    gemList,
	}, nil
}

// Hook implements the Parser interface, returns a function for use by mapstructure when parsing config files
func (r Ruby) Hook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		stringToRuby,
		versionToRubyImage,
		arrayToRuby,
	)
}

// Setup creates the items that are needed to fetch artifacts for the Python worker. It does not actually fetch artifacts.
func (r *Ruby) Setup() error {
	Debug("Called Ruby.Setup()")
	_ = os.MkdirAll(r.dir(), os.ModePerm)

	gemfile, err := os.Create(path.Join(r.dir(), "Gemfile"))
	if err != nil {
		return fmt.Errorf("Unable to create Ruby Gemfile: %s", err)
	}

	return asset.RenderFile(rbGems, r, gemfile)
}

// Run fetches all artifacts for the Python configuration
func (r *Ruby) Run() error {
	Debug("Called Ruby.Setup()")
	if err := r.Setup(); err != nil {
		return err
	}

	shell, err := asset.Load("ruby.sh") //no parsing needed, so just Load is fine here
	if err != nil {
		return err
	}

	batcher := newBatch(r.Image().Name(), r.dir(), path.Join(r.dir(), "Gemfile"), "/Gemfile")
	return batcher.runContainer("bridgr_ruby", shell)
}
