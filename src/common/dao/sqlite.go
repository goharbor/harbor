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

	"github.com/beego/beego/orm"
	// _ "github.com/mattn/go-sqlite3" // register sqlite driver
)

type sqlite struct {
	file string
}

// NewSQLite returns an instance of sqlite
func NewSQLite(file string) Database {
	return &sqlite{
		file: file,
	}
}

// Register registers SQLite as the underlying database used
func (s *sqlite) Register(alias ...string) error {
	if err := orm.RegisterDriver("sqlite3", orm.DRSqlite); err != nil {
		return err
	}

	an := "default"
	if len(alias) != 0 {
		an = alias[0]
	}
	return orm.RegisterDataBase(an, "sqlite3", s.file)
}

// Name returns the name of SQLite
func (s *sqlite) Name() string {
	return "SQLite"
}

// String returns the details of database
func (s *sqlite) String() string {
	return fmt.Sprintf("type-%s file:%s", s.Name(), s.file)
}

// UpgradeSchema is not supported for SQLite, it assumes the schema is initialized and up to date in the DB instance.
func (s *sqlite) UpgradeSchema() error {
	return nil
}
