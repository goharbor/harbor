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
	"time"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/artifact/manager/dao"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

// Artifact is the abstract object managed by Harbor. It hides the
// underlying concrete detail and provides an unified artifact view
// for all users.
type Artifact struct {
	ID                int64
	Type              string // image, chart, etc
	MediaType         string // the media type of artifact. Mostly, it's the value of `manifest.config.mediatype`
	ManifestMediaType string // the media type of manifest/index
	Repository        *models.RepoRecord
	Tags              []*Tag // the list of tags that attached to the artifact
	Digest            string
	Size              int64
	PushTime          time.Time
	Platform          *v1.Platform               // when the parent of the artifact is an index, populate the platform information here
	ExtraAttrs        map[string]interface{}     // only contains the simple attributes specific for the different artifact type, most of them should come from the config layer
	SubResourceLinks  map[string][]*ResourceLink // the resource link for build history(image), values.yaml(chart), dependency(chart), etc
	Annotations       map[string]string
	References        []int64 // child artifacts referenced by the parent artifact if the artifact is an index
	Revision          string  // record data revision
	// TODO: As the labels and signature aren't handled inside the artifact module,
	// we should move it to the API level artifact model rather than
	// keeping it here. The same to scan information
	// Labels                  []*models.Label
	// Signature               *Signature                 // add the signature in the artifact level rather than tag level as we cannot make sure the signature always apply to tag
}

// From converts the database level artifact to the business level object
func (a *Artifact) From(art *dao.Artifact) {
	a.ID = art.ID
	a.Type = art.Type
	a.MediaType = art.MediaType
	a.ManifestMediaType = art.ManifestMediaType
	a.Repository = &models.RepoRecord{
		ProjectID:    art.ProjectID,
		RepositoryID: art.RepositoryID,
	}
	a.Digest = art.Digest
	a.Size = art.Size
	a.PushTime = art.PushTime
	a.ExtraAttrs = map[string]interface{}{}
	a.Annotations = map[string]string{}
	a.Revision = art.Revision
	if len(art.Platform) > 0 {
		if err := json.Unmarshal([]byte(art.Platform), &a.Platform); err != nil {
			log.Errorf("failed to unmarshal the platform of artifact %d: %v", art.ID, err)
		}
	}
	if len(art.ExtraAttrs) > 0 {
		if err := json.Unmarshal([]byte(art.ExtraAttrs), &a.ExtraAttrs); err != nil {
			log.Errorf("failed to unmarshal the extra attrs of artifact %d: %v", art.ID, err)
		}
	}
	if len(art.Annotations) > 0 {
		if err := json.Unmarshal([]byte(art.Annotations), &a.Annotations); err != nil {
			log.Errorf("failed to unmarshal the annotations of artifact %d: %v", art.ID, err)
		}
	}
}

// To converts the artifact to the database level object
func (a *Artifact) To() *dao.Artifact {
	art := &dao.Artifact{
		ID:                a.ID,
		Type:              a.Type,
		MediaType:         a.MediaType,
		ManifestMediaType: a.ManifestMediaType,
		ProjectID:         a.Repository.ProjectID,
		RepositoryID:      a.Repository.RepositoryID,
		Digest:            a.Digest,
		Size:              a.Size,
		PushTime:          a.PushTime,
		Revision:          a.Revision,
	}

	if a.Platform != nil {
		platform, err := json.Marshal(a.Platform)
		if err != nil {
			log.Errorf("failed to marshal the platform of artifact %d: %v", a.ID, err)
		}
		art.Platform = string(platform)
	}
	if len(a.ExtraAttrs) > 0 {
		attrs, err := json.Marshal(a.ExtraAttrs)
		if err != nil {
			log.Errorf("failed to marshal the extra attrs of artifact %d: %v", a.ID, err)
		}
		art.ExtraAttrs = string(attrs)
	}
	if len(a.Annotations) > 0 {
		annotations, err := json.Marshal(a.Annotations)
		if err != nil {
			log.Errorf("failed to marshal the annotations of artifact %d: %v", a.ID, err)
		}
		art.Annotations = string(annotations)
	}
	return art
}

// ResourceLink is a link via that a resource can be fetched
type ResourceLink struct {
	HREF     string
	Absolute bool // specify the href is an absolute URL or not
}

// TODO: move it to the API level artifact model
// Signature information
// type Signature struct {
// 	Signatures map[string]bool // tag: signed or not
// }

// Tag belongs to one repository and can only be attached to a single
// one artifact under the repository
type Tag struct {
	ID       int64
	Name     string
	PushTime time.Time
	PullTime time.Time
	Revision string // record data revision
}

// From converts the database level tag to the business level object
func (t *Tag) From(tag *dao.Tag) {
	t.ID = tag.ID
	t.Name = tag.Name
	t.PushTime = tag.PushTime
	t.PullTime = tag.PullTime
	t.Revision = tag.Revision
}

// To converts the tag to the database level model
func (t *Tag) To() *dao.Tag {
	return &dao.Tag{
		ID:       t.ID,
		Name:     t.Name,
		PushTime: t.PushTime,
		PullTime: t.PullTime,
		Revision: t.Revision,
	}
}
