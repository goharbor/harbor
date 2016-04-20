package util

import (
	"errors"
	"fmt"
	"strings"
)

// possible image uris are
//
// 1	host.com:1000/library/drone:v2.1
// 2	host.com:1000/library/drone
// 3	host.com/library/drone:v2.1
// 4	library/drone:v2.1
// 5	library/drone

var (
	ErrImageNotValid = errors.New("image uri not valid")
)

type ImageStruct struct {
	Repository string
	Host       string
	Port       string
	Version    string
	Namespace  string
}

func ParseImage(imageUri string) (*ImageStruct, error) {
	imageStruct := &ImageStruct{
		Version:   "latest",
		Namespace: "library",
		Host:      "index.docker.io",
		Port:      "443",
	}

	fields := strings.Split(imageUri, ":")
	if len(fields) == 1 {
		imageStruct.Repository = fields[0]
		tokens := strings.Split(fields[0], "/")
		if len(tokens) == 2 {
			imageStruct.Namespace = tokens[0]
			imageStruct.Repository = tokens[1]
		} else if len(tokens) == 3 {
			imageStruct.Host = tokens[0]
			imageStruct.Namespace = tokens[1]
			imageStruct.Repository = tokens[2]
		} else {
			return nil, ErrImageNotValid
		}
		return imageStruct, nil
	}

	if len(fields) == 2 {
		// host.com:1000/library/drone
		if strings.Contains(fields[1], "/") {
			tokens := strings.Split(fields[1], "/")
			if len(tokens) != 3 {
				return nil, ErrImageNotValid
			}
			imageStruct.Host = fields[0]
			imageStruct.Port = tokens[0]
			imageStruct.Namespace = tokens[1]
			imageStruct.Repository = tokens[2]
			return imageStruct, nil

		} else { // host.com/library/drone:v2.1 or library/drone:v2.1
			tokens := strings.Split(fields[0], "/")
			imageStruct.Version = fields[1]
			if len(tokens) == 2 {
				imageStruct.Namespace = tokens[0]
				imageStruct.Repository = tokens[1]
			}

			if len(tokens) == 3 {
				imageStruct.Host = tokens[0]
				imageStruct.Namespace = tokens[1]
				imageStruct.Repository = tokens[2]
			}

			return imageStruct, nil
		}
	}

	// host.com:1000/library/drone:v2.1
	if len(fields) == 3 {
		imageStruct.Host = fields[0]
		imageStruct.Version = fields[2]
		tokens := strings.Split(fields[1], "/")
		if len(tokens) != 3 {
			return nil, ErrImageNotValid
		}
		imageStruct.Port = tokens[0]
		imageStruct.Namespace = tokens[1]
		imageStruct.Repository = tokens[2]
		return imageStruct, nil
	}

	return nil, ErrImageNotValid
}

func (imageStruct *ImageStruct) ImageName() string {
	if len(imageStruct.Host) > 0 && imageStruct.Port != "443" {
		return fmt.Sprintf("%s:%s/%s/%s", imageStruct.Host, imageStruct.Port, imageStruct.Namespace, imageStruct.Repository)
	} else if len(imageStruct.Host) > 0 && imageStruct.Port == "443" {
		return fmt.Sprintf("%s/%s/%s", imageStruct.Host, imageStruct.Namespace, imageStruct.Repository)
	} else if len(imageStruct.Host) == 0 {
		return fmt.Sprintf("%s/%s", imageStruct.Host, imageStruct.Namespace, imageStruct.Repository)
	}
	return ErrImageNotValid.Error()
}
