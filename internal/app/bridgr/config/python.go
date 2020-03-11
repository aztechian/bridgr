package config

import (
	"fmt"
	"path"
	"reflect"

	"github.com/docker/distribution/reference"
)

var defaultPyImg reference.Named

const defaultPySrc = "https://pypi.org"
const basePyImage = "python"

func init() {
	defaultPyImg, _ = reference.ParseNormalizedNamed(basePyImage + ":3.7") // https://github.com/wolever/pip2pi/issues/96 3.8 doesn't work
}

// Python is the configuration object specifically for the Python section of the config file
type Python struct {
	Packages []pythonPackage
	Version  pythonVersion
	Sources  []string
}

type pythonVersion reference.Named

type pythonPackage struct {
	Package string
	Version string
}

// BaseDir is the top-level directory name for all objects written out under the Python worker
func (p Python) BaseDir() string {
	return path.Join(BaseDir(), "python")
}

func (p Python) Count() int {
	return len(p.Packages)
}

func (p Python) Image() reference.Named {
	if p.Version == nil {
		return defaultPyImg
	}
	return p.Version
}

func (p Python) Repositories() []string {
	return p.Sources
}

func NewPython() *Python {
	return &Python{
		Version: defaultPyImg,
	}
}

func versionToPythonImage(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf((*pythonVersion)(nil)).Elem() {
		return data, nil
	}
	if f.Kind() == reflect.Float64 {
		return reference.ParseAnyReference(basePyImage + ":" + fmt.Sprintf("%.1f", data.(float64)))
	}
	if f.Kind() == reflect.String {
		return reference.ParseAnyReference(basePyImage + ":" + data.(string))
	}
	return data, nil
}

func arrayToPython(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || t != reflect.TypeOf(Python{}) {
		return data, nil
	}
	var pkgs []pythonPackage
	for _, p := range data.([]interface{}) {
		pkgs = append(pkgs, pythonPackage{Package: p.(string)})
	}

	return &Python{
		Version:  defaultPyImg,
		Packages: pkgs,
		Sources:  []string{defaultPySrc},
	}, nil
}
