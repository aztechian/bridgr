package bridgr

import (
	"os"
	"path"

	"github.com/distribution/reference"
	"github.com/mitchellh/mapstructure"
	"helm.sh/helm/v3/pkg/repo"
	log "unknwon.dev/clog/v2"
)

// Helm is a list of files that represent Helm charts, and have both a source and target
type Helm []*FileItem

func (h Helm) dir() string {
	return BaseDir(h.Name())
}

// Name returns the name string of this Helm worker
func (h Helm) Name() string {
	return "helm"
}

// Image returns the docker image used for collecting Helm charts in this worker. It is always nil, as the Helm worker doesn't use docker.
func (h Helm) Image() reference.Named {
	return nil
}

// Hook returns a list of DecodeHookFunc objects that are needed to decode a generic object into a Helm struct. It is used with mapstructure library.
func (h *Helm) Hook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(stringToFileItem)
}

// Setup prepares the Helm charts for fetching and indexing.
func (h Helm) Setup() error {
	log.Trace("Called Helm.Setup()")
	for _, chart := range h {
		chart.Normalize(h.dir())
	}
	return os.MkdirAll(h.dir(), DefaultDirPerms)
}

// Run downloads the requested Helm charts, and creates an index for statically hosting them
func (h Helm) Run() error {
	err := h.Setup()
	if err != nil {
		return err
	}

	for _, chart := range h {
		writer, createErr := os.Create(chart.Target)
		if createErr != nil {
			log.Warn("Unable to create local file '%s' (for %s) %s", chart.Target, chart.Source.String(), createErr)
			continue
		}
		if err := chart.fetch(&fileFetcher{}, &WorkerCredentialReader{}, writer); err != nil {
			log.Info("Files '%s' - %+s", chart.Source.String(), err)
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
