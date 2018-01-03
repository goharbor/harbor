// Actually start up a notary server and run through basic TUF and key
// interactions via the client.

// Note - if using Yubikey, retrieving pins/touch doesn't seem to work right
// when running in the midst of all tests.

package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	ctxu "github.com/docker/distribution/context"
	"github.com/docker/notary"
	"github.com/docker/notary/client"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/server"
	"github.com/docker/notary/server/storage"
	nstorage "github.com/docker/notary/storage"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

var testPassphrase = "passphrase"
var NewNotaryCommand func() *cobra.Command

// run a command and return the output as a string
func runCommand(t *testing.T, tempDir string, args ...string) (string, error) {
	b := new(bytes.Buffer)

	// Create an empty config file so we don't load the default on ~/.notary/config.json
	configFile := filepath.Join(tempDir, "config.json")

	cmd := NewNotaryCommand()
	cmd.SetArgs(append([]string{"-c", configFile, "-d", tempDir}, args...))
	cmd.SetOutput(b)
	retErr := cmd.Execute()
	output, err := ioutil.ReadAll(b)
	require.NoError(t, err)

	// Clean up state to mimic running a fresh command next time
	for _, command := range cmd.Commands() {
		command.ResetFlags()
	}

	return string(output), retErr
}

func setupServerHandler(metaStore storage.MetaStore) http.Handler {
	ctx := context.WithValue(context.Background(), notary.CtxKeyMetaStore, metaStore)

	ctx = context.WithValue(ctx, notary.CtxKeyKeyAlgo, data.ECDSAKey)

	// Eat the logs instead of spewing them out
	var b bytes.Buffer
	l := logrus.New()
	l.Out = &b
	ctx = ctxu.WithLogger(ctx, logrus.NewEntry(l))

	cryptoService := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(passphrase.ConstantRetriever("pass")))
	return server.RootHandler(ctx, nil, cryptoService, nil, nil, nil)
}

// makes a testing notary-server
func setupServer() *httptest.Server {
	return httptest.NewServer(setupServerHandler(storage.NewMemStorage()))
}

// Initializes a repo with existing key
func TestInitWithRootKey(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	tempFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// -- tests --

	// create encrypted root key
	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	// if the key has a root role, AddKey sets the gun to "" so we have done the same here
	encryptedPEMPrivKey, err := utils.EncryptPrivateKey(privKey, data.CanonicalRootRole, "", testPassphrase)
	require.NoError(t, err)
	encryptedPEMKeyFilename := filepath.Join(tempDir, "encrypted_key.key")
	err = ioutil.WriteFile(encryptedPEMKeyFilename, encryptedPEMPrivKey, 0644)
	require.NoError(t, err)

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun", "--rootkey", encryptedPEMKeyFilename)
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// check that the root key used for init is the one listed as root key
	output, err := runCommand(t, tempDir, "key", "list")
	require.NoError(t, err)
	require.Contains(t, output, data.PublicKeyFromPrivate(privKey).ID())

	// check error if file doesn't exist
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun2", "--rootkey", "bad_file")
	require.Error(t, err, "Init with nonexistent key file should error")

	// check error if file is invalid format
	badKeyFilename := filepath.Join(tempDir, "bad_key.key")
	nonPEMKey := []byte("thisisnotapemkey")
	err = ioutil.WriteFile(badKeyFilename, nonPEMKey, 0644)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun2", "--rootkey", badKeyFilename)
	require.Error(t, err, "Init with non-PEM key should error")

	// check error if unencrypted PEM used
	unencryptedPrivKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	unencryptedPEMPrivKey, err := utils.KeyToPEM(unencryptedPrivKey, data.CanonicalRootRole, "")
	require.NoError(t, err)
	unencryptedPEMKeyFilename := filepath.Join(tempDir, "unencrypted_key.key")
	err = ioutil.WriteFile(unencryptedPEMKeyFilename, unencryptedPEMPrivKey, 0644)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun2", "--rootkey", unencryptedPEMKeyFilename)
	require.Error(t, err, "Init with unencrypted PEM key should error")

	// check error if invalid password used
	// instead of using a new retriever, we create a new key with a different pass
	badPassPrivKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	// Blank gun name since it is a root key
	badPassPEMPrivKey, err := utils.EncryptPrivateKey(badPassPrivKey, data.CanonicalRootRole, "", "bad_pass")
	require.NoError(t, err)
	badPassPEMKeyFilename := filepath.Join(tempDir, "badpass_key.key")
	err = ioutil.WriteFile(badPassPEMKeyFilename, badPassPEMPrivKey, 0644)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun2", "--rootkey", badPassPEMKeyFilename)
	require.Error(t, err, "Init with wrong password should error")

	// check error if wrong role specified
	snapshotPrivKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	snapshotPEMPrivKey, err := utils.KeyToPEM(snapshotPrivKey, data.CanonicalSnapshotRole, "gun2")
	require.NoError(t, err)
	snapshotPEMKeyFilename := filepath.Join(tempDir, "snapshot_key.key")
	err = ioutil.WriteFile(snapshotPEMKeyFilename, snapshotPEMPrivKey, 0644)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun2", "--rootkey", snapshotPEMKeyFilename)
	require.Error(t, err, "Init with wrong role should error")
}

// Initializes a repo, adds a target, publishes the target, lists the target,
// verifies the target, and then removes the target.
func TestClientTUFInteraction(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	tempFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	var (
		output string
		target = "sdgkadga"
	)
	// -- tests --

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	// add a target
	_, err = runCommand(t, tempDir, "add", "gun", target, tempFile.Name())
	require.NoError(t, err)

	// check status - see target
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// check status - no targets
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.False(t, strings.Contains(string(output), target))

	// list repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target)

	// lookup target and repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "lookup", "gun", target)
	require.NoError(t, err)
	require.Contains(t, output, target)

	// verify repo - empty file
	_, err = runCommand(t, tempDir, "-s", server.URL, "verify", "gun", target)
	require.NoError(t, err)

	// remove target
	_, err = runCommand(t, tempDir, "remove", "gun", target)
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list repo - don't see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.False(t, strings.Contains(string(output), target))
}

func TestClientDeleteTUFInteraction(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	tempFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Setup certificate
	certFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	cert, _, _ := generateCertPrivKeyPair(t, "gun", data.ECDSAKey)
	_, err = certFile.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	defer os.Remove(certFile.Name())

	var (
		output string
		target = "helloIamanotarytarget"
	)
	// -- tests --

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	// add a target
	_, err = runCommand(t, tempDir, "add", "gun", target, tempFile.Name())
	require.NoError(t, err)

	// check status - see target
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.True(t, strings.Contains(output, target))

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(output), target))

	// add a delegation and publish
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", certFile.Name())
	require.NoError(t, err)
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - see role
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(output), "targets/delegation"))

	// Delete the repo metadata locally, so no need for server URL
	_, err = runCommand(t, tempDir, "delete", "gun")
	require.NoError(t, err)
	assertLocalMetadataForGun(t, tempDir, "gun", false)

	// list repo - see target still because remote data exists
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(output), target))

	// list delegations - see role because remote data still exists
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(output), "targets/delegation"))

	// Trying to delete the repo with the remote flag fails if it's given a badly formed URL
	_, err = runCommand(t, tempDir, "-s", "//invalidURLType", "delete", "gun", "--remote")
	require.Error(t, err)
	// since the connection fails to parse the URL before we can delete anything, local data should exist
	assertLocalMetadataForGun(t, tempDir, "gun", true)

	// Trying to delete the repo with the remote flag fails if it's given a well-formed URL that doesn't point to a server
	_, err = runCommand(t, tempDir, "-s", "https://invalid-server", "delete", "gun", "--remote")
	require.Error(t, err)
	require.IsType(t, nstorage.ErrOffline{}, err)
	// In this case, local notary metadata does not exist since local deletion operates first if we have a valid transport
	assertLocalMetadataForGun(t, tempDir, "gun", false)

	// Delete the repo remotely and locally, pointing to the correct server
	_, err = runCommand(t, tempDir, "-s", server.URL, "delete", "gun", "--remote")
	require.NoError(t, err)
	assertLocalMetadataForGun(t, tempDir, "gun", false)
	_, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.Error(t, err)
	require.IsType(t, client.ErrRepositoryNotExist{}, err)

	// Silent success on extraneous deletes
	_, err = runCommand(t, tempDir, "-s", server.URL, "delete", "gun", "--remote")
	require.NoError(t, err)
	assertLocalMetadataForGun(t, tempDir, "gun", false)
	_, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.Error(t, err)
	require.IsType(t, client.ErrRepositoryNotExist{}, err)

	// Now check that we can re-publish the same repo
	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	// add a target
	_, err = runCommand(t, tempDir, "add", "gun", target, tempFile.Name())
	require.NoError(t, err)

	// check status - see target
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.True(t, strings.Contains(output, target))

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.True(t, strings.Contains(string(output), target))
}

