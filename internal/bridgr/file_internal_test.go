package bridgr

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
)

type httpMock struct {
	http.RoundTripper
}

func (m httpMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       ioutil.NopCloser(bytes.NewBufferString("OK")),
		Header:     make(http.Header),
	}, nil
}

type fakeFetcher struct {
	mock.Mock
}

func (mfi *fakeFetcher) httpFetch(httpClient *http.Client, source string, out io.WriteCloser, creds Credential) error {
	args := mfi.Called(httpClient, source, out, creds)
	return args.Error(0)
}

func (mfi *fakeFetcher) ftpFetch(source string, out io.WriteCloser, creds Credential) error {
	args := mfi.Called(source, out, creds)
	return args.Error(0)
}

func (mfi *fakeFetcher) fileFetch(source string, out io.WriteCloser) error {
	args := mfi.Called(source, out)
	return args.Error(0)
}

func (mfi *fakeFetcher) s3Fetch(client *s3.S3, source *url.URL, out io.WriteCloser, cred Credential) error {
	args := mfi.Called(client, source, out, cred)
	return args.Error(0)
}

func TestStringToFileItem(t *testing.T) {
	fileItemType := reflect.TypeOf(FileItem{})
	src, _ := url.Parse("ksanchez.mov")
	tests := []struct {
		name   string
		input  interface{}
		expect interface{}
	}{
		{"success", src.String(), FileItem{Source: src}},
		{"not string", 124, 124},
		{"parse error", "\007kitty.doc", nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := stringToFileItem(reflect.TypeOf(test.input), fileItemType, test.input)
			if err == nil && test.expect == nil {
				t.Errorf("Expected an error, but got %+v", err)
			}
			if test.expect != nil && !cmp.Equal(test.expect, result) {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestFilesDir(t *testing.T) {
	expected := BaseDir("files")
	result := File{}.dir()
	if !cmp.Equal(expected, result) {
		t.Error(cmp.Diff(expected, result))
	}
}

func TestFilesHttp(t *testing.T) {
	client := &http.Client{Transport: &httpMock{}}
	defaultSrc, _ := url.Parse("https://bluth.com/arock.pdf")
	tests := []struct {
		name   string
		item   FileItem
		target io.WriteCloser
		creds  Credential
		expect string
	}{
		{"success", FileItem{Source: defaultSrc}, newMockCloser(false), Credential{}, "OK"},
		{"with credentials", FileItem{Source: defaultSrc}, newMockCloser(false), Credential{Username: "myself", Password: "blued"}, "OK"},
	}

	fetcher := fileFetcher{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := fetcher.httpFetch(client, test.item.Source.String(), test.target, test.creds)
			if err != nil {
				t.Errorf("Unable to fetch HTTP source: %s", err)
			}
			if !cmp.Equal(test.expect, "OK") {
				t.Errorf(cmp.Diff(test.expect, "OK"))
			}
		})
	}
}

func TestFilesFtp(t *testing.T) {
	writer := fakeWriteCloser{}
	fetcher := fileFetcher{}
	err := fetcher.ftpFetch("ftp://bluth.com/arock.pdf", &writer, Credential{})
	if err == nil {
		t.Error("Expected FTP source to be unimplemented")
	}
}

func TestFilesFile(t *testing.T) {
	// want := "Awesome File Content."
	src := ("/arock.pdf")
	got := newMockCloser(false)
	fetcher := fileFetcher{}
	err := fetcher.fileFetch(src, got)
	if err == nil {
		t.Errorf("Unexpected success when opening %s from filesystem", src)
	}
	// if !cmp.Equal(want, got.String()) {
	// 	t.Errorf("Expected %s to be written to output file, but got %s", want, got.String())
	// }
}

func TestFileFetch(t *testing.T) {
	httpSrc, _ := url.Parse("https://bluth.com/solid/as/arock.ppt")
	ftpSrc, _ := url.Parse("ftp://bluth.com/solid/as/arock.ppt")
	localSrc, _ := url.Parse("/file.zip")
	otherSrc, _ := url.Parse("other://illusion")
	setupHttp := func(f *fakeFetcher) *mock.Call {
		return f.On("httpFetch", httpClient, mock.AnythingOfType("string"), &goodWriteCloser{}, Credential{})
	}
	setupFtp := func(f *fakeFetcher) *mock.Call {
		return f.On("ftpFetch", mock.AnythingOfType("string"), &goodWriteCloser{}, Credential{})
	}
	setupFile := func(f *fakeFetcher) *mock.Call {
		return f.On("fileFetch", mock.AnythingOfType("string"), &goodWriteCloser{})
	}

	tests := []struct {
		name   string
		item   FileItem
		setup  func(*fakeFetcher) *mock.Call
		expect interface{}
	}{
		{"http", FileItem{Source: httpSrc}, setupHttp, nil},
		{"ftp", FileItem{Source: ftpSrc}, setupFtp, nil},
		{"local", FileItem{Source: localSrc}, setupFile, nil},
		{"other", FileItem{Source: otherSrc}, nil, errors.New("gob")},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dest := newMockCloser(false)
			fetcher := fakeFetcher{}
			if test.setup != nil {
				test.setup(&fetcher).Return(test.expect)
			}
			err := test.item.fetch(&fetcher, &WorkerCredentialReader{}, dest)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
