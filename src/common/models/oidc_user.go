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
)

// OIDCUser ...
type OIDCUser struct {
	ID     int64 `orm:"pk;auto;column(id)" json:"id"`
	UserID int   `orm:"column(user_id)" json:"user_id"`
	// encrypted secret
	Secret string `orm:"column(secret)" json:"-"`
	// secret in plain text
	PlainSecret  string    `orm:"-" json:"secret"`
	SubIss       string    `orm:"column(subiss)" json:"subiss"`
	Token        string    `orm:"column(token)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (o *OIDCUser) TableName() string {
	return "oidc_user"
}
