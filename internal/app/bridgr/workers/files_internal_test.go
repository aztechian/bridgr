package workers

import (
	"bridgr/internal/app/bridgr/config"
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
)

type MemWriteCloser struct {
	bytes.Buffer
}

type HTTPMock struct {
	http.RoundTripper
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
	HTTP:   &http.Client{Transport: &HTTPMock{}},
}

func (m HTTPMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
		Header:     make(http.Header),
	}, nil
}

func TestFilesHttp(t *testing.T) {
	writer := MemWriteCloser{}
	err := stubWorker.httpFetch(defaultConf.Files.Items[1], &writer)
	if err != nil {
		t.Errorf("Unable to fetch HTTP source: %s", err)
	}
	if writer.String() != "OK" {
		t.Errorf("Expected HTTP response of OK, but got %s", writer.String())
	}
}

func TestFilesFtp(t *testing.T) {
	err := stubWorker.ftpFetch(defaultConf.Files.Items[2], nil)
	if err == nil {
		t.Error("Expected FTP source to be unimplemented")
	}
}

// It doesn't make sense to test fileFetch(), because this relies on the OS's file system. The only other call
//  here is io.Copy() - which we'll assume is working. I don't like moving the file opening to Run(), then that becomes
//  untestable instead.
// func TestFilesFile(t *testing.T) {
// 	err := stubWorker.fileFetch(defaultConf.Files.Items[0], nil)
// 	if err != nil {
// 		t.Errorf("Unable to fetch FILE source: %s", err)
// 	}
// }
