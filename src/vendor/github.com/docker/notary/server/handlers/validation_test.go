package handlers

import (
	"bytes"
	"path"
	"testing"
	"time"

	"github.com/docker/go/canonical/json"

	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/testutils"
	"github.com/docker/notary/tuf/validation"
	"github.com/stretchr/testify/require"
)

// this is a fake storage that serves errors
type getFailStore struct {
	errsToReturn map[string]error
	storage.MetaStore
}

// GetCurrent returns the current metadata, or an error depending on whether
// getFailStore is configured to return an error for this role
func (f getFailStore) GetCurrent(gun data.GUN, tufRole data.RoleName) (*time.Time, []byte, error) {
	err := f.errsToReturn[tufRole.String()]
	if err == nil {
		return f.MetaStore.GetCurrent(gun, tufRole)
	}
	return nil, nil, err
}

// GetChecksum returns the metadata with this checksum, or an error depending on
// whether getFailStore is configured to return an error for this role
func (f getFailStore) GetChecksum(gun data.GUN, tufRole data.RoleName, checksum string) (*time.Time, []byte, error) {
	err := f.errsToReturn[tufRole.String()]
	if err == nil {
		return f.MetaStore.GetChecksum(gun, tufRole, checksum)
	}
	return nil, nil, err
}

// Returns a mapping of role name to `MetaUpdate` objects
func getUpdates(r, tg, sn, ts *data.Signed) (
	root, targets, snapshot, timestamp storage.MetaUpdate, err error) {

	rs, tgs, sns, tss, err := testutils.Serialize(r, tg, sn, ts)
	if err != nil {
		return
	}

	root = storage.MetaUpdate{
		Role:    data.CanonicalRootRole,
		Version: 1,
		Data:    rs,
	}
	targets = storage.MetaUpdate{
		Role:    data.CanonicalTargetsRole,
		Version: 1,
		Data:    tgs,
	}
	snapshot = storage.MetaUpdate{
		Role:    data.CanonicalSnapshotRole,
		Version: 1,
		Data:    sns,
	}
	timestamp = storage.MetaUpdate{
		Role:    data.CanonicalTimestampRole,
		Version: 1,
		Data:    tss,
	}
	return
}

func TestValidateEmptyNew(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	updates, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)

	// we generated our own timestamp, and did not take the other timestamp,
	// but all other metadata should come from updates
	founds := make(map[data.RoleName]bool)
	for _, update := range updates {
		switch update.Role {
		case data.CanonicalRootRole:
			require.True(t, bytes.Equal(update.Data, root.Data))
			founds[data.CanonicalRootRole] = true
		case data.CanonicalSnapshotRole:
			require.True(t, bytes.Equal(update.Data, snapshot.Data))
			founds[data.CanonicalSnapshotRole] = true
		case data.CanonicalTargetsRole:
			require.True(t, bytes.Equal(update.Data, targets.Data))
			founds[data.CanonicalTargetsRole] = true
		case data.CanonicalTimestampRole:
			require.False(t, bytes.Equal(update.Data, timestamp.Data))
			founds[data.CanonicalTimestampRole] = true
		}
	}
	require.Len(t, founds, 4)
}

func TestValidateRootCanContainOnlyx509KeysWithRightGun(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo("wrong/gun")
	require.NoError(t, err)
	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)

	// if the root has the wrong gun, the server will fail to validate
	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	_, err = validateUpdate(serverCrypto, gun,
		[]storage.MetaUpdate{root, targets, snapshot, timestamp},
		storage.NewMemStorage())
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)

	// create regular non-x509 keys - change the root keys to one that is not
	// an x509 key - it should also fail to validate
	newRootKey, err := cs.Create(data.CanonicalRootRole, gun, data.ECDSAKey)
	require.NoError(t, err)
	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, newRootKey))

	r, tg, sn, ts, err = testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err = getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	_, err = validateUpdate(serverCrypto, gun,
		[]storage.MetaUpdate{root, targets, snapshot, timestamp},
		storage.NewMemStorage())
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

