// make sure that the swizzler actually sort of works, so our tests that use it actually test what we
// think

package testutils

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	store "github.com/docker/notary/storage"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/stretchr/testify/require"
)

// creates a new swizzler with 3 delegation targets (and only 2 metadata files
// for those targets), and returns the swizzler along with a copy of the original
// metadata
func createNewSwizzler(t *testing.T) (*MetadataSwizzler, map[data.RoleName][]byte) {
	var gun data.GUN = "docker.com/notary"
	m, cs, err := NewRepoMetadata(gun, "targets/a", "targets/a/b", "targets/a/b/c")
	require.NoError(t, err)

	return NewMetadataSwizzler(gun, m, cs), CopyRepoMetadata(m)
}

// A new swizzler should have metadata for all roles, and a snapshot of all roles
func TestNewSwizzler(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	for _, role := range f.Roles {
		metaBytes, ok := origMeta[role]
		require.True(t, ok)
		require.NotNil(t, metaBytes)
		require.NotEmpty(t, metaBytes)
	}

	snapshot, timestamp := &data.SignedSnapshot{}, &data.SignedTimestamp{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalSnapshotRole], snapshot))
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalTimestampRole], timestamp))

	for _, role := range f.Roles {
		filemeta, ok := snapshot.Signed.Meta[role.String()]
		if role != data.CanonicalTimestampRole && role != data.CanonicalSnapshotRole {
			require.True(t, ok)
			require.NotNil(t, filemeta)
			require.NotEmpty(t, filemeta)
		} else {
			require.False(t, ok)
		}
	}

	require.Len(t, timestamp.Signed.Meta, 1)
	filemeta, ok := timestamp.Signed.Meta[data.CanonicalSnapshotRole.String()]
	require.True(t, ok)
	require.NotNil(t, filemeta)
	require.NotEmpty(t, filemeta)

	// targets should have 1 delegated role, as should targets/a
	targets, targetsA := &data.SignedTargets{}, &data.SignedTargets{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalTargetsRole], targets))
	require.NoError(t, json.Unmarshal(origMeta["targets/a"], targetsA))

	require.Len(t, targets.Signed.Delegations.Roles, 1)
	require.EqualValues(t, "targets/a", targets.Signed.Delegations.Roles[0].Name)

	require.Len(t, targetsA.Signed.Delegations.Roles, 1)
	require.EqualValues(t, "targets/a/b", targetsA.Signed.Delegations.Roles[0].Name)
}

// This invalidates the metadata so that it can no longer be unmarshalled as
// JSON as any sort
func TestSwizzlerSetInvalidJSON(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.SetInvalidJSON(data.CanonicalSnapshotRole)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalSnapshotRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			// it should not be JSON unmarshallable
			var generic interface{}
			require.Error(t, json.Unmarshal(newMeta, &generic))
		}
	}
}

// This adds a single byte of whitespace to the metadata file, so it should be parsed
// and deserialized the same way, but checksums against snapshot/timestamp may fail
func TestSwizzlerAddExtraSpace(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.AddExtraSpace(data.CanonicalTargetsRole)

	snapshot := &data.SignedSnapshot{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalSnapshotRole], snapshot))

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalTargetsRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			require.True(t, bytes.Equal(metaBytes, newMeta[1:len(metaBytes)+1]))
			require.Equal(t, byte(' '), newMeta[0])
			require.Equal(t, byte(' '), newMeta[len(newMeta)-1])

			// make sure the hash is not the same as the hash in snapshot
			newHash := sha256.Sum256(newMeta)
			require.False(t, bytes.Equal(
				snapshot.Signed.Meta[data.CanonicalTargetsRole.String()].Hashes["sha256"],
				newHash[:]))
			require.NotEqual(t,
				snapshot.Signed.Meta[data.CanonicalTargetsRole.String()].Length,
				len(newMeta))
		}
	}
}

