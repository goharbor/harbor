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
	"os"
	"strings"
	"sync"

	"github.com/astaxie/beego/orm"
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
	switch strings.ToLower(os.Getenv("DATABASE")) {
	case "", "mysql":
		host, port, usr, pwd, database := getMySQLConnInfo()
		db = NewMySQL(host, port, usr, pwd, database)
	case "sqlite":
		file := getSQLiteConnInfo()
		db = NewSQLite(file)
	default:
		err = fmt.Errorf("invalid database: %s", os.Getenv("DATABASE"))
	}

	return
}

// TODO read from config
func getMySQLConnInfo() (host, port, username, password, database string) {
	host = os.Getenv("MYSQL_HOST")
	port = os.Getenv("MYSQL_PORT")
	username = os.Getenv("MYSQL_USR")
	password = os.Getenv("MYSQL_PWD")
	database = os.Getenv("MYSQL_DATABASE")
	if len(database) == 0 {
		database = "registry"
	}
	return
}

// TODO read from config
func getSQLiteConnInfo() string {
	file := os.Getenv("SQLITE_FILE")
	if len(file) == 0 {
		file = "registry.db"
	}
	return file
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
