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
	"github.com/astaxie/beego/orm"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/models"
	"time"
)

func init() {
	orm.RegisterModel(&Blob{})
}

// TODO: move ArtifactAndBlob, ProjectBlob to here

// ArtifactAndBlob alias ArtifactAndBlob model
type ArtifactAndBlob = models.ArtifactAndBlob

// Blob holds the details of a blob.
type Blob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Digest       string    `orm:"column(digest)" json:"digest"`
	ContentType  string    `orm:"column(content_type)" json:"content_type"`
	Size         int64     `orm:"column(size)" json:"size"`
	Status       string    `orm:"column(status)" json:"status"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now_add" json:"update_time"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName ...
func (b *Blob) TableName() string {
	return "blob"
}

// IsForeignLayer returns true if the blob is foreign layer
func (b *Blob) IsForeignLayer() bool {
	return b.ContentType == schema2.MediaTypeForeignLayer
}

// ProjectBlob alias ProjectBlob model
type ProjectBlob = models.ProjectBlob

// ListParams list params
type ListParams struct {
	ArtifactDigest  string   // list blobs which associated with the artifact
	ArtifactDigests []string // list blobs which associated with these artifacts
	BlobDigests     []string // list blobs which digest in the digests
	ProjectID       int64    // list blobs which associated with the project
}
