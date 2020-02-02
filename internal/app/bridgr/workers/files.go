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
	Config *config.Files
	HTTP   *http.Client
	workerCredentialReader
}

// NewFiles is the constructor for a new Files worker struct
func NewFiles(conf *config.BridgrConf) Worker {
	_ = os.MkdirAll(conf.Files.BaseDir(), os.ModePerm)
	return &Files{
		Config: &conf.Files,
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
func (f *Files) Name() string {
	return "Files"
}

// Run sets up, creates and fetches static files based on the settings from the config file
func (f *Files) Run() error {
	setupErr := f.Setup()
	if setupErr != nil {
		return setupErr
	}
	for _, item := range f.Config.Items {
		var err error
		_ = os.MkdirAll(path.Dir(item.Target), os.ModePerm)
		out, err := os.Create(item.Target)
		if err != nil {
			bridgr.Printf("Unable to create target: %s", err)
			continue
		}
		creds, ok := f.Read(item.Source)
		if ok {
			bridgr.Printf("Found credentials for File %s", item.Source.String())
		}
		switch item.Source.Scheme {
		case "http", "https":
			err = f.httpFetch(item, out, creds)
		case "ftp":
			err = f.ftpFetch(item, out, creds)
		case "file", "":
			err = f.fileFetch(item, out)
		}
		if err != nil {
			bridgr.Printf("Files '%s' - %+s", item.Source.String(), err)
			_ = os.Remove(out.Name())
		}
	}
	return nil
}

// Setup only does the setup step of the Files worker
func (f *Files) Setup() error {
	bridgr.Print("Called Files.Setup()")
	return nil
}

func (f *Files) fileFetch(item config.FileItem, out io.WriteCloser) error {
	if in, openErr := os.Open(item.Source.String()); openErr == nil {
		bridgr.Debugf("Copying local file: %s", item.Source.String())
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

func (f *Files) ftpFetch(item config.FileItem, out io.WriteCloser, creds Credential) error {
	defer out.Close()
	bridgr.Debugf("Downloading FTP file: %s", item.Source.String())
	return fmt.Errorf("FTP support is not yet implemented, skipping %+s", item.Source.String())
}

func (f *Files) httpFetch(item config.FileItem, out io.WriteCloser, creds Credential) error {
	// Get the data
	defer out.Close()

	bridgr.Debugf("Downloading HTTP/S file: %s", item.Source)
	req, err := http.NewRequest(http.MethodGet, item.Source.String(), nil)
	if err != nil {
		return err
	}
	if len(creds.Username+creds.Password) > 0 {
		req.SetBasicAuth(creds.Username, creds.Password)
	}
	resp, err := f.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
