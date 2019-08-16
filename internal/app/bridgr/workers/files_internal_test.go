package workers

import (
	"bridgr/internal/app/bridgr/config"
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
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

var defaultConf = config.BridgrConf{
	Files: config.Files{
		Items: []config.FileItem{
			{Source: "/source1", Target: "file1", Protocol: "file"},
			{Source: "http://nothing.net/file2", Target: "file2", Protocol: "http"},
			{Source: "ftp://nothing.net/file3", Target: "file3", Protocol: "ftp"},
		},
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
	err := stubWorker.httpFetch(defaultConf.Files.Items[1], &writer)
	if err != nil {
		t.Errorf("Unable to fetch HTTP source: %s", err)
	}
	if writer.String() != "OK" {
		t.Errorf("Expected HTTP response of OK, but got %s", writer.String())
	}
}

func TestFilesFtp(t *testing.T) {
	writer := fakeWriteCloser{}
	err := stubWorker.ftpFetch(defaultConf.Files.Items[2], &writer)
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
