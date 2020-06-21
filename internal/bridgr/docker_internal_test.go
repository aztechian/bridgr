package bridgr

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/mock"
)

var (
	namedComparer = cmp.Comparer(func(got, want reference.Named) bool {
		return got.String() == want.String()
	})

	defaultImg, _   = reference.ParseNormalizedNamed("jade:dragon-triad")
	defaultResponse = []byte("Stan Sitwell")
	defaultError    = fmt.Errorf("Gene Parmesean")
)

func dockerMust(ref reference.Named, err error) reference.Named {
	if err != nil {
		panic(err)
	}
	return ref
}

type dockMock struct {
	mock.Mock
}

func (d *dockMock) ImagePush(ctx context.Context, ref string, options types.ImagePushOptions) (io.ReadCloser, error) {
	args := d.Called(ctx, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (d *dockMock) ImageTag(ctx context.Context, image, ref string) error {
	args := d.Called(ctx, image, ref)
	return args.Error(0)
}

func (d *dockMock) ImageSave(ctx context.Context, images []string) (io.ReadCloser, error) {
	args := d.Called(ctx, images)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func TestArrayToDocker(t *testing.T) {
	imageSrc := []interface{}{"cinco:5.4", "norman-md", "bluth.com/cinco/cuatro:latest"}
	images := []reference.Named{dockerMust(reference.ParseNormalizedNamed("cinco:5.4")), dockerMust(reference.ParseNormalizedNamed("norman-md")), dockerMust(reference.ParseNormalizedNamed("bluth.com/cinco/cuatro:latest"))}
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid type", reflect.TypeOf(""), "", false, ""},
		{"valid", reflect.TypeOf(Docker{}), imageSrc, false, Docker{Images: images}},
		{"error parsing", reflect.TypeOf(Docker{}), []interface{}{`\/.`}, false, Docker{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := arrayToDocker(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(test.expect, result, namedComparer) && !test.isError {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestDockerMapToImage(t *testing.T) {
	img, _ := reference.ParseNormalizedNamed("cinco:cuatro")
	tests := []struct {
		name    string
		target  reflect.Type
		input   interface{}
		isError bool
		expect  interface{}
	}{
		{"invalid type", reflect.TypeOf(""), "", false, ""},
		{"valid", reflect.TypeOf(&img).Elem(), map[interface{}]interface{}{"image": "cinco", "version": "cuatro"}, false, img},
		{"parse error", reflect.TypeOf(&img).Elem(), map[interface{}]interface{}{"version": "fakeblock"}, true, img},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := mapToImage(reflect.TypeOf(test.input), test.target, test.input)
			if test.isError && err != nil {
				return
			}
			if err != nil {
				t.Error(err)
			}
			if !cmp.Equal(test.expect, result, namedComparer) && !test.isError {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestDockerParse(t *testing.T) {
	tests := []struct {
		name   string
		input  map[interface{}]interface{}
		expect string
	}{
		{"happy path", map[interface{}]interface{}{"image": "jade"}, "jade"},
		{"error missing image", map[interface{}]interface{}{"version": "triad"}, ""},
		{"host", map[interface{}]interface{}{"image": "dragon/jade", "host": "repo.lite"}, "repo.lite/dragon/jade"},
		{"host image and version", map[interface{}]interface{}{"image": "dragon/jade", "host": "repo.lite", "version": "triad"}, "repo.lite/dragon/jade:triad"},
		{"error float version", map[interface{}]interface{}{"image": "jade", "version": 1.4}, ""},
		{"int version", map[interface{}]interface{}{"image": "jade", "version": 12}, "jade:12"},
		{"stringified version", map[interface{}]interface{}{"image": "jade", "version": "1.4"}, "jade:1.4"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := dockerParse(test.input)
			if strings.Contains(test.name, "error") && err == nil {
				t.Error(err)
			}
			if !cmp.Equal(test.expect, result) {
				t.Error(cmp.Diff(test.expect, result))
			}
		})
	}
}

func TestDockerDir(t *testing.T) {
	expected := BaseDir("docker")
	result := Docker{}.dir()
	if !cmp.Equal(expected, result) {
		t.Error(cmp.Diff(expected, result))
	}
}

func TestDockerWriteRemote(t *testing.T) {
	tests := []struct {
		name    string
		verbose bool
	}{
		{"succcess quiet", false},
		{"success verbose", true},
		{"error quiet", false},
		{"error verbose", true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			Verbose = test.verbose
			docker := Docker{}
			cli := dockMock{}
			if strings.Contains(test.name, "error") {
				cli.On("ImagePush", context.Background(), defaultImg.String(), mock.Anything).Return(ioutil.NopCloser(bytes.NewReader([]byte{})), defaultError)
			} else {
				cli.On("ImagePush", context.Background(), defaultImg.String(), mock.Anything).Return(ioutil.NopCloser(bytes.NewReader(defaultResponse)), nil)
			}
			err := docker.writeRemote(&cli, defaultImg.String(), defaultImg)
			if err == nil && strings.Contains(test.name, "error") {
				t.Error(err)
			}
			cli.AssertNotCalled(t, "ImageSave")
		})
	}
}

func TestDockerWriteLocal(t *testing.T) {
	tests := []struct {
		name   string
		writer io.WriteCloser
	}{
		{"success", newMockCloser(false)},
		{"docker error", newMockCloser(false)},
		{"writer error", newMockCloser(true)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			docker := Docker{}
			cli := dockMock{}
			if strings.Contains(test.name, "error") {
				cli.On("ImageSave", mock.Anything, []string{defaultImg.String()}).Return(ioutil.NopCloser(bytes.NewReader([]byte{})), defaultError)
			} else {
				cli.On("ImageSave", mock.Anything, []string{defaultImg.String()}).Return(ioutil.NopCloser(bytes.NewReader(defaultResponse)), nil)
			}

			err := docker.writeLocal(&cli, test.writer, defaultImg)
			if err == nil && strings.Contains(test.name, "error") {
				t.Error("Expected an error condition from writeLocal() but got none")
			}
		})
	}
}

func TestDockerTagForRemote(t *testing.T) {
	customImg, _ := reference.ParseNormalizedNamed("myproj/myimage:1.0")
	tests := []struct {
		name  string
		image reference.Named
		want  string
	}{
		{"succcess", defaultImg, "repo.lite/library/jade:dragon-triad"},
		{"full path", customImg, "repo.lite/myproj/myimage:1.0"},
		{"error", defaultImg, "repo.lite/library/jade:dragon-triad"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := Docker{Destination: "repo.lite"}
			cli := dockMock{}
			if strings.Contains(test.name, "error") {
				cli.On("ImageTag", context.Background(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(fmt.Errorf("tag error"))
			} else {
				cli.On("ImageTag", context.Background(), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)
			}
			result := d.tagForRemote(&cli, test.image)

			if !cmp.Equal(test.want, result) {
				t.Error(cmp.Diff(test.want, result))
			}
		})
	}
}