func assertLocalMetadataForGun(t *testing.T, configDir, gun string, shouldExist bool) {
	for _, role := range data.BaseRoles {
		fileInfo, err := os.Stat(filepath.Join(configDir, "tuf", gun, "metadata", role.String()+".json"))
		if shouldExist {
			require.NoError(t, err)
			require.NotNil(t, fileInfo)
		} else {
			require.Error(t, err)
			require.Nil(t, fileInfo)
		}
	}
}

// Initializes a repo, adds a target, publishes the target by hash, lists the target,
// verifies the target, and then removes the target.
func TestClientTUFAddByHashInteraction(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	targetData := []byte{'a', 'b', 'c'}
	target256Bytes := sha256.Sum256(targetData)
	targetSHA256Hex := hex.EncodeToString(target256Bytes[:])
	target512Bytes := sha512.Sum512(targetData)
	targetSha512Hex := hex.EncodeToString(target512Bytes[:])

	err := ioutil.WriteFile(filepath.Join(tempDir, "tempfile"), targetData, 0644)
	require.NoError(t, err)

	var (
		output  string
		target1 = "sdgkadga"
		target2 = "asdfasdf"
		target3 = "qwerty"
	)
	// -- tests --

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	// add a target just by sha256
	_, err = runCommand(t, tempDir, "addhash", "gun", target1, "3", "--sha256", targetSHA256Hex)
	require.NoError(t, err)

	// check status - see target
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target1)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// check status - no targets
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.False(t, strings.Contains(string(output), target1))

	// list repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target1)

	// lookup target and repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "lookup", "gun", target1)
	require.NoError(t, err)
	require.Contains(t, output, target1)

	// remove target
	_, err = runCommand(t, tempDir, "remove", "gun", target1)
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list repo - don't see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.False(t, strings.Contains(string(output), target1))

	// add a target just by sha512
	_, err = runCommand(t, tempDir, "addhash", "gun", target2, "3", "--sha512", targetSha512Hex)
	require.NoError(t, err)

	// check status - see target
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target2)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// check status - no targets
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.False(t, strings.Contains(string(output), target2))

	// list repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target2)

	// lookup target and repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "lookup", "gun", target2)
	require.NoError(t, err)
	require.Contains(t, output, target2)

	// remove target
	_, err = runCommand(t, tempDir, "remove", "gun", target2)
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// add a target by sha256 and sha512
	_, err = runCommand(t, tempDir, "addhash", "gun", target3, "3", "--sha256", targetSHA256Hex, "--sha512", targetSha512Hex)
	require.NoError(t, err)

	// check status - see target
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target3)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// check status - no targets
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.False(t, strings.Contains(string(output), target3))

	// list repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target3)

	// lookup target and repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "lookup", "gun", target3)
	require.NoError(t, err)
	require.Contains(t, output, target3)

	// remove target
	_, err = runCommand(t, tempDir, "remove", "gun", target3)
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)
}

// Initialize repo and test delegations commands by adding, listing, and removing delegations
func TestClientDelegationsInteraction(t *testing.T) {
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	// Setup certificate
	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	cert, _, keyID := generateCertPrivKeyPair(t, "gun", data.ECDSAKey)
	_, err = tempFile.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	var output string

	// -- tests --

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - none yet
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No delegations present in this repository.")

	// add new valid delegation with single new cert, and no path
	output, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", tempFile.Name())
	require.NoError(t, err)
	require.Contains(t, output, "Addition of delegation role")
	require.Contains(t, output, keyID)
	require.NotContains(t, output, "path")

	// check status - see delegation
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "Unpublished changes for gun")

	// list delegations - none yet because still unpublished
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No delegations present in this repository.")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// check status - no changelist
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No unpublished changes for gun")

	// list delegations - we should see our added delegation, with no paths
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "targets/delegation")
	require.Contains(t, output, keyID)
	require.NotContains(t, output, "\"\"")

	// add all paths to this delegation
	output, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", "--all-paths")
	require.NoError(t, err)
	require.Contains(t, output, "Addition of delegation role")
	require.Contains(t, output, "\"\"")
	require.Contains(t, output, "<all paths>")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see our added delegation, with no paths
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "targets/delegation")
	require.Contains(t, output, "\"\"")
	require.Contains(t, output, "<all paths>")

	// Setup another certificate
	tempFile2, err := ioutil.TempFile("", "pemfile2")
	require.NoError(t, err)

	cert2, _, keyID2 := generateCertPrivKeyPair(t, "gun", data.ECDSAKey)
	_, err = tempFile2.Write(utils.CertToPEM(cert2))
	require.NoError(t, err)
	tempFile2.Close()
	defer os.Remove(tempFile2.Name())

	// add to the delegation by specifying the same role, this time add a scoped path
	output, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", tempFile2.Name(), "--paths", "path")
	require.NoError(t, err)
	require.Contains(t, output, "Addition of delegation role")
	require.Contains(t, output, keyID2)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see two keys
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "path")
	require.Contains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// remove the delegation's first key
	output, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/delegation", keyID)
	require.NoError(t, err)
	require.Contains(t, output, "Removal of delegation role")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see the delegation but with only the second key
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// remove the delegation's second key
	output, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/delegation", keyID2)
	require.NoError(t, err)
	require.Contains(t, output, "Removal of delegation role")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see no delegations
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, keyID)
	require.NotContains(t, output, keyID2)

	// add delegation with multiple certs and multiple paths
	output, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", tempFile.Name(), tempFile2.Name(), "--paths", "path1,path2")
	require.NoError(t, err)
	require.Contains(t, output, "Addition of delegation role")
	require.Contains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see two keys
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "path1")
	require.Contains(t, output, "path2")
	require.Contains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// add delegation with multiple certs and multiple paths
	output, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", "--paths", "path3")
	require.NoError(t, err)
	require.Contains(t, output, "Addition of delegation role")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see two keys
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "path1")
	require.Contains(t, output, "path2")
	require.Contains(t, output, "path3")
	require.Contains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// just remove two paths from this delegation
	output, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/delegation", "--paths", "path2,path3")
	require.NoError(t, err)
	require.Contains(t, output, "Removal of delegation role")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see the same two keys, and only path1
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "path1")
	require.NotContains(t, output, "path2")
	require.NotContains(t, output, "path3")
	require.Contains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// remove the remaining path, should not remove the delegation entirely
	output, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/delegation", "--paths", "path1")
	require.NoError(t, err)
	require.Contains(t, output, "Removal of delegation role")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see the same two keys, and no paths
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, "path1")
	require.NotContains(t, output, "path2")
	require.NotContains(t, output, "path3")
	require.Contains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// Add a bunch of individual paths so we can test a delegation remove --all-paths
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", "--paths", "abcdef,123456")
	require.NoError(t, err)

	// Add more individual paths so we can test a delegation remove --all-paths
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", "--paths", "banana/split,apple/crumble/pie,orange.peel,kiwi")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see all of our paths
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "abcdef")
	require.Contains(t, output, "123456")
	require.Contains(t, output, "banana/split")
	require.Contains(t, output, "apple/crumble/pie")
	require.Contains(t, output, "orange.peel")
	require.Contains(t, output, "kiwi")

	// Try adding "", and check that adding it with other paths clears out the others
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", "--paths", "\"\",grapefruit,pomegranate")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see all of our old paths, and ""
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "abcdef")
	require.Contains(t, output, "123456")
	require.Contains(t, output, "banana/split")
	require.Contains(t, output, "apple/crumble/pie")
	require.Contains(t, output, "orange.peel")
	require.Contains(t, output, "kiwi")
	require.Contains(t, output, "\"\"")
	require.NotContains(t, output, "grapefruit")
	require.NotContains(t, output, "pomegranate")

	// Try removing just ""
	_, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/delegation", "--paths", "\"\"")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see all of our old paths without ""
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "abcdef")
	require.Contains(t, output, "123456")
	require.Contains(t, output, "banana/split")
	require.Contains(t, output, "apple/crumble/pie")
	require.Contains(t, output, "orange.peel")
	require.Contains(t, output, "kiwi")
	require.NotContains(t, output, "\"\"")

	// Remove --all-paths to clear out all paths from this delegation
	_, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/delegation", "--all-paths")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see all of our paths
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, "abcdef")
	require.NotContains(t, output, "123456")
	require.NotContains(t, output, "banana/split")
	require.NotContains(t, output, "apple/crumble/pie")
	require.NotContains(t, output, "orange.peel")
	require.NotContains(t, output, "kiwi")

	// Check that we ignore other --paths if we pass in --all-paths on an add
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", "--all-paths", "--paths", "grapefruit,pomegranate")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should only see "", and not the other paths specified
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "\"\"")
	require.NotContains(t, output, "grapefruit")
	require.NotContains(t, output, "pomegranate")

	// Add those extra paths we ignored to set up the next test
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/delegation", "--paths", "grapefruit,pomegranate")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// Check that we ignore other --paths if we pass in --all-paths on a remove
	_, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/delegation", "--all-paths", "--paths", "pomegranate")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see no paths
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, "\"\"")
	require.NotContains(t, output, "grapefruit")
	require.NotContains(t, output, "pomegranate")

	// remove by force to delete the delegation entirely
	output, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/delegation", "-y")
	require.NoError(t, err)
	require.Contains(t, output, "Forced removal (including all keys and paths) of delegation role")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see no delegations
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No delegations present in this repository.")
}

