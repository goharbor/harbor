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
	"syscall"
	"time"

	"github.com/astaxie/beego"
	_ "github.com/astaxie/beego/session/redis"

	"github.com/goharbor/harbor/src/common/dao"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	_ "github.com/goharbor/harbor/src/controller/event/handler"
	"github.com/goharbor/harbor/src/core/api"
	_ "github.com/goharbor/harbor/src/core/auth/authproxy"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	_ "github.com/goharbor/harbor/src/core/auth/oidc"
	_ "github.com/goharbor/harbor/src/core/auth/uaa"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/migration"
	"github.com/goharbor/harbor/src/pkg/notification"
	_ "github.com/goharbor/harbor/src/pkg/notifier/topic"
	"github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/version"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/server"
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
	beego.BConfig.WebConfig.Session.SessionName = config.SessionCookieName

	redisURL := os.Getenv("_REDIS_URL")
	if len(redisURL) > 0 {
		gob.Register(models.User{})
		beego.BConfig.WebConfig.Session.SessionProvider = "redis"
		beego.BConfig.WebConfig.Session.SessionProviderConfig = redisURL
	}
	beego.AddTemplateExt("htm")

	log.Info("initializing configurations...")
	config.Init()
	log.Info("configurations initialization completed")
	token.InitCreators()
	database, err := config.Database()
	if err != nil {
		log.Fatalf("failed to get database configuration: %v", err)
	}
	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	if err = migration.Migrate(database); err != nil {
		log.Fatalf("failed to migrate: %v", err)
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

	registerScanners()

	closing := make(chan struct{})
	done := make(chan struct{})
	go gracefulShutdown(closing, done)
	if err := replication.Init(closing, done); err != nil {
		log.Fatalf("failed to init for replication: %v", err)
	}

	log.Info("initializing notification...")
	notification.Init()

	server.RegisterRoutes()

	if common_http.InternalTLSEnabled() {
		log.Info("internal TLS enabled, Init TLS ...")
		iTLSKeyPath := os.Getenv("INTERNAL_TLS_KEY_PATH")
		iTLSCertPath := os.Getenv("INTERNAL_TLS_CERT_PATH")

		log.Infof("load client key: %s client cert: %s", iTLSKeyPath, iTLSCertPath)
		beego.BConfig.Listen.EnableHTTP = false
		beego.BConfig.Listen.EnableHTTPS = true
		beego.BConfig.Listen.HTTPSPort = 8443
		beego.BConfig.Listen.HTTPSKeyFile = iTLSKeyPath
		beego.BConfig.Listen.HTTPSCertFile = iTLSCertPath
		beego.BeeApp.Server.TLSConfig = common_http.NewServerTLSConfig()
	}

	log.Infof("Version: %s, Git commit: %s", version.ReleaseVersion, version.GitCommit)
	beego.RunWithMiddleWares("", middlewares.MiddleWares()...)
}

func registerScanners() {
	wantedScanners := make([]scanner.Registration, 0)
	uninstallURLs := make([]string, 0)

	if config.WithTrivy() {
		log.Info("Registering Trivy scanner")
		wantedScanners = append(wantedScanners, scanner.Registration{
			Name:            "Trivy",
			Description:     "The Trivy scanner adapter",
			URL:             config.TrivyAdapterURL(),
			UseInternalAddr: true,
			Immutable:       true,
		})
	} else {
		log.Info("Removing Trivy scanner")
		uninstallURLs = append(uninstallURLs, config.TrivyAdapterURL())
	}

	if config.WithClair() {
		clairDB, err := config.ClairDB()
		if err != nil {
			log.Fatalf("failed to load clair database information: %v", err)
		}
		if err := dao.InitClairDB(clairDB); err != nil {
			log.Fatalf("failed to initialize clair database: %v", err)
		}

		log.Info("Registering Clair scanner")
		wantedScanners = append(wantedScanners, scanner.Registration{
			Name:            "Clair",
			Description:     "The Clair scanner adapter",
			URL:             config.ClairAdapterEndpoint(),
			UseInternalAddr: true,
			Immutable:       true,
		})
	} else {
		log.Info("Removing Clair scanner")
		uninstallURLs = append(uninstallURLs, config.ClairAdapterEndpoint())
	}

	if err := scan.EnsureScanners(wantedScanners); err != nil {
		log.Fatalf("failed to register scanners: %v", err)
	}

	if defaultScannerURL := getDefaultScannerURL(); defaultScannerURL != "" {
		log.Infof("Setting %s as default scanner", defaultScannerURL)
		if err := scan.EnsureDefaultScanner(defaultScannerURL); err != nil {
			log.Fatalf("failed to set default scanner: %v", err)
		}
	}

	if err := scan.RemoveImmutableScanners(uninstallURLs); err != nil {
		log.Warningf("failed to remove scanners: %v", err)
	}

}

func getDefaultScannerURL() string {
	if config.WithTrivy() {
		return config.TrivyAdapterURL()
	}
	if config.WithClair() {
		return config.ClairAdapterEndpoint()
	}
	return ""
}
