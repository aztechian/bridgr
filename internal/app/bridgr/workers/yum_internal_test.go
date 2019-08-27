package workers

import (
	"strings"
	"testing"
)

func TestYumScript(t *testing.T) {
	y := Yum{}
	script, err := y.script([]string{"my", "test"})
	if err != nil {
		t.Error(err)
	}
	if len(script) <= 0 {
		t.Errorf("yum shell script expected non-zero length, but got %s", script)
	}
	if !strings.Contains(script, "test") {
		t.Errorf("Expected 'test' string to be in script output, but got %s", script)
	}
}
