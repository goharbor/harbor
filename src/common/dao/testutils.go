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

	"github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/log"
)

var defaultRegistered = false
var o orm.Ormer

// PrepareTestForSQLite is for test only.
func PrepareTestForSQLite() {
}

// PrepareTestForPostgresSQL is for test only.
func PrepareTestForPostgresSQL() {
	dbHost := os.Getenv("POSTGRESQL_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable POSTGRESQL_HOST is not set")
	}
	dbUser := os.Getenv("POSTGRESQL_USR")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable POSTGRESQL_USR is not set")
	}
	dbPortStr := os.Getenv("POSTGRESQL_PORT")
	if len(dbPortStr) == 0 {
		log.Fatalf("environment variable POSTGRESQL_PORT is not set")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("invalid POSTGRESQL_PORT: %v", err)
	}

	dbPassword := os.Getenv("POSTGRESQL_PWD")
	dbDatabase := os.Getenv("POSTGRESQL_DATABASE")
	if len(dbDatabase) == 0 {
		log.Fatalf("environment variable POSTGRESQL_DATABASE is not set")
	}

	database := &models.Database{
		Type: "postgresql",
		PostGreSQL: &models.PostGreSQL{
			Host:         dbHost,
			Port:         dbPort,
			Username:     dbUser,
			Password:     dbPassword,
			Database:     dbDatabase,
			MaxIdleConns: 50,
			MaxOpenConns: 100,
		},
	}

	log.Infof("POSTGRES_HOST: %s, POSTGRES_USR: %s, POSTGRES_PORT: %d, POSTGRES_PWD: %s\n", dbHost, dbUser, dbPort, dbPassword)
	o = initDatabaseForTest(database)
}

func initDatabaseForTest(db *models.Database) orm.Ormer {
	database, err := getDatabase(db)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	log.Infof("initializing database: %s", database.String())

	alias := database.Name()
	if !defaultRegistered {
		defaultRegistered = true
		alias = "default"
	}
	if err := database.Register(alias); err != nil {
		log.Fatal(err)
		return nil
	}
	if err := database.UpgradeSchema(); err != nil {
		log.Fatal(err)
		return nil
	}

	if alias != "default" {
		return orm.NewOrmUsingDB(alias)
	}
	return GetOrmer()
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

// ExecuteBatchSQL ...
func ExecuteBatchSQL(sqls []string) {
	o := GetOrmer()

	for _, sql := range sqls {
		fmt.Printf("Exec sql:%v\n", sql)
		_, err := o.Raw(sql).Exec()
		if err != nil {
			fmt.Printf("failed to execute batch sql, sql:%v, error: %v", sql, err)
		}
	}
}

// ArrayEqual ...
func ArrayEqual(arrayA, arrayB []int) bool {
	if len(arrayA) != len(arrayB) {
		return false
	}
	size := len(arrayA)
	for i := 0; i < size; i++ {
		if arrayA[i] != arrayB[i] {
			return false
		}
	}
	return true
}

// ClearHTTPAuthProxyUsers remove the records from harbor_users to delete all user imported via
// HTTP Auth Proxy
func ClearHTTPAuthProxyUsers() error {
	o := GetOrmer()
	sql := "DELETE FROM harbor_user WHERE comment='By Authproxy'"
	_, err := o.Raw(sql).Exec()
	return err
}
