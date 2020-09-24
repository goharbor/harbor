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
	"github.com/Masterminds/semver"
	"reflect"

	"github.com/gomodule/redigo/redis"

	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
)

// Manager for managing the related migrators
type Manager interface {
	// Register the specified migrator to the execution chain
	Register(migratorFactory MigratorFactory)

	// Migrate data
	Migrate() error
}

// MigratorChainNode is a wrapper to append the migrator to the chain with a next reference
type MigratorChainNode struct {
	// Migrator implementation
	migrator RDBMigrator
	// Refer the next migration of the chain if existing
	next *MigratorChainNode
}

// BasicManager is the default implementation of manager interface
type BasicManager struct {
	// The head of migrator chain
	head *MigratorChainNode
	// Pool for connecting to redis
	pool *redis.Pool
	// RDB namespace
	namespace string
}

// New a basic manager
func New(pool *redis.Pool, ns string) Manager {
	return &BasicManager{
		pool:      pool,
		namespace: ns,
	}
}

// Register the migrator to the chain
func (bm *BasicManager) Register(migratorFactory MigratorFactory) {
	if migratorFactory == nil {
		return // ignore, do nothing
	}

	migrator, err := migratorFactory(bm.pool, bm.namespace)
	if err != nil {
		logger.Errorf("migrator register error: %s", err)
		return
	}

	newNode := &MigratorChainNode{
		migrator: migrator,
		next:     nil,
	}

	if bm.head == nil {
		bm.head = newNode
		return
	}

	bm.head.next = newNode
}

// Migrate data
func (bm *BasicManager) Migrate() error {
	conn := bm.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	// Read schema version first
	v, err := redis.String(conn.Do("GET", VersionKey(bm.namespace)))
	if err != nil && err != redis.ErrNil {
		return errors.Wrap(err, "read schema version failed")
	}

	if len(v) > 0 {
		current, err := semver.NewVersion(v)
		if err != nil {
			return errors.Wrap(err, "malformed schema version")
		}
		nowV, _ := semver.NewVersion(SchemaVersion)

		diff := nowV.Compare(current)
		if diff < 0 {
			return errors.Errorf("the schema version of migrator is smaller that the one in the rdb: %s<%s", nowV.String(), current.String())
		} else if diff == 0 {
			logger.Info("No migration needed")
			return nil
		}
	}

	if bm.head == nil {
		logger.Warning("No migrator registered, passed migration")
		return nil
	}

	logger.Info("Process for migrating data is started")

	h := bm.head
	for h != nil {
		meta := h.migrator.Metadata()
		if meta == nil {
			// Make metadata required
			return errors.Errorf("no metadata provided for the migrator %s", reflect.TypeOf(h.migrator).String())
		}

		logger.Infof("Migrate %s from %s to %s", meta.ObjectRef, meta.FromVersion, meta.ToVersion)
		if err := h.migrator.Migrate(); err != nil {
			return errors.Wrap(err, "migration chain calling failed")
		}

		// Next one if existing
		h = h.next
	}

	// Set schema version
	if _, err = conn.Do("SET", VersionKey(bm.namespace), SchemaVersion); err != nil {
		return errors.Wrap(err, "write schema version failed")
	}

	logger.Infof("Data schema version upgraded to %s", SchemaVersion)

	return nil
}
