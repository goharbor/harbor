// Copyright 2018 The Harbor Authors. All rights reserved.

package runtime

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/vmware/harbor/src/common/job"
	"github.com/vmware/harbor/src/jobservice/api"
	"github.com/vmware/harbor/src/jobservice/config"
	"github.com/vmware/harbor/src/jobservice/core"
	"github.com/vmware/harbor/src/jobservice/env"
	"github.com/vmware/harbor/src/jobservice/job/impl"
	"github.com/vmware/harbor/src/jobservice/job/impl/replication"
	"github.com/vmware/harbor/src/jobservice/job/impl/scan"
	"github.com/vmware/harbor/src/jobservice/logger"
	"github.com/vmware/harbor/src/jobservice/pool"
)

const (
	dialConnectionTimeout = 30 * time.Second
	healthCheckPeriod     = time.Minute
	dialReadTimeout       = healthCheckPeriod + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
)

//JobService ...
var JobService = &Bootstrap{}

//Bootstrap is coordinating process to help load and start the other components to serve.
type Bootstrap struct {
	jobConextInitializer env.JobContextInitializer
}

//SetJobContextInitializer set the job context initializer
func (bs *Bootstrap) SetJobContextInitializer(initializer env.JobContextInitializer) {
	if initializer != nil {
		bs.jobConextInitializer = initializer
	}
}

//LoadAndRun will load configurations, initialize components and then start the related process to serve requests.
//Return error if meet any problems.
func (bs *Bootstrap) LoadAndRun() {
	//Create the root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootContext := &env.Context{
		SystemContext: ctx,
		WG:            &sync.WaitGroup{},
		ErrorChan:     make(chan error, 1), //with 1 buffer
	}

	//Build specified job context
	if bs.jobConextInitializer != nil {
		if jobCtx, err := bs.jobConextInitializer(rootContext); err == nil {
			rootContext.JobContext = jobCtx
		} else {
			logger.Fatalf("Failed to initialize job context: %s\n", err)
		}
	}

	//Start the pool
	var (
		backendPool pool.Interface
		wpErr       error
	)
	if config.DefaultConfig.PoolConfig.Backend == config.JobServicePoolBackendRedis {
		backendPool, wpErr = bs.loadAndRunRedisWorkerPool(rootContext, config.DefaultConfig)
		if wpErr != nil {
			logger.Fatalf("Failed to load and run worker pool: %s\n", wpErr.Error())
		}
	} else {
		logger.Fatalf("Worker pool backend '%s' is not supported", config.DefaultConfig.PoolConfig.Backend)
	}

	//Initialize controller
	ctl := core.NewController(backendPool)
	//Start the API server
	apiServer := bs.loadAndRunAPIServer(rootContext, config.DefaultConfig, ctl)
	logger.Infof("Server is started at %s:%d with %s", "", config.DefaultConfig.Port, config.DefaultConfig.Protocol)

	//Start outdated log files sweeper
	logSweeper := logger.NewSweeper(ctx, config.GetLogBasePath(), config.GetLogArchivePeriod())
	logSweeper.Start()

	//Block here
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, os.Kill)
	select {
	case <-sig:
	case err := <-rootContext.ErrorChan:
		logger.Errorf("Server error:%s\n", err)
	}

	//Call cancel to send termination signal to other interested parts.
	cancel()

	//Gracefully shutdown
	apiServer.Stop()

	//In case stop is called before the server is ready
	close := make(chan bool, 1)
	go func() {
		timer := time.NewTimer(10 * time.Second)
		defer timer.Stop()

		select {
		case <-timer.C:
			//Try again
			apiServer.Stop()
		case <-close:
			return
		}

	}()

	rootContext.WG.Wait()
	close <- true

	logger.Infof("Server gracefully exit")
}

//Load and run the API server.
func (bs *Bootstrap) loadAndRunAPIServer(ctx *env.Context, cfg *config.Configuration, ctl *core.Controller) *api.Server {
	//Initialized API server
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

	server := api.NewServer(ctx, router, serverConfig)
	//Start processes
	server.Start()

	return server
}

//Load and run the worker pool
func (bs *Bootstrap) loadAndRunRedisWorkerPool(ctx *env.Context, cfg *config.Configuration) (pool.Interface, error) {
	redisPool := &redis.Pool{
		MaxActive: 6,
		MaxIdle:   6,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.DialURL(
				cfg.PoolConfig.RedisPoolCfg.RedisURL,
				redis.DialConnectTimeout(dialConnectionTimeout),
				redis.DialReadTimeout(dialReadTimeout),
				redis.DialWriteTimeout(dialWriteTimeout),
			)
		},
	}

	redisWorkerPool := pool.NewGoCraftWorkPool(ctx,
		cfg.PoolConfig.RedisPoolCfg.Namespace,
		cfg.PoolConfig.WorkerCount,
		redisPool)
	//Register jobs here
	if err := redisWorkerPool.RegisterJob(impl.KnownJobDemo, (*impl.DemoJob)(nil)); err != nil {
		//exit
		return nil, err
	}
	if err := redisWorkerPool.RegisterJobs(
		map[string]interface{}{
			job.ImageScanJob:   (*scan.ClairJob)(nil),
			job.ImageTransfer:  (*replication.Transfer)(nil),
			job.ImageDelete:    (*replication.Deleter)(nil),
			job.ImageReplicate: (*replication.Replicator)(nil),
		}); err != nil {
		//exit
		return nil, err
	}

	if err := redisWorkerPool.Start(); err != nil {
		return nil, err
	}

	return redisWorkerPool, nil
}
