// copyright (c) 2017 vmware, inc. all rights reserved.
//
// licensed under the apache license, version 2.0 (the "license");
// you may not use this file except in compliance with the license.
// you may obtain a copy of the license at
//
//    http://www.apache.org/licenses/license-2.0
//
// unless required by applicable law or agreed to in writing, software
// distributed under the license is distributed on an "as is" basis,
// without warranties or conditions of any kind, either express or implied.
// see the license for the specific language governing permissions and
// limitations under the license.

package dao

import (
	"fmt"
	"os"
	"strconv"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

var defaultRegistered = false

// PrepareTestForMySQL is for test only.
func PrepareTestForMySQL() {
	dbHost := os.Getenv("MYSQL_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable MYSQL_HOST is not set")
	}
	dbUser := os.Getenv("MYSQL_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable MYSQL_USR is not set")
	}
	dbPortStr := os.Getenv("MYSQL_PORT")
	if len(dbPortStr) == 0 {
		log.Fatalf("environment variable MYSQL_PORT is not set")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("invalid MYSQL_PORT: %v", err)
	}

	dbPassword := os.Getenv("MYSQL_PWD")
	dbDatabase := os.Getenv("MYSQL_DATABASE")
	if len(dbDatabase) == 0 {
		log.Fatalf("environment variable MYSQL_DATABASE is not set")
	}

	database := &models.Database{
		Type: "mysql",
		MySQL: &models.MySQL{
			Host:     dbHost,
			Port:     dbPort,
			Username: dbUser,
			Password: dbPassword,
			Database: dbDatabase,
		},
	}

	log.Infof("MYSQL_HOST: %s, MYSQL_USR: %s, MYSQL_PORT: %d, MYSQL_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)
	initDatabaseForTest(database)
}

// PrepareTestForSQLite is for test only.
func PrepareTestForSQLite() {
	file := os.Getenv("SQLITE_FILE")
	if len(file) == 0 {
		log.Fatalf("environment variable SQLITE_FILE is not set")
	}

	database := &models.Database{
		Type: "sqlite",
		SQLite: &models.SQLite{
			File: file,
		},
	}
	initDatabaseForTest(database)
}

func initDatabaseForTest(db *models.Database) {
	database, err := getDatabase(db)
	if err != nil {
		panic(err)
	}

	log.Infof("initializing database: %s", database.String())

	alias := database.Name()
	if !defaultRegistered {
		defaultRegistered = true
		alias = "default"
	}
	if err := database.Register(alias); err != nil {
		panic(err)
	}

	if alias != "default" {
		if err = globalOrm.Using(alias); err != nil {
			log.Fatalf("failed to create new orm: %v", err)
		}
	}
}

// PrepareTestData -- Clean and Create data
func PrepareTestData(clearSqls []string, initSqls []string) {
	o := GetOrmer()

	for _, sql := range clearSqls {
		fmt.Printf("Exec sql:%v\n", sql)
		_, err := o.Raw(sql).Exec()
		if err != nil {
			fmt.Printf("failed to clear database, sql:%v, error: %v", sql, err)
		}
	}

	for _, sql := range initSqls {
		_, err := o.Raw(sql).Exec()
		if err != nil {
			fmt.Printf("failed to init database, sql:%v, error: %v", sql, err)
		}
	}
}
