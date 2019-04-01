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

	"github.com/davecgh/go-spew/spew"
)

// Files is the work type for fetching plain files of various protocols
type Files struct{}

// Run sets up, creates and fetches static files based on the settings from the config file
func (f *Files) Run(conf config.BridgrConf) error {
	f.Setup(conf)
	for _, file := range conf.Files.Items {
		var err error
		switch file.Protocol {
		case "http", "https":
			err = httpFetch(file)
		case "ftp":
			err = ftpFetch(file)
		case "file":
			err = fileFetch(file)
		}
		if err != nil {
			log.Printf("Files - %+s", err)
		}
	}
	return nil
}

// Setup only does the setup step of the Files worker
func (f *Files) Setup(conf config.BridgrConf) error {
	log.Println("Called Files.setup()")
	spew.Dump(conf.Files)
	os.Mkdir(conf.Files.BaseDir(), os.ModePerm)
	return nil
}

func ftpFetch(f config.FileItem) error {
	return fmt.Errorf("FTP support is not yet implemented, skipping %+s", f.Source)
}

func httpFetch(f config.FileItem) error {
	// TODO this would be much better to do as a fallback - if regular (InsecureSkipVerify: true) fails first
	// this will be _really_ bad if someday we supported 2-way SSL
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore SSL certificates
	}

	client := &http.Client{
		Timeout:   time.Second * 10,
		Transport: transport,
	}

	// Get the data
	resp, err := client.Get(f.Source)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(f.Target)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func fileFetch(f config.FileItem) error {
	in, err := os.Open(f.Source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(f.Target)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
