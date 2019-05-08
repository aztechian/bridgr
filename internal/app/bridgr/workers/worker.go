package workers

import (
	"bridgr/internal/app/bridgr/assets"
	"io/ioutil"
)

// Worker is the interface for how to talk to all instances of worker structs
type Worker interface {
	Setup() error
	Run() error
}

func loadTemplate(name string) (string, error) {
	f, err := assets.Templates.Open(name)
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
