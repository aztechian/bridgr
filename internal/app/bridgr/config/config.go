package config

import (
	"github.com/davecgh/go-spew/spew"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v3"
	"github.com/mitchellh/mapstructure"
)

// BridgrConf is the in-memory representation of the provided YAML config file
//
type BridgrConf struct {
	Yum      Yum
	Files    Files
	Ruby     Ruby
	Python   Python
	Jenkins  interface{}
	Docker   Docker
	Npm      interface{}
	Maven    interface{}
	Git      Git
	Settings interface{}
}

// type tempConfig struct {
// 	Yum      interface{}
// 	Files    []interface{}
// 	Ruby     interface{}
// 	Python   interface{}
// 	Jenkins  interface{}
// 	Docker   interface{}
// 	Npm      interface{}
// 	Maven    []interface{}
// 	Git      []interface{}
// 	Settings []interface{}
// }

// Helper interface translates top-level config file sections into normalized structs for use by workers
// type Helper interface {
// 	parse(BridgrConf) (interface{}, error)
// }

// New is a factory method that instantiates and populates a BridgrConf object
func New(f io.ReadCloser) (*BridgrConf, error) {
	var c BridgrConf
	confData, err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		log.Printf("Unable to read config file: %s", err)
		return &c, err
	}

	var temp = make([]interface{}, 1)
	// temp := tempConfig{}
	err = yaml.Unmarshal(confData, &temp)
	if err != nil {
		return &c, err
	}

	mapstructure.Decode(temp, &c)
	spew.Dump(c)
	// c.Files = parseFiles(temp)
	// c.Yum = parseYum(temp)
	// c.Docker = parseDocker(temp)
	// c.Python = parsePython(temp)
	// c.Ruby = parseRuby(temp)
	// c.Git = parseGit(temp)
	return &c, nil
}

// BaseDir gives the runtime absolute directory of the base "packages" directory
// See the individual repo type struct for the type-specific path
func BaseDir() string {
	var cwd, _ = os.Getwd()
	return path.Join(cwd, "packages")
}
