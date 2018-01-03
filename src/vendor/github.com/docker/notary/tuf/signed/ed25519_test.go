package signed

import (
	"testing"

	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

// ListKeys only returns the keys for that role
func TestListKeys(t *testing.T) {
	c := NewEd25519()
	tskey, err := c.Create(data.CanonicalTimestampRole, "", data.ED25519Key)
	require.NoError(t, err)

	_, err = c.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)

	tsKeys := c.ListKeys(data.CanonicalTimestampRole)
	require.Len(t, tsKeys, 1)
	require.Equal(t, tskey.ID(), tsKeys[0])

	require.Len(t, c.ListKeys(data.CanonicalTargetsRole), 0)
}

// GetKey and GetPrivateKey only gets keys that we've added to this service
func TestGetKeys(t *testing.T) {
	c := NewEd25519()
	tskey, err := c.Create(data.CanonicalTimestampRole, "", data.ED25519Key)
	require.NoError(t, err)

	pubKey := c.GetKey(tskey.ID())
	require.NotNil(t, pubKey)
	require.Equal(t, tskey.Public(), pubKey.Public())
	require.Equal(t, tskey.Algorithm(), pubKey.Algorithm())
	require.Equal(t, tskey.ID(), pubKey.ID())

	privKey, role, err := c.GetPrivateKey(tskey.ID())
	require.NoError(t, err)
	require.Equal(t, data.CanonicalTimestampRole, role)
	require.Equal(t, tskey.Public(), privKey.Public())
	require.Equal(t, tskey.Algorithm(), privKey.Algorithm())
	require.Equal(t, tskey.ID(), privKey.ID())

	// if the key doesn't exist, GetKey returns nil and GetPrivateKey errors out
	randomKey := c.GetKey("someID")
	require.Nil(t, randomKey)
	_, _, err = c.GetPrivateKey("someID")
	require.Error(t, err)
}
