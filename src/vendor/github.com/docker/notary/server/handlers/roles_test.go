package handlers

import (
	"testing"

	"github.com/docker/distribution/registry/api/errcode"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"github.com/docker/notary"
	"github.com/docker/notary/server/errors"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
)

func TestGetMaybeServerSignedNoCrypto(t *testing.T) {
	_, _, err := getMaybeServerSigned(context.Background(), nil, "", "")
	require.Error(t, err)
	require.IsType(t, errcode.Error{}, err)

	errc, ok := err.(errcode.Error)
	require.True(t, ok)
	require.Equal(t, errors.ErrNoCryptoService, errc.Code)
}

func TestGetMaybeServerSignedNoKey(t *testing.T) {
	crypto := signed.NewEd25519()
	store := storage.NewMemStorage()
	ctx := context.WithValue(context.Background(), notary.CtxKeyMetaStore, store)
	ctx = context.WithValue(ctx, notary.CtxKeyCryptoSvc, crypto)
	ctx = context.WithValue(ctx, notary.CtxKeyKeyAlgo, data.ED25519Key)

	_, _, err := getMaybeServerSigned(
		ctx,
		store,
		"gun",
		data.CanonicalTimestampRole,
	)
	require.Error(t, err)
	require.IsType(t, errcode.Error{}, err)

	errc, ok := err.(errcode.Error)
	require.True(t, ok)
	require.Equal(t, errors.ErrMetadataNotFound, errc.Code)
}
