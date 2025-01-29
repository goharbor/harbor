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

package artifact

import (
	"encoding/json"
	"fmt"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib/encode/repository"
	accessoryModel "github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/label/model"
)

// Artifact is the overall view of artifact
type Artifact struct {
	artifact.Artifact
	Tags          []*tag.Tag                 `json:"tags"`           // the list of tags that attached to the artifact
	AdditionLinks map[string]*AdditionLink   `json:"addition_links"` // the resource link for build history(image), values.yaml(chart), dependency(chart), etc
	Labels        []*model.Label             `json:"labels"`
	Accessories   []accessoryModel.Accessory `json:"-"`
}

// UnmarshalJSON to customize the accessories unmarshal
func (artifact *Artifact) UnmarshalJSON(data []byte) error {
	type Alias Artifact
	ali := &struct {
		*Alias
		AccessoryItems []interface{} `json:"accessories,omitempty"`
	}{
		Alias: (*Alias)(artifact),
	}

	if err := json.Unmarshal(data, &ali); err != nil {
		return err
	}

	if len(ali.AccessoryItems) > 0 {
		for _, item := range ali.AccessoryItems {
			data, err := json.Marshal(item)
			if err != nil {
				return err
			}
			acc, err := accessoryModel.ToAccessory(data)
			if err != nil {
				return err
			}
			artifact.Accessories = append(artifact.Accessories, acc)
		}
	}

	return nil
}

// SetAdditionLink set a addition link
func (artifact *Artifact) SetAdditionLink(addition, version string) {
	if artifact.AdditionLinks == nil {
		artifact.AdditionLinks = make(map[string]*AdditionLink)
	}

	projectName, repo := utils.ParseRepository(artifact.RepositoryName)
	// encode slash as %252F
	repo = repository.Encode(repo)
	href := fmt.Sprintf("/api/%s/projects/%s/repositories/%s/artifacts/%s/additions/%s", version, projectName, repo, artifact.Digest, addition)

	artifact.AdditionLinks[addition] = &AdditionLink{HREF: href, Absolute: false}
}

// SetSBOMAdditionLink set the link of SBOM addition
func (artifact *Artifact) SetSBOMAdditionLink(sbomDgst string, version string) {
	if artifact.AdditionLinks == nil {
		artifact.AdditionLinks = make(map[string]*AdditionLink)
	}
	addition := "sboms"
	projectName, repo := utils.ParseRepository(artifact.RepositoryName)
	// encode slash as %252F
	repo = repository.Encode(repo)
	href := fmt.Sprintf("/api/%s/projects/%s/repositories/%s/artifacts/%s/additions/sbom", version, projectName, repo, sbomDgst)

	artifact.AdditionLinks[addition] = &AdditionLink{HREF: href, Absolute: false}
}

// AbstractLabelNames abstracts the label names from the artifact.
func (artifact *Artifact) AbstractLabelNames() []string {
	var names []string
	for _, label := range artifact.Labels {
		names = append(names, label.Name)
	}

	return names
}

// AdditionLink is a link via that the addition can be fetched
type AdditionLink struct {
	HREF     string `json:"href"`
	Absolute bool   `json:"absolute"` // specify the href is an absolute URL or not
}

// Option is used to specify the properties returned when listing/getting artifacts
type Option struct {
	WithTag            bool
	TagOption          *tag.Option // only works when WithTag is set to true
	WithLabel          bool
	WithAccessory      bool
	LatestInRepository bool
}
