package bridgr

import (
	"os"
	"path"

	"github.com/docker/distribution/reference"
	"github.com/mitchellh/mapstructure"
	"helm.sh/helm/v3/pkg/repo"
)

type Helm []*FileItem

func (h Helm) dir() string {
	return BaseDir(h.Name())
}

func (h Helm) Name() string {
	return "helm"
}

func (h Helm) Image() reference.Named {
	return nil
}

func (h *Helm) Hook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(stringToFileItem)
}

func (h Helm) Setup() error {
	Print("Called Helm.Setup()")
	for _, chart := range h {
		chart.Normalize(h.dir())
	}
	return os.MkdirAll(h.dir(), os.ModePerm)
}

func (h Helm) Run() error {
	err := h.Setup()
	if err != nil {
		return err
	}

	for _, chart := range h {
		writer, createErr := os.Create(chart.Target)
		if createErr != nil {
			Printf("Unable to create local file '%s' (for %s) %s", chart.Target, chart.Source.String(), createErr)
			continue
		}
		if err := chart.fetch(&fileFetcher{}, &WorkerCredentialReader{}, writer); err != nil {
			Printf("Files '%s' - %+s", chart.Source.String(), err)
			_ = os.Remove(chart.Target)
		}
	}
	return h.createHelmIndex()
}

func (h Helm) createHelmIndex() error {
	helmIndex, err := repo.IndexDirectory(h.dir(), "/"+h.Name())
	if err != nil {
		return err
	}
	helmIndex.SortEntries()
	return helmIndex.WriteFile(path.Join(h.dir(), "index.yaml"), os.ModePerm)
}
