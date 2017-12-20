package storage

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary"
	"github.com/docker/notary/storage"
	"github.com/docker/notary/tuf/data"
)

// TUFMetaStorage wraps a MetaStore in order to walk the TUF tree for GetCurrent in a consistent manner,
// by always starting from a current timestamp and then looking up other data by hash
type TUFMetaStorage struct {
	MetaStore
	// cached metadata by checksum
	cachedMeta map[string]*storedMeta
}

// NewTUFMetaStorage instantiates a TUFMetaStorage instance
func NewTUFMetaStorage(m MetaStore) *TUFMetaStorage {
	return &TUFMetaStorage{
		MetaStore:  m,
		cachedMeta: make(map[string]*storedMeta),
	}
}

type storedMeta struct {
	data         []byte
	createupdate *time.Time
}

// GetCurrent gets a specific TUF record, by walking from the current Timestamp to other metadata by checksum
func (tms TUFMetaStorage) GetCurrent(gun data.GUN, tufRole data.RoleName) (*time.Time, []byte, error) {
	timestampTime, timestampJSON, err := tms.MetaStore.GetCurrent(gun, data.CanonicalTimestampRole)
	if err != nil {
		return nil, nil, err
	}
	// If we wanted data for the timestamp role, we're done here
	if tufRole == data.CanonicalTimestampRole {
		return timestampTime, timestampJSON, nil
	}

	// If we want to lookup another role, walk to it via current timestamp --> snapshot by checksum --> desired role
	timestampMeta := &data.SignedTimestamp{}
	if err := json.Unmarshal(timestampJSON, timestampMeta); err != nil {
		return nil, nil, fmt.Errorf("could not parse current timestamp")
	}
	snapshotChecksums, err := timestampMeta.GetSnapshot()
	if err != nil || snapshotChecksums == nil {
		return nil, nil, fmt.Errorf("could not retrieve latest snapshot checksum")
	}
	snapshotSHA256Bytes, ok := snapshotChecksums.Hashes[notary.SHA256]
	if !ok {
		return nil, nil, fmt.Errorf("could not retrieve latest snapshot sha256")
	}
	snapshotSHA256Hex := hex.EncodeToString(snapshotSHA256Bytes[:])

	// Check the cache if we have our snapshot data
	var snapshotTime *time.Time
	var snapshotJSON []byte
	if cachedSnapshotData, ok := tms.cachedMeta[snapshotSHA256Hex]; ok {
		snapshotTime = cachedSnapshotData.createupdate
		snapshotJSON = cachedSnapshotData.data
	} else {
		// Get the snapshot from the underlying store by checksum if it isn't cached yet
		snapshotTime, snapshotJSON, err = tms.GetChecksum(gun, data.CanonicalSnapshotRole, snapshotSHA256Hex)
		if err != nil {
			return nil, nil, err
		}
		// cache for subsequent lookups
		tms.cachedMeta[snapshotSHA256Hex] = &storedMeta{data: snapshotJSON, createupdate: snapshotTime}
	}
	// If we wanted data for the snapshot role, we're done here
	if tufRole == data.CanonicalSnapshotRole {
		return snapshotTime, snapshotJSON, nil
	}

	// If it's a different role, we should have the checksum in snapshot metadata, and we can use it to GetChecksum()
	snapshotMeta := &data.SignedSnapshot{}
	if err := json.Unmarshal(snapshotJSON, snapshotMeta); err != nil {
		return nil, nil, fmt.Errorf("could not parse current snapshot")
	}
	roleMeta, err := snapshotMeta.GetMeta(tufRole)
	if err != nil {
		return nil, nil, err
	}
	roleSHA256Bytes, ok := roleMeta.Hashes[notary.SHA256]
	if !ok {
		return nil, nil, fmt.Errorf("could not retrieve latest %s sha256", tufRole)
	}
	roleSHA256Hex := hex.EncodeToString(roleSHA256Bytes[:])
	// check if we can retrieve this data from cache
	if cachedRoleData, ok := tms.cachedMeta[roleSHA256Hex]; ok {
		return cachedRoleData.createupdate, cachedRoleData.data, nil
	}

	roleTime, roleJSON, err := tms.MetaStore.GetChecksum(gun, tufRole, roleSHA256Hex)
	if err != nil {
		return nil, nil, err
	}
	// cache for subsequent lookups
	tms.cachedMeta[roleSHA256Hex] = &storedMeta{data: roleJSON, createupdate: roleTime}
	return roleTime, roleJSON, nil
}

// GetChecksum gets a specific TUF record by checksum, also checking the internal cache
func (tms TUFMetaStorage) GetChecksum(gun data.GUN, tufRole data.RoleName, checksum string) (*time.Time, []byte, error) {
	if cachedRoleData, ok := tms.cachedMeta[checksum]; ok {
		return cachedRoleData.createupdate, cachedRoleData.data, nil
	}
	roleTime, roleJSON, err := tms.MetaStore.GetChecksum(gun, tufRole, checksum)
	if err != nil {
		return nil, nil, err
	}
	// cache for subsequent lookups
	tms.cachedMeta[checksum] = &storedMeta{data: roleJSON, createupdate: roleTime}
	return roleTime, roleJSON, nil
}

// Bootstrap the store with tables if possible
func (tms TUFMetaStorage) Bootstrap() error {
	if s, ok := tms.MetaStore.(storage.Bootstrapper); ok {
		return s.Bootstrap()
	}
	return fmt.Errorf("store does not support bootstrapping")
}
