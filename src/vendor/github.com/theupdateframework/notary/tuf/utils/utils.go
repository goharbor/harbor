package utils

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/theupdateframework/notary/tuf/data"
)

// StrSliceContains checks if the given string appears in the slice
func StrSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// RoleNameSliceContains checks if the given string appears in the slice
func RoleNameSliceContains(ss []data.RoleName, s data.RoleName) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

// RoleNameSliceRemove removes the the given RoleName from the slice, returning a new slice
func RoleNameSliceRemove(ss []data.RoleName, s data.RoleName) []data.RoleName {
	res := []data.RoleName{}
	for _, v := range ss {
		if v != s {
			res = append(res, v)
		}
	}
	return res
}

// NoopCloser is a simple Reader wrapper that does nothing when Close is
// called
type NoopCloser struct {
	io.Reader
}

// Close does nothing for a NoopCloser
func (nc *NoopCloser) Close() error {
	return nil
}

// DoHash returns the digest of d using the hashing algorithm named
// in alg
func DoHash(alg string, d []byte) []byte {
	switch alg {
	case "sha256":
		digest := sha256.Sum256(d)
		return digest[:]
	case "sha512":
		digest := sha512.Sum512(d)
		return digest[:]
	}
	return nil
}

// UnusedDelegationKeys prunes a list of keys, returning those that are no
// longer in use for a given targets file
func UnusedDelegationKeys(t data.SignedTargets) []string {
	// compare ids to all still active key ids in all active roles
	// with the targets file
	found := make(map[string]bool)
	for _, r := range t.Signed.Delegations.Roles {
		for _, id := range r.KeyIDs {
			found[id] = true
		}
	}
	var discard []string
	for id := range t.Signed.Delegations.Keys {
		if !found[id] {
			discard = append(discard, id)
		}
	}
	return discard
}

// RemoveUnusedKeys determines which keys in the slice of IDs are no longer
// used in the given targets file and removes them from the delegated keys
// map
func RemoveUnusedKeys(t *data.SignedTargets) {
	unusedIDs := UnusedDelegationKeys(*t)
	for _, id := range unusedIDs {
		delete(t.Signed.Delegations.Keys, id)
	}
}

// FindRoleIndex returns the index of the role named <name> or -1 if no
// matching role is found.
func FindRoleIndex(rs []*data.Role, name data.RoleName) int {
	for i, r := range rs {
		if r.Name == name {
			return i
		}
	}
	return -1
}

// ConsistentName generates the appropriate HTTP URL path for the role,
// based on whether the repo is marked as consistent. The RemoteStore
// is responsible for adding file extensions.
func ConsistentName(role string, hashSHA256 []byte) string {
	if len(hashSHA256) > 0 {
		hash := hex.EncodeToString(hashSHA256)
		return fmt.Sprintf("%s.%s", role, hash)
	}
	return role
}
