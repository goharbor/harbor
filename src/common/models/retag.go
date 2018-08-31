// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

type RetagRequest struct {
	SrcImage  string  `json:"src_image"`
	DestImage string  `json:"dest_image"`
}

type Image struct {
	Project string
	Repo    string
	Tag     string
}

func ParseImage(image string) (*Image, error) {
	repo := strings.SplitN(image, "/", 2)
	if len(repo) < 2 {
		return nil, fmt.Errorf("unable to parse image from string: %s", image)
	}
	i := strings.SplitN(repo[1], ":", 2)
	res := &Image{
		Project: repo[0],
		Repo:      i[0],
	}
	if len(i) == 2 {
		res.Tag = i[1]
	}
	return res, nil
}