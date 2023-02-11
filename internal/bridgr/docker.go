package bridgr

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/distribution/reference"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/mitchellh/mapstructure"
	log "unknwon.dev/clog/v2"
)

var cli, _ = client.NewClientWithOpts(client.FromEnv)

// Docker struct is the configuration holder for the Docker worker type
type Docker struct {
	Destination string `mapstructure:"repository,omitempty"`
	Images      []reference.Named
}

// Image implements the Imager interface
func (d Docker) Image() reference.Named {
	return nil
}

func (d Docker) dir() string {
	return BaseDir(d.Name())
}

// Name returns the name of this Configuration
func (d Docker) Name() string {
	return "docker"
}

func mapToImage(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() == reflect.Map && t == reflect.TypeOf((*reference.Named)(nil)).Elem() {
		imageStr, err := dockerParse(data.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return reference.ParseNormalizedNamed(imageStr)
	}
	return data, nil
}

func arrayToDocker(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() != reflect.Slice || t != reflect.TypeOf(Docker{}) {
		return data, nil
	}

	var images []reference.Named
	for _, g := range data.([]interface{}) {
		if pkg, ok := g.(string); ok {
			img, err := reference.ParseNormalizedNamed(pkg)
			if err != nil {
				continue
			}
			images = append(images, img)
		}
	}

	return Docker{
		Images: images,
	}, nil
}

// Hook implements the Parser interface, returns a function for use by mapstructure when parsing config files
func (d *Docker) Hook() mapstructure.DecodeHookFunc {
	return mapstructure.ComposeDecodeHookFunc(
		arrayToDocker,
		mapToImage,
	)
}

func dockerParse(imageObj map[string]interface{}) (string, error) {
	i := ""
	if image, ok := imageObj["image"]; ok {
		i = i + image.(string)
	} else {
		return "", fmt.Errorf("Docker image configuration in map format must contain an `image` key: %v", imageObj)
	}

	if host, ok := imageObj["host"]; ok {
		i = host.(string) + "/" + i
	}
	if version, ok := imageObj["version"]; ok {
		switch v := version.(type) {
		case string:
			i = i + ":" + v
		case int:
			i = i + ":" + strconv.Itoa(v)
		default:
			return "", fmt.Errorf("unable to convert `version` field of Docker entry %s, please enclose the value in quotes", imageObj)
		}
	}
	return i, nil
}

// Run executes the Docker worker to fetch artifacts
func (d *Docker) Run() error {
	setupErr := d.Setup()
	if setupErr != nil {
		return setupErr
	}
	for _, img := range d.Images {
		if d.Destination != "" {
			dest := d.tagForRemote(cli, img)
			err := d.writeRemote(cli, dest, img)
			if err != nil {
				log.Info(err.Error())
			}
		} else {
			re := regexp.MustCompile(`[:/]`)
			outFile := re.ReplaceAllString(reference.Path(img), "_") + ".tar"
			out, err := os.Create(path.Join(d.dir(), outFile))
			if err != nil {
				log.Info("error creating %s for saving Docker image %s - %s", outFile, img.String(), err)
				continue
			}
			err = d.writeLocal(cli, out, img)
			if err != nil {
				log.Info("error saving %s - %s", img.String(), err)
				os.Remove(out.Name())
				continue
			}
			log.Trace("saved Docker image %s to %s", img.String(), out.Name())
		}
	}
	return nil
}

// Setup gets the environment ready to run the Docker worker
func (d *Docker) Setup() error {
	log.Trace("Called Docker.Setup()")
	_ = os.MkdirAll(d.dir(), os.ModePerm)

	// filter nil images from parse errors
	filtered := d.Images[:0]
	for _, img := range d.Images {
		if img == nil {
			continue
		}
		filtered = append(filtered, img)
		log.Trace("pulling image %s", img.String())
		if err := PullImage(cli, img); err != nil {
			log.Error("Error pulling Docker image `%s`: %s", img.String(), err)
		}
	}
	d.Images = filtered
	return nil
}

type imageSaver interface {
	ImageSave(ctx context.Context, images []string) (io.ReadCloser, error)
}

func (d *Docker) writeLocal(cli imageSaver, out io.WriteCloser, in reference.Named) error {
	ctx := context.Background()
	reader, err := cli.ImageSave(ctx, []string{in.String()})
	if err != nil {
		return err
	}
	defer reader.Close()
	defer out.Close()
	_, err = io.Copy(out, reader)
	return err
}

type imagePusher interface {
	ImagePush(ctx context.Context, ref string, options types.ImagePushOptions) (io.ReadCloser, error)
}

func (d *Docker) writeRemote(cli imagePusher, remote string, in reference.Named) error {
	writer := ioutil.Discard
	if Verbose {
		writer = os.Stderr
	}
	output, err := cli.ImagePush(context.Background(), remote, types.ImagePushOptions{})
	if err != nil {
		return err
	}
	defer output.Close()
	_, err = io.Copy(writer, output)
	return err
}

type imageTagger interface {
	ImageTag(ctx context.Context, image, ref string) error
}

func (d *Docker) tagForRemote(cli imageTagger, local reference.Named) string {
	destReg := d.Destination
	remoteTag := strings.Replace(local.String(), reference.Domain(local), destReg, -1)

	_ = cli.ImageTag(context.Background(), local.String(), remoteTag)
	return remoteTag
}
