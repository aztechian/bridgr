package config

import (
	"bridgr/internal/app/bridgr"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/docker/distribution/reference"

	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

var decodeHooks = mapstructure.ComposeDecodeHookFunc(
	debugHook,
	versionToRubyImage,
	versionToYumImage,
	versionToPythonImage,
	stringToImage,
	stringToURL,
	stringToGitItem,
	stringToFileItem,
	stringToRuby,
	arrayToRuby,
	arrayToPython,
	arrayToYum,
	mapToImage,
	mapToGitItem,
)

// BridgrConf is the in-memory representation of the provided YAML config file
//
type BridgrConf struct {
	Yum      *Yum
	Files    *Files
	Ruby     *Ruby
	Python   *Python
	Jenkins  *Jenkins
	Docker   *Docker
	Npm      *Npm
	Maven    *Maven
	Git      *Git
	Settings *Settings
}

// place holders for types until they get implemented

type Jenkins interface{}
type Npm interface{}
type Maven interface{}
type Settings interface{}

type Imager interface {
	Image() reference.Named
}

// New is a factory method that instantiates and populates a BridgrConf object
func New(f io.ReadCloser) (*BridgrConf, error) {
	c := BridgrConf{}
	confData, err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		log.Printf("Unable to read config file: %s", err)
		return &c, err
	}

	var temp map[string]interface{}
	err = yaml.Unmarshal(confData, &temp)
	if err != nil {
		return &c, err
	}

	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       decodeHooks,
		WeaklyTypedInput: true,
		Result:           &c,
		ZeroFields:       true,
	})
	decoder.Decode(temp)
	spew.Dump(c)
	return &c, nil
}

// BaseDir gives the runtime absolute directory of the base "packages" directory
// See the individual repo type struct for the type-specific path
func BaseDir() string {
	var cwd, _ = os.Getwd()
	return path.Join(cwd, "packages")
}

func stringToImage(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf((*reference.Reference)(nil)).Elem() {
		return data, nil
	}
	return reference.ParseNormalizedNamed(data.(string))
}

func stringToURL(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(&url.URL{}) {
		return data, nil
	}
	return url.Parse(data.(string))
}

func debugHook(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	bridgr.Debugf("%s -> %s\n%s\n\n", f, t, data)
	return data, nil
}
