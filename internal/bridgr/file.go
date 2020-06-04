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

type fetcher interface {
	httpFetch(*http.Client, string, io.WriteCloser, Credential) error
	ftpFetch(string, io.WriteCloser, Credential) error
	fileFetch(string, io.WriteCloser) error
}

type fileFetcher struct{}

// BaseDir is the top-level directory name for all objects written out under the Files worker
func (f File) dir() string {
	return BaseDir(f.Name())
}

// Normalize sets the FileItems' Target filed to the proper destination string
func (fi FileItem) Normalize() string {
	return filepath.Join(BaseDir("files"), fi.Target, filepath.Base(fi.Source.String()))
}

// Fetch gets a FileItem from it's source and writes it to the destination
func (fi *FileItem) fetch(fetcher fetcher, cr CredentialReader, output io.WriteCloser) error {
	creds, ok := cr.Read(fi.Source)
	if ok {
		Debugf("Found credentials for File %s", fi.Source.String())
	}
	switch fi.Source.Scheme {
	case "http", "https":
		return fetcher.httpFetch(httpClient, fi.Source.String(), output, creds)
	case "ftp":
		return fetcher.ftpFetch(fi.Source.String(), output, creds)
	case "file", "":
		return fetcher.fileFetch(fi.Source.String(), output)
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
	url, err := url.Parse(data.(string))
	return FileItem{Source: url}, err
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
	credentials := WorkerCredentialReader{}
	fetcher := fileFetcher{}
	for _, item := range f {
		out, createErr := os.Create(item.Target)
		if createErr != nil {
			Printf("Unable to create local file %s (for %s)", item.Target, item.Source.String())
			continue
		}
		if err := item.fetch(&fetcher, &credentials, out); err != nil {
			Printf("Files '%s' - %+s", item.Source.String(), err)
			_ = os.Remove(item.Target)
		}
	}
	return nil
}

// Setup only does the setup step of the Files worker
func (f File) Setup() error {
	Debug("Called Files.Setup()")
	_ = os.MkdirAll(f.dir(), os.ModePerm)
	for _, item := range f {
		item.Target = item.Normalize()
	}
	return nil
}

func (ff *fileFetcher) fileFetch(source string, out io.WriteCloser) error {
	in, openErr := os.Open(source)
	if openErr != nil {
		return openErr
	}

	Debugf("Copying local file: %s", source)
	defer out.Close()
	defer in.Close()
	_, err := io.Copy(out, in)
	return err
}

func (ff *fileFetcher) ftpFetch(source string, out io.WriteCloser, creds Credential) error {
	defer out.Close()
	Debugf("Downloading FTP file: %s", source)
	return fmt.Errorf("FTP support is not yet implemented, skipping %+s", source)
}

func (ff *fileFetcher) httpFetch(httpClient *http.Client, source string, out io.WriteCloser, creds Credential) error {
	// Get the data
	defer out.Close()

	Debugf("Downloading HTTP/S file: %s", source)
	req, err := http.NewRequest(http.MethodGet, source, nil)
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
