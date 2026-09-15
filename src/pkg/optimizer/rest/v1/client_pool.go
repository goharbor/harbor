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

package v1

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	defaultDeadCheckInterval = 1 * time.Minute
	defaultExpireTime        = 5 * time.Minute
)

// DefaultClientPool is a default client pool.
var DefaultClientPool = NewClientPool(nil)

// ClientPool defines operations for the client pool which provides v1 client cache.
type ClientPool interface {
	// Get a v1 client interface for the specified registration.
	Get(url, authType, accessCredential string, skipCertVerify bool) (Client, error)
}

// PoolConfig provides configurations for the client pool.
type PoolConfig struct {
	// Interval for checking dead instance.
	DeadCheckInterval time.Duration
	// Expire time for the instance to be marked as dead.
	ExpireTime time.Duration
}

// poolItem append timestamp for the caching client instance.
type poolItem struct {
	c         Client
	timestamp time.Time
}

// basicClientPool is default implementation of client pool interface.
type basicClientPool struct {
	pool   *sync.Map
	config *PoolConfig
}

// NewClientPool news a basic client pool.
func NewClientPool(config *PoolConfig) ClientPool {
	bcp := &basicClientPool{
		pool:   &sync.Map{},
		config: config,
	}

	if bcp.config == nil {
		bcp.config = &PoolConfig{}
	}

	if bcp.config.DeadCheckInterval == 0 {
		bcp.config.DeadCheckInterval = defaultDeadCheckInterval
	}

	if bcp.config.ExpireTime == 0 {
		bcp.config.ExpireTime = defaultExpireTime
	}

	return bcp
}

// Get client for the specified registration.
func (bcp *basicClientPool) Get(url, authType, accessCredential string, skipCertVerify bool) (Client, error) {
	k := fmt.Sprintf("%s:%s:%s:%v", url, authType, accessCredential, skipCertVerify)

	item, ok := bcp.pool.Load(k)
	if !ok {
		nc, err := NewClient(url, authType, accessCredential, skipCertVerify)
		if err != nil {
			return nil, errors.Wrap(err, "optimizer client pool: get")
		}

		npi := &poolItem{
			c:         nc,
			timestamp: time.Now().UTC(),
		}

		bcp.pool.Store(k, npi)
		item = npi

		bcp.deadCheck(k, npi)
	}

	return item.(*poolItem).c, nil
}

func (bcp *basicClientPool) deadCheck(key string, item *poolItem) {
	// Run in a separate goroutine; watch the system signal to exit correctly.
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		tk := time.NewTicker(bcp.config.DeadCheckInterval)
		defer tk.Stop()

		for {
			select {
			case t := <-tk.C:
				if item.timestamp.Add(bcp.config.ExpireTime).Before(t.UTC()) {
					bcp.pool.Delete(key)
					return
				}
			case <-sig:
				return
			}
		}
	}()
}
