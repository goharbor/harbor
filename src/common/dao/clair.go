// Copyright Project Harbor Authors
//
// licensed under the apache license, version 2.0 (the "license");
// you may not use this file except in compliance with the license.
// you may obtain a copy of the license at
//
//    http://www.apache.org/licenses/license-2.0
//
// unless required by applicable law or agreed to in writing, software
// distributed under the license is distributed on an "as is" basis,
// without warranties or conditions of any kind, either express or implied.
// see the license for the specific language governing permissions and
// limitations under the license.

package dao

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"

	"time"
)

// SetClairVulnTimestamp update the last_update of a namespace. If there's no record for this namespace, one will be created.
func SetClairVulnTimestamp(namespace string, timestamp time.Time) error {
	o := GetOrmer()
	rec := &models.ClairVulnTimestamp{
		Namespace:  namespace,
		LastUpdate: timestamp,
	}
	created, _, err := o.ReadOrCreate(rec, "Namespace")
	if err != nil {
		return err
	}
	if !created {
		rec.LastUpdate = timestamp
		n, err := o.Update(rec)
		if n == 0 {
			log.Warningf("no records are updated for %v", *rec)
		}
		return err
	}
	return nil
}

// ListClairVulnTimestamps return a list of all records in vuln timestamp table.
func ListClairVulnTimestamps() ([]*models.ClairVulnTimestamp, error) {
	var res []*models.ClairVulnTimestamp
	o := GetOrmer()
	_, err := o.QueryTable(models.ClairVulnTimestampTable).Limit(-1).All(&res)
	return res, err
}
