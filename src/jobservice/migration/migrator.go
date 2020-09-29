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
	"github.com/gomodule/redigo/redis"
)

// RDBMigrator defines the action to migrate redis data
type RDBMigrator interface {
	// Metadata info of the migrator
	Metadata() *MigratorMeta

	// Migrate executes the real migration work
	Migrate() error
}

// MigratorMeta keeps the base info of the migrator
type MigratorMeta struct {
	FromVersion string
	ToVersion   string
	ObjectRef   string
}

// MigratorFactory is factory function to create RDBMigrator interface
type MigratorFactory func(pool *redis.Pool, namespace string) (RDBMigrator, error)
