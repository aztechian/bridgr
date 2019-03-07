package config

import (
	"log"
	"path/filepath"
	"strings"
)

// Files is the worker implementation for static File repositories
type Files struct {
	Items []FileItem
}

// FileItem is a discreet file defintion object
type FileItem struct {
	Source   string
	Target   string
	Protocol string
}

// BaseDir is the top-level directory name for all objects written out under the Files worker
func BaseDir() string {
	return "files"
}

func parseFiles(conf tempConfig) (Files, error) {
	files := Files{}
	for _, val := range conf.Files {
		log.Println("DEBUG: Parsing Files entry for:", val)
		newItem := FileItem{}
		var err error
		switch o := val.(type) {
		case string: //simple string entry
			err = newItem.parseSimple(o)
		case map[interface{}]interface{}: // complex type
			err = newItem.parseComplex(o)
		}
		if err != nil {
			log.Println("Error while parsing Files entry:", val)
		} else {
			files.Items = append(files.Items, newItem)
		}
	}
	return files, nil
}

func (f *FileItem) parseSimple(s string) error {
	f.Protocol = getFileProtocol(s)
	f.Source = s
	f.Target = getFileTarget(s)
	log.Println("populated FileItem", f)
	return nil
}

func (f *FileItem) parseComplex(s map[interface{}]interface{}) error {
	source := s["source"].(string)
	target := s["target"].(string)
	f.Protocol = getFileProtocol(source)
	f.Source = source
	if strings.HasSuffix(target, "/") {
		f.Target = filepath.Join(target, filepath.Base(source))
	} else {
		f.Target = target
	}
	return nil
}

func getFileProtocol(src string) string {
	if strings.HasPrefix(src, "/") {
		return "file"
	}
	return strings.Split(src, "://")[0]
}

func getFileTarget(src string) string {
	return filepath.Join(BaseDir(), filepath.Base(src))
}
