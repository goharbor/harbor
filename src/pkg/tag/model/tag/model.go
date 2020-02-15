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

package tag

import (
	"time"
)

// Tag model in database
type Tag struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	RepositoryID int64     `orm:"column(repository_id)" json:"repository_id"` // tags are the resources of repository, one repository only contains one same name tag
	ArtifactID   int64     `orm:"column(artifact_id)" json:"artifact_id"`     // the artifact ID that the tag attaches to, it changes when pushing a same name but different digest artifact
	Name         string    `orm:"column(name)" json:"name"`
	PushTime     time.Time `orm:"column(push_time)" json:"push_time"`
	PullTime     time.Time `orm:"column(pull_time)" json:"pull_time"`
}
