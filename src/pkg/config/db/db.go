//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package db

import (
	"context"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/encrypt"
	"github.com/goharbor/harbor/src/pkg/config/db/dao"
	"os"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/lib/log"
)

// Database - Used to load/save configuration in database
type Database struct {
	cfgDAO dao.DAO
}

// Load - load config from database, only user setting will be load from database.
func (d *Database) Load(ctx context.Context) (map[string]interface{}, error) {
	resultMap := map[string]interface{}{}
	configEntries, err := d.cfgDAO.GetConfigEntries(ctx)
	if err != nil {
		return resultMap, err
	}
	for _, item := range configEntries {

		itemMetadata, ok := metadata.Instance().GetByName(item.Key)
		if !ok {
			log.Debugf("failed to get metadata, key:%v, error:%v, skip to load item", item.Key, err)
			continue
		}
		if itemMetadata.Scope == metadata.SystemScope {
			continue
		}
		if _, ok := itemMetadata.ItemType.(*metadata.PasswordType); ok {
			if decryptPassword, err := encrypt.Instance().Decrypt(item.Value); err == nil {
				item.Value = decryptPassword
			} else {
				log.Errorf("decrypt password failed, key: %s, error %v,", item.Key, err)
			}
		}
		resultMap[itemMetadata.Name] = item.Value
	}
	return resultMap, nil
}

// Save - Only save user config items in the cfgs map
func (d *Database) Save(ctx context.Context, cfgs map[string]interface{}) error {
	var configEntries []models.ConfigEntry
	for key, value := range cfgs {
		if item, ok := metadata.Instance().GetByName(key); ok {
			if os.Getenv("UTTEST") != "true" && item.Scope == metadata.SystemScope {
				// skip to save system setting to db
				continue
			}
			strValue := utils.GetStrValueOfAnyType(value)
			entry := &models.ConfigEntry{Key: key, Value: strValue}
			if _, ok := item.ItemType.(*metadata.PasswordType); ok {
				if encryptPassword, err := encrypt.Instance().Encrypt(strValue); err == nil {
					entry.Value = encryptPassword
				} else {
					log.Errorf("encrypt password failed, error: %v", err)
				}
			}
			configEntries = append(configEntries, *entry)
		} else {
			log.Errorf("failed to get metadata, skip to save key:%v", key)
		}
	}
	return d.cfgDAO.SaveConfigEntries(ctx, configEntries)
}
