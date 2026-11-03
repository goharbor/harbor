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

// Package policy gives tests the policy Harbor seeds, without a database.
//
// It is built from the same definition the seeder writes, so a test that
// passes here is evaluating what a real Harbor holds rather than a second copy
// that can drift.
package policy

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/permission/policy"
)

// NewStore returns a store holding the grants of the roles Harbor ships.
func NewStore() (*policy.Store, error) {
	return policy.NewInMemory(policy.Shipped())
}

// NewStoreWith returns those grants plus whatever the caller adds, for tests
// about a role Harbor does not ship.
func NewStoreWith(extra ...policy.Grant) (*policy.Store, error) {
	return policy.NewInMemory(append(policy.Shipped(), extra...))
}

// Seed installs the seeded store as the process default for the duration of the
// test, and puts the previous one back afterwards.
func Seed(t testing.TB) *policy.Store {
	t.Helper()
	s, err := NewStore()
	if err != nil {
		t.Fatalf("failed to build the seeded policy store: %v", err)
	}
	previous := policy.SetDefault(s)
	t.Cleanup(func() { policy.SetDefault(previous) })
	return s
}

// SeedWith installs the shipped policy plus extra grants as the process
// default, for a test about a role Harbor does not ship.
func SeedWith(t testing.TB, extra ...policy.Grant) *policy.Store {
	t.Helper()
	s, err := NewStoreWith(extra...)
	if err != nil {
		t.Fatalf("failed to build the policy store: %v", err)
	}
	previous := policy.SetDefault(s)
	t.Cleanup(func() { policy.SetDefault(previous) })
	return s
}
