package bridgr

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"text/template"

	"github.com/aztechian/bridgr/internal/bridgr/asset"
	"github.com/distribution/reference"
	"github.com/mitchellh/mapstructure"
	log "unknwon.dev/clog/v2"
)

var (
	pyImage reference.Named
	pyReqt  *template.Template
)

const defaultPySource = "https://pypi.org"

func init() {
	pyImage, _ = reference.ParseNormalizedNamed(baseImage["python"] + ":3.7") // https://github.com/wolever/pip2pi/issues/96 3.8 doesn't work
	pyReqt = asset.Template("requirements.txt")
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

// dir is the top-level directory name for all objects written out under the Python worker
func (p Python) dir() string {
	return BaseDir(p.Name())
}

// Image implements the Imager interface
func (p Python) Image() reference.Named {
	if p.Version == nil {
		return pyImage
	}
	return p.Version
}

// Name returns the name of this Configuration
func (p Python) Name() string {
	return "python"
}

func versionToPythonImage(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if t != reflect.TypeOf((*pythonVersion)(nil)).Elem() {
		return data, nil
	}
	if f.Kind() == reflect.Float64 {
		return reference.ParseAnyReference(baseImage["python"] + ":" + fmt.Sprintf("%.1f", data.(float64)))
	}
	if f.Kind() == reflect.String {
		return reference.ParseAnyReference(baseImage["python"] + ":" + data.(string))
	}
	return data, nil
}

func arrayToPython(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || t != reflect.TypeOf(Python{}) {
		return data, nil
	}
	var pkgs []pythonPackage
	for _, p := range data.([]interface{}) {
		if pkg, ok := p.(string); ok {
			pkgs = append(pkgs, pythonPackage{Package: pkg})
		}
	}
	return Python{
		Version:  pyImage,
		Sources:  []string{defaultPySource},
		Packages: pkgs,
	}, nil
}

// Hook implements the Parser interface, returns a function for use by mapstructure when parsing config files
func (p *Python) Hook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		versionToPythonImage,
		arrayToPython,
	)
}

// Setup creates the items that are needed to fetch artifacts for the Python worker. It does not actually fetch artifacts.
func (p Python) Setup() error {
	log.Trace("Called Python Setup()")
	_ = os.MkdirAll(p.dir(), DefaultDirPerms)
	reqt, err := os.Create(path.Join(p.dir(), "requirements.txt"))
	if err != nil {
		return fmt.Errorf("unable to create Python requirements file: %s", err)
	}

	return asset.RenderFile(pyReqt, p.Packages, reqt)
}

// Run fetches all artifacts for the Python configuration
func (p Python) Run() error {
	if err := p.Setup(); err != nil {
		return err
	}
	shell, err := asset.Load("python.sh")
	if err != nil {
		return err
	}

	batcher := newBatch(p.Image().String(), p.dir(), path.Join(p.dir(), "requirements.txt"), "/requirements.txt")
	return batcher.runContainer("bridgr_python", shell)
}
