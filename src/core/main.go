// Copyright 2018 Project Harbor Authors
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
	"encoding/gob"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/models"
	common_quota "github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api"
	quota "github.com/goharbor/harbor/src/core/api/quota"
	_ "github.com/goharbor/harbor/src/core/api/quota/chart"
	_ "github.com/goharbor/harbor/src/core/api/quota/registry"
	_ "github.com/goharbor/harbor/src/core/auth/authproxy"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	_ "github.com/goharbor/harbor/src/core/auth/oidc"
	_ "github.com/goharbor/harbor/src/core/auth/uaa"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/filter"
	"github.com/goharbor/harbor/src/core/middlewares"
	_ "github.com/goharbor/harbor/src/core/notifier/topic"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/event"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/goharbor/harbor/src/replication"
)

const (
	adminUserID = 1
)

func updateInitPassword(userID int, password string) error {
	queryUser := models.User{UserID: userID}
	user, err := dao.GetUser(queryUser)
	if err != nil {
		return fmt.Errorf("Failed to get user, userID: %d %v", userID, err)
	}
	if user == nil {
		return fmt.Errorf("user id: %d does not exist", userID)
	}
	if user.Salt == "" {
		salt := utils.GenerateRandomString()

		user.Salt = salt
		user.Password = password
		err = dao.ChangeUserPassword(*user)
		if err != nil {
			return fmt.Errorf("Failed to update user encrypted password, userID: %d, err: %v", userID, err)
		}

		log.Infof("User id: %d updated its encrypted password successfully.", userID)
	} else {
		log.Infof("User id: %d already has its encrypted password.", userID)
	}
	return nil
}

// Quota migration
func quotaSync() error {
	projects, err := dao.GetProjects(nil)
	if err != nil {
		log.Errorf("list project error, %v", err)
		return err
	}

	var pids []string
	for _, project := range projects {
		pids = append(pids, strconv.FormatInt(project.ProjectID, 10))
	}
	usages, err := dao.ListQuotaUsages(&models.QuotaUsageQuery{Reference: "project", ReferenceIDs: pids})
	if err != nil {
		log.Errorf("list quota usage error, %v", err)
		return err
	}

	// The condition handles these two cases:
	// 1, len(project) > 1 && len(usages) == 1. existing projects without usage, as we do always has 'library' usage in DB.
	// 2, migration fails at the phase of inserting usage into DB, and parts of them are inserted successfully.
	if len(projects) != len(usages) {
		log.Info("Start to sync quota data .....")
		if err := quota.Sync(config.GlobalProjectMgr, true); err != nil {
			log.Errorf("Fail to sync quota data, %v", err)
			return err
		}
		log.Info("Success to sync quota data .....")
		return nil
	}

	// Only has one project without usage
	zero := common_quota.ResourceList{
		common_quota.ResourceCount:   0,
		common_quota.ResourceStorage: 0,
	}
	if len(projects) == 1 && len(usages) == 1 {
		totalRepo, err := dao.GetTotalOfRepositories()
		if totalRepo == 0 {
			return nil
		}
		refID, err := strconv.ParseInt(usages[0].ReferenceID, 10, 64)
		if err != nil {
			log.Error(err)
			return err
		}
		usedRes, err := types.NewResourceList(usages[0].Used)
		if err != nil {
			log.Error(err)
			return err
		}
		if types.Equals(usedRes, zero) && refID == projects[0].ProjectID {
			log.Info("Start to sync quota data .....")
			if err := quota.Sync(config.GlobalProjectMgr, true); err != nil {
				log.Errorf("Fail to sync quota data, %v", err)
				return err
			}
			log.Info("Success to sync quota data .....")
		}
	}
	return nil
}

func gracefulShutdown(closing, done chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Infof("capture system signal %s, to close \"closing\" channel", <-signals)
	close(closing)
	select {
	case <-done:
		log.Infof("Goroutines exited normally")
	case <-time.After(time.Second * 3):
		log.Infof("Timeout waiting goroutines to exit")
	}
	os.Exit(0)
}

func main() {
	beego.BConfig.WebConfig.Session.SessionOn = true
	beego.BConfig.WebConfig.Session.SessionName = "sid"

	redisURL := os.Getenv("_REDIS_URL")
	if len(redisURL) > 0 {
		gob.Register(models.User{})
		beego.BConfig.WebConfig.Session.SessionProvider = "redis"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = redisURL
	}
	beego.AddTemplateExt("htm")

	log.Info("initializing configurations...")
	if err := config.Init(); err != nil {
		log.Fatalf("failed to initialize configurations: %v", err)
	}
	log.Info("configurations initialization completed")
	token.InitCreators()
	database, err := config.Database()
	if err != nil {
		log.Fatalf("failed to get database configuration: %v", err)
	}
	if err := dao.InitAndUpgradeDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	if err := config.Load(); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// init the jobservice client
	job.Init()
	// init the scheduler
	scheduler.Init()

	password, err := config.InitialAdminPassword()
	if err != nil {
		log.Fatalf("failed to get admin's initial password: %v", err)
	}
	if err := updateInitPassword(adminUserID, password); err != nil {
		log.Error(err)
	}

	// Init API handler
	if err := api.Init(); err != nil {
		log.Fatalf("Failed to initialize API handlers with error: %s", err.Error())
	}

	if config.WithClair() {
		clairDB, err := config.ClairDB()
		if err != nil {
			log.Fatalf("failed to load clair database information: %v", err)
		}
		if err := dao.InitClairDB(clairDB); err != nil {
			log.Fatalf("failed to initialize clair database: %v", err)
		}

		// TODO: change to be internal adapter
		reg := &scanner.Registration{
			Name:            "Clair",
			Description:     "The clair scanner adapter",
			URL:             config.ClairAdapterEndpoint(),
			IsDefault:       true,
			UseInternalAddr: true,
			Immutable:       true,
		}

		if err := scan.EnsureScanner(reg); err != nil {
			log.Fatalf("failed to initialize clair scanner: %v", err)
		}
	}

	closing := make(chan struct{})
	done := make(chan struct{})
	go gracefulShutdown(closing, done)
	if err := replication.Init(closing, done); err != nil {
		log.Fatalf("failed to init for replication: %v", err)
	}

	log.Info("initializing notification...")
	notification.Init()
	// Initialize the event handlers for handling artifact cascade deletion
	event.Init()

	filter.Init()
	beego.InsertFilter("/*", beego.BeforeRouter, filter.SecurityFilter)
	beego.InsertFilter("/*", beego.BeforeRouter, filter.ReadonlyFilter)
	beego.InsertFilter("/api/*", beego.BeforeRouter, filter.MediaTypeFilter("application/json", "multipart/form-data", "application/octet-stream"))

	initRouters()

	syncRegistry := os.Getenv("SYNC_REGISTRY")
	sync, err := strconv.ParseBool(syncRegistry)
	if err != nil {
		log.Errorf("Failed to parse SYNC_REGISTRY: %v", err)
		// if err set it default to false
		sync = false
	}
	if sync {
		if err := api.SyncRegistry(config.GlobalProjectMgr); err != nil {
			log.Error(err)
		}
	} else {
		log.Infof("Because SYNC_REGISTRY set false , no need to sync registry \n")
	}

	log.Info("Init proxy")
	if err := middlewares.Init(); err != nil {
		log.Fatalf("init proxy error, %v", err)
	}

	if err := quotaSync(); err != nil {
		log.Fatalf("quota migration error, %v", err)
	}

	beego.Run()
}
