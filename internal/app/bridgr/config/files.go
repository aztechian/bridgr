package config

import (
	"net/url"
	"path"
	"path/filepath"
	"reflect"

	"github.com/docker/distribution/reference"
)

// Files is the worker implementation for static File repositories
type Files []FileItem

// FileItem is a discreet file definition object
type FileItem struct {
	Source *url.URL
	Target string
}

// BaseDir is the top-level directory name for all objects written out under the Files worker
func (f *Files) BaseDir() string {
	return path.Join(BaseDir(), "files")
}

func (f *Files) Count() int {
	return len(*f)
}

func (f *Files) Image() reference.Named {
	return nil
}

func (fi *FileItem) GetTarget() string {
	src := fi.Target
	if src == "" {
		src = fi.Source.EscapedPath()
	}
	return getFileTarget(src)
}

func NewFileItem(s string) FileItem {
	url, _ := url.Parse(s)
	return FileItem{Source: url}
}

func getFileTarget(src string) string {
	return filepath.Join(new(Files).BaseDir(), filepath.Base(src))
}

func stringToFileItem(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(FileItem{}) {
		return data, nil
	}
	return NewFileItem(data.(string)), nil
}
