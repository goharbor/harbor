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

package migration

import (
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
)

const (
	// SchemaVersion identifies the schema version of RDB
	SchemaVersion = "1.8.1"
)

// VersionKey returns the key of redis schema
func VersionKey(ns string) string {
	return fmt.Sprintf("%s%s", rds.KeyNamespacePrefix(ns), "_schema_version")
}
