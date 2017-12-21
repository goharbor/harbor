package storage

import (
	"time"

	"github.com/docker/notary/tuf/data"
)

// KeyStore provides a minimal interface for managing key persistence
type KeyStore interface {
	// GetKey returns the algorithm and public key for the given GUN and role.
	// If the GUN+role don't exist, returns an error.
	GetKey(gun, role string) (algorithm string, public []byte, err error)

	// SetKey sets the algorithm and public key for the given GUN and role if
	// it doesn't already exist.  Otherwise an error is returned.
	SetKey(gun, role, algorithm string, public []byte) error
}

// MetaStore holds the methods that are used for a Metadata Store
type MetaStore interface {
	// UpdateCurrent adds new metadata version for the given GUN if and only
	// if it's a new role, or the version is greater than the current version
	// for the role. Otherwise an error is returned.
	UpdateCurrent(gun data.GUN, update MetaUpdate) error

	// UpdateMany adds multiple new metadata for the given GUN.  It can even
	// add multiple versions for the same role, so long as those versions are
	// all unique and greater than any current versions.  Otherwise,
	// none of the metadata is added, and an error is be returned.
	UpdateMany(gun data.GUN, updates []MetaUpdate) error

	// GetCurrent returns the modification date and data part of the metadata for
	// the latest version of the given GUN and role.  If there is no data for
	// the given GUN and role, an error is returned.
	GetCurrent(gun data.GUN, tufRole data.RoleName) (created *time.Time, data []byte, err error)

	// GetChecksum returns the given TUF role file and creation date for the
	// GUN with the provided checksum. If the given (gun, role, checksum) are
	// not found, it returns storage.ErrNotFound
	GetChecksum(gun data.GUN, tufRole data.RoleName, checksum string) (created *time.Time, data []byte, err error)

	// GetVersion returns the given TUF role file and creation date for the
	// GUN with the provided version. If the given (gun, role, version) are
	// not found, it returns storage.ErrNotFound
	GetVersion(gun data.GUN, tufRole data.RoleName, version int) (created *time.Time, data []byte, err error)

	// Delete removes all metadata for a given GUN.  It does not return an
	// error if no metadata exists for the given GUN.
	Delete(gun data.GUN) error

	// GetChanges returns an ordered slice of changes. It starts from
	// the change matching changeID, but excludes this change from the results
	// on the assumption that if a user provides an ID, they've seen that change.
	// If changeID is 0, it starts from the
	// beginning, and if changeID is -1, it starts from the most recent
	// change. The number of results returned is limited by records.
	// If records is negative, we will return that number of changes preceding
	// the given changeID.
	// The returned []Change should always be ordered oldest to newest.
	GetChanges(changeID string, records int, filterName string) ([]Change, error)
}
