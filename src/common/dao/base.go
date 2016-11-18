/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package dao

import (
	"fmt"
	"strings"
	"sync"

	"github.com/astaxie/beego/orm"
	"github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/utils/log"
)

// NonExistUserID : if a user does not exist, the ID of the user will be 0.
const NonExistUserID = 0

// Database is an interface of different databases
type Database interface {
	// Name returns the name of database
	Name() string
	// String returns the details of database
	String() string
	// Register registers the database which will be used
	Register(alias ...string) error
}

// InitDatabase initializes the database
func InitDatabase() {
	database, err := getDatabase()
	if err != nil {
		panic(err)
	}

	log.Infof("initializing database: %s", database.String())
	if err := database.Register(); err != nil {
		panic(err)
	}
}

func getDatabase() (db Database, err error) {
	switch config.Database() {
	case "", "mysql":
		db = NewMySQL(config.MySQL().Host, config.MySQL().Port, config.MySQL().User,
			config.MySQL().Password, config.MySQL().Database)
	case "sqlite":
		db = NewSQLite(config.SQLite().FilePath)
	default:
		err = fmt.Errorf("invalid database: %s", config.Database())
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

func paginateForRawSQL(sql string, limit, offset int64) string {
	return fmt.Sprintf("%s limit %d offset %d", sql, limit, offset)
}

func escape(str string) string {
	str = strings.Replace(str, `%`, `\%`, -1)
	str = strings.Replace(str, `_`, `\_`, -1)
	return str
}
