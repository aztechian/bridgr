package config

import (
	"net/url"
	"os"
	"path"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var rootFile, _ = url.Parse("/afile.xyz")
var relativeFile, _ = url.Parse("my/file.abc")
var file, _ = url.Parse("file://some/file/toget.zip")
var httpFile, _ = url.Parse("http://mysite.com/file.gz")
var httpsFile, _ = url.Parse("https://mysite.com/archive.tar")
var blahFile, _ = url.Parse("blah://mysite.com/file.js")
var cwd, _ = os.Getwd()
var dir = path.Join(cwd, "packages", "files")

func TestParseFiles(t *testing.T) {
	tests := []struct {
		name         string
		in           tempConfig
		numFiles     int
		simpleCalls  int
		complexCalls int
	}{
		{"relative file", tempConfig{Files: []interface{}{"some/file.xyz"}}, 1, 1, 0},
		{"absolute file", tempConfig{Files: []interface{}{"/some/file.xyz"}}, 1, 1, 0},
		{"complex file", tempConfig{Files: []interface{}{map[interface{}]interface{}{"source": "some/file.xyz", "target": "myfile.abc"}}}, 1, 1, 0},
		{"nil", tempConfig{Files: nil}, 0, 0, 0},
		{"non string", tempConfig{Files: []interface{}{2}}, 0, 0, 0},
		{"multiple entries", tempConfig{Files: []interface{}{"file1.zip", "file2.tar"}}, 2, 2, 0},
		{"error - bad url", tempConfig{Files: []interface{}{"\x7f"}}, 0, 0, 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := parseFiles(test.in)
			if len(result.Items) != test.numFiles {
				t.Errorf("Expected %d files in File struct, got %d", test.numFiles, len(result.Items))
			}
			//TODO test for spies on simple and complex calls
		})
	}
}

func TestParseSimple(t *testing.T) {
	tests := []struct {
		given    string
		expected FileItem
		isError  bool
	}{
		{rootFile.String(), FileItem{rootFile, path.Join(dir, "afile.xyz")}, false},
		{relativeFile.String(), FileItem{relativeFile, path.Join(dir, "file.abc")}, false},
		{file.String(), FileItem{file, path.Join(dir, "toget.zip")}, false},
		{httpFile.String(), FileItem{httpFile, path.Join(dir, "file.gz")}, false},
		{httpsFile.String(), FileItem{httpsFile, path.Join(dir, "archive.tar")}, false},
		{blahFile.String(), FileItem{blahFile, path.Join(dir, "file.js")}, false},
		{"\x7f", FileItem{}, true},
	}

	for _, test := range tests {
		t.Run(test.given, func(t *testing.T) {
			result := FileItem{}
			err := result.parseSimple(test.given)
			if err != nil && !test.isError {
				t.Error(err)
			}
			if !cmp.Equal(result, test.expected) {
				t.Errorf("unexpected result from parseSimple() %s", cmp.Diff(result, test.expected))
			}
		})
	}
}

func TestParseComplex(t *testing.T) {
	tests := []struct {
		given    map[interface{}]interface{}
		expected FileItem
		isError  bool
	}{
		{map[interface{}]interface{}{"source": rootFile.String(), "target": "/afile.zyx"}, FileItem{rootFile, path.Join(dir, "afile.zyx")}, false},
		{map[interface{}]interface{}{"source": relativeFile.String(), "target": "file.xyz"}, FileItem{relativeFile, path.Join(dir, "file.xyz")}, false},
		{map[interface{}]interface{}{"source": relativeFile.String(), "target": "myfolder/"}, FileItem{relativeFile, path.Join(dir, "myfolder", "file.abc")}, false},
		{map[interface{}]interface{}{"source": file.String(), "target": "file.toget.zip"}, FileItem{file, path.Join(dir, "file.toget.zip")}, false},
		{map[interface{}]interface{}{"source": httpFile.String(), "target": "file.gz"}, FileItem{httpFile, path.Join(dir, "file.gz")}, false},
		{map[interface{}]interface{}{"source": httpsFile.String(), "target": "archive.tgz"}, FileItem{httpsFile, path.Join(dir, "archive.tgz")}, false},
		{map[interface{}]interface{}{"source": blahFile.String(), "target": "myfile.js"}, FileItem{blahFile, path.Join(dir, "myfile.js")}, false},
		{map[interface{}]interface{}{"source": "\x7f", "target": "*shrug*"}, FileItem{}, true},
	}

	for _, test := range tests {
		t.Run(test.given["source"].(string), func(t *testing.T) {
			result := FileItem{}
			err := result.parseComplex(test.given)
			if err != nil && !test.isError {
				t.Error(err)
			}
			if !cmp.Equal(result, test.expected) {
				t.Errorf("unexpected result from parseComplex() %s", cmp.Diff(result, test.expected))
			}
		})
	}
}

func TestGetFileTarget(t *testing.T) {
	tests := []struct {
		given    string
		expected string
	}{
		{"afile.zyx", path.Join(dir, "afile.zyx")},
		{"afile", path.Join(dir, "afile")},
		{"/my/bfile", path.Join(dir, "bfile")},
		{"my/file.abc", path.Join(dir, "file.abc")},
		{"file://some/file/toget.zip", path.Join(dir, "toget.zip")},
	}

	for _, test := range tests {
		t.Run(test.given, func(t *testing.T) {
			result := getFileTarget(test.given)
			if result != test.expected {
				t.Errorf("Expected %s from getFileTarget(), got %s", test.expected, result)
			}
		})
	}
}
