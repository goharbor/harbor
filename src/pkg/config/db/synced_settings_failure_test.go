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
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/orm"
	cfgdao "github.com/goharbor/harbor/src/pkg/config/db/dao"
	"github.com/goharbor/harbor/src/pkg/config/store"
)

func holdAllPoolConnections(t *testing.T, pool *sql.DB) (release func()) {
	t.Helper()
	restoreMax := capOpenConns(pool)
	var held []*sql.Conn
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		c, err := pool.Conn(ctx)
		cancel()
		if err != nil {
			break
		}
		held = append(held, c)
	}
	require.Equal(t, pool.Stats().MaxOpenConnections, pool.Stats().InUse)
	var once sync.Once
	return func() {
		once.Do(func() {
			for _, c := range held {
				_ = c.Close()
			}
			restoreMax()
		})
	}
}

func saveAuthMode(t *testing.T, mode string) {
	t.Helper()
	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(orm.Context(), map[string]any{common.AUTHMode: mode}))
}

func stopWithin(t *testing.T, stop func(), d time.Duration) time.Duration {
	t.Helper()
	start := time.Now()
	done := make(chan struct{})
	go func() { stop(); close(done) }()
	select {
	case <-done:
	case <-time.After(d):
		t.Fatalf("stop did not return within %s", d)
	}
	return time.Since(start)
}

func TestListenerReconnectsOncePoolFreesUp(t *testing.T) {
	pool := testDB(t)
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(pool)
	defer stop()
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)
	saveAuthMode(t, "db_auth")
	require.Eventually(t, authModeIs(s, "db_auth"), 5*time.Second, 10*time.Millisecond)

	// drop the listener, then take every connection so it cannot reconnect
	_, err := pool.ExecContext(context.Background(),
		"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = current_database() AND application_name = '"+listenerAppName+"'")
	require.NoError(t, err)
	require.Eventually(t, func() bool { return listenerSessions(t) == 0 }, 5*time.Second, 5*time.Millisecond)
	release := holdAllPoolConnections(t, pool)
	defer release()

	// a change committed by another instance while this one is starved
	other := otherDB(t)
	_, err = other.ExecContext(context.Background(), "UPDATE properties SET v = 'ldap_auth' WHERE k = 'auth_mode'")
	require.NoError(t, err)
	_, err = other.ExecContext(context.Background(), "SELECT pg_notify('harbor_configuration_changed', '')")
	require.NoError(t, err)

	for range 100 {
		v, err := s.Load(orm.Context())
		require.NoError(t, err)
		require.Equal(t, "db_auth", v[common.AUTHMode])
	}
	time.Sleep(2 * time.Second)
	release()
	start := time.Now()
	require.Eventually(t, authModeIs(s, "ldap_auth"), 30*time.Second, 20*time.Millisecond)
	t.Logf("missed change applied %s after the pool freed up", time.Since(start))
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 30*time.Second, 20*time.Millisecond)
	saveAuthMode(t, "db_auth")
}

func TestRefreshWaitsForPoolAndKeepsPreviousSettings(t *testing.T) {
	pool := testDB(t)
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(pool)
	defer stop()
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)
	saveAuthMode(t, "db_auth")
	require.Eventually(t, authModeIs(s, "db_auth"), 5*time.Second, 10*time.Millisecond)

	release := holdAllPoolConnections(t, pool)
	defer release()
	other := otherDB(t)
	// the notification reaches the listener, but the refresh cannot get a connection
	_, err := other.ExecContext(context.Background(), "UPDATE properties SET v = 'ldap_auth' WHERE k = 'auth_mode'")
	require.NoError(t, err)
	_, err = other.ExecContext(context.Background(), "SELECT pg_notify('harbor_configuration_changed', '')")
	require.NoError(t, err)
	time.Sleep(time.Second)
	v, err := s.Load(orm.Context())
	require.NoError(t, err)
	assert.Equal(t, "db_auth", v[common.AUTHMode], "previous settings are kept while the refresh waits")

	release()
	start := time.Now()
	require.Eventually(t, authModeIs(s, "ldap_auth"), 35*time.Second, 20*time.Millisecond)
	t.Logf("refresh completed %s after the pool freed up", time.Since(start))
	saveAuthMode(t, "db_auth")
}

