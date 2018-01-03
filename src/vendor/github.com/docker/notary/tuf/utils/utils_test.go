package utils

import (
	"testing"

	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

func TestUnusedDelegationKeys(t *testing.T) {
	targets := data.NewTargets()

	role, err := data.NewRole("targets/test", 1, []string{}, []string{""})
	require.NoError(t, err)

	discard := UnusedDelegationKeys(*targets)
	require.Len(t, discard, 0)

	targets.Signed.Delegations.Roles = []*data.Role{role}
	targets.Signed.Delegations.Keys["123"] = nil

	discard = UnusedDelegationKeys(*targets)
	require.Len(t, discard, 1)

	role.KeyIDs = []string{"123"}

	discard = UnusedDelegationKeys(*targets)
	require.Len(t, discard, 0)
}

func TestRemoveUnusedKeys(t *testing.T) {
	targets := data.NewTargets()

	role, err := data.NewRole("targets/test", 1, []string{"123"}, []string{""})
	require.NoError(t, err)

	targets.Signed.Delegations.Keys["123"] = nil

	RemoveUnusedKeys(targets)
	require.Len(t, targets.Signed.Delegations.Keys, 0)

	// when role is present that uses key, it shouldn't get removed
	targets.Signed.Delegations.Roles = []*data.Role{role}
	targets.Signed.Delegations.Keys["123"] = nil

	RemoveUnusedKeys(targets)
	require.Len(t, targets.Signed.Delegations.Keys, 1)
}

func TestFindRoleIndexFound(t *testing.T) {
	role, err := data.NewRole("targets/test", 1, []string{}, []string{""})
	require.NoError(t, err)

	require.Equal(
		t,
		0,
		FindRoleIndex([]*data.Role{role}, role.Name),
	)
}

func TestFindRoleIndexNotFound(t *testing.T) {
	role, err := data.NewRole("targets/test", 1, []string{}, []string{""})
	require.NoError(t, err)

	require.Equal(
		t,
		-1,
		FindRoleIndex(nil, role.Name),
	)
}
