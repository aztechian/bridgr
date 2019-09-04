package config

import (
	"bridgr/internal/app/bridgr"
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"
)

// Files is the worker implementation for static File repositories
type Files struct {
	Items []FileItem
}

// FileItem is a discreet file definition object
type FileItem struct {
	Source   string
	Target   string
	Protocol string
}

// BaseDir is the top-level directory name for all objects written out under the Files worker
func (f *Files) BaseDir() string {
	return path.Join(BaseDir(), "files")
}

func parseFiles(conf tempConfig) Files {
	files := Files{}
	for _, val := range conf.Files {
		newItem := FileItem{}
		var err error
		switch o := val.(type) {
		case string: //simple string entry
			err = newItem.parseSimple(o)
		case map[interface{}]interface{}: // complex type
			err = newItem.parseComplex(o)
		default:
			err = fmt.Errorf("Unsupported File type in config - %T", o)
		}
		if err != nil {
			bridgr.Println(err)
		} else {
			files.Items = append(files.Items, newItem)
		}
	}
	bridgr.Debugf("Final Files configuration %+v", files)
	return files
}

func (f *FileItem) parseSimple(s string) error {
	f.Protocol = getFileProtocol(s)
	f.Source = s
	f.Target = getFileTarget(s)
	return nil
}

func (f *FileItem) parseComplex(s map[interface{}]interface{}) error {
	source := s["source"].(string)
	target := s["target"].(string)
	f.Protocol = getFileProtocol(source)
	f.Source = source
	if strings.HasSuffix(target, "/") {
		f.Target = filepath.Join(new(Files).BaseDir(), target, filepath.Base(source))
	} else {
		f.Target = getFileTarget(target)
	}
	return nil
}

func getFileProtocol(src string) string {
	url, _ := url.Parse(src)
	if url.Scheme == "" {
		return "file"
	}
	return url.Scheme
}

func getFileTarget(src string) string {
	return filepath.Join(new(Files).BaseDir(), filepath.Base(src))
}
