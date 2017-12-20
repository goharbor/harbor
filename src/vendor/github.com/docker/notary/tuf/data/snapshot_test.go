package data

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	rjson "encoding/json"
	"reflect"
	"testing"
	"time"

	cjson "github.com/docker/go/canonical/json"
	"github.com/stretchr/testify/require"
)

func validSnapshotTemplate() *SignedSnapshot {
	return &SignedSnapshot{
		Signed: Snapshot{
			SignedCommon: SignedCommon{Type: TUFTypes[CanonicalSnapshotRole], Version: 1, Expires: time.Now()},
			Meta: Files{
				CanonicalRootRole.String():    FileMeta{Hashes: Hashes{"sha256": bytes.Repeat([]byte("a"), sha256.Size)}},
				CanonicalTargetsRole.String(): FileMeta{Hashes: Hashes{"sha256": bytes.Repeat([]byte("a"), sha256.Size)}},
				"targets/a":                   FileMeta{},
			}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
		},
	}
}

func TestSnapshotToSignedMarshalsSignedPortionWithCanonicalJSON(t *testing.T) {
	sn := SignedSnapshot{Signed: Snapshot{SignedCommon: SignedCommon{
		Type: TUFTypes[CanonicalSnapshotRole], Version: 1, Expires: time.Now()}}}
	signedCanonical, err := sn.ToSigned()
	require.NoError(t, err)

	canonicalSignedPortion, err := cjson.MarshalCanonical(sn.Signed)
	require.NoError(t, err)

	castedCanonical := rjson.RawMessage(canonicalSignedPortion)

	// don't bother testing regular JSON because it might not be different

	require.True(t, bytes.Equal(*signedCanonical.Signed, castedCanonical),
		"expected %v == %v", signedCanonical.Signed, castedCanonical)
}

func TestSnapshotToSignCopiesSignatures(t *testing.T) {
	sn := SignedSnapshot{
		Signed: Snapshot{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalSnapshotRole], Version: 2, Expires: time.Now()}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
		},
	}
	signed, err := sn.ToSigned()
	require.NoError(t, err)

	require.True(t, reflect.DeepEqual(sn.Signatures, signed.Signatures),
		"expected %v == %v", sn.Signatures, signed.Signatures)

	sn.Signatures[0].KeyID = "changed"
	require.False(t, reflect.DeepEqual(sn.Signatures, signed.Signatures),
		"expected %v != %v", sn.Signatures, signed.Signatures)
}

func TestSnapshotToSignedMarshallingErrorsPropagated(t *testing.T) {
	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})
	sn := SignedSnapshot{
		Signed: Snapshot{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalSnapshotRole], Version: 2, Expires: time.Now()}},
	}
	_, err := sn.ToSigned()
	require.EqualError(t, err, "bad")
}

func TestSnapshotMarshalJSONMarshalsSignedWithRegularJSON(t *testing.T) {
	sn := SignedSnapshot{
		Signed: Snapshot{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalSnapshotRole], Version: 1, Expires: time.Now()}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
			{KeyID: "key2", Method: "method2", Signature: []byte("there")},
		},
	}
	serialized, err := sn.MarshalJSON()
	require.NoError(t, err)

	signed, err := sn.ToSigned()
	require.NoError(t, err)

	// don't bother testing canonical JSON because it might not be different

	regular, err := rjson.Marshal(signed)
	require.NoError(t, err)

	require.True(t, bytes.Equal(serialized, regular),
		"expected %v != %v", serialized, regular)
}

func TestSnapshotMarshalJSONMarshallingErrorsPropagated(t *testing.T) {
	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})
	sn := SignedSnapshot{
		Signed: Snapshot{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalSnapshotRole], Version: 2, Expires: time.Now()}},
	}
	_, err := sn.MarshalJSON()
	require.EqualError(t, err, "bad")
}

func TestSnapshotFromSignedUnmarshallingErrorsPropagated(t *testing.T) {
	signed, err := validSnapshotTemplate().ToSigned()
	require.NoError(t, err)

	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})

	_, err = SnapshotFromSigned(signed)
	require.EqualError(t, err, "bad")
}

