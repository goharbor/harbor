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
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql" // register mysql driver
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql" // import mysql driver for migrator
)

const defaultMysqlMigrationPath = "migrations/mysql/"

type mysql struct {
	host         string
	port         string
	usr          string
	pwd          string
	database     string
	maxIdleConns int
	maxOpenConns int
}

// NewMySQL returns an instance of mysql
func NewMySQL(host, port, usr, pwd, database string, maxIdleConns int, maxOpenConns int) Database {
	return &mysql{
		host:         host,
		port:         port,
		usr:          usr,
		pwd:          pwd,
		database:     database,
		maxIdleConns: maxIdleConns,
		maxOpenConns: maxOpenConns,
	}
}

// Register registers MySQL as the underlying database used
func (m *mysql) Register(alias ...string) error {

	if err := utils.TestTCPConn(m.host+":"+m.port, 60, 2); err != nil {
		return err
	}

	if err := orm.RegisterDriver("mysql", orm.DRMySQL); err != nil {
		return err
	}

	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}
	conn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", m.usr,
		m.pwd, m.host, m.port, m.database)
	if err := orm.RegisterDataBase(an, "mysql", conn, m.maxIdleConns, m.maxOpenConns); err != nil {
		return err
	}
	db, _ := orm.GetDB(an)
	db.SetMaxOpenConns(m.maxOpenConns)
	db.SetConnMaxLifetime(5 * time.Minute)

	return nil
}

// Name returns the name of MySQL
func (m *mysql) Name() string {
	return "MySQL"
}

// UpgradeSchema is not supported for MySQL, it assumes the schema is initialized and up to date in the DB instance.
func (m *mysql) UpgradeSchema() error {
	port, err := strconv.ParseInt(m.port, 10, 64)
	if err != nil {
		return err
	}
	mg, err := NewMysqlMigrator(&models.MySQL{
		Host:     m.host,
		Port:     int(port),
		Username: m.usr,
		Password: m.pwd,
		Database: m.database,
	})
	if err != nil {
		return err
	}
	defer func() {
		srcErr, dbErr := mg.Close()
		if srcErr != nil || dbErr != nil {
			log.Warningf("Failed to close migrator, source error: %v, db error: %v", srcErr, dbErr)
		}
	}()
	log.Infof("Upgrading schema for mysql ...")
	err = mg.Up()
	if err == migrate.ErrNoChange {
		log.Infof("No change in schema, skip.")
	} else if err != nil { // migrate.ErrLockTimeout will be thrown when another process is doing migration and timeout.
		log.Errorf("Failed to upgrade schema, error: %q", err)
		return err
	}
	return nil
}

// String returns the details of database
func (m *mysql) String() string {
	return fmt.Sprintf("type-%s host-%s port-%s user-%s database-%s",
		m.Name(), m.host, m.port, m.usr, m.database)
}

// NewMysqlMigrator creates a migrator base on the information
func NewMysqlMigrator(database *models.MySQL) (*migrate.Migrate, error) {
	dbURL := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s", database.Username,
		database.Password, database.Host, database.Port, database.Database)
	// For UT
	path := os.Getenv("MYSQL_MIGRATION_SCRIPTS_PATH")
	if len(path) == 0 {
		path = defaultMysqlMigrationPath
	}
	srcURL := fmt.Sprintf("file://%s", path)
	m, err := migrate.New(srcURL, dbURL)
	if err != nil {
		return nil, err
	}
	m.Log = newMigrateLogger()
	return m, nil
}
