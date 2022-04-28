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
	"time"

	"github.com/beego/beego/orm"
	_ "github.com/go-sql-driver/mysql" // register mysql driver
	"github.com/goharbor/harbor/src/common/utils"
)

type mysql struct {
	host     string
	port     string
	usr      string
	pwd      string
	database string
}

// NewMySQL returns an instance of mysql
func NewMySQL(host, port, usr, pwd, database string) Database {
	return &mysql{
		host:     host,
		port:     port,
		usr:      usr,
		pwd:      pwd,
		database: database,
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
	if err := orm.RegisterDataBase(an, "mysql", conn); err != nil {
		return err
	}
	db, _ := orm.GetDB(an)
	db.SetConnMaxLifetime(5 * time.Minute)
	return nil
}

// Name returns the name of MySQL
func (m *mysql) Name() string {
	return "MySQL"
}

// UpgradeSchema is not supported for MySQL, it assumes the schema is initialized and up to date in the DB instance.
func (m *mysql) UpgradeSchema() error {
	return nil
}

// String returns the details of database
func (m *mysql) String() string {
	return fmt.Sprintf("type-%s host-%s port-%s user-%s database-%s",
		m.Name(), m.host, m.port, m.usr, m.database)
}
