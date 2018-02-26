package testutils

import (
	"bytes"
	"fmt"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/passphrase"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
)

// ErrNoKeyForRole returns an error when the cryptoservice provided to
// MetadataSwizzler has no key for a particular role
type ErrNoKeyForRole struct {
	Role data.RoleName
}

func (e ErrNoKeyForRole) Error() string {
	return "Swizzler's cryptoservice has no key for role " + e.Role.String()
}

// MetadataSwizzler fuzzes the metadata in a MetadataStore
type MetadataSwizzler struct {
	Gun           data.GUN
	MetadataCache store.MetadataStore
	CryptoService signed.CryptoService
	Roles         []data.RoleName // list of Roles in the metadataStore
}

func getPubKeys(cs signed.CryptoService, s *data.Signed, role data.RoleName) ([]data.PublicKey, error) {
	var pubKeys []data.PublicKey
	if role == data.CanonicalRootRole {
		// if this is root metadata, we have to get the keys from the root because they
		// are certs
		root := &data.Root{}
		if err := json.Unmarshal(*s.Signed, root); err != nil {
			return nil, err
		}
		rootRole, ok := root.Roles[data.CanonicalRootRole]
		if !ok || rootRole == nil {
			return nil, tuf.ErrNotLoaded{}
		}
		for _, pubKeyID := range rootRole.KeyIDs {
			pubKeys = append(pubKeys, root.Keys[pubKeyID])
		}
	} else {
		pubKeyIDs := cs.ListKeys(role)
		for _, pubKeyID := range pubKeyIDs {
			pubKey := cs.GetKey(pubKeyID)
			if pubKey != nil {
				pubKeys = append(pubKeys, pubKey)
			}
		}
	}
	return pubKeys, nil
}

// signs the new metadata, replacing whatever signature was there
func serializeMetadata(cs signed.CryptoService, s *data.Signed, role data.RoleName,
	pubKeys ...data.PublicKey) ([]byte, error) {

	// delete the existing signatures
	s.Signatures = []data.Signature{}

	if len(pubKeys) < 1 {
		return nil, ErrNoKeyForRole{role}
	}

	if err := signed.Sign(cs, s, pubKeys, 1, nil); err != nil {
		if _, ok := err.(signed.ErrInsufficientSignatures); ok {
			return nil, ErrNoKeyForRole{Role: role}
		}
		return nil, err
	}

	metaBytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	return metaBytes, nil
}

// gets a Signed from the metadata store
func signedFromStore(cache store.MetadataStore, role data.RoleName) (*data.Signed, error) {
	b, err := cache.GetSized(role.String(), store.NoSizeLimit)
	if err != nil {
		return nil, err
	}

	signed := &data.Signed{}
	if err := json.Unmarshal(b, signed); err != nil {
		return nil, err
	}

	return signed, nil
}

// NewMetadataSwizzler returns a new swizzler when given a gun,
// mapping of roles to initial metadata bytes, and a cryptoservice
func NewMetadataSwizzler(gun data.GUN, initialMetadata map[data.RoleName][]byte,
	cryptoService signed.CryptoService) *MetadataSwizzler {

	var roles []data.RoleName
	for roleName := range initialMetadata {
		roles = append(roles, roleName)
	}

	return &MetadataSwizzler{
		Gun:           gun,
		MetadataCache: store.NewMemoryStore(initialMetadata),
		CryptoService: cryptoService,
		Roles:         roles,
	}
}

// SetInvalidJSON corrupts metadata into something that is no longer valid JSON
func (m *MetadataSwizzler) SetInvalidJSON(role data.RoleName) error {
	metaBytes, err := m.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(role.String(), metaBytes[5:])
}

// AddExtraSpace adds an extra space to the beginning and end of the serialized
// JSON bytes, which should not affect serialization, but will change the checksum
// of the file.
func (m *MetadataSwizzler) AddExtraSpace(role data.RoleName) error {
	metaBytes, err := m.MetadataCache.GetSized(role.String(), store.NoSizeLimit)
	if err != nil {
		return err
	}
	newBytes := append(append([]byte{' '}, metaBytes...), ' ')
	return m.MetadataCache.Set(role.String(), newBytes)
}

