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
	"github.com/goharbor/harbor/src/common/models"
)

const (
	// SchemaVersion is the version of database schema
	SchemaVersion = "1.6.0"
)

// GetSchemaVersion return the version of database schema
func GetSchemaVersion() (*models.SchemaVersion, error) {
	version := &models.SchemaVersion{}
	if err := GetOrmer().Raw("select version_num from alembic_version").
		QueryRow(version); err != nil {
		return nil, err
	}
	return version, nil
}
