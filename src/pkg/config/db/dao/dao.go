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

package dao

import (
	"context"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
)

// DAO the dao for configure items
type DAO interface {
	// GetConfigEntries get all configure items
	GetConfigEntries(ctx context.Context) ([]*models.ConfigEntry, error)
	// SaveConfigEntries save configure items provided
	SaveConfigEntries(ctx context.Context, entries []models.ConfigEntry) error
}

type dao struct {
}

// New ...
func New() DAO {
	return &dao{}
}

func (d *dao) GetConfigEntries(ctx context.Context) ([]*models.ConfigEntry, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
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

func (d *dao) SaveConfigEntries(ctx context.Context, entries []models.ConfigEntry) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Key == common.LDAPGroupAdminDn {
			entry.Value = utils.TrimLower(entry.Value)
		}
		tempEntry := models.ConfigEntry{}
		tempEntry.Key = entry.Key
		tempEntry.Value = entry.Value
		created, _, err := o.ReadOrCreate(&tempEntry, "k")
		if err != nil && !orm.IsDuplicateKeyError(err) {
			return errors.Wrap(err, "failed to create configuration entry")
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
