package changelist

import (
	"testing"

	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/stretchr/testify/require"
)

func TestTUFDelegation(t *testing.T) {
	cs := signed.NewEd25519()
	key, err := cs.Create("targets/new_name", "gun", data.ED25519Key)
	require.NoError(t, err)
	kl := data.KeyList{key}
	td := TUFDelegation{
		NewName:      "targets/new_name",
		NewThreshold: 1,
		AddKeys:      kl,
		AddPaths:     []string{""},
	}

	r, err := td.ToNewRole("targets/old_name")
	require.NoError(t, err)
	require.Equal(t, td.NewName, r.Name)
	require.Len(t, r.KeyIDs, 1)
	require.Equal(t, kl[0].ID(), r.KeyIDs[0])
	require.Len(t, r.Paths, 1)
}
