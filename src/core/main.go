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
	"context"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/beego/beego/v2/server/web"

	"github.com/goharbor/harbor/src/common/dao"
	common_http "github.com/goharbor/harbor/src/common/http"
	configCtl "github.com/goharbor/harbor/src/controller/config"
	_ "github.com/goharbor/harbor/src/controller/event/handler"
	"github.com/goharbor/harbor/src/controller/health"
	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/controller/systemartifact"
	"github.com/goharbor/harbor/src/controller/task"
	"github.com/goharbor/harbor/src/core/api"
	_ "github.com/goharbor/harbor/src/core/auth/authproxy"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	_ "github.com/goharbor/harbor/src/core/auth/oidc"
	_ "github.com/goharbor/harbor/src/core/auth/uaa"
	"github.com/goharbor/harbor/src/core/middlewares"
	"github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/core/session"
	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/memory" // memory cache
	_ "github.com/goharbor/harbor/src/lib/cache/redis"  // redis cache
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/gtask"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/metric"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/retry"
	tracelib "github.com/goharbor/harbor/src/lib/trace"
	"github.com/goharbor/harbor/src/migration"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/notation"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/nydus"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/sbom"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/subject"
	"github.com/goharbor/harbor/src/pkg/audit"
	dbCfg "github.com/goharbor/harbor/src/pkg/config/db"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/notification"
	_ "github.com/goharbor/harbor/src/pkg/notifier/topic"
	"github.com/goharbor/harbor/src/pkg/oidc"
	"github.com/goharbor/harbor/src/pkg/scan"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	pkguser "github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/pkg/version"
	"github.com/goharbor/harbor/src/server"
)

const (
	adminUserID = 1
)

func updateInitPassword(ctx context.Context, userID int, password string) error {
	userMgr := pkguser.Mgr
	user, err := userMgr.Get(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user, userID: %d %v", userID, err)
	}
	if user.Salt == "" {
		err = userMgr.UpdatePassword(ctx, userID, password)
		if err != nil {
			return fmt.Errorf("failed to update user encrypted password, userID: %d, err: %v", userID, err)
		}

		log.Infof("User id: %d updated its encrypted password successfully.", userID)
	} else {
		log.Infof("User id: %d already has its encrypted password.", userID)
	}
	return nil
}

func gracefulShutdown(closing, done chan struct{}, shutdowns ...func()) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	log.Infof("capture system signal %s, to close \"closing\" channel", <-signals)
	close(closing)
	shutdownChan := make(chan struct{}, 1)
	go func() {
		for _, s := range shutdowns {
			s()
		}
		<-done
		log.Infof("Goroutines exited normally")
		shutdownChan <- struct{}{}
	}()
	select {
	case <-shutdownChan:
		log.Infof("all shutdown jobs done")
	case <-time.After(time.Second * 3):
		log.Infof("Timeout waiting goroutines to exit")
	}

	os.Exit(0)
}

