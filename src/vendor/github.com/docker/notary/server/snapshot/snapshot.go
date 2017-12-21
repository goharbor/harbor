package snapshot

import (
	"time"

	"github.com/Sirupsen/logrus"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
)

// GetOrCreateSnapshotKey either creates a new snapshot key, or returns
// the existing one. Only the PublicKey is returned. The private part
// is held by the CryptoService.
func GetOrCreateSnapshotKey(gun data.GUN, store storage.MetaStore, crypto signed.CryptoService, createAlgorithm string) (data.PublicKey, error) {
	_, rootJSON, err := store.GetCurrent(gun, data.CanonicalRootRole)
	if err != nil {
		// If the error indicates we couldn't find the root, create a new key
		if _, ok := err.(storage.ErrNotFound); !ok {
			logrus.Errorf("Error when retrieving root role for GUN %s: %v", gun.String(), err)
			return nil, err
		}
		return crypto.Create(data.CanonicalSnapshotRole, gun, createAlgorithm)
	}

	// If we have a current root, parse out the public key for the snapshot role, and return it
	repoSignedRoot := new(data.SignedRoot)
	if err := json.Unmarshal(rootJSON, repoSignedRoot); err != nil {
		logrus.Errorf("Failed to unmarshal existing root for GUN %s to retrieve snapshot key ID", gun)
		return nil, err
	}

	snapshotRole, err := repoSignedRoot.BuildBaseRole(data.CanonicalSnapshotRole)
	if err != nil {
		logrus.Errorf("Failed to extract snapshot role from root for GUN %s", gun)
		return nil, err
	}

	// We currently only support single keys for snapshot and timestamp, so we can return the first and only key in the map if the signer has it
	for keyID := range snapshotRole.Keys {
		if pubKey := crypto.GetKey(keyID); pubKey != nil {
			return pubKey, nil
		}
	}
	logrus.Debugf("Failed to find any snapshot keys in cryptosigner from root for GUN %s, generating new key", gun)
	return crypto.Create(data.CanonicalSnapshotRole, gun, createAlgorithm)
}

// RotateSnapshotKey attempts to rotate a snapshot key in the signer, but might be rate-limited by the signer
func RotateSnapshotKey(gun data.GUN, store storage.MetaStore, crypto signed.CryptoService, createAlgorithm string) (data.PublicKey, error) {
	// Always attempt to create a new key, but this might be rate-limited
	key, err := crypto.Create(data.CanonicalSnapshotRole, gun, createAlgorithm)
	if err != nil {
		return nil, err
	}
	logrus.Debug("Created new pending snapshot key ", key.ID(), "to rotate to for ", gun, ". With algo: ", key.Algorithm())
	return key, nil
}

// GetOrCreateSnapshot either returns the existing latest snapshot, or uses
// whatever the most recent snapshot is to generate the next one, only updating
// the expiry time and version.  Note that this function does not write generated
// snapshots to the underlying data store, and will either return the latest snapshot time
// or nil as the time modified
func GetOrCreateSnapshot(gun data.GUN, checksum string, store storage.MetaStore, cryptoService signed.CryptoService) (
	*time.Time, []byte, error) {

	lastModified, currentJSON, err := store.GetChecksum(gun, data.CanonicalSnapshotRole, checksum)
	if err != nil {
		return nil, nil, err
	}

	prev := new(data.SignedSnapshot)
	if err := json.Unmarshal(currentJSON, prev); err != nil {
		logrus.Error("Failed to unmarshal existing snapshot for GUN ", gun)
		return nil, nil, err
	}

	if !snapshotExpired(prev) {
		return lastModified, currentJSON, nil
	}

	builder := tuf.NewRepoBuilder(gun, cryptoService, trustpinning.TrustPinConfig{})

	// load the current root to ensure we use the correct snapshot key.
	_, rootJSON, err := store.GetCurrent(gun, data.CanonicalRootRole)
	if err != nil {
		logrus.Debug("Previous snapshot, but no root for GUN ", gun)
		return nil, nil, err
	}
	if err := builder.Load(data.CanonicalRootRole, rootJSON, 1, false); err != nil {
		logrus.Debug("Could not load valid previous root for GUN ", gun)
		return nil, nil, err
	}

	meta, _, err := builder.GenerateSnapshot(prev)
	if err != nil {
		return nil, nil, err
	}

	return nil, meta, nil
}

// snapshotExpired simply checks if the snapshot is past its expiry time
func snapshotExpired(sn *data.SignedSnapshot) bool {
	return signed.IsExpired(sn.Signed.Expires)
}
