package bridgr

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
	log "unknwon.dev/clog/v2"
)

var (
	s3session                   = session.Must(session.NewSessionWithOptions(session.Options{SharedConfigState: session.SharedConfigEnable}))
	defaultS3     s3iface.S3API = s3.New(s3session) // used for getting the desired files bucket location. A new client is created when the files region is known
	headerTimeout               = time.Second * 5   // we expect to get headers coming back in 5 seconds
	keepAlive                   = time.Second * 3   // we create a new client for each file, so no keepalive needed as we won't reuse the client
	httpClient                  = &http.Client{
		// TODO: this would be much better to do as a fallback - if regular (InsecureSkipVerify: false) fails first
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   FileTimeout,
				KeepAlive: keepAlive,
			}).Dial,
			// this will be _really_ bad if someday we supported 2-way SSL
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, //nolint:gosec  // ignore SSL certificates
			ResponseHeaderTimeout: headerTimeout,
		},
	}
)

// File is the implementation for static File repositories
type File []*FileItem

// FileItem is a discreet file definition object
type FileItem struct {
	Source     *url.URL
	Target     string
	normalized bool
}

type fetcher interface {
	httpFetch(*http.Client, string, io.WriteCloser, Credential) error
	ftpFetch(string, io.WriteCloser, Credential) error
	fileFetch(string, io.WriteCloser) error
	s3Fetch(s3iface.S3API, *url.URL, io.WriteCloser) error
	regionalClient(*url.URL, Credential) *s3.S3
}

type fileFetcher struct{}

// BaseDir is the top-level directory name for all objects written out under the Files worker
func (f File) dir() string {
	return BaseDir(f.Name())
}

// Normalize sets the FileItems' Target field to the proper destination string
func (fi *FileItem) Normalize(basedir string) string {
	if fi.normalized {
		return fi.Target
	}
	fi.normalized = true
	fi.Target = filepath.Join(basedir, fi.Target, filepath.Base(fi.Source.String()))
	return fi.Target
}

func (fi FileItem) String() string {
	return fi.Source.String()
}

// Fetch gets a FileItem from it's source and writes it to the destination
func (fi *FileItem) fetch(fetcher fetcher, cr CredentialReader, output io.WriteCloser) error {
	creds, ok := cr.Read(fi.Source)
	if ok {
		log.Trace("Found credentials for File %s", fi.Source.String())
	}
	switch fi.Source.Scheme {
	case "http", "https":
		return fetcher.httpFetch(httpClient, fi.Source.String(), output, creds)
	case "ftp":
		return fetcher.ftpFetch(fi.Source.String(), output, creds)
	case "file", "":
		return fetcher.fileFetch(fi.Source.String(), output)
	case "s3":
		client := fetcher.regionalClient(fi.Source, creds)
		return fetcher.s3Fetch(client, fi.Source, output)
	default:
		log.Info("unsupported FileItem schema: %s, from %s", fi.Source.Scheme, fi.Source)
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
			log.Info("Unable to create local file '%s' (for %s) %s", item.Target, item.Source.String(), createErr)
			continue
		}
		if err := item.fetch(&fetcher, &credentials, out); err != nil {
			log.Info("Files '%s' - %+s", item.Source.String(), err)
			_ = os.Remove(item.Target)
		}
	}
	return nil
}

// Setup only does the setup step of the Files worker
func (f File) Setup() error {
	log.Trace("Called Files.Setup()")
	_ = os.MkdirAll(f.dir(), os.ModePerm)
	for _, item := range f {
		_ = item.Normalize(f.dir())
	}
	return nil
}

func (ff *fileFetcher) fileFetch(source string, out io.WriteCloser) error {
	in, openErr := os.Open(source)
	if openErr != nil {
		return openErr
	}

	log.Trace("Copying local file: %s", source)
	defer out.Close()
	defer in.Close()
	_, err := io.Copy(out, in)
	return err
}

func (ff *fileFetcher) ftpFetch(source string, out io.WriteCloser, creds Credential) error {
	defer out.Close()
	log.Trace("Downloading FTP file: %s", source)
	return fmt.Errorf("FTP support is not yet implemented, skipping %+s", source)
}

func (ff *fileFetcher) httpFetch(httpClient *http.Client, source string, out io.WriteCloser, creds Credential) error {
	// Get the data
	defer out.Close()

	log.Trace("Downloading HTTP/S file: %s", source)
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

func (ff *fileFetcher) s3Fetch(client s3iface.S3API, source *url.URL, out io.WriteCloser) error {
	defer out.Close()
	if client == (*s3.S3)(nil) {
		return errors.New("Invalid S3 client, unable to copy file")
	}
	log.Trace("Downloading S3 file: %s", source.String())
	resp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(source.Host),
		Key:    aws.String(source.Path),
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func (ff *fileFetcher) regionalClient(source *url.URL, creds Credential) *s3.S3 {
	loc, err := defaultS3.GetBucketLocation(&s3.GetBucketLocationInput{Bucket: &source.Host})
	if err != nil {
		log.Info("unable to find bucket '%s': %s. Validate bucket is correct and credentials are valid.", source.Host, err)
		return nil
	}
	region := s3.NormalizeBucketLocation(aws.StringValue(loc.LocationConstraint))

	cfg := aws.NewConfig().WithRegion(aws.StringValue(&region))
	if len(creds.Username) > 0 {
		cfg.WithCredentials(credentials.NewStaticCredentials(creds.Username, creds.Password, ""))
		log.Trace("Using static AWS credentials from bridgr environment")
	}
	return s3.New(s3session.Copy(cfg))
}
