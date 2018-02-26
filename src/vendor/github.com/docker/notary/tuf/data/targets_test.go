package data

import (
	"bytes"
	rjson "encoding/json"
	"path"
	"reflect"
	"testing"
	"time"

	cjson "github.com/docker/go/canonical/json"
	"github.com/stretchr/testify/require"
)

func validTargetsTemplate() *SignedTargets {
	return &SignedTargets{
		Signed: Targets{
			SignedCommon: SignedCommon{Type: "Targets", Version: 1, Expires: time.Now()},
			Targets:      Files{},
			Delegations: Delegations{
				Roles: []*Role{},
				Keys: Keys{
					"key1": NewPublicKey(RSAKey, []byte("key1")),
					"key2": NewPublicKey(RSAKey, []byte("key2")),
				},
			},
		},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
		},
	}
}

func TestTargetsToSignedMarshalsSignedPortionWithCanonicalJSON(t *testing.T) {
	tg := SignedTargets{Signed: Targets{SignedCommon: SignedCommon{Type: "Targets", Version: 1, Expires: time.Now()}}}
	signedCanonical, err := tg.ToSigned()
	require.NoError(t, err)

	canonicalSignedPortion, err := cjson.MarshalCanonical(tg.Signed)
	require.NoError(t, err)

	castedCanonical := rjson.RawMessage(canonicalSignedPortion)

	// don't bother testing regular JSON because it might not be different

	require.True(t, bytes.Equal(*signedCanonical.Signed, castedCanonical),
		"expected %v == %v", signedCanonical.Signed, castedCanonical)
}

func TestTargetsToSignCopiesSignatures(t *testing.T) {
	tg := SignedTargets{
		Signed: Targets{SignedCommon: SignedCommon{Type: "Targets", Version: 2, Expires: time.Now()}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
		},
	}
	signed, err := tg.ToSigned()
	require.NoError(t, err)

	require.True(t, reflect.DeepEqual(tg.Signatures, signed.Signatures),
		"expected %v == %v", tg.Signatures, signed.Signatures)

	tg.Signatures[0].KeyID = "changed"
	require.False(t, reflect.DeepEqual(tg.Signatures, signed.Signatures),
		"expected %v != %v", tg.Signatures, signed.Signatures)
}

func TestTargetsToSignedMarshallingErrorsPropagated(t *testing.T) {
	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})
	tg := SignedTargets{
		Signed: Targets{SignedCommon: SignedCommon{Type: "Targets", Version: 2, Expires: time.Now()}},
	}
	_, err := tg.ToSigned()
	require.EqualError(t, err, "bad")
}

func TestTargetsMarshalJSONMarshalsSignedWithRegularJSON(t *testing.T) {
	tg := SignedTargets{
		Signed: Targets{SignedCommon: SignedCommon{Type: "Targets", Version: 1, Expires: time.Now()}},
		Signatures: []Signature{
			{KeyID: "key1", Method: "method1", Signature: []byte("hello")},
			{KeyID: "key2", Method: "method2", Signature: []byte("there")},
		},
	}
	serialized, err := tg.MarshalJSON()
	require.NoError(t, err)

	signed, err := tg.ToSigned()
	require.NoError(t, err)

	// don't bother testing canonical JSON because it might not be different

	regular, err := rjson.Marshal(signed)
	require.NoError(t, err)

	require.True(t, bytes.Equal(serialized, regular),
		"expected %v != %v", serialized, regular)
}

func TestTargetsMarshalJSONMarshallingErrorsPropagated(t *testing.T) {
	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})
	tg := SignedTargets{
		Signed: Targets{SignedCommon: SignedCommon{Type: "Targets", Version: 2, Expires: time.Now()}},
	}
	_, err := tg.MarshalJSON()
	require.EqualError(t, err, "bad")
}

func TestTargetsFromSignedUnmarshallingErrorsPropagated(t *testing.T) {
	signed, err := validTargetsTemplate().ToSigned()
	require.NoError(t, err)

	setDefaultSerializer(errorSerializer{})
	defer setDefaultSerializer(canonicalJSON{})

	_, err = TargetsFromSigned(signed, CanonicalTargetsRole)
	require.EqualError(t, err, "bad")
}

// TargetsFromSigned succeeds if the targets is valid, and copies the signatures
// rather than assigns them
func TestTargetsFromSignedCopiesSignatures(t *testing.T) {
	for _, roleName := range []RoleName{CanonicalTargetsRole, RoleName(path.Join(CanonicalTargetsRole.String(), "a"))} {
		signed, err := validTargetsTemplate().ToSigned()
		require.NoError(t, err)

		signedTargets, err := TargetsFromSigned(signed, roleName)
		require.NoError(t, err)

		signed.Signatures[0] = Signature{KeyID: "key3", Method: "method3", Signature: []byte("world")}

		require.Equal(t, "key3", signed.Signatures[0].KeyID)
		require.Equal(t, "key1", signedTargets.Signatures[0].KeyID)
	}
}