func main() {
	runMode := flag.String("mode", "normal", "The harbor-core container run mode, it could be normal, migrate or skip-migrate, default is normal")
	flag.Parse()

	web.BConfig.WebConfig.Session.SessionOn = true
	web.BConfig.WebConfig.Session.SessionName = config.SessionCookieName
	// the core db used for beego session
	redisCoreURL := os.Getenv("_REDIS_URL_CORE")
	if len(redisCoreURL) > 0 {
		_, err := url.Parse(redisCoreURL)
		if err != nil {
			panic("bad _REDIS_URL_CORE")
		}
		// configure the beego session redis
		web.BConfig.WebConfig.Session.SessionProvider = session.HarborProviderName
		web.BConfig.WebConfig.Session.SessionProviderConfig = redisCoreURL
	}

	log.Info("initializing cache ...")
	// the harbor db used for harbor business, use core db if not specified
	redisHarborURL := os.Getenv("_REDIS_URL_HARBOR")
	if redisHarborURL == "" {
		redisHarborURL = redisCoreURL
	}
	u, err := url.Parse(redisHarborURL)
	if err != nil {
		panic("bad _REDIS_URL_HARBOR")
	}
	if err := cache.Initialize(u.Scheme, redisHarborURL); err != nil {
		log.Fatalf("failed to initialize cache: %v", err)
	}
	// when config/db init function is called, the cache is not ready,
	// enable config cache explicitly when the cache is ready
	dbCfg.EnableConfigCache()

	web.AddTemplateExt("htm")

	log.Info("initializing configurations...")
	config.Init()
	log.Info("configurations initialization completed")

	// default beego max memory and max upload size is 128GB, consider from some AI related image would be large,
	// also support customize it from the environment variables if the default value cannot satisfy some scenarios.
	web.BConfig.MaxMemory = config.GetBeegoMaxMemoryBytes()
	web.BConfig.MaxUploadSize = config.GetBeegoMaxUploadSizeBytes()

	metricCfg := config.Metric()
	if metricCfg.Enabled {
		metric.RegisterCollectors()
		go metric.ServeProm(metricCfg.Path, metricCfg.Port)
	}
	ctx := context.Background()
	config.InitTraceConfig(ctx)
	shutdownTracerProvider := tracelib.InitGlobalTracer(ctx)
	token.InitCreators()
	database, err := config.Database()
	if err != nil {
		log.Fatalf("failed to get database configuration: %v", err)
	}
	if err := dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}
	if strings.EqualFold(*runMode, "migrate") {
		// Used by Harbor helm preinstall, preupgrade hook container
		if err = migration.Migrate(database); err != nil {
			log.Fatalf("failed to migrate the database, error: %v", err)
		}
		log.Info("the database migrate success")
		os.Exit(0)
	} else if strings.EqualFold(*runMode, "skip-migrate") {
		log.Info("skip the database migrate")
	} else {
		// Run migrator as normal
		if err = migration.Migrate(database); err != nil {
			log.Fatalf("failed to migrate the database, error: %v", err)
		}

		log.Info("The database has been migrated successfully")
	}

	ctx = orm.Clone(ctx)
	if err := config.Load(ctx); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	if err := configCtl.Ctl.OverwriteConfig(ctx); err != nil {
		log.Fatalf("failed to init config from CONFIG_OVERWRITE_JSON, error %v", err)
	}
	password, err := config.InitialAdminPassword()
	if err != nil {
		log.Fatalf("failed to get admin's initial password: %v", err)
	}
	if err := updateInitPassword(ctx, adminUserID, password); err != nil {
		log.Error(err)
	}

	// Init API handler
	if err := api.Init(); err != nil {
		log.Fatalf("Failed to initialize API handlers with error: %s", err.Error())
	}

	health.RegisterHealthCheckers()
	registerScanners(orm.Context())

	// start global task pool, do not stop in the gracefulShutdown because it may take long time to finish.
	gtask.DefaultPool().Start(ctx)

	closing := make(chan struct{})
	done := make(chan struct{})
	go gracefulShutdown(closing, done, shutdownTracerProvider)
	// Start health checker for registries
	go registry.Ctl.StartRegularHealthCheck(orm.Context(), closing, done)
	// Init audit log
	auditEP := config.AuditLogForwardEndpoint(ctx)
	audit.LogMgr.Init(ctx, auditEP)

	log.Info("initializing notification...")
	notification.Init()

	server.RegisterRoutes()

	if common_http.InternalTLSEnabled() {
		log.Info("internal TLS enabled, Init TLS ...")
		iTLSKeyPath := os.Getenv("INTERNAL_TLS_KEY_PATH")
		iTLSCertPath := os.Getenv("INTERNAL_TLS_CERT_PATH")

		log.Infof("load client key: %s client cert: %s", iTLSKeyPath, iTLSCertPath)
		web.BConfig.Listen.EnableHTTP = false
		web.BConfig.Listen.EnableHTTPS = true
		web.BConfig.Listen.HTTPSPort = 8443
		web.BConfig.Listen.HTTPSKeyFile = iTLSKeyPath
		web.BConfig.Listen.HTTPSCertFile = iTLSCertPath
		web.BeeApp.Server.TLSConfig = common_http.NewServerTLSConfig()
	}

	log.Infof("Version: %s, Git commit: %s", version.ReleaseVersion, version.GitCommit)

	log.Info("Fix empty subiss for meta info data.")
	_, err = oidc.FixEmptySubIss(orm.Context())
	if err != nil {
		log.Warningf("oidc.FixEmptySubIss() errors out, error: %v", err)
	}
	// Scheduling of system artifact depends on the jobservice, where gorountine is used to avoid the circular
	// dependencies between core and jobservice.
	go func() {
		url := config.InternalJobServiceURL() + "/api/v1/stats"
		checker := health.HTTPStatusCodeHealthChecker(http.MethodGet, url, nil, 60*time.Second, http.StatusOK)
		options := []retry.Option{
			retry.InitialInterval(time.Millisecond * 500),
			retry.MaxInterval(time.Second * 10),
			retry.Timeout(time.Minute * 5),
			retry.Callback(func(err error, sleep time.Duration) {
				log.Debugf("failed to ping %s, retry after %s : %v", url, sleep, err)
			}),
		}
		if err := retry.Retry(checker.Check, options...); err != nil {
			log.Errorf("failed to check the jobservice health status: timeout, error: %v", err)
			return
		}
		// schedule the system jobs with retry as the operation depends on the jobservice,
		// retry to handle the failure case caused by jobservice.
		ctx := orm.Context()
		options = []retry.Option{
			retry.InitialInterval(time.Millisecond * 500),
			retry.MaxInterval(time.Second * 10),
			retry.Timeout(time.Minute * 5),
			retry.Callback(func(err error, sleep time.Duration) {
				log.Debugf("failed to schedule system job, retry after %s : %v", sleep, err)
			}),
		}
		// schedule system artifact cleanup job
		if err := retry.Retry(func() error {
			return systemartifact.ScheduleCleanupTask(ctx)
		}, options...); err != nil {
			log.Errorf("failed to schedule system artifact cleanup job, error: %v", err)
		}
		// schedule system execution sweep job
		if err := retry.Retry(func() error {
			return task.ScheduleSweepJob(ctx)
		}, options...); err != nil {
			log.Errorf("failed to schedule system execution sweep job, error: %v", err)
		}
	}()
	web.RunWithMiddleWares("", middlewares.MiddleWares()...)
}

