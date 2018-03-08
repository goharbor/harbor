// Copyright 2018 The Harbor Authors. All rights reserved.

package pool

import (
	"errors"
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/gocraft/work"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/jobservice_v2/core"
)

//GoCraftWorkPool is the pool implementation based on gocraft/work powered by redis.
type GoCraftWorkPool struct {
	redisPool *redis.Pool
	pool      *work.WorkerPool
	context   core.BaseContext
}

//RedisPoolConfig defines configurations for GoCraftWorkPool.
type RedisPoolConfig struct {
	RedisHost   string
	RedisPort   uint
	Namespace   string
	WorkerCount uint
}

//NewGoCraftWorkPool is constructor of goCraftWorkPool.
func NewGoCraftWorkPool(ctx core.BaseContext, cfg RedisPoolConfig) *GoCraftWorkPool {
	redisPool := &redis.Pool{
		MaxActive: 5,
		MaxIdle:   5,
		Wait:      true,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", fmt.Sprintf("%s:%d", cfg.RedisHost, cfg.RedisPort))
		},
	}
	pool := work.NewWorkerPool(ctx, cfg.WorkerCount, cfg.Namespace, redisPool)
	return &GoCraftWorkPool{
		redisPool: redisPool,
		pool:      pool,
		context:   ctx,
	}
}

//Start to serve
//Unblock action
func (gcwp *GoCraftWorkPool) Start() error {
	if gcwp.redisPool == nil ||
		gcwp.pool == nil ||
		gcwp.context.SystemContext == nil {
		return errors.New("Redis worker pool can not start as it's not correctly configured")
	}

	go func() {
		defer func() {
			if gcwp.context.WG != nil {
				gcwp.context.WG.Done()
			}
		}()
		gcwp.pool.Start()
		log.Infof("Redis worker pool is started")

		//Block on listening context signal
		<-gcwp.context.SystemContext.Done()
		gcwp.pool.Stop()
		log.Infof("Redis worker pool is stopped")
	}()

	return nil
}