func TestValidatePrevTimestamp(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot}

	store := storage.NewMemStorage()
	store.UpdateCurrent(gun, timestamp)

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	updates, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)

	// we generated our own timestamp, and did not take the other timestamp,
	// but all other metadata should come from updates
	var foundTimestamp bool
	for _, update := range updates {
		if update.Role == data.CanonicalTimestampRole {
			foundTimestamp = true
			oldTimestamp, newTimestamp := &data.SignedTimestamp{}, &data.SignedTimestamp{}
			require.NoError(t, json.Unmarshal(timestamp.Data, oldTimestamp))
			require.NoError(t, json.Unmarshal(update.Data, newTimestamp))
			require.Equal(t, oldTimestamp.Signed.Version+1, newTimestamp.Signed.Version)
		}
	}
	require.True(t, foundTimestamp)
}

func TestValidatePreviousTimestampCorrupt(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot}

	// have corrupt timestamp data in the storage already
	store := storage.NewMemStorage()
	timestamp.Data = timestamp.Data[1:]
	store.UpdateCurrent(gun, timestamp)

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, &json.SyntaxError{}, err)
}

func TestValidateGetCurrentTimestampBroken(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot}

	store := getFailStore{
		MetaStore:    storage.NewMemStorage(),
		errsToReturn: map[string]error{data.CanonicalTimestampRole.String(): data.ErrNoSuchRole{}},
	}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, data.ErrNoSuchRole{}, err)
}

func TestValidateNoNewRoot(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	store.UpdateCurrent(gun, root)
	updates := []storage.MetaUpdate{targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)
}

func TestValidateNoNewTargets(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	store.UpdateCurrent(gun, targets)
	updates := []storage.MetaUpdate{root, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)
}

func TestValidateOnlySnapshot(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	store.UpdateCurrent(gun, root)
	store.UpdateCurrent(gun, targets)

	updates := []storage.MetaUpdate{snapshot}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)
}

func TestValidateOldRoot(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	store.UpdateCurrent(gun, root)
	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)
}

func TestValidateOldRootCorrupt(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	badRoot := storage.MetaUpdate{
		Version: root.Version,
		Role:    root.Role,
		Data:    root.Data[1:],
	}
	store.UpdateCurrent(gun, badRoot)
	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, &json.SyntaxError{}, err)
}

// We cannot validate a new root if the old root is corrupt, because there might
// have been a root key rotation.
func TestValidateOldRootCorruptRootRole(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	// so a valid root, but missing the root role
	signedRoot, err := data.RootFromSigned(r)
	require.NoError(t, err)
	delete(signedRoot.Signed.Roles, data.CanonicalRootRole)
	badRootJSON, err := json.Marshal(signedRoot)
	require.NoError(t, err)
	badRoot := storage.MetaUpdate{
		Version: root.Version,
		Role:    root.Role,
		Data:    badRootJSON,
	}
	store.UpdateCurrent(gun, badRoot)
	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidMetadata{}, err)
}

// We cannot validate a new root if we cannot get the old root from the DB (
// and cannot detect whether there was an old root or not), because there might
// have been an old root and we can't determine if the new root represents a
// root key rotation.
func TestValidateRootGetCurrentRootBroken(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := getFailStore{
		MetaStore:    storage.NewMemStorage(),
		errsToReturn: map[string]error{data.CanonicalRootRole.String(): data.ErrNoSuchRole{}},
	}

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, data.ErrNoSuchRole{}, err)
}