// SetInvalidSigned corrupts the metadata into something that is valid JSON,
// but not unmarshallable into signed JSON
func (m *MetadataSwizzler) SetInvalidSigned(role data.RoleName) error {
	signedThing, err := signedFromStore(m.MetadataCache, role)
	if err != nil {
		return err
	}
	metaBytes, err := json.MarshalCanonical(map[string]interface{}{
		"signed":     signedThing.Signed,
		"signatures": "not list",
	})
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(role.String(), metaBytes)
}

// SetInvalidSignedMeta corrupts the metadata into something that is unmarshallable
// as a Signed object, but not unmarshallable into a SignedMeta object
func (m *MetadataSwizzler) SetInvalidSignedMeta(role data.RoleName) error {
	signedThing, err := signedFromStore(m.MetadataCache, role)
	if err != nil {
		return err
	}

	pubKeys, err := getPubKeys(m.CryptoService, signedThing, role)
	if err != nil {
		return err
	}

	var unmarshalled map[string]interface{}
	if err := json.Unmarshal(*signedThing.Signed, &unmarshalled); err != nil {
		return err
	}

	unmarshalled["_type"] = []string{"not a string"}
	unmarshalled["version"] = "string not int"
	unmarshalled["expires"] = "cannot be parsed as time"

	metaBytes, err := json.MarshalCanonical(unmarshalled)
	if err != nil {
		return err
	}
	signedThing.Signed = (*json.RawMessage)(&metaBytes)

	metaBytes, err = serializeMetadata(m.CryptoService, signedThing, role, pubKeys...)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(role.String(), metaBytes)
}

// TODO: corrupt metadata in such a way that it can be unmarshalled as a
// SignedMeta, but not as a SignedRoot or SignedTarget, etc. (Signed*)

// SetInvalidMetadataType unmarshallable, but has the wrong metadata type (not
// actually a metadata type)
func (m *MetadataSwizzler) SetInvalidMetadataType(role data.RoleName) error {
	signedThing, err := signedFromStore(m.MetadataCache, role)
	if err != nil {
		return err
	}

	var unmarshalled map[string]interface{}
	if err := json.Unmarshal(*signedThing.Signed, &unmarshalled); err != nil {
		return err
	}

	unmarshalled["_type"] = "not_real"

	metaBytes, err := json.MarshalCanonical(unmarshalled)
	if err != nil {
		return err
	}
	signedThing.Signed = (*json.RawMessage)(&metaBytes)

	pubKeys, err := getPubKeys(m.CryptoService, signedThing, role)
	if err == nil {
		metaBytes, err = serializeMetadata(m.CryptoService, signedThing, role, pubKeys...)
	}

	if err != nil {
		return err
	}
	return m.MetadataCache.Set(role.String(), metaBytes)
}

// InvalidateMetadataSignatures signs with the right key(s) but wrong hash
func (m *MetadataSwizzler) InvalidateMetadataSignatures(role data.RoleName) error {
	signedThing, err := signedFromStore(m.MetadataCache, role)
	if err != nil {
		return err
	}
	sigs := make([]data.Signature, len(signedThing.Signatures))
	for i, origSig := range signedThing.Signatures {
		sigs[i] = data.Signature{
			KeyID:     origSig.KeyID,
			Signature: []byte("invalid signature"),
			Method:    origSig.Method,
		}
	}
	signedThing.Signatures = sigs

	metaBytes, err := json.Marshal(signedThing)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(role.String(), metaBytes)
}

// TODO: AddExtraSignedInfo - add an extra field to Signed that doesn't get
//	unmarshalled, and the whole thing is correctly signed, so shouldn't cause
//  problems there.  Should this fail a canonical JSON check?

// RemoveMetadata deletes the metadata entirely
func (m *MetadataSwizzler) RemoveMetadata(role data.RoleName) error {
	return m.MetadataCache.Remove(role.String())
}

