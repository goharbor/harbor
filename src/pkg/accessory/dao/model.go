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
)

func init() {
	orm.RegisterModel(&Accessory{})
}

// Accessory model in database
type Accessory struct {
	ID                int64     `orm:"pk;auto;column(id)" json:"id"`
	ArtifactID        int64     `orm:"column(artifact_id)" json:"artifact_id"`
	SubjectArtifactID int64     `orm:"column(subject_artifact_id)" json:"subject_artifact_id"`
	Type              string    `orm:"column(type)" json:"type"`
	Size              int64     `orm:"column(size)" json:"size"`
	Digest            string    `orm:"column(digest)" json:"digest"`
	CreationTime      time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName for artifact reference
func (a *Accessory) TableName() string {
	return "artifact_accessory"
}
