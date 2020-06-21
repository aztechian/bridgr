package cmd

import (
	"io"
	"io/ioutil"
	"net/url"
	"reflect"
	"strings"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/davecgh/go-spew/spew"
	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v3"
)

// Bridgr needs documentation
type Bridgr []bridgr.Configuration

// New is a factory method that instantiates and populates a BridgrConf object
func New(f io.ReadCloser) (*Bridgr, error) {
	c := Bridgr{}
	confData, err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		bridgr.Printf("Unable to read config file: %s", err)
		return &c, err
	}

	var temp map[string]interface{}
	if err = yaml.Unmarshal(confData, &temp); err != nil {
		return &c, err
	}

	for key, cfg := range temp {
		var section bridgr.Configuration
		switch key {
		case "yum":
			section = &bridgr.Yum{}
		case "docker":
			section = &bridgr.Docker{}
		case "files":
			section = &bridgr.File{}
		case "ruby":
			section = &bridgr.Ruby{}
		case "python":
			section = &bridgr.Python{}
		case "git":
			section = &bridgr.Git{}
		default:
			bridgr.Printf("Unable to create repository \"%s\", skipping.", key)
			continue
		}
		err = decode(section, cfg)
		if err != nil {
			bridgr.Printf("error decoding section \"%s\": %s", key, err)
		}
		c = append(c, section)
	}
	bridgr.Debug(spew.Sdump(c))
	return &c, nil
}

// Execute runs the specified workers from the configuration
func Execute(config Bridgr, filter []string) error {
	for _, w := range config {
		if len(filter) > 0 && !contains(w.Name(), filter) {
			bridgr.Debugf("skipping worker %s, not in %s", w.Name(), filter)
			continue
		}
		bridgr.Printf("Processing %s...", w.Name())
		var err error
		if bridgr.DryRun {
			err = w.Setup()
		} else {
			err = w.Run()
		}
		if err != nil {
			bridgr.Printf("Error processing %s: %s", w.Name(), err)
		}
	}

	return nil
}

func contains(item string, list []string) bool {
	if len(list) <= 0 || strings.ToLower(list[0]) == "all" {
		return true
	}
	for _, x := range list {
		if item == x {
			return true
		}
	}
	return false
}

func decode(p bridgr.Configuration, configSection interface{}) error {
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(debugHook, stringToImage, stringToURL, p.Hook()),
		WeaklyTypedInput: true,
		Result:           p,
		ZeroFields:       true,
	})
	return decoder.Decode(configSection)
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
