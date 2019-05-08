// +build ignore

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/shurcooL/vfsgen"
)

func main() {
	var cwd, _ = os.Getwd()
	templates := http.Dir(filepath.Join(cwd, "templates"))
	if err := vfsgen.Generate(templates, vfsgen.Options{
		Filename:     "templates.go",
		PackageName:  "assets",
		BuildTags:    "dist",
		VariableName: "Templates",
	}); err != nil {
		log.Fatalln(err)
	}
}
