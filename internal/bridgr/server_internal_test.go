package bridgr

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func GetTestHandler() http.HandlerFunc {
	fn := func(rw http.ResponseWriter, req *http.Request) {
		fmt.Fprint(rw, "Hello")
	}
	return http.HandlerFunc(fn)
}

func TestCustomHeaders(t *testing.T) {
	ts := httptest.NewServer(customHeaders(GetTestHandler()))
	res, err := http.Get(ts.URL + "/blah ")
	if err != nil {
		t.Error(err)
	}

	server := res.Header.Get("Server")
	if server == "" {
		t.Error("Expected a Server header, but got none")
	}
	if !strings.Contains(server, "Bridgr") {
		t.Errorf("Expected Server header to contain Bridgr, but got %q", server)
	}
}

func TestLogMiddleware(t *testing.T) {
	ts := httptest.NewServer(logMiddleware(GetTestHandler()))
	res, err := http.Get(ts.URL + "/blah ")
	if err != nil {
		t.Error(err)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "Hello") {
		t.Error("logMiddleware handler did not call next")
	}
}
