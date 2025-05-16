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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func init() {
	orm.RegisterModel(&Blob{})
	orm.RegisterModel(&ArtifactAndBlob{})
	orm.RegisterModel(&ProjectBlob{})
}

/*
the status are used for Garbage Collection
StatusNone, the blob is using in Harbor as normal.
StatusDelete, the blob is marked as GC candidate.
StatusDeleting, the blob undergo a GC blob deletion.
StatusDeleteFailed, the blob is failed to delete from the backend storage.

The status transitions
StatusNone -> StatusDelete : Mark the blob as candidate.
StatusDelete -> StatusDeleting : Select the blob and call the API to delete asset in the backend storage.
StatusDeleting -> Trash : Delete success from the backend storage.
StatusDelete -> StatusNone : Client asks the existence of blob, remove it from the candidate.
StatusDelete -> StatusDeleteFailed : The storage driver returns fail when to delete the real data from the configurated file system.
StatusDelete -> StatusDelete : Encounter failure in the GC sweep phase. When to rerun the GC job, all of blob candidates are marked as StatusDelete again.
StatusDeleteFailed -> StatusNone : The delete failed blobs can be pushed again, and back to normal.
StatusDeleteFailed -> StatusDelete : The delete failed blobs should be in the candidate.
*/
const (
	StatusNone         = "none"
	StatusDelete       = "delete"
	StatusDeleting     = "deleting"
	StatusDeleteFailed = "deletefailed"
)

// StatusMap key is the target status, values are the accepted source status.
// For example, only StatusDelete can be convert to StatusDeleting.
var StatusMap = map[string][]string{
	StatusNone:         {StatusNone, StatusDelete, StatusDeleteFailed},
	StatusDelete:       {StatusNone, StatusDelete, StatusDeleteFailed},
	StatusDeleting:     {StatusDelete},
	StatusDeleteFailed: {StatusDeleting},
}

// ArtifactAndBlob holds the relationship between manifest and blob.
type ArtifactAndBlob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	DigestAF     string    `orm:"column(digest_af)" json:"digest_af"`
	DigestBlob   string    `orm:"column(digest_blob)" json:"digest_blob"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName ...
func (afb *ArtifactAndBlob) TableName() string {
	return "artifact_blob"
}

// ProjectBlob holds the relationship between manifest and blob.
type ProjectBlob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	BlobID       int64     `orm:"column(blob_id)" json:"blob_id"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName ...
func (*ProjectBlob) TableName() string {
	return "project_blob"
}

// Blob holds the details of a blob.
type Blob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Digest       string    `orm:"column(digest)" json:"digest"`
	ContentType  string    `orm:"column(content_type)" json:"content_type"`
	Size         int64     `orm:"column(size)" json:"size"`
	Status       string    `orm:"column(status)" json:"status"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now_add" json:"update_time"`
	Version      int64     `orm:"column(version)" json:"version"`
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

// IsManifest returns true if the blob is manifest layer
func (b *Blob) IsManifest() bool {
	return b.ContentType == schema2.MediaTypeManifest ||
		b.ContentType == schema1.MediaTypeManifest || b.ContentType == schema1.MediaTypeSignedManifest ||
		b.ContentType == v1.MediaTypeImageManifest || b.ContentType == v1.MediaTypeImageIndex ||
		b.ContentType == manifestlist.MediaTypeManifestList
}

// FilterByArtifactDigest returns orm.QuerySeter with artifact digest filter
func (b *Blob) FilterByArtifactDigest(_ context.Context, qs orm.QuerySeter, _ string, value any) orm.QuerySeter {
	v, ok := value.(string)
	if !ok {
		return qs
	}
	sql := fmt.Sprintf("IN (SELECT digest_blob FROM artifact_blob WHERE digest_af IN (%s))", `'`+v+`'`)
	return qs.FilterRaw("digest", sql)
}

// FilterByArtifactDigests returns orm.QuerySeter with artifact digests filter
func (b *Blob) FilterByArtifactDigests(_ context.Context, qs orm.QuerySeter, _ string, value any) orm.QuerySeter {
	artifactDigests, ok := value.([]string)
	if !ok {
		return qs
	}
	var afs []string
	for _, v := range artifactDigests {
		afs = append(afs, `'`+v+`'`)
	}

	sql := fmt.Sprintf("IN (SELECT digest_blob FROM artifact_blob WHERE digest_af IN (%s))", strings.Join(afs, ","))
	return qs.FilterRaw("digest", sql)
}

// FilterByProjectID returns orm.QuerySeter with project id filter
func (b *Blob) FilterByProjectID(_ context.Context, qs orm.QuerySeter, _ string, value any) orm.QuerySeter {
	projectID, ok := value.(int64)
	if !ok {
		return qs
	}

	return qs.FilterRaw("id", fmt.Sprintf("IN (SELECT blob_id FROM project_blob WHERE project_id = %d)", projectID))
}
