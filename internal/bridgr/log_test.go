package bridgr_test

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/aztechian/bridgr/internal/bridgr"
)

var logBuff bytes.Buffer

func setup(verbose bool) {
	logBuff.Reset()
	log.SetOutput(&logBuff)
	bridgr.Verbose = verbose
}

func TestLogDebugf(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   []interface{}
		verbose  bool
		expected string
	}{
		{"without debug", "thats why you %s %s", []interface{}{"always leave", "a note"}, false, "X"},
		{"verbose", "Bob %s", []interface{}{"Loblaw"}, true, "Bob Loblaw"},
	}

	for _, test := range tests {
		setup(test.verbose)

		t.Run(test.name, func(t *testing.T) {
			bridgr.Debugf(test.input, test.params...)
			if !test.verbose && len(logBuff.String()) > 0 {
				t.Errorf("negative test failed, got %s but was not expecting to", logBuff.String())
			}
			if test.verbose && !strings.Contains(logBuff.String(), test.expected) {
				t.Errorf("expected log to contain %s, got %s", test.expected, logBuff.String())
			}
		})
	}
}

func TestLogPrintf(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   []interface{}
		verbose  bool
		expected string
	}{
		{"without debug", "thats why you %s %s", []interface{}{"always leave", "a note"}, false, "thats why you always leave a note"},
		{"verbose", "Bob %s", []interface{}{"Loblaw"}, true, "Bob Loblaw"},
		{"verbose with simple string", "Theres always money in the banana stand", nil, true, "Theres always money in the banana stand"},
	}

	for _, test := range tests {
		setup(test.verbose)

		t.Run(test.name, func(t *testing.T) {
			bridgr.Printf(test.input, test.params...)
			if !strings.Contains(logBuff.String(), test.expected) {
				t.Errorf("expected log to contain %s, got %s", test.expected, logBuff.String())
			}
		})
	}
}

func TestLogPrintln(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		verbose  bool
		expected string
	}{
		{"without debug", "thats why you always leave a note", false, "thats why you always leave a note"},
		{"verbose", "Bob Loblaws Law Blog", true, "Bob Loblaws Law Blog"},
		{"verbose with simple string", "Theres always money in the banana stand", true, "Theres always money in the banana stand"},
	}

	for _, test := range tests {
		setup(test.verbose)

		t.Run(test.name, func(t *testing.T) {
			bridgr.Println(test.input)
			if !strings.Contains(logBuff.String(), test.expected) {
				t.Errorf("expected log to contain %s, got %s", test.expected, logBuff.String())
			}
		})
	}
}

func TestLogDebugln(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		verbose  bool
		expected string
	}{
		{"without debug", "thats why you always leave a note", false, "thats why you always leave a note"},
		{"verbose", "Bob Loblaws Law Blog", true, "Bob Loblaws Law Blog"},
		{"verbose with simple string", "Theres always money in the banana stand", true, "Theres always money in the banana stand"},
	}

	for _, test := range tests {
		setup(test.verbose)

		t.Run(test.name, func(t *testing.T) {
			bridgr.Debugln(test.input)
			if !test.verbose && len(logBuff.String()) > 0 {
				t.Errorf("negative test failed, got %s but was not expecting to", logBuff.String())
			}
			if test.verbose && !strings.Contains(logBuff.String(), test.expected) {
				t.Errorf("expected log to contain %s, got %s", test.expected, logBuff.String())
			}
		})
	}
}

func TestLogPrint(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		verbose  bool
		expected string
	}{
		{"without debug", map[string]string{"thats why": "you always", "leave": "a note"}, false, "leave:a note"},
		{"verbose", map[string]string{"Bob Loblaws": "Law Blog"}, true, "Bob Loblaws:Law Blog"},
		{"verbose with simple string", "Theres always money in the banana stand", true, "Theres always money in the banana stand"},
	}

	for _, test := range tests {
		setup(test.verbose)

		t.Run(test.name, func(t *testing.T) {
			bridgr.Print(test.input)
			if !strings.Contains(logBuff.String(), test.expected) {
				t.Errorf("expected log to contain %s, got %s", test.expected, logBuff.String())
			}
		})
	}
}

func TestLogDebug(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		verbose  bool
		expected string
	}{
		{"without debug", map[string]string{"thats why": "you always", "leave": "a note"}, false, "leave:a note"},
		{"verbose", map[string]string{"Bob Loblaws": "Law Blog"}, true, "Bob Loblaws:Law Blog"},
		{"verbose with simple string", "Theres always money in the banana stand", true, "Theres always money in the banana stand"},
	}

	for _, test := range tests {
		setup(test.verbose)

		t.Run(test.name, func(t *testing.T) {
			bridgr.Debug(test.input)
			if !test.verbose && len(logBuff.String()) > 0 {
				t.Errorf("negative test failed, got %s but was not expecting to", logBuff.String())
			}
			if test.verbose && !strings.Contains(logBuff.String(), test.expected) {
				t.Errorf("expected log to contain %s, got %s", test.expected, logBuff.String())
			}
		})
	}
}

func TestLog(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		params   []interface{}
		verbose  bool
		expected string
	}{
		{"without debug", "thats why you %s %s", []interface{}{"always leave", "a note"}, false, "thats why you always leave a note"},
		{"verbose", "Bob %s", []interface{}{"Loblaw"}, true, "Bob Loblaw"},
		{"verbose with simple string", "Theres always money in the banana stand", nil, true, "Theres always money in the banana stand"},
	}

	for _, test := range tests {
		logBuff.Reset()
		bridgr.Out = &logBuff
		bridgr.Verbose = test.verbose

		t.Run(test.name, func(t *testing.T) {
			bridgr.Log(test.input, test.params...)
			if test.verbose && !strings.Contains(logBuff.String(), test.expected) {
				t.Errorf("expected log to contain %s, got %s", test.expected, logBuff.String())
			}
			if !test.verbose && len(logBuff.String()) > 0 {
				t.Errorf("Log buffer should be empty, but is %s", logBuff.String())
			}
		})
	}
	bridgr.Out = os.Stdout
}
