//go:build !windows
// +build !windows

package fsutil

import (
	"fmt"
	"os"
)

// EnsureMaxPermissions tests the provided file info, returning an error if the
// file's permission bits contain excess permissions not set in maxPerms.
//
// For example, a file with permissions -rw------- will successfully validate
// with maxPerms -rw-r--r-- or -rw-rw-r--, but will not validate with maxPerms
// -r-------- (due to excess --w------- permission) or --w------- (due to
// excess -r-------- permission).
//
// Only permission bits of the file modes are considered.
func EnsureMaxPermissions(fi os.FileInfo, maxPerms os.FileMode) error {
	gotPerm := fi.Mode().Perm()
	forbiddenPerms := (^maxPerms).Perm()
	excessPerms := gotPerm & forbiddenPerms

	if excessPerms != 0 {
		return fmt.Errorf("permission bits for file %v failed validation: want at most %v, got %v with excess perms %v", fi.Name(), maxPerms.Perm(), gotPerm, excessPerms)
	}

	return nil
}
