package workers

import (
	"bridgr/internal/app/bridgr"
	"bridgr/internal/app/bridgr/config"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path"
	"time"
)

// Files is the work type for fetching plain files of various protocols
type Files struct {
	Config *config.BridgrConf
	HTTP   *http.Client
}

// NewFiles is the constructor for a new Files worker struct
func NewFiles(conf *config.BridgrConf) Worker {
	_ = os.MkdirAll(conf.Files.BaseDir(), os.ModePerm)
	return &Files{
		Config: conf,
		HTTP: &http.Client{
			// TODO: this would be much better to do as a fallback - if regular (InsecureSkipVerify: false) fails first
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   time.Second * 20,
					KeepAlive: time.Second * 3, // we don't expect more than one connection to a server before we move on
				}).Dial,
				// this will be _really_ bad if someday we supported 2-way SSL
				TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // ignore SSL certificates
				ResponseHeaderTimeout: time.Second * 5,
			},
		},
	}
}

// Name returns the string name of the Files worker
func (worker *Files) Name() string {
	return "Files"
}

// Run sets up, creates and fetches static files based on the settings from the config file
func (worker *Files) Run() error {
	setupErr := worker.Setup()
	if setupErr != nil {
		return setupErr
	}
	for _, file := range worker.Config.Files.Items {
		var err error
		_ = os.MkdirAll(path.Dir(file.Target), os.ModePerm)
		out, err := os.Create(file.Target)
		if err != nil {
			bridgr.Printf("Unable to create target: %s", err)
			continue
		}
		switch file.Protocol {
		case "http", "https":
			err = worker.httpFetch(file, out)
		case "ftp":
			err = worker.ftpFetch(file, out)
		case "file":
			in, openErr := os.Open(file.Source)
			if openErr == nil {
				bridgr.Debugf("Copying local file: %s", file.Source)
				err = worker.fileFetch(in, out)
			}
		}
		if err != nil {
			bridgr.Printf("Files '%s' - %+s", file.Source, err)
			_ = os.Remove(out.Name())
		}
	}
	return nil
}

// Setup only does the setup step of the Files worker
func (worker *Files) Setup() error {
	return nil
}

func (worker *Files) ftpFetch(f config.FileItem, out io.WriteCloser) error {
	defer out.Close()
	bridgr.Debugf("Downloading FTP file: %s", f.Source)
	return fmt.Errorf("FTP support is not yet implemented, skipping %+s", f.Source)
}

func (worker *Files) httpFetch(f config.FileItem, out io.WriteCloser) error {
	// Get the data
	defer out.Close()
	bridgr.Debugf("Downloading HTTP/S file: %s", f.Source)
	resp, err := worker.HTTP.Get(f.Source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (worker *Files) fileFetch(in io.ReadCloser, out io.WriteCloser) error {
	defer out.Close()
	defer in.Close()
	_, err := io.Copy(out, in)
	return err
}
