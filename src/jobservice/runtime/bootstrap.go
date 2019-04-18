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

package runtime

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/goharbor/harbor/src/jobservice/api"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/core"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/hook"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/job/impl/gc"
	"github.com/goharbor/harbor/src/jobservice/job/impl/replication"
	"github.com/goharbor/harbor/src/jobservice/job/impl/sample"
	"github.com/goharbor/harbor/src/jobservice/job/impl/scan"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/worker"
	"github.com/goharbor/harbor/src/jobservice/worker/cworker"
	"github.com/gomodule/redigo/redis"
)

const (
	dialConnectionTimeout = 30 * time.Second
	healthCheckPeriod     = time.Minute
	dialReadTimeout       = healthCheckPeriod + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
)

// JobService ...
var JobService = &Bootstrap{}

// Bootstrap is coordinating process to help load and start the other components to serve.
type Bootstrap struct {
	jobConextInitializer job.JobContextInitializer
}

// SetJobContextInitializer set the job context initializer
func (bs *Bootstrap) SetJobContextInitializer(initializer job.JobContextInitializer) {
	if initializer != nil {
		bs.jobConextInitializer = initializer
	}
}

// LoadAndRun will load configurations, initialize components and then start the related process to serve requests.
// Return error if meet any problems.
func (bs *Bootstrap) LoadAndRun(ctx context.Context) {
	rootContext := &env.Context{
		SystemContext: ctx,
		WG:            &sync.WaitGroup{},
		ErrorChan:     make(chan error, 3), // with 3 buffers
	}

	// Build specified job context
	if bs.jobConextInitializer != nil {
		if jobCtx, err := bs.jobConextInitializer(ctx); err == nil {
			rootContext.JobContext = jobCtx
		} else {
			logger.Fatalf("Failed to initialize job context: %s\n", err)
		}
	}

	// Alliance to config
	cfg := config.DefaultConfig

	var (
		backendWorker worker.Interface
		lcmCtl        lcm.Controller
		wErr          error
	)
	if cfg.PoolConfig.Backend == config.JobServicePoolBackendRedis {
		// Number of workers
		workerNum := cfg.PoolConfig.WorkerCount
		// Add {} to namespace to void slot issue
		namespace := fmt.Sprintf("{%s}", cfg.PoolConfig.RedisPoolCfg.Namespace)
		// Get redis connection pool
		redisPool := bs.getRedisPool(cfg.PoolConfig.RedisPoolCfg.RedisURL)
		// Create hook agent, it's a singleton object
		hookAgent := hook.NewAgent(ctx, namespace, redisPool)
		hookCallback := func(URL string, change *job.StatusChange) error {
			msg := fmt.Sprintf("status change: job=%s, status=%s", change.JobID, change.Status)
			if !utils.IsEmptyStr(change.CheckIn) {
				msg = fmt.Sprintf("%s, check_in=%s", msg, change.CheckIn)
			}

			evt := &hook.Event{
				URL:       URL,
				Timestamp: time.Now().Unix(),
				Data:      change,
				Message:   msg,
			}

			return hookAgent.Trigger(evt)
		}

		// Create job life cycle management controller
		lcmCtl = lcm.NewController(ctx, namespace, redisPool, hookCallback)

		// Start the backend worker
		backendWorker, wErr = bs.loadAndRunRedisWorkerPool(rootContext, namespace, workerNum, redisPool, lcmCtl)
		if wErr != nil {
			logger.Fatalf("Failed to load and run worker worker: %s\n", wErr.Error())
		}

		// Start agent
		// Non blocking call
		hookAgent.Serve()
	} else {
		logger.Fatalf("Worker worker backend '%s' is not supported", cfg.PoolConfig.Backend)
	}

	// Initialize controller
	ctl := core.NewController(backendWorker, lcmCtl)
	// Start the API server
	apiServer := bs.createAPIServer(ctx, cfg, ctl)

	// Listen to the system signals
	go func(errChan chan error) {
		defer func() {
			// Gracefully shutdown
			if err := apiServer.Stop(); err != nil {
				logger.Error(err)
			}
		}()
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, os.Kill)

		select {
		case <-sig:
			return
		case err := <-errChan:
			logger.Errorf("error received from error chan: %s", err)
			return
		}
	}(rootContext.ErrorChan)

	// Blocking here
	logger.Infof("API server is serving at %d with %s mode", cfg.Port, cfg.Protocol)
	if err := apiServer.Start(); err != nil {
		logger.Errorf("API server error: %s", err)
	} else {
		logger.Info("API server is gracefully shut down")
	}
}

// Load and run the API server.
func (bs *Bootstrap) createAPIServer(ctx context.Context, cfg *config.Configuration, ctl core.Interface) *api.Server {
	// Initialized API server
	authProvider := &api.SecretAuthenticator{}
	handler := api.NewDefaultHandler(ctl)
	router := api.NewBaseRouter(handler, authProvider)
	serverConfig := api.ServerConfig{
		Protocol: cfg.Protocol,
		Port:     cfg.Port,
	}
	if cfg.HTTPSConfig != nil {
		serverConfig.Cert = cfg.HTTPSConfig.Cert
		serverConfig.Key = cfg.HTTPSConfig.Key
	}

	return api.NewServer(ctx, router, serverConfig)
}

// Load and run the worker worker
func (bs *Bootstrap) loadAndRunRedisWorkerPool(
	ctx *env.Context,
	ns string,
	workers uint,
	redisPool *redis.Pool,
	lcmCtl lcm.Controller,
) (worker.Interface, error) {
	redisWorker := cworker.NewWorker(ctx, ns, workers, redisPool, lcmCtl)
	// Register jobs here
	if err := redisWorker.RegisterJobs(
		map[string]interface{}{
			// Only for debugging and testing purpose
			job.SampleJob: (*sample.Job)(nil),
			// Functional jobs
			job.ImageScanJob:    (*scan.ClairJob)(nil),
			job.ImageScanAllJob: (*scan.All)(nil),
			job.ImageTransfer:   (*replication.Transfer)(nil),
			job.ImageDelete:     (*replication.Deleter)(nil),
			job.ImageReplicate:  (*replication.Replicator)(nil),
			job.ImageGC:         (*gc.GarbageCollector)(nil),
		}); err != nil {
		// exit
		return nil, err
	}

	if err := redisWorker.Start(); err != nil {
		return nil, err
	}

	return redisWorker, nil
}

// Get a redis connection pool
func (bs *Bootstrap) getRedisPool(redisURL string) *redis.Pool {
	return &redis.Pool{
		MaxActive: 6,
		MaxIdle:   6,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(
				redisURL,
				redis.DialConnectTimeout(dialConnectionTimeout),
				redis.DialReadTimeout(dialReadTimeout),
				redis.DialWriteTimeout(dialWriteTimeout),
			)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}

			_, err := c.Do("PING")
			return err
		},
	}
}