// A valid root rotation only cares about the immediately previous old root keys,
// whether or not there are old root roles
func TestValidateRootRotationWithOldSigs(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, crypto, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	serverCrypto := testutils.CopyKeys(t, crypto, data.CanonicalTimestampRole)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	// set the original root in the store
	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}
	require.NoError(t, store.UpdateMany(gun, updates))

	// rotate the root key, sign with both keys, and update - update should succeed
	newRootKey, err := testutils.CreateKey(crypto, gun, data.CanonicalRootRole, data.ECDSAKey)
	require.NoError(t, err)
	newRootID := newRootKey.ID()

	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, newRootKey))
	r, _, sn, _, err = testutils.Sign(repo)
	require.NoError(t, err)
	root, _, snapshot, _, err = getUpdates(r, tg, sn, ts)
	require.NoError(t, err)
	root.Version = repo.Root.Signed.Version
	snapshot.Version = repo.Snapshot.Signed.Version

	updates, err = validateUpdate(serverCrypto, gun, []storage.MetaUpdate{root, snapshot}, store)
	require.NoError(t, err)
	require.NoError(t, store.UpdateMany(gun, updates))

	// the next root does NOT need to be signed by both keys, because we only care
	// about signing with both keys if the root keys have changed (signRoot again to bump the version)

	r, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	// delete all signatures except the one with the new key
	for _, sig := range repo.Root.Signatures {
		if sig.KeyID == newRootID {
			r.Signatures = []data.Signature{sig}
			repo.Root.Signatures = r.Signatures
			break
		}
	}
	sn, err = repo.SignSnapshot(data.DefaultExpires(data.CanonicalSnapshotRole))
	require.NoError(t, err)

	root, _, snapshot, _, err = getUpdates(r, tg, sn, ts)
	require.NoError(t, err)
	root.Version = repo.Root.Signed.Version
	snapshot.Version = repo.Snapshot.Signed.Version
	updates, err = validateUpdate(serverCrypto, gun, []storage.MetaUpdate{root, snapshot}, store)
	require.NoError(t, err)
	require.NoError(t, store.UpdateMany(gun, updates))

	// another root rotation requires only the previous and new keys, and not the
	// original root key even though that original role is still in the metadata

	newRootKey2, err := testutils.CreateKey(crypto, gun, data.CanonicalRootRole, data.ECDSAKey)
	require.NoError(t, err)
	newRootID2 := newRootKey2.ID()

	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, newRootKey2))
	r, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	// delete all signatures except the ones with the first and second new keys
	sigs := make([]data.Signature, 0, 2)
	for _, sig := range repo.Root.Signatures {
		if sig.KeyID == newRootID || sig.KeyID == newRootID2 {
			sigs = append(sigs, sig)
		}
	}
	require.Len(t, sigs, 2)
	repo.Root.Signatures = sigs
	r.Signatures = sigs

	sn, err = repo.SignSnapshot(data.DefaultExpires(data.CanonicalSnapshotRole))
	require.NoError(t, err)

	root, _, snapshot, _, err = getUpdates(r, tg, sn, ts)
	require.NoError(t, err)
	root.Version = repo.Root.Signed.Version
	snapshot.Version = repo.Snapshot.Signed.Version
	_, err = validateUpdate(serverCrypto, gun, []storage.MetaUpdate{root, snapshot}, store)
	require.NoError(t, err)
}

// A valid root rotation requires the immediately previous root ROLE be satisfied,
// not just that there is a single root signature.  So if there were 2 keys, either
// of which can sign the root rotation, then either one of those keys can be used
// to sign the root rotation - not necessarily the one that signed the previous root.
func TestValidateRootRotationMultipleKeysThreshold1(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, crypto, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	serverCrypto := testutils.CopyKeys(t, crypto, data.CanonicalTimestampRole)
	store := storage.NewMemStorage()

	// add a new root key to the root so that either can sign have to sign
	additionalRootKey, err := testutils.CreateKey(crypto, gun, data.CanonicalRootRole, data.ECDSAKey)
	require.NoError(t, err)
	additionalRootID := additionalRootKey.ID()
	repo.Root.Signed.Keys[additionalRootID] = additionalRootKey
	repo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs = append(
		repo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs, additionalRootID)

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	require.Len(t, r.Signatures, 2)
	// make sure the old root was just signed with the first key
	for _, sig := range r.Signatures {
		if sig.KeyID != additionalRootID {
			r.Signatures = []data.Signature{sig}
			break
		}
	}
	require.Len(t, r.Signatures, 1)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	// set the original root in the store
	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}
	require.NoError(t, store.UpdateMany(gun, updates))

	// replace the keys with just 1 key
	rotatedRootKey, err := testutils.CreateKey(crypto, gun, data.CanonicalRootRole, data.ECDSAKey)
	require.NoError(t, err)
	rotatedRootID := rotatedRootKey.ID()
	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, rotatedRootKey))

	r, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	require.Len(t, r.Signatures, 3)
	// delete all signatures except the additional key (which didn't sign the
	// previous root) and the new key
	sigs := make([]data.Signature, 0, 2)
	for _, sig := range repo.Root.Signatures {
		if sig.KeyID == additionalRootID || sig.KeyID == rotatedRootID {
			sigs = append(sigs, sig)
		}
	}
	require.Len(t, sigs, 2)
	repo.Root.Signatures = sigs
	r.Signatures = sigs

	sn, err = repo.SignSnapshot(data.DefaultExpires(data.CanonicalSnapshotRole))
	require.NoError(t, err)

	root, _, snapshot, _, err = getUpdates(r, tg, sn, ts)
	require.NoError(t, err)
	root.Version = repo.Root.Signed.Version
	snapshot.Version = repo.Snapshot.Signed.Version
	_, err = validateUpdate(serverCrypto, gun, []storage.MetaUpdate{root, snapshot}, store)
	require.NoError(t, err)
}

