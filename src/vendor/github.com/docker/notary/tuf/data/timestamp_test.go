package data

import (
	"bytes"
	"crypto/sha256"
	"crypto/sha512"
	rjson "encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	cjson "github.com/docker/go/canonical/json"
	"github.com/stretchr/testify/require"
)

func validTimestampTemplate() *SignedTimestamp {
	return &SignedTimestamp{
		Signed: Timestamp{
			SignedCommon: SignedCommon{Type: TUFTypes[CanonicalTimestampRole], Version: 1, Expires: time.Now()},
			Meta: Files{
				CanonicalSnapshotRole.String(): FileMeta{Hashes: Hashes{"sha256": bytes.Repeat([]byte("a"), sha256.Size)}},
			}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
		},
	}
}

func TestTimestampToSignedMarshalsSignedPortionWithCanonicalJSON(t *testing.T) {
	ts := SignedTimestamp{Signed: Timestamp{
		SignedCommon: SignedCommon{Type: TUFTypes[CanonicalTimestampRole], Version: 1, Expires: time.Now()}}}
	signedCanonical, err := ts.ToSigned()
	require.NoError(t, err)

	canonicalSignedPortion, err := cjson.MarshalCanonical(ts.Signed)
	require.NoError(t, err)

	castedCanonical := rjson.RawMessage(canonicalSignedPortion)

	// don't bother testing regular JSON because it might not be different

	require.True(t, bytes.Equal(*signedCanonical.Signed, castedCanonical),
		"expected %v == %v", signedCanonical.Signed, castedCanonical)
}

func TestTimestampToSignCopiesSignatures(t *testing.T) {
	ts := SignedTimestamp{
		Signed: Timestamp{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalTimestampRole], Version: 2, Expires: time.Now()}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
		},
	}
	signed, err := ts.ToSigned()
	require.NoError(t, err)

	require.True(t, reflect.DeepEqual(ts.Signatures, signed.Signatures),
		"expected %v == %v", ts.Signatures, signed.Signatures)

	ts.Signatures[0].KeyID = "changed"
	require.False(t, reflect.DeepEqual(ts.Signatures, signed.Signatures),
		"expected %v != %v", ts.Signatures, signed.Signatures)
}

func TestTimestampToSignedMarshallingErrorsPropagated(t *testing.T) {
	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})
	ts := SignedTimestamp{
		Signed: Timestamp{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalTimestampRole], Version: 2, Expires: time.Now()}},
	}
	_, err := ts.ToSigned()
	require.EqualError(t, err, "bad")
}

func TestTimestampMarshalJSONMarshalsSignedWithRegularJSON(t *testing.T) {
	ts := SignedTimestamp{
		Signed: Timestamp{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalTimestampRole], Version: 1, Expires: time.Now()}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
			{KeyID: "key2", Method: "method2", Signature: []byte("there")},
		},
	}
	serialized, err := ts.MarshalJSON()
	require.NoError(t, err)

	signed, err := ts.ToSigned()
	require.NoError(t, err)

	// don't bother testing canonical JSON because it might not be different

	regular, err := rjson.Marshal(signed)
	require.NoError(t, err)

	require.True(t, bytes.Equal(serialized, regular),
		"expected %v != %v", serialized, regular)
}

func TestTimestampMarshalJSONMarshallingErrorsPropagated(t *testing.T) {
	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})
	ts := SignedTimestamp{
		Signed: Timestamp{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalTimestampRole], Version: 2, Expires: time.Now()}},
	}
	_, err := ts.MarshalJSON()
	require.EqualError(t, err, "bad")
}

func TestTimestampFromSignedUnmarshallingErrorsPropagated(t *testing.T) {
	signed, err := validTimestampTemplate().ToSigned()
	require.NoError(t, err)

	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})

	_, err = TimestampFromSigned(signed)
	require.EqualError(t, err, "bad")
}