func TestStopReturnsWhilePoolExhausted(t *testing.T) {
	pool := testDB(t)
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(pool)
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)

	release := holdAllPoolConnections(t, pool)
	defer release()
	// a refresh queued behind the exhausted pool, and a listener that must reconnect
	s.scheduleRefresh()
	time.Sleep(500 * time.Millisecond)
	t.Logf("stop returned after %s", stopWithin(t, stop, 10*time.Second))
}

func TestStartGivesUpOnExhaustedPoolThenCatchesUp(t *testing.T) {
	defer func(d time.Duration) { refreshTimeout = d }(refreshTimeout)
	refreshTimeout = 2 * time.Second
	pool := testDB(t)
	release := holdAllPoolConnections(t, pool)
	defer release()

	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	start := time.Now()
	started := make(chan func(), 1)
	go func() { started <- s.StartSync(pool) }()
	var stop func()
	select {
	case stop = <-started:
		t.Logf("StartSync returned after %s with revision %d", time.Since(start), s.Revision())
	case <-time.After(10 * time.Second):
		t.Fatal("Start blocked on an exhausted pool")
	}
	release()
	require.Eventually(t, func() bool { return s.Revision() > 0 }, 35*time.Second, 20*time.Millisecond)
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 35*time.Second, 20*time.Millisecond)
	stopWithin(t, stop, 10*time.Second)
}

func TestConcurrentSavesConvergeOnLastValue(t *testing.T) {
	driver := &countingDriver{Driver: &Database{cfgDAO: cfgdao.New()}}
	s := newSyncedSettings(driver)
	stop := s.StartSync(testDB(t))
	defer stop()
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)
	base := s.Revision()

	var wg sync.WaitGroup
	for i := range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range 50 {
				if err := (&Database{cfgDAO: cfgdao.New()}).Save(orm.Context(), map[string]any{common.LDAPURL: fmt.Sprintf("ldap://%d-%d", i, j)}); err != nil {
					t.Errorf("save failed: %v", err)
				}
			}
		}()
	}
	stopReaders := make(chan struct{})
	var readers sync.WaitGroup
	for range 16 {
		readers.Add(1)
		go func() {
			defer readers.Done()
			for {
				select {
				case <-stopReaders:
					return
				default:
					_, err := s.Load(orm.Context())
					if err != nil {
						t.Error(err)
						return
					}
				}
			}
		}()
	}
	wg.Wait()
	final := "ldap://final"
	require.NoError(t, (&Database{cfgDAO: cfgdao.New()}).Save(orm.Context(), map[string]any{common.LDAPURL: final}))
	require.Eventually(t, func() bool {
		v, _ := s.Load(orm.Context())
		return v[common.LDAPURL] == final
	}, 10*time.Second, 10*time.Millisecond)
	close(stopReaders)
	readers.Wait()
	t.Logf("401 committed saves caused %d refreshes", s.Revision()-base)
	assert.Equal(t, int64(0), driver.loads.Load(), "no read went through the Beego driver")
}

func TestManagersShareOneListener(t *testing.T) {
	stop := StartSettingsSync(testDB(t))
	defer stop()
	for range 20 {
		_ = NewDBCfgManager()
	}
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)
	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, 1, listenerSessions(t))
}
func authModeIs(s *SyncedSettings, want string) func() bool {
	return func() bool { v, _ := s.Load(orm.Context()); return v[common.AUTHMode] == want }
}

