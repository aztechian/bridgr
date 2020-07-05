package bridgr

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestHelmDir(t *testing.T) {
	expected := BaseDir("helm")
	result := Helm{}.dir()
	if !cmp.Equal(expected, result) {
		t.Error(cmp.Diff(expected, result))
	}
}

func TestHelmIndex(t *testing.T) {
	helm := Helm{}
	helm.createHelmIndex()
}
