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

package session

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/beego/beego/v2/server/web/session"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/cache/redis"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// HarborProviderName is the harbor session provider name
	HarborProviderName = "harbor"
)

var harborpder = &Provider{}

// Store redis session store
type Store struct {
	c           cache.Cache
	sid         string
	lock        sync.RWMutex
	values      map[any]any
	maxlifetime int64
}

// Set value in redis session
func (rs *Store) Set(_ context.Context, key, value any) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values[key] = value
	return nil
}

// Get value in redis session
func (rs *Store) Get(_ context.Context, key any) any {
	rs.lock.RLock()
	defer rs.lock.RUnlock()
	if v, ok := rs.values[key]; ok {
		return v
	}
	return nil
}

// Delete value in redis session
func (rs *Store) Delete(_ context.Context, key any) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	delete(rs.values, key)
	return nil
}

// Flush clear all values in redis session
func (rs *Store) Flush(_ context.Context) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values = make(map[any]any)
	return nil
}

// SessionID get redis session id
func (rs *Store) SessionID(_ context.Context) string {
	return rs.sid
}

func (rs *Store) releaseSession(ctx context.Context, _ http.ResponseWriter, requirePresent bool) {
	rs.lock.RLock()
	values := rs.values
	rs.lock.RUnlock()
	b, err := session.EncodeGob(values)
	if err != nil {
		return
	}
	if ctx == nil {
		ctx = context.TODO()
	}
	maxlifetime := time.Duration(systemSessionTimeout(ctx, rs.maxlifetime))
	if rdb, ok := rs.c.(*redis.Cache); ok {
		if requirePresent {
			cmd := rdb.Client.SetXX(ctx, rs.sid, string(b), maxlifetime)
			if cmd.Err() != nil {
				log.Debugf("release session error: %v", err)
			}
		} else {
			cmd := rdb.Client.Set(ctx, rs.sid, string(b), maxlifetime)
			if cmd.Err() != nil {
				log.Debugf("release session error: %v", err)
			}
		}
	}
}

// SessionRelease save session values to redis
func (rs *Store) SessionRelease(ctx context.Context, w http.ResponseWriter) {
	rs.releaseSession(ctx, w, false)
}

// added by beego version v2.3.4, commit https://github.com/beego/beego/commit/06d869664a9c55aea6c2bb6ac3866f8a39b1100c#diff-bc81cfdba9f5250f9bf95ccaae2e4e34b37af87e2091dda11ef49dc58bd91c2c
// SessionReleaseIfPresent save session values to redis when key is present
func (rs *Store) SessionReleaseIfPresent(ctx context.Context, w http.ResponseWriter) {
	rs.releaseSession(ctx, w, true)
}

// Provider redis session provider
type Provider struct {
	maxlifetime int64
	c           cache.Cache
}

// SessionInit init redis session
func (rp *Provider) SessionInit(ctx context.Context, maxlifetime int64, url string) (err error) {
	rp.maxlifetime = maxlifetime * int64(time.Second)
	rp.c, err = redis.New(cache.Options{Address: url, Codec: codec})
	if err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.TODO()
	}
	return rp.c.Ping(ctx)
}

// SessionRead read redis session by sid
func (rp *Provider) SessionRead(ctx context.Context, sid string) (session.Store, error) {
	kv := make(map[any]any)
	if ctx == nil {
		ctx = context.TODO()
	}
	err := rp.c.Fetch(ctx, sid, &kv)
	if err != nil && !errors.Is(err, cache.ErrNotFound) {
		return nil, err
	}

	rs := &Store{c: rp.c, sid: sid, values: kv, maxlifetime: rp.maxlifetime}
	return rs, nil
}

// SessionExist check redis session exist by sid
func (rp *Provider) SessionExist(ctx context.Context, sid string) (bool, error) {
	if ctx == nil {
		ctx = context.TODO()
	}
	return rp.c.Contains(ctx, sid), nil
}

// SessionRegenerate generate new sid for redis session
func (rp *Provider) SessionRegenerate(ctx context.Context, oldsid, sid string) (session.Store, error) {
	if ctx == nil {
		ctx = context.TODO()
	}
	maxlifetime := time.Duration(systemSessionTimeout(ctx, rp.maxlifetime))
	if isExist, _ := rp.SessionExist(ctx, oldsid); !isExist {
		err := rp.c.Save(ctx, sid, map[any]any{}, time.Duration(rp.maxlifetime))
		if err != nil {
			log.Debugf("failed to save sid=%s, where oldsid=%s, error: %s", sid, oldsid, err)
		}
	} else {
		if rdb, ok := rp.c.(*redis.Cache); ok {
			// redis has rename command
			rdb.Rename(ctx, oldsid, sid)
			rdb.Expire(ctx, sid, maxlifetime)
		} else {
			kv := make(map[any]any)
			err := rp.c.Fetch(ctx, sid, &kv)
			if err != nil && !errors.Is(err, cache.ErrNotFound) {
				return nil, err
			}

			err = rp.c.Delete(ctx, oldsid)
			if err != nil {
				log.Debugf("failed to delete oldsid=%s, error: %s", oldsid, err)
			}
			err = rp.c.Save(ctx, sid, kv, maxlifetime)
			if err != nil {
				log.Debugf("failed to save sid=%s, error: %s", sid, err)
			}
		}
	}

	return rp.SessionRead(ctx, sid)
}

// SessionDestroy delete redis session by id
func (rp *Provider) SessionDestroy(ctx context.Context, sid string) error {
	if ctx == nil {
		ctx = context.TODO()
	}
	return rp.c.Delete(ctx, sid)
}

// SessionGC Implement method, no used.
func (rp *Provider) SessionGC(_ context.Context) {
}

// SessionAll return all activeSession
func (rp *Provider) SessionAll(_ context.Context) int {
	return 0
}

// systemSessionTimeout return the system session timeout set by user.
func systemSessionTimeout(ctx context.Context, beegoTimeout int64) int64 {
	// read from system config if it is meaningful to support change session timeout in runtime for user.
	// otherwise, use parameters beegoTimeout which set from beego.
	timeout := beegoTimeout
	if sysTimeout := config.SessionTimeout(ctx); sysTimeout > 0 {
		timeout = sysTimeout * int64(time.Minute)
	}

	return timeout
}

func init() {
	session.Register(HarborProviderName, harborpder)
}
