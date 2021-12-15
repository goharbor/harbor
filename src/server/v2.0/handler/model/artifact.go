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
	pkg_art "github.com/goharbor/harbor/src/pkg/artifact"
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
		Icon:              a.Icon,
		PullTime:          strfmt.DateTime(a.PullTime),
		PushTime:          strfmt.DateTime(a.PushTime),
		ExtraAttrs:        a.ExtraAttrs,
		Annotations:       a.Annotations,
	}

	for _, reference := range a.References {
		art.References = append(art.References, NewReference(reference).ToSwagger())
	}
	for _, acc := range a.Accessories {
		art.Accessories = append(art.Accessories, NewAccessory(acc.GetData()).ToSwagger())
	}
	for _, tag := range a.Tags {
		art.Tags = append(art.Tags, NewTag(tag).ToSwagger())
	}
	for addition, link := range a.AdditionLinks {
		if art.AdditionLinks == nil {
			art.AdditionLinks = make(map[string]models.AdditionLink)
		}
		art.AdditionLinks[addition] = NewAdditionLink(link).ToSwagger()
	}
	for _, label := range a.Labels {
		art.Labels = append(art.Labels, NewLabel(label).ToSwagger())
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

// AdditionLink is a link via that the addition can be fetched
type AdditionLink struct {
	*artifact.AdditionLink
}

// ToSwagger converts the addition link to the swagger model
func (a *AdditionLink) ToSwagger() models.AdditionLink {
	return models.AdditionLink{
		Absolute: a.Absolute,
		Href:     a.HREF,
	}
}

// NewAdditionLink ...
func NewAdditionLink(a *artifact.AdditionLink) *AdditionLink {
	return &AdditionLink{AdditionLink: a}
}

// Reference records the child artifact referenced by parent artifact
type Reference struct {
	*pkg_art.Reference
}

// ToSwagger converts the reference to the swagger model
func (r *Reference) ToSwagger() *models.Reference {
	ref := &models.Reference{
		ChildDigest: r.ChildDigest,
		ChildID:     r.ChildID,
		ParentID:    r.ParentID,
		Annotations: r.Annotations,
		Urls:        r.URLs,
	}
	if r.Platform != nil {
		ref.Platform = &models.Platform{
			Architecture: r.Platform.Architecture,
			Os:           r.Platform.OS,
			OsFeatures:   r.Platform.OSFeatures,
			OsVersion:    r.Platform.OSVersion,
			Variant:      r.Platform.Variant,
		}
	}
	return ref
}

// NewReference ...
func NewReference(r *pkg_art.Reference) *Reference {
	return &Reference{Reference: r}
}
