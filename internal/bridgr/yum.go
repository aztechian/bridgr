package bridgr

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"reflect"
	"strings"
	"text/template"

	"github.com/aztechian/bridgr/internal/bridgr/asset"
	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
	log "unknwon.dev/clog/v2"
)

var (
	yumImage  reference.Named
	yumScript *template.Template
	yumRepo   *template.Template
)

func init() {
	yumImage, _ = reference.ParseNormalizedNamed(baseImage["yum"] + ":7")
	yumScript = asset.Template("yum.sh")
	yumRepo = asset.Template("yum.repo")
}

// Yum sets up and creates an YUM repository based on user configuration
type Yum struct {
	Repos    []string
	Packages []string
	Version  yumVersion
}

type yumVersion reference.Named

// Dir is the top-level directory name for all objects written out under the Yum worker
func (y Yum) dir() string {
	return BaseDir(y.Name())
}

// Name returns the name of this Configuration
func (y Yum) Name() string {
	return "yum"
}

// Image returns the docker image that will be used for the batch execution
func (y Yum) Image() reference.Named {
	if y.Version == nil {
		return yumImage
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
		return reference.ParseAnyReference(baseImage["yum"] + ":" + data.(string))
	}
	return data, nil
}

func arrayToYum(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || t != reflect.TypeOf(Yum{}) {
		return data, nil
	}
	var pkgList []string
	for _, pkg := range data.([]interface{}) {
		if pkg, ok := pkg.(string); ok {
			pkgList = append(pkgList, pkg)
		}
	}
	return Yum{
		Version:  yumImage,
		Packages: pkgList,
	}, nil
}

// Hook implements the Parser interface, returns a function for use by mapstructure when parsing config files
func (y *Yum) Hook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		versionToYumImage,
		arrayToYum,
	)
}

// Run sets up, creates and fetches a YUM repository based on the settings from the config file
func (y Yum) Run() error {
	if err := y.Setup(); err != nil {
		return err
	}

	script := bytes.Buffer{}
	if err := asset.Render(yumScript, y.Packages, &script); err != nil {
		return err
	}

	batcher := newBatch(y.Image().String(), y.dir(), path.Join(y.dir(), "bridgr.repo"), "/etc/yum.repos.d/bridgr.repo")
	return batcher.runContainer("bridgr_yum", script.String())
}

// Setup only does the setup step of the YUM worker
func (y Yum) Setup() error {
	log.Trace("Called Yum Setup()")
	_ = os.MkdirAll(y.dir(), os.ModePerm)

	repoFile, err := os.Create(path.Join(y.dir(), "bridgr.repo"))
	if err != nil {
		return fmt.Errorf("Unable to create YUM repo file: %s", err)
	}

	err = asset.RenderFile(yumRepo, y.Repos, repoFile)
	if err != nil {
		return err
	}
	return nil
}
