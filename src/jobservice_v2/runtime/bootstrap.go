// Copyright 2018 The Harbor Authors. All rights reserved.

package runtime

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/api"
	"github.com/vmware/harbor/src/jobservice_v2/config"
	"github.com/vmware/harbor/src/jobservice_v2/core"
	"github.com/vmware/harbor/src/jobservice_v2/pool"
)

//JobService ...
var JobService = &Bootstrap{}

//Bootstrap is coordinating process to help load and start the other components to serve.
type Bootstrap struct {
	apiServer  *api.Server
	workerPool pool.Interface
}

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
	rootContext := core.BaseContext{
		SystemContext: ctx,
		WG:            &sync.WaitGroup{},
	}

	//Start the pool
	if cfg.PoolConfig.Backend == config.JobServicePoolBackendRedis {
		if err := bs.loadAndRunRedisWorkerPool(rootContext, cfg); err != nil {
			log.Errorf("Failed to start the redis worker pool with error: %s\n", err)
			return
		}
		rootContext.WG.Add(1)
	}

	//Initialize controller
	ctl := core.NewController()

	//Start the API server
	bs.loadAndRunAPIServer(rootContext, cfg, ctl)
	rootContext.WG.Add(1)
	log.Infof("Server is starting at %s:%d with %s", "", cfg.Port, cfg.Protocol)

	//Block here
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM, os.Kill)
	<-sig

	//Call cancel to send termination signal to other interested parts.
	cancel()

	//Gracefully shutdown
	bs.apiServer.Stop()

	rootContext.WG.Wait()
	log.Infof("Server gracefully exit")
}

//Load and run the API server.
func (bs *Bootstrap) loadAndRunAPIServer(ctx core.BaseContext, cfg config.Configuration, ctl *core.Controller) {
	//Initialized API server
	handler := api.NewDefaultHandler(ctx, ctl)
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
	bs.apiServer = server

	//Start processes
	server.Start()
}

//Load and run the worker pool
func (bs *Bootstrap) loadAndRunRedisWorkerPool(ctx core.BaseContext, cfg config.Configuration) error {
	redisPoolCfg := pool.RedisPoolConfig{
		RedisHost:   cfg.PoolConfig.RedisPoolCfg.Host,
		RedisPort:   cfg.PoolConfig.RedisPoolCfg.Port,
		Namespace:   cfg.PoolConfig.RedisPoolCfg.Namespace,
		WorkerCount: cfg.PoolConfig.WorkerCount,
	}

	redisWorkerPool := pool.NewGoCraftWorkPool(ctx, redisPoolCfg)
	bs.workerPool = redisWorkerPool
	return redisWorkerPool.Start()
}
