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
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Artifact is the overall view of artifact
type Artifact struct {
	artifact.Artifact
	Tags             []*Tag                     // the list of tags that attached to the artifact
	SubResourceLinks map[string][]*ResourceLink // the resource link for build history(image), values.yaml(chart), dependency(chart), etc
	// TODO add other attrs: signature, scan result, etc
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
		ref := &models.Reference{
			ChildID:     reference.ChildID,
			ChildDigest: reference.ChildDigest,
			ParentID:    reference.ParentID,
		}
		if reference.Platform != nil {
			ref.Platform = &models.Platform{
				Architecture: reference.Platform.Architecture,
				Os:           reference.Platform.OS,
				OsFeatures:   reference.Platform.OSFeatures,
				OsVersion:    reference.Platform.OSVersion,
				Variant:      reference.Platform.Variant,
			}
		}
		art.References = append(art.References, ref)
	}
	for _, tag := range a.Tags {
		art.Tags = append(art.Tags, &models.Tag{
			ArtifactID:   tag.ArtifactID,
			ID:           tag.ID,
			Name:         tag.Name,
			PullTime:     strfmt.DateTime(tag.PullTime),
			PushTime:     strfmt.DateTime(tag.PushTime),
			RepositoryID: tag.RepositoryID,
		})
	}
	for resource, links := range a.SubResourceLinks {
		for _, link := range links {
			art.SubResourceLinks[resource] = []models.ResourceLink{}
			if link != nil {
				art.SubResourceLinks[resource] = append(art.SubResourceLinks[resource], models.ResourceLink{
					Absolute: link.Absolute,
					Href:     link.HREF,
				})
			}
		}
	}
	return art
}

// Tag is the overall view of tag
type Tag struct {
	tag.Tag
	// TODO add other attrs: signature, label, immutable status, etc
}

// Resource defines the specific resource of different artifacts: build history for image, values.yaml for chart, etc
type Resource struct {
	Content     []byte // the content of the resource
	ContentType string // the content type of the resource, returned as "Content-Type" header in API
}

// ResourceLink is a link via that a resource can be fetched
type ResourceLink struct {
	HREF     string
	Absolute bool // specify the href is an absolute URL or not
}

// Option is used to specify the properties returned when listing/getting artifacts
type Option struct {
	WithTag          bool
	TagOption        *TagOption // only works when WithTag is set to true
	WithLabel        bool
	WithScanOverview bool
	// TODO move it to TagOption?
	WithSignature bool
}

// TagOption is used to specify the properties returned when listing/getting tags
type TagOption struct {
	WithImmutableStatus bool
}

// TODO move this to GC controller?
// Option for pruning artifact records
// type Option struct {
//	 KeepUntagged bool // keep the untagged artifacts or not
// }
