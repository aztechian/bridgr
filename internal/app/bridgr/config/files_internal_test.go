package config

import (
	"os"
	"path"
	"testing"
)

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
	}{
		{"/afile.zyx", FileItem{"/afile.zyx", path.Join(dir, "afile.zyx"), "file"}},
		{"my/file.abc", FileItem{"my/file.abc", path.Join(dir, "file.abc"), "file"}},
		{"file://some/file/toget.zip", FileItem{"file://some/file/toget.zip", path.Join(dir, "toget.zip"), "file"}},
		{"http://mysite.com/file.gz", FileItem{"http://mysite.com/file.gz", path.Join(dir, "file.gz"), "http"}},
		{"https://mysite.com/archive.tar", FileItem{"https://mysite.com/archive.tar", path.Join(dir, "archive.tar"), "https"}},
		{"blah://mysite.com/file.js", FileItem{"blah://mysite.com/file.js", path.Join(dir, "file.js"), "blah"}},
	}

	for _, test := range tests {
		t.Run(test.given, func(t *testing.T) {
			result := FileItem{}
			result.parseSimple(test.given)
			if result != test.expected {
				t.Errorf("Expected %+v from parseSimple(), got %+v", test.expected, result)
			}
		})
	}
}

func TestParseComplex(t *testing.T) {
	tests := []struct {
		given    map[interface{}]interface{}
		expected FileItem
	}{
		{map[interface{}]interface{}{"source": "/afile.zyx", "target": "/afile.zyx"}, FileItem{"/afile.zyx", path.Join(dir, "afile.zyx"), "file"}},
		{map[interface{}]interface{}{"source": "my/file.abc", "target": "file.xyz"}, FileItem{"my/file.abc", path.Join(dir, "file.xyz"), "file"}},
		{map[interface{}]interface{}{"source": "my/file.bac", "target": "myfolder/"}, FileItem{"my/file.bac", path.Join(dir, "myfolder", "file.bac"), "file"}},
		{map[interface{}]interface{}{"source": "file://some/file/toget.zip", "target": "file.toget.zip"}, FileItem{"file://some/file/toget.zip", path.Join(dir, "file.toget.zip"), "file"}},
		{map[interface{}]interface{}{"source": "http://mysite.com/file.gz", "target": "file.gz"}, FileItem{"http://mysite.com/file.gz", path.Join(dir, "file.gz"), "http"}},
		{map[interface{}]interface{}{"source": "https://mysite.com/archive.tar", "target": "archive.tgz"}, FileItem{"https://mysite.com/archive.tar", path.Join(dir, "archive.tgz"), "https"}},
		{map[interface{}]interface{}{"source": "blah://mysite.com/file.js", "target": "myfile.js"}, FileItem{"blah://mysite.com/file.js", path.Join(dir, "myfile.js"), "blah"}},
	}

	for _, test := range tests {
		t.Run(test.given["source"].(string), func(t *testing.T) {
			result := FileItem{}
			result.parseComplex(test.given)
			if result != test.expected {
				t.Errorf("Expected %s from parseComplex(), got %s", test.expected, result)
			}
		})
	}
}

func TestGetFileProtocol(t *testing.T) {
	tests := []struct {
		given    string
		expected string
	}{
		{"/afile.zyx", "file"},
		{"my/file.abc", "file"},
		{"file://some/file/toget.zip", "file"},
		{"http://mysite.com/file.gz", "http"},
		{"https://mysite.com/archive.tar", "https"},
		{"blah://mysite.com/file.js", "blah"},
	}

	for _, test := range tests {
		t.Run(test.given, func(t *testing.T) {
			result := getFileProtocol(test.given)
			if result != test.expected {
				t.Errorf("Expected %s from getFileProtocol(), got %s", test.expected, result)
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
