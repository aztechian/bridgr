package bridgr

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
)

var httpClient = &http.Client{
	// TODO: this would be much better to do as a fallback - if regular (InsecureSkipVerify: false) fails first
	Transport: &http.Transport{
		Dial: (&net.Dialer{
			Timeout:   time.Second * 20,
			KeepAlive: time.Second * 3, // we don't expect more than one connection to a server before we move on
		}).Dial,
		// this will be _really_ bad if someday we supported 2-way SSL
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //nolint:gosec  // ignore SSL certificates
		ResponseHeaderTimeout: time.Second * 5,
	},
}

// File is the implementation for static File repositories
type File []FileItem

// FileItem is a discreet file definition object
type FileItem struct {
	Source *url.URL
	Target string
}

// BaseDir is the top-level directory name for all objects written out under the Files worker
func (f File) dir() string {
	return BaseDir(f.Name())
}

// Normalize sets the FileItems' Target filed to the proper destination string
func (fi FileItem) Normalize() string {
	return filepath.Join(BaseDir("files"), filepath.Base(fi.Source.String()))
}

// Fetch gets a FileItem from it's source and writes it to the destination
func (fi *FileItem) Fetch() error {
	fi.Target = fi.Normalize()
	out, err := os.Create(fi.Target)
	if err != nil {
		return err
	}
	creds, ok := new(WorkerCredentialReader).Read(fi.Source)
	if ok {
		Debugf("Found credentials for File %s", fi.Source.String())
	}
	switch fi.Source.Scheme {
	case "http", "https":
		return httpFetch(httpClient, fi.Source.String(), out, creds)
	case "ftp":
		return ftpFetch(fi.Source.String(), out, creds)
	case "file", "":
		return fileFetch(fi.Source.String(), out)
	default:
		Printf("unsupported FileItem schema: %s, from %s", fi.Source.Scheme, fi.Source)
	}
	return nil
}

// Image returns the Named image for executing
func (f File) Image() reference.Named {
	return nil
}

// Name returns the name of this Configuration
func (f File) Name() string {
	return "files"
}

func stringToFileItem(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.String || t != reflect.TypeOf(FileItem{}) {
		return data, nil
	}
	url, _ := url.Parse(data.(string))
	return FileItem{Source: url}, nil
}

// Hook implements the Parser interface, returns a function for use by mapstructure when parsing config files
func (f *File) Hook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		stringToFileItem,
	)
}

// Run sets up, creates and fetches static files based on the settings from the config file
func (f File) Run() error {
	if err := f.Setup(); err != nil {
		return err
	}
	for _, item := range f {
		if err := item.Fetch(); err != nil {
			Printf("Files '%s' - %+s", item.Source.String(), err)
			_ = os.Remove(item.Normalize())
		}
	}
	return nil
}

// Setup only does the setup step of the Files worker
func (f File) Setup() error {
	Debug("Called Files.Setup()")
	_ = os.MkdirAll(f.dir(), os.ModePerm)
	return nil
}

func fileFetch(source string, out io.WriteCloser) error {
	if in, openErr := os.Open(source); openErr == nil {
		Debugf("Copying local file: %s", source)
		defer out.Close()
		defer in.Close()
		_, err := io.Copy(out, in)
		if err != nil {
			return err
		}
	} else {
		return openErr
	}
	return nil
}

func ftpFetch(source string, out io.WriteCloser, creds Credential) error {
	defer out.Close()
	Debugf("Downloading FTP file: %s", source)
	return fmt.Errorf("FTP support is not yet implemented, skipping %+s", source)
}

func httpFetch(httpClient *http.Client, url string, out io.WriteCloser, creds Credential) error {
	// Get the data
	defer out.Close()

	Debugf("Downloading HTTP/S file: %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	if len(creds.Username+creds.Password) > 0 {
		req.SetBasicAuth(creds.Username, creds.Password)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
