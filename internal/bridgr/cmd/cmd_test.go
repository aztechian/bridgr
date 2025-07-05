package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/distribution/reference"
	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/mock"
)

type fakeConfig struct {
	mock.Mock
}

func (c *fakeConfig) Image() reference.Named {
	args := c.Called()
	return args.Get(0).(reference.Named)
}

func (c *fakeConfig) Hook() mapstructure.DecodeHookFunc {
	args := c.Called()
	return args.Get(0).(mapstructure.DecodeHookFunc)
}

func (c *fakeConfig) Name() string {
	args := c.Called()
	return args.String(0)
}

func (c *fakeConfig) Run() error {
	args := c.Called()
	return args.Error(0)
}

func (c *fakeConfig) Setup() error {
	args := c.Called()
	return args.Error(0)
}

type errReader int

func (r *errReader) Read(p []byte) (int, error) {
	return 0, errors.New("Fake read error")
}

var (
	yamlGit = []byte(`---
git:
  - https://repo.org/something.git
`)

	yamlYum = []byte(`---
yum:
  - forgetmenow
`)

	yamlRuby = []byte(`---
ruby:
  - wall
`)

	yamlBlah = []byte(`---
blah:
  - bobloblaw
`)

	yamlDocker = []byte(`---
docker:
  - hub.bluth.org/gob:latest
`)

	yamlFile = []byte(`---
files:
  - /buster.gif
`)

	yamlPython = []byte(`---
python:
  - bobloblaw
`)

	yamlHelm = []byte(`---
helm:
  - https://repo.bluth.org/illusions/trick-1.0.0.tgz
`)

	namedComparer = cmp.Comparer(func(got, want reference.Named) bool {
		return got.String() == want.String()
	})
)

func TestStringToImage(t *testing.T) {
	img, _ := reference.ParseNormalizedNamed("tobias:nevernude")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid type", reflect.TypeOf(39), 39, false, 39},
		{"valid", reflect.TypeOf((*reference.Reference)(nil)).Elem(), "tobias:nevernude", false, img},
		{"invalid image", reflect.TypeOf((*reference.Reference)(nil)).Elem(), "", true, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := stringToImage(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(test.expect, result, namedComparer) && !test.isError {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestStringToURL(t *testing.T) {
	expectedURL, _ := url.Parse("tobias.com/analrapist.html")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid image", reflect.TypeOf(4.302), 4.302, false, 4.302},
		{"valid", reflect.TypeOf(&url.URL{}), expectedURL.String(), false, expectedURL},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := stringToURL(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(test.expect, result, namedComparer) && !test.isError {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestDebugHook(t *testing.T) {
	result, err := debugHook(reflect.TypeOf(false), reflect.TypeOf(42), false)
	if !cmp.Equal(false, result) {
		t.Error(cmp.Diff(false, result))
	}
	if err != nil {
		t.Error(err)
	}
}

func TestDecode(t *testing.T) {
	config := bridgr.Ruby{}
	err := decode(&config, []interface{}{"all"})
	if err != nil {
		t.Error(err)
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		item   string
		filter []string
		expect bool
	}{
		{"success", "gob", []string{"gob"}, true},
		{"empty", "kitty", []string{}, true},
		{"all", "oscar", []string{"all"}, true},
		{"all empty", "", []string{"all"}, true},
		{"multiple filters with all", "george-sr", []string{"oscar", "all"}, false},
		{"multiple filter with all first", "george-sr", []string{"all", "oscar"}, true},
		{"multiple with match", "michael", []string{"george", "michael"}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := contains(test.item, test.filter)
			if !cmp.Equal(test.expect, result) {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		config  *fakeConfig
		dryrun  bool
		isError bool
	}{
		{"basic", &fakeConfig{}, false, false},
		{"with error", &fakeConfig{}, false, true},
		{"setup", &fakeConfig{}, true, false},
		{"setup error", &fakeConfig{}, true, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := []bridgr.Configuration{test.config}
			test.config.On("Name").Return("fake")
			if test.dryrun {
				if test.isError {
					test.config.On("Setup").Return(fmt.Errorf("%s", "fake error"))
				} else {
					test.config.On("Setup").Return(nil)
				}
			} else {
				if test.isError {
					test.config.On("Run").Return(fmt.Errorf("%s", "fake error"))
				} else {
					test.config.On("Run").Return(nil)
				}
			}
			bridgr.DryRun = test.dryrun
			err := Bridgr(c).Execute([]string{"fake"})
			if err != nil {
				t.Error(err)
			}
		})
	}
}

func TestNewCmd(t *testing.T) {
	tests := []struct {
		name     string
		yaml     *bytes.Reader
		negative bool
	}{
		{"git", bytes.NewReader(yamlGit), false},
		{"yum", bytes.NewReader(yamlYum), false},
		{"ruby", bytes.NewReader(yamlRuby), false},
		{"docker", bytes.NewReader(yamlDocker), false},
		{"python", bytes.NewReader(yamlPython), false},
		{"files", bytes.NewReader(yamlFile), false},
		{"helm", bytes.NewReader(yamlHelm), false},
		{"blah", bytes.NewReader(yamlBlah), false},
		{"failed read", bytes.NewReader(yamlBlah), true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var reader io.Reader = test.yaml
			if test.negative {
				reader = new(errReader)
			}

			cfg, err := New(io.NopCloser(reader))
			if cfg == nil {
				t.Error(cfg)
			}
			if err != nil && !test.negative {
				t.Error(err)
			}
		})
	}
}
