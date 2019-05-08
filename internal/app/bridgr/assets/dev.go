// +build !dist

package assets

import (
	"net/http"
	"os"
	"path"
)

var cwd, _ = os.Getwd()

// Templates is the development version of assets.Templates, which is an http Filesystem struct.
// This must match what is generated from "generate.go". See vfsgen package for details.
var Templates http.FileSystem = http.Dir(path.Join(cwd, "internal", "app", "bridgr", "assets", "templates"))