// A root rotation must be signed with old and new root keys such that it satisfies
// the old and new roles, otherwise the new root fails to validate
func TestRootRotationNotSignedWithOldKeysForOldRole(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, crypto, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()
	serverCrypto := testutils.CopyKeys(t, crypto, data.CanonicalTimestampRole)

	oldRootKeyID := repo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs[0]

	// make the original root have 2 keys with a threshold of 2
	pairedRootKeys := make([]data.PublicKey, 2)
	for i := 0; i < len(pairedRootKeys); i++ {
		pairedRootKeys[i], err = testutils.CreateKey(crypto, gun, data.CanonicalRootRole, data.ECDSAKey)
		require.NoError(t, err)
	}
	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, pairedRootKeys...))
	repo.Root.Signed.Roles[data.CanonicalRootRole].Threshold = 2

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	require.Len(t, r.Signatures, 3)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}
	require.NoError(t, store.UpdateMany(gun, updates))

	finalRootKey, err := testutils.CreateKey(crypto, gun, data.CanonicalRootRole, data.ECDSAKey)
	require.NoError(t, err)
	repo.Root.Signed.Roles[data.CanonicalRootRole].Threshold = 1
	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, finalRootKey))

	r, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	origSigs := r.Signatures

	// make sure it's signed with only one of the previous keys and the new key
	sigs := make([]data.Signature, 0, 2)
	for _, sig := range origSigs {
		if sig.KeyID == pairedRootKeys[0].ID() || sig.KeyID == finalRootKey.ID() {
			sigs = append(sigs, sig)
		}
	}
	require.Len(t, sigs, 2)
	repo.Root.Signatures = sigs
	r.Signatures = sigs

	sn, err = repo.SignSnapshot(data.DefaultExpires(data.CanonicalSnapshotRole))
	require.NoError(t, err)
	root, _, snapshot, _, err = getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	_, err = validateUpdate(serverCrypto, gun, []storage.MetaUpdate{root, snapshot}, store)
	require.Error(t, err)
	require.Contains(t, err.Error(), "could not rotate trust to a new trusted root")

	// now sign with both of the pair and the new one
	sigs = make([]data.Signature, 0, 3)
	for _, sig := range origSigs {
		if sig.KeyID != oldRootKeyID {
			sigs = append(sigs, sig)
		}
	}
	require.Len(t, sigs, 3)
	repo.Root.Signatures = sigs
	r.Signatures = sigs

	sn, err = repo.SignSnapshot(data.DefaultExpires(data.CanonicalSnapshotRole))
	require.NoError(t, err)
	root, _, snapshot, _, err = getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	_, err = validateUpdate(serverCrypto, gun, []storage.MetaUpdate{root, snapshot}, store)
	require.NoError(t, err)
}

// A root rotation must increment the root version by 1
func TestRootRotationVersionIncrement(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, crypto, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	serverCrypto := testutils.CopyKeys(t, crypto, data.CanonicalTimestampRole)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	// set the original root in the store
	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}
	require.NoError(t, store.UpdateMany(gun, updates))

	// rotate the root key, sign with both keys, and update - update should succeed
	newRootKey, err := testutils.CreateKey(crypto, gun, data.CanonicalRootRole, data.ECDSAKey)
	require.NoError(t, err)

	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, newRootKey))
	r, _, sn, _, err = testutils.Sign(repo)
	require.NoError(t, err)
	root, _, snapshot, _, err = getUpdates(r, tg, sn, ts)
	require.NoError(t, err)
	snapshot.Version = repo.Snapshot.Signed.Version

	// Wrong root version
	root.Version = repo.Root.Signed.Version + 1

	_, err = validateUpdate(serverCrypto, gun, []storage.MetaUpdate{root, snapshot}, store)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Root modifications must increment the version")

	// correct root version
	root.Version = root.Version - 1
	updates, err = validateUpdate(serverCrypto, gun, []storage.MetaUpdate{root, snapshot}, store)
	require.NoError(t, err)
	require.NoError(t, store.UpdateMany(gun, updates))
}