// SignMetadataWithInvalidKey signs the metadata with the wrong key
func (m *MetadataSwizzler) SignMetadataWithInvalidKey(role data.RoleName) error {
	signedThing, err := signedFromStore(m.MetadataCache, role)
	if err != nil {
		return err
	}

	// create an invalid key, but not in the existing CryptoService
	cs := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(passphrase.ConstantRetriever("")))
	key, err := CreateKey(cs, m.Gun, role, data.ECDSAKey)
	if err != nil {
		return err
	}

	metaBytes, err := serializeMetadata(cs, signedThing, "root", key)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(role.String(), metaBytes)
}

// OffsetMetadataVersion updates the metadata version
func (m *MetadataSwizzler) OffsetMetadataVersion(role data.RoleName, offset int) error {
	signedThing, err := signedFromStore(m.MetadataCache, role)
	if err != nil {
		return err
	}

	var unmarshalled map[string]interface{}
	if err := json.Unmarshal(*signedThing.Signed, &unmarshalled); err != nil {
		return err
	}

	if role == data.CanonicalRootRole {
		// store old versions of roots accessible by version
		version, ok := unmarshalled["version"].(float64)
		if !ok {
			version = float64(0) // just ignore the error and set it to 0
		}

		versionedRole := fmt.Sprintf("%d.%s", int(version), data.CanonicalRootRole)
		pubKeys, err := getPubKeys(m.CryptoService, signedThing, role)
		if err != nil {
			return err
		}
		versionedMetaBytes, err := serializeMetadata(m.CryptoService, signedThing, role, pubKeys...)
		if err != nil {
			return err
		}
		err = m.MetadataCache.Set(versionedRole, versionedMetaBytes)
		if err != nil {
			return err
		}
	}

	oldVersion, ok := unmarshalled["version"].(float64)
	if !ok {
		oldVersion = float64(0) // just ignore the error and set it to 0
	}
	unmarshalled["version"] = int(oldVersion) + offset

	metaBytes, err := json.MarshalCanonical(unmarshalled)
	if err != nil {
		return err
	}
	signedThing.Signed = (*json.RawMessage)(&metaBytes)

	pubKeys, err := getPubKeys(m.CryptoService, signedThing, role)
	if err == nil {
		metaBytes, err = serializeMetadata(m.CryptoService, signedThing, role, pubKeys...)
	}

	if err != nil {
		return err
	}
	return m.MetadataCache.Set(role.String(), metaBytes)
}

// ExpireMetadata expires the metadata, which would make it invalid - don't do anything if
// we don't have the timestamp key
func (m *MetadataSwizzler) ExpireMetadata(role data.RoleName) error {
	signedThing, err := signedFromStore(m.MetadataCache, role)
	if err != nil {
		return err
	}

	var unmarshalled map[string]interface{}
	if err := json.Unmarshal(*signedThing.Signed, &unmarshalled); err != nil {
		return err
	}

	unmarshalled["expires"] = time.Now().AddDate(-1, -1, -1)

	metaBytes, err := json.MarshalCanonical(unmarshalled)
	if err != nil {
		return err
	}
	signedThing.Signed = (*json.RawMessage)(&metaBytes)

	pubKeys, err := getPubKeys(m.CryptoService, signedThing, role)
	if err == nil {
		metaBytes, err = serializeMetadata(m.CryptoService, signedThing, role, pubKeys...)
	}

	if err != nil {
		return err
	}
	return m.MetadataCache.Set(role.String(), metaBytes)
}

