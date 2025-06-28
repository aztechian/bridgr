package bridgr

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"testing"

	"github.com/distribution/reference"
	"github.com/google/go-cmp/cmp"
)

type mockFoundCredentials struct {
	Credential
	ShouldFail bool
}

func (mc mockFoundCredentials) Read(*url.URL) (Credential, bool) {
	if mc.ShouldFail {
		return Credential{}, false
	}
	return Credential{Username: mc.Username, Password: mc.Password}, true
}

func (mc *mockFoundCredentials) Write(c Credential) error {
	mc.Username = c.Username
	mc.Password = c.Password
	return nil
}

type fakeWriteCloser struct {
	bytes.Buffer
}

func (f *fakeWriteCloser) Close() error {
	return nil
}

type badWriteCloser struct{ fakeWriteCloser }
type goodWriteCloser struct{ fakeWriteCloser }

func newMockCloser(err bool) io.WriteCloser {
	if err {
		return &badWriteCloser{}
	}
	return &goodWriteCloser{}
}

func (c *badWriteCloser) Write(p []byte) (int, error) {
	fmt.Println("called badWriteCloser")
	return 0, errors.New("write error")
}

func TestDockerAuth(t *testing.T) {
	ioImg, _ := reference.ParseNormalizedNamed("centos:7")
	customImg, _ := reference.ParseNormalizedNamed("fakeblock.com/gmaharis/block:0.0.0")
	ioCred := mockFoundCredentials{Credential: Credential{Username: "maeby", Password: "marryme!"}}
	customCred := mockFoundCredentials{Credential: Credential{Username: "gmaharis", Password: "hackertraps"}}
	tests := []struct {
		name   string
		image  reference.Named
		expect CredentialReaderWriter
	}{
		{"image with matching", ioImg, &ioCred},
		{"custom image with match", customImg, &customCred},
		{"image unfound credentials", ioImg, &mockFoundCredentials{ShouldFail: true}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			copy := *(test.expect.(*mockFoundCredentials))
			result := &copy

			dockerAuth(test.image, result)
			if !cmp.Equal(test.expect, result) {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}
