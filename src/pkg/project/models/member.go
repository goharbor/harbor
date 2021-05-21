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
	"time"

	"github.com/goharbor/harbor/src/lib/orm"
)

func init() {
	orm.RegisterModel(
		new(Member),
	)
}

// Member holds the details of a member.
type Member struct {
	ID           int       `orm:"pk;auto;column(id)" json:"id"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	Role         int       `orm:"column(role)" json:"role_id"`
	EntityID     int       `orm:"column(entity_id)" json:"entity_id"`
	EntityType   string    `orm:"column(entity_type)" json:"entity_type"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// MemberQuery ...
type MemberQuery struct {
	UserID   int    // the user id
	Name     string // the username of member
	Role     int    // the role of the member has to the project
	GroupIDs []int  // the group ID of current user belongs to

	WithPublic bool // include the public projects for the member
}

// TableName ...
func (*Member) TableName() string {
	return "project_member"
}
