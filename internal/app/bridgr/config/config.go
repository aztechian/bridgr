package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// BridgrConf is the in-memory representation of the provided YAML config file
//
type BridgrConf struct {
	Yum      interface{}
	Files    Files
	Ruby     interface{}
	Python   interface{}
	Jenkins  interface{}
	Docker   interface{}
	Npm      interface{}
	Maven    interface{}
	Git      interface{}
	Settings interface{}
}

type tempConfig struct {
	Yum      interface{}
	Files    []interface{}
	Ruby     interface{}
	Python   interface{}
	Jenkins  interface{}
	Docker   interface{}
	Npm      interface{}
	Maven    []interface{}
	Git      []interface{}
	Settings []interface{}
}

// Helper interface translates top-level config file sections into normalized structs for use by workers
type Helper interface {
	parse(BridgrConf) (interface{}, error)
}

// Config is a factory method that instantiates and populates a BridgrConf object
func Config(f string) (BridgrConf, error) {
	var c BridgrConf

	if !fileExists(f) {
		return c, fmt.Errorf("config file %+s not found", f)
	}

	confData, err := ioutil.ReadFile(f)
	if err != nil {
		log.Println("Unable to read config file", f)
		return c, fmt.Errorf("unable to read config file %+s", f)
	}

	temp := tempConfig{}
	yaml.Unmarshal(confData, &temp)
	c.Files, _ = parseFiles(temp)
	return c, nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
