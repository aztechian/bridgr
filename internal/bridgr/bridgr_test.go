package bridgr_test

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/stretchr/testify/mock"
)

type fakeCLI struct {
	mock.Mock
}

func (f *fakeCLI) ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error) {
	args := f.Called(ctx, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func TestBaseDir(t *testing.T) {
	v := bridgr.BaseDir("")
	expect, _ := os.Getwd()
	if len(v) == 0 {
		t.Error("BaseDir() has 0 length string")
	}
	if !strings.HasPrefix(v, expect) {
		t.Errorf("Expected BaseDir prefix of %s, but got %s", expect, v)
	}
	if v != path.Join(expect, "packages") {
		t.Errorf("Expected BaseDir to be %s, but got %s", path.Join(expect, "packages"), v)
	}
}

func TestPullImage(t *testing.T) {
	cli := fakeCLI{}
	img, _ := reference.ParseNormalizedNamed("nginx:2")
	cli.On("ImagePull", context.Background(), img.String(), types.ImagePullOptions{}).Return(ioutil.NopCloser(bytes.NewReader([]byte("hello world"))), nil)
	bridgr.PullImage(&cli, img)
}
