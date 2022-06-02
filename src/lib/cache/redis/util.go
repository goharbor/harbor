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

package redis

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/goharbor/harbor/src/lib/errors"
)

// ParseSentinelURL parses sentinel url to redis FailoverOptions.
// It's a modified version of go-redis ParseURL(https://github.com/go-redis/redis/blob/997118894af9d4244d4a471f2b317eead9c9ca62/options.go#L222) because official version does
// not support parse sentinel mode.
func ParseSentinelURL(redisURL string) (*redis.FailoverOptions, error) {
	u, err := url.Parse(redisURL)
	if err != nil {
		return nil, err
	}

	o := &redis.FailoverOptions{}

	o.Username, o.Password = getUserPassword(u)
	o.SentinelAddrs = strings.Split(u.Host, ",")

	f := strings.FieldsFunc(u.Path, func(r rune) bool {
		return r == '/'
	})
	// expect path length is 2, example: [mymaster 1]
	if len(f) != 2 {
		return nil, errors.Errorf("redis: invalid redis URL path: %s", u.Path)
	}

	o.MasterName = f[0]
	if o.DB, err = strconv.Atoi(f[1]); err != nil {
		return nil, errors.Errorf("redis: invalid database number: %q", f[1])
	}

	return setupConnParams(u, o)
}

func getUserPassword(u *url.URL) (string, string) {
	var user, password string
	if u.User != nil {
		user = u.User.Username()
		if p, ok := u.User.Password(); ok {
			password = p
		}
	}
	return user, password
}

type queryOptions struct {
	q   url.Values
	err error
}

func (o *queryOptions) string(name string) string {
	vs := o.q[name]
	if len(vs) == 0 {
		return ""
	}
	delete(o.q, name) // enable detection of unknown parameters
	return vs[len(vs)-1]
}

func (o *queryOptions) int(name string) int {
	s := o.string(name)
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err == nil {
		return i
	}
	if o.err == nil {
		o.err = errors.Errorf("redis: invalid %s number: %s", name, err)
	}
	return 0
}

func (o *queryOptions) duration(name string) time.Duration {
	s := o.string(name)
	if s == "" {
		return 0
	}
	// try plain number first
	if i, err := strconv.Atoi(s); err == nil {
		if i <= 0 {
			// disable timeouts
			return -1
		}
		return time.Duration(i) * time.Second
	}
	dur, err := time.ParseDuration(s)
	if err == nil {
		return dur
	}
	if o.err == nil {
		o.err = fmt.Errorf("redis: invalid %s duration: %w", name, err)
	}
	return 0
}

func (o *queryOptions) bool(name string) bool {
	switch s := o.string(name); s {
	case "true", "1":
		return true
	case "false", "0", "":
		return false
	default:
		if o.err == nil {
			o.err = errors.Errorf("redis: invalid %s boolean: expected true/false/1/0 or an empty string, got %q", name, s)
		}
		return false
	}
}

func (o *queryOptions) remaining() []string {
	if len(o.q) == 0 {
		return nil
	}
	keys := make([]string, 0, len(o.q))
	for k := range o.q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// setupConnParams converts query parameters in u to option value in o.
func setupConnParams(u *url.URL, o *redis.FailoverOptions) (*redis.FailoverOptions, error) {
	q := queryOptions{q: u.Query()}

	// compat: a future major release may use q.int("db")
	if tmp := q.string("db"); tmp != "" {
		db, err := strconv.Atoi(tmp)
		if err != nil {
			return nil, fmt.Errorf("redis: invalid database number: %w", err)
		}
		o.DB = db
	}

	o.MaxRetries = q.int("max_retries")
	o.MinRetryBackoff = q.duration("min_retry_backoff")
	o.MaxRetryBackoff = q.duration("max_retry_backoff")
	o.DialTimeout = q.duration("dial_timeout")
	o.ReadTimeout = q.duration("read_timeout")
	o.WriteTimeout = q.duration("write_timeout")
	o.PoolFIFO = q.bool("pool_fifo")
	o.PoolSize = q.int("pool_size")
	o.MinIdleConns = q.int("min_idle_conns")
	o.MaxConnAge = q.duration("max_conn_age")
	o.PoolTimeout = q.duration("pool_timeout")
	o.IdleTimeout = q.duration("idle_timeout")
	// For compatibility
	if t := q.duration("idle_timeout_seconds"); t != 0 {
		o.IdleTimeout = t
	}
	o.IdleCheckFrequency = q.duration("idle_check_frequency")
	if q.err != nil {
		return nil, q.err
	}

	// any parameters left?
	if r := q.remaining(); len(r) > 0 {
		return nil, errors.Errorf("redis: unexpected option: %s", strings.Join(r, ", "))
	}

	return o, nil
}
