// Copyright 2018 The Harbor Authors. All rights reserved.

package runtime

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/vmware/harbor/src/common/job"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/api"
	"github.com/vmware/harbor/src/jobservice_v2/config"
	"github.com/vmware/harbor/src/jobservice_v2/core"
	"github.com/vmware/harbor/src/jobservice_v2/env"
	"github.com/vmware/harbor/src/jobservice_v2/job/impl"
	"github.com/vmware/harbor/src/jobservice_v2/job/impl/scan"
	"github.com/vmware/harbor/src/jobservice_v2/pool"
)

//JobService ...
var JobService = &Bootstrap{}

//Bootstrap is coordinating process to help load and start the other components to serve.
type Bootstrap struct{}

//LoadAndRun will load configurations, initialize components and then start the related process to serve requests.
//Return error if meet any problems.
func (bs *Bootstrap) LoadAndRun(configFile string, detectEnv bool) {
	//Load configurations
	cfg := config.Configuration{}
	if err := cfg.Load(configFile, detectEnv); err != nil {
		log.Errorf("Failed to load configurations with error: %s\n", err)
		return
	}

	//Create the root context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rootContext := &env.Context{
		SystemContext: ctx,
		WG:            &sync.WaitGroup{},
		ErrorChan:     make(chan error, 1), //with 1 buffer
	}

	//Build specified job context
	//TODO:
	jobCtx := impl.NewContext(ctx)
	if err := jobCtx.InitDao(); err != nil {
		log.Errorf("Failed to build job conetxt with error: %s\n", err)
		return
	}
	rootContext.JobContext = jobCtx

	//Start the pool
	var backendPool pool.Interface
	if cfg.PoolConfig.Backend == config.JobServicePoolBackendRedis {
		backendPool = bs.loadAndRunRedisWorkerPool(rootContext, cfg)
	}

	//Initialize controller
	ctl := core.NewController(backendPool)

	//Start the API server
	apiServer := bs.loadAndRunAPIServer(rootContext, cfg, ctl)
	log.Infof("Server is starting at %s:%d with %s", "", cfg.Port, cfg.Protocol)

	//Block here
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, os.Kill)
	select {
	case <-sig:
	case err := <-rootContext.ErrorChan:
		log.Errorf("Server error:%s\n", err)
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

	log.Infof("Server gracefully exit")
}

//Load and run the API server.
func (bs *Bootstrap) loadAndRunAPIServer(ctx *env.Context, cfg config.Configuration, ctl *core.Controller) *api.Server {
	//Initialized API server
	handler := api.NewDefaultHandler(ctl)
	router := api.NewBaseRouter(handler)
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
func (bs *Bootstrap) loadAndRunRedisWorkerPool(ctx *env.Context, cfg config.Configuration) pool.Interface {
	redisPoolCfg := pool.RedisPoolConfig{
		RedisHost:   cfg.PoolConfig.RedisPoolCfg.Host,
		RedisPort:   cfg.PoolConfig.RedisPoolCfg.Port,
		Namespace:   cfg.PoolConfig.RedisPoolCfg.Namespace,
		WorkerCount: cfg.PoolConfig.WorkerCount,
	}

	redisWorkerPool := pool.NewGoCraftWorkPool(ctx, redisPoolCfg)
	//Register jobs here
	if err := redisWorkerPool.RegisterJob(impl.KnownJobReplication, (*impl.ReplicationJob)(nil)); err != nil {
		//exit
		ctx.ErrorChan <- err
		return redisWorkerPool //avoid nil pointer issue
	}
	if err := redisWorkerPool.RegisterJob(job.ImageScanJob, (*scan.ClairJob)(nil)); err != nil {
		//exit
		ctx.ErrorChan <- err
		return redisWorkerPool //avoid nil pointer issue
	}

	redisWorkerPool.Start()

	return redisWorkerPool
}