// An update is not valid without the root metadata.
func TestValidateNoRoot(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	_, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrValidation{}, err)
}

func TestValidateSnapshotMissingNoSnapshotKey(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, _, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadHierarchy{}, err)
}

func TestValidateSnapshotGenerateNoPrev(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, _, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole, data.CanonicalSnapshotRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)
}

func TestValidateSnapshotGenerateWithPrev(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets}

	// set the current snapshot in the store manually so we find it when generating
	// the next version
	store.UpdateCurrent(gun, snapshot)

	prev, err := data.SnapshotFromSigned(sn)
	require.NoError(t, err)

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole, data.CanonicalSnapshotRole)
	updates, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)

	for _, u := range updates {
		if u.Role == data.CanonicalSnapshotRole {
			curr := &data.SignedSnapshot{}
			err = json.Unmarshal(u.Data, curr)
			require.NoError(t, err)
			require.Equal(t, prev.Signed.Version+1, curr.Signed.Version)
			require.Equal(t, u.Version, curr.Signed.Version)
		}
	}
}

func TestValidateSnapshotGeneratePrevCorrupt(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets}

	// corrupt the JSON structure of prev snapshot
	snapshot.Data = snapshot.Data[1:]
	// set the current snapshot in the store manually so we find it when generating
	// the next version
	store.UpdateCurrent(gun, snapshot)

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole, data.CanonicalSnapshotRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, &json.SyntaxError{}, err)
}

// Store is broken when getting the current snapshot
func TestValidateSnapshotGenerateStoreGetCurrentSnapshotBroken(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := getFailStore{
		MetaStore:    storage.NewMemStorage(),
		errsToReturn: map[string]error{data.CanonicalSnapshotRole.String(): data.ErrNoSuchRole{}},
	}

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, _, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole, data.CanonicalSnapshotRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, data.ErrNoSuchRole{}, err)
}

func TestValidateSnapshotGenerateNoTargets(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, _, _, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole, data.CanonicalSnapshotRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
}

func TestValidateSnapshotGenerate(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, _, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{targets}

	store.UpdateCurrent(gun, root)

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole, data.CanonicalSnapshotRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.NoError(t, err)
}

// If there is no timestamp key in the store, validation fails.  This could
// happen if pushing an existing repository from one server to another that
// does not have the repo.
func TestValidateRootNoTimestampKey(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	oldRepo, _, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	r, tg, sn, ts, err := testutils.Sign(oldRepo)
	require.NoError(t, err)
	root, targets, snapshot, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	store := storage.NewMemStorage()
	updates := []storage.MetaUpdate{root, targets, snapshot}

	// do not copy the targets key to the storage, and try to update the root
	serverCrypto := signed.NewEd25519()
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)

	// there should still be no timestamp keys - one should not have been
	// created
	require.Empty(t, serverCrypto.ListAllKeys())
}