// This modifies metdata so that it is unmarshallable as JSON, but cannot be
// unmarshalled as a Signed object
func TestSwizzlerSetInvalidSigned(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.SetInvalidSigned(data.CanonicalTargetsRole)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalTargetsRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			// it be JSON unmarshallable, but not data.Signed marshallable
			var generic interface{}
			require.NoError(t, json.Unmarshal(newMeta, &generic))
			signedThing := data.Signed{}
			require.Error(t, json.Unmarshal(newMeta, &signedThing))
		}
	}
}

// This modifies metdata so that it is unmarshallable as JSON, but cannot be
// unmarshalled as a Signed object
func TestSwizzlerSetInvalidSignedMeta(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	err := f.SetInvalidSignedMeta(data.CanonicalRootRole)
	require.NoError(t, err)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalRootRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			// can be unmarshaled as Signed, but not as SignedMeta
			signedThing := data.Signed{}
			require.NoError(t, json.Unmarshal(newMeta, &signedThing))
			signedMeta := data.SignedMeta{}
			require.Error(t, json.Unmarshal(newMeta, &signedMeta))
		}
	}
}

// This modifies metdata so that it is unmarshallable as JSON, but cannot be
// unmarshalled as a Signed object
func TestSwizzlerSetInvalidMetadataType(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.SetInvalidMetadataType(data.CanonicalTargetsRole)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalTargetsRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			signedMeta := data.SignedMeta{}
			require.NoError(t, json.Unmarshal(newMeta, &signedMeta))
			require.NotEqual(t, data.CanonicalTargetsRole, signedMeta.Signed.Type)
		}
	}
}

// This modifies the metadata so that the signed part has an extra, extraneous
// field.  This does not prevent it from being unmarshalled as Signed* object,
// but the signature is no longer valid because the hash is different.
func TestSwizzlerInvalidateMetadataSignatures(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.InvalidateMetadataSignatures(data.CanonicalRootRole)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalRootRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			// it be JSON unmarshallable into a data.Signed, and it's signed by
			// root, but it is NOT the correct signature because the hash
			// does not match
			origSigned, newSigned := &data.Signed{}, &data.Signed{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.Len(t, newSigned.Signatures, len(origSigned.Signatures))
			for i := range origSigned.Signatures {
				require.Equal(t, origSigned.Signatures[i].KeyID, newSigned.Signatures[i].KeyID)
				require.Equal(t, origSigned.Signatures[i].Method, newSigned.Signatures[i].Method)
				require.NotEqual(t, origSigned.Signatures[i].Signature, newSigned.Signatures[i].Signature)
				require.Equal(t, []byte("invalid signature"), newSigned.Signatures[i].Signature)
			}
			require.True(t, bytes.Equal(*origSigned.Signed, *newSigned.Signed))
		}
	}
}

// This just deletes the metadata entirely from the cache
func TestSwizzlerRemoveMetadata(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.RemoveMetadata("targets/a")

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		if role != "targets/a" {
			require.NoError(t, err)
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.Error(t, err)
			require.IsType(t, store.ErrMetaNotFound{}, err)
		}
	}
}

// This signs the metadata with the wrong key
func TestSwizzlerSignMetadataWithInvalidKey(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.SignMetadataWithInvalidKey(data.CanonicalTimestampRole)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalTimestampRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			// it is JSON unmarshallable as a timestamp, but the signature ID
			// does not match.
			require.NoError(t, json.Unmarshal(newMeta, &data.SignedTimestamp{}))
			origSigned, newSigned := &data.Signed{}, &data.Signed{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.Len(t, origSigned.Signatures, 1)
			require.Len(t, newSigned.Signatures, 1)
			require.NotEqual(t, origSigned.Signatures[0].KeyID, newSigned.Signatures[0].KeyID)
		}
	}
}

