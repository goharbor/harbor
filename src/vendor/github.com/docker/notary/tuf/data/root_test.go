package data

import (
	"bytes"
	rjson "encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

	cjson "github.com/docker/go/canonical/json"
	"github.com/stretchr/testify/require"
)

type errorSerializer struct {
	canonicalJSON
}

func (e errorSerializer) MarshalCanonical(interface{}) ([]byte, error) {
	return nil, fmt.Errorf("bad")
}

func (e errorSerializer) Unmarshal([]byte, interface{}) error {
	return fmt.Errorf("bad")
}

func validRootTemplate() *SignedRoot {
	return &SignedRoot{
		Signed: Root{
			SignedCommon: SignedCommon{
				Type:    TUFTypes[CanonicalRootRole],
				Version: 1,
				Expires: time.Now(),
			},
			Keys: Keys{
				"key1":  NewPublicKey(RSAKey, []byte("key1")),
				"key2":  NewPublicKey(RSAKey, []byte("key2")),
				"key3":  NewPublicKey(RSAKey, []byte("key3")),
				"snKey": NewPublicKey(RSAKey, []byte("snKey")),
				"tgKey": NewPublicKey(RSAKey, []byte("tgKey")),
				"tsKey": NewPublicKey(RSAKey, []byte("tsKey")),
			},
			Roles: map[RoleName]*RootRole{
				CanonicalRootRole:      {KeyIDs: []string{"key1"}, Threshold: 1},
				CanonicalSnapshotRole:  {KeyIDs: []string{"snKey"}, Threshold: 1},
				CanonicalTimestampRole: {KeyIDs: []string{"tsKey"}, Threshold: 1},
				CanonicalTargetsRole:   {KeyIDs: []string{"tgKey"}, Threshold: 1},
			},
		},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
		},
	}
}

func TestRootToSignedMarshalsSignedPortionWithCanonicalJSON(t *testing.T) {
	r := SignedRoot{Signed: Root{SignedCommon: SignedCommon{
		Type: TUFTypes[CanonicalRootRole], Version: 2, Expires: time.Now()}}}
	signedCanonical, err := r.ToSigned()
	require.NoError(t, err)

	canonicalSignedPortion, err := cjson.MarshalCanonical(r.Signed)
	require.NoError(t, err)

	castedCanonical := rjson.RawMessage(canonicalSignedPortion)

	// don't bother testing regular JSON because it might not be different

	require.True(t, bytes.Equal(*signedCanonical.Signed, castedCanonical),
		"expected %v == %v", signedCanonical.Signed, castedCanonical)
}

func TestRootToSignCopiesSignatures(t *testing.T) {
	r := SignedRoot{
		Signed: Root{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalRootRole], Version: 2, Expires: time.Now()}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
		},
	}
	signed, err := r.ToSigned()
	require.NoError(t, err)

	require.True(t, reflect.DeepEqual(r.Signatures, signed.Signatures),
		"expected %v == %v", r.Signatures, signed.Signatures)

	r.Signatures[0].KeyID = "changed"
	require.False(t, reflect.DeepEqual(r.Signatures, signed.Signatures),
		"expected %v != %v", r.Signatures, signed.Signatures)
}

func TestRootToSignedMarshallingErrorsPropagated(t *testing.T) {
	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})
	r := SignedRoot{
		Signed: Root{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalRootRole], Version: 2, Expires: time.Now()}},
	}
	_, err := r.ToSigned()
	require.EqualError(t, err, "bad")
}

func TestRootMarshalJSONMarshalsSignedWithRegularJSON(t *testing.T) {
	r := SignedRoot{
		Signed: Root{SignedCommon: SignedCommon{Type: "root", Version: 2, Expires: time.Now()}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
			{KeyID: "key2", Method: "method2", Signature: []byte("there")},
		},
	}
	serialized, err := r.MarshalJSON()
	require.NoError(t, err)

	signed, err := r.ToSigned()
	require.NoError(t, err)

	// don't bother testing canonical JSON because it might not be different

	regular, err := rjson.Marshal(signed)
	require.NoError(t, err)

	require.True(t, bytes.Equal(serialized, regular),
		"expected %v != %v", serialized, regular)
}

