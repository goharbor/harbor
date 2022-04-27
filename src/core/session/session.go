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
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/session"
	goredis "github.com/go-redis/redis/v8"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/cache/redis"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// HarborProviderName is the harbor session provider name
	HarborProviderName = "harbor"
)

var harborpder = &Provider{}

// SessionStore redis session store
type SessionStore struct {
	c           cache.Cache
	sid         string
	lock        sync.RWMutex
	values      map[interface{}]interface{}
	maxlifetime int64
}

// Set value in redis session
func (rs *SessionStore) Set(key, value interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values[key] = value
	return nil
}

// Get value in redis session
func (rs *SessionStore) Get(key interface{}) interface{} {
	rs.lock.RLock()
	defer rs.lock.RUnlock()
	if v, ok := rs.values[key]; ok {
		return v
	}
	return nil
}

// Delete value in redis session
func (rs *SessionStore) Delete(key interface{}) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	delete(rs.values, key)
	return nil
}

// Flush clear all values in redis session
func (rs *SessionStore) Flush() error {
	rs.lock.Lock()
	defer rs.lock.Unlock()
	rs.values = make(map[interface{}]interface{})
	return nil
}

// SessionID get redis session id
func (rs *SessionStore) SessionID() string {
	return rs.sid
}

// SessionRelease save session values to redis
func (rs *SessionStore) SessionRelease(w http.ResponseWriter) {
	b, err := session.EncodeGob(rs.values)
	if err != nil {
		return
	}

	if rdb, ok := rs.c.(*redis.Cache); ok {
		cmd := rdb.Client.Set(context.TODO(), rs.sid, string(b), time.Duration(rs.maxlifetime))
		if cmd.Err() != nil {
			log.Debugf("release session error: %v", err)
		}
	}
}

// Provider redis session provider
type Provider struct {
	maxlifetime int64
	c           cache.Cache
}

// SessionInit init redis session
func (rp *Provider) SessionInit(maxlifetime int64, url string) (err error) {
	rp.maxlifetime = maxlifetime * int64(time.Second)
	rp.c, err = redis.New(cache.Options{Address: url, Codec: codec})
	if err != nil {
		return err
	}

	return rp.c.Ping(context.TODO())
}

// SessionRead read redis session by sid
func (rp *Provider) SessionRead(sid string) (session.Store, error) {
	kv := make(map[interface{}]interface{})
	err := rp.c.Fetch(context.TODO(), sid, &kv)
	if err != nil && !strings.Contains(err.Error(), goredis.Nil.Error()) {
		return nil, err
	}

	rs := &SessionStore{c: rp.c, sid: sid, values: kv, maxlifetime: rp.maxlifetime}
	return rs, nil
}

// SessionExist check redis session exist by sid
func (rp *Provider) SessionExist(sid string) bool {
	return rp.c.Contains(context.TODO(), sid)
}

// SessionRegenerate generate new sid for redis session
func (rp *Provider) SessionRegenerate(oldsid, sid string) (session.Store, error) {
	ctx := context.TODO()
	if !rp.SessionExist(oldsid) {
		rp.c.Save(ctx, sid, "", time.Duration(rp.maxlifetime))
	} else {
		if rdb, ok := rp.c.(*redis.Cache); ok {
			// redis has rename command
			rdb.Rename(ctx, oldsid, sid)
			rdb.Expire(ctx, sid, time.Duration(rp.maxlifetime))
		} else {
			kv := make(map[interface{}]interface{})
			err := rp.c.Fetch(ctx, sid, &kv)
			if err != nil && !strings.Contains(err.Error(), goredis.Nil.Error()) {
				return nil, err
			}

			rp.c.Delete(ctx, oldsid)
			rp.c.Save(ctx, sid, kv)
		}
	}

	return rp.SessionRead(sid)
}

// SessionDestroy delete redis session by id
func (rp *Provider) SessionDestroy(sid string) error {
	return rp.c.Delete(context.TODO(), sid)
}

// SessionGC Implement method, no used.
func (rp *Provider) SessionGC() {
}

// SessionAll return all activeSession
func (rp *Provider) SessionAll() int {
	return 0
}

func init() {
	session.Register(HarborProviderName, harborpder)
}