// This updates the metadata version with a particular number
func TestSwizzlerOffsetMetadataVersion(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.OffsetMetadataVersion("targets/a", -2)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != "targets/a" {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			origSigned, newSigned := &data.SignedMeta{}, &data.SignedMeta{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.Equal(t, 1, origSigned.Signed.Version)
			require.Equal(t, -1, newSigned.Signed.Version)
		}
	}
}

// This causes the metadata to be expired
func TestSwizzlerExpireMetadata(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	err := f.ExpireMetadata(data.CanonicalRootRole)
	require.NoError(t, err)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalRootRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			origSigned, newSigned := &data.SignedMeta{}, &data.SignedMeta{}
			now := time.Now()
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.True(t, now.Before(origSigned.Signed.Expires))
			require.True(t, now.After(newSigned.Signed.Expires))
		}
	}
}

// This sets the threshold for a base role
func TestSwizzlerSetThresholdBaseRole(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	err := f.SetThreshold(data.CanonicalTargetsRole, 3)
	require.NoError(t, err)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		// the threshold for base roles is set in root
		if role != data.CanonicalRootRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			signedRoot := &data.SignedRoot{}
			require.NoError(t, json.Unmarshal(newMeta, signedRoot))
			for r, roleInfo := range signedRoot.Signed.Roles {
				if r != data.CanonicalTargetsRole {
					require.Equal(t, 1, roleInfo.Threshold)
				} else {
					require.Equal(t, 3, roleInfo.Threshold)
				}
			}
		}
	}
}

// This sets the threshold for a delegation
func TestSwizzlerSetThresholdDelegatedRole(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	f.SetThreshold("targets/a/b", 3)

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		// the threshold for "targets/a/b" is in "targets/a"
		if role != "targets/a" {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			signedTargets := &data.SignedTargets{}
			require.NoError(t, json.Unmarshal(newMeta, signedTargets))
			require.Len(t, signedTargets.Signed.Delegations.Roles, 1)
			require.EqualValues(t, "targets/a/b", signedTargets.Signed.Delegations.Roles[0].Name)
			require.Equal(t, 3, signedTargets.Signed.Delegations.Roles[0].Threshold)
		}
	}
}

// This changes the root key
func TestSwizzlerChangeRootKey(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	err := f.ChangeRootKey()
	require.NoError(t, err)

	// we want to test these in a specific order
	roles := []data.RoleName{data.CanonicalRootRole, data.CanonicalTargetsRole, data.CanonicalSnapshotRole,
		data.CanonicalTimestampRole, "targets/a", "targets/a/b"}

	for _, role := range roles {
		origMeta := origMeta[role]
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		// the threshold for base roles is set in root
		switch role {
		case data.CanonicalRootRole:
			require.False(t, bytes.Equal(origMeta, newMeta))
			origRoot, newRoot := &data.SignedRoot{}, &data.SignedRoot{}
			require.NoError(t, json.Unmarshal(origMeta, origRoot))
			require.NoError(t, json.Unmarshal(newMeta, newRoot))

			require.NotEqual(t, len(origRoot.Signed.Keys), len(newRoot.Signed.Keys))

			for r, origRole := range origRoot.Signed.Roles {
				newRole := newRoot.Signed.Roles[r]
				require.Len(t, origRole.KeyIDs, 1)
				require.Len(t, newRole.KeyIDs, 1)
				if r == data.CanonicalRootRole {
					require.NotEqual(t, origRole.KeyIDs[0], newRole.KeyIDs[0])
				} else {
					require.Equal(t, origRole.KeyIDs[0], newRole.KeyIDs[0])
				}
			}

			rootRole, err := newRoot.BuildBaseRole(data.CanonicalRootRole)
			require.NoError(t, err)
			signedThing, err := newRoot.ToSigned()
			require.NoError(t, err)
			require.NoError(t, signed.VerifySignatures(signedThing, rootRole))
			require.NoError(t, signed.VerifyVersion(&(newRoot.Signed.SignedCommon), 1))
		default:
			require.True(t, bytes.Equal(origMeta, newMeta), "bytes have changed for role %s", role)
		}
	}
}

