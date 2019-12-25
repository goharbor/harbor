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
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// AuthModeCanBeModified determines whether auth mode can be
// modified or not. Auth mode can modified when there is only admin
// user in database.
func AuthModeCanBeModified() (bool, error) {
	c, err := GetOrmer().QueryTable(&models.User{}).Count()
	if err != nil {
		return false, err
	}
	// admin and anonymous
	return c == 2, nil
}

// GetConfigEntries Get configuration from database
func GetConfigEntries() ([]*models.ConfigEntry, error) {
	o := GetOrmer()
	var p []*models.ConfigEntry
	sql := "select * from properties"
	n, err := o.Raw(sql, []interface{}{}).QueryRows(&p)

	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, nil
	}
	return p, nil
}

// SaveConfigEntries Save configuration to database.
func SaveConfigEntries(entries []models.ConfigEntry) error {
	o := GetOrmer()
	for _, entry := range entries {
		if entry.Key == common.LDAPGroupAdminDn {
			entry.Value = utils.TrimLower(entry.Value)
		}
		tempEntry := models.ConfigEntry{}
		tempEntry.Key = entry.Key
		tempEntry.Value = entry.Value
		created, _, err := o.ReadOrCreate(&tempEntry, "k")
		if err != nil && !IsDupRecErr(err) {
			log.Errorf("Error create configuration entry: %v", err)
			return err
		}
		if !created {
			entry.ID = tempEntry.ID
			_, err := o.Update(&entry, "v")
			if err != nil {
				return err
			}
		}
	}
	return nil
}