// If the targets metadata contains delegations which are invalid, the targets metadata
// fails to validate and thus fails to convert into a SignedTargets
func TestTargetsFromSignedValidatesDelegations(t *testing.T) {
	for _, roleName := range []RoleName{CanonicalTargetsRole, RoleName(path.Join(CanonicalTargetsRole.String(), "a"))} {
		targets := validTargetsTemplate()
		delgRole, err := NewRole(RoleName(path.Join(roleName.String(), "b")), 1, []string{"key1"}, nil)
		require.NoError(t, err)
		targets.Signed.Delegations.Roles = []*Role{delgRole}

		// delegation has invalid threshold
		delgRole.Threshold = 0
		s, err := targets.ToSigned()
		require.NoError(t, err)
		_, err = TargetsFromSigned(s, roleName)
		require.Error(t, err)
		require.IsType(t, ErrInvalidMetadata{}, err)

		delgRole.Threshold = 1

		// Keys that aren't in the list of keys
		delgRole.KeyIDs = []string{"keys11"}
		s, err = targets.ToSigned()
		require.NoError(t, err)
		_, err = TargetsFromSigned(s, roleName)
		require.Error(t, err)
		require.IsType(t, ErrInvalidMetadata{}, err)

		delgRole.KeyIDs = []string{"keys1"}

		// not delegation role
		delgRole.Name = CanonicalRootRole
		s, err = targets.ToSigned()
		require.NoError(t, err)
		_, err = TargetsFromSigned(s, roleName)
		require.Error(t, err)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// more than one level deep
		delgRole.Name = RoleName(path.Join(roleName.String(), "x", "y"))
		s, err = targets.ToSigned()
		require.NoError(t, err)
		_, err = TargetsFromSigned(s, roleName)
		require.Error(t, err)
		require.IsType(t, ErrInvalidMetadata{}, err)

		// not in delegation hierarchy
		if IsDelegation(roleName) {
			delgRole.Name = RoleName(path.Join(CanonicalTargetsRole.String(), "z"))
			s, err := targets.ToSigned()
			require.NoError(t, err)
			_, err = TargetsFromSigned(s, roleName)
			require.Error(t, err)
			require.IsType(t, ErrInvalidMetadata{}, err)
		}
	}
}

// Type must be "Targets"
func TestTargetsFromSignedValidatesRoleType(t *testing.T) {
	for _, roleName := range []RoleName{CanonicalTargetsRole, RoleName(path.Join(CanonicalTargetsRole.String(), "a"))} {
		tg := validTargetsTemplate()

		for _, invalid := range []string{" Targets", CanonicalTargetsRole.String(), "TARGETS"} {
			tg.Signed.Type = invalid
			s, err := tg.ToSigned()
			require.NoError(t, err)
			_, err = TargetsFromSigned(s, roleName)
			require.IsType(t, ErrInvalidMetadata{}, err)
		}

		tg = validTargetsTemplate()
		tg.Signed.Type = "Targets"
		s, err := tg.ToSigned()
		require.NoError(t, err)
		sTargets, err := TargetsFromSigned(s, roleName)
		require.NoError(t, err)
		require.Equal(t, "Targets", sTargets.Signed.Type)
	}
}

// The rolename passed to TargetsFromSigned must be a valid targets role name
func TestTargetsFromSignedValidatesRoleName(t *testing.T) {
	for _, roleName := range []RoleName{"TARGETS", "root/a"} {
		tg := validTargetsTemplate()
		s, err := tg.ToSigned()
		require.NoError(t, err)

		_, err = TargetsFromSigned(s, roleName)
		require.IsType(t, ErrInvalidRole{}, err)
	}
}

// The version cannot be negative
func TestTargetsFromSignedValidatesVersion(t *testing.T) {
	tg := validTargetsTemplate()
	tg.Signed.Version = -1
	s, err := tg.ToSigned()
	require.NoError(t, err)
	_, err = TargetsFromSigned(s, "targets/a")
	require.IsType(t, ErrInvalidMetadata{}, err)

	tg.Signed.Version = 0
	s, err = tg.ToSigned()
	require.NoError(t, err)
	_, err = TargetsFromSigned(s, "targets/a")
	require.IsType(t, ErrInvalidMetadata{}, err)

	tg.Signed.Version = 1
	s, err = tg.ToSigned()
	require.NoError(t, err)
	_, err = TargetsFromSigned(s, "targets/a")
	require.NoError(t, err)
}