// UpdateSnapshotHashes will recreate all snapshot hashes, useful if some metadata
// has been fuzzed and we want all the hashes to be correct.  If roles are provided,
// only hashes for those roles will be re-generated.
func TestSwizzlerUpdateSnapshotHashesSpecifiedRoles(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	// nothing has changed, signed data should be the same (signatures might
	// change because signatures may have random elements
	f.UpdateSnapshotHashes(data.CanonicalTargetsRole)
	newMeta, err := f.MetadataCache.GetSized(data.CanonicalSnapshotRole.String(), store.NoSizeLimit)
	require.NoError(t, err)

	origSigned, newSigned := &data.Signed{}, &data.Signed{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalSnapshotRole], origSigned))
	require.NoError(t, json.Unmarshal(newMeta, newSigned))
	require.True(t, bytes.Equal(*origSigned.Signed, *newSigned.Signed))

	// change these 3 metadata items
	f.InvalidateMetadataSignatures(data.CanonicalTargetsRole)
	f.InvalidateMetadataSignatures("targets/a")
	f.InvalidateMetadataSignatures("targets/a/b")
	// update the snapshot with just 1 role
	f.UpdateSnapshotHashes(data.CanonicalTargetsRole)

	newMeta, err = f.MetadataCache.GetSized(data.CanonicalSnapshotRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	require.False(t, bytes.Equal(origMeta[data.CanonicalSnapshotRole], newMeta))

	origSnapshot, newSnapshot := &data.SignedSnapshot{}, &data.SignedSnapshot{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalSnapshotRole], origSnapshot))
	require.NoError(t, json.Unmarshal(newMeta, newSnapshot))

	// only the targets checksum was regenerated, since that was specified
	for _, role := range f.Roles {
		switch role {
		case data.CanonicalTimestampRole:
			continue
		case data.CanonicalTargetsRole:
			require.NotEqual(t, origSnapshot.Signed.Meta[role.String()], newSnapshot.Signed.Meta[role.String()])
		default:
			require.Equal(t, origSnapshot.Signed.Meta[role.String()], newSnapshot.Signed.Meta[role.String()])
		}
	}
}

// UpdateSnapshotHashes will recreate all snapshot hashes, useful if some metadata
// has been fuzzed and we want all the hashes to be correct.  If no roles are provided,
// all hashes are regenerated
func TestSwizzlerUpdateSnapshotHashesNoSpecifiedRoles(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	// nothing has changed, signed data should be the same (signatures might
	// change because signatures may have random elements
	f.UpdateSnapshotHashes()
	newMeta, err := f.MetadataCache.GetSized(data.CanonicalSnapshotRole.String(), store.NoSizeLimit)
	require.NoError(t, err)

	origSigned, newSigned := &data.Signed{}, &data.Signed{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalSnapshotRole], origSigned))
	require.NoError(t, json.Unmarshal(newMeta, newSigned))
	require.True(t, bytes.Equal(*origSigned.Signed, *newSigned.Signed))

	// change these 2 metadata items
	f.InvalidateMetadataSignatures(data.CanonicalTargetsRole)
	f.InvalidateMetadataSignatures("targets/a")
	f.InvalidateMetadataSignatures("targets/a/b")
	// update the snapshot with just no specified roles
	f.UpdateSnapshotHashes()

	newMeta, err = f.MetadataCache.GetSized(data.CanonicalSnapshotRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	require.False(t, bytes.Equal(origMeta[data.CanonicalSnapshotRole], newMeta))

	origSnapshot, newSnapshot := &data.SignedSnapshot{}, &data.SignedSnapshot{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalSnapshotRole], origSnapshot))
	require.NoError(t, json.Unmarshal(newMeta, newSnapshot))

	for _, role := range f.Roles {
		switch role {
		case data.CanonicalTimestampRole:
			continue
		case data.CanonicalTargetsRole:
			fallthrough
		case "targets/a":
			fallthrough
		case "targets/a/b":
			require.NotEqual(t, origSnapshot.Signed.Meta[role.String()], newSnapshot.Signed.Meta[role.String()])
		default:
			require.Equal(t, origSnapshot.Signed.Meta[role.String()], newSnapshot.Signed.Meta[role.String()])
		}
	}
}

