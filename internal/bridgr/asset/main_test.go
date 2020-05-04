package asset_test

import (
	"bytes"
	"io"
	"regexp"
	"testing"
	"text/template"

	"github.com/aztechian/bridgr/internal/bridgr/asset"
	"github.com/google/go-cmp/cmp"
)

const exampleTemplate = `And As It Is {{.}}, So Also As {{.}} Is It Unto You`

type myWriteCloser struct {
	io.Writer
}

func (mwc myWriteCloser) Close() error {
	return nil
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name     string
		load     string
		expected *regexp.Regexp
		isNull   bool
	}{
		{"negative test", "gob", regexp.MustCompile(`.*`), true},
		{"successful", "yum.sh", regexp.MustCompile(`createrepo`), false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := asset.Load(test.load)
			if test.isNull && err == nil {
				t.Error("expected an error, but got none")
			}
			if !test.expected.MatchString(result) {
				t.Errorf("results don't match %s", cmp.Diff(test.expected, result))
			}
		})
	}
}

func TestTemplate(t *testing.T) {
	result := asset.Template("blah")
	if result == nil {
		t.Error(cmp.Diff(result, nil))
	}
}

func TestRender(t *testing.T) {
	expected := "And As It Is Such, So Also As Such Is It Unto You"
	tmpl, _ := template.New("example").Parse(exampleTemplate)
	result := bytes.Buffer{}
	asset.Render(tmpl, "Such", &result)
	if !cmp.Equal(result.String(), expected) {
		t.Error(cmp.Diff(result.String(), expected))
	}
}

func TestRenderFile(t *testing.T) {
	expected := "And As It Is Such, So Also As Such Is It Unto You"
	tmpl, _ := template.New("example").Parse(exampleTemplate)
	buffer := bytes.Buffer{}
	result := myWriteCloser{&buffer}
	asset.RenderFile(tmpl, "Such", &result)
	if !cmp.Equal(buffer.String(), expected) {
		t.Error(cmp.Diff(buffer.String(), expected))
	}
}