// SnapshotFromSigned succeeds if the snapshot is valid, and copies the signatures
// rather than assigns them
func TestSnapshotFromSignedCopiesSignatures(t *testing.T) {
	signed, err := validSnapshotTemplate().ToSigned()
	require.NoError(t, err)

	signedSnapshot, err := SnapshotFromSigned(signed)
	require.NoError(t, err)

	signed.Signatures[0] = Signature{KeyID: "key3", Method: "method3", Signature: []byte("world")}

	require.Equal(t, "key3", signed.Signatures[0].KeyID)
	require.Equal(t, "key1", signedSnapshot.Signatures[0].KeyID)
}

func snapshotToSignedAndBack(t *testing.T, snapshot *SignedSnapshot) (*SignedSnapshot, error) {
	s, err := snapshot.ToSigned()
	require.NoError(t, err)
	return SnapshotFromSigned(s)
}

// If the root or targets metadata is missing, the snapshot metadata fails to validate
// and thus fails to convert into a SignedSnapshot
func TestSnapshotFromSignedValidatesMeta(t *testing.T) {
	var err error
	for _, roleName := range []RoleName{CanonicalRootRole, CanonicalTargetsRole} {
		sn := validSnapshotTemplate()

		// invalid checksum length
		sn.Signed.Meta[roleName.String()].Hashes["sha256"] = []byte("too short")
		_, err = snapshotToSignedAndBack(t, sn)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// missing sha256 checksum
		delete(sn.Signed.Meta[roleName.String()].Hashes, "sha256")
		_, err = snapshotToSignedAndBack(t, sn)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// add a different checksum to make sure it's not failing because of the hash length
		sn.Signed.Meta[roleName.String()].Hashes["sha512"] = bytes.Repeat([]byte("a"), sha512.Size)
		_, err = snapshotToSignedAndBack(t, sn)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// delete the ckechsum metadata entirely for the role
		delete(sn.Signed.Meta, roleName.String())
		_, err = snapshotToSignedAndBack(t, sn)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// add some extra metadata to make sure it's not failing because the metadata
		// is empty
		sn.Signed.Meta[CanonicalSnapshotRole.String()] = FileMeta{}
		_, err = snapshotToSignedAndBack(t, sn)
		require.IsType(t, ErrInvalidMetadata{}, err)
	}
}

// Type must be "Snapshot"
func TestSnapshotFromSignedValidatesRoleType(t *testing.T) {
	sn := validSnapshotTemplate()

	for _, invalid := range []string{" Snapshot", CanonicalSnapshotRole.String(), "TIMESTAMP"} {
		sn.Signed.Type = invalid
		_, err := snapshotToSignedAndBack(t, sn)
		require.IsType(t, ErrInvalidMetadata{}, err)
	}

	sn = validSnapshotTemplate()
	sn.Signed.Type = TUFTypes[CanonicalSnapshotRole]
	sSnapshot, err := snapshotToSignedAndBack(t, sn)
	require.NoError(t, err)
	require.Equal(t, TUFTypes[CanonicalSnapshotRole], sSnapshot.Signed.Type)
}

// The version cannot be negative
func TestSnapshotFromSignedValidatesVersion(t *testing.T) {
	sn := validSnapshotTemplate()
	sn.Signed.Version = -1
	_, err := snapshotToSignedAndBack(t, sn)
	require.IsType(t, ErrInvalidMetadata{}, err)

	sn.Signed.Version = 0
	_, err = snapshotToSignedAndBack(t, sn)
	require.IsType(t, ErrInvalidMetadata{}, err)

	sn.Signed.Version = 1
	_, err = snapshotToSignedAndBack(t, sn)
	require.NoError(t, err)
}

// GetMeta returns the checksum, or an error if it is missing.
func TestSnapshotGetMeta(t *testing.T) {
	ts := validSnapshotTemplate()
	f, err := ts.GetMeta(CanonicalRootRole)
	require.NoError(t, err)
	require.IsType(t, &FileMeta{}, f)

	// now one that doesn't exist
	f, err = ts.GetMeta("targets/a/b")
	require.Error(t, err)
	require.IsType(t, ErrMissingMeta{}, err)
	require.Nil(t, f)
}
