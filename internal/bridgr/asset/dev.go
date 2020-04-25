// +build !dist

package asset

import (
	"net/http"
	"path"
	"runtime"
)

var (
	_, b, _, _ = runtime.Caller(0)
	dir        = path.Dir(b)
)

// Templates is the development version of asset.Templates, which is an http Filesystem struct.
// This must match what is generated from "generate.go". See vfsgen package for details.
var Templates http.FileSystem = http.Dir(path.Join(dir, "templates"))