// Initialize repo and test publishing targets with delegation roles
func TestClientDelegationsPublishing(t *testing.T) {
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	// Setup certificate for delegation role
	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)

	cert, privKey, canonicalKeyID := generateCertPrivKeyPair(t, "gun", data.RSAKey)
	_, err = tempFile.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	privKeyBytesNoRole, err := utils.KeyToPEM(privKey, "", "")
	require.NoError(t, err)
	privKeyBytesWithRole, err := utils.KeyToPEM(privKey, "user", "")
	require.NoError(t, err)

	// Set up targets for publishing
	tempTargetFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempTargetFile.Close()
	defer os.Remove(tempTargetFile.Name())

	var target = "sdgkadga"

	var output string

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - none yet
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No delegations present in this repository.")

	// validate that we have all keys, including snapshot
	assertNumKeys(t, tempDir, 1, 2, true)

	// rotate the snapshot key to server
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", "snapshot", "-r")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// validate that we lost the snapshot signing key
	_, signingKeyIDs := assertNumKeys(t, tempDir, 1, 1, true)
	targetKeyID := signingKeyIDs[0]

	// add new valid delegation with single new cert
	output, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/releases", tempFile.Name(), "--paths", "\"\"")
	require.NoError(t, err)
	require.Contains(t, output, "Addition of delegation role")
	require.Contains(t, output, canonicalKeyID)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see our one delegation
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, "No delegations present in this repository.")

	// remove the targets key to demonstrate that delegates don't need this key
	keyDir := filepath.Join(tempDir, notary.PrivDir)
	require.NoError(t, os.Remove(filepath.Join(keyDir, targetKeyID+".key")))

	// Note that we need to use the canonical key ID, followed by the base of the role here
	// Since, for a delegation- the filename is the canonical key ID. We have no role header in the PEM
	err = ioutil.WriteFile(filepath.Join(keyDir, canonicalKeyID+".key"), privKeyBytesNoRole, 0700)
	require.NoError(t, err)

	// add a target using the delegation -- will only add to targets/releases
	_, err = runCommand(t, tempDir, "add", "gun", target, tempTargetFile.Name(), "--roles", "targets/releases")
	require.NoError(t, err)

	// list targets for targets/releases - we should see no targets until we publish
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "No targets")

	_, err = runCommand(t, tempDir, "-s", server.URL, "status", "gun")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list targets for targets/releases - we should see our target!
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "targets/releases")

	// remove the target for this role only
	_, err = runCommand(t, tempDir, "remove", "gun", target, "--roles", "targets/releases")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list targets for targets/releases - we should see no targets
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "No targets present")

	// Try adding a target with a different key style - private/tuf_keys/canonicalKeyID.key with "user" set as the "role" PEM header
	// First remove the old key and add the new style
	require.NoError(t, os.Remove(filepath.Join(keyDir, canonicalKeyID+".key")))
	err = ioutil.WriteFile(filepath.Join(keyDir, canonicalKeyID+".key"), privKeyBytesWithRole, 0700)
	require.NoError(t, err)

	// add a target using the delegation -- will only add to targets/releases
	_, err = runCommand(t, tempDir, "add", "gun", target, tempTargetFile.Name(), "--roles", "targets/releases")
	require.NoError(t, err)

	// list targets for targets/releases - we should see no targets until we publish
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "No targets")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list targets for targets/releases - we should see our target!
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "targets/releases")

	// remove the target for this role only
	_, err = runCommand(t, tempDir, "remove", "gun", target, "--roles", "targets/releases")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// add a target using the delegation -- will only add to targets/releases
	_, err = runCommand(t, tempDir, "add", "gun", target, tempTargetFile.Name(), "--roles", "targets/releases")
	require.NoError(t, err)

	// list targets for targets/releases - we should see no targets until we publish
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "No targets")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list targets for targets/releases - we should see our target!
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "targets/releases")

	// Setup another certificate
	tempFile2, err := ioutil.TempFile("", "pemfile2")
	require.NoError(t, err)
	cert2, _, keyID2 := generateCertPrivKeyPair(t, "gun", data.RSAKey)
	_, err = tempFile2.Write(utils.CertToPEM(cert2))
	require.NoError(t, err)
	tempFile2.Close()
	defer os.Remove(tempFile2.Name())

	// add a nested delegation under this releases role
	output, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/releases/nested", tempFile2.Name(), "--paths", "nested/path")
	require.NoError(t, err)
	require.Contains(t, output, "Addition of delegation role")
	require.Contains(t, output, keyID2)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see two roles
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "targets/releases")
	require.Contains(t, output, "targets/releases/nested")
	require.Contains(t, output, canonicalKeyID)
	require.Contains(t, output, keyID2)
	require.Contains(t, output, "nested/path")
	require.Contains(t, output, "\"\"")
	require.Contains(t, output, "<all paths>")

	// remove by force to delete the nested delegation entirely
	output, err = runCommand(t, tempDir, "delegation", "remove", "gun", "targets/releases/nested", "-y")
	require.NoError(t, err)
	require.Contains(t, output, "Forced removal (including all keys and paths) of delegation role")

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

}

// Splits a string into lines, and returns any lines that are not empty (
// striped of whitespace)
func splitLines(chunk string) []string {
	splitted := strings.Split(strings.TrimSpace(chunk), "\n")
	var results []string

	for _, line := range splitted {
		line := strings.TrimSpace(line)
		if line != "" {
			results = append(results, line)
		}
	}

	return results
}

// List keys, parses the output, and returns the unique key IDs as an array
// of root key IDs and an array of signing key IDs.  Output expected looks like:
//     ROLE      GUN          KEY ID                   LOCATION
// ----------------------------------------------------------------
//   root               8bd63a896398b558ac...   file (.../private)
//   snapshot   repo    e9e9425cd9a85fc7a5...   file (.../private)
//   targets    repo    f5b84e2d92708c5acb...   file (.../private)
func getUniqueKeys(t *testing.T, tempDir string) ([]string, []string) {
	output, err := runCommand(t, tempDir, "key", "list")
	require.NoError(t, err)
	lines := splitLines(output)
	if len(lines) == 1 && lines[0] == "No signing keys found." {
		return []string{}, []string{}
	}
	if len(lines) < 3 { // 2 lines of header, at least 1 line with keys
		t.Logf("This output is not what is expected by the test:\n%s", output)
	}

	var (
		rootMap    = make(map[string]bool)
		nonrootMap = make(map[string]bool)
		root       []string
		nonroot    []string
	)
	// first two lines are header
	for _, line := range lines[2:] {
		parts := strings.Fields(line)
		var (
			placeToGo map[string]bool
			keyID     string
		)
		if strings.TrimSpace(parts[0]) == data.CanonicalRootRole.String() {
			// no gun, so there are only 3 fields
			placeToGo, keyID = rootMap, parts[1]
		} else {
			// gun comes between role and key ID
			if len(parts) == 3 {
				// gun is empty as this may be a delegation key
				placeToGo, keyID = nonrootMap, parts[1]
			} else {
				placeToGo, keyID = nonrootMap, parts[2]
			}

		}
		// keys are 32-chars long (32 byte shasum, hex-encoded)
		require.Len(t, keyID, 64)
		placeToGo[keyID] = true
	}
	for k := range rootMap {
		root = append(root, k)
	}
	for k := range nonrootMap {
		nonroot = append(nonroot, k)
	}

	return root, nonroot
}

// List keys, parses the output, and asserts something about the number of root
// keys and number of signing keys, as well as returning them.
func assertNumKeys(t *testing.T, tempDir string, numRoot, numSigning int,
	rootOnDisk bool) ([]string, []string) {

	root, signing := getUniqueKeys(t, tempDir)
	require.Len(t, root, numRoot)
	require.Len(t, signing, numSigning)
	for _, rootKeyID := range root {
		_, err := os.Stat(filepath.Join(
			tempDir, notary.PrivDir, rootKeyID+".key"))
		// os.IsExist checks to see if the error is because a file already
		// exists, and hence it isn't actually the right function to use here
		require.Equal(t, rootOnDisk, !os.IsNotExist(err))

		// this function is declared is in the build-tagged setup files
		verifyRootKeyOnHardware(t, rootKeyID)
	}
	return root, signing
}

// Adds the given target to the gun, publishes it, and lists it to ensure that
// it appears.  Returns the listing output.
func assertSuccessfullyPublish(
	t *testing.T, tempDir, url, gun, target, fname string) string {

	_, err := runCommand(t, tempDir, "add", gun, target, fname)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", url, "publish", gun)
	require.NoError(t, err)

	output, err := runCommand(t, tempDir, "-s", url, "list", gun)
	require.NoError(t, err)
	require.Contains(t, output, target)

	return output
}

