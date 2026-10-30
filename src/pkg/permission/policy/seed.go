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
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

// seedLock is the advisory lock key seeding holds. Several cores start at once
// and would otherwise race to write the same rows.
const seedLock = 8_050_190

// policyScope is the scope string controller/role writes for a project role's
// permissions. It is a label on the row, not the object the matcher sees.
const policyScope = "/project/*"

// builtinRoles maps the role ids Harbor ships to the key their grants are
// under in shippedRoles. project_member has referred to these ids since Harbor
// 1.x, so they are the ids the rows already have.
var builtinRoles = map[int64]string{
	1: "projectAdmin",
	2: "developer",
	3: "guest",
	4: "maintainer",
	5: "limitedGuest",
}

// EnsureSeeded writes the grants of the roles Harbor ships, the first time a
// Harbor starts against a database that does not have them yet.
//
// Until now those grants were a map compiled into the binary, so there was
// nothing to seed. They go into the same two tables a custom role already
// uses, which is what leaves one way to resolve any role.
//
// It runs on every boot of every replica and writes nothing it does not have
// to. Whatever is already there wins: a role that already has grants is left
// exactly as it is, so an operator who edited one keeps that edit across
// restarts and upgrades.
func (s *Store) EnsureSeeded(ctx context.Context) error {
	if s.db == nil {
		return nil
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to open the seeding transaction")
	}
	defer func() { _ = tx.Rollback() }()

	// Cores start together. One of them seeds, the rest wait here and then
	// find there is nothing left to do.
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, seedLock); err != nil {
		return errors.Wrap(err, "failed to take the seeding lock")
	}

	seeded, err := seedRoles(ctx, tx)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit the seeded grants")
	}
	if seeded > 0 {
		log.Infof("seeded %d grants for the roles Harbor ships", seeded)
	}
	// Reload whether or not this replica was the one that wrote. A replica
	// that started at the same time read an empty policy in New, then waited
	// here for the lock while another replica seeded and committed. Its
	// watcher is not listening yet, so the notification for that write reaches
	// nobody, and without this it would refuse valid role access until the
	// resync thirty seconds later. Taking the lock puts us strictly after the
	// writer's commit, so what we read here is what was written.
	return s.Reload(ctx, "seed")
}

// seedRoles writes the grants of every shipped role that has none.
//
// It is two statements rather than one per grant. Each statement fires the
// AFTER STATEMENT trigger, so a row-at-a-time seed would advance
// policy_version by several hundred on first boot and make that counter read
// like noise to anyone watching it.
func seedRoles(ctx context.Context, tx *sql.Tx) (int, error) {
	type want struct {
		roleID   int64
		resource string
		action   string
	}

	var wanted []want
	for _, roleID := range sortedBuiltinIDs() {
		seed, err := roleNeedsSeeding(ctx, tx, roleID)
		if err != nil {
			return 0, err
		}
		if !seed {
			continue
		}
		for _, p := range shippedRoles[builtinRoles[roleID]] {
			wanted = append(wanted, want{roleID, p.Resource.String(), p.Action.String()})
		}
	}
	if len(wanted) == 0 {
		return 0, nil
	}

	// permission_policy is a dictionary of distinct permissions shared by every
	// role that holds one, so this adds only the ones that are not there yet.
	seen := map[string]bool{}
	var polArgs []any
	var polRows []string
	for _, w := range wanted {
		key := w.resource + "\x00" + w.action
		if seen[key] {
			continue
		}
		seen[key] = true
		n := len(polArgs)
		polRows = append(polRows, fmt.Sprintf("($%d, $%d, $%d, 'allow')", n+1, n+2, n+3))
		polArgs = append(polArgs, policyScope, w.resource, w.action)
	}
	// nolint:gosec // G202: the concatenated fragment is generated placeholders
	// ($1, $2, ...) whose count follows the number of rows; every value is passed
	// as a query parameter.
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO permission_policy (scope, resource, action, effect) VALUES `+
			strings.Join(polRows, ", ")+
			` ON CONFLICT ON CONSTRAINT unique_rbac_policy DO NOTHING`, polArgs...); err != nil {
		return 0, errors.Wrap(err, "failed to write the permissions of the shipped roles")
	}

	// Then link each role to them by looking the ids back up, so a permission
	// that was already there is reused rather than duplicated.
	var linkArgs []any
	var linkRows []string
	for _, w := range wanted {
		n := len(linkArgs)
		linkRows = append(linkRows, fmt.Sprintf("($%d::bigint, $%d::varchar, $%d::varchar)", n+1, n+2, n+3))
		linkArgs = append(linkArgs, w.roleID, w.resource, w.action)
	}
	linkArgs = append(linkArgs, RoleType, policyScope)
	roleTypeArg := len(linkArgs) - 1
	scopeArg := len(linkArgs)

	res, err := tx.ExecContext(ctx, fmt.Sprintf(
		`INSERT INTO role_permission (role_type, role_id, permission_policy_id)
		 SELECT $%d, w.role_id, pp.id
		   FROM (VALUES %s) AS w(role_id, resource, action)
		   JOIN permission_policy pp
		     ON pp.scope = $%d AND pp.resource = w.resource
		    AND pp.action = w.action AND pp.effect = 'allow'`,
		roleTypeArg, strings.Join(linkRows, ", "), scopeArg), linkArgs...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to grant the permissions of the shipped roles")
	}
	n, err := res.RowsAffected()
	if err != nil {
		return len(wanted), nil
	}
	return int(n), nil
}

// roleNeedsSeeding reports whether this shipped role exists and has no grants.
// Whatever is already there wins, so an operator who edited one keeps it.
func roleNeedsSeeding(ctx context.Context, tx *sql.Tx, roleID int64) (bool, error) {
	var exists bool
	if err := tx.QueryRowContext(ctx,
		`SELECT EXISTS (SELECT 1 FROM role WHERE role_id = $1)`, roleID).Scan(&exists); err != nil {
		return false, errors.Wrapf(err, "failed to look for role %d", roleID)
	}
	if !exists {
		return false, nil // not every deployment has every id
	}

	var have int
	if err := tx.QueryRowContext(ctx,
		`SELECT count(*) FROM role_permission WHERE role_type = $1 AND role_id = $2`,
		RoleType, roleID).Scan(&have); err != nil {
		return false, errors.Wrapf(err, "failed to count the grants of role %d", roleID)
	}
	return have == 0, nil
}

func sortedBuiltinIDs() []int64 {
	ids := make([]int64, 0, len(builtinRoles))
	for id := range builtinRoles {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Shipped is what the roles Harbor ships grant, as grants against role ids.
// Seeding writes it into the database; tests build the same thing in memory.
func Shipped() []Grant {
	var out []Grant
	for id, key := range builtinRoles {
		for _, p := range shippedRoles[key] {
			out = append(out, Grant{RoleID: id, Resource: p.Resource.String(), Action: p.Action.String()})
		}
	}
	return out
}
