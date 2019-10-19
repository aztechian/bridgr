package workers

import (
	"bridgr/internal/app/bridgr/config"
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

type fakeWriteCloser struct {
	bytes.Buffer
	isError bool
}

type httpMock struct {
	http.RoundTripper
}

func (wc *fakeWriteCloser) Close() error {
	return nil
}

func (wc *fakeWriteCloser) Write(p []byte) (n int, err error) {
	if wc.isError {
		return 0, errors.New("write error")
	}
	return wc.Write(p)
}

var fileSource, _ = url.Parse("/source1")
var httpSource, _ = url.Parse("http://nothing.net/file2")
var ftpSource, _ = url.Parse("ftp://nothing.net/file3")
var defaultConf = config.Files{
	Items: []config.FileItem{
		{Source: fileSource, Target: "file1"},
		{Source: httpSource, Target: "file2"},
		{Source: ftpSource, Target: "file3"},
	},
}

var stubWorker = Files{
	Config: &defaultConf,
	HTTP:   &http.Client{Transport: &httpMock{}},
}

func (m httpMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
		Header:     make(http.Header),
	}, nil
}

func TestFilesHttp(t *testing.T) {
	writer := fakeWriteCloser{}
	err := stubWorker.httpFetch(defaultConf.Items[1], &writer)
	if err != nil {
		t.Errorf("Unable to fetch HTTP source: %s", err)
	}
	if writer.String() != "OK" {
		t.Errorf("Expected HTTP response of OK, but got %s", writer.String())
	}
}

func TestFilesFtp(t *testing.T) {
	writer := fakeWriteCloser{}
	err := stubWorker.ftpFetch(defaultConf.Items[2], &writer)
	if err == nil {
		t.Error("Expected FTP source to be unimplemented")
	}
}

func TestFilesFile(t *testing.T) {
	want := "Awesome File Content."
	in := ioutil.NopCloser(bytes.NewBufferString(want))
	got := fakeWriteCloser{}
	err := stubWorker.fileFetch(in, &got)
	if err != nil {
		t.Errorf("Unable to fetch FILE source: %s", err)
	}
	if want != got.String() {
		t.Errorf("Expected %s to be written to output file, but got %s", want, got.String())
	}
}