// Tests root key generation and key rotation
func TestClientKeyGenerationRotation(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	tempfiles := make([]string, 2)
	for i := 0; i < 2; i++ {
		tempFile, err := ioutil.TempFile("", "targetfile")
		require.NoError(t, err)
		tempFile.Close()
		tempfiles[i] = tempFile.Name()
		defer os.Remove(tempFile.Name())
	}

	server := setupServer()
	defer server.Close()

	var target = "sdgkadga"

	// -- tests --

	// starts out with no keys
	assertNumKeys(t, tempDir, 0, 0, true)

	// generate root key produces a single root key and no other keys
	_, err := runCommand(t, tempDir, "key", "generate", data.ECDSAKey)
	require.NoError(t, err)
	assertNumKeys(t, tempDir, 1, 0, true)

	// initialize a repo, should have signing keys and no new root key
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)
	origRoot, origSign := assertNumKeys(t, tempDir, 1, 2, true)

	// publish using the original keys
	assertSuccessfullyPublish(t, tempDir, server.URL, "gun", target, tempfiles[0])

	// rotate the signing keys
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalSnapshotRole.String())
	require.NoError(t, err)
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalTargetsRole.String())
	require.NoError(t, err)
	root, sign := assertNumKeys(t, tempDir, 1, 2, true)
	require.Equal(t, origRoot[0], root[0])

	// just do a cursory rotation check that the keys aren't equal anymore
	for _, origKey := range origSign {
		for _, key := range sign {
			require.NotEqual(
				t, key, origKey, "One of the signing keys was not removed")
		}
	}

	// publish using the new keys
	output := assertSuccessfullyPublish(
		t, tempDir, server.URL, "gun", target+"2", tempfiles[1])
	// assert that the previous target is still there
	require.True(t, strings.Contains(string(output), target))

	// rotate the snapshot and timestamp keys on the server, multiple times
	for i := 0; i < 10; i++ {
		_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalSnapshotRole.String(), "-r")
		require.NoError(t, err)
		_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalTimestampRole.String(), "-r")
		require.NoError(t, err)
	}
}

// Tests key rotation
func TestKeyRotation(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	tempfiles := make([]string, 2)
	for i := 0; i < 2; i++ {
		tempFile, err := ioutil.TempFile("", "targetfile")
		require.NoError(t, err)
		tempFile.Close()
		tempfiles[i] = tempFile.Name()
		defer os.Remove(tempFile.Name())
	}

	server := setupServer()
	defer server.Close()

	var target = "sdgkadga"

	// -- tests --

	// starts out with no keys
	assertNumKeys(t, tempDir, 0, 0, true)

	// generate root key produces a single root key and no other keys
	_, err := runCommand(t, tempDir, "key", "generate", data.ECDSAKey)
	require.NoError(t, err)
	assertNumKeys(t, tempDir, 1, 0, true)

	// initialize a repo, should have signing keys and no new root key
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)
	assertNumKeys(t, tempDir, 1, 2, true)

	// publish using the original keys
	assertSuccessfullyPublish(t, tempDir, server.URL, "gun", target, tempfiles[0])

	// invalid keys
	badKeyFile, err := ioutil.TempFile("", "badKey")
	require.NoError(t, err)
	defer os.Remove(badKeyFile.Name())
	_, err = badKeyFile.Write([]byte{0, 0, 0, 0})
	require.NoError(t, err)
	badKeyFile.Close()

	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalRootRole.String(), "--key", "123")
	require.Error(t, err)
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalRootRole.String(), "--key", badKeyFile.Name())
	require.Error(t, err)

	// create encrypted root keys
	rootPrivKey1, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	encryptedPEMPrivKey1, err := utils.EncryptPrivateKey(rootPrivKey1, data.CanonicalRootRole, "", testPassphrase)
	require.NoError(t, err)
	encryptedPEMKeyFilename1 := filepath.Join(tempDir, "encrypted_key.key")
	err = ioutil.WriteFile(encryptedPEMKeyFilename1, encryptedPEMPrivKey1, 0644)
	require.NoError(t, err)

	rootPrivKey2, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	encryptedPEMPrivKey2, err := utils.EncryptPrivateKey(rootPrivKey2, data.CanonicalRootRole, "", testPassphrase)
	require.NoError(t, err)
	encryptedPEMKeyFilename2 := filepath.Join(tempDir, "encrypted_key2.key")
	err = ioutil.WriteFile(encryptedPEMKeyFilename2, encryptedPEMPrivKey2, 0644)
	require.NoError(t, err)

	// rotate the root key
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalRootRole.String(), "--key", encryptedPEMKeyFilename1, "--key", encryptedPEMKeyFilename2)
	require.NoError(t, err)
	// 3 root keys - 1 prev, 1 new
	assertNumKeys(t, tempDir, 3, 2, true)

	// rotate the root key again
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalRootRole.String())
	require.NoError(t, err)
	// 3 root keys, 2 prev, 1 new
	assertNumKeys(t, tempDir, 3, 2, true)

	// publish using the new keys
	output := assertSuccessfullyPublish(
		t, tempDir, server.URL, "gun", target+"2", tempfiles[1])
	// assert that the previous target is still there
	require.True(t, strings.Contains(string(output), target))
}

// Tests rotating non-root keys
func TestKeyRotationNonRoot(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	tempfiles := make([]string, 2)
	for i := 0; i < 2; i++ {
		tempFile, err := ioutil.TempFile("", "targetfile")
		require.NoError(t, err)
		tempFile.Close()
		tempfiles[i] = tempFile.Name()
		defer os.Remove(tempFile.Name())
	}

	server := setupServer()
	defer server.Close()

	var target = "sdgkadgad"

	// -- tests --

	// starts out with no keys
	assertNumKeys(t, tempDir, 0, 0, true)

	// generate root key produces a single root key and no other keys
	_, err := runCommand(t, tempDir, "key", "generate", data.ECDSAKey)
	require.NoError(t, err)
	assertNumKeys(t, tempDir, 1, 0, true)

	// initialize a repo, should have signing keys and no new root key
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)
	assertNumKeys(t, tempDir, 1, 2, true)

	// publish using the original keys
	assertSuccessfullyPublish(t, tempDir, server.URL, "gun", target, tempfiles[0])

	// create new target keys
	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	defer os.Remove(tempFile.Name())

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err := utils.EncryptPrivateKey(privKey, data.CanonicalTargetsRole, "", testPassphrase)
	require.NoError(t, err)

	nBytes, err := tempFile.Write(pemBytes)
	require.NoError(t, err)
	tempFile.Close()
	require.Equal(t, len(pemBytes), nBytes)

	tempFile2, err := ioutil.TempFile("", "pemfile2")
	require.NoError(t, err)
	defer os.Remove(tempFile2.Name())

	privKey2, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes2, err := utils.KeyToPEM(privKey2, data.CanonicalTargetsRole, "")
	require.NoError(t, err)

	nBytes2, err := tempFile2.Write(pemBytes2)
	require.NoError(t, err)
	tempFile2.Close()
	require.Equal(t, len(pemBytes2), nBytes2)

	// rotate the targets key
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalTargetsRole.String(), "--key", tempFile.Name(), "--key", tempFile2.Name())
	require.NoError(t, err)

	// publish using the new keys
	output := assertSuccessfullyPublish(
		t, tempDir, server.URL, "gun", target+"2", tempfiles[1])
	// assert that the previous target is still there
	require.True(t, strings.Contains(string(output), target))

	// rotate to nonexistant key
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalTargetsRole.String(), "--key", "nope.pem")
	require.Error(t, err)
}

// Tests default root key generation
func TestDefaultRootKeyGeneration(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	// -- tests --

	// starts out with no keys
	assertNumKeys(t, tempDir, 0, 0, true)

	// generate root key with no algorithm produces a single ECDSA root key and no other keys
	_, err := runCommand(t, tempDir, "key", "generate")
	require.NoError(t, err)
	assertNumKeys(t, tempDir, 1, 0, true)
}

// Tests the interaction with the verbose and log-level flags
func TestLogLevelFlags(t *testing.T) {
	// Test default to fatal
	n := notaryCommander{}
	n.setVerbosityLevel()
	require.Equal(t, "fatal", logrus.GetLevel().String())

	// Test that verbose (-v) sets to error
	n.verbose = true
	n.setVerbosityLevel()
	require.Equal(t, "error", logrus.GetLevel().String())

	// Test that debug (-D) sets to debug
	n.debug = true
	n.setVerbosityLevel()
	require.Equal(t, "debug", logrus.GetLevel().String())

	// Test that unsetting verboseError still uses verboseDebug
	n.verbose = false
	n.setVerbosityLevel()
	require.Equal(t, "debug", logrus.GetLevel().String())
}

