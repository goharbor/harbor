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

package main

import (
	"os"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/migration"
)

// key: env var, value: default value
var defaultAttrs = map[string]string{
	"POSTGRESQL_HOST":     "localhost",
	"POSTGRESQL_PORT":     "5432",
	"POSTGRESQL_USERNAME": "postgres",
	"POSTGRESQL_PASSWORD": "password",
	"POSTGRESQL_DATABASE": "registry",
	"POSTGRESQL_SSLMODE":  "disable",
}

func main() {
	p, _ := strconv.Atoi(getAttr("POSTGRESQL_PORT"))
	db := &models.Database{
		Type: "postgresql",
		PostGreSQL: &models.PostGreSQL{
			Host:         getAttr("POSTGRESQL_HOST"),
			Port:         p,
			Username:     getAttr("POSTGRESQL_USERNAME"),
			Password:     getAttr("POSTGRESQL_PASSWORD"),
			Database:     getAttr("POSTGRESQL_DATABASE"),
			SSLMode:      getAttr("POSTGRESQL_SSLMODE"),
			MaxIdleConns: 5,
			MaxOpenConns: 5,
		},
	}

	log.Info("Migrating the data to latest schema...")
	log.Infof("DB info: postgres://%s@%s:%d/%s?sslmode=%s", db.PostGreSQL.Username, db.PostGreSQL.Host,
		db.PostGreSQL.Port, db.PostGreSQL.Database, db.PostGreSQL.SSLMode)

	if err := dao.InitDatabase(db); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	if err := migration.Migrate(db); err != nil {
		log.Fatalf("failed to migrate DB: %v", err)
	}
	log.Info("Migration done.  The data schema in DB is now update to date.")
}

func getAttr(k string) string {
	v := os.Getenv(k)
	if len(v) > 0 {
		return v
	}
	return defaultAttrs[k]
}
