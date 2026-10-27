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
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	beegoorm "github.com/beego/beego/v2/client/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/orm"
	cfgdao "github.com/goharbor/harbor/src/pkg/config/db/dao"
	"github.com/goharbor/harbor/src/pkg/config/store"
)

type countingDriver struct {
	store.Driver
	loads atomic.Int64
}

func (c *countingDriver) Load(ctx context.Context) (map[string]any, error) {
	c.loads.Add(1)
	return c.Driver.Load(ctx)
}

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := beegoorm.GetDB()
	require.NoError(t, err)
	return db
}

// otherDB is a separate client, standing in for another Harbor instance.
func otherDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRESQL_HOST"), os.Getenv("POSTGRESQL_PORT"), os.Getenv("POSTGRESQL_USR"),
		os.Getenv("POSTGRESQL_PWD"), os.Getenv("POSTGRESQL_DATABASE")))
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// capOpenConns bounds the pool so tests can exhaust it; the test setup leaves it unlimited.
func capOpenConns(db *sql.DB) (restore func()) {
	prev := db.Stats().MaxOpenConnections
	if prev == 0 {
		db.SetMaxOpenConns(4)
	}
	return func() { db.SetMaxOpenConns(prev) }
}

func waitForRevisionAbove(t *testing.T, s *SyncedSettings, after uint64) {
	t.Helper()
	require.Eventually(t, func() bool { return s.Revision() > after }, 10*time.Second, 10*time.Millisecond)
}

func TestCommittedSaveRefreshesSettings(t *testing.T) {
	driver := &countingDriver{Driver: &Database{cfgDAO: cfgdao.New()}}
	s := newSyncedSettings(driver)
	stop := s.StartSync(testDB(t))
	defer stop()

	ctx := orm.Context()
	// initial load plus the catch-up refresh once LISTEN is active
	waitForRevisionAbove(t, s, 1)

	loads := driver.loads.Load()
	for range 100 {
		_, err := s.Load(ctx)
		require.NoError(t, err)
	}
	assert.Equal(t, loads, driver.loads.Load())

	before := s.Revision()
	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.AUTHMode: "ldap_auth"}))
	waitForRevisionAbove(t, s, before)
	v, err := s.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, "ldap_auth", v[common.AUTHMode])

	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.AUTHMode: "db_auth"}))
}

func TestRolledBackSavePublishesNoChange(t *testing.T) {
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(testDB(t))
	defer stop()

	ctx := orm.Context()
	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.AUTHMode: "db_auth"}))
	waitForRevisionAbove(t, s, 1)
	time.Sleep(500 * time.Millisecond)
	before := s.Revision()

	rollback := errors.New("rollback")
	err := orm.WithTransaction(func(ctx context.Context) error {
		if err := (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.AUTHMode: "ldap_auth"}); err != nil {
			return err
		}
		return rollback
	})(ctx)
	require.ErrorIs(t, err, rollback)

	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, before, s.Revision(), "a rolled back save must not notify")
	v, err := s.Load(ctx)
	require.NoError(t, err)
	assert.Equal(t, "db_auth", v[common.AUTHMode])
}

func listenerSessions(t *testing.T) int {
	t.Helper()
	var n int
	require.NoError(t, testDB(t).QueryRowContext(context.Background(),
		"SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND application_name = '"+listenerAppName+"'").Scan(&n))
	return n
}

func TestStopClosesListenerSession(t *testing.T) {
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(testDB(t))
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)

	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("stop did not return")
	}
	// destroyed, not returned to the pool for requests
	require.Eventually(t, func() bool { return listenerSessions(t) == 0 }, 10*time.Second, 10*time.Millisecond)
}

func TestLoadBeforeSyncReadsDatabase(t *testing.T) {
	driver := &countingDriver{Driver: &Database{cfgDAO: cfgdao.New()}}
	s := newSyncedSettings(driver)
	_, err := s.Load(orm.Context())
	require.NoError(t, err)
	assert.Equal(t, int64(1), driver.loads.Load())
	assert.Equal(t, uint64(0), s.Revision())
}

func TestListenerReconnectsAfterTermination(t *testing.T) {
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(testDB(t))
	defer stop()
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)

	_, err := testDB(t).ExecContext(context.Background(),
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = current_database() AND application_name = '"+listenerAppName+"'")
	require.NoError(t, err)
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 15*time.Second, 50*time.Millisecond)

	ctx := orm.Context()
	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.AUTHMode: "ldap_auth"}))
	require.Eventually(t, func() bool {
		v, _ := s.Load(ctx)
		return v[common.AUTHMode] == "ldap_auth"
	}, 10*time.Second, 10*time.Millisecond)
	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.AUTHMode: "db_auth"}))
}

// every pool connection held is the state of the #92 deadlock
func TestReadsSucceedWithExhaustedPool(t *testing.T) {
	pool := testDB(t)
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(pool)
	defer stop()
	waitForRevisionAbove(t, s, 1)

	mgr := NewDBCfgManager()
	mgr.Store = store.NewConfigStore(s)

	release := holdAllPoolConnections(t, pool)
	defer release()

	done := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(orm.Context(), 5*time.Second)
		defer cancel()
		done <- mgr.Load(ctx)
	}()
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("config load blocked on the exhausted pool")
	}
	assert.NotEmpty(t, mgr.Get(context.Background(), common.AUTHMode).GetString())
}

func TestReconnectAppliesChangesMissedWhileDisconnected(t *testing.T) {
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(testDB(t))
	defer stop()
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)

	ctx := orm.Context()
	_, err := testDB(t).ExecContext(context.Background(),
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = current_database() AND application_name = '"+listenerAppName+"'")
	require.NoError(t, err)
	require.Eventually(t, func() bool { return listenerSessions(t) == 0 }, 5*time.Second, 5*time.Millisecond)
	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.AUTHMode: "ldap_auth"}))

	require.Eventually(t, func() bool {
		v, _ := s.Load(ctx)
		return v[common.AUTHMode] == "ldap_auth"
	}, 15*time.Second, 20*time.Millisecond)
	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.AUTHMode: "db_auth"}))
}