func TestClientKeyPassphraseChange(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	target := "sdgkadga"
	tempFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// -- tests --
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun1")
	require.NoError(t, err)

	// we should have three keys stored locally in total: root, targets, snapshot
	rootIDs, signingIDs := assertNumKeys(t, tempDir, 1, 2, true)
	for _, keyID := range signingIDs {
		// try changing the private key passphrase
		_, err = runCommand(t, tempDir, "-s", server.URL, "key", "passwd", keyID)
		require.NoError(t, err)

		// assert that the signing keys (number and IDs) didn't change
		_, signingIDs = assertNumKeys(t, tempDir, 1, 2, true)
		require.Contains(t, signingIDs, keyID)

		// make sure we can still publish with this signing key
		assertSuccessfullyPublish(t, tempDir, server.URL, "gun1", target, tempFile.Name())
	}

	// only one rootID, try changing the private key passphrase
	rootID := rootIDs[0]
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "passwd", rootID)
	require.NoError(t, err)

	// make sure we can init a new repo with this key
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun2")
	require.NoError(t, err)

	// assert that the root key ID didn't change
	rootIDs, _ = assertNumKeys(t, tempDir, 1, 4, true)
	require.Equal(t, rootID, rootIDs[0])
}

func tempDirWithConfig(t *testing.T, config string) string {
	tempDir, err := ioutil.TempDir("", "repo")
	require.NoError(t, err)
	err = ioutil.WriteFile(filepath.Join(tempDir, "config.json"), []byte(config), 0644)
	require.NoError(t, err)
	return tempDir
}

func TestMain(m *testing.M) {
	if testing.Short() {
		// skip
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestPurgeSingleKey(t *testing.T) {
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	// Setup certificates
	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)

	cert, _, keyID := generateCertPrivKeyPair(t, "gun", data.ECDSAKey)
	_, err = tempFile.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Setup another certificate
	tempFile2, err := ioutil.TempFile("", "pemfile2")
	require.NoError(t, err)

	cert2, _, keyID2 := generateCertPrivKeyPair(t, "gun", data.ECDSAKey)
	_, err = tempFile2.Write(utils.CertToPEM(cert2))
	require.NoError(t, err)
	tempFile2.Close()
	defer os.Remove(tempFile2.Name())

	delgName := "targets/delegation1"
	delgName2 := "targets/delegation2"

	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// Add two delegations with two keys
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", delgName, tempFile.Name(), tempFile2.Name(), "--all-paths")
	require.NoError(t, err)
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", delgName2, tempFile.Name(), tempFile2.Name(), "--all-paths")
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	out, err := runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, out, delgName)
	require.Contains(t, out, delgName2)
	require.Contains(t, out, keyID)
	require.Contains(t, out, keyID2)

	// auto-publish doesn't error because purge only updates the roles we have signing keys for
	_, err = runCommand(t, tempDir, "delegation", "purge", "-s", server.URL, "-p", "gun", "--key", keyID)
	require.NoError(t, err)

	// check the delegations weren't removed, and that the key we purged isn't present
	out, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, out, keyID)
	require.Contains(t, out, delgName)
	require.Contains(t, out, delgName2)
	require.Contains(t, out, keyID2)
}

// Initialize repo and test witnessing. The following steps are performed:
//   1. init a repo
//   2. add a delegation with a key and --all-paths
//   3. add a target to the delegation
//   4. list targets and ensure it really is in the delegation
//   5  witness the valid delegation, make sure everything is successful
//   6. add a new (different) key to the delegation
//   7. remove the key from the delegation
//   8. list targets and ensure the target is no longer visible
//   9. witness the delegation
//  10. list targets and ensure target is visible again
//  11. witness an invalid role and check for error on publish
//  12. check non-targets base roles all fail
//  13. test auto-publish functionality
//  14. remove all keys from the delegation and publish
//  15. witnessing the delegation should now fail
func TestWitness(t *testing.T) {
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	// Setup certificates
	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)

	cert, privKey, keyID := generateCertPrivKeyPair(t, "gun", data.ECDSAKey)
	_, err = tempFile.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Setup another certificate
	tempFile2, err := ioutil.TempFile("", "pemfile2")
	require.NoError(t, err)

	cert2, privKey2, keyID2 := generateCertPrivKeyPair(t, "gun", data.ECDSAKey)
	_, err = tempFile2.Write(utils.CertToPEM(cert2))
	require.NoError(t, err)
	tempFile2.Close()
	defer os.Remove(tempFile2.Name())

	delgName := "targets/delegation"
	targetName := "test_target"
	targetHash := "9d9e890af64dd0f44b8a1538ff5fa0511cc31bf1ab89f3a3522a9a581a70fad8" // sha256 of README.md at time of writing test

	keyStore, err := trustmanager.NewKeyFileStore(tempDir, passphrase.ConstantRetriever(testPassphrase))
	require.NoError(t, err)
	err = keyStore.AddKey(
		trustmanager.KeyInfo{
			Gun:  "gun",
			Role: data.RoleName(delgName),
		},
		privKey,
	)
	require.NoError(t, err)

	// 1. init a repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// 2. add a delegation with a key and --all-paths
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", delgName, tempFile.Name(), "--all-paths")
	require.NoError(t, err)

	// 3. add a target to the delegation
	_, err = runCommand(t, tempDir, "addhash", "gun", targetName, "100", "--sha256", targetHash, "-r", delgName)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// 4. list targets and ensure it really is in the delegation
	output, err := runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, targetName)
	require.Contains(t, output, targetHash)

	// 5. witness the valid delegation, make sure everything is successful
	_, err = runCommand(t, tempDir, "witness", "gun", delgName)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, targetName)
	require.Contains(t, output, targetHash)

	// 6. add a new (different) key to the delegation
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", delgName, tempFile2.Name(), "--all-paths")
	require.NoError(t, err)

	// 7. remove the key from the delegation
	_, err = runCommand(t, tempDir, "delegation", "remove", "gun", delgName, keyID)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// 8. list targets and ensure the target is no longer visible
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, targetName)
	require.NotContains(t, output, targetHash)

	err = keyStore.AddKey(
		trustmanager.KeyInfo{
			Gun:  "gun",
			Role: data.RoleName(delgName),
		},
		privKey2,
	)
	require.NoError(t, err)

	// 9. witness the delegation
	_, err = runCommand(t, tempDir, "witness", "gun", delgName)
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// 10. list targets and ensure target is visible again
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, targetName)
	require.Contains(t, output, targetHash)

	// 11. witness an invalid role and check for error on publish
	_, err = runCommand(t, tempDir, "witness", "gun", "targets/made/up")
	require.NoError(t, err)

	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.Error(t, err)

	// 12. check non-targets base roles all fail
	for _, role := range []string{data.CanonicalRootRole.String(), data.CanonicalSnapshotRole.String(), data.CanonicalTimestampRole.String()} {
		// clear any pending changes to ensure errors are only related to the specific role we're trying to witness
		_, err = runCommand(t, tempDir, "reset", "gun", "--all")
		require.NoError(t, err)

		_, err = runCommand(t, tempDir, "witness", "gun", role)
		require.NoError(t, err)

		_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
		require.Error(t, err)
	}

	// 13. test auto-publish functionality (just for witness)

	// purge the old staged witness
	_, err = runCommand(t, tempDir, "reset", "gun", "--all")
	require.NoError(t, err)

	// remove key2 and add back key1
	_, err = runCommand(t, tempDir, "delegation", "add", "gun", delgName, tempFile.Name(), "--all-paths")
	require.NoError(t, err)
	_, err = runCommand(t, tempDir, "delegation", "remove", "gun", delgName, keyID2)
	require.NoError(t, err)
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// the role now won't show its target because it's invalid
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, targetName)
	require.NotContains(t, output, targetHash)

	// auto-publish with witness, check that the target is back
	_, err = runCommand(t, tempDir, "-s", server.URL, "witness", "-p", "gun", delgName)
	require.NoError(t, err)
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, targetName)
	require.Contains(t, output, targetHash)

	_, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "remove", "-p", "gun", delgName, keyID, keyID2)
	require.NoError(t, err)
	_, err = runCommand(t, tempDir, "-s", server.URL, "witness", "-p", "gun", delgName)
	require.Error(t, err)
	require.Contains(t, err.Error(), "role does not specify enough valid signing keys to meet its required threshold")
}

func generateCertPrivKeyPair(t *testing.T, gun, keyAlgorithm string) (*x509.Certificate, data.PrivateKey, string) {
	// Setup certificate
	var privKey data.PrivateKey
	var err error
	switch keyAlgorithm {
	case data.ECDSAKey:
		privKey, err = utils.GenerateECDSAKey(rand.Reader)
	case data.RSAKey:
		privKey, err = utils.GenerateRSAKey(rand.Reader, 4096)
	default:
		err = fmt.Errorf("invalid key algorithm provided: %s", keyAlgorithm)
	}
	require.NoError(t, err)
	startTime := time.Now()
	endTime := startTime.AddDate(10, 0, 0)
	cert, err := cryptoservice.GenerateCertificate(privKey, data.GUN(gun), startTime, endTime)
	require.NoError(t, err)
	parsedPubKey, _ := utils.ParsePEMPublicKey(utils.CertToPEM(cert))
	keyID, err := utils.CanonicalKeyID(parsedPubKey)
	require.NoError(t, err)
	return cert, privKey, keyID
}

