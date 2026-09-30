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

// Package policy holds what every project role grants, for the life of the
// process.
//
// Harbor used to answer "may this role do this" by building a casbin enforcer
// per request out of a map compiled into the binary. The roles Harbor ships
// and the roles an administrator writes now come from the same place, so there
// is one way to resolve a role and the database is read when the policy
// changes rather than when a user asks a question.
package policy

import (
	"context"
	"database/sql"
	"strconv"
	"sync/atomic"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	"github.com/casbin/casbin/v3/persist"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

// RoleType is the role_permission.role_type a project role's grants carry. It
// is the value controller/role already writes.
const RoleType = "project-role"

// ptypeLive is the casbin policy definition the matcher reads.
const ptypeLive = "p"

// EffectAllow is the effect a grant carries when permission_policy left it
// unset. A deny is carried through as it is stored: the model denies over any
// allow, and dropping it here would hand out a permission the policy refuses.
const EffectAllow = "allow"

// selectGrants reads every project role's grants in one query. It runs at
// startup and when the policy changes, never on a request.
const selectGrants = `
SELECT rp.role_id, pp.resource, pp.action, COALESCE(NULLIF(pp.effect, ''), 'allow')
  FROM role_permission rp
  JOIN permission_policy pp ON pp.id = rp.permission_policy_id
 WHERE rp.role_type = $1`

// Store answers permission questions from memory.
//
// The enforcer is replaced wholesale on reload rather than mutated, so a
// request in flight keeps answering from the policy it started with and no
// reader ever sees a half-applied change.
type Store struct {
	enf atomic.Pointer[casbin.SyncedCachedEnforcer]
	db  *sql.DB

	origin  string
	watcher *Watcher

	generation atomic.Int64

	// applied is the policy_version this replica last loaded. Resync compares
	// it against the database so a lost notification cannot strand us.
	applied atomic.Int64
}

// Grant is one row of the policy: a role, and one thing it may do.
type Grant struct {
	RoleID   int64
	Resource string
	Action   string
	// Effect is what permission_policy stores for the grant. Empty means
	// allow, the same default the policy model has always used.
	Effect string
}

// New builds a store over an already-open database and loads the policy once.
// origin stamps this replica's connections so it can recognise its own writes.
func New(ctx context.Context, db *sql.DB, origin string) (*Store, error) {
	s := &Store{db: db, origin: origin}
	if err := s.Reload(ctx, "startup"); err != nil {
		return nil, err
	}
	return s, nil
}

// NewInMemory builds a store over a fixed set of grants, for tests and for the
// empty store the process holds before the database is up.
func NewInMemory(grants []Grant) (*Store, error) {
	s := &Store{}
	enf, err := build(grants)
	if err != nil {
		return nil, err
	}
	s.enf.Store(enf)
	s.generation.Add(1)
	return s, nil
}

// Close stops the watcher, if there is one.
func (s *Store) Close() {
	if s.watcher != nil {
		s.watcher.Close()
	}
}

// Subject is how a role appears to casbin. The id is what project_member
// stores and what the API hands out, so nothing has to agree on a name.
func Subject(roleID int64) string { return "role:" + strconv.FormatInt(roleID, 10) }

// Enforce answers whether the role may take the action on the object.
//
// object is already in the role's own object space: the caller maps a resource
// into it, because only the caller knows the membership is in project 57 and
// not in project 58.
func (s *Store) Enforce(subject string, object string, action string) (bool, error) {
	enf := s.enf.Load()
	if enf == nil || subject == "" {
		return false, nil
	}
	return enf.Enforce(subject, object, action)
}

// Generation counts how many times this replica has loaded the policy.
func (s *Store) Generation() int64 { return s.generation.Load() }

// AppliedVersion is the policy version this replica is serving. An operator
// comparing it across replicas can see a stale one without waiting for a user
// to be refused something.
func (s *Store) AppliedVersion() int64 { return s.applied.Load() }

// build makes an enforcer holding exactly these grants.
func build(grants []Grant) (*casbin.SyncedCachedEnforcer, error) {
	m, err := model.NewModelFromString(Model)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse the policy model")
	}

	for _, g := range grants {
		// A fresh slice per rule: casbin keeps the one it is given.
		effect := g.Effect
		if effect == "" {
			effect = EffectAllow
		}
		rule := []string{ptypeLive, Subject(g.RoleID), Object(g.Resource), g.Action, effect}
		if err := persist.LoadPolicyArray(rule, m); err != nil {
			return nil, errors.Wrap(err, "failed to load the policy")
		}
	}

	enf, err := casbin.NewSyncedCachedEnforcer(m)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build the enforcer")
	}
	enf.AddFunction("keyMatch2", KeyMatch2Func)
	return enf, nil
}

