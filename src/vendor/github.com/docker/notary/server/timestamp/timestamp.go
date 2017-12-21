package timestamp

import (
	"encoding/hex"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"

	"github.com/Sirupsen/logrus"
	"github.com/docker/notary/server/snapshot"
	"github.com/docker/notary/server/storage"
)

// GetOrCreateTimestampKey returns the timestamp key for the gun. It uses the store to
// lookup an existing timestamp key and the crypto to generate a new one if none is
// found. It attempts to handle the race condition that may occur if 2 servers try to
// create the key at the same time by simply querying the store a second time if it
// receives a conflict when writing.
func GetOrCreateTimestampKey(gun data.GUN, store storage.MetaStore, crypto signed.CryptoService, createAlgorithm string) (data.PublicKey, error) {
	_, rootJSON, err := store.GetCurrent(gun, data.CanonicalRootRole)
	if err != nil {
		// If the error indicates we couldn't find the root, create a new key
		if _, ok := err.(storage.ErrNotFound); !ok {
			logrus.Errorf("Error when retrieving root role for GUN %s: %v", gun, err)
			return nil, err
		}
		return crypto.Create(data.CanonicalTimestampRole, gun, createAlgorithm)
	}

	// If we have a current root, parse out the public key for the timestamp role, and return it
	repoSignedRoot := new(data.SignedRoot)
	if err := json.Unmarshal(rootJSON, repoSignedRoot); err != nil {
		logrus.Errorf("Failed to unmarshal existing root for GUN %s to retrieve timestamp key ID", gun)
		return nil, err
	}

	timestampRole, err := repoSignedRoot.BuildBaseRole(data.CanonicalTimestampRole)
	if err != nil {
		logrus.Errorf("Failed to extract timestamp role from root for GUN %s", gun)
		return nil, err
	}

	// We currently only support single keys for snapshot and timestamp, so we can return the first and only key in the map if the signer has it
	for keyID := range timestampRole.Keys {
		if pubKey := crypto.GetKey(keyID); pubKey != nil {
			return pubKey, nil
		}
	}
	logrus.Debugf("Failed to find any timestamp keys in cryptosigner from root for GUN %s, generating new key", gun)
	return crypto.Create(data.CanonicalTimestampRole, gun, createAlgorithm)
}

// RotateTimestampKey attempts to rotate a timestamp key in the signer, but might be rate-limited by the signer
func RotateTimestampKey(gun data.GUN, store storage.MetaStore, crypto signed.CryptoService, createAlgorithm string) (data.PublicKey, error) {
	// Always attempt to create a new key, but this might be rate-limited
	key, err := crypto.Create(data.CanonicalTimestampRole, gun, createAlgorithm)
	if err != nil {
		return nil, err
	}
	logrus.Debug("Created new pending timestamp key ", key.ID(), "to rotate to for ", gun, ". With algo: ", key.Algorithm())
	return key, nil
}