func TestSecondStartIsIgnoredAndRestartWorks(t *testing.T) {
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(testDB(t))
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)
	s.StartSync(testDB(t))()
	time.Sleep(300 * time.Millisecond)
	assert.Equal(t, 1, listenerSessions(t))

	stopWithin(t, stop, 10*time.Second)
	require.Eventually(t, func() bool { return listenerSessions(t) == 0 }, 10*time.Second, 10*time.Millisecond)
	stop = s.StartSync(testDB(t))
	require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 10*time.Second, 10*time.Millisecond)
	stopWithin(t, stop, 10*time.Second)
}

// PUT /configurations runs in the request transaction
func TestRolledBackUpdateIsNotServed(t *testing.T) {
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(testDB(t))
	defer stop()
	saveAuthMode(t, "db_auth")
	require.Eventually(t, authModeIs(s, "db_auth"), 10*time.Second, 10*time.Millisecond)

	mgr := NewDBCfgManager()
	mgr.Store = store.NewConfigStore(s)
	ctx := orm.Context()
	require.NoError(t, mgr.Load(ctx))

	rollback := errors.New("rollback")
	err := orm.WithTransaction(func(ctx context.Context) error {
		if err := mgr.UpdateConfig(ctx, map[string]any{common.AUTHMode: "ldap_auth"}); err != nil {
			return err
		}
		return rollback
	})(ctx)
	require.ErrorIs(t, err, rollback)

	require.NoError(t, mgr.Load(ctx))
	assert.Equal(t, "db_auth", mgr.Get(ctx, common.AUTHMode).GetString())
}

func TestSingleConnectionPoolReadsDatabase(t *testing.T) {
	db := otherDB(t)
	db.SetMaxOpenConns(1)

	driver := &countingDriver{Driver: &Database{cfgDAO: cfgdao.New()}}
	s := newSyncedSettings(driver)
	stop := s.StartSync(db)
	defer stop()
	assert.Equal(t, uint64(0), s.Revision(), "settings are not synced without a listener")
	assert.Equal(t, 0, db.Stats().InUse, "no connection is held")

	saveAuthMode(t, "ldap_auth")
	v, err := s.Load(orm.Context())
	require.NoError(t, err)
	assert.Equal(t, "ldap_auth", v[common.AUTHMode])
	assert.Equal(t, int64(1), driver.loads.Load())
	saveAuthMode(t, "db_auth")
}

// a duplicate-key error on first creation would abort the transaction
func TestConcurrentFirstSaveOfPropertySucceeds(t *testing.T) {
	ctx := orm.Context()
	o, err := orm.FromContext(ctx)
	require.NoError(t, err)
	var original []string
	_, err = o.Raw("SELECT v FROM properties WHERE k = ?", common.LDAPURL).QueryRows(&original)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, err := o.Raw("DELETE FROM properties WHERE k = ?", common.LDAPURL).Exec()
		require.NoError(t, err)
		if len(original) == 1 {
			_, err = o.Raw("INSERT INTO properties (k, v) VALUES (?, ?)", common.LDAPURL, original[0]).Exec()
			require.NoError(t, err)
		}
	})
	for round := range 20 {
		_, err := o.Raw("DELETE FROM properties WHERE k = ?", common.LDAPURL).Exec()
		require.NoError(t, err)
		start := make(chan struct{})
		errs := make(chan error, 8)
		for i := range 8 {
			go func() {
				<-start
				errs <- orm.WithTransaction(func(ctx context.Context) error {
					return (&Database{cfgDAO: cfgdao.New()}).Save(ctx, map[string]any{common.LDAPURL: fmt.Sprintf("ldap://%d-%d", round, i)})
				})(orm.Context())
			}()
		}
		close(start)
		for range 8 {
			require.NoError(t, <-errs)
		}
	}
}

// zombieProxy forwards TCP to Postgres. zombify makes every existing link drop its
// bytes from then on while the sockets stay open, like a peer lost without a reset;
// links opened afterwards work normally.
type zombieProxy struct {
	ln    net.Listener
	mu    sync.Mutex
	dead  []*atomic.Bool
	conns []net.Conn
}

