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
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
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

func (mfi *fakeFetcher) s3Fetch(client s3iface.S3API, source *url.URL, out io.WriteCloser) error {
	args := mfi.Called(client, source, out)
	return args.Error(0)
}

func (mfi *fakeFetcher) regionalClient(source *url.URL, creds Credential) *s3.S3 {
	args := mfi.Called(source, creds)
	return args.Get(0).(*s3.S3)
}

type mockS3Client struct {
	s3iface.S3API
	mock.Mock
}

func (m *mockS3Client) GetObject(input *s3.GetObjectInput) (*s3.GetObjectOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3.GetObjectOutput), args.Error(1)
}

func (m *mockS3Client) GetBucketLocation(input *s3.GetBucketLocationInput) (*s3.GetBucketLocationOutput, error) {
	args := m.Called(input)
	return args.Get(0).(*s3.GetBucketLocationOutput), args.Error(1)
}

func TestStringToFileItem(t *testing.T) {
	cmpOpts := cmpopts.IgnoreUnexported(FileItem{})
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
			if test.expect != nil && !cmp.Equal(test.expect, result, cmpOpts) {
				t.Error(cmp.Diff(test.expect, result, cmpOpts))
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
				t.Error(cmp.Diff(test.expect, "OK"))
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

func TestFilesS3(t *testing.T) {
	Verbose = true
	fetcher := fileFetcher{}
	s3Client := mockS3Client{}
	src, _ := url.Parse("s3://bluth.com/arock.pdf")
	tests := []struct {
		name   string
		client s3iface.S3API
		source *url.URL
		target *fakeWriteCloser
		creds  Credential
		region string
		expect string
	}{
		{"success", &s3Client, src, &fakeWriteCloser{}, Credential{}, "us-oc-1", "illusion"},
		{"with credentials", &s3Client, src, &fakeWriteCloser{}, Credential{Username: "myself", Password: "blued"}, "us-oc-1", "OK"},
		{"empty region", &s3Client, src, &fakeWriteCloser{}, Credential{}, "", "illusion"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s3Client := mockS3Client{}
			s3Client.On("GetBucketLocation", mock.Anything).Return(&s3.GetBucketLocationOutput{LocationConstraint: &test.region}, nil)
			s3Client.On("GetObject", mock.Anything).Return(&s3.GetObjectOutput{Body: io.NopCloser(bytes.NewReader([]byte(test.expect)))}, nil)

			err := fetcher.s3Fetch(&s3Client, test.source, test.target)
			if err != nil {
				t.Errorf("expected no errors, got %s", err)
			}
			if !cmp.Equal(test.expect, test.target.String()) {
				t.Error(cmp.Diff(test.expect, test.target.String()))
			}
		})
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
	s3Src, _ := url.Parse("s3://bluth-accounting/kitty/files.zip")
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
	setupS3 := func(f *fakeFetcher) *mock.Call {
		f.On("regionalClient", mock.Anything, Credential{}).Return(defaultS3)
		return f.On("s3Fetch", mock.Anything, mock.Anything, &goodWriteCloser{})
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
		{"s3", FileItem{Source: s3Src}, setupS3, nil},
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

func TestRegionalClient(t *testing.T) {
	originalS3 := defaultS3
	defer func() { defaultS3 = originalS3 }()
	src, _ := url.Parse("s3://bluth.com/plans/sudden-valley.doc")
	fakeRegion := "us-fake-1"
	fetcher := fileFetcher{}
	mock := &mockS3Client{}
	mock.On("GetBucketLocation", &s3.GetBucketLocationInput{Bucket: &src.Host}).Return(&s3.GetBucketLocationOutput{LocationConstraint: &fakeRegion}, nil)
	defaultS3 = mock
	client := fetcher.regionalClient(src, Credential{})

	if !cmp.Equal(fakeRegion, *client.Client.Config.Region) {
		t.Error(cmp.Diff(fakeRegion, *client.Client.Config.Region))
	}
}
