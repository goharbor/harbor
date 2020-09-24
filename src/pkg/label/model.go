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

package label

import "time"

// Reference is the reference of label and artifact
type Reference struct {
	ID           int64     `orm:"pk;auto;column(id)"`
	LabelID      int64     `orm:"column(label_id)"`
	ArtifactID   int64     `orm:"column(artifact_id)"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now"`
}

// TableName defines the database table name
func (r *Reference) TableName() string {
	return "label_reference"
}
