//go:generate go run generate.go

// This file is just a placeholder to have the above go generate directive placed in the "asset" directory.

package asset

import (
	"io"
	"io/ioutil"
	"log"
	"strings"
	"text/template"
)

// Load reads a given asset name from the VFS, and returns it as a string
func Load(name string) (string, error) {
	f, err := Templates.Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// Template gets a template from the asset directory, and parses it into a *template.Template object, ready to be Execute'd
func Template(name string) *template.Template {
	tmpl, err := Load(name)
	if err != nil {
		log.Printf("Error loading %s template: %s", name, err)
	}
	return template.Must(template.New(name).Funcs(template.FuncMap{"Join": strings.Join}).Parse(tmpl))
}

// Render takes a template and renders it complete with data to the given output. This is useful if you want to render a template to a string.
// See RenderFile if you want a more convenient way to render to a file and have the output closed for you when complete.
func Render(tmpl *template.Template, data interface{}, output io.Writer) error {
	return tmpl.Execute(output, data)
}

// RenderFile renders a template with the given data to an output. As opposed to Render, this function will close the output when rendering is complete.
func RenderFile(tmpl *template.Template, data interface{}, output io.WriteCloser) error {
	defer output.Close()
	return Render(tmpl, data, output)
}
