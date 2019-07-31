package workers

import "testing"

func TestPythonScript(t *testing.T) {
	p := Python{}
	script, err := p.script()
	if err != nil {
		t.Error(err)
	}
	if len(script) <= 0 {
		t.Errorf("python shell script expected non-zero length, but got %s", script)
	}
}
