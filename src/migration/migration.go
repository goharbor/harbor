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
	"context"
	"fmt"
	"time"

	beegorm "github.com/beego/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/golang-migrate/migrate/v4"
)

const (
	schemaVersionV1_10_0 = 15
	// data version for tracking the data integrity in the DB, it can be different from schema version
	dataversionV2_0_0 = 30
)

// MigrateDB upgrades DB schema and do necessary transformation of the data in DB
func MigrateDB(database *models.Database) error {
	// check the database schema version
	migrator, err := dao.NewMigrator(database.PostGreSQL)
	if err != nil {
		return err
	}
	defer migrator.Close()
	schemaVersion, _, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return err
	}
	log.Debugf("current database schema version: %v", schemaVersion)
	// prior to 1.9, version = 0 means fresh install
	if schemaVersion > 0 && schemaVersion < schemaVersionV1_10_0 {
		return fmt.Errorf("please upgrade to version 1.10 first")
	}
	// update database schema
	return dao.UpgradeSchema(database)
}

// AbstractArtifactData accesses the registry to
func AbstractArtifactData() error {
	log.Info("Abstracting artifact data to DB...")
	ctx := orm.NewContext(context.Background(), beegorm.NewOrm())
	dataVersion, err := getDataVersion(ctx)
	if err != nil {
		return err
	}
	log.Debugf("current data version: %v", dataVersion)
	// the abstract logic already done before, skip
	if dataVersion >= dataversionV2_0_0 {
		log.Info("No need to abstract artifact data. Skip")
		return nil
	}
	if err = abstractArtData(ctx); err != nil {
		return err
	}
	log.Info("Abstract artifact data to DB done")
	return nil
}

// Migrate the database schema and abstract artifact data
func Migrate(database *models.Database) error {
	if err := MigrateDB(database); err != nil {
		return err
	}
	if err := AbstractArtifactData(); err != nil {
		return err
	}
	return nil
}

type dataVersion struct {
	ID           int64
	Version      int
	CreationTime time.Time
	UpdateTime   time.Time
}

func getDataVersion(ctx context.Context) (int, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	versions := []*dataVersion{}
	if _, err = ormer.Raw("select * from data_migrations order by id").QueryRows(&versions); err != nil {
		return 0, err
	}
	n := len(versions)
	if n == 0 {
		return 0, nil
	}
	if n > 1 {
		return 0, fmt.Errorf("there should be only one record in the table data_migrations, but found %d records", n)
	}
	return versions[0].Version, nil
}

func setDataVersion(ctx context.Context, version int) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = ormer.Raw("update data_migrations set version=?, update_time=?", version, time.Now()).Exec()
	return err
}