// If the timestamp key in the store does not match the timestamp key in
// the root.json, validation fails.  This could happen if pushing an existing
// repository from one server to another that had already initialized the same
// repo.
func TestValidateRootInvalidTimestampKey(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	oldRepo, _, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	r, tg, sn, ts, err := testutils.Sign(oldRepo)
	require.NoError(t, err)
	root, targets, snapshot, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	store := storage.NewMemStorage()
	updates := []storage.MetaUpdate{root, targets, snapshot}

	serverCrypto := signed.NewEd25519()
	_, err = serverCrypto.Create(data.CanonicalTimestampRole, gun, data.ED25519Key)
	require.NoError(t, err)

	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

// If the timestamp role has a threshold > 1, validation fails.
func TestValidateRootInvalidTimestampThreshold(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	oldRepo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	tsKey2, err := testutils.CreateKey(cs, gun, "timestamp2", data.ECDSAKey)
	require.NoError(t, err)
	oldRepo.AddBaseKeys(data.CanonicalTimestampRole, tsKey2)
	tsRole, ok := oldRepo.Root.Signed.Roles[data.CanonicalTimestampRole]
	require.True(t, ok)
	tsRole.Threshold = 2

	r, tg, sn, ts, err := testutils.Sign(oldRepo)
	require.NoError(t, err)
	root, targets, snapshot, _, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	store := storage.NewMemStorage()
	updates := []storage.MetaUpdate{root, targets, snapshot}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

// If any role has a threshold < 1, validation fails
func TestValidateRootInvalidZeroThreshold(t *testing.T) {
	for _, role := range data.BaseRoles {
		var gun data.GUN = "docker.com/notary"
		oldRepo, cs, err := testutils.EmptyRepo(gun)
		require.NoError(t, err)
		tsRole, ok := oldRepo.Root.Signed.Roles[role]
		require.True(t, ok)
		tsRole.Threshold = 0

		r, tg, sn, ts, err := testutils.Sign(oldRepo)
		require.NoError(t, err)
		root, targets, snapshot, _, err := getUpdates(r, tg, sn, ts)
		require.NoError(t, err)

		store := storage.NewMemStorage()
		updates := []storage.MetaUpdate{root, targets, snapshot}

		serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
		_, err = validateUpdate(serverCrypto, gun, updates, store)
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid threshold")
	}
}

// ### Role missing negative tests ###
// These tests remove a role from the Root file and
// check for a validation.ErrBadRoot
func TestValidateRootRoleMissing(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	delete(repo.Root.Signed.Roles, data.CanonicalRootRole)

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

func TestValidateTargetsRoleMissing(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	delete(repo.Root.Signed.Roles, "targets")

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

func TestValidateSnapshotRoleMissing(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	delete(repo.Root.Signed.Roles, "snapshot")

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

// ### End role missing negative tests ###

// ### Signature missing negative tests ###
func TestValidateRootSigMissing(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	delete(repo.Root.Signed.Roles, "snapshot")

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)

	r.Signatures = nil

	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

func TestValidateTargetsSigMissing(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)

	tg.Signatures = nil

	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadTargets{}, err)
}

func TestValidateSnapshotSigMissing(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)

	sn.Signatures = nil

	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadSnapshot{}, err)
}

// ### End signature missing negative tests ###

// ### Corrupted metadata negative tests ###
func TestValidateRootCorrupt(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	// flip all the bits in the first byte
	root.Data[0] = root.Data[0] ^ 0xff

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

func TestValidateTargetsCorrupt(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	// flip all the bits in the first byte
	targets.Data[0] = targets.Data[0] ^ 0xff

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadTargets{}, err)
}

func TestValidateSnapshotCorrupt(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)
	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	// flip all the bits in the first byte
	snapshot.Data[0] = snapshot.Data[0] ^ 0xff

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadSnapshot{}, err)
}

// ### End corrupted metadata negative tests ###

// ### Snapshot size mismatch negative tests ###
func TestValidateRootModifiedSize(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)

	// add another copy of the signature so the hash is different
	r.Signatures = append(r.Signatures, r.Signatures[0])

	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	// flip all the bits in the first byte
	root.Data[0] = root.Data[0] ^ 0xff

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadRoot{}, err)
}

func TestValidateTargetsModifiedSize(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)

	// add another copy of the signature so the hash is different
	tg.Signatures = append(tg.Signatures, tg.Signatures[0])

	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadSnapshot{}, err)
}

// ### End snapshot size mismatch negative tests ###

// ### Snapshot hash mismatch negative tests ###
func TestValidateRootModifiedHash(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)

	snap, err := data.SnapshotFromSigned(sn)
	require.NoError(t, err)
	snap.Signed.Meta[data.CanonicalRootRole.String()].Hashes["sha256"][0] = snap.Signed.Meta[data.CanonicalRootRole.String()].Hashes["sha256"][0] ^ 0xff

	sn, err = snap.ToSigned()
	require.NoError(t, err)

	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadSnapshot{}, err)
}