// SetThreshold sets a threshold for a metadata role - can invalidate metadata for which
// the threshold is increased, if there aren't enough signatures or can be invalid because
// the threshold is 0
func (m *MetadataSwizzler) SetThreshold(role data.RoleName, newThreshold int) error {
	roleSpecifier := data.CanonicalRootRole
	if data.IsDelegation(role) {
		roleSpecifier = role.Parent()
	}

	b, err := m.MetadataCache.GetSized(roleSpecifier.String(), store.NoSizeLimit)
	if err != nil {
		return err
	}

	signedThing := &data.Signed{}
	if err := json.Unmarshal(b, signedThing); err != nil {
		return err
	}

	if roleSpecifier == data.CanonicalRootRole {
		signedRoot, err := data.RootFromSigned(signedThing)
		if err != nil {
			return err
		}
		signedRoot.Signed.Roles[role].Threshold = newThreshold
		if signedThing, err = signedRoot.ToSigned(); err != nil {
			return err
		}
	} else {
		signedTargets, err := data.TargetsFromSigned(signedThing, roleSpecifier)
		if err != nil {
			return err
		}
		for _, roleObject := range signedTargets.Signed.Delegations.Roles {
			if roleObject.Name == role {
				roleObject.Threshold = newThreshold
				break
			}
		}
		if signedThing, err = signedTargets.ToSigned(); err != nil {
			return err
		}
	}

	var metaBytes []byte
	pubKeys, err := getPubKeys(m.CryptoService, signedThing, roleSpecifier)
	if err == nil {
		metaBytes, err = serializeMetadata(m.CryptoService, signedThing, roleSpecifier, pubKeys...)
	}

	if err != nil {
		return err
	}
	return m.MetadataCache.Set(roleSpecifier.String(), metaBytes)
}

// RotateKey rotates the key for a role - this can invalidate that role's metadata
// if it is not signed by that key.  Particularly if the key being rotated is the
// root key, because it is not signed by the new key, only the old key.
func (m *MetadataSwizzler) RotateKey(role data.RoleName, key data.PublicKey) error {
	roleSpecifier := data.CanonicalRootRole
	if data.IsDelegation(role) {
		roleSpecifier = role.Parent()
	}

	b, err := m.MetadataCache.GetSized(roleSpecifier.String(), store.NoSizeLimit)
	if err != nil {
		return err
	}

	signedThing := &data.Signed{}
	if err := json.Unmarshal(b, signedThing); err != nil {
		return err
	}

	// get keys before the keys are rotated
	pubKeys, err := getPubKeys(m.CryptoService, signedThing, roleSpecifier)
	if err != nil {
		return err
	}

	if roleSpecifier == data.CanonicalRootRole {
		signedRoot, err := data.RootFromSigned(signedThing)
		if err != nil {
			return err
		}
		signedRoot.Signed.Roles[role].KeyIDs = []string{key.ID()}
		signedRoot.Signed.Keys[key.ID()] = key
		if signedThing, err = signedRoot.ToSigned(); err != nil {
			return err
		}
	} else {
		signedTargets, err := data.TargetsFromSigned(signedThing, roleSpecifier)
		if err != nil {
			return err
		}
		for _, roleObject := range signedTargets.Signed.Delegations.Roles {
			if roleObject.Name == role {
				roleObject.KeyIDs = []string{key.ID()}
				break
			}
		}
		signedTargets.Signed.Delegations.Keys[key.ID()] = key
		if signedThing, err = signedTargets.ToSigned(); err != nil {
			return err
		}
	}

	metaBytes, err := serializeMetadata(m.CryptoService, signedThing, roleSpecifier, pubKeys...)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(roleSpecifier.String(), metaBytes)
}

// ChangeRootKey swaps out the root key with a new key, and re-signs the metadata
// with the new key
func (m *MetadataSwizzler) ChangeRootKey() error {
	key, err := CreateKey(m.CryptoService, m.Gun, data.CanonicalRootRole, data.ECDSAKey)
	if err != nil {
		return err
	}

	b, err := m.MetadataCache.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
	if err != nil {
		return err
	}

	signedRoot := &data.SignedRoot{}
	if err := json.Unmarshal(b, signedRoot); err != nil {
		return err
	}

	signedRoot.Signed.Keys[key.ID()] = key
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{key.ID()}

	var signedThing *data.Signed
	if signedThing, err = signedRoot.ToSigned(); err != nil {
		return err
	}

	var metaBytes []byte
	pubKeys, err := getPubKeys(m.CryptoService, signedThing, data.CanonicalRootRole)
	if err == nil {
		metaBytes, err = serializeMetadata(m.CryptoService, signedThing, data.CanonicalRootRole, pubKeys...)
	}

	if err != nil {
		return err
	}
	return m.MetadataCache.Set(data.CanonicalRootRole.String(), metaBytes)
}

