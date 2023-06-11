package fsutil

import (
	"os"
)

// EnsureMaxPermissions tests the provided file info to make sure the
// permission bits matches the provided.
// On Windows system the permission bits are not really compatible with
// UNIX-like permission bits. By setting the UNIX-like permission bits
// on a Windows system only read/write by all users can be achieved.
// See this issue for tracking and more details:
// https://github.com/theupdateframework/go-tuf/issues/360
// Currently this method will always return nil.
func EnsureMaxPermissions(fi os.FileInfo, perm os.FileMode) error {
	return nil
}