// TimestampFromSigned succeeds if the timestamp is valid, and copies the signatures
// rather than assigns them
func TestTimestampFromSignedCopiesSignatures(t *testing.T) {
	signed, err := validTimestampTemplate().ToSigned()
	require.NoError(t, err)

	signedTimestamp, err := TimestampFromSigned(signed)
	require.NoError(t, err)

	signed.Signatures[0] = Signature{KeyID: "key3", Method: "method3", Signature: []byte("world")}

	require.Equal(t, "key3", signed.Signatures[0].KeyID)
	require.Equal(t, "key1", signedTimestamp.Signatures[0].KeyID)
}

func timestampToSignedAndBack(t *testing.T, timestamp *SignedTimestamp) (*SignedTimestamp, error) {
	s, err := timestamp.ToSigned()
	require.NoError(t, err)
	return TimestampFromSigned(s)
}

// If the snapshot metadata is missing, the timestamp metadata fails to validate
// and thus fails to convert into a SignedTimestamp
func TestTimestampFromSignedValidatesMeta(t *testing.T) {
	var err error
	ts := validTimestampTemplate()

	// invalid checksum length
	ts.Signed.Meta[CanonicalSnapshotRole.String()].Hashes["sha256"] = []byte("too short")
	_, err = timestampToSignedAndBack(t, ts)
	require.IsType(t, ErrInvalidMetadata{}, err)

	// missing sha256 checksum
	delete(ts.Signed.Meta[CanonicalSnapshotRole.String()].Hashes, "sha256")
	_, err = timestampToSignedAndBack(t, ts)
	require.IsType(t, ErrInvalidMetadata{}, err)

	// add a different checksum to make sure it's not failing because of the hash length
	ts.Signed.Meta[CanonicalSnapshotRole.String()].Hashes["sha512"] = bytes.Repeat([]byte("a"), sha512.Size)
	_, err = timestampToSignedAndBack(t, ts)
	require.IsType(t, ErrInvalidMetadata{}, err)

	// delete the ckechsum metadata entirely for the role
	delete(ts.Signed.Meta, CanonicalSnapshotRole.String())
	_, err = timestampToSignedAndBack(t, ts)
	require.IsType(t, ErrInvalidMetadata{}, err)

	// add some extra metadata to make sure it's not failing because the metadata
	// is empty
	ts.Signed.Meta[CanonicalSnapshotRole.String()] = FileMeta{}
	_, err = timestampToSignedAndBack(t, ts)
	require.IsType(t, ErrInvalidMetadata{}, err)
}

// Type must be "Timestamp"
func TestTimestampFromSignedValidatesRoleType(t *testing.T) {
	ts := validTimestampTemplate()
	tsType := TUFTypes[CanonicalTimestampRole]

	for _, invalid := range []string{" " + tsType, CanonicalSnapshotRole.String(), strings.ToUpper(tsType)} {
		ts.Signed.Type = invalid
		s, err := ts.ToSigned()
		require.NoError(t, err)
		_, err = TimestampFromSigned(s)
		require.IsType(t, ErrInvalidMetadata{}, err)
	}

	ts = validTimestampTemplate()
	ts.Signed.Type = tsType
	sTimestamp, err := timestampToSignedAndBack(t, ts)
	require.NoError(t, err)
	require.Equal(t, tsType, sTimestamp.Signed.Type)
}

// The version cannot be negative
func TestTimestampFromSignedValidatesVersion(t *testing.T) {
	ts := validTimestampTemplate()
	ts.Signed.Version = -1
	_, err := timestampToSignedAndBack(t, ts)
	require.IsType(t, ErrInvalidMetadata{}, err)

	ts.Signed.Version = 0
	_, err = timestampToSignedAndBack(t, ts)
	require.IsType(t, ErrInvalidMetadata{}, err)

	ts.Signed.Version = 1
	_, err = timestampToSignedAndBack(t, ts)
	require.NoError(t, err)
}

// GetSnapshot returns the snapshot checksum, or an error if it is missing.
func TestTimestampGetSnapshot(t *testing.T) {
	ts := validTimestampTemplate()
	f, err := ts.GetSnapshot()
	require.NoError(t, err)
	require.IsType(t, &FileMeta{}, f)

	// no timestamp meta
	delete(ts.Signed.Meta, CanonicalSnapshotRole.String())
	f, err = ts.GetSnapshot()
	require.Error(t, err)
	require.IsType(t, ErrMissingMeta{}, err)
	require.Nil(t, f)
}