func startZombieProxy(t *testing.T, target string) *zombieProxy {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	p := &zombieProxy{ln: ln}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			u, err := net.Dial("tcp", target)
			if err != nil {
				c.Close()
				continue
			}
			dead := &atomic.Bool{}
			p.mu.Lock()
			p.dead = append(p.dead, dead)
			p.conns = append(p.conns, c, u)
			p.mu.Unlock()
			go pipe(c, u, dead)
			go pipe(u, c, dead)
		}
	}()
	t.Cleanup(func() {
		ln.Close()
		p.mu.Lock()
		defer p.mu.Unlock()
		for _, c := range p.conns {
			c.Close()
		}
	})
	return p
}

func pipe(src, dst net.Conn, dead *atomic.Bool) {
	buf := make([]byte, 32*1024)
	for {
		n, err := src.Read(buf)
		if err != nil {
			return
		}
		if !dead.Load() {
			if _, err := dst.Write(buf[:n]); err != nil {
				return
			}
		}
	}
}

func (p *zombieProxy) zombify() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, d := range p.dead {
		d.Store(true)
	}
}

func (p *zombieProxy) port() int { return p.ln.Addr().(*net.TCPAddr).Port }

// A dead connection that never errors must not block the listener: pgx pings a reused
// connection with the caller's context before handing it out.
func TestListenerRecoversFromDeadConnectionWithoutReset(t *testing.T) {
	defer func(r, i, p time.Duration) { refreshTimeout, listenerPingInterval, pingTimeout = r, i, p }(refreshTimeout, listenerPingInterval, pingTimeout)
	refreshTimeout, listenerPingInterval, pingTimeout = 2*time.Second, time.Second, time.Second

	proxy := startZombieProxy(t, net.JoinHostPort(os.Getenv("POSTGRESQL_HOST"), os.Getenv("POSTGRESQL_PORT")))
	pool, err := sql.Open("pgx", fmt.Sprintf("host=127.0.0.1 port=%d user=%s password=%s dbname=%s sslmode=disable",
		proxy.port(), os.Getenv("POSTGRESQL_USR"), os.Getenv("POSTGRESQL_PWD"), os.Getenv("POSTGRESQL_DATABASE")))
	require.NoError(t, err)
	defer pool.Close()
	pool.SetMaxOpenConns(4)

	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(pool)
	defer stop()
	saveAuthMode(t, "db_auth")
	require.Eventually(t, authModeIs(s, "db_auth"), 10*time.Second, 10*time.Millisecond)
	// the refresh after the save leaves an idle connection next to the listener's
	require.Eventually(t, func() bool {
		st := pool.Stats()
		return st.InUse == 1 && st.Idle >= 1
	}, 10*time.Second, 10*time.Millisecond)

	proxy.zombify()
	saveAuthMode(t, "ldap_auth")
	start := time.Now()
	require.Eventually(t, authModeIs(s, "ldap_auth"), 20*time.Second, 50*time.Millisecond)
	t.Logf("recovered %s after the connections died", time.Since(start))
	saveAuthMode(t, "db_auth")
}

// A working listener that drops is re-established at once, not after a growing backoff.
func TestListenerReconnectsQuicklyAfterRepeatedDrops(t *testing.T) {
	s := newSyncedSettings(&Database{cfgDAO: cfgdao.New()})
	stop := s.StartSync(testDB(t))
	defer stop()
	for i := range 6 {
		require.Eventually(t, func() bool { return listenerSessions(t) == 1 }, 3*time.Second, 20*time.Millisecond, "drop %d", i)
		_, err := testDB(t).ExecContext(context.Background(),
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = current_database() AND application_name = '"+listenerAppName+"'")
		require.NoError(t, err)
		require.Eventually(t, func() bool { return listenerSessions(t) == 0 }, 3*time.Second, 10*time.Millisecond)
	}
}
