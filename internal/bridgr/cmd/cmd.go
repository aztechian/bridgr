package cmd

import (
	"io"
	"io/ioutil"
	"net/url"
	"reflect"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/davecgh/go-spew/spew"
	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v1"
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
	err = yaml.Unmarshal(confData, &temp)
	if err != nil {
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
		}
		decode(section, cfg)
		c = append(c, section)
	}
	bridgr.Debug(spew.Sdump(c))
	return &c, nil
}

// Execute runs the specified workers from the configuration
func Execute(config Bridgr, filter []string) error {
	// workers := filterWorkers(workers, filter)
	for _, w := range config {
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

func decode(p bridgr.Configuration, configSection interface{}) error {
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(debugHook, stringToImage, stringToURL, p.Hook()),
		WeaklyTypedInput: true,
		Result:           p,
		ZeroFields:       true,
	})
	err := decoder.Decode(configSection)
	return err
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
