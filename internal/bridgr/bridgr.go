// Package bridgr downloads artifacts based on a configuration file
package bridgr

import (
	"context"
	"io"
	"net/url"
	"os"
	"path"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/mitchellh/mapstructure"
)

// Verbose determines whether debug logging is printed
var Verbose = false

// Version is the built version of Bridgr
var Version = "development"

// DryRun holds whether workers should actually retrieve artifacts, or just do setup
var DryRun = false

// BaseDir gives the runtime absolute directory of the base "packages" directory
// See the individual repo type struct for the type-specific path
func BaseDir(repo string) string {
	var cwd, _ = os.Getwd()
	return path.Join(cwd, "packages", repo)
}

// Configuration is a type that unifies all of the sub-types of repository configurations
// It has a Hook function that returns
// a list of functions used for parsing its types with mapstructure decoding.
// Image is a configuration that can return a reference to a docker image that should be used
// for running its worker. This is a type that must be "batch" run in a docker image to create its
// repository
type Configuration interface {
	Name() string
	Image() reference.Named
	Hook() mapstructure.DecodeHookFunc
	Setup() error
	Run() error
}

// ImagePuller is a simplified interface to docker ImageAPIClient, that only defines the ImagePull function
type ImagePuller interface {
	ImagePull(ctx context.Context, ref string, options types.ImagePullOptions) (io.ReadCloser, error)
}

// PullImage is a helper function that Pulls a docker image to the local docker daemon
func PullImage(cli ImagePuller, image reference.Named) error {
	creds := &DockerCredential{}
	dockerAuth(image, creds)
	output, err := cli.ImagePull(context.Background(), image.String(), types.ImagePullOptions{RegistryAuth: creds.String()})
	defer output.Close() // nolint
	if err != nil {
		return err
	}

	_, err = io.Copy(writer(), output) // must wait for output before returning
	return err
}

func dockerAuth(image reference.Named, rw CredentialReaderWriter) {
	imgDomain := "https://" + reference.Domain(image) // by putting scheme in front, it forces url.Parse to correctly identify the host portion
	url, _ := url.Parse(imgDomain)

	if creds, ok := rw.Read(url); ok {
		Debugf("Docker: Found credentials for %s", url.Hostname())
		_ = rw.Write(creds)
	}
}
