package config

import (
	"fmt"
	"path"
	"reflect"
	"strconv"

	"github.com/docker/distribution/reference"
)

// Docker struct is the configuration holder for the Docker worker type
type Docker struct {
	Destination string `mapstructure:"repository,omitempty"`
	Images      []reference.Named
}

// BaseDir is the top-level directory name for all objects written out under the Docker worker
func (d Docker) BaseDir() string {
	return path.Join(BaseDir(), "docker")
}

func (d Docker) Count() int {
	return len(d.Images)
}

func (d *Docker) Image() reference.Named {
	return nil
}

func parseDocker(imageObj map[string]interface{}) (string, error) {
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

func mapToImage(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
	if f.Kind() == reflect.Map && t == reflect.TypeOf((*reference.Named)(nil)).Elem() {
		imageStr, err := parseDocker(data.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
		return reference.ParseNormalizedNamed(imageStr)
	}
	return data, nil
}
