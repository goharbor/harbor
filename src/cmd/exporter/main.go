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
	"net/http"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" // registry pgx driver
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"

	"github.com/goharbor/harbor/src/common/dao"
	commonthttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/exporter"
)

func main() {
	// Start pprof server
	lib.StartPprof()

	viper.AutomaticEnv()
	viper.SetEnvPrefix("harbor")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	dbCfg := &models.Database{
		Type: "postgresql",
		PostGreSQL: &models.PostGreSQL{
			Host:            viper.GetString("database.host"),
			Port:            viper.GetInt("database.port"),
			Username:        viper.GetString("database.username"),
			Password:        viper.GetString("database.password"),
			Database:        viper.GetString("database.dbname"),
			SSLMode:         viper.GetString("database.sslmode"),
			MaxIdleConns:    viper.GetInt("database.max_idle_conns"),
			MaxOpenConns:    viper.GetInt("database.max_open_conns"),
			ConnMaxLifetime: getConnMaxLifetime(viper.GetString("database.conn_max_lifetime")),
			ConnMaxIdleTime: getConnMaxIdleTime(viper.GetString("database.conn_max_idle_time")),
		},
	}
	if err := dao.InitDatabase(dbCfg); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	exporter.InitHarborClient(&exporter.HarborClient{
		HarborScheme: viper.GetString("service.scheme"),
		HarborHost:   viper.GetString("service.host"),
		HarborPort:   viper.GetInt("service.port"),
		Client: &http.Client{
			Transport: commonthttp.GetHTTPTransport(),
		},
	})

	exporter.InitBackendWorker(&exporter.RedisPoolConfig{
		URL:               viper.GetString("redis.url"),
		Namespace:         viper.GetString("redis.namespace"),
		IdleTimeoutSecond: viper.GetInt("redis.timeout"),
	})

	exporterOpt := &exporter.Opt{
		Port:                   viper.GetInt("exporter.port"),
		MetricsPath:            viper.GetString("exporter.metrics_path"),
		ExporterMetricsEnabled: viper.GetBool("exporter.metrics_enabled"),
		MaxRequests:            viper.GetInt("exporter.max_requests"),
		TLSEnabled:             viper.GetBool("exporter.tls_enabled"),
		Certificate:            viper.GetString("exporter.tls_cert"),
		Key:                    viper.GetString("exporter.tls_key"),
		CacheDuration:          viper.GetInt64("exporter.cache_time"),
		CacheCleanInterval:     viper.GetInt64("exporter.cache_clean_interval"),
	}
	harborExporter := exporter.NewExporter(exporterOpt)
	log.Infof("Starting harbor_exporter with port=%v path=%v metrics=%v max_request=%v tls=%v cert=%v key=%v cache_time=%v clean_interval=%v",
		exporterOpt.Port,
		exporterOpt.MetricsPath,
		exporterOpt.ExporterMetricsEnabled,
		exporterOpt.MaxRequests,
		exporterOpt.TLSEnabled,
		exporterOpt.Certificate,
		exporterOpt.Key,
		exporterOpt.CacheDuration,
		exporterOpt.CacheCleanInterval,
	)
	prometheus.MustRegister(harborExporter)
	if err := harborExporter.ListenAndServe(); err != nil {
		log.Errorf("Error starting Harbor exporter %s", err)
		os.Exit(1)
	}
}

func getConnMaxLifetime(duration string) time.Duration {
	// set conn max life time, 5m by default
	connMaxLifetime := 5 * time.Minute
	if duration != "" {
		maxLifetime, err := time.ParseDuration(duration)
		if err == nil {
			connMaxLifetime = maxLifetime
		} else {
			log.Warningf("Failed to parse database.conn_max_lifetime, use default value: %s, err: %v", connMaxLifetime, err)
		}
	}

	return connMaxLifetime
}

func getConnMaxIdleTime(duration string) time.Duration {
	// set conn max idle time, 0 by default
	connMaxIdleTime := time.Duration(0)
	if duration != "" {
		maxIdleTime, err := time.ParseDuration(duration)
		if err == nil {
			connMaxIdleTime = maxIdleTime
		} else {
			log.Warningf("Failed to parse database.conn_max_idle_time, use default value: %s, err: %v", connMaxIdleTime, err)
		}
	}

	return connMaxIdleTime
}