func TestClientTUFInitWithAutoPublish(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	tempFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	var (
		gun          = "MistsOfPandaria"
		gunNoPublish = "Legion"

		// This might be changed via the implementation, please be careful.
		emptyList = "\nNo targets present in this repository.\n\n"
	)
	// -- tests --

	// init repo with auto publish being enabled but with a malformed URL.
	_, err = runCommand(t, tempDir, "-s", "For the Horde!", "init", "-p", gun)
	require.Error(t, err, "Trust server url has to be in the form of http(s)://URL:PORT.")
	// init repo with auto publish being enabled but with an unaccessible URL.
	_, err = runCommand(t, tempDir, "-s", "https://notary-server-on-the-moon:12306", "init", "-p", gun)
	require.NotNil(t, err)
	require.Equal(t, err, nstorage.ErrOffline{})

	// init repo with auto publish being enabled
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "-p", gun)
	require.NoError(t, err)
	// list repo - expect empty list
	output, err := runCommand(t, tempDir, "-s", server.URL, "list", gun)
	require.NoError(t, err)
	require.Equal(t, output, emptyList)

	// init repo without auto publish being enabled
	//
	// Use this test to guarantee that we won't break the normal init process.
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", gunNoPublish)
	require.NoError(t, err)
	// list repo - expect error
	_, err = runCommand(t, tempDir, "-s", server.URL, "list", gunNoPublish)
	require.NotNil(t, err)
	require.IsType(t, client.ErrRepositoryNotExist{}, err)
}

func TestClientTUFAddWithAutoPublish(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	tempFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	var (
		target          = "ShangXi"
		target2         = "ChenStormstout"
		targetNoPublish = "Shen-zinSu"
		gun             = "MistsOfPandaria"
	)
	// -- tests --

	// init repo with auto publish being enabled
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "-p", gun)
	require.NoError(t, err)

	// add a target with auto publish being enabled, but without the server URL
	_, err = runCommand(t, tempDir, "add", "-p", gun, target, tempFile.Name())
	require.NotNil(t, err)
	require.Equal(t, err, nstorage.ErrOffline{})
	// check status, since we only fail the auto publishment in the previous step,
	// the change should still exists.
	output, err := runCommand(t, tempDir, "status", gun)
	require.NoError(t, err)
	require.Contains(t, output, target)

	// add a target with auto publish being enabled but with a malformed URL.
	_, err = runCommand(t, tempDir, "-s", "For the Horde!", "add", "-p", gun, target, tempFile.Name())
	require.Error(t, err, "Trust server url has to be in the form of http(s)://URL:PORT.")
	// add a target with auto publish being enabled but with an unaccessible URL.
	_, err = runCommand(t, tempDir, "-s", "https://notary-server-on-the-moon:12306", "add", "-p", gun, target, tempFile.Name())
	require.NotNil(t, err)
	require.Equal(t, err, nstorage.ErrOffline{})

	// add a target with auto publish being enabled, and with the server URL
	_, err = runCommand(t, tempDir, "-s", server.URL, "add", "-p", gun, target2, tempFile.Name())
	require.NoError(t, err)
	// list repo, since the auto publish flag will try to publish all the staged changes,
	// so the target and target2 should be in the list.
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", gun)
	require.NoError(t, err)
	require.Contains(t, output, target)
	require.Contains(t, output, target2)

	// add a target without auto publish being enabled
	//
	// Use this test to guarantee that we won't break the normal add process.
	_, err = runCommand(t, tempDir, "add", gun, targetNoPublish, tempFile.Name())
	require.NoError(t, err)
	// check status - expect the targetNoPublish
	output, err = runCommand(t, tempDir, "status", gun)
	require.NoError(t, err)
	require.Contains(t, output, targetNoPublish)
	// list repo - expect only the target, not the targetNoPublish
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", gun)
	require.NoError(t, err)
	require.Contains(t, output, target)
	require.False(t, strings.Contains(output, targetNoPublish))
}

func TestClientTUFRemoveWithAutoPublish(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	tempFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	var (
		target              = "ShangXi"
		targetWillBeRemoved = "Shen-zinSu"
		gun                 = "MistsOfPandaria"
	)
	// -- tests --

	// init repo with auto publish being enabled
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "-p", gun)
	require.NoError(t, err)
	// add a target with auto publish being enabled
	_, err = runCommand(t, tempDir, "add", "-s", server.URL, "-p", gun, target, tempFile.Name())
	require.NoError(t, err)
	_, err = runCommand(t, tempDir, "add", "-s", server.URL, "-p", gun, targetWillBeRemoved, tempFile.Name())
	require.NoError(t, err)
	// remove a target with auto publish being enabled
	_, err = runCommand(t, tempDir, "remove", "-s", server.URL, "-p", gun, targetWillBeRemoved, tempFile.Name())
	require.NoError(t, err)
	// list repo - expect target
	output, err := runCommand(t, tempDir, "-s", server.URL, "list", gun)
	require.NoError(t, err)
	require.Contains(t, output, target)
	require.False(t, strings.Contains(output, targetWillBeRemoved))

	// remove a target without auto publish being enabled
	//
	// Use this test to guarantee that we won't break the normal remove process.
	_, err = runCommand(t, tempDir, "add", "-s", server.URL, "-p", gun, targetWillBeRemoved, tempFile.Name())
	require.NoError(t, err)
	// remove the targetWillBeRemoved without auto publish being enabled
	_, err = runCommand(t, tempDir, "remove", gun, targetWillBeRemoved, tempFile.Name())
	require.NoError(t, err)
	// check status - expect the targetWillBeRemoved
	output, err = runCommand(t, tempDir, "status", gun)
	require.NoError(t, err)
	require.Contains(t, output, targetWillBeRemoved)
	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", gun)
	require.NoError(t, err)
	// list repo - expect only the target, not the targetWillBeRemoved
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", gun)
	require.NoError(t, err)
	require.Contains(t, output, target)
	require.False(t, strings.Contains(output, targetWillBeRemoved))
}

func TestClientDelegationAddWithAutoPublish(t *testing.T) {
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	// Setup certificate
	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	startTime := time.Now()
	endTime := startTime.AddDate(10, 0, 0)
	cert, err := cryptoservice.GenerateCertificate(privKey, "gun", startTime, endTime)
	require.NoError(t, err)

	_, err = tempFile.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	rawPubBytes, _ := ioutil.ReadFile(tempFile.Name())
	parsedPubKey, _ := utils.ParsePEMPublicKey(rawPubBytes)
	keyID, err := utils.CanonicalKeyID(parsedPubKey)
	require.NoError(t, err)

	var output string

	// -- tests --

	// init and publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun", "-p")
	require.NoError(t, err)

	// list delegations - none yet
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No delegations present in this repository.")

	// add new valid delegation with single new cert, and no path
	_, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "add", "-p", "gun", "targets/delegation", tempFile.Name())
	require.NoError(t, err)

	// check status - no changelist
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No unpublished changes for gun")

	// list delegations - we should see our added delegation, with no paths
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "targets/delegation")
	require.Contains(t, output, keyID)
}

func TestClientDelegationRemoveWithAutoPublish(t *testing.T) {
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	// Setup certificate
	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	startTime := time.Now()
	endTime := startTime.AddDate(10, 0, 0)
	cert, err := cryptoservice.GenerateCertificate(privKey, "gun", startTime, endTime)
	require.NoError(t, err)

	_, err = tempFile.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	rawPubBytes, _ := ioutil.ReadFile(tempFile.Name())
	parsedPubKey, _ := utils.ParsePEMPublicKey(rawPubBytes)
	keyID, err := utils.CanonicalKeyID(parsedPubKey)
	require.NoError(t, err)

	var output string

	// -- tests --

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun", "-p")
	require.NoError(t, err)

	// add new valid delegation with single new cert, and no path
	_, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "add", "-p", "gun", "targets/delegation", tempFile.Name())
	require.NoError(t, err)

	// list delegations - we should see our added delegation, with no paths
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "targets/delegation")

	// Setup another certificate
	tempFile2, err := ioutil.TempFile("", "pemfile2")
	require.NoError(t, err)

	privKey, err = utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	startTime = time.Now()
	endTime = startTime.AddDate(10, 0, 0)
	cert, err = cryptoservice.GenerateCertificate(privKey, "gun", startTime, endTime)
	require.NoError(t, err)

	_, err = tempFile2.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	tempFile2.Close()
	defer os.Remove(tempFile2.Name())

	rawPubBytes2, _ := ioutil.ReadFile(tempFile2.Name())
	parsedPubKey2, _ := utils.ParsePEMPublicKey(rawPubBytes2)
	keyID2, err := utils.CanonicalKeyID(parsedPubKey2)
	require.NoError(t, err)

	// add to the delegation by specifying the same role, this time add a scoped path
	_, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "add", "-p", "gun", "targets/delegation", tempFile2.Name(), "--paths", "path")
	require.NoError(t, err)

	// list delegations - we should see two keys
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "path")
	require.Contains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// remove the delegation's first key
	output, err = runCommand(t, tempDir, "delegation", "-s", server.URL, "remove", "-p", "gun", "targets/delegation", keyID)
	require.NoError(t, err)
	require.Contains(t, output, "Removal of delegation role")

	// list delegations - we should see the delegation but with only the second key
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, keyID)
	require.Contains(t, output, keyID2)

	// remove the delegation's second key
	output, err = runCommand(t, tempDir, "delegation", "-s", server.URL, "remove", "-p", "gun", "targets/delegation", keyID2)
	require.NoError(t, err)
	require.Contains(t, output, "Removal of delegation role")

	// list delegations - we should see no delegations
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, keyID)
	require.NotContains(t, output, keyID2)
}

