package config

import (
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
	Source *url.URL
	Target string
}

// BaseDir is the top-level directory name for all objects written out under the Files worker
func (f *Files) BaseDir() string {
	return path.Join(BaseDir(), "files")
}

// func parseFiles(conf tempConfig) Files {
//   files := Files{}
//   for _, val := range conf.Files {
//     newItem := FileItem{}
//     var err error
//     switch o := val.(type) {
//     case string: //simple string entry
//       err = newItem.parseSimple(o)
//     case map[interface{}]interface{}: // complex type
//       err = newItem.parseComplex(o)
//     default:
//       err = fmt.Errorf("Unsupported File type in config - %T", o)
//     }
//     if err != nil {
//       bridgr.Println(err)
//     } else {
//       files.Items = append(files.Items, newItem)
//     }
//   }
//   bridgr.Debugf("Final Files configuration %+v", files)
//   return files
// }

func (f *FileItem) parseSimple(s string) error {
	url, err := url.Parse(s)
	if err != nil {
		return err
	}
	// setProtocol(url)
	f.Source = url
	f.Target = getFileTarget(s)
	return nil
}

func (f *FileItem) parseComplex(s map[interface{}]interface{}) error {
	source := s["source"].(string)
	target := s["target"].(string)
	url, err := url.Parse(source)
	if err != nil {
		return err
	}
	f.Source = url
	if strings.HasSuffix(target, "/") {
		f.Target = filepath.Join(new(Files).BaseDir(), target, filepath.Base(source))
	} else {
		f.Target = getFileTarget(target)
	}
	return nil
}

func getFileTarget(src string) string {
	return filepath.Join(new(Files).BaseDir(), filepath.Base(src))
}
