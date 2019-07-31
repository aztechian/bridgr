package config

import (
	"bridgr/internal/app/bridgr"
	"path"

	"github.com/docker/distribution/reference"
)

// Python is the configuration object specifically for the Python section of the config file
type Python struct {
	Items []string
	Image reference.Named
}

var defaultPyImg, _ = reference.ParseNormalizedNamed("python:2")

// BaseDir is the top-level directory name for all objects written out under the Python worker
func (p *Python) BaseDir() string {
	return path.Join(BaseDir(), "python")
}

func parsePython(config tempConfig) Python {
	py := Python{Image: defaultPyImg}
	switch c := config.Python.(type) {
	case []interface{}:
		_ = py.parsePackages(c)
	case map[interface{}]interface{}:
		if _, present := c["version"]; present {
			var err error
			py.Image, err = reference.ParseNormalizedNamed("python:" + c["version"].(string))
			if err != nil {
				bridgr.Debugf("Error using Python image of 'python:%s', falling back to %s", c["version"].(string), defaultPyImg.String())
				py.Image = defaultPyImg
			}
		}
		pkgList := c["packages"].([]interface{})
		_ = py.parsePackages(pkgList)
	default:
		bridgr.Debugf("Unknown configuration section for Python: %+s", c)
	}
	bridgr.Print(py)
	return py
}

func (p *Python) parsePackages(pkgList []interface{}) error {
	for _, pkg := range pkgList {
		switch pkgObj := pkg.(type) {
		case string:
			p.Items = append(p.Items, pkgObj)
		case map[interface{}]interface{}:
			p.Items = append(p.Items, pkgObj["package"].(string)+pkgObj["version"].(string))
		}
	}
	return nil
}
