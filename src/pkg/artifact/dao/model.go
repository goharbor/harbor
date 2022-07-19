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

package dao

import (
	"time"

	"github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

func init() {
	orm.RegisterModel(&Artifact{})
	orm.RegisterModel(&ArtifactReference{})
}

// Artifact model in database
type Artifact struct {
	ID                int64     `orm:"pk;auto;column(id)"`
	Type              string    `orm:"column(type)"`                // image or chart
	MediaType         string    `orm:"column(media_type)"`          // the media type of artifact
	ManifestMediaType string    `orm:"column(manifest_media_type)"` // the media type of manifest/index
	ProjectID         int64     `orm:"column(project_id)"`          // needed for quota
	RepositoryID      int64     `orm:"column(repository_id)"`
	RepositoryName    string    `orm:"column(repository_name)"`
	Digest            string    `orm:"column(digest)"`
	Size              int64     `orm:"column(size)"`
	Icon              string    `orm:"column(icon)"`
	PushTime          time.Time `orm:"column(push_time)"`
	PullTime          time.Time `orm:"column(pull_time)"`
	ExtraAttrs        string    `orm:"column(extra_attrs)"`             // json string
	Annotations       string    `orm:"column(annotations);type(jsonb)"` // json string
}

// TableName for artifact
func (a *Artifact) TableName() string {
	return "artifact"
}

// GetDefaultSorts specifies the default sorts
func (a *Artifact) GetDefaultSorts() []*q.Sort {
	return []*q.Sort{
		{
			Key:  "PushTime",
			DESC: true,
		},
		{
			Key:  "ID",
			DESC: true,
		},
	}
}

// ArtifactReference records the child artifact referenced by parent artifact
type ArtifactReference struct {
	ID          int64  `orm:"pk;auto;column(id)"`
	ParentID    int64  `orm:"column(parent_id)"`
	ChildID     int64  `orm:"column(child_id)"`
	ChildDigest string `orm:"column(child_digest)"`
	Platform    string `orm:"column(platform)"` // json string
	URLs        string `orm:"column(urls)"`     // json string
	Annotations string `orm:"column(annotations);type(jsonb)"`
}

// TableName for artifact reference
func (a *ArtifactReference) TableName() string {
	return "artifact_reference"
}