func TestValidateTargetsModifiedHash(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	r, tg, sn, ts, err := testutils.Sign(repo)
	require.NoError(t, err)

	snap, err := data.SnapshotFromSigned(sn)
	require.NoError(t, err)
	snap.Signed.Meta["targets"].Hashes["sha256"][0] = snap.Signed.Meta["targets"].Hashes["sha256"][0] ^ 0xff

	sn, err = snap.ToSigned()
	require.NoError(t, err)

	root, targets, snapshot, timestamp, err := getUpdates(r, tg, sn, ts)
	require.NoError(t, err)

	updates := []storage.MetaUpdate{root, targets, snapshot, timestamp}

	serverCrypto := testutils.CopyKeys(t, cs, data.CanonicalTimestampRole)
	_, err = validateUpdate(serverCrypto, gun, updates, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadSnapshot{}, err)
}

// ### End snapshot hash mismatch negative tests ###

// ### generateSnapshot tests ###
func TestGenerateSnapshotRootNotLoaded(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	_, err := generateSnapshot(gun, builder, storage.NewMemStorage())
	require.Error(t, err)
	require.IsType(t, validation.ErrValidation{}, err)
}

func TestGenerateSnapshotNoKey(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	metadata, cs, err := testutils.NewRepoMetadata(gun)
	require.NoError(t, err)
	store := storage.NewMemStorage()

	// delete snapshot key in the cryptoservice
	for _, keyID := range cs.ListKeys(data.CanonicalSnapshotRole) {
		require.NoError(t, cs.RemoveKey(keyID))
	}

	builder := tuf.NewRepoBuilder(gun, cs, trustpinning.TrustPinConfig{})
	// only load root and targets
	require.NoError(t, builder.Load(data.CanonicalRootRole, metadata[data.CanonicalRootRole], 0, false))
	require.NoError(t, builder.Load(data.CanonicalTargetsRole, metadata[data.CanonicalTargetsRole], 0, false))

	_, err = generateSnapshot(gun, builder, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadHierarchy{}, err)
}

// ### End generateSnapshot tests ###

// ### Target validation with delegations tests
func TestLoadTargetsLoadsNothingIfNoUpdates(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	metadata, _, err := testutils.NewRepoMetadata(gun)
	require.NoError(t, err)

	// load the root into the builder, else we can't load anything else
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	require.NoError(t, builder.Load(data.CanonicalRootRole, metadata[data.CanonicalRootRole], 0, false))

	store := storage.NewMemStorage()
	store.UpdateCurrent(gun, storage.MetaUpdate{
		Role:    data.CanonicalTargetsRole,
		Version: 1,
		Data:    metadata[data.CanonicalTargetsRole],
	})

	// if no updates, nothing is loaded
	targetsToUpdate, err := loadAndValidateTargets(gun, builder, nil, store)
	require.Empty(t, targetsToUpdate)
	require.NoError(t, err)
	require.False(t, builder.IsLoaded(data.CanonicalTargetsRole))
}

// When a delegation role appears in the update and the parent does not, the
// parent is loaded from the DB if it can
func TestValidateTargetsRequiresStoredParent(t *testing.T) {
	var (
		gun      data.GUN      = "docker.com/notary"
		delgName data.RoleName = "targets/level1"
	)
	metadata, _, err := testutils.NewRepoMetadata(gun, delgName, data.RoleName(path.Join(delgName.String(), "other")))
	require.NoError(t, err)

	// load the root into the builder, else we can't load anything else
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	require.NoError(t, builder.Load(data.CanonicalRootRole, metadata[data.CanonicalRootRole], 0, false))

	delUpdate := storage.MetaUpdate{
		Role:    delgName,
		Version: 1,
		Data:    metadata[delgName],
	}

	upload := map[data.RoleName]storage.MetaUpdate{delgName: delUpdate}

	store := storage.NewMemStorage()

	// if the DB has no "targets" role
	_, err = loadAndValidateTargets(gun, builder, upload, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadTargets{}, err)

	// ensure the "targets" (the parent) is in the "db"
	store.UpdateCurrent(gun, storage.MetaUpdate{
		Role:    data.CanonicalTargetsRole,
		Version: 1,
		Data:    metadata[data.CanonicalTargetsRole],
	})

	updates, err := loadAndValidateTargets(gun, builder, upload, store)
	require.NoError(t, err)
	require.Len(t, updates, 1)
	require.Equal(t, delgName, updates[0].Role)
	require.Equal(t, metadata[delgName], updates[0].Data)
}

