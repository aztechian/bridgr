package bridgr

import "strings"

// DockerImage takes an informal docker image string representation, and returns the canonicalized form of a Docker image specification string
func DockerImage(input string) string {
	var (
		repo    = ""
		image   = input
		version = "latest"
	)
	slash := strings.Index(input, "/")
	if slash >= 0 {
		tmp := strings.LastIndex(input, "/")
		image = input[tmp+1:]
		repo = input[:tmp]
	}
	if strings.Contains(image, ":") {
		version = strings.Split(image, ":")[1]
		image = strings.TrimSuffix(image, ":"+version)
	}
	if len(repo) <= 0 || repo == "_" {
		repo = "library"
	}
	return repo + "/" + image + ":" + version
}
