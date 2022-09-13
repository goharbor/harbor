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
	"fmt"
	"time"

	"github.com/beego/beego/orm"
)

func init() {
	orm.RegisterModel(&ArtifactTrash{})
}

// ArtifactTrash records the deleted artifact
type ArtifactTrash struct {
	ID                int64     `orm:"pk;auto;column(id)"`
	MediaType         string    `orm:"column(media_type)"`
	ManifestMediaType string    `orm:"column(manifest_media_type)"`
	RepositoryName    string    `orm:"column(repository_name)"`
	Digest            string    `orm:"column(digest)"`
	CreationTime      time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName for artifact trash
func (at *ArtifactTrash) TableName() string {
	return "artifact_trash"
}

func (at *ArtifactTrash) String() string {
	return fmt.Sprintf("ID-%d MediaType-%s ManifestMediaType-%s RepositoryName-%s Digest-%s CreationTime-%s",
		at.ID, at.MediaType, at.ManifestMediaType, at.RepositoryName, at.Digest, at.CreationTime.Format("2006-01-02 15:04:05"))
}