func TestRootMarshalJSONMarshallingErrorsPropagated(t *testing.T) {
	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})
	r := SignedRoot{
		Signed: Root{SignedCommon: SignedCommon{
			Type: TUFTypes[CanonicalRootRole], Version: 2, Expires: time.Now()}},
	}
	_, err := r.MarshalJSON()
	require.EqualError(t, err, "bad")
}

func TestRootFromSignedUnmarshallingErrorsPropagated(t *testing.T) {
	signed, err := validRootTemplate().ToSigned()
	require.NoError(t, err)

	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})

	_, err = RootFromSigned(signed)
	require.EqualError(t, err, "bad")
}

// RootFromSigned succeeds if the root is valid, and copies the signatures
// rather than assigns them
func TestRootFromSignedCopiesSignatures(t *testing.T) {
	signed, err := validRootTemplate().ToSigned()
	require.NoError(t, err)

	signedRoot, err := RootFromSigned(signed)
	require.NoError(t, err)

	signed.Signatures[0] = Signature{KeyID: "key3", Method: "method3", Signature: []byte("world")}

	require.Equal(t, "key3", signed.Signatures[0].KeyID)
	require.Equal(t, "key1", signedRoot.Signatures[0].KeyID)
}

func rootToSignedAndBack(t *testing.T, root *SignedRoot) (*SignedRoot, error) {
	s, err := root.ToSigned()
	require.NoError(t, err)
	return RootFromSigned(s)
}

// If the role data specified is nil, or has an invalid threshold, or doesn't have enough
// keys to cover the threshold, or has key IDs that are not in the key list, the root
// metadata fails to validate and thus fails to convert into a SignedRoot
func TestRootFromSignedValidatesRoleData(t *testing.T) {
	var err error
	for _, roleName := range BaseRoles {
		root := validRootTemplate()

		// Invalid threshold
		root.Signed.Roles[roleName].Threshold = 0
		_, err = rootToSignedAndBack(t, root)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// Keys that aren't in the list of keys
		root.Signed.Roles[roleName].Threshold = 1
		root.Signed.Roles[roleName].KeyIDs = []string{"key11"}
		_, err = rootToSignedAndBack(t, root)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// role is nil
		root.Signed.Roles[roleName] = nil
		_, err = rootToSignedAndBack(t, root)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// too few roles
		delete(root.Signed.Roles, roleName)
		_, err = rootToSignedAndBack(t, root)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// add an extra role that doesn't belong, so that the number of roles
		// is correct a required one is still missing
		root.Signed.Roles["extraneous"] = &RootRole{KeyIDs: []string{"key3"}, Threshold: 1}
		_, err = rootToSignedAndBack(t, root)
		require.IsType(t, ErrInvalidMetadata{}, err)
	}
}

// The type must be "Root"
func TestRootFromSignedValidatesRoleType(t *testing.T) {
	root := validRootTemplate()

	for _, invalid := range []string{"Root ", CanonicalSnapshotRole.String(), "rootroot", "RoOt", "root"} {
		root.Signed.Type = invalid
		_, err := rootToSignedAndBack(t, root)
		require.IsType(t, ErrInvalidMetadata{}, err)
	}

	root.Signed.Type = "Root"
	sRoot, err := rootToSignedAndBack(t, root)
	require.NoError(t, err)
	require.Equal(t, "Root", sRoot.Signed.Type)
}

// The version cannot be negative
func TestRootFromSignedValidatesVersion(t *testing.T) {
	root := validRootTemplate()
	root.Signed.Version = -1
	_, err := rootToSignedAndBack(t, root)
	require.IsType(t, ErrInvalidMetadata{}, err)

	root.Signed.Version = 0
	_, err = rootToSignedAndBack(t, root)
	require.IsType(t, ErrInvalidMetadata{}, err)

	root.Signed.Version = 1
	_, err = rootToSignedAndBack(t, root)
	require.NoError(t, err)
}
