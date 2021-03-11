package main

import (
	"github.com/goharbor/harbor/src/common/models"
	"os"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
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
	if err := migration.MigrateDB(db); err != nil {
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
