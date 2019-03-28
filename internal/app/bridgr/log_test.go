package bridgr_test

import (
	"bridgr/internal/app/bridgr"
	"testing"
)

func TestLog(t *testing.T) {
	err := bridgr.Log("Some string")
	if err != nil {
		t.Error("Error calling Log with simple string")
	}
}

func TestLogf(t *testing.T) {
	err := bridgr.Logf("Some formatted %s", "string")
	if err != nil {
		t.Error("Error calling Logf with formatted string")
	}
}