// UpdateSnapshotHashes updates the snapshot to reflect the latest hash changes, to
// ensure that failure isn't because the snapshot has the wrong hash.
func (m *MetadataSwizzler) UpdateSnapshotHashes(roles ...data.RoleName) error {
	var (
		metaBytes      []byte
		snapshotSigned *data.Signed
		err            error
	)
	if metaBytes, err = m.MetadataCache.GetSized(data.CanonicalSnapshotRole.String(), store.NoSizeLimit); err != nil {
		return err
	}

	snapshot := data.SignedSnapshot{}
	if err = json.Unmarshal(metaBytes, &snapshot); err != nil {
		return err
	}

	// just rebuild everything if roles is not specified
	if len(roles) == 0 {
		roles = m.Roles
	}

	for _, role := range roles {
		if role != data.CanonicalSnapshotRole && role != data.CanonicalTimestampRole {
			if metaBytes, err = m.MetadataCache.GetSized(role.String(), store.NoSizeLimit); err != nil {
				return err
			}

			meta, err := data.NewFileMeta(bytes.NewReader(metaBytes), data.NotaryDefaultHashes...)
			if err != nil {
				return err
			}

			snapshot.Signed.Meta[role.String()] = meta
		}
	}

	if snapshotSigned, err = snapshot.ToSigned(); err != nil {
		return err
	}
	pubKeys, err := getPubKeys(m.CryptoService, snapshotSigned, data.CanonicalSnapshotRole)
	if err == nil {
		metaBytes, err = serializeMetadata(m.CryptoService, snapshotSigned, data.CanonicalSnapshotRole, pubKeys...)
	}

	if err != nil {
		return err
	}
	return m.MetadataCache.Set(data.CanonicalSnapshotRole.String(), metaBytes)
}

// UpdateTimestampHash updates the timestamp to reflect the latest snapshot changes, to
// ensure that failure isn't because the timestamp has the wrong hash.
func (m *MetadataSwizzler) UpdateTimestampHash() error {
	var (
		metaBytes       []byte
		timestamp       = &data.SignedTimestamp{}
		timestampSigned *data.Signed
		err             error
	)
	if metaBytes, err = m.MetadataCache.GetSized(data.CanonicalTimestampRole.String(), store.NoSizeLimit); err != nil {
		return err
	}
	// we can't just create a new timestamp, because then the expiry would be
	// different
	if err = json.Unmarshal(metaBytes, timestamp); err != nil {
		return err
	}

	if metaBytes, err = m.MetadataCache.GetSized(data.CanonicalSnapshotRole.String(), store.NoSizeLimit); err != nil {
		return err
	}

	snapshotMeta, err := data.NewFileMeta(bytes.NewReader(metaBytes), data.NotaryDefaultHashes...)
	if err != nil {
		return err
	}

	timestamp.Signed.Meta[data.CanonicalSnapshotRole.String()] = snapshotMeta

	timestampSigned, err = timestamp.ToSigned()
	if err != nil {
		return err
	}
	pubKeys, err := getPubKeys(m.CryptoService, timestampSigned, data.CanonicalTimestampRole)
	if err == nil {
		metaBytes, err = serializeMetadata(m.CryptoService, timestampSigned, data.CanonicalTimestampRole, pubKeys...)
	}

	if err != nil {
		return err
	}
	return m.MetadataCache.Set(data.CanonicalTimestampRole.String(), metaBytes)
}

