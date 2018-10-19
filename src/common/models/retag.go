// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package models

import (
	"fmt"
	"strings"
)

// RetagRequest gives the source image and target image of retag
type RetagRequest struct {
	Tag      string `json:"tag"`       // The new tag
	SrcImage string `json:"src_image"` // Source images in format <project>/<repo>:<reference>
	Override bool   `json:"override"`  // If target tag exists, whether override it
}

// Image holds each part (project, repo, tag) of an image name
type Image struct {
	Project string
	Repo    string
	Tag     string
}

// ParseImage parses an image name such as 'library/app:v1.0' to a structure with
// project, repo, and tag fields
func ParseImage(image string) (*Image, error) {
	repo := strings.SplitN(image, "/", 2)
	if len(repo) < 2 {
		return nil, fmt.Errorf("Unable to parse image from string: %s", image)
	}
	i := strings.SplitN(repo[1], ":", 2)
	res := &Image{
		Project: repo[0],
		Repo:    i[0],
	}
	if len(i) == 2 {
		res.Tag = i[1]
	}
	return res, nil
}