// UpdateTimestamp will re-calculate the snapshot hash
func TestSwizzlerUpdateTimestamp(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	// nothing has changed, signed data should be the same (signatures might
	// change because signatures may have random elements
	f.UpdateTimestampHash()
	newMeta, err := f.MetadataCache.GetSized(data.CanonicalTimestampRole.String(), store.NoSizeLimit)
	require.NoError(t, err)

	origSigned, newSigned := &data.Signed{}, &data.Signed{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalTimestampRole], origSigned))
	require.NoError(t, json.Unmarshal(newMeta, newSigned))
	require.True(t, bytes.Equal(*origSigned.Signed, *newSigned.Signed))

	// update snapshot
	f.OffsetMetadataVersion(data.CanonicalSnapshotRole, 1)
	// update the timestamp
	f.UpdateTimestampHash()

	newMeta, err = f.MetadataCache.GetSized(data.CanonicalTimestampRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	require.False(t, bytes.Equal(origMeta[data.CanonicalTimestampRole], newMeta))

	origTimestamp, newTimestamp := &data.SignedTimestamp{}, &data.SignedTimestamp{}
	require.NoError(t, json.Unmarshal(origMeta[data.CanonicalTimestampRole], origTimestamp))
	require.NoError(t, json.Unmarshal(newMeta, newTimestamp))

	require.Len(t, origTimestamp.Signed.Meta, 1)
	require.Len(t, newTimestamp.Signed.Meta, 1)
	require.False(t, reflect.DeepEqual(
		origTimestamp.Signed.Meta[data.CanonicalSnapshotRole.String()],
		newTimestamp.Signed.Meta[data.CanonicalSnapshotRole.String()]))
}

// functions which require re-signing the metadata will return ErrNoKeyForRole if
// the signing key is missing
func TestMissingSigningKey(t *testing.T) {
	f, _ := createNewSwizzler(t)

	// delete the snapshot, timestamp, and root keys
	noKeys := []data.RoleName{
		data.CanonicalSnapshotRole, data.CanonicalTimestampRole, data.CanonicalRootRole}
	for _, role := range noKeys {
		k := f.CryptoService.ListKeys(role)
		require.Len(t, k, 1)
		require.NoError(t, f.CryptoService.RemoveKey(k[0]))
	}

	// these are all the functions that require re-signing
	require.IsType(t, ErrNoKeyForRole{}, f.OffsetMetadataVersion(data.CanonicalSnapshotRole, 1))
	require.IsType(t, ErrNoKeyForRole{}, f.ExpireMetadata(data.CanonicalSnapshotRole))
	require.IsType(t, ErrNoKeyForRole{}, f.SetThreshold(data.CanonicalSnapshotRole, 2))
	require.IsType(t, ErrNoKeyForRole{}, f.UpdateSnapshotHashes())
	require.IsType(t, ErrNoKeyForRole{}, f.UpdateTimestampHash())
}

// This mutates the root
func TestSwizzlerMutateRoot(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	require.NoError(t, f.MutateRoot(func(r *data.Root) { r.Roles["hello"] = nil }))

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalRootRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			origSigned, newSigned := &data.SignedRoot{}, &data.SignedRoot{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			// it may not exactly equal 4 or 5 because if the metadata was
			// produced by calling SignedRoot, it could have saved a previous
			// root role
			require.True(t, len(origSigned.Signed.Roles) >= 4)
			require.True(t, len(newSigned.Signed.Roles) >= 5)
		}
	}
}

