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

import "time"

// CVEWhitelist defines the data model for a CVE whitelist
type CVEWhitelist struct {
	ID           int64              `orm:"pk;auto;column(id)" json:"id"`
	ProjectID    int64              `orm:"column(project_id)" json:"project_id"`
	ExpiresAt    *int64             `orm:"column(expires_at)" json:"expires_at,omitempty"`
	Items        []CVEWhitelistItem `orm:"-" json:"items"`
	ItemsText    string             `orm:"column(items)" json:"-"`
	CreationTime time.Time          `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time          `orm:"column(update_time);auto_now" json:"update_time"`
}

// CVEWhitelistItem defines one item in the CVE whitelist
type CVEWhitelistItem struct {
	CVEID string `json:"cve_id"`
}

// TableName ...
func (r *CVEWhitelist) TableName() string {
	return "cve_whitelist"
}
