//+build !pkcs11

package main

import (
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/notary"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/passphrase"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestImportKeysNoYubikey(t *testing.T) {
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
	bytes, err := memStore.Get(pubK.ID())
	require.NoError(t, err)
	b, _ := pem.Decode(bytes)
	b.Headers["path"] = "ankh"

	pubK, err = cs.Create(data.CanonicalTargetsRole, "morpork", data.ECDSAKey)
	require.NoError(t, err)
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

	fileStore, err := store.NewPrivateKeyFileStorage(tempBaseDir, notary.KeyExtension)
	require.NoError(t, err)
	bResult, err := fileStore.Get("ankh")
	require.NoError(t, err)
	cResult, err := fileStore.Get("morpork")
	require.NoError(t, err)

	block, rest := pem.Decode(bResult)
	require.Equal(t, b.Bytes, block.Bytes)
	require.Len(t, rest, 0)

	block, rest = pem.Decode(cResult)
	require.Equal(t, c.Bytes, block.Bytes)
	require.Len(t, rest, 0)
}

func TestExportImportKeysNoYubikey(t *testing.T) {
	setUp(t)
	exportTempDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(exportTempDir)
	tempfile, err := ioutil.TempFile("", "notary-test-import-")
	require.NoError(t, err)
	tempfile.Close()
	defer os.RemoveAll(tempfile.Name())
	exportCommander := &keyCommander{
		configGetter: func() (*viper.Viper, error) {
			v := viper.New()
			v.SetDefault("trust_dir", exportTempDir)
			return v, nil
		},
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
	}
	exportCommander.outFile = tempfile.Name()

	exportStore, err := store.NewPrivateKeyFileStorage(exportTempDir, notary.KeyExtension)
	require.NoError(t, err)
	ks := trustmanager.NewGenericKeyStore(exportStore, exportCommander.getRetriever())
	cs := cryptoservice.NewCryptoService(ks)

	pubK, err := cs.Create(data.CanonicalRootRole, "ankh", data.ECDSAKey)
	require.NoError(t, err)
	bID := pubK.ID()
	bOrigBytes, err := exportStore.Get(bID)
	require.NoError(t, err)
	bOrig, _ := pem.Decode(bOrigBytes)

	pubK, err = cs.Create(data.CanonicalTargetsRole, "morpork", data.ECDSAKey)
	require.NoError(t, err)
	cID := pubK.ID()
	cOrigBytes, err := exportStore.Get(cID)
	require.NoError(t, err)
	cOrig, _ := pem.Decode(cOrigBytes)

	exportCommander.exportKeys(&cobra.Command{}, nil)

	importTempDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err)
	defer os.RemoveAll(importTempDir)
	importCommander := &keyCommander{
		configGetter: func() (*viper.Viper, error) {
			v := viper.New()
			v.SetDefault("trust_dir", importTempDir)
			return v, nil
		},
		getRetriever: func() notary.PassRetriever { return passphrase.ConstantRetriever("pass") },
	}

	err = importCommander.importKeys(&cobra.Command{}, []string{tempfile.Name()})
	require.NoError(t, err)

	importStore, err := store.NewPrivateKeyFileStorage(importTempDir, notary.KeyExtension)
	require.NoError(t, err)
	bResult, err := importStore.Get(bID)
	require.NoError(t, err)
	cResult, err := importStore.Get(cID)
	require.NoError(t, err)

	block, rest := pem.Decode(bResult)
	require.Equal(t, bOrig.Bytes, block.Bytes)
	require.Len(t, rest, 0)

	block, rest = pem.Decode(cResult)
	require.Equal(t, cOrig.Bytes, block.Bytes)
	require.Len(t, rest, 0)
}
