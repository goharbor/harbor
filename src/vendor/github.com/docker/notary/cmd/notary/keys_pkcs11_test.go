// +build pkcs11

package main

import (
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/docker/notary"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/passphrase"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/trustmanager/yubikey"
	"github.com/docker/notary/tuf/data"
)

func TestImportWithYubikey(t *testing.T) {
	if !yubikey.IsAccessible() {
		t.Skip("Must have Yubikey access.")
	}
	setUp(t)
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)
	input, err := ioutil.TempFile("", "notary-test-import-")
	require.NoError(t, err)
	defer os.RemoveAll(input.Name())
	k := &keyCommander{
		configGetter: func() (*viper.Viper, error) {
			v := viper.New()
			v.SetDefault("trust_dir", tempBaseDir)
			return v, nil
		},
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
	}

	memStore := store.NewMemoryStore(nil)
	ks := trustmanager.NewGenericKeyStore(memStore, k.getRetriever())
	cs := cryptoservice.NewCryptoService(ks)

	pubK, err := cs.Create(data.CanonicalRootRole, "ankh", data.ECDSAKey)
	require.NoError(t, err)
	bID := pubK.ID() // need to check presence in yubikey later
	bytes, err := memStore.Get(pubK.ID())
	require.NoError(t, err)
	b, _ := pem.Decode(bytes)
	b.Headers["path"] = "ankh"
	require.Equal(t, "root", b.Headers["role"])

	pubK, err = cs.Create(data.CanonicalTargetsRole, "morpork", data.ECDSAKey)
	require.NoError(t, err)
	cID := pubK.ID()
	bytes, err = memStore.Get(pubK.ID())
	require.NoError(t, err)
	c, _ := pem.Decode(bytes)
	c.Headers["path"] = "morpork"

	bBytes := pem.EncodeToMemory(b)
	cBytes := pem.EncodeToMemory(c)
	input.Write(bBytes)
	input.Write(cBytes)

	file := input.Name()
	err = input.Close() // close so import can open
	require.NoError(t, err)

	err = k.importKeys(&cobra.Command{}, []string{file})
	require.NoError(t, err)

	yks, err := yubikey.NewYubiStore(nil, k.getRetriever())
	require.NoError(t, err)
	_, _, err = yks.GetKey(bID)
	require.NoError(t, err)
	_, _, err = yks.GetKey(cID)
	require.Error(t, err) // c is non-root, should not be in yubikey

	fileStore, err := store.NewPrivateKeyFileStorage(tempBaseDir, notary.KeyExtension)
	require.NoError(t, err)
	_, err = fileStore.Get("ankh")
	require.Error(t, err) // b should only be in yubikey, not in filestore

	cResult, err := fileStore.Get("morpork")
	require.NoError(t, err)

	block, rest := pem.Decode(cResult)
	require.Equal(t, c.Bytes, block.Bytes)
	require.Len(t, rest, 0)
}

func TestGetImporters(t *testing.T) {
	if !yubikey.IsAccessible() {
		t.Skip("Must have Yubikey access.")
	}
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)
	importers, err := getImporters(tempBaseDir, passphrase.ConstantRetriever("pass"))
	require.NoError(t, err)
	require.Len(t, importers, 2)
}