// Reload reads every role's grants and swaps in an enforcer holding them.
//
// It reloads everything rather than applying a delta, on the grounds that a
// replica which missed an unknown number of changes should not try to be
// clever. A full reload converges after a lost notification; a delta stream
// does not.
func (s *Store) Reload(ctx context.Context, reason string) error {
	if s.db == nil {
		return nil
	}

	// Read the version before the grants, never after. A write that lands
	// mid-reload then leaves us recorded as older than we are, and the next
	// resync picks it up. The other order would record us as current while
	// holding a policy from before that write.
	version, err := s.policyVersion(ctx)
	if err != nil {
		return err
	}

	grants, err := s.load(ctx)
	if err != nil {
		return err
	}
	enf, err := build(grants)
	if err != nil {
		return err
	}

	s.enf.Store(enf)
	s.applied.Store(version)
	s.generation.Add(1)
	log.Debugf("policy reloaded (%s): %d grants, generation %d, version %d",
		reason, len(grants), s.generation.Load(), version)
	return nil
}

func (s *Store) load(ctx context.Context) ([]Grant, error) {
	rows, err := s.db.QueryContext(ctx, selectGrants, RoleType)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the role permissions")
	}
	defer rows.Close()

	var out []Grant
	for rows.Next() {
		var g Grant
		if err := rows.Scan(&g.RoleID, &g.Resource, &g.Action, &g.Effect); err != nil {
			return nil, errors.Wrap(err, "failed to read a role permission")
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to read the role permissions")
	}
	return out, nil
}

// policyVersion reads the counter every policy write bumps.
func (s *Store) policyVersion(ctx context.Context) (int64, error) {
	if s.db == nil {
		return 0, nil
	}
	var v int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT version FROM policy_version WHERE only_row`).Scan(&v); err != nil {
		return 0, errors.Wrap(err, "failed to read the policy version")
	}
	return v, nil
}

// Resync reloads when the database has moved past what this replica applied.
//
// This is the safety net under the watcher rather than a replacement for it.
// A notification arrives in milliseconds; this catches the cases where no
// notification arrives at all, which in a fleet of cores is the case that
// matters: a connection that dropped between the write and the reconnect, a
// Postgres restart, a replica paused long enough to miss the message.
func (s *Store) Resync(ctx context.Context) (bool, error) {
	current, err := s.policyVersion(ctx)
	if err != nil {
		return false, err
	}
	if current == s.applied.Load() {
		return false, nil
	}
	log.Infof("policy version %d in the database, %d here: reloading", current, s.applied.Load())
	if err := s.Reload(ctx, "resync"); err != nil {
		return false, err
	}
	return true, nil
}

// Watch starts the LISTEN/NOTIFY watcher. connString is a libpq-style DSN for a
// connection of this replica's own: a connection parked in LISTEN answers no
// queries, so it must not come out of the pool.
func (s *Store) Watch(connString string) error {
	w, err := NewWatcher(connString, Channel, s.origin,
		func(ctx context.Context, payload string) error {
			return s.Reload(ctx, "notify "+payload)
		},
		func(ctx context.Context) error {
			_, err := s.Resync(ctx)
			return err
		})
	if err != nil {
		return err
	}
	s.watcher = w
	return nil
}

// WatcherStats reports what the watcher has seen, or nil when there is none.
func (s *Store) WatcherStats() *WatcherStats {
	if s.watcher == nil {
		return nil
	}
	st := s.watcher.Stats()
	return &st
}
