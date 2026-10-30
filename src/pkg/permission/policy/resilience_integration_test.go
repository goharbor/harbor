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

package policy_test

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/permission/policy"
)

// grantsPerShippedRole is what each of the five roles Harbor ships carries,
// counted from testdata/shipped_grants_on_main.golden. Pinning it here as well
// asserts that what reaches the database is what the Go side holds.
var grantsPerShippedRole = map[int64]int{1: 76, 2: 36, 3: 20, 4: 53, 5: 14}

func countGrants(t *testing.T, db *sql.DB, roleID int64) int {
	t.Helper()
	var n int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM role_permission WHERE role_type = $1 AND role_id = $2`,
		policy.RoleType, roleID).Scan(&n))
	return n
}

func policyVersion(t *testing.T, db *sql.DB) int64 {
	t.Helper()
	var v int64
	require.NoError(t, db.QueryRow(`SELECT version FROM policy_version WHERE only_row`).Scan(&v))
	return v
}

// restoreShippedGrants puts the five roles back the way a seeded database has
// them, for a test that had to take them apart.
func restoreShippedGrants(t *testing.T, conn string, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`DELETE FROM role_permission WHERE role_type = $1 AND role_id BETWEEN 1 AND 5`,
		policy.RoleType)
	require.NoError(t, err)
	require.NoError(t, deaf(t, conn, "test-restore").EnsureSeeded(context.Background()))
}

// Seeding has to put the grants of all five roles in the database, not merely
// leave the process able to answer from memory.
func TestIntegrationSeedingWritesTheExpectedRows(t *testing.T) {
	conn := dsn(t)
	db := open(t, conn, "test-rows")
	require.NoError(t, deaf(t, conn, "test-rows-seed").EnsureSeeded(context.Background()))

	total := 0
	for roleID, want := range grantsPerShippedRole {
		got := countGrants(t, db, roleID)
		assert.Equal(t, want, got, "role %d", roleID)
		total += got
	}
	assert.Equal(t, 199, total, "the five roles Harbor ships, in full")

	// permission_policy is a dictionary, so the same permission held by two
	// roles is one row rather than two.
	var distinct int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM permission_policy WHERE scope = '/project/*'`).Scan(&distinct))
	assert.GreaterOrEqual(t, distinct, 76, "the distinct permissions the five roles draw on")
}