const (
	trivyScanner = "Trivy"
)

func registerScanners(ctx context.Context) {
	wantedScanners := make([]scanner.Registration, 0)
	uninstallScannerNames := make([]string, 0)

	if config.WithTrivy() {
		log.Info("Registering Trivy scanner")
		wantedScanners = append(wantedScanners, scanner.Registration{
			Name:            trivyScanner,
			Description:     "The Trivy scanner adapter",
			URL:             config.TrivyAdapterURL(),
			UseInternalAddr: true,
			Immutable:       true,
		})
	} else {
		log.Info("Removing Trivy scanner")
		uninstallScannerNames = append(uninstallScannerNames, trivyScanner)
	}

	if err := scan.RemoveImmutableScanners(ctx, uninstallScannerNames); err != nil {
		log.Warningf("failed to remove scanners: %v", err)
	}

	if err := scan.EnsureScanners(ctx, wantedScanners); err != nil {
		log.Fatalf("failed to register scanners: %v", err)
	}

	if defaultScannerName := getDefaultScannerName(); defaultScannerName != "" {
		log.Infof("Setting %s as default scanner", defaultScannerName)
		if err := scan.EnsureDefaultScanner(ctx, defaultScannerName); err != nil {
			log.Fatalf("failed to set default scanner: %v", err)
		}
	}
}

func getDefaultScannerName() string {
	if config.WithTrivy() {
		return trivyScanner
	}
	return ""
}
