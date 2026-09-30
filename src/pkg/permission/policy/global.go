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

package policy

import (
	"context"

	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

var def atomic.Pointer[Store]

// Default is the store this process answers permission questions from.
//
// It is never nil. Before Init has run it is a store with no policy, which
// grants nothing -- a component that has not been given a database must not
// accidentally allow something.
func Default() *Store {
	if s := def.Load(); s != nil {
		return s
	}
	s, err := Empty()
	if err != nil {
		log.Fatalf("failed to build the empty policy store: %v", err)
	}
	if !def.CompareAndSwap(nil, s) {
		return def.Load()
	}
	return s
}

// SetDefault installs the process-wide store. It returns the previous one, so
// a test can put it back.
func SetDefault(s *Store) *Store { return def.Swap(s) }

// Init opens the policy store against Harbor's database and starts watching
// for changes made by the other replicas.
//
// origin identifies this replica in the notifications it causes, so it can
// skip its own writes instead of reloading for nothing.
func Init(ctx context.Context, cfg *models.PostGreSQL, origin string) error {
	conn, err := connString(cfg)
	if err != nil {
		return err
	}

	pgxCfg, err := pgx.ParseConfig(conn)
	if err != nil {
		return errors.Wrap(err, "failed to parse the database connection string")
	}
	// Rides along on every connection this store opens, so the trigger can
	// stamp a notification with the replica that caused it.
	pgxCfg.RuntimeParams["options"] = "-c harbor.origin=" + origin

	db := stdlib.OpenDB(*pgxCfg)
	db.SetMaxOpenConns(maxStoreConns)
	db.SetMaxIdleConns(2)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return errors.Wrap(err, "failed to reach the database")
	}

	s, err := New(ctx, db, origin)
	if err != nil {
		_ = db.Close()
		return err
	}
	// The roles Harbor ships used to carry their grants in the binary, so the
	// first start after this change is the one that writes them down.
	if err := s.EnsureSeeded(ctx); err != nil {
		s.Close()
		_ = db.Close()
		return err
	}
	if err := s.Watch(conn); err != nil {
		s.Close()
		_ = db.Close()
		return err
	}

	SetDefault(s)
	log.Infof("policy store ready as %q, generation %d", origin, s.Generation())
	return nil
}

// maxStoreConns bounds the pool the store keeps for itself. It reads the whole
// policy on a change and nothing on a request, so it needs very little.
const maxStoreConns = 4

// Empty is a store with no policy and no database. It denies everything, which
// is what the process holds until Init has run.
func Empty() (*Store, error) {
	return NewInMemory(nil)
}

func connString(cfg *models.PostGreSQL) (string, error) {
	if cfg == nil {
		return "", errors.New("no database configuration")
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Username, cfg.Password),
		Host:     net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		Path:     cfg.Database,
		RawQuery: url.Values{"sslmode": {sslMode(cfg.SSLMode)}}.Encode(),
	}
	return u.String(), nil
}

func sslMode(mode string) string {
	if mode == "" {
		return "disable"
	}
	return mode
}

// Origin names this replica in the notifications it causes. The hostname is
// the pod name under Kubernetes and the container id under Compose, which is
// what an operator reading the logs wants to see.
func Origin() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		return fmt.Sprintf("core-%d", os.Getpid())
	}
	return host
}
