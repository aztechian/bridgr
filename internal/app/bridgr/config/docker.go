package config

import (
	"path"

	"github.com/docker/distribution/reference"
)

// Docker struct is the configuration holder for the Docker worker type
type Docker struct {
	Destination string
	Items       []reference.Named
}

// BaseDir is the top-level directory name for all objects written out under the Docker worker
func (d *Docker) BaseDir() string {
	return path.Join(BaseDir(), "docker")
}

// Parse the top-level Docker section of the YAML config file
// parseDocker should only determine the next level of the config - mainly:
//  is this just a simple array of strings?
//  is this an array of objects?
//  is this a map of objects? (ie, "images" and/or "repository" were specified)
// func parseDocker(conf tempConfig) Docker {
// 	d := Docker{}
// 	switch cfgBlock := conf.Docker.(type) {
// 	case []interface{}: // this matches string and complex image spec
// 		err := d.parseItems(cfgBlock)
// 		if err != nil {
// 			log.Println(err)
// 		}
// 	case map[interface{}]interface{}:
// 		for key, block := range cfgBlock {
// 			if key.(string) == "repository" {
// 				d.Destination = block.(string)
// 			}
// 			if key.(string) == "images" {
// 				err := d.parseItems(block.([]interface{}))
// 				if err != nil {
// 					log.Println(err)
// 				}
// 			}
// 		}
// 	case nil:

// 	default:
// 		log.Printf("DEBUG: Unknown configuration section for Docker: %+v", cfgBlock)
// 	}
// 	bridgr.Debugf("Final Docker configuration %+v", d)
// 	return d
// }

// func (d *Docker) parseItems(imageList []interface{}) error {
// 	for _, imageObj := range imageList {
// 		switch imgSpec := imageObj.(type) {
// 		case string:
// 			err := d.parseSimple(imgSpec)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 		case map[interface{}]interface{}:
// 			err := d.parseComplex(imgSpec)
// 			if err != nil {
// 				log.Println(err)
// 			}
// 		default:
// 			log.Printf("Got unknown set of items: %+v \n %T", imageObj, imageObj)
// 		}
// 	}
// 	return nil
// }

// func (d *Docker) parseSimple(s string) error {
// 	named, err := reference.ParseNormalizedNamed(s)
// 	if err != nil {
// 		return err
// 	}
// 	d.Items = append(d.Items, named)
// 	return nil
// }

// func (d *Docker) parseComplex(imageObj map[interface{}]interface{}) error {
// 	i := ""
// 	if image, ok := imageObj["image"]; ok {
// 		i = i + image.(string)
// 	} else {
// 		return fmt.Errorf("Docker image configuration in map format must contain an `image` key: %v", imageObj)
// 	}

// 	if host, ok := imageObj["host"]; ok {
// 		i = host.(string) + "/" + i
// 	}
// 	if version, ok := imageObj["version"]; ok {
// 		switch v := version.(type) {
// 		case string:
// 			i = i + ":" + v
// 		case int:
// 			i = i + ":" + strconv.Itoa(v)
// 		default:
// 			return fmt.Errorf("unable to convert `version` field of Docker entry %s, please enclose the value in quotes", imageObj)
// 		}
// 	}
// 	return d.parseSimple(i)
// }
