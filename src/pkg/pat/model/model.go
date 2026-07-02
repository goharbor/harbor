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
	"time"

	"github.com/beego/beego/v2/client/orm"
)

func init() {
	orm.RegisterModel(&PersonalAccessToken{})
}

// AccessLevel defines the access level for PAT scope
type AccessLevel struct {
	Resource string   `json:"resource"`
	Actions  []string `json:"actions"`
}

// ProjectScope defines the scope for a specific project
type ProjectScope struct {
	ProjectID   int64         `json:"project_id"`
	ProjectName string        `json:"project_name"`
	Access      []AccessLevel `json:"access"`
}

// PersonalAccessToken represents a personal access token for user authentication
type PersonalAccessToken struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	UserID       int       `orm:"column(user_id)" json:"user_id"`
	Name         string    `orm:"column(name)" json:"name" sort:"default"`
	Secret       string    `orm:"column(secret)" filter:"false" json:"-"`
	Salt         string    `orm:"column(salt)" filter:"false" json:"-"`
	Description  string    `orm:"column(description)" json:"description"`
	ExpiresAt    int64     `orm:"column(expires_at)" json:"expires_at"`
	LastUsedAt   int64     `orm:"column(last_used_at)" json:"last_used_at"`
	Disabled     bool      `orm:"column(disabled)" json:"disabled"`
	IsLegacy     bool      `orm:"column(is_legacy)" json:"is_legacy"`
	Scope        string    `orm:"column(scope)" json:"scope"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName returns the table name for PersonalAccessToken
func (p *PersonalAccessToken) TableName() string {
	return "personal_access_token"
}
