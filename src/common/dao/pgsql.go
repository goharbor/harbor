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
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/beego/beego/v2/client/orm"
	migrate "github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx" // import pgx driver for migrator
	_ "github.com/golang-migrate/migrate/v4/source/file"  // import local file driver for migrator
	_ "github.com/jackc/pgx/v4/stdlib"                    // registry pgx driver

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
)

const defaultMigrationPath = "migrations/postgresql/"

type pgsql struct {
	host            string
	port            string
	usr             string
	pwd             string
	database        string
	sslmode         string
	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration
	connMaxIdleTime time.Duration
}

// Name returns the name of PostgreSQL
func (p *pgsql) Name() string {
	return "PostgreSQL"
}

// String ...
func (p *pgsql) String() string {
	return fmt.Sprintf("type-%s host-%s port-%s database-%s sslmode-%q",
		p.Name(), p.host, p.port, p.database, p.sslmode)
}

// NewPGSQL returns an instance of postgres
func NewPGSQL(host string, port string, usr string, pwd string, database string, sslmode string, maxIdleConns int, maxOpenConns int, connMaxLifetime time.Duration, connMaxIdleTime time.Duration) Database {
	if len(sslmode) == 0 {
		sslmode = "disable"
	}
	return &pgsql{
		host:            host,
		port:            port,
		usr:             usr,
		pwd:             pwd,
		database:        database,
		sslmode:         sslmode,
		maxIdleConns:    maxIdleConns,
		maxOpenConns:    maxOpenConns,
		connMaxLifetime: connMaxLifetime,
		connMaxIdleTime: connMaxIdleTime,
	}
}

// Register registers pgSQL to orm with the info wrapped by the instance.
func (p *pgsql) Register(alias ...string) error {
	if err := utils.TestTCPConn(net.JoinHostPort(p.host, p.port), 60, 2); err != nil {
		return err
	}

	if err := orm.RegisterDriver("pgx", orm.DRPostgres); err != nil {
		return err
	}

	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}
	info := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s timezone=UTC options='-c statement_timeout=20min'",
		p.host, p.port, p.usr, p.pwd, p.database, p.sslmode)

	if err := orm.RegisterDataBase(an, "pgx", info, orm.MaxIdleConnections(p.maxIdleConns),
		orm.MaxOpenConnections(p.maxOpenConns), orm.ConnMaxLifetime(p.connMaxLifetime)); err != nil {
		return err
	}

	db, err := orm.GetDB(an)
	if err != nil {
		return err
	}
	db.SetConnMaxIdleTime(p.connMaxIdleTime)

	return nil
}

// UpgradeSchema calls migrate tool to upgrade schema to the latest based on the SQL scripts.
func (p *pgsql) UpgradeSchema() error {
	port, err := strconv.Atoi(p.port)
	if err != nil {
		return err
	}
	m, err := NewMigrator(&models.PostGreSQL{
		Host:     p.host,
		Port:     port,
		Username: p.usr,
		Password: p.pwd,
		Database: p.database,
		SSLMode:  p.sslmode,
	})
	if err != nil {
		return err
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil || dbErr != nil {
			log.Warningf("Failed to close migrator, source error: %v, db error: %v", srcErr, dbErr)
		}
	}()
	log.Infof("Upgrading schema for pgsql ...")
	err = m.Up()
	if err == migrate.ErrNoChange {
		log.Infof("No change in schema, skip.")
	} else if err != nil { // migrate.ErrLockTimeout will be thrown when another process is doing migration and timeout.
		log.Errorf("Failed to upgrade schema, error: %q", err)
		return err
	}
	return nil
}

// NewMigrator creates a migrator base on the information
func NewMigrator(database *models.PostGreSQL) (*migrate.Migrate, error) {
	dbURL := url.URL{
		Scheme:   "pgx",
		User:     url.UserPassword(database.Username, database.Password),
		Host:     net.JoinHostPort(database.Host, strconv.Itoa(database.Port)),
		Path:     database.Database,
		RawQuery: fmt.Sprintf("sslmode=%s", database.SSLMode),
	}

	// For UT
	path := os.Getenv("POSTGRES_MIGRATION_SCRIPTS_PATH")
	if len(path) == 0 {
		path = defaultMigrationPath
	}
	srcURL := fmt.Sprintf("file://%s", path)
	m, err := migrate.New(srcURL, dbURL.String())
	if err != nil {
		return nil, err
	}
	m.Log = newMigrateLogger()
	return m, nil
}
