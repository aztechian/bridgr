package config

import (
	"path"
	"reflect"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
)

var defaultYumImage reference.Named

const baseYumImage = "centos"

func init() {
	defaultYumImage, _ = reference.ParseNormalizedNamed(baseYumImage + ":7")
}

// Yum is the normalized structure for workers to get YUM information from the config file
type Yum struct {
	Repos    []string
	Packages []string
	Version  yumVersion
}

type yumVersion reference.Named

// BaseDir is the top-level directory name for all objects written out under the Yum worker
func (y Yum) BaseDir() string {
	return path.Join(BaseDir(), "yum")
}

func (y Yum) Repositories() []string {
	return y.Repos
}

func (y Yum) Image() reference.Named {
	if y.Version == nil {
		return defaultYumImage
	}
	return y.Version
}

func versionToYumImage(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() == reflect.String && t == reflect.TypeOf((*yumVersion)(nil)).Elem() {
		if strings.Contains(data.(string), ":") {
			// if the config file has a "full" image given, use that.
			// Otherwise use our default image name, assuming version field is _just_ the version
			return reference.ParseNormalizedNamed(data.(string))
		}
		return reference.ParseAnyReference(baseYumImage + ":" + data.(string))
	}
	return data, nil
}

func arrayToYum(t reflect.Type, f reflect.Type, data interface{}) (interface{}, error) {
	if t.Kind() == reflect.Slice && f == reflect.TypeOf(Yum{}) {
		var pkgList []string
		for _, pkg := range data.([]interface{}) {
			pkgList = append(pkgList, pkg.(string))
		}
		return Yum{
			Version:  defaultYumImage,
			Packages: pkgList,
		}, nil
	}
	return data, nil
}

func (y Yum) hookFunction() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		debugHook,
		arrayToYum,
	)
}
