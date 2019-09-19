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
	"encoding/base64"
	"fmt"
	"sync"

	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/pkg/errors"
)

// DefaultClientPool is a default client pool.
var DefaultClientPool = NewClientPool()

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

// basicClientPool is default implementation of client pool interface.
type basicClientPool struct {
	pool *sync.Map
}

// NewClientPool news a basic client pool.
func NewClientPool() ClientPool {
	return &basicClientPool{
		pool: &sync.Map{},
	}
}

// Get client for the specified registration.
// So far, there will not be too many scanner registrations. An then
// no need to do client instance clear work.
// If one day, we have to clear unactivated client instances in the pool,
// add the following func after the first time initializing the client.
// pool item represents the client with a timestamp of last accessed.
//
// type poolItem struct {
// 	 c         Client
// 	 timestamp time.Time
// }
//
// func (bcp *basicClientPool) deadCheck(key string, item *poolItem) {
//	// Run in a separate goroutine
//	go func() {
//		// As we do not have a global context, let's watch the system signal to
//		// exit the goroutine correctly.
//		sig := make(chan os.Signal, 1)
//		signal.Notify(sig, os.Interrupt, syscall.SIGTERM, os.Kill)
//
//		tk := time.NewTicker(bcp.config.DeadCheckInterval)
//		defer tk.Stop()
//
//		for {
//			select {
//			case t := <-tk.C:
//				if item.timestamp.Add(bcp.config.ExpireTime).Before(t.UTC()) {
//					// Expired
//					bcp.pool.Delete(key)
//					return
//				}
//			case <-sig:
//				// Terminated by system
//				return
//			}
//		}
//	}()
// }
func (bcp *basicClientPool) Get(r *scanner.Registration) (Client, error) {
	if r == nil {
		return nil, errors.New("nil scanner registration")
	}

	if err := r.Validate(true); err != nil {
		return nil, errors.Wrap(err, "client pool: get")
	}

	k := key(r)

	c, ok := bcp.pool.Load(k)
	if !ok {
		nc, err := NewClient(r)
		if err != nil {
			return nil, errors.Wrap(err, "client pool: get")
		}

		// Cache it
		bcp.pool.Store(k, nc)
		c = nc
	}

	return c.(Client), nil
}

func key(r *scanner.Registration) string {
	raw := fmt.Sprintf("%s:%s:%s:%s:%v",
		r.UUID,
		r.URL,
		r.Auth,
		r.AccessCredential,
		r.SkipCertVerify,
	)

	return base64.StdEncoding.EncodeToString([]byte(raw))
}