// MutateRoot takes a function that mutates the root metadata - once done, it
// serializes the root again
func (m *MetadataSwizzler) MutateRoot(mutate func(*data.Root)) error {
	signedThing, err := signedFromStore(m.MetadataCache, data.CanonicalRootRole)
	if err != nil {
		return err
	}

	var root data.Root
	if err := json.Unmarshal(*signedThing.Signed, &root); err != nil {
		return err
	}

	// get the original keys, in case the mutation messes with the signing keys
	oldPubKeys, err := getPubKeys(m.CryptoService, signedThing, data.CanonicalRootRole)
	if err != nil {
		return err
	}

	mutate(&root)

	sRoot := &data.SignedRoot{Signed: root, Signatures: signedThing.Signatures}
	signedThing, err = sRoot.ToSigned()
	if err != nil {
		return err
	}

	pubKeys, err := getPubKeys(m.CryptoService, signedThing, data.CanonicalRootRole)
	if err != nil || len(pubKeys) == 0 { // we have to sign it somehow - might as well use the old keys
		pubKeys = oldPubKeys
	}

	metaBytes, err := serializeMetadata(m.CryptoService, signedThing, data.CanonicalRootRole, pubKeys...)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(data.CanonicalRootRole.String(), metaBytes)
}

// MutateTimestamp takes a function that mutates the timestamp metadata - once done, it
// serializes the timestamp again
func (m *MetadataSwizzler) MutateTimestamp(mutate func(*data.Timestamp)) error {
	signedThing, err := signedFromStore(m.MetadataCache, data.CanonicalTimestampRole)
	if err != nil {
		return err
	}

	var timestamp data.Timestamp
	if err := json.Unmarshal(*signedThing.Signed, &timestamp); err != nil {
		return err
	}

	mutate(&timestamp)

	sTimestamp := &data.SignedTimestamp{Signed: timestamp, Signatures: signedThing.Signatures}
	signedThing, err = sTimestamp.ToSigned()
	if err != nil {
		return err
	}

	pubKeys, err := getPubKeys(m.CryptoService, signedThing, data.CanonicalTimestampRole)
	if err != nil {
		return err
	}

	metaBytes, err := serializeMetadata(m.CryptoService, signedThing, data.CanonicalTimestampRole, pubKeys...)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(data.CanonicalTimestampRole.String(), metaBytes)
}

// MutateSnapshot takes a function that mutates the snapshot metadata - once done, it
// serializes the snapshot again
func (m *MetadataSwizzler) MutateSnapshot(mutate func(*data.Snapshot)) error {
	signedThing, err := signedFromStore(m.MetadataCache, data.CanonicalSnapshotRole)
	if err != nil {
		return err
	}

	var snapshot data.Snapshot
	if err := json.Unmarshal(*signedThing.Signed, &snapshot); err != nil {
		return err
	}

	mutate(&snapshot)

	sSnapshot := &data.SignedSnapshot{Signed: snapshot, Signatures: signedThing.Signatures}
	signedThing, err = sSnapshot.ToSigned()
	if err != nil {
		return err
	}

	pubKeys, err := getPubKeys(m.CryptoService, signedThing, data.CanonicalSnapshotRole)
	if err != nil {
		return err
	}

	metaBytes, err := serializeMetadata(m.CryptoService, signedThing, data.CanonicalSnapshotRole, pubKeys...)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(data.CanonicalSnapshotRole.String(), metaBytes)
}

// MutateTargets takes a function that mutates the targets metadata - once done, it
// serializes the targets again
func (m *MetadataSwizzler) MutateTargets(mutate func(*data.Targets)) error {
	signedThing, err := signedFromStore(m.MetadataCache, data.CanonicalTargetsRole)
	if err != nil {
		return err
	}

	var targets data.Targets
	if err := json.Unmarshal(*signedThing.Signed, &targets); err != nil {
		return err
	}

	mutate(&targets)

	sTargets := &data.SignedTargets{Signed: targets, Signatures: signedThing.Signatures}
	signedThing, err = sTargets.ToSigned()
	if err != nil {
		return err
	}

	pubKeys, err := getPubKeys(m.CryptoService, signedThing, data.CanonicalTargetsRole)
	if err != nil {
		return err
	}

	metaBytes, err := serializeMetadata(m.CryptoService, signedThing, data.CanonicalTargetsRole, pubKeys...)
	if err != nil {
		return err
	}
	return m.MetadataCache.Set(data.CanonicalTargetsRole.String(), metaBytes)
}
