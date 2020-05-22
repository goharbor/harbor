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

package model

import (
	"encoding/json"

	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Artifact model
type Artifact struct {
	artifact.Artifact
	ScanOverview map[string]interface{} `json:"scan_overview"`
}

// ToSwagger converts the artifact to the swagger model
func (a *Artifact) ToSwagger() *models.Artifact {
	art := &models.Artifact{
		ID:                a.ID,
		Type:              a.Type,
		MediaType:         a.MediaType,
		ManifestMediaType: a.ManifestMediaType,
		ProjectID:         a.ProjectID,
		RepositoryID:      a.RepositoryID,
		Digest:            a.Digest,
		Size:              a.Size,
		PullTime:          strfmt.DateTime(a.PullTime),
		PushTime:          strfmt.DateTime(a.PushTime),
		ExtraAttrs:        a.ExtraAttrs,
		Annotations:       a.Annotations,
	}
	for _, reference := range a.References {
		art.References = append(art.References, reference.ToSwagger())
	}
	for _, tag := range a.Tags {
		art.Tags = append(art.Tags, tag.ToSwagger())
	}
	for addition, link := range a.AdditionLinks {
		if art.AdditionLinks == nil {
			art.AdditionLinks = make(map[string]models.AdditionLink)
		}
		art.AdditionLinks[addition] = link.ToSwagger()
	}
	for _, label := range a.Labels {
		art.Labels = append(art.Labels, label.ToSwagger())
	}
	if len(a.ScanOverview) > 0 {
		art.ScanOverview = models.ScanOverview{}
		for key, value := range a.ScanOverview {
			js, err := json.Marshal(value)
			if err != nil {
				log.Warningf("convert summary of %s failed, error: %v", key, err)
				continue
			}
			var summary models.NativeReportSummary
			if err := summary.UnmarshalBinary(js); err != nil {
				log.Warningf("convert summary of %s failed, error: %v", key, err)
				continue
			}

			art.ScanOverview[key] = summary
		}
	}
	return art
}