// TestClientTUFAddByHashWithAutoPublish is similar to TestClientTUFAddByHashInteraction,
// but with the auto publish flag "-p".
func TestClientTUFAddByHashWithAutoPublish(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	targetData := []byte{'a', 'b', 'c'}
	target256Bytes := sha256.Sum256(targetData)
	targetSHA256Hex := hex.EncodeToString(target256Bytes[:])

	err := ioutil.WriteFile(filepath.Join(tempDir, "tempfile"), targetData, 0644)
	require.NoError(t, err)

	var (
		output  string
		target1 = "sdgkadga"
	)
	// -- tests --

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun", "-p")
	require.NoError(t, err)

	// add a target just by sha256
	_, err = runCommand(t, tempDir, "-s", server.URL, "addhash", "-p", "gun", target1, "3", "--sha256", targetSHA256Hex)
	require.NoError(t, err)

	// check status - no targets
	output, err = runCommand(t, tempDir, "status", "gun")
	require.NoError(t, err)
	require.False(t, strings.Contains(string(output), target1))

	// list repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, target1)

	// lookup target and repo - see target
	output, err = runCommand(t, tempDir, "-s", server.URL, "lookup", "gun", target1)
	require.NoError(t, err)
	require.Contains(t, output, target1)
}

// Tests import/export keys
func TestClientKeyImport(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	var (
		rootKeyID string
	)

	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	// close later, because we might need to write to it
	defer os.Remove(tempFile.Name())

	// -- tests --
	// test 1, no path but role=root included with non-encrypted key
	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err := utils.EncryptPrivateKey(privKey, data.CanonicalRootRole, "", "")
	require.NoError(t, err)

	nBytes, err := tempFile.Write(pemBytes)
	require.NoError(t, err)
	tempFile.Close()
	require.Equal(t, len(pemBytes), nBytes)
	rootKeyID = privKey.ID()

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", tempFile.Name())
	require.NoError(t, err)

	// if there is hardware available, root will only be on hardware, and not
	// on disk
	newRoot, _ := assertNumKeys(t, tempDir, 1, 0, !rootOnHardware())
	require.Equal(t, rootKeyID, newRoot[0])

	// test 2, no path but role flag included with unencrypted key

	tempFile2, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	// close later, because we might need to write to it
	defer os.Remove(tempFile2.Name())

	privKey, err = utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err = utils.EncryptPrivateKey(privKey, "", "", "")
	require.NoError(t, err)

	nBytes, err = tempFile2.Write(pemBytes)
	require.NoError(t, err)
	tempFile2.Close()
	require.Equal(t, len(pemBytes), nBytes)

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", tempFile2.Name(), "-r", data.CanonicalRootRole.String())
	require.NoError(t, err)

	// if there is hardware available, root will only be on hardware, and not
	// on disk
	assertNumKeys(t, tempDir, 2, 0, !rootOnHardware())

	// test 3, no path no role included with unencrypted key

	tempFile3, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	// close later, because we might need to write to it
	defer os.Remove(tempFile3.Name())

	privKey, err = utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err = utils.EncryptPrivateKey(privKey, "", "", "")
	require.NoError(t, err)

	nBytes, err = tempFile3.Write(pemBytes)
	require.NoError(t, err)
	tempFile3.Close()
	require.Equal(t, len(pemBytes), nBytes)

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", tempFile3.Name())
	require.NoError(t, err)

	// if there is hardware available, root will only be on hardware, and not
	// on disk
	assertNumKeys(t, tempDir, 2, 1, !rootOnHardware())
	file, err := os.OpenFile(filepath.Join(tempDir, notary.PrivDir, privKey.ID()+".key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	filebytes, _ := ioutil.ReadAll(file)
	require.Contains(t, string(filebytes), ("role: " + notary.DefaultImportRole))

	// test 4, no path non root role with non canonical role and gun flag with unencrypted key

	tempFile4, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	// close later, because we might need to write to it
	defer os.Remove(tempFile4.Name())

	privKey, err = utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err = utils.EncryptPrivateKey(privKey, "", "", "")
	require.NoError(t, err)

	nBytes, err = tempFile4.Write(pemBytes)
	require.NoError(t, err)
	tempFile4.Close()
	require.Equal(t, len(pemBytes), nBytes)

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", tempFile4.Name(), "-r", "somerole", "-g", "somegun")
	require.NoError(t, err)

	// if there is hardware available, root will only be on hardware, and not
	// on disk
	assertNumKeys(t, tempDir, 2, 2, !rootOnHardware())
	file, err = os.OpenFile(filepath.Join(tempDir, notary.PrivDir, privKey.ID()+".key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	filebytes, _ = ioutil.ReadAll(file)
	require.Contains(t, string(filebytes), ("role: " + "somerole"))
	require.NotContains(t, string(filebytes), ("gun: " + "somegun"))

	// test 5, no path non root role with canonical role and gun flag with unencrypted key

	tempFile5, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	// close later, because we might need to write to it
	defer os.Remove(tempFile5.Name())

	privKey, err = utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err = utils.EncryptPrivateKey(privKey, "", "", "")
	require.NoError(t, err)

	nBytes, err = tempFile5.Write(pemBytes)
	require.NoError(t, err)
	tempFile5.Close()
	require.Equal(t, len(pemBytes), nBytes)

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", tempFile5.Name(), "-r", data.CanonicalSnapshotRole.String(), "-g", "somegun")
	require.NoError(t, err)

	// if there is hardware available, root will only be on hardware, and not
	// on disk
	assertNumKeys(t, tempDir, 2, 3, !rootOnHardware())
	file, err = os.OpenFile(filepath.Join(tempDir, notary.PrivDir, privKey.ID()+".key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	filebytes, _ = ioutil.ReadAll(file)
	require.Contains(t, string(filebytes), ("role: " + data.CanonicalSnapshotRole.String()))
	require.Contains(t, string(filebytes), ("gun: " + "somegun"))

	// test6, no path but role=root included with encrypted key, should fail since we don't know what keyid to save to

	tempFile6, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	// close later, because we might need to write to it
	defer os.Remove(tempFile6.Name())

	privKey, err = utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err = utils.EncryptPrivateKey(privKey, data.CanonicalRootRole, "", testPassphrase)
	require.NoError(t, err)

	nBytes, err = tempFile6.Write(pemBytes)
	require.NoError(t, err)
	tempFile6.Close()
	require.Equal(t, len(pemBytes), nBytes)

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", tempFile6.Name())
	require.NoError(t, err)

	// if there is hardware available, root will only be on hardware, and not
	// on disk
	assertNumKeys(t, tempDir, 2, 3, !rootOnHardware())

	// test7, non root key with no path with no gun

	tempFile7, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	// close later, because we might need to write to it
	defer os.Remove(tempFile7.Name())

	privKey, err = utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err = utils.EncryptPrivateKey(privKey, "", "", "")
	require.NoError(t, err)

	nBytes, err = tempFile7.Write(pemBytes)
	require.NoError(t, err)
	tempFile7.Close()
	require.Equal(t, len(pemBytes), nBytes)

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", tempFile7.Name(), "-r", "somerole")
	require.NoError(t, err)

	// if there is hardware available, root will only be on hardware, and not
	// on disk
	assertNumKeys(t, tempDir, 2, 4, !rootOnHardware())
	file, err = os.OpenFile(filepath.Join(tempDir, notary.PrivDir, privKey.ID()+".key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	filebytes, _ = ioutil.ReadAll(file)
	require.Contains(t, string(filebytes), ("role: " + "somerole"))

	// test 8, non root canonical key with no gun

	tempFile8, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	// close later, because we might need to write to it
	defer os.Remove(tempFile8.Name())

	privKey, err = utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	pemBytes, err = utils.EncryptPrivateKey(privKey, data.CanonicalSnapshotRole, "", "")
	require.NoError(t, err)

	nBytes, err = tempFile8.Write(pemBytes)
	require.NoError(t, err)
	tempFile8.Close()
	require.Equal(t, len(pemBytes), nBytes)
	newKeyID := privKey.ID()

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", tempFile8.Name())
	require.NoError(t, err)

	// if there is hardware available, root will only be on hardware, and not
	// on disk
	assertNumKeys(t, tempDir, 2, 4, !rootOnHardware())
	_, err = os.Open(filepath.Join(tempDir, notary.PrivDir, newKeyID+".key"))
	require.Error(t, err)
}

func TestAddDelImportKeyPublishFlow(t *testing.T) {
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	// Setup certificate for delegation role
	tempFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)

	privKey, err := utils.GenerateRSAKey(rand.Reader, 2048)
	require.NoError(t, err)
	startTime := time.Now()
	endTime := startTime.AddDate(10, 0, 0)
	cert, err := cryptoservice.GenerateCertificate(privKey, "gun", startTime, endTime)
	require.NoError(t, err)

	// Setup key in a file for import
	keyFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	defer os.Remove(keyFile.Name())
	pemBytes, err := utils.EncryptPrivateKey(privKey, "", "", "")
	require.NoError(t, err)
	nBytes, err := keyFile.Write(pemBytes)
	require.NoError(t, err)
	keyFile.Close()
	require.Equal(t, len(pemBytes), nBytes)

	_, err = tempFile.Write(utils.CertToPEM(cert))
	require.NoError(t, err)
	tempFile.Close()
	defer os.Remove(tempFile.Name())

	rawPubBytes, _ := ioutil.ReadFile(tempFile.Name())
	parsedPubKey, _ := utils.ParsePEMPublicKey(rawPubBytes)
	canonicalKeyID, err := utils.CanonicalKeyID(parsedPubKey)
	require.NoError(t, err)

	// Set up targets for publishing
	tempTargetFile, err := ioutil.TempFile("", "targetfile")
	require.NoError(t, err)
	tempTargetFile.Close()
	defer os.Remove(tempTargetFile.Name())

	var target = "sdgkadga"

	var output string

	// init repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - none yet
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No delegations present in this repository.")

	// validate that we have all keys, including snapshot
	assertNumKeys(t, tempDir, 1, 2, true)

	// rotate the snapshot key to server
	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "rotate", "gun", data.CanonicalSnapshotRole.String(), "-r")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// validate that we lost the snapshot signing key
	_, signingKeyIDs := assertNumKeys(t, tempDir, 1, 1, true)
	targetKeyID := signingKeyIDs[0]

	// add new valid delegation with single new cert
	output, err = runCommand(t, tempDir, "delegation", "add", "gun", "targets/releases", tempFile.Name(), "--paths", "\"\"")
	require.NoError(t, err)
	require.Contains(t, output, "Addition of delegation role")
	require.Contains(t, output, canonicalKeyID)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - we should see our one delegation
	output, err = runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, "No delegations present in this repository.")

	// remove the targets key to demonstrate that delegates don't need this key
	require.NoError(t, os.Remove(filepath.Join(tempDir, notary.PrivDir, targetKeyID+".key")))

	// we are now set up with the first part, now import the delegation key- add a target- publish

	// first test the negative case, should fail without the key import

	// add a target using the delegation -- will only add to targets/releases
	_, err = runCommand(t, tempDir, "add", "gun", target, tempTargetFile.Name(), "--roles", "targets/releases")
	require.NoError(t, err)

	// list targets for targets/releases - we should see no targets until we publish
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "No targets")

	// check that our change is staged
	output, err = runCommand(t, tempDir, "-s", server.URL, "status", "gun")
	require.Contains(t, output, "targets/releases")
	require.Contains(t, output, "sdgkadga")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.Error(t, err)
	// list targets for targets/releases - we should not see our target!
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.Contains(t, output, "No targets present")

	// now test for the positive case, import the key and publish and it should work

	// the changelist still exists so no need to add the target again
	// just import the key and publish
	output, err = runCommand(t, tempDir, "-s", server.URL, "status", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "targets/releases")
	require.Contains(t, output, "sdgkadga")

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", keyFile.Name(), "-r", "targets/releases")
	require.NoError(t, err)

	// make sure that it has been imported fine
	// if there is hardware available, root will only be on hardware, and not
	// on disk
	_, err = os.Open(filepath.Join(tempDir, notary.PrivDir, privKey.ID()+".key"))
	require.NoError(t, err)

	// now try to publish
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// check that changelist is applied
	output, err = runCommand(t, tempDir, "-s", server.URL, "status", "gun")
	require.NoError(t, err)
	require.NotContains(t, output, "targets/releases")

	// list targets for targets/releases - we should see our target!
	output, err = runCommand(t, tempDir, "-s", server.URL, "list", "gun", "--roles", "targets/releases")
	require.NoError(t, err)
	require.NotContains(t, output, "No targets present")
	require.Contains(t, output, "sdgkadga")
	require.Contains(t, output, "targets/releases")
}

func TestExportImportFlow(t *testing.T) {
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	server := setupServer()
	defer server.Close()

	// init repo
	_, err := runCommand(t, tempDir, "-s", server.URL, "init", "gun")
	require.NoError(t, err)

	// publish repo
	_, err = runCommand(t, tempDir, "-s", server.URL, "publish", "gun")
	require.NoError(t, err)

	// list delegations - none yet
	output, err := runCommand(t, tempDir, "-s", server.URL, "delegation", "list", "gun")
	require.NoError(t, err)
	require.Contains(t, output, "No delegations present in this repository.")

	// validate that we have all keys, including snapshot
	assertNumKeys(t, tempDir, 1, 2, true)

	_, err = runCommand(t, tempDir, "-s", server.URL, "key", "export", "-o", filepath.Join(tempDir, "exported"))
	require.NoError(t, err)

	// make sure the export has been done properly
	from, err := os.OpenFile(filepath.Join(tempDir, "exported"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer from.Close()
	fromBytes, _ := ioutil.ReadAll(from)
	fromString := string(fromBytes)
	require.Contains(t, fromString, "role: snapshot")
	require.Contains(t, fromString, "role: root")
	require.Contains(t, fromString, "role: targets")

	// now setup new filestore
	newTempDir := tempDirWithConfig(t, "{}")
	defer os.Remove(newTempDir)

	// and new server
	newServer := setupServer()
	defer newServer.Close()

	// make sure there are no keys
	if !rootOnHardware() {
		assertNumKeys(t, newTempDir, 0, 0, true)
	}

	// import keys from our exported file
	_, err = runCommand(t, newTempDir, "-s", newServer.URL, "key", "import", filepath.Join(tempDir, "exported"))
	require.NoError(t, err)

	// validate that we have all keys, including snapshot
	assertNumKeys(t, newTempDir, 1, 2, !rootOnHardware())
	root, signing := getUniqueKeys(t, newTempDir)

	fileList := []string{}
	err = filepath.Walk(newTempDir, func(path string, f os.FileInfo, err error) error {
		fileList = append(fileList, path)
		return nil
	})
	require.NoError(t, err)

	if !rootOnHardware() {
		// validate root is imported correctly
		rootKey, err := os.OpenFile(filepath.Join(newTempDir, notary.PrivDir, root[0]+".key"), os.O_RDONLY, notary.PrivExecPerms)
		require.NoError(t, err)
		defer rootKey.Close()
		rootBytes, _ := ioutil.ReadAll(rootKey)
		rootString := string(rootBytes)
		require.Contains(t, rootString, "role: root")
	} else {
		verifyRootKeyOnHardware(t, root[0])
	}

	// validate snapshot is imported correctly
	snapKey, err := os.OpenFile(filepath.Join(newTempDir, notary.PrivDir, signing[0]+".key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer snapKey.Close()
	snapBytes, _ := ioutil.ReadAll(snapKey)
	snapString := string(snapBytes)
	require.Contains(t, snapString, "gun: gun")
	require.True(t, strings.Contains(snapString, "role: snapshot") || strings.Contains(snapString, "role: target"))

	// validate targets is imported correctly
	targKey, err := os.OpenFile(filepath.Join(newTempDir, notary.PrivDir, signing[1]+".key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer targKey.Close()
	targBytes, _ := ioutil.ReadAll(targKey)
	targString := string(targBytes)
	require.Contains(t, targString, "gun: gun")
	require.True(t, strings.Contains(snapString, "role: snapshot") || strings.Contains(snapString, "role: target"))
}

// Tests import/export keys with delegations, which don't require a gun
func TestDelegationKeyImportExport(t *testing.T) {
	// -- setup --
	setUp(t)

	tempDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	tempExportedDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	tempImportingDir := tempDirWithConfig(t, "{}")
	defer os.RemoveAll(tempDir)

	// Setup key in a file for import
	keyFile, err := ioutil.TempFile("", "pemfile")
	require.NoError(t, err)
	defer os.Remove(keyFile.Name())
	privKey, err := utils.GenerateRSAKey(rand.Reader, 2048)
	require.NoError(t, err)
	pemBytes, err := utils.EncryptPrivateKey(privKey, "", "", "")
	require.NoError(t, err)
	nBytes, err := keyFile.Write(pemBytes)
	require.NoError(t, err)
	keyFile.Close()
	require.Equal(t, len(pemBytes), nBytes)

	// import the key
	_, err = runCommand(t, tempDir, "key", "import", keyFile.Name(), "-r", "user")
	require.NoError(t, err)

	// export the key
	_, err = runCommand(t, tempDir, "key", "export", "-o", filepath.Join(tempExportedDir, "exported"))
	require.NoError(t, err)

	// re-import the key from the exported store to a new tempDir
	_, err = runCommand(t, tempImportingDir, "key", "import", filepath.Join(tempExportedDir, "exported"))
	require.NoError(t, err)
}
