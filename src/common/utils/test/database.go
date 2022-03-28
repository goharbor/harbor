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

package test

import (
	"fmt"
	"os"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	pkguser "github.com/goharbor/harbor/src/pkg/user"
)

// InitDatabaseFromEnv is used to initialize database for testing
func InitDatabaseFromEnv() {
	dbHost := os.Getenv("DB_HOST")
	if len(dbHost) == 0 {
		log.Fatalf("environment variable DB_HOST is not set")
	}
	dbUser := os.Getenv("DB_USERNAME")
	if len(dbUser) == 0 {
		log.Fatalf("environment variable DB_USERNAME is not set")
	}
	dbPortStr := os.Getenv("DB_PORT")
	if len(dbPortStr) == 0 {
		log.Fatalf("environment variable DB_PORT is not set")
	}
	dbPort, err := strconv.Atoi(dbPortStr)
	if err != nil {
		log.Fatalf("invalid DB_PORT: %v", err)
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	dbDatabase := os.Getenv("DB_DATABASE")
	dbCollation := os.Getenv("DB_COLLATION")
	adminPwd := os.Getenv("HARBOR_ADMIN_PASSWD")
	if len(dbDatabase) == 0 {
		log.Fatalf("environment variable DB_DATABASE is not set")
	}

	database := &models.Database{}
	switch {
	case utils.IsDBPostgresql():
		database = &models.Database{
			Type: "postgresql",
			PostGreSQL: &models.PostGreSQL{
				Host:     dbHost,
				Port:     dbPort,
				Username: dbUser,
				Password: dbPassword,
				Database: dbDatabase,
			},
		}
	case utils.IsDBMysql():
		database = &models.Database{
			Type: "mysql",
			MySQL: &models.MySQL{
				Host:      dbHost,
				Port:      dbPort,
				Username:  dbUser,
				Password:  dbPassword,
				Database:  dbDatabase,
				Collation: dbCollation,
			},
		}
	default:
		log.Fatalf("invalid db type %s", os.Getenv("DATABASE_TYPE"))
	}

	log.Infof("DB_HOST: %s, DB_USERNAME: %s, DB_PORT: %d, DB_PASSWORD: %s\n", dbHost, dbUser, dbPort, dbPassword)

	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to init database : %v", err)
	}
	if err := dao.UpgradeSchema(database); err != nil {
		log.Fatalf("failed to upgrade database : %v", err)
	}
	if err := updateUserInitialPassword(1, adminPwd); err != nil {
		log.Fatalf("failed to init password for admin: %v", err)
	}
}

func updateUserInitialPassword(userID int, password string) error {
	ctx := orm.Context()
	user, err := pkguser.Mgr.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user, userID: %d %v", userID, err)
	}
	if user.Salt == "" {
		err = pkguser.Mgr.UpdatePassword(ctx, userID, password)
		if err != nil {
			return fmt.Errorf("failed to update user encrypted password, userID: %d, err: %v", userID, err)
		}
	}
	return nil
}