// This mutates the timestamp
func TestSwizzlerMutateTimestamp(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	require.NoError(t, f.MutateTimestamp(func(t *data.Timestamp) { t.Meta["hello"] = data.FileMeta{} }))

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalTimestampRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			origSigned, newSigned := &data.SignedTimestamp{}, &data.SignedTimestamp{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.Len(t, origSigned.Signed.Meta, 1)
			require.Len(t, newSigned.Signed.Meta, 2)
		}
	}
}

// This mutates the snapshot
func TestSwizzlerMutateSnapshot(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	require.NoError(t, f.MutateSnapshot(func(s *data.Snapshot) { s.Meta["hello"] = data.FileMeta{} }))

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalSnapshotRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			origSigned, newSigned := &data.SignedSnapshot{}, &data.SignedSnapshot{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.Len(t, origSigned.Signed.Meta, 4)
			require.Len(t, newSigned.Signed.Meta, 5)
		}
	}
}

// This mutates the targets
func TestSwizzlerMutateTargets(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	require.NoError(t, f.MutateTargets(func(t *data.Targets) { t.Targets["hello"] = data.FileMeta{} }))

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalTargetsRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			origSigned, newSigned := &data.SignedTargets{}, &data.SignedTargets{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.Len(t, origSigned.Signed.Targets, 0)
			require.Len(t, newSigned.Signed.Targets, 1)
		}
	}
}

// This rotates the key of some base role
func TestSwizzlerRotateKeyBaseRole(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	theRole := data.CanonicalSnapshotRole
	cs := signed.NewEd25519()
	pubKey, err := cs.Create(theRole, f.Gun, data.ED25519Key)
	require.NoError(t, err)

	require.NoError(t, f.RotateKey(theRole, pubKey))

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != data.CanonicalRootRole {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			origSigned, newSigned := &data.SignedRoot{}, &data.SignedRoot{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.NotEqual(t, []string{pubKey.ID()}, origSigned.Signed.Roles[theRole].KeyIDs)
			require.Equal(t, []string{pubKey.ID()}, newSigned.Signed.Roles[theRole].KeyIDs)
			_, ok := origSigned.Signed.Keys[pubKey.ID()]
			require.False(t, ok)
			_, ok = newSigned.Signed.Keys[pubKey.ID()]
			require.True(t, ok)
		}
	}
}

// This rotates the key of some delegation role
func TestSwizzlerRotateKeyDelegationRole(t *testing.T) {
	f, origMeta := createNewSwizzler(t)

	var theRole data.RoleName = "targets/a/b"
	cs := signed.NewEd25519()
	pubKey, err := cs.Create(theRole, f.Gun, data.ED25519Key)
	require.NoError(t, err)

	require.NoError(t, f.RotateKey(theRole, pubKey))

	for role, metaBytes := range origMeta {
		newMeta, err := f.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
		require.NoError(t, err)

		if role != "targets/a" {
			require.True(t, bytes.Equal(metaBytes, newMeta), "bytes have changed for role %s", role)
		} else {
			require.False(t, bytes.Equal(metaBytes, newMeta))
			origSigned, newSigned := &data.SignedTargets{}, &data.SignedTargets{}
			require.NoError(t, json.Unmarshal(metaBytes, origSigned))
			require.NoError(t, json.Unmarshal(newMeta, newSigned))
			require.NotEqual(t, []string{pubKey.ID()}, origSigned.Signed.Delegations.Roles[0].KeyIDs)
			require.Equal(t, []string{pubKey.ID()}, newSigned.Signed.Delegations.Roles[0].KeyIDs)
			_, ok := origSigned.Signed.Delegations.Keys[pubKey.ID()]
			require.False(t, ok)
			_, ok = newSigned.Signed.Delegations.Keys[pubKey.ID()]
			require.True(t, ok)
		}
	}
}
