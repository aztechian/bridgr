package bridgr_test

import (
	"net/url"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
	"github.com/google/go-cmp/cmp"
)

func TestFileImage(t *testing.T) {
	file := bridgr.File{}
	if file.Image() != nil {
		t.Errorf("expected nil, but got %+v", file.Image())
	}
}

func TestFileName(t *testing.T) {
	expected := "files"
	file := bridgr.File{}
	if !cmp.Equal(expected, file.Name()) {
		t.Error(cmp.Diff(expected, file.Name()))
	}
}

func TestFileHook(t *testing.T) {
	file := bridgr.File{}
	result := reflect.TypeOf(file.Hook())
	if strings.HasPrefix(result.Name(), "func(") {
		t.Error(cmp.Diff(result.Name(), reflect.Func))
	}
}

func TestFileNormalize(t *testing.T) {
	simpleSrc, _ := url.Parse("https://bluth.com/pub/plans/saddam.pdf")

	tests := []struct {
		name   string
		item   bridgr.FileItem
		expect string
	}{
		{"simple", bridgr.FileItem{Source: simpleSrc}, "saddam.pdf"},
		{"target dir", bridgr.FileItem{Source: simpleSrc, Target: "assets"}, "assets/saddam.pdf"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expect := path.Join(bridgr.BaseDir("files"), test.expect)
			result := test.item.Normalize(bridgr.BaseDir("files"))
			if !cmp.Equal(expect, result) {
				t.Error(cmp.Diff(expect, result))
			}

			second := test.item.Normalize(bridgr.BaseDir("files"))
			if !cmp.Equal(expect, second) {
				t.Error(cmp.Diff(expect, second))
			}
		})
	}
}

func TestFileItemString(t *testing.T) {
	simpleSrc, _ := url.Parse("https://bluth.com/pub/plans/saddam.pdf")
	item := bridgr.FileItem{Source: simpleSrc}

	if !cmp.Equal(item.Source.String(), item.String()) {
		t.Error(cmp.Diff(item.Source.String(), item.String()))
	}
}
