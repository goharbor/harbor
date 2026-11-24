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

package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"maps"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/config/db/dao"
	"github.com/goharbor/harbor/src/pkg/config/store"
)

const (
	listenerAppName = "harbor_configuration_listener"
	// catches missed notifications and edits made directly in SQL
	fullRefreshInterval = 5 * time.Minute
	reconnectBackoff    = time.Second
	maxReconnectBackoff = 30 * time.Second
	selectUserSettings  = "SELECT k, v FROM properties"
)

var (
	// also bounds the listener setup: pgx pings a reused connection with the caller's
	// context, and a dead peer would otherwise block that forever
	refreshTimeout = 30 * time.Second
	// a peer lost in a failover or partition leaves the socket open, so only a ping notices
	listenerPingInterval = 30 * time.Second
	pingTimeout          = 5 * time.Second
)

// SyncedSettings keeps the user settings in memory so reading configuration needs no
// database connection or lock, which is what let requests deadlock on the pool.
type SyncedSettings struct {
	driver   store.Driver
	db       *sql.DB // set by StartSync; refreshes use it because Beego ignores contexts
	settings atomic.Pointer[map[string]any]
	revision atomic.Uint64

	refreshMu sync.Mutex // orders refreshes so an older read never replaces a newer one
	refreshCh chan struct{}
	running   atomic.Bool
}

var _ store.Revisioned = (*SyncedSettings)(nil)

func newSyncedSettings(driver store.Driver) *SyncedSettings {
	return &SyncedSettings{driver: driver, refreshCh: make(chan struct{}, 1)}
}

func (s *SyncedSettings) Load(ctx context.Context) (map[string]any, error) {
	if settings := s.settings.Load(); settings != nil {
		return maps.Clone(*settings), nil
	}
	return s.driver.Load(ctx)
}

func (s *SyncedSettings) Save(ctx context.Context, cfg map[string]any) error {
	return s.driver.Save(ctx, cfg)
}

func (s *SyncedSettings) Get(ctx context.Context, key string) (map[string]any, error) {
	return s.driver.Get(ctx, key)
}

// Revision 0 means the settings are not synced and every Load reads the database.
func (s *SyncedSettings) Revision() uint64 {
	return s.revision.Load()
}

// userSettingsReader bypasses Beego, which ignores contexts and so cannot bound the wait.
type userSettingsReader func(ctx context.Context) (map[string]any, error)

func (s *SyncedSettings) refresh(ctx context.Context) error {
	return s.refreshFrom(ctx, s.readFromPool)
}

func (s *SyncedSettings) refreshFrom(ctx context.Context, read userSettingsReader) error {
	s.refreshMu.Lock()
	defer s.refreshMu.Unlock()
	ctx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()
	settings, err := read(ctx)
	if err != nil {
		return err
	}
	s.settings.Store(&settings)
	s.revision.Add(1)
	return nil
}

func (s *SyncedSettings) readFromPool(ctx context.Context) (map[string]any, error) {
	rows, err := s.db.QueryContext(ctx, selectUserSettings)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []*models.ConfigEntry
	for rows.Next() {
		e := &models.ConfigEntry{}
		if err := rows.Scan(&e.Key, &e.Value); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return userSettingsFrom(entries), nil
}

func readFromConn(conn *pgx.Conn) userSettingsReader {
	return func(ctx context.Context) (map[string]any, error) {
		rows, err := conn.Query(ctx, selectUserSettings)
		if err != nil {
			return nil, err
		}
		entries, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (*models.ConfigEntry, error) {
			e := &models.ConfigEntry{}
			return e, row.Scan(&e.Key, &e.Value)
		})
		if err != nil {
			return nil, err
		}
		return userSettingsFrom(entries), nil
	}
}

func (s *SyncedSettings) scheduleRefresh() {
	select {
	case s.refreshCh <- struct{}{}:
	default:
	}
}