// If the parent is not in the store, then the parent must be in the update else
// validation fails.
func TestValidateTargetsParentInUpdate(t *testing.T) {
	var (
		gun      data.GUN      = "docker.com/notary"
		delgName data.RoleName = "targets/level1"
	)
	metadata, _, err := testutils.NewRepoMetadata(gun, delgName, data.RoleName(path.Join(delgName.String(), "other")))
	require.NoError(t, err)
	store := storage.NewMemStorage()

	// load the root into the builder, else we can't load anything else
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	require.NoError(t, builder.Load(data.CanonicalRootRole, metadata[data.CanonicalRootRole], 0, false))

	targetsUpdate := storage.MetaUpdate{
		Role:    data.CanonicalTargetsRole,
		Version: 1,
		Data:    []byte("Invalid metadata"),
	}

	delgUpdate := storage.MetaUpdate{
		Role:    delgName,
		Version: 1,
		Data:    metadata[delgName],
	}

	upload := map[data.RoleName]storage.MetaUpdate{
		"targets/level1":          delgUpdate,
		data.CanonicalTargetsRole: targetsUpdate,
	}

	// parent update not readable - fail
	_, err = loadAndValidateTargets(gun, builder, upload, store)
	require.Error(t, err)
	require.IsType(t, validation.ErrBadTargets{}, err)

	// because we sort the roles, the list of returned updates
	// will contain shallower roles first, in this case "targets",
	// and then "targets/level1"
	targetsUpdate.Data = metadata[data.CanonicalTargetsRole]
	upload[data.CanonicalTargetsRole] = targetsUpdate
	updates, err := loadAndValidateTargets(gun, builder, upload, store)
	require.NoError(t, err)
	require.Equal(t, []storage.MetaUpdate{targetsUpdate, delgUpdate}, updates)
}

// If the parent, either from the DB or from an update, does not contain the role
// of the delegation update, validation fails
func TestValidateTargetsRoleNotInParent(t *testing.T) {
	// no delegation at first
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)

	// load the root into the builder, else we can't load anything else
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 0, false))

	// prepare the original targets file, without a delegation role, as an update
	origTargetsUpdate := storage.MetaUpdate{
		Role:    data.CanonicalTargetsRole,
		Version: 1,
		Data:    meta[data.CanonicalTargetsRole],
	}
	emptyStore := storage.NewMemStorage()
	storeWithParent := storage.NewMemStorage()
	storeWithParent.UpdateCurrent(gun, origTargetsUpdate)

	// add a delegation role now
	var delgName data.RoleName = "targets/level1"
	level1Key, err := testutils.CreateKey(cs, gun, delgName, data.ECDSAKey)
	require.NoError(t, err)
	require.NoError(t, repo.UpdateDelegationKeys(delgName, []data.PublicKey{level1Key}, []string{}, 1))
	// create the delegation metadata too
	repo.InitTargets(delgName)

	// re-serialize
	meta, err = testutils.SignAndSerialize(repo)
	require.NoError(t, err)
	delgMeta, ok := meta[delgName]
	require.True(t, ok)

	delgUpdate := storage.MetaUpdate{
		Role:    delgName,
		Version: 1,
		Data:    delgMeta,
	}

	// parent in update does not have this role, whether or not there's a parent in the store,
	// so validation fails
	roles := map[data.RoleName]storage.MetaUpdate{
		delgName:                  delgUpdate,
		data.CanonicalTargetsRole: origTargetsUpdate,
	}
	for _, metaStore := range []storage.MetaStore{emptyStore, storeWithParent} {
		updates, err := loadAndValidateTargets(gun, builder, roles, metaStore)
		require.Error(t, err)
		require.Empty(t, updates)
		require.IsType(t, validation.ErrBadTargets{}, err)
	}

	// if the update is provided without the parent, then the parent from the
	// store is loaded - if it doesn't have the role, then the update fails
	updates, err := loadAndValidateTargets(gun, builder,
		map[data.RoleName]storage.MetaUpdate{delgName: delgUpdate}, storeWithParent)
	require.Error(t, err)
	require.Empty(t, updates)
	require.IsType(t, validation.ErrBadTargets{}, err)
}

// ### End target validation with delegations tests
