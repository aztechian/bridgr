package workers

import (
	"bridgr/internal/app/bridgr/config"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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
				// this will be _really_ bad if someday we supported 2-way SSL
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore SSL certificates
			},
			Timeout: time.Second * 10,
		},
	}
}

// Name returns the string name of the Files worker
func (worker *Files) Name() string {
	return "Files"
}

// Run sets up, creates and fetches static files based on the settings from the config file
func (worker *Files) Run() error {
	err := worker.Setup()
	if err != nil {
		return err
	}
	for _, file := range worker.Config.Files.Items {
		out, err := os.Create(file.Target)
		if err != nil {
			log.Printf("Unable to create target: %s", err)
			continue
		}
		switch file.Protocol {
		case "http", "https":
			err = worker.httpFetch(file, out)
		case "ftp":
			err = worker.ftpFetch(file, out)
		case "file":
			err = worker.fileFetch(file, out)
		}
		if err != nil {
			log.Printf("Files - %+s", err)
		}
		out.Close()
	}
	return nil
}

// Setup only does the setup step of the Files worker
func (worker *Files) Setup() error {
	return nil
}

func (worker *Files) ftpFetch(f config.FileItem, out io.Writer) error {
	return fmt.Errorf("FTP support is not yet implemented, skipping %+s", f.Source)
}

func (worker *Files) httpFetch(f config.FileItem, out io.Writer) error {
	// Get the data
	resp, err := worker.HTTP.Get(f.Source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func (worker *Files) fileFetch(f config.FileItem, out io.Writer) error {
	in, err := os.Open(f.Source)
	if err != nil {
		return err
	}
	defer in.Close()

	_, err = io.Copy(out, in)
	return err
}