func (s *SyncedSettings) StartSync(db *sql.DB) (stop func()) {
	// Without a listener an instance would serve stale values for its own writes until
	// the full refresh; reading the database is correct and, with no cache lock, safe.
	if db == nil || db.Stats().MaxOpenConnections == 1 { // 0 means unlimited
		log.Warning("database pool too small to hold a connection for the configuration change listener, reading configuration from the database on every request")
		return func() {}
	}
	if !s.running.CompareAndSwap(false, true) {
		log.Warning("configuration sync is already running")
		return func() {}
	}
	ctx, cancel := context.WithCancel(orm.Context())
	var done sync.WaitGroup
	s.db = db
	if err := s.refresh(ctx); err != nil {
		log.Errorf("failed to read the user settings, reading configuration from the database until it succeeds: %v", err)
	}
	done.Add(2)
	go func() { defer done.Done(); s.refreshLoop(ctx) }()
	go func() { defer done.Done(); s.watchChanges(ctx, db) }()
	var once sync.Once
	return func() {
		once.Do(func() {
			cancel()
			done.Wait()
			s.running.Store(false)
		})
	}
}

func (s *SyncedSettings) refreshLoop(ctx context.Context) {
	ticker := time.NewTicker(fullRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		case <-s.refreshCh:
		}
		if err := s.refresh(ctx); err != nil && ctx.Err() == nil {
			log.Warningf("failed to refresh the user settings, keeping the previous ones: %v", err)
		}
	}
}

func (s *SyncedSettings) watchChanges(ctx context.Context, db *sql.DB) {
	backoff := reconnectBackoff
	for ctx.Err() == nil {
		listened, err := s.subscribe(ctx, db)
		if ctx.Err() != nil {
			return
		}
		log.Warningf("configuration change listener stopped, falling back to the %s full refresh until it reconnects: %v", fullRefreshInterval, err)
		if listened {
			// the backoff is for failing connects; a dropped working listener retries at once
			backoff = reconnectBackoff
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, maxReconnectBackoff)
	}
}

// subscribe reports whether it got as far as listening before the error.
func (s *SyncedSettings) subscribe(ctx context.Context, db *sql.DB) (bool, error) {
	setupCtx, cancel := context.WithTimeout(ctx, refreshTimeout)
	defer cancel()
	conn, err := db.Conn(setupCtx)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	// ErrBadConn makes database/sql discard it: a LISTEN session must never serve requests
	var (
		listened bool
		cause    error
	)
	err = conn.Raw(func(driverConn any) error {
		c, ok := driverConn.(*stdlib.Conn)
		if !ok {
			cause = fmt.Errorf("unexpected driver connection %T", driverConn)
		} else {
			listened, cause = s.awaitChanges(ctx, setupCtx, c.Conn())
		}
		return driver.ErrBadConn
	})
	if cause != nil {
		return listened, cause
	}
	return listened, err
}

func (s *SyncedSettings) awaitChanges(ctx, setupCtx context.Context, conn *pgx.Conn) (bool, error) {
	// explains the long-lived session to operators in pg_stat_activity
	if _, err := conn.Exec(setupCtx, "SET application_name = '"+listenerAppName+"'"); err != nil {
		return false, err
	}
	if _, err := conn.Exec(setupCtx, "LISTEN "+dao.ConfigurationChangedChannel); err != nil {
		return false, err
	}
	// Changes committed while disconnected were announced to nobody. Reading on this
	// connection avoids waiting for a second one from an exhausted pool.
	if err := s.refreshFrom(setupCtx, readFromConn(conn)); err != nil {
		return false, fmt.Errorf("sync after connect: %w", err)
	}
	for {
		waitCtx, cancel := context.WithTimeout(ctx, listenerPingInterval)
		_, err := conn.WaitForNotification(waitCtx)
		idle := errors.Is(waitCtx.Err(), context.DeadlineExceeded) // read before cancel overwrites it
		cancel()
		switch {
		case err == nil:
			s.scheduleRefresh()
		case ctx.Err() != nil:
			return true, ctx.Err()
		case idle:
			// pgconn keeps the connection open on a timeout, so it can be probed
			pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
			err = conn.Ping(pingCtx)
			cancel()
			if err != nil {
				return true, fmt.Errorf("listener connection lost: %w", err)
			}
		default:
			return true, err
		}
	}
}
