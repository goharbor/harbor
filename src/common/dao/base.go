// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

const (
	// NonExistUserID : if a user does not exist, the ID of the user will be 0.
	NonExistUserID = 0
	// ClairDBAlias ...
	ClairDBAlias = "clair-db"
)

// Database is an interface of different databases
type Database interface {
	// Name returns the name of database
	Name() string
	// String returns the details of database
	String() string
	// Register registers the database which will be used
	Register(alias ...string) error
	// UpgradeSchema upgrades the DB schema to the latest version
	UpgradeSchema() error
}

// InitClairDB ...
func InitClairDB(clairDB *models.PostGreSQL) error {
	p := &pgsql{
		host:     clairDB.Host,
		port:     strconv.Itoa(clairDB.Port),
		usr:      clairDB.Username,
		pwd:      clairDB.Password,
		database: clairDB.Database,
		sslmode:  false,
	}
	if err := p.Register(ClairDBAlias); err != nil {
		return err
	}
	log.Info("initialized clair database")
	return nil
}

// InitDatabase registers the database
func InitDatabase(database *models.Database) error {
	db, err := getDatabase(database)
	if err != nil {
		return err
	}

	log.Infof("Registering database: %s", db.String())
	if err := db.Register(); err != nil {
		return err
	}
	log.Info("Register database completed")

	log.Info("Upgrading schema...")
	if err = db.UpgradeSchema(); err != nil {
		return err
	}
	log.Info("Upgrade schema completed")

	version, err := GetSchemaVersion()
	if err != nil {
		return err
	}
	if version.Version != SchemaVersion {
		return fmt.Errorf("unexpected database schema version, expected %s, got %s",
			SchemaVersion, version.Version)
	}

	return nil
}

func getDatabase(database *models.Database) (db Database, err error) {

	switch database.Type {
	case "", "postgresql":
		db = NewPGSQL(database.PostGreSQL.Host,
			strconv.Itoa(database.PostGreSQL.Port),
			database.PostGreSQL.Username,
			database.PostGreSQL.Password,
			database.PostGreSQL.Database,
			false)
	default:
		err = fmt.Errorf("invalid database: %s", database.Type)
	}
	return
}

var globalOrm orm.Ormer
var once sync.Once

// GetOrmer :set ormer singleton
func GetOrmer() orm.Ormer {
	once.Do(func() {
		globalOrm = orm.NewOrm()
	})
	return globalOrm
}

// ClearTable is the shortcut for test cases, it should be called only in test cases.
func ClearTable(table string) error {
	o := GetOrmer()
	sql := fmt.Sprintf("delete from %s where 1=1", table)
	if table == models.ProjectTable {
		sql = fmt.Sprintf("delete from %s where project_id > 1", table)
	}
	if table == models.UserTable {
		sql = fmt.Sprintf("delete from %s where user_id > 2", table)
	}
	if table == "project_metadata" { //make sure library is public
		sql = fmt.Sprintf("delete from %s where id > 1", table)
	}
	_, err := o.Raw(sql).Exec()
	return err
}

func paginateForRawSQL(sql string, limit, offset int64) string {
	return fmt.Sprintf("%s limit %d offset %d", sql, limit, offset)
}

func paginateForQuerySetter(qs orm.QuerySeter, page, size int64) orm.QuerySeter {
	if size > 0 {
		qs = qs.Limit(size)
		if page > 0 {
			qs = qs.Offset((page - 1) * size)
		}
	}
	return qs
}

//Escape ..
func Escape(str string) string {
	str = strings.Replace(str, `%`, `\%`, -1)
	str = strings.Replace(str, `_`, `\_`, -1)
	return str
}
