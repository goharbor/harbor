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

	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/pkg/errors"
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
	//
	//  Arguments:
	//   r *scanner.Registration : registration for client connecting to
	//
	//  Returns:
	//   Client : v1 client
	//   error  : non nil error if any errors occurred
	Get(r *scanner.Registration) (Client, error)
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

	// Set config
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
// So far, there will not be too many scanner registrations. An then
// no need to do client instance clear work.
// If one day, we have to clear unactivated client instances in the pool,
// add the following func after the first time initializing the client.
// pool item represents the client with a timestamp of last accessed.

func (bcp *basicClientPool) Get(r *scanner.Registration) (Client, error) {
	if r == nil {
		return nil, errors.New("nil scanner registration")
	}

	if err := r.Validate(false); err != nil {
		return nil, errors.Wrap(err, "client pool: get")
	}

	k := key(r)

	item, ok := bcp.pool.Load(k)
	if !ok {
		nc, err := NewClient(r)
		if err != nil {
			return nil, errors.Wrap(err, "client pool: get")
		}

		// Cache it
		npi := &poolItem{
			c:         nc,
			timestamp: time.Now().UTC(),
		}

		bcp.pool.Store(k, npi)
		item = npi

		// dead check
		bcp.deadCheck(k, npi)
	}

	return item.(*poolItem).c, nil
}

func (bcp *basicClientPool) deadCheck(key string, item *poolItem) {
	// Run in a separate goroutine
	go func() {
		// As we do not have a global context, let's watch the system signal to
		// exit the goroutine correctly.
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, os.Kill)

		tk := time.NewTicker(bcp.config.DeadCheckInterval)
		defer tk.Stop()

		for {
			select {
			case t := <-tk.C:
				if item.timestamp.Add(bcp.config.ExpireTime).Before(t.UTC()) {
					// Expired
					bcp.pool.Delete(key)
					return
				}
			case <-sig:
				// Terminated by system
				return
			}
		}
	}()
}

func key(r *scanner.Registration) string {
	return fmt.Sprintf("%s:%s:%s:%v",
		r.URL,
		r.Auth,
		r.AccessCredential,
		r.SkipCertVerify,
	)
}