// Every core runs EnsureSeeded on the way up, so they collide by design. The
// advisory lock is what stops five of them writing five copies.
func TestIntegrationFiveReplicasSeedingAtOnceWriteOneSet(t *testing.T) {
	conn := dsn(t)
	db := open(t, conn, "test-race-count")
	restoreShippedGrants(t, conn, db)

	// Back to the state a database is in before anyone has seeded.
	_, err := db.Exec(`DELETE FROM role_permission WHERE role_type = $1 AND role_id BETWEEN 1 AND 5`,
		policy.RoleType)
	require.NoError(t, err)
	t.Cleanup(func() { restoreShippedGrants(t, conn, db) })

	const replicas = 5
	stores := make([]*policy.Store, replicas)
	for i := range stores {
		stores[i] = deaf(t, conn, "test-replica")
	}

	var wg sync.WaitGroup
	errs := make(chan error, replicas)
	for i := range stores {
		wg.Add(1)
		go func(s *policy.Store) {
			defer wg.Done()
			errs <- s.EnsureSeeded(context.Background())
		}(stores[i])
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	total := 0
	for roleID, want := range grantsPerShippedRole {
		got := countGrants(t, db, roleID)
		assert.Equal(t, want, got, "role %d was written once, not five times", roleID)
		total += got
	}
	assert.Equal(t, 199, total)

	// Whichever one lost the race still came out holding the policy.
	for i, s := range stores {
		granted, err := s.Enforce(policy.Subject(1), policy.Object("repository"), "push")
		require.NoError(t, err)
		assert.True(t, granted, "replica %d", i)
	}
}

// Decided behaviour: any grant on a shipped role means an operator owns that
// role, and seeding leaves it alone. Seeding writes all of a role's grants in
// one transaction, so a half-filled role is never something Harbor produced --
// it is something a person did, and putting the rows back would undo it on
// every restart.
func TestIntegrationARoleAnOperatorEditedIsLeftAlone(t *testing.T) {
	conn := dsn(t)
	db := open(t, conn, "test-partial")
	restoreShippedGrants(t, conn, db)
	t.Cleanup(func() { restoreShippedGrants(t, conn, db) })

	// An operator who cut guest down to a single grant.
	_, err := db.Exec(`
		DELETE FROM role_permission
		 WHERE role_type = $1 AND role_id = 3
		   AND id <> (SELECT min(id) FROM role_permission WHERE role_type = $1 AND role_id = 3)`,
		policy.RoleType)
	require.NoError(t, err)
	require.Equal(t, 1, countGrants(t, db, 3), "the edit this test is about")

	require.NoError(t, deaf(t, conn, "test-partial-seed").EnsureSeeded(context.Background()))

	assert.Equal(t, 1, countGrants(t, db, 3),
		"a role with grants is left as it is, even when they are fewer than shipped")
	assert.Equal(t, grantsPerShippedRole[2], countGrants(t, db, 2),
		"and the roles nobody touched are untouched")
}

// The seed is two statements rather than two hundred because the trigger is
// AFTER STATEMENT. If that ever became AFTER ROW, a first start would ask the
// fleet to reload several hundred times.
func TestIntegrationABulkWriteBumpsTheVersionOnce(t *testing.T) {
	conn := dsn(t)
	db := open(t, conn, "test-bulk")

	before := policyVersion(t, db)
	_, err := db.Exec(`
		INSERT INTO permission_policy (scope, resource, action, effect)
		SELECT '/project/*test-bulk', 'resource-' || i, 'read', 'allow'
		  FROM generate_series(1, 100) AS i
		ON CONFLICT ON CONSTRAINT unique_rbac_policy DO NOTHING`)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM role_permission WHERE role_id = 3 AND permission_policy_id IN
			(SELECT id FROM permission_policy WHERE scope = '/project/*test-bulk')`)
		_, _ = db.Exec(`DELETE FROM permission_policy WHERE scope = '/project/*test-bulk'`)
	})

	// permission_policy is shared with robot accounts, so a policy row nobody
	// has linked to a project role yet is not the fleet's business.
	assert.Equal(t, before, policyVersion(t, db), "policy rows on their own tell nobody")

	_, err = db.Exec(`
		INSERT INTO role_permission (role_type, role_id, permission_policy_id)
		SELECT $1, 3, id FROM permission_policy WHERE scope = '/project/*test-bulk'`,
		policy.RoleType)
	require.NoError(t, err)

	assert.Equal(t, before+1, policyVersion(t, db), "100 links, one statement, one version")
}

// The trigger fires from inside the writing transaction, so a write that never
// lands cannot ask anyone to reload.
func TestIntegrationANotificationDoesNotSurviveARollback(t *testing.T) {
	conn := dsn(t)
	db := open(t, conn, "test-rollback")

	before := policyVersion(t, db)

	tx, err := db.Begin()
	require.NoError(t, err)
	var policyID int64
	require.NoError(t, tx.QueryRow(`
		INSERT INTO permission_policy (scope, resource, action, effect)
		VALUES ('/project/*test-rollback', 'repository', 'push', 'allow')
		ON CONFLICT ON CONSTRAINT unique_rbac_policy
		DO UPDATE SET scope = EXCLUDED.scope
		RETURNING id`).Scan(&policyID))
	// The link is what would notify, so it is the write that has to be rolled
	// back for this test to be about a rollback at all.
	_, err = tx.Exec(
		`INSERT INTO role_permission (role_type, role_id, permission_policy_id) VALUES ($1, 3, $2)`,
		policy.RoleType, policyID)
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())

	assert.Equal(t, before, policyVersion(t, db), "the version went back with the transaction")

	var rows int
	require.NoError(t, db.QueryRow(
		`SELECT count(*) FROM permission_policy WHERE scope = '/project/*test-rollback'`).Scan(&rows))
	assert.Zero(t, rows)
}

// A replica that writes gets the notification it caused, and has to act on it.
// Writing the row does not update this replica's enforcer and nothing else
// does, so the core an operator edited a role on would otherwise be the last
// one in the fleet to serve that edit.
func TestIntegrationAReplicaLoadsItsOwnWrite(t *testing.T) {
	conn := dsn(t)
	const origin = "test-self"

	store := replica(t, conn, origin)
	require.NoError(t, store.EnsureSeeded(context.Background()))
	writer := open(t, conn, origin) // same origin: this is the replica writing

	granted, err := store.Enforce(policy.Subject(3), policy.Object("repository"), "scanner-pull")
	require.NoError(t, err)
	require.False(t, granted, "the grant this test adds is not there yet")

	before := store.Generation()

	var policyID int64
	require.NoError(t, writer.QueryRow(`
		INSERT INTO permission_policy (scope, resource, action, effect)
		VALUES ('/project/*', 'repository', 'scanner-pull', 'allow')
		ON CONFLICT ON CONSTRAINT unique_rbac_policy
		DO UPDATE SET scope = EXCLUDED.scope
		RETURNING id`).Scan(&policyID))
	_, err = writer.Exec(
		`INSERT INTO role_permission (role_type, role_id, permission_policy_id) VALUES ($1, 3, $2)`,
		policy.RoleType, policyID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = writer.Exec(
			`DELETE FROM role_permission WHERE role_id = 3 AND permission_policy_id = $1`, policyID)
	})

	waitFor(t, 5*time.Second, func() bool { return store.Generation() > before })
	assert.Positive(t, store.WatcherStats().SelfOriginated,
		"it recognised the notification as its own")

	granted, err = store.Enforce(policy.Subject(3), policy.Object("repository"), "scanner-pull")
	require.NoError(t, err)
	assert.True(t, granted, "and serves what it just wrote, without waiting for the resync")
}

// A reload that fails must leave the replica answering from the policy it
// already has. Emptying the enforcer on a failed read would turn a database
// blip into a fleet-wide denial.
func TestIntegrationAFailedReloadKeepsTheLastGoodPolicy(t *testing.T) {
	conn := dsn(t)
	db := open(t, conn, "test-failed-reload")
	require.NoError(t, deaf(t, conn, "test-failed-reload-seed").EnsureSeeded(context.Background()))

	store, err := policy.New(context.Background(), db, "test-failed-reload")
	require.NoError(t, err)
	t.Cleanup(store.Close)

	granted, err := store.Enforce(policy.Subject(1), policy.Object("repository"), "push")
	require.NoError(t, err)
	require.True(t, granted)

	generation := store.Generation()
	require.NoError(t, db.Close()) // the database goes away under it

	assert.Error(t, store.Reload(context.Background(), "test"), "the reload cannot read anything")
	assert.Equal(t, generation, store.Generation(), "so it did not install anything either")

	granted, err = store.Enforce(policy.Subject(1), policy.Object("repository"), "push")
	require.NoError(t, err)
	assert.True(t, granted, "it still answers from the policy it had")

	granted, err = store.Enforce(policy.Subject(3), policy.Object("repository"), "push")
	require.NoError(t, err)
	assert.False(t, granted, "and it is the real policy, not an allow-everything fallback")
}

// The whole point of holding the policy in memory: a permission question is
// answered without going near the database.
func TestIntegrationAPermissionCheckDoesNotTouchTheDatabase(t *testing.T) {
	conn := dsn(t)
	db := open(t, conn, "test-no-db")
	require.NoError(t, deaf(t, conn, "test-no-db-seed").EnsureSeeded(context.Background()))

	store, err := policy.New(context.Background(), db, "test-no-db")
	require.NoError(t, err)
	t.Cleanup(store.Close)
	require.NoError(t, db.Close())

	for _, tc := range []struct {
		roleID   int64
		resource string
		action   string
		allowed  bool
	}{
		{1, "repository", "push", true},
		{2, "repository", "push", true},
		{3, "repository", "push", false},
		{3, "repository", "pull", true},
		{4, "member", "create", false},
		{1, "member", "create", true},
		{5, "repository", "pull", true},
	} {
		granted, err := store.Enforce(policy.Subject(tc.roleID), policy.Object(tc.resource), tc.action)
		require.NoError(t, err, "role %d %s:%s", tc.roleID, tc.resource, tc.action)
		assert.Equal(t, tc.allowed, granted, "role %d %s:%s", tc.roleID, tc.resource, tc.action)
	}
}

// Connections drop. The watcher has to come back on its own and read the
// policy when it does, rather than waiting for a notification that was sent
// while it was away.
func TestIntegrationTheWatcherReconnectsAndReloads(t *testing.T) {
	conn := dsn(t)
	db := open(t, conn, "test-reconnect")
	store := replica(t, conn, "test-reconnect-reader")
	require.NoError(t, store.EnsureSeeded(context.Background()))

	reconnects := store.WatcherStats().Reconnects

	// Cut every listening connection, this replica's included.
	_, err := db.Exec(`
		SELECT pg_terminate_backend(pid)
		  FROM pg_stat_activity
		 WHERE pid <> pg_backend_pid()
		   AND query ILIKE 'listen %'`)
	require.NoError(t, err)

	waitFor(t, 30*time.Second, func() bool {
		return store.WatcherStats().Reconnects > reconnects
	})

	// And it is listening again: a write made now still reaches it.
	writer := open(t, conn, "test-reconnect-writer")
	generation := store.Generation()

	var policyID int64
	require.NoError(t, writer.QueryRow(`
		INSERT INTO permission_policy (scope, resource, action, effect)
		VALUES ('/project/*', 'repository', 'delete', 'allow')
		ON CONFLICT ON CONSTRAINT unique_rbac_policy
		DO UPDATE SET scope = EXCLUDED.scope
		RETURNING id`).Scan(&policyID))
	_, err = writer.Exec(
		`INSERT INTO role_permission (role_type, role_id, permission_policy_id) VALUES ($1, 5, $2)`,
		policy.RoleType, policyID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = writer.Exec(
			`DELETE FROM role_permission WHERE role_id = 5 AND permission_policy_id = $1`, policyID)
	})

	waitFor(t, 15*time.Second, func() bool { return store.Generation() > generation })

	granted, err := store.Enforce(policy.Subject(5), policy.Object("repository"), "delete")
	require.NoError(t, err)
	assert.True(t, granted, "the reconnected watcher is carrying writes again")
}
