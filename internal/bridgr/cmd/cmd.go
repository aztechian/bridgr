package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"reflect"
	"strings"
	"syscall"
	"time"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/briandowns/spinner"
	"github.com/davecgh/go-spew/spew"
	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
	"golang.org/x/crypto/ssh/terminal"
	"gopkg.in/yaml.v3"
	log "unknwon.dev/clog/v2"
)

const spinnerSpeed = 400 * time.Millisecond

// Bridgr needs documentation
type Bridgr []bridgr.Configuration

// New is a factory method that instantiates and populates a BridgrConf object
func New(f io.ReadCloser) (*Bridgr, error) {
	c := Bridgr{}
	confData, err := ioutil.ReadAll(f)
	defer f.Close()
	if err != nil {
		log.Error("Unable to read config file: %s", err)
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
		case "helm":
			section = &bridgr.Helm{}
		default:
			log.Warn("Unable to create repository \"%s\", skipping.", key)
			continue
		}
		err = decode(section, cfg)
		if err != nil {
			log.Warn("error decoding section \"%s\": %s", key, err)
		}
		c = append(c, section)
	}
	log.Trace(spew.Sdump(c))
	return &c, nil
}

// Execute runs the specified workers from the configuration
func (b Bridgr) Execute(filter []string) error {
	spin := spinner.New(spinner.CharSets[11], spinnerSpeed, spinner.WithWriter(os.Stderr))
	spin.Suffix = "  | Starting Bridgr"
	spin.FinalMSG = "Bridgr Completed!\n"
	_ = spin.Color("fgHiGreen")
	if isTty() && !bridgr.Verbose {
		spin.Start()
	}
	for _, w := range b {
		if len(filter) > 0 && !contains(w.Name(), filter) {
			log.Trace("skipping worker %s, not in %s", w.Name(), filter)
			continue
		}
		spin.Lock()
		spin.Suffix = fmt.Sprintf("  | Processing %s...", w.Name())
		spin.Unlock()
		log.Info("Processing %s...", w.Name())
		var err error
		if bridgr.DryRun {
			err = w.Setup()
		} else {
			err = w.Run()
		}
		if err != nil {
			log.Warn("Error processing %s: %s", w.Name(), err)
		}
	}
	spin.Stop()
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
	log.Trace("%s -> %s\n%s\n\n", f, t, data)
	return data, nil
}

func isTty() bool {
	return terminal.IsTerminal(syscall.Stdout)
}
