package bridgr_test

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
)

func TestBaseDir(t *testing.T) {
	v := bridgr.BaseDir("")
	expect, _ := os.Getwd()
	if len(v) == 0 {
		t.Error("BaseDir() has 0 length string")
	}
	if !strings.HasPrefix(v, expect) {
		t.Errorf("Expected BaseDir prefix of %s, but got %s", expect, v)
	}
	if v != path.Join(expect, "packages") {
		t.Errorf("Expected BaseDir to be %s, but got %s", path.Join(expect, "packages"), v)
	}
}

func TestPullImage(t *testing.T) {

}
