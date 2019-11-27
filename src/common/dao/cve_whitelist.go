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
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// CreateCVEWhitelist creates the CVE whitelist
func CreateCVEWhitelist(l models.CVEWhitelist) (int64, error) {
	o := GetOrmer()
	itemsBytes, _ := json.Marshal(l.Items)
	l.ItemsText = string(itemsBytes)
	return o.Insert(&l)
}

// UpdateCVEWhitelist Updates the vulnerability white list to DB
func UpdateCVEWhitelist(l models.CVEWhitelist) (int64, error) {
	o := GetOrmer()
	itemsBytes, _ := json.Marshal(l.Items)
	l.ItemsText = string(itemsBytes)
	id, err := o.InsertOrUpdate(&l, "project_id")
	return id, err
}

// GetCVEWhitelist Gets the CVE whitelist of the project based on the project ID in parameter
func GetCVEWhitelist(pid int64) (*models.CVEWhitelist, error) {
	o := GetOrmer()
	qs := o.QueryTable(&models.CVEWhitelist{})
	qs = qs.Filter("ProjectID", pid)
	r := []*models.CVEWhitelist{}
	_, err := qs.All(&r)
	if err != nil {
		return nil, fmt.Errorf("failed to get CVE whitelist for project %d, error: %v", pid, err)
	}
	if len(r) == 0 {
		return nil, nil
	} else if len(r) > 1 {
		log.Infof("Multiple CVE whitelists found for project %d, length: %d, returning first element.", pid, len(r))
	}
	items := []models.CVEWhitelistItem{}
	err = json.Unmarshal([]byte(r[0].ItemsText), &items)
	if err != nil {
		log.Errorf("Failed to decode item list, err: %v, text: %s", err, r[0].ItemsText)
		return nil, err
	}
	r[0].Items = items
	return r[0], nil
}
