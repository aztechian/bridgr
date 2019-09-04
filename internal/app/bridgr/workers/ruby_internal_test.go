package workers

import "testing"

func TestRubyScript(t *testing.T) {
	r := Ruby{}
	script, err := r.script()
	if err != nil {
		t.Error(err)
	}
	if len(script) <= 0 {
		t.Errorf("ruby shell script expected non-zero length, but got %s", script)
	}
}