// GetOrCreateTimestamp returns the current timestamp for the gun. This may mean
// a new timestamp is generated either because none exists, or because the current
// one has expired. Once generated, the timestamp is saved in the store.
// Additionally, if we had to generate a new snapshot for this timestamp,
// it is also saved in the store
func GetOrCreateTimestamp(gun data.GUN, store storage.MetaStore, cryptoService signed.CryptoService) (
	*time.Time, []byte, error) {

	updates := []storage.MetaUpdate{}

	lastModified, timestampJSON, err := store.GetCurrent(gun, data.CanonicalTimestampRole)
	if err != nil {
		logrus.Debug("error retrieving timestamp: ", err.Error())
		return nil, nil, err
	}

	prev := &data.SignedTimestamp{}
	if err := json.Unmarshal(timestampJSON, prev); err != nil {
		logrus.Error("Failed to unmarshal existing timestamp")
		return nil, nil, err
	}
	snapChecksums, err := prev.GetSnapshot()
	if err != nil || snapChecksums == nil {
		return nil, nil, err
	}
	snapshotSHA256Bytes, ok := snapChecksums.Hashes[notary.SHA256]
	if !ok {
		return nil, nil, data.ErrMissingMeta{Role: data.CanonicalSnapshotRole.String()}
	}
	snapshotSHA256Hex := hex.EncodeToString(snapshotSHA256Bytes[:])
	snapshotTime, snapshot, err := snapshot.GetOrCreateSnapshot(gun, snapshotSHA256Hex, store, cryptoService)
	if err != nil {
		logrus.Debug("Previous timestamp, but no valid snapshot for GUN ", gun)
		return nil, nil, err
	}
	snapshotRole := &data.SignedSnapshot{}
	if err := json.Unmarshal(snapshot, snapshotRole); err != nil {
		logrus.Error("Failed to unmarshal retrieved snapshot")
		return nil, nil, err
	}

	// If the snapshot was generated, we should write it with the timestamp
	if snapshotTime == nil {
		updates = append(updates, storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: snapshotRole.Signed.Version, Data: snapshot})
	}

	if !timestampExpired(prev) && !snapshotExpired(prev, snapshot) {
		return lastModified, timestampJSON, nil
	}

	tsUpdate, err := createTimestamp(gun, prev, snapshot, store, cryptoService)
	if err != nil {
		logrus.Error("Failed to create a new timestamp")
		return nil, nil, err
	}
	updates = append(updates, *tsUpdate)

	c := time.Now()

	// Write the timestamp, and potentially snapshot
	if err = store.UpdateMany(gun, updates); err != nil {
		return nil, nil, err
	}
	return &c, tsUpdate.Data, nil
}

// timestampExpired compares the current time to the expiry time of the timestamp
func timestampExpired(ts *data.SignedTimestamp) bool {
	return signed.IsExpired(ts.Signed.Expires)
}

// snapshotExpired verifies the checksum(s) for the given snapshot using metadata from the timestamp
func snapshotExpired(ts *data.SignedTimestamp, snapshot []byte) bool {
	// If this check failed, it means the current snapshot was not exactly what we expect
	// via the timestamp. So we can consider it to be "expired."
	return data.CheckHashes(snapshot, data.CanonicalSnapshotRole.String(),
		ts.Signed.Meta[data.CanonicalSnapshotRole.String()].Hashes) != nil
}

// CreateTimestamp creates a new timestamp. If a prev timestamp is provided, it
// is assumed this is the immediately previous one, and the new one will have a
// version number one higher than prev. The store is used to lookup the current
// snapshot, this function does not save the newly generated timestamp.
func createTimestamp(gun data.GUN, prev *data.SignedTimestamp, snapshot []byte, store storage.MetaStore,
	cryptoService signed.CryptoService) (*storage.MetaUpdate, error) {

	builder := tuf.NewRepoBuilder(gun, cryptoService, trustpinning.TrustPinConfig{})

	// load the current root to ensure we use the correct timestamp key.
	_, root, err := store.GetCurrent(gun, data.CanonicalRootRole)
	if err != nil {
		logrus.Debug("Previous timestamp, but no root for GUN ", gun)
		return nil, err
	}
	if err := builder.Load(data.CanonicalRootRole, root, 1, false); err != nil {
		logrus.Debug("Could not load valid previous root for GUN ", gun)
		return nil, err
	}

	// load snapshot so we can include it in timestamp
	if err := builder.Load(data.CanonicalSnapshotRole, snapshot, 1, false); err != nil {
		logrus.Debug("Could not load valid previous snapshot for GUN ", gun)
		return nil, err
	}

	meta, ver, err := builder.GenerateTimestamp(prev)
	if err != nil {
		return nil, err
	}
	return &storage.MetaUpdate{
		Role:    data.CanonicalTimestampRole,
		Version: ver,
		Data:    meta,
	}, nil
}
