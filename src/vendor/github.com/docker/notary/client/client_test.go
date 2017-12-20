package client

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	ctxu "github.com/docker/distribution/context"
	"github.com/docker/go/canonical/json"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"

	"github.com/docker/notary"
	"github.com/docker/notary/client/changelist"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/server"
	"github.com/docker/notary/server/storage"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/utils"
	"github.com/docker/notary/tuf/validation"
)

const password = "passphrase"

type passRoleRecorder struct {
	rolesCreated []string
	rolesAsked   []string
}

func newRoleRecorder() *passRoleRecorder {
	return &passRoleRecorder{}
}

func (p *passRoleRecorder) clear() {
	p.rolesCreated = nil
	p.rolesAsked = nil
}

func (p *passRoleRecorder) retriever(_, alias string, createNew bool, _ int) (string, bool, error) {
	if createNew {
		p.rolesCreated = append(p.rolesCreated, alias)
	} else {
		p.rolesAsked = append(p.rolesAsked, alias)
	}
	return password, false, nil
}

func (p *passRoleRecorder) compareRolesRecorded(t *testing.T, expected []string, created bool,
	args ...interface{}) {

	var actual, useExpected sort.StringSlice
	copy(expected, useExpected) // don't sort expected, since we don't want to mutate it
	sort.Stable(useExpected)

	if created {
		copy(p.rolesCreated, actual)
	} else {
		copy(p.rolesAsked, actual)
	}
	sort.Stable(actual)

	require.Equal(t, useExpected, actual, args...)
}

// requires the following keys be created: order does not matter
func (p *passRoleRecorder) requireCreated(t *testing.T, expected []string, args ...interface{}) {
	p.compareRolesRecorded(t, expected, true, args...)
}

// requires that passwords be asked for the following keys: order does not matter
func (p *passRoleRecorder) requireAsked(t *testing.T, expected []string, args ...interface{}) {
	p.compareRolesRecorded(t, expected, false, args...)
}

var passphraseRetriever = passphrase.ConstantRetriever(password)

func simpleTestServer(t *testing.T, roles ...string) (
	*httptest.Server, *http.ServeMux, map[string]data.PrivateKey) {

	if len(roles) == 0 {
		roles = []string{data.CanonicalTimestampRole.String(), data.CanonicalSnapshotRole.String()}
	}
	keys := make(map[string]data.PrivateKey)
	mux := http.NewServeMux()

	for _, role := range roles {
		key, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err)

		keys[role] = key
		pubKey := data.PublicKeyFromPrivate(key)
		jsonBytes, err := json.MarshalCanonical(&pubKey)
		require.NoError(t, err)
		keyJSON := string(jsonBytes)

		// TUF will request /v2/docker.com/notary/_trust/tuf/<role>.key
		mux.HandleFunc(
			fmt.Sprintf("/v2/docker.com/notary/_trust/tuf/%s.key", role),
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, keyJSON)
			})
	}

	ts := httptest.NewServer(mux)
	return ts, mux, keys
}

func fullTestServer(t *testing.T) *httptest.Server {
	// Set up server
	ctx := context.WithValue(
		context.Background(), notary.CtxKeyMetaStore, storage.NewMemStorage())

	// Do not pass one of the const KeyAlgorithms here as the value! Passing a
	// string is in itself good test that we are handling it correctly as we
	// will be receiving a string from the configuration.
	ctx = context.WithValue(ctx, notary.CtxKeyKeyAlgo, "ecdsa")

	// Eat the logs instead of spewing them out
	var b bytes.Buffer
	l := logrus.New()
	l.Out = &b
	ctx = ctxu.WithLogger(ctx, logrus.NewEntry(l))

	cryptoService := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(passphraseRetriever))
	return httptest.NewServer(server.RootHandler(ctx, nil, cryptoService, nil, nil, nil))
}

// server that returns some particular error code all the time
func errorTestServer(t *testing.T, errorCode int) *httptest.Server {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(errorCode)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	return server
}

// initializes a repository in a temporary directory
func initializeRepo(t *testing.T, rootType, gun, url string,
	serverManagesSnapshot bool) (*NotaryRepository, string) {

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	serverManagedRoles := []data.RoleName{}
	if serverManagesSnapshot {
		serverManagedRoles = []data.RoleName{data.CanonicalSnapshotRole}
	}

	repo, rec, rootPubKeyID := createRepoAndKey(t, rootType, tempBaseDir, gun, url)

	err = repo.Initialize([]string{rootPubKeyID}, serverManagedRoles...)
	if err != nil {
		os.RemoveAll(tempBaseDir)
	}
	require.NoError(t, err, "error creating repository: %s", err)

	// generates the target role, maybe the snapshot role
	if serverManagesSnapshot {
		rec.requireCreated(t, []string{data.CanonicalTargetsRole.String()})
	} else {
		rec.requireCreated(t, []string{data.CanonicalTargetsRole.String(), data.CanonicalSnapshotRole.String()})
	}
	// root key is cached by the cryptoservice, so when signing we don't actually ask
	// for the passphrase
	rec.requireAsked(t, nil)
	return repo, rootPubKeyID
}

// Creates a new repository and adds a root key.  Returns the repo and key ID.
func createRepoAndKey(t *testing.T, rootType, tempBaseDir, gun, url string) (
	*NotaryRepository, *passRoleRecorder, string) {

	rec := newRoleRecorder()
	repo, err := NewFileCachedNotaryRepository(
		tempBaseDir, data.GUN(gun), url, http.DefaultTransport, rec.retriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	rootPubKey, err := repo.CryptoService.Create(data.CanonicalRootRole, repo.gun, rootType)
	require.NoError(t, err, "error generating root key: %s", err)

	rec.requireCreated(t, []string{data.CanonicalRootRole.String()},
		"root passphrase should have been required to generate a root key")
	rec.requireAsked(t, nil)
	rec.clear()

	return repo, rec, rootPubKey.ID()
}

// creates a new notary repository with the same gun and url as the previous
// repo, in order to eliminate caches (for instance, cryptoservice cache)
// if a new directory is to be created, it also eliminates the TUF metadata
// cache
func newRepoToTestRepo(t *testing.T, existingRepo *NotaryRepository, newDir bool) (
	*NotaryRepository, *passRoleRecorder) {

	repoDir := existingRepo.baseDir
	if newDir {
		tempBaseDir, err := ioutil.TempDir("", "notary-test-")
		require.NoError(t, err, "failed to create a temporary directory")
		repoDir = tempBaseDir
	}

	rec := newRoleRecorder()
	repo, err := NewFileCachedNotaryRepository(
		repoDir, existingRepo.gun, existingRepo.baseURL,
		http.DefaultTransport, rec.retriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repository: %s", err)
	if err != nil && newDir {
		defer os.RemoveAll(repoDir)
	}

	return repo, rec
}

// Initializing a new repo while specifying that the server should manage the root
// role will fail.
func TestInitRepositoryManagedRolesIncludingRoot(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	repo, rec, rootPubKeyID := createRepoAndKey(
		t, data.ECDSAKey, tempBaseDir, "docker.com/notary", "http://localhost")
	err = repo.Initialize([]string{rootPubKeyID}, data.CanonicalRootRole)
	require.Error(t, err)
	require.IsType(t, ErrInvalidRemoteRole{}, err)
	// Just testing the error message here in this one case
	require.Equal(t, err.Error(),
		"notary does not permit the server managing the root key")
	// no key creation happened
	rec.requireCreated(t, nil)
}

// Initializing a new repo while specifying that the server should manage some
// invalid role will fail.
func TestInitRepositoryManagedRolesInvalidRole(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	repo, rec, rootPubKeyID := createRepoAndKey(
		t, data.ECDSAKey, tempBaseDir, "docker.com/notary", "http://localhost")
	err = repo.Initialize([]string{rootPubKeyID}, "randomrole")
	require.Error(t, err)
	require.IsType(t, ErrInvalidRemoteRole{}, err)
	// no key creation happened
	rec.requireCreated(t, nil)
}

// Initializing a new repo while specifying that the server should manage the
// targets role will fail.
func TestInitRepositoryManagedRolesIncludingTargets(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	repo, rec, rootPubKeyID := createRepoAndKey(
		t, data.ECDSAKey, tempBaseDir, "docker.com/notary", "http://localhost")
	err = repo.Initialize([]string{rootPubKeyID}, data.CanonicalTargetsRole)
	require.Error(t, err)
	require.IsType(t, ErrInvalidRemoteRole{}, err)
	// no key creation happened
	rec.requireCreated(t, nil)
}

// Initializing a new repo while specifying that the server should manage the
// timestamp key is fine - that's what it already does, so no error.
func TestInitRepositoryManagedRolesIncludingTimestamp(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rec, rootPubKeyID := createRepoAndKey(
		t, data.ECDSAKey, tempBaseDir, "docker.com/notary", ts.URL)
	err = repo.Initialize([]string{rootPubKeyID}, data.CanonicalTimestampRole)
	require.NoError(t, err)

	// generates the target role, the snapshot role
	rec.requireCreated(t, []string{data.CanonicalTargetsRole.String(), data.CanonicalSnapshotRole.String()})
}

func TestInitRepositoryMultipleRootKeys(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rec, rootPubKeyID := createRepoAndKey(
		t, data.ECDSAKey, tempBaseDir, "docker.com/notary", ts.URL)
	rootPubKey2, err := repo.CryptoService.Create(data.CanonicalRootRole, repo.gun, data.ECDSAKey)
	require.NoError(t, err, "error generating second root key: %s", err)

	err = repo.Initialize([]string{rootPubKeyID, rootPubKey2.ID()}, data.CanonicalTimestampRole)
	require.NoError(t, err)

	// generates the target role, the snapshot role
	rec.requireCreated(t, []string{data.CanonicalTargetsRole.String(), data.CanonicalSnapshotRole.String()})

	// has two root keys
	require.Len(t, repo.tufRepo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs, 2)
}

// Initializing a new repo fails if unable to get the timestamp key, even if
// the snapshot key is available
func TestInitRepositoryNeedsRemoteTimestampKey(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	ts, _, _ := simpleTestServer(t, data.CanonicalSnapshotRole.String())
	defer ts.Close()

	repo, rec, rootPubKeyID := createRepoAndKey(
		t, data.ECDSAKey, tempBaseDir, "docker.com/notary", ts.URL)
	err = repo.Initialize([]string{rootPubKeyID}, data.CanonicalTimestampRole)
	require.Error(t, err)
	require.IsType(t, store.ErrMetaNotFound{}, err)

	// locally managed keys are created first, to avoid unnecssary network calls,
	// so they would have been generated
	rec.requireCreated(t, []string{data.CanonicalTargetsRole.String(), data.CanonicalSnapshotRole.String()})
}

// Initializing a new repo with remote server signing fails if unable to get
// the snapshot key, even if the timestamp key is available
func TestInitRepositoryNeedsRemoteSnapshotKey(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory")
	defer os.RemoveAll(tempBaseDir)

	ts, _, _ := simpleTestServer(t, data.CanonicalTimestampRole.String())
	defer ts.Close()

	repo, rec, rootPubKeyID := createRepoAndKey(
		t, data.ECDSAKey, tempBaseDir, "docker.com/notary", ts.URL)
	err = repo.Initialize([]string{rootPubKeyID}, data.CanonicalSnapshotRole)
	require.Error(t, err)
	require.IsType(t, store.ErrMetaNotFound{}, err)

	// locally managed keys are created first, to avoid unnecssary network calls,
	// so they would have been generated
	rec.requireCreated(t, []string{data.CanonicalTargetsRole.String()})
}

// passing timestamp + snapshot, or just snapshot, is tested in the next two
// test cases.

// TestInitRepoServerOnlyManagesTimestampKey runs through the process of
// initializing a repository and makes sure the repository looks correct on disk.
// We test this with both an RSA and ECDSA root key.
// This test case covers the default case where the server only manages the
// timestamp key.
func TestInitRepoServerOnlyManagesTimestampKey(t *testing.T) {
	testInitRepoMetadata(t, data.ECDSAKey, false)
	testInitRepoSigningKeys(t, data.ECDSAKey, false)
	if !testing.Short() {
		testInitRepoMetadata(t, data.RSAKey, false)
		testInitRepoSigningKeys(t, data.RSAKey, false)
	}
}

// TestInitRepoServerManagesTimestampAndSnapshotKeys runs through the process of
// initializing a repository and makes sure the repository looks correct on disk.
// We test this with both an RSA and ECDSA root key.
// This test case covers the server managing both the timestamp and snapshot keys.
func TestInitRepoServerManagesTimestampAndSnapshotKeys(t *testing.T) {
	testInitRepoMetadata(t, data.ECDSAKey, true)
	testInitRepoSigningKeys(t, data.ECDSAKey, true)
	if !testing.Short() {
		testInitRepoMetadata(t, data.RSAKey, true)
		testInitRepoSigningKeys(t, data.RSAKey, false)
	}
}

// This creates a new KeyFileStore in the repo's base directory and makes sure
// the repo has the right number of keys
func requireRepoHasExpectedKeys(t *testing.T, repo *NotaryRepository,
	rootKeyID string, expectedSnapshotKey bool) {

	// The repo should have a keyFileStore and have created keys using it,
	// so create a new KeyFileStore, and check that the keys do exist and are
	// valid
	ks, err := trustmanager.NewKeyFileStore(repo.baseDir, passphraseRetriever)
	require.NoError(t, err)

	roles := make(map[string]bool)
	for keyID, keyInfo := range ks.ListKeys() {
		if keyInfo.Role == data.CanonicalRootRole {
			require.Equal(t, rootKeyID, keyID, "Unexpected root key ID")
		}
		// just to ensure the content of the key files created are valid
		_, r, err := ks.GetKey(keyID)
		require.NoError(t, err)
		require.Equal(t, keyInfo.Role, r)
		roles[keyInfo.Role.String()] = true
	}
	// there is a root key and a targets key
	alwaysThere := []string{data.CanonicalRootRole.String(), data.CanonicalTargetsRole.String()}
	for _, role := range alwaysThere {
		_, ok := roles[role]
		require.True(t, ok, "missing %s key", role)
	}

	// there may be a snapshots key, depending on whether the server is managing
	// the snapshots key
	_, ok := roles[data.CanonicalSnapshotRole.String()]
	if expectedSnapshotKey {
		require.True(t, ok, "missing snapshot key")
	} else {
		require.False(t, ok,
			"there should be no snapshot key because the server manages it")
	}

	// The server manages the timestamp key - there should not be a timestamp
	// key
	_, ok = roles[data.CanonicalTimestampRole.String()]
	require.False(t, ok,
		"there should be no timestamp key because the server manages it")
}

// Sanity check the TUF metadata files. Verify that it exists for a particular
// role, the JSON is well-formed, and the signatures exist.
// For the root.json file, also check that the root, snapshot, and
// targets key IDs are present.
func requireRepoHasExpectedMetadata(t *testing.T, repo *NotaryRepository,
	role data.RoleName, expected bool) {

	filename := filepath.Join(tufDir, filepath.FromSlash(repo.gun.String()),
		"metadata", role.String()+".json")
	fullPath := filepath.Join(repo.baseDir, filename)
	_, err := os.Stat(fullPath)

	if expected {
		require.NoError(t, err, "missing TUF metadata file: %s", filename)
	} else {
		require.Error(t, err,
			"%s metadata should not exist, but does: %s", role.String(), filename)
		return
	}

	jsonBytes, err := ioutil.ReadFile(fullPath)
	require.NoError(t, err, "error reading TUF metadata file %s: %s", filename, err)

	var decoded data.Signed
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err, "error parsing TUF metadata file %s: %s", filename, err)

	require.Len(t, decoded.Signatures, 1,
		"incorrect number of signatures in TUF metadata file %s", filename)

	require.NotEmpty(t, decoded.Signatures[0].KeyID,
		"empty key ID field in TUF metadata file %s", filename)
	require.NotEmpty(t, decoded.Signatures[0].Method,
		"empty method field in TUF metadata file %s", filename)
	require.NotEmpty(t, decoded.Signatures[0].Signature,
		"empty signature in TUF metadata file %s", filename)

	// Special case for root.json: also check that the signed
	// content for keys and roles
	if role == data.CanonicalRootRole {
		var decodedRoot data.Root
		err := json.Unmarshal(*decoded.Signed, &decodedRoot)
		require.NoError(t, err, "error parsing root.json signed section: %s", err)

		require.Equal(t, "Root", decodedRoot.Type, "_type mismatch in root.json")

		// Expect 1 key for each valid role in the Keys map - one for
		// each of root, targets, snapshot, timestamp
		require.Len(t, decodedRoot.Keys, len(data.BaseRoles),
			"wrong number of keys in root.json")
		require.True(t, len(decodedRoot.Roles) >= len(data.BaseRoles),
			"wrong number of roles in root.json")

		for _, role := range data.BaseRoles {
			_, ok := decodedRoot.Roles[role]
			require.True(t, ok, "Missing role %s in root.json", role)
		}
	}
}

func testInitRepoMetadata(t *testing.T, rootType string, serverManagesSnapshot bool) {
	gun := "docker.com/notary"

	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rootKeyID := initializeRepo(t, rootType, gun, ts.URL, serverManagesSnapshot)
	defer os.RemoveAll(repo.baseDir)

	requireRepoHasExpectedKeys(t, repo, rootKeyID, !serverManagesSnapshot)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalRootRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTargetsRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole,
		!serverManagesSnapshot)
}

func testInitRepoSigningKeys(t *testing.T, rootType string, serverManagesSnapshot bool) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	repo, _, rootPubKeyID := createRepoAndKey(
		t, data.ECDSAKey, tempBaseDir, "docker.com/notary", ts.URL)

	// create a new repository, so we can wipe out the cryptoservice's cached
	// keys, so we can test which keys it asks for passwords for
	repo, rec := newRepoToTestRepo(t, repo, false)

	if serverManagesSnapshot {
		err = repo.Initialize([]string{rootPubKeyID}, data.CanonicalSnapshotRole)
	} else {
		err = repo.Initialize([]string{rootPubKeyID})
	}

	require.NoError(t, err, "error initializing repository")

	// generates the target role, maybe the snapshot role
	if serverManagesSnapshot {
		rec.requireCreated(t, []string{data.CanonicalTargetsRole.String()})
	} else {
		rec.requireCreated(t, []string{data.CanonicalTargetsRole.String(), data.CanonicalSnapshotRole.String()})
	}
	// root is asked for signing the root role
	rec.requireAsked(t, []string{data.CanonicalRootRole.String()})
}

// TestInitRepoAttemptsExceeded tests error handling when passphrase.Retriever
// (or rather the user) insists on an incorrect password.
func TestInitRepoAttemptsExceeded(t *testing.T) {
	testInitRepoAttemptsExceeded(t, data.ECDSAKey)
	if !testing.Short() {
		testInitRepoAttemptsExceeded(t, data.RSAKey)
	}
}

func testInitRepoAttemptsExceeded(t *testing.T, rootType string) {
	var gun data.GUN = "docker.com/notary"
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	defer os.RemoveAll(tempBaseDir)

	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	retriever := passphrase.ConstantRetriever("password")
	repo, err := NewFileCachedNotaryRepository(tempBaseDir, gun, ts.URL, http.DefaultTransport, retriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)
	rootPubKey, err := repo.CryptoService.Create(data.CanonicalRootRole, repo.gun, rootType)
	require.NoError(t, err, "error generating root key: %s", err)

	retriever = passphrase.ConstantRetriever("incorrect password")
	// repo.CryptoService’s FileKeyStore caches the unlocked private key, so to test
	// private key unlocking we need a new repo instance.
	repo, err = NewFileCachedNotaryRepository(tempBaseDir, gun, ts.URL, http.DefaultTransport, retriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)
	err = repo.Initialize([]string{rootPubKey.ID()})
	require.EqualError(t, err, trustmanager.ErrAttemptsExceeded{}.Error())
}

// TestInitRepoPasswordInvalid tests error handling when passphrase.Retriever
// (or rather the user) fails to provide a correct password.
func TestInitRepoPasswordInvalid(t *testing.T) {
	testInitRepoPasswordInvalid(t, data.ECDSAKey)
	if !testing.Short() {
		testInitRepoPasswordInvalid(t, data.RSAKey)
	}
}

func giveUpPassphraseRetriever(_, _ string, _ bool, _ int) (string, bool, error) {
	return "", true, nil
}

func testInitRepoPasswordInvalid(t *testing.T, rootType string) {
	var gun data.GUN = "docker.com/notary"
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	defer os.RemoveAll(tempBaseDir)

	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	retriever := passphrase.ConstantRetriever("password")
	repo, err := NewFileCachedNotaryRepository(tempBaseDir, gun, ts.URL, http.DefaultTransport, retriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)
	rootPubKey, err := repo.CryptoService.Create(data.CanonicalRootRole, repo.gun, rootType)
	require.NoError(t, err, "error generating root key: %s", err)

	// repo.CryptoService’s FileKeyStore caches the unlocked private key, so to test
	// private key unlocking we need a new repo instance.
	repo, err = NewFileCachedNotaryRepository(tempBaseDir, gun, ts.URL, http.DefaultTransport, giveUpPassphraseRetriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)
	err = repo.Initialize([]string{rootPubKey.ID()})
	require.EqualError(t, err, trustmanager.ErrPasswordInvalid{}.Error())
}

func addTarget(t *testing.T, repo *NotaryRepository, targetName, targetFile string,
	roles ...data.RoleName) *Target {
	target, err := NewTarget(targetName, targetFile)
	require.NoError(t, err, "error creating target")
	err = repo.AddTarget(target, roles...)
	require.NoError(t, err, "error adding target")
	return target
}

// calls GetChangelist and gets the actual changes out
func getChanges(t *testing.T, repo *NotaryRepository) []changelist.Change {
	changeList, err := repo.GetChangelist()
	require.NoError(t, err)
	return changeList.List()
}

// TestAddTargetToTargetRoleByDefault adds a target without specifying a role
// to a repo without delegations.  Confirms that the changelist is created
// correctly, for the targets scope.
func TestAddTargetToTargetRoleByDefault(t *testing.T) {
	testAddTargetToTargetRoleByDefault(t, false)
	testAddTargetToTargetRoleByDefault(t, true)
}

func testAddTargetToTargetRoleByDefault(t *testing.T, clearCache bool) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	var rec *passRoleRecorder
	if clearCache {
		repo, rec = newRepoToTestRepo(t, repo, false)
	}

	testAddOrDeleteTarget(t, repo, changelist.ActionCreate, nil,
		[]string{data.CanonicalTargetsRole.String()})

	if clearCache {
		// no key creation or signing happened, because add doesn't ever require signing
		rec.requireCreated(t, nil)
		rec.requireAsked(t, nil)
	}
}

// Tests that adding a target to a repo or deleting a target from a repo,
// with the given roles, makes a change to the expected scopes
func testAddOrDeleteTarget(t *testing.T, repo *NotaryRepository, action string,
	rolesToChange []data.RoleName, expectedScopes []string) {

	require.Len(t, getChanges(t, repo), 0, "should start with zero changes")

	if action == changelist.ActionCreate {
		// Add fixtures/intermediate-ca.crt as a target. There's no particular
		// reason for using this file except that it happens to be available as
		// a fixture.
		addTarget(t, repo, "latest", "../fixtures/intermediate-ca.crt", rolesToChange...)
	} else {
		err := repo.RemoveTarget("latest", rolesToChange...)
		require.NoError(t, err, "error removing target")
	}

	changes := getChanges(t, repo)
	require.Len(t, changes, len(expectedScopes), "wrong number of changes files found")

	foundScopes := make(map[string]bool)
	for _, c := range changes { // there is only one
		require.EqualValues(t, action, c.Action())
		foundScopes[c.Scope().String()] = true
		require.Equal(t, "target", c.Type())
		require.Equal(t, "latest", c.Path())
		if action == changelist.ActionCreate {
			require.NotEmpty(t, c.Content())
		} else {
			require.Empty(t, c.Content())
		}
	}
	require.Len(t, foundScopes, len(expectedScopes))
	for _, expectedScope := range expectedScopes {
		_, ok := foundScopes[expectedScope]
		require.True(t, ok, "Target was not added/removed from %s", expectedScope)
	}

	// add/delete a second time
	if action == changelist.ActionCreate {
		addTarget(t, repo, "current", "../fixtures/intermediate-ca.crt", rolesToChange...)
	} else {
		err := repo.RemoveTarget("current", rolesToChange...)
		require.NoError(t, err, "error removing target")
	}

	changes = getChanges(t, repo)
	require.Len(t, changes, 2*len(expectedScopes),
		"wrong number of changelist files found")

	newFileFound := false
	foundScopes = make(map[string]bool)
	for _, c := range changes {
		if c.Path() != "latest" {
			require.EqualValues(t, action, c.Action())
			foundScopes[c.Scope().String()] = true
			require.Equal(t, "target", c.Type())
			require.Equal(t, "current", c.Path())
			if action == changelist.ActionCreate {
				require.NotEmpty(t, c.Content())
			} else {
				require.Empty(t, c.Content())
			}

			newFileFound = true
		}
	}
	require.True(t, newFileFound, "second changelist file not found")
	require.Len(t, foundScopes, len(expectedScopes))
	for _, expectedScope := range expectedScopes {
		_, ok := foundScopes[expectedScope]
		require.True(t, ok, "Target was not added/removed from %s", expectedScope)
	}
}

// TestAddTargetToSpecifiedValidRoles adds a target to the specified roles.
// Confirms that the changelist is created correctly, one for each of the
// the specified roles as scopes.
func TestAddTargetToSpecifiedValidRoles(t *testing.T) {
	testAddTargetToSpecifiedValidRoles(t, false)
	testAddTargetToSpecifiedValidRoles(t, true)
}

func testAddTargetToSpecifiedValidRoles(t *testing.T, clearCache bool) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	var rec *passRoleRecorder
	if clearCache {
		repo, rec = newRepoToTestRepo(t, repo, false)
	}

	roleName := filepath.Join(data.CanonicalTargetsRole.String(), "a")
	testAddOrDeleteTarget(t, repo, changelist.ActionCreate,
		[]data.RoleName{
			data.CanonicalTargetsRole,
			data.RoleName(roleName),
		},
		[]string{data.CanonicalTargetsRole.String(), roleName})

	if clearCache {
		// no key creation or signing happened, because add doesn't ever require signing
		rec.requireCreated(t, nil)
		rec.requireAsked(t, nil)
	}
}

// TestAddTargetToSpecifiedInvalidRoles expects errors to be returned if
// adding a target to an invalid role.  If any of the roles are invalid,
// no targets are added to any roles.
func TestAddTargetToSpecifiedInvalidRoles(t *testing.T) {
	testAddTargetToSpecifiedInvalidRoles(t, false)
	testAddTargetToSpecifiedInvalidRoles(t, true)
}

func testAddTargetToSpecifiedInvalidRoles(t *testing.T, clearCache bool) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	var rec *passRoleRecorder
	if clearCache {
		repo, rec = newRepoToTestRepo(t, repo, false)
	}

	invalidRoles := []data.RoleName{
		data.CanonicalRootRole,
		data.CanonicalSnapshotRole,
		data.CanonicalTimestampRole,
		"target/otherrole",
		"otherrole",
		"TARGETS/ALLCAPSROLE",
	}

	for _, invalidRole := range invalidRoles {
		target, err := NewTarget("latest", "../fixtures/intermediate-ca.crt")
		require.NoError(t, err, "error creating target")

		err = repo.AddTarget(target, data.CanonicalTargetsRole, invalidRole)
		require.Error(t, err, "Expected an ErrInvalidRole error")
		require.IsType(t, data.ErrInvalidRole{}, err)

		changes := getChanges(t, repo)
		require.Len(t, changes, 0)
	}

	if clearCache {
		// no key creation or signing happened, because add doesn't ever require signing
		rec.requireCreated(t, nil)
		rec.requireAsked(t, nil)
	}
}

// General way to require that errors writing a changefile are propagated up
func testErrorWritingChangefiles(t *testing.T, writeChangeFile func(*NotaryRepository) error) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()
	gun := "docker.com/notary"
	repo, _ := initializeRepo(t, data.ECDSAKey, gun, ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	// first, make the actual changefile unwritable by making the changelist
	// directory unwritable
	changelistPath := filepath.Join(
		filepath.Join(repo.baseDir, tufDir, filepath.FromSlash(gun)), "changelist",
	)
	err := os.MkdirAll(changelistPath, 0744)
	require.NoError(t, err, "could not create changelist dir")
	err = os.Chmod(changelistPath, 0600)
	require.NoError(t, err, "could not change permission of changelist dir")

	err = writeChangeFile(repo)
	require.Error(t, err, "Expected an error writing the change")
	require.IsType(t, &os.PathError{}, err)

	// then break prevent the changlist directory from being able to be created
	err = os.Chmod(changelistPath, 0744)
	require.NoError(t, err, "could not change permission of temp dir")
	err = os.RemoveAll(changelistPath)
	require.NoError(t, err, "could not remove changelist dir")
	// creating a changelist file so the directory can't be created
	err = ioutil.WriteFile(changelistPath, []byte("hi"), 0644)
	require.NoError(t, err, "could not write temporary file")

	err = writeChangeFile(repo)
	require.Error(t, err, "Expected an error writing the change")
	require.IsType(t, &os.PathError{}, err)
}

// Ensures that AddTarget errors on invalid target input (no hashes)
func TestAddTargetWithInvalidTarget(t *testing.T) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	target, err := NewTarget("latest", "../fixtures/intermediate-ca.crt")
	require.NoError(t, err, "error creating target")

	// Clear the hashes
	target.Hashes = data.Hashes{}
	require.Error(t, repo.AddTarget(target, data.CanonicalTargetsRole))
}

// TestAddTargetErrorWritingChanges expects errors writing a change to file
// to be propagated.
func TestAddTargetErrorWritingChanges(t *testing.T) {
	testErrorWritingChangefiles(t, func(repo *NotaryRepository) error {
		target, err := NewTarget("latest", "../fixtures/intermediate-ca.crt")
		require.NoError(t, err, "error creating target")
		return repo.AddTarget(target, data.CanonicalTargetsRole)
	})
}

// TestRemoveTargetToTargetRoleByDefault removes a target without specifying a
// role from a repo.  Confirms that the changelist is created correctly for
// the targets scope.
func TestRemoveTargetToTargetRoleByDefault(t *testing.T) {
	testRemoveTargetToTargetRoleByDefault(t, false)
	testRemoveTargetToTargetRoleByDefault(t, true)
}

func testRemoveTargetToTargetRoleByDefault(t *testing.T, clearCache bool) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	var rec *passRoleRecorder
	if clearCache {
		repo, rec = newRepoToTestRepo(t, repo, false)
	}

	testAddOrDeleteTarget(t, repo, changelist.ActionDelete, nil,
		[]string{data.CanonicalTargetsRole.String()})

	if clearCache {
		// no key creation or signing happened, because remove doesn't ever require signing
		rec.requireCreated(t, nil)
		rec.requireAsked(t, nil)
	}
}

// TestRemoveTargetFromSpecifiedValidRoles removes a target from the specified
// roles. Confirms that the changelist is created correctly, one for each of
// the the specified roles as scopes.
func TestRemoveTargetFromSpecifiedValidRoles(t *testing.T) {
	testRemoveTargetFromSpecifiedValidRoles(t, false)
	testRemoveTargetFromSpecifiedValidRoles(t, true)
}

func testRemoveTargetFromSpecifiedValidRoles(t *testing.T, clearCache bool) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	var rec *passRoleRecorder
	if clearCache {
		repo, rec = newRepoToTestRepo(t, repo, false)
	}

	roleName := filepath.Join(data.CanonicalTargetsRole.String(), "a")
	testAddOrDeleteTarget(t, repo, changelist.ActionDelete,
		[]data.RoleName{
			data.CanonicalTargetsRole,
			data.RoleName(roleName),
		},
		[]string{data.CanonicalTargetsRole.String(), roleName})

	if clearCache {
		// no key creation or signing happened, because remove doesn't ever require signing
		rec.requireCreated(t, nil)
		rec.requireAsked(t, nil)
	}
}

// TestRemoveTargetFromSpecifiedInvalidRoles expects errors to be returned if
// removing a target to an invalid role.  If any of the roles are invalid,
// no targets are removed from any roles.
func TestRemoveTargetToSpecifiedInvalidRoles(t *testing.T) {
	testRemoveTargetToSpecifiedInvalidRoles(t, false)
	testRemoveTargetToSpecifiedInvalidRoles(t, true)
}

func testRemoveTargetToSpecifiedInvalidRoles(t *testing.T, clearCache bool) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	var rec *passRoleRecorder
	if clearCache {
		repo, rec = newRepoToTestRepo(t, repo, false)
	}

	invalidRoles := []data.RoleName{
		data.CanonicalRootRole,
		data.CanonicalSnapshotRole,
		data.CanonicalTimestampRole,
		"target/otherrole",
		"otherrole",
	}

	for _, invalidRole := range invalidRoles {
		err := repo.RemoveTarget("latest", data.CanonicalTargetsRole, invalidRole)
		require.Error(t, err, "Expected an ErrInvalidRole error")
		require.IsType(t, data.ErrInvalidRole{}, err)

		changes := getChanges(t, repo)
		require.Len(t, changes, 0)
	}

	if clearCache {
		// no key creation or signing happened, because remove doesn't ever require signing
		rec.requireCreated(t, nil)
		rec.requireAsked(t, nil)
	}
}

// TestRemoveTargetErrorWritingChanges expects errors writing a change to file
// to be propagated.
func TestRemoveTargetErrorWritingChanges(t *testing.T) {
	testErrorWritingChangefiles(t, func(repo *NotaryRepository) error {
		return repo.RemoveTarget("latest", data.CanonicalTargetsRole)
	})
}

// TestListTarget fakes serving signed metadata files over the test's
// internal HTTP server to ensure that ListTargets returns the correct number
// of listed targets.
// We test this with both an RSA and ECDSA root key
func TestListTarget(t *testing.T) {
	testListEmptyTargets(t, data.ECDSAKey)
	testListTarget(t, data.ECDSAKey)
	testListTargetWithDelegates(t, data.ECDSAKey)
	if !testing.Short() {
		testListEmptyTargets(t, data.RSAKey)
		testListTarget(t, data.RSAKey)
		testListTargetWithDelegates(t, data.RSAKey)
	}
}

func testListEmptyTargets(t *testing.T, rootType string) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	_, err := repo.ListTargets(data.CanonicalTargetsRole)
	require.Error(t, err) // no trust data
}

// reads data from the repository in order to fake data being served via
// the ServeMux.
func fakeServerData(t *testing.T, repo *NotaryRepository, mux *http.ServeMux,
	keys map[string]data.PrivateKey) {

	timestampKey, ok := keys[data.CanonicalTimestampRole.String()]
	require.True(t, ok)
	// Add timestamp key via the server's cryptoservice so it can sign
	repo.CryptoService.AddKey(data.CanonicalTimestampRole, repo.gun, timestampKey)

	savedTUFRepo := repo.tufRepo // in case this is overwritten

	rootJSONFile := filepath.Join(repo.baseDir, "tuf",
		filepath.FromSlash(repo.gun.String()), "metadata", "root.json")
	rootFileBytes, err := ioutil.ReadFile(rootJSONFile)

	signedTargets, err := savedTUFRepo.SignTargets(
		"targets", data.DefaultExpires("targets"))
	require.NoError(t, err)

	signedLevel1, err := savedTUFRepo.SignTargets(
		"targets/level1",
		data.DefaultExpires(data.CanonicalTargetsRole),
	)
	if _, ok := savedTUFRepo.Targets["targets/level1"]; ok {
		require.NoError(t, err)
	}

	signedLevel2, err := savedTUFRepo.SignTargets(
		"targets/level2",
		data.DefaultExpires(data.CanonicalTargetsRole),
	)
	if _, ok := savedTUFRepo.Targets["targets/level2"]; ok {
		require.NoError(t, err)
	}

	nested, err := savedTUFRepo.SignTargets(
		"targets/level1/level2",
		data.DefaultExpires(data.CanonicalTargetsRole),
	)

	if _, ok := savedTUFRepo.Targets["targets/level1/level2"]; ok {
		require.NoError(t, err)
	}

	signedSnapshot, err := savedTUFRepo.SignSnapshot(
		data.DefaultExpires("snapshot"))
	require.NoError(t, err)

	signedTimestamp, err := savedTUFRepo.SignTimestamp(
		data.DefaultExpires("timestamp"))
	require.NoError(t, err)

	timestampJSON, _ := json.Marshal(signedTimestamp)
	snapshotJSON, _ := json.Marshal(signedSnapshot)
	targetsJSON, _ := json.Marshal(signedTargets)
	level1JSON, _ := json.Marshal(signedLevel1)
	level2JSON, _ := json.Marshal(signedLevel2)
	nestedJSON, _ := json.Marshal(nested)

	cksmBytes := sha256.Sum256(rootFileBytes)
	rootChecksum := hex.EncodeToString(cksmBytes[:])

	cksmBytes = sha256.Sum256(snapshotJSON)
	snapshotChecksum := hex.EncodeToString(cksmBytes[:])

	cksmBytes = sha256.Sum256(targetsJSON)
	targetsChecksum := hex.EncodeToString(cksmBytes[:])

	cksmBytes = sha256.Sum256(level1JSON)
	level1Checksum := hex.EncodeToString(cksmBytes[:])

	cksmBytes = sha256.Sum256(level2JSON)
	level2Checksum := hex.EncodeToString(cksmBytes[:])

	cksmBytes = sha256.Sum256(nestedJSON)
	nestedChecksum := hex.EncodeToString(cksmBytes[:])

	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/root.json",
		func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, err)
			fmt.Fprint(w, string(rootFileBytes))
		})
	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/root."+rootChecksum+".json",
		func(w http.ResponseWriter, r *http.Request) {
			require.NoError(t, err)
			fmt.Fprint(w, string(rootFileBytes))
		})

	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/timestamp.json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(timestampJSON))
		})

	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/snapshot.json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(snapshotJSON))
		})
	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/snapshot."+snapshotChecksum+".json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(snapshotJSON))
		})

	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/targets.json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(targetsJSON))
		})
	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/targets."+targetsChecksum+".json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(targetsJSON))
		})

	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/targets/level1.json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(level1JSON))
		})
	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/targets/level1."+level1Checksum+".json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(level1JSON))
		})

	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/targets/level2.json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(level2JSON))
		})
	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/targets/level2."+level2Checksum+".json",
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, string(level2JSON))
		})
	mux.HandleFunc("/v2/docker.com/notary/_trust/tuf/targets/level1/level2."+nestedChecksum+".json",
		func(w http.ResponseWriter, r *http.Request) {
			level2JSON, err := json.Marshal(nested)
			require.NoError(t, err)
			fmt.Fprint(w, string(level2JSON))
		})
}

// We want to sort by name, so we can guarantee ordering.
type targetSorter []*TargetWithRole

func (k targetSorter) Len() int           { return len(k) }
func (k targetSorter) Swap(i, j int)      { k[i], k[j] = k[j], k[i] }
func (k targetSorter) Less(i, j int) bool { return k[i].Name < k[j].Name }

func testListTarget(t *testing.T, rootType string) {
	ts, mux, keys := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	// tests need to manually bootstrap timestamp as client doesn't generate it
	err := repo.tufRepo.InitTimestamp()
	require.NoError(t, err, "error creating repository: %s", err)

	latestTarget := addTarget(t, repo, "latest", "../fixtures/intermediate-ca.crt")
	currentTarget := addTarget(t, repo, "current", "../fixtures/intermediate-ca.crt")

	// Apply the changelist. Normally, this would be done by Publish

	// load the changelist for this repo
	cl, err := changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")

	// apply the changelist to the repo
	err = applyChangelist(repo.tufRepo, nil, cl)
	require.NoError(t, err, "could not apply changelist")

	fakeServerData(t, repo, mux, keys)

	targets, err := repo.ListTargets(data.CanonicalTargetsRole)
	require.NoError(t, err)

	// Should be two targets
	require.Len(t, targets, 2, "unexpected number of targets returned by ListTargets")

	sort.Stable(targetSorter(targets))

	// the targets should both be found in the targets role
	for _, foundTarget := range targets {
		require.Equal(t, data.CanonicalTargetsRole, foundTarget.Role)
	}

	// current should be first
	require.True(t, reflect.DeepEqual(*currentTarget, targets[0].Target), "current target does not match")
	require.True(t, reflect.DeepEqual(*latestTarget, targets[1].Target), "latest target does not match")

	// Also test GetTargetByName
	newLatestTarget, err := repo.GetTargetByName("latest")
	require.NoError(t, err)
	require.Equal(t, data.CanonicalTargetsRole, newLatestTarget.Role)
	require.True(t, reflect.DeepEqual(*latestTarget, newLatestTarget.Target), "latest target does not match")

	newCurrentTarget, err := repo.GetTargetByName("current")
	require.NoError(t, err)
	require.Equal(t, data.CanonicalTargetsRole, newCurrentTarget.Role)
	require.True(t, reflect.DeepEqual(*currentTarget, newCurrentTarget.Target), "current target does not match")
}

func testListTargetWithDelegates(t *testing.T, rootType string) {
	ts, mux, keys := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	// tests need to manually bootstrap timestamp as client doesn't generate it
	err := repo.tufRepo.InitTimestamp()
	require.NoError(t, err, "error creating repository: %s", err)

	latestTarget := addTarget(t, repo, "latest", "../fixtures/intermediate-ca.crt")
	currentTarget := addTarget(t, repo, "current", "../fixtures/intermediate-ca.crt")

	// setup delegated targets/level1 role
	k, err := repo.CryptoService.Create("targets/level1", repo.gun, rootType)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationKeys("targets/level1", []data.PublicKey{k}, []string{}, 1)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1", []string{""}, []string{}, false)
	require.NoError(t, err)
	delegatedTarget := addTarget(t, repo, "current", "../fixtures/root-ca.crt", "targets/level1")
	otherTarget := addTarget(t, repo, "other", "../fixtures/root-ca.crt", "targets/level1")

	// setup delegated targets/level2 role
	k, err = repo.CryptoService.Create("targets/level2", repo.gun, rootType)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationKeys("targets/level2", []data.PublicKey{k}, []string{}, 1)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level2", []string{""}, []string{}, false)
	require.NoError(t, err)
	// this target should not show up as the one in targets/level1 takes higher priority
	_ = addTarget(t, repo, "current", "../fixtures/notary-server.crt", "targets/level2")
	// this target should show up as the name doesn't exist elsewhere
	level2Target := addTarget(t, repo, "level2", "../fixtures/notary-server.crt", "targets/level2")

	// Apply the changelist. Normally, this would be done by Publish

	// load the changelist for this repo
	cl, err := changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")

	// apply the changelist to the repo, then clear it
	err = applyChangelist(repo.tufRepo, nil, cl)
	require.NoError(t, err, "could not apply changelist")
	require.NoError(t, cl.Clear(""))

	_, ok := repo.tufRepo.Targets["targets/level1"].Signed.Targets["current"]
	require.True(t, ok)
	_, ok = repo.tufRepo.Targets["targets/level1"].Signed.Targets["other"]
	require.True(t, ok)
	_, ok = repo.tufRepo.Targets["targets/level2"].Signed.Targets["level2"]
	require.True(t, ok)

	// setup delegated targets/level1/level2 role separately, which can only modify paths prefixed with "level2"
	// This is done separately due to target shadowing
	k, err = repo.CryptoService.Create("targets/level1/level2", repo.gun, rootType)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationKeys("targets/level1/level2", []data.PublicKey{k}, []string{}, 1)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1/level2", []string{"level2"}, []string{}, false)
	require.NoError(t, err)
	nestedTarget := addTarget(t, repo, "level2", "../fixtures/notary-signer.crt", "targets/level1/level2")
	// load the changelist for this repo
	cl, err = changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")
	// apply the changelist to the repo
	err = applyChangelist(repo.tufRepo, nil, cl)
	require.NoError(t, err, "could not apply changelist")
	// check the changelist was applied
	_, ok = repo.tufRepo.Targets["targets/level1/level2"].Signed.Targets["level2"]
	require.True(t, ok)

	fakeServerData(t, repo, mux, keys)

	// test default listing
	targets, err := repo.ListTargets()
	require.NoError(t, err)

	// Should be four targets
	require.Len(t, targets, 4, "unexpected number of targets returned by ListTargets")

	sort.Stable(targetSorter(targets))

	// current should be first.
	require.True(t, reflect.DeepEqual(*currentTarget, targets[0].Target), "current target does not match")
	require.Equal(t, data.CanonicalTargetsRole, targets[0].Role)

	require.True(t, reflect.DeepEqual(*latestTarget, targets[1].Target), "latest target does not match")
	require.Equal(t, data.CanonicalTargetsRole, targets[1].Role)

	// This target shadows the "level2" target in level1/level2
	require.True(t, reflect.DeepEqual(*level2Target, targets[2].Target), "level2 target does not match")
	require.Equal(t, "targets/level2", targets[2].Role.String())

	require.True(t, reflect.DeepEqual(*otherTarget, targets[3].Target), "other target does not match")
	require.Equal(t, "targets/level1", targets[3].Role.String())

	// test listing with priority specified
	targets, err = repo.ListTargets("targets/level1", data.CanonicalTargetsRole)
	require.NoError(t, err)

	// Should be four targets
	require.Len(t, targets, 4, "unexpected number of targets returned by ListTargets")

	sort.Stable(targetSorter(targets))

	// current (in delegated role) should be first
	require.True(t, reflect.DeepEqual(*delegatedTarget, targets[0].Target), "current target does not match")
	require.Equal(t, "targets/level1", string(targets[0].Role))

	require.True(t, reflect.DeepEqual(*latestTarget, targets[1].Target), "latest target does not match")
	require.Equal(t, data.CanonicalTargetsRole, targets[1].Role)

	// Now the level1/level2 target shadows the level2 target
	require.True(t, reflect.DeepEqual(*nestedTarget, targets[2].Target), "level1/level2 target does not match")
	require.Equal(t, "targets/level1/level2", targets[2].Role.String())

	require.True(t, reflect.DeepEqual(*otherTarget, targets[3].Target), "other target does not match")
	require.Equal(t, "targets/level1", targets[3].Role.String())

	// Also test GetTargetByName
	newLatestTarget, err := repo.GetTargetByName("latest")
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(*latestTarget, newLatestTarget.Target), "latest target does not match")
	require.Equal(t, data.CanonicalTargetsRole, newLatestTarget.Role)

	newCurrentTarget, err := repo.GetTargetByName("current", "targets/level1", "targets")
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(*delegatedTarget, newCurrentTarget.Target), "current target does not match")
	require.Equal(t, "targets/level1", newCurrentTarget.Role.String())

	newOtherTarget, err := repo.GetTargetByName("other")
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(*otherTarget, newOtherTarget.Target), "other target does not match")
	require.Equal(t, "targets/level1", newOtherTarget.Role.String())

	newLevel2Target, err := repo.GetTargetByName("level2")
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(*level2Target, newLevel2Target.Target), "level2 target does not match")
	require.Equal(t, "targets/level2", newLevel2Target.Role.String())

	// Shadow by prioritizing level1, but exclude level1/level2, so we should still get targets/level2's level2 target
	newLevel2Target, err = repo.GetTargetByName("level2", "targets/level1", "targets/level2", "targets/level1/level2")
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(*level2Target, newLevel2Target.Target), "level2 target does not match")
	require.Equal(t, "targets/level2", newLevel2Target.Role.String())

	// Prioritize level1 to get level1/level2's level2 target
	newLevel2Target, err = repo.GetTargetByName("level2", "targets/level1")
	require.NoError(t, err)
	require.True(t, reflect.DeepEqual(*nestedTarget, newLevel2Target.Target), "level2 target does not match")
	require.Equal(t, "targets/level1/level2", newLevel2Target.Role.String())
}

func TestListTargetRestrictsDelegationPaths(t *testing.T) {
	ts, mux, keys := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	// tests need to manually bootstrap timestamp as client doesn't generate it
	err := repo.tufRepo.InitTimestamp()
	require.NoError(t, err, "error creating repository: %s", err)

	// setup delegated targets/level1 role
	k, err := repo.CryptoService.Create("targets/level1", repo.gun, data.ECDSAKey)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationKeys("targets/level1", []data.PublicKey{k}, []string{}, 1)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1", []string{""}, []string{}, false)
	require.NoError(t, err)
	addTarget(t, repo, "level1-target", "../fixtures/root-ca.crt", "targets/level1")
	addTarget(t, repo, "incorrectly-named-target", "../fixtures/root-ca.crt", "targets/level1")

	// setup delegated targets/level2 role
	err = repo.tufRepo.UpdateDelegationKeys("targets/level1/level2", []data.PublicKey{k}, []string{}, 1)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1/level2", []string{""}, []string{}, false)
	require.NoError(t, err)
	addTarget(t, repo, "level2-target", "../fixtures/notary-server.crt", "targets/level1/level2")
	addTarget(t, repo, "level1-level2-target", "../fixtures/notary-server.crt", "targets/level1/level2")

	// Apply the changelist. Normally, this would be done by Publish

	// load the changelist for this repo
	cl, err := changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")

	// apply the changelist to the repo
	err = applyChangelist(repo.tufRepo, nil, cl)
	require.NoError(t, err, "could not apply changelist")

	require.NoError(t, cl.Clear(""))

	// Now restrict the paths
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1", []string{"level1"}, []string{}, false)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1/level2", []string{"level1-level2", "level2"}, []string{}, false)
	require.NoError(t, err)

	err = repo.tufRepo.UpdateDelegationPaths("targets/level1", []string{}, []string{""}, false)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1/level2", []string{}, []string{""}, false)
	require.NoError(t, err)

	// load the changelist for this repo
	cl, err = changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")

	// apply the changelist to the repo
	err = applyChangelist(repo.tufRepo, nil, cl)
	require.NoError(t, err, "could not apply changelist")

	fakeServerData(t, repo, mux, keys)

	// test default listing
	targets, err := repo.ListTargets("targets/level1")
	require.NoError(t, err)

	// Should be four targets
	require.Len(t, targets, 2, "unexpected number of targets returned by ListTargets")

	sort.Stable(targetSorter(targets))

	var foundLevel1, foundLevel2 bool

	for _, tgts := range targets {
		switch tgts.Name {
		case "level1-target":
			require.Equal(t, "targets/level1", tgts.Role.String())
			foundLevel1 = true
		case "level1-level2-target":
			require.Equal(t, "targets/level1/level2", tgts.Role.String())
			foundLevel2 = true
		}
	}

	require.True(t, foundLevel1)
	require.True(t, foundLevel2)

	// test GetTargetByName
	tgt, err := repo.GetTargetByName("level1-target", "targets/level1")
	require.NoError(t, err)
	require.NotNil(t, tgt)
	require.EqualValues(t, tgt.Role, "targets/level1")

	tgt, err = repo.GetTargetByName("level1-level2-target", "targets/level1")
	require.NoError(t, err)
	require.NotNil(t, tgt)
	require.EqualValues(t, tgt.Role, "targets/level1/level2")

	tgt, err = repo.GetTargetByName("level2-target", "targets/level1/level2")
	require.Error(t, err)
	require.Nil(t, tgt)
}

// TestValidateRootKey verifies that the public data in root.json for the root
// key is a valid x509 certificate.
func TestValidateRootKey(t *testing.T) {
	testValidateRootKey(t, data.ECDSAKey)
	if !testing.Short() {
		testValidateRootKey(t, data.RSAKey)
	}
}

func testValidateRootKey(t *testing.T, rootType string) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	rootJSONFile := filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()),
		"metadata", "root.json")

	jsonBytes, err := ioutil.ReadFile(rootJSONFile)
	require.NoError(t, err, "error reading TUF metadata file %s: %s", rootJSONFile, err)

	var decoded data.Signed
	err = json.Unmarshal(jsonBytes, &decoded)
	require.NoError(t, err, "error parsing TUF metadata file %s: %s", rootJSONFile, err)

	var decodedRoot data.Root
	err = json.Unmarshal(*decoded.Signed, &decodedRoot)
	require.NoError(t, err, "error parsing root.json signed section: %s", err)

	keyids := []string{}
	for role, roleData := range decodedRoot.Roles {
		if role == data.CanonicalRootRole {
			keyids = append(keyids, roleData.KeyIDs...)
		}
	}
	require.NotEmpty(t, keyids)

	for _, keyid := range keyids {
		key, ok := decodedRoot.Keys[keyid]
		require.True(t, ok, "key id not found in keys")
		_, err := utils.LoadCertFromPEM(key.Public())
		require.NoError(t, err, "key is not a valid cert")
	}
}

// TestGetChangelist ensures that the changelist returned matches the changes
// added.
// We test this with both an RSA and ECDSA root key
func TestGetChangelist(t *testing.T) {
	testGetChangelist(t, data.ECDSAKey)
	if !testing.Short() {
		testGetChangelist(t, data.RSAKey)
	}
}

func testGetChangelist(t *testing.T, rootType string) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	require.Len(t, getChanges(t, repo), 0, "No changes should be in changelist yet")

	// Create 2 targets
	addTarget(t, repo, "latest", "../fixtures/intermediate-ca.crt")
	addTarget(t, repo, "current", "../fixtures/intermediate-ca.crt")

	// Test loading changelist
	chgs := getChanges(t, repo)
	require.Len(t, chgs, 2, "Wrong number of changes returned from changelist")

	changes := make(map[string]changelist.Change)
	for _, ch := range chgs {
		changes[ch.Path()] = ch
	}

	currentChange := changes["current"]
	require.NotNil(t, currentChange, "Expected changelist to contain a change for path 'current'")
	require.EqualValues(t, changelist.ActionCreate, currentChange.Action())
	require.EqualValues(t, "targets", currentChange.Scope())
	require.EqualValues(t, "target", currentChange.Type())
	require.EqualValues(t, "current", currentChange.Path())

	latestChange := changes["latest"]
	require.NotNil(t, latestChange, "Expected changelist to contain a change for path 'latest'")
	require.EqualValues(t, changelist.ActionCreate, latestChange.Action())
	require.EqualValues(t, "targets", latestChange.Scope())
	require.EqualValues(t, "target", latestChange.Type())
	require.EqualValues(t, "latest", latestChange.Path())
}

// Create a repo, instantiate a notary server, and publish the bare repo to the
// server, signing all the non-timestamp metadata.  Root, targets, and snapshots
// (if locally signing) should be sent.
func TestPublishBareRepo(t *testing.T) {
	testPublishNoData(t, data.ECDSAKey, false, true)
	testPublishNoData(t, data.ECDSAKey, false, false)
	testPublishNoData(t, data.ECDSAKey, true, true)
	testPublishNoData(t, data.ECDSAKey, true, false)
	if !testing.Short() {
		testPublishNoData(t, data.RSAKey, false, true)
		testPublishNoData(t, data.RSAKey, false, false)
		testPublishNoData(t, data.RSAKey, true, true)
		testPublishNoData(t, data.RSAKey, true, false)
	}
}

func testPublishNoData(t *testing.T, rootType string, clearCache, serverManagesSnapshot bool) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo1, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL,
		serverManagesSnapshot)
	defer os.RemoveAll(repo1.baseDir)

	var rec *passRoleRecorder
	if clearCache {
		rec = newRoleRecorder()
		repo1, rec = newRepoToTestRepo(t, repo1, false)
	}

	require.NoError(t, repo1.Publish())

	if clearCache {
		// signing is only done by the target/snapshot keys
		rec.requireCreated(t, nil)
		if serverManagesSnapshot {
			rec.requireAsked(t, []string{data.CanonicalTargetsRole.String()})
		} else {
			rec.requireAsked(t, []string{data.CanonicalTargetsRole.String(), data.CanonicalSnapshotRole.String()})
		}
	}

	// use another repo to check metadata
	repo2, _ := newRepoToTestRepo(t, repo1, true)
	defer os.RemoveAll(repo2.baseDir)

	targets, err := repo2.ListTargets()
	require.NoError(t, err)
	require.Empty(t, targets)

	for _, role := range data.BaseRoles {
		// we don't cache timstamp metadata
		if role != data.CanonicalTimestampRole {
			requireRepoHasExpectedMetadata(t, repo2, role, true)
		}
	}
}

// Publishing an uninitialized repo will fail, but initializing and republishing
// after should succeed
func TestPublishUninitializedRepo(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	ts := fullTestServer(t)
	defer ts.Close()

	// uninitialized repo should fail to publish
	tempBaseDir, err := ioutil.TempDir("", "notary-tests")
	require.NoError(t, err)
	defer os.RemoveAll(tempBaseDir)

	repo, err := NewFileCachedNotaryRepository(tempBaseDir, gun, ts.URL,
		http.DefaultTransport, passphraseRetriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repository: %s", err)
	err = repo.Publish()
	require.Error(t, err)

	// no metadata created
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalRootRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTargetsRole, false)

	// now, initialize and republish in the same directory
	rootPubKey, err := repo.CryptoService.Create(data.CanonicalRootRole, repo.gun, data.ECDSAKey)
	require.NoError(t, err, "error generating root key: %s", err)

	require.NoError(t, repo.Initialize([]string{rootPubKey.ID()}))

	// now metadata is created
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalRootRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTargetsRole, true)

	require.NoError(t, repo.Publish())
}

// Create a repo, instantiate a notary server, and publish the repo with
// some targets to the server, signing all the non-timestamp metadata.
// We test this with both an RSA and ECDSA root key
func TestPublishClientHasSnapshotKey(t *testing.T) {
	testPublishWithData(t, data.ECDSAKey, true, false)
	testPublishWithData(t, data.ECDSAKey, false, false)
	if !testing.Short() {
		testPublishWithData(t, data.RSAKey, true, false)
		testPublishWithData(t, data.RSAKey, false, false)
	}
}

// Create a repo, instantiate a notary server (designating the server as the
// snapshot signer) , and publish the repo with some targets to the server,
// signing the root and targets metadata only.  The server should sign just fine.
// We test this with both an RSA and ECDSA root key
func TestPublishAfterInitServerHasSnapshotKey(t *testing.T) {
	testPublishWithData(t, data.ECDSAKey, true, true)
	testPublishWithData(t, data.ECDSAKey, false, true)
	if !testing.Short() {
		testPublishWithData(t, data.RSAKey, true, true)
		testPublishWithData(t, data.RSAKey, false, true)
	}
}

func testPublishWithData(t *testing.T, rootType string, clearCache, serverManagesSnapshot bool) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL,
		serverManagesSnapshot)
	defer os.RemoveAll(repo.baseDir)

	var rec *passRoleRecorder
	if clearCache {
		rec = newRoleRecorder()
		repo, rec = newRepoToTestRepo(t, repo, false)
	}

	requirePublishToRolesSucceeds(t, repo, nil, []data.RoleName{data.CanonicalTargetsRole})

	if clearCache {
		// signing is only done by the target/snapshot keys
		rec.requireCreated(t, nil)
		if serverManagesSnapshot {
			rec.requireAsked(t, []string{data.CanonicalTargetsRole.String()})
		} else {
			rec.requireAsked(t, []string{data.CanonicalTargetsRole.String(), data.CanonicalSnapshotRole.String()})
		}
	}
}

// requires that adding to the given roles results in the targets actually being
// added only to the expected roles and no others
func requirePublishToRolesSucceeds(t *testing.T, repo1 *NotaryRepository,
	publishToRoles []data.RoleName, expectedPublishedRoles []data.RoleName) {

	// were there unpublished changes before?
	changesOffset := len(getChanges(t, repo1))

	// Create 2 targets - (actually 3, but we delete 1)
	addTarget(t, repo1, "toDelete", "../fixtures/intermediate-ca.crt", publishToRoles...)
	latestTarget := addTarget(
		t, repo1, "latest", "../fixtures/intermediate-ca.crt", publishToRoles...)
	currentTarget := addTarget(
		t, repo1, "current", "../fixtures/intermediate-ca.crt", publishToRoles...)
	repo1.RemoveTarget("toDelete", publishToRoles...)

	// if no roles are provided, the default role is target
	numRoles := int(math.Max(1, float64(len(publishToRoles))))
	require.Len(t, getChanges(t, repo1), changesOffset+4*numRoles,
		"wrong number of changelist files found")

	// Now test Publish
	err := repo1.Publish()
	require.NoError(t, err)
	require.Len(t, getChanges(t, repo1), 0, "wrong number of changelist files found")

	// use another repo to check metadata
	repo2, _ := newRepoToTestRepo(t, repo1, true)
	defer os.RemoveAll(repo2.baseDir)

	// Should be two targets per role
	for _, role := range expectedPublishedRoles {
		for _, repo := range []*NotaryRepository{repo1, repo2} {
			targets, err := repo.ListTargets(role)
			require.NoError(t, err)

			require.Len(t, targets, 2,
				"unexpected number of targets returned by ListTargets(%s)", role)

			sort.Stable(targetSorter(targets))

			require.True(t, reflect.DeepEqual(*currentTarget, targets[0].Target), "current target does not match")
			require.EqualValues(t, role, targets[0].Role)
			require.True(t, reflect.DeepEqual(*latestTarget, targets[1].Target), "latest target does not match")
			require.EqualValues(t, role, targets[1].Role)

			// Also test GetTargetByName
			newLatestTarget, err := repo.GetTargetByName("latest", role)
			require.NoError(t, err)
			require.True(t, reflect.DeepEqual(*latestTarget, newLatestTarget.Target), "latest target does not match")
			require.EqualValues(t, role, newLatestTarget.Role)

			newCurrentTarget, err := repo.GetTargetByName("current", role)
			require.NoError(t, err)
			require.True(t, reflect.DeepEqual(*currentTarget, newCurrentTarget.Target), "current target does not match")
			require.EqualValues(t, role, newCurrentTarget.Role)
		}
	}
}

// After pulling a repo from the server, so there is a snapshots metadata file,
// push a different target to the server (the server is still the snapshot
// signer).  The server should sign just fine.
// We test this with both an RSA and ECDSA root key
func TestPublishAfterPullServerHasSnapshotKey(t *testing.T) {
	testPublishAfterPullServerHasSnapshotKey(t, data.ECDSAKey)
	if !testing.Short() {
		testPublishAfterPullServerHasSnapshotKey(t, data.RSAKey)
	}
}

func testPublishAfterPullServerHasSnapshotKey(t *testing.T, rootType string) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL, true)
	defer os.RemoveAll(repo.baseDir)
	// no timestamp metadata because that comes from the server
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTimestampRole, false)
	// no snapshot metadata because that comes from the server
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, false)

	// Publish something
	published := addTarget(t, repo, "v1", "../fixtures/intermediate-ca.crt")
	require.NoError(t, repo.Publish())

	// still no timestamp or snapshot metadata info
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTimestampRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, false)

	// list, so that the snapshot metadata is pulled from server
	targets, err := repo.ListTargets(data.CanonicalTargetsRole)
	require.NoError(t, err)
	require.Equal(t, []*TargetWithRole{{Target: *published, Role: data.CanonicalTargetsRole}}, targets)
	// listing downloaded the timestamp and snapshot metadata info
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTimestampRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, true)

	// Publish again should succeed
	addTarget(t, repo, "v2", "../fixtures/intermediate-ca.crt")
	err = repo.Publish()
	require.NoError(t, err)
}

// If neither the client nor the server has the snapshot key, signing will fail
// with an ErrNoKeys error.
// We test this with both an RSA and ECDSA root key
func TestPublishNoOneHasSnapshotKey(t *testing.T) {
	testPublishNoOneHasSnapshotKey(t, data.ECDSAKey)
	if !testing.Short() {
		testPublishNoOneHasSnapshotKey(t, data.RSAKey)
	}
}

func testPublishNoOneHasSnapshotKey(t *testing.T, rootType string) {
	ts := fullTestServer(t)
	defer ts.Close()

	// create repo and delete the snapshot key and metadata
	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	snapshotRole, ok := repo.tufRepo.Root.Signed.Roles[data.CanonicalSnapshotRole]
	require.True(t, ok)
	for _, keyID := range snapshotRole.KeyIDs {
		repo.CryptoService.RemoveKey(keyID)
	}

	// ensure that the cryptoservice no longer has any snapshot keys
	require.Len(t, repo.CryptoService.ListKeys(data.CanonicalSnapshotRole), 0)

	// Publish something
	addTarget(t, repo, "v1", "../fixtures/intermediate-ca.crt")
	err := repo.Publish()
	require.Error(t, err)
	require.IsType(t, validation.ErrBadHierarchy{}, err)
}

// If the snapshot metadata is corrupt or the snapshot metadata is unreadable,
// we can't publish for the first time (whether the client or server has the
// snapshot key), because there is no existing data for us to download. If the
// repo has already been published, it doesn't matter if the metadata is corrupt
// because we can just redownload if it is.
func TestPublishSnapshotCorrupt(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// do not publish first - publish should fail with corrupt snapshot data even with server signing snapshot
	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary1", ts.URL, true)
	defer os.RemoveAll(repo.baseDir)
	testPublishBadMetadata(t, data.CanonicalSnapshotRole.String(), repo, false, false)

	// do not publish first - publish should fail with corrupt snapshot data with local snapshot signing
	repo, _ = initializeRepo(t, data.ECDSAKey, "docker.com/notary2", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	testPublishBadMetadata(t, data.CanonicalSnapshotRole.String(), repo, false, false)

	// publish first - publish again should succeed despite corrupt snapshot data (server signing snapshot)
	repo, _ = initializeRepo(t, data.ECDSAKey, "docker.com/notary3", ts.URL, true)
	defer os.RemoveAll(repo.baseDir)
	testPublishBadMetadata(t, data.CanonicalSnapshotRole.String(), repo, true, true)

	// publish first - publish again should succeed despite corrupt snapshot data (local snapshot signing)
	repo, _ = initializeRepo(t, data.ECDSAKey, "docker.com/notary4", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	testPublishBadMetadata(t, data.CanonicalSnapshotRole.String(), repo, true, true)
}

// If the targets metadata is corrupt or the targets metadata is unreadable,
// we can't publish for the first time, because there is no existing data for.
// us to download. If the repo has already been published, it doesn't matter
// if the metadata is corrupt because we can just redownload if it is.
func TestPublishTargetsCorrupt(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// do not publish first - publish should fail with corrupt snapshot data
	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary1", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	testPublishBadMetadata(t, data.CanonicalTargetsRole.String(), repo, false, false)

	// publish first - publish again should succeed despite corrupt snapshot data
	repo, _ = initializeRepo(t, data.ECDSAKey, "docker.com/notary2", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	testPublishBadMetadata(t, data.CanonicalTargetsRole.String(), repo, true, true)
}

// If the root metadata is corrupt or the root metadata is unreadable,
// we can't publish for the first time.  If there is already a remote root,
// we just download that and verify (using our trusted certificate trust
// anchors) that it is signed with the same keys, and if so, we just use the
// remote root.
func TestPublishRootCorrupt(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// do not publish first - publish should fail with corrupt snapshot data
	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary1", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	testPublishBadMetadata(t, data.CanonicalRootRole.String(), repo, false, false)

	// publish first - publish should still fail if the local root is corrupt since
	// we can't determine whether remote root is signed with the same key.
	repo, _ = initializeRepo(t, data.ECDSAKey, "docker.com/notary2", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	testPublishBadMetadata(t, data.CanonicalRootRole.String(), repo, true, false)
}

// When publishing snapshot, root, or target, if the repo hasn't been published
// before, if the metadata is corrupt, it can't be published.
func testPublishBadMetadata(t *testing.T, roleName string, repo *NotaryRepository,
	publishFirst, succeeds bool) {

	if publishFirst {
		require.NoError(t, repo.Publish())
	}

	addTarget(t, repo, "v1", "../fixtures/intermediate-ca.crt")

	// readable, but corrupt file
	repo.cache.Set(roleName, []byte("this isn't JSON"))
	err := repo.Publish()
	if succeeds {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
		require.IsType(t, &json.SyntaxError{}, err)
	}

	// make an unreadable file by creating a directory instead of a file
	path := fmt.Sprintf("%s.%s",
		filepath.Join(repo.baseDir, tufDir, filepath.FromSlash(repo.gun.String()),
			"metadata", roleName), "json")
	os.RemoveAll(path)
	require.NoError(t, os.Mkdir(path, 0755))
	defer os.RemoveAll(path)

	err = repo.Publish()
	if succeeds || publishFirst {
		require.NoError(t, err)
	} else {
		require.Error(t, err)
		require.IsType(t, &os.PathError{}, err)
	}
}

// If the repo is not initialized, calling repo.Publish() should return ErrRepoNotInitialized
func TestNotInitializedOnPublish(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	gun := "docker.com/notary"
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _, _ := createRepoAndKey(t, data.ECDSAKey, tempBaseDir, gun, ts.URL)

	addTarget(t, repo, "v1", "../fixtures/intermediate-ca.crt")

	err = repo.Publish()
	require.Error(t, err)
	require.IsType(t, ErrRepoNotInitialized{}, err)
}

type cannotCreateKeys struct {
	signed.CryptoService
}

func (cs cannotCreateKeys) Create(_ data.RoleName, _ data.GUN, _ string) (data.PublicKey, error) {
	return nil, fmt.Errorf("Oh no I cannot create keys")
}

// If there is an error creating the local keys, no call is made to get a
// remote key.
func TestPublishSnapshotLocalKeysCreatedFirst(t *testing.T) {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	defer os.RemoveAll(tempBaseDir)
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	var gun data.GUN = "docker.com/notary"

	requestMade := false
	ts := httptest.NewServer(http.HandlerFunc(
		func(http.ResponseWriter, *http.Request) { requestMade = true }))
	defer ts.Close()

	repo, err := NewFileCachedNotaryRepository(
		tempBaseDir, gun, ts.URL, http.DefaultTransport, passphraseRetriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	cs := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(passphraseRetriever))

	rootPubKey, err := cs.Create(data.CanonicalRootRole, gun, data.ECDSAKey)
	require.NoError(t, err, "error generating root key: %s", err)

	repo.CryptoService = cannotCreateKeys{CryptoService: cs}

	err = repo.Initialize([]string{rootPubKey.ID()}, data.CanonicalSnapshotRole)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Oh no I cannot create keys")
	require.False(t, requestMade)
}

func createKey(t *testing.T, repo *NotaryRepository, role data.RoleName, x509 bool) data.PublicKey {
	key, err := repo.CryptoService.Create(role, repo.gun, data.ECDSAKey)
	require.NoError(t, err, "error creating key")

	if x509 {
		start := time.Now().AddDate(0, 0, -1)
		privKey, _, err := repo.CryptoService.GetPrivateKey(key.ID())
		require.NoError(t, err)
		cert, err := cryptoservice.GenerateCertificate(
			privKey, data.GUN(role), start, start.AddDate(1, 0, 0),
		)
		require.NoError(t, err)
		return data.NewECDSAx509PublicKey(utils.CertToPEM(cert))
	}
	return key
}

// Publishing delegations works so long as the delegation parent exists by the
// time that delegation addition change is applied.  Most of the tests for
// applying delegation changes in in helpers_test.go (applyTargets tests), so
// this is just a sanity test to make sure Publish calls it correctly
func TestPublishDelegations(t *testing.T) {
	testPublishDelegations(t, true, false)
	testPublishDelegations(t, false, false)
}

func TestPublishDelegationsX509(t *testing.T) {
	testPublishDelegations(t, true, true)
	testPublishDelegations(t, false, true)
}

func testPublishDelegations(t *testing.T, clearCache, x509Keys bool) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo1, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo1.baseDir)

	delgKey := createKey(t, repo1, "targets/a", x509Keys)

	// This should publish fine, even though targets/a/b is dependent upon
	// targets/a, because these should execute in order
	for _, delgName := range []data.RoleName{"targets/a", "targets/a/b", "targets/c"} {
		require.NoError(t,
			repo1.AddDelegation(delgName, []data.PublicKey{delgKey}, []string{""}),
			"error creating delegation")
	}
	require.Len(t, getChanges(t, repo1), 6, "wrong number of changelist files found")

	var rec *passRoleRecorder
	if clearCache {
		repo1, rec = newRepoToTestRepo(t, repo1, false)
	}

	require.NoError(t, repo1.Publish())
	require.Len(t, getChanges(t, repo1), 0, "wrong number of changelist files found")

	if clearCache {
		// when publishing, only the parents of the delegations created need to be signed
		// (and snapshot)
		rec.requireAsked(t, []string{data.CanonicalTargetsRole.String(), "targets/a", data.CanonicalSnapshotRole.String()})
		rec.clear()
	}

	// this should not publish, because targets/z doesn't exist
	require.NoError(t,
		repo1.AddDelegation("targets/z/y", []data.PublicKey{delgKey}, []string{""}),
		"error creating delegation")
	require.Len(t, getChanges(t, repo1), 2, "wrong number of changelist files found")
	require.Error(t, repo1.Publish())
	require.Len(t, getChanges(t, repo1), 2, "wrong number of changelist files found")

	if clearCache {
		rec.requireAsked(t, nil)
	}

	// use another repo to check metadata
	repo2, _ := newRepoToTestRepo(t, repo1, false)
	defer os.RemoveAll(repo2.baseDir)

	// pull
	_, err := repo2.ListTargets()
	require.NoError(t, err, "unable to pull repo")

	for _, repo := range []*NotaryRepository{repo1, repo2} {
		// targets should have delegations targets/a and targets/c
		targets := repo.tufRepo.Targets[data.CanonicalTargetsRole]
		require.Len(t, targets.Signed.Delegations.Roles, 2)
		require.Len(t, targets.Signed.Delegations.Keys, 1)

		_, ok := targets.Signed.Delegations.Keys[delgKey.ID()]
		require.True(t, ok)

		foundRoleNames := make(map[data.RoleName]bool)
		for _, r := range targets.Signed.Delegations.Roles {
			foundRoleNames[r.Name] = true
		}
		require.True(t, foundRoleNames["targets/a"])
		require.True(t, foundRoleNames["targets/c"])

		// targets/a should have delegation targets/a/b only
		a := repo.tufRepo.Targets["targets/a"]
		require.Len(t, a.Signed.Delegations.Roles, 1)
		require.Len(t, a.Signed.Delegations.Keys, 1)

		_, ok = a.Signed.Delegations.Keys[delgKey.ID()]
		require.True(t, ok)

		require.EqualValues(t, "targets/a/b", a.Signed.Delegations.Roles[0].Name)
	}
}

// If a changelist specifies a particular role to push targets to, and there
// is a role but no key, publish should just fail outright.
func TestPublishTargetsDelegationScopeFailIfNoKeys(t *testing.T) {
	testPublishTargetsDelegationScopeFailIfNoKeys(t, true)
	testPublishTargetsDelegationScopeFailIfNoKeys(t, false)
}

func testPublishTargetsDelegationScopeFailIfNoKeys(t *testing.T, clearCache bool) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	// generate a key that isn't in the cryptoservice, so we can't sign this
	// one
	aPrivKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, "error generating key that is not in our cryptoservice")
	aPubKey := data.PublicKeyFromPrivate(aPrivKey)

	var rec *passRoleRecorder
	if clearCache {
		repo, rec = newRepoToTestRepo(t, repo, false)
	}

	// ensure that the role exists
	require.NoError(t, repo.AddDelegation("targets/a", []data.PublicKey{aPubKey}, []string{""}))
	require.NoError(t, repo.Publish())

	if clearCache {
		rec.requireAsked(t, []string{data.CanonicalTargetsRole.String(), data.CanonicalSnapshotRole.String()})
		rec.clear()
	}

	// add a target to targets/a/b - no role b, so it falls back on a, which
	// exists but there is no signing key for
	addTarget(t, repo, "latest", "../fixtures/intermediate-ca.crt", "targets/a/b")
	require.Len(t, getChanges(t, repo), 1, "wrong number of changelist files found")

	// Now Publish should fail
	require.Error(t, repo.Publish())
	require.Len(t, getChanges(t, repo), 1, "wrong number of changelist files found")
	if clearCache {
		rec.requireAsked(t, nil)
		rec.clear()
	}

	targets, err := repo.ListTargets("targets", "targets/a", "targets/a/b")
	require.NoError(t, err)
	require.Empty(t, targets)
}

// If a changelist specifies a particular role to push targets to, and such
// a role and the keys are present, publish will write to that role only, and
// not its parents.  This tests the case where the local machine knows about
// all the roles (in fact, the role creations will be applied before the
// targets)
func TestPublishTargetsDelegationSuccessLocallyHasRoles(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	for _, delgName := range []data.RoleName{"targets/a", "targets/a/b"} {
		delgKey := createKey(t, repo, delgName, false)
		require.NoError(t,
			repo.AddDelegation(delgName, []data.PublicKey{delgKey}, []string{""}),
			"error creating delegation")
	}

	// just always check signing now, we've already established we can publish
	// delegations with and without the metadata and key cache
	var rec *passRoleRecorder
	repo, rec = newRepoToTestRepo(t, repo, false)

	requirePublishToRolesSucceeds(t, repo, []data.RoleName{"targets/a/b"}, []data.RoleName{"targets/a/b"})

	// first time publishing, so everything gets signed
	rec.requireAsked(t, []string{data.CanonicalTargetsRole.String(), "targets/a", "targets/a/b",
		data.CanonicalSnapshotRole.String()})
}

// If a changelist specifies a particular role to push targets to, and the role
// is present, publish will write to that role only.  The targets keys are not
// needed.
func TestPublishTargetsDelegationNoTargetsKeyNeeded(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	for _, delgName := range []data.RoleName{"targets/a", "targets/a/b"} {
		delgKey := createKey(t, repo, delgName, false)
		require.NoError(t,
			repo.AddDelegation(delgName, []data.PublicKey{delgKey}, []string{""}),
			"error creating delegation")
	}

	// just always check signing now, we've already established we can publish
	// delegations with and without the metadata and key cache
	var rec *passRoleRecorder
	repo, rec = newRepoToTestRepo(t, repo, false)

	require.NoError(t, repo.Publish())
	// first time publishing, so all delegation parents get signed
	rec.requireAsked(t, []string{data.CanonicalTargetsRole.String(), "targets/a", data.CanonicalSnapshotRole.String()})
	rec.clear()

	// remove targets key - it is not even needed
	targetsKeys := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)
	require.Len(t, targetsKeys, 1)
	require.NoError(t, repo.CryptoService.RemoveKey(targetsKeys[0]))

	requirePublishToRolesSucceeds(t, repo,
		[]data.RoleName{"targets/a/b"}, []data.RoleName{"targets/a/b"})

	// only the target delegation gets signed - snapshot key has already been cached
	rec.requireAsked(t, []string{"targets/a/b"})
}

// If a changelist specifies a particular role to push targets to, and is such
// a role and the keys are present, publish will write to that role only, and
// not its parents.  Tests:
// - case where the local doesn't know about all the roles, and has to download
//   them before publish.
// - owner of a repo may not have the delegated keys, so can't sign a delegated
//   role
func TestPublishTargetsDelegationSuccessNeedsToDownloadRoles(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	ts := fullTestServer(t)
	defer ts.Close()

	// this is the original repo - it owns the root/targets keys and creates
	// the delegation to which it doesn't have the key (so server snapshot
	// signing would be required)
	ownerRepo, _ := initializeRepo(t, data.ECDSAKey, gun.String(), ts.URL, true)
	defer os.RemoveAll(ownerRepo.baseDir)

	// this is a user, or otherwise a repo that only has access to the delegation
	// key so it can publish targets to the delegated role
	delgRepo, _ := newRepoToTestRepo(t, ownerRepo, true)
	defer os.RemoveAll(delgRepo.baseDir)

	// create a key on the owner repo
	aKey, err := ownerRepo.CryptoService.Create("targets/a", gun, data.ECDSAKey)
	require.NoError(t, err, "error creating delegation key")

	// create a key on the delegated repo
	bKey, err := delgRepo.CryptoService.Create("targets/a/b", gun, data.ECDSAKey)
	require.NoError(t, err, "error creating delegation key")

	// clear metadata and unencrypted private key cache
	var ownerRec, delgRec *passRoleRecorder
	ownerRepo, ownerRec = newRepoToTestRepo(t, ownerRepo, false)
	delgRepo, delgRec = newRepoToTestRepo(t, delgRepo, false)

	// owner creates delegations, adds the delegated key to them, and publishes them
	require.NoError(t,
		ownerRepo.AddDelegation("targets/a", []data.PublicKey{aKey}, []string{""}),
		"error creating delegation")
	require.NoError(t,
		ownerRepo.AddDelegation("targets/a/b", []data.PublicKey{bKey}, []string{""}),
		"error creating delegation")

	require.NoError(t, ownerRepo.Publish())
	// delegation parents all get signed
	ownerRec.requireAsked(t, []string{data.CanonicalTargetsRole.String(), "targets/a"})

	// assert both delegation roles appear to the other repo in a call to GetDelegationRoles
	delgRoleList, err := delgRepo.GetDelegationRoles()
	require.NoError(t, err)
	require.Len(t, delgRoleList, 2)
	// The walk is a pre-order so we can enforce ordering.  Also check that canonical key IDs are reported from this walk
	require.EqualValues(t, delgRoleList[0].Name, "targets/a")
	require.NotContains(t, delgRoleList[0].KeyIDs, ownerRepo.tufRepo.Targets[data.CanonicalTargetsRole].Signed.Delegations.Roles[0].KeyIDs)
	canonicalAKeyID, err := utils.CanonicalKeyID(aKey)
	require.NoError(t, err)
	require.Contains(t, delgRoleList[0].KeyIDs, canonicalAKeyID)
	require.EqualValues(t, delgRoleList[1].Name, "targets/a/b")
	require.NotContains(t, delgRoleList[1].KeyIDs, ownerRepo.tufRepo.Targets["targets/a"].Signed.Delegations.Roles[0].KeyIDs)
	canonicalBKeyID, err := utils.CanonicalKeyID(bKey)
	require.NoError(t, err)
	require.Contains(t, delgRoleList[1].KeyIDs, canonicalBKeyID)
	// assert that the key ID data didn't somehow change between the two repos, since we translated to canonical key IDs
	require.Equal(t, ownerRepo.tufRepo.Targets[data.CanonicalTargetsRole].Signed.Delegations.Roles[0].KeyIDs, delgRepo.tufRepo.Targets[data.CanonicalTargetsRole].Signed.Delegations.Roles[0].KeyIDs)
	require.EqualValues(t, ownerRepo.tufRepo.Targets["targets/a"].Signed.Delegations.Roles[0].KeyIDs, delgRepo.tufRepo.Targets["targets/a"].Signed.Delegations.Roles[0].KeyIDs)

	// delegated repo now publishes to delegated roles, but it will need
	// to download those roles first, since it doesn't know about them
	requirePublishToRolesSucceeds(t, delgRepo, []data.RoleName{data.RoleName("targets/a/b")}, []data.RoleName{data.RoleName("targets/a/b")})
	delgRec.requireAsked(t, []string{"targets/a/b"})
}

// Ensure that two clients can publish delegations with two different keys and
// the changes will not clobber each other.
func TestPublishTargetsDelegationFromTwoRepos(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// this happens to be the client that creates the repo, but can also
	// write a delegation
	repo1, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, true)
	defer os.RemoveAll(repo1.baseDir)

	// this is the second writable repo
	repo2, _ := newRepoToTestRepo(t, repo1, true)
	defer os.RemoveAll(repo2.baseDir)

	// create keys for each repo
	key1, err := repo1.CryptoService.Create("targets/a", repo1.gun, data.ECDSAKey)
	require.NoError(t, err, "error creating delegation key")

	// create a key on the delegated repo
	key2, err := repo2.CryptoService.Create("targets/a", repo2.gun, data.ECDSAKey)
	require.NoError(t, err, "error creating delegation key")

	// delegation includes both keys
	require.NoError(t,
		repo1.AddDelegation("targets/a", []data.PublicKey{key1, key2}, []string{""}),
		"error creating delegation")

	require.NoError(t, repo1.Publish())

	// clear metadata and unencrypted private key cache
	var rec1, rec2 *passRoleRecorder
	repo1, rec1 = newRepoToTestRepo(t, repo1, false)
	repo2, rec2 = newRepoToTestRepo(t, repo2, false)

	// both repos add targets and publish
	addTarget(t, repo1, "first", "../fixtures/root-ca.crt", "targets/a")
	require.NoError(t, repo1.Publish())
	rec1.requireAsked(t, []string{"targets/a"})
	rec1.clear()

	addTarget(t, repo2, "second", "../fixtures/root-ca.crt", "targets/a")
	require.NoError(t, repo2.Publish())
	rec2.requireAsked(t, []string{"targets/a"})
	rec2.clear()

	// first repo can publish again
	addTarget(t, repo1, "third", "../fixtures/root-ca.crt", "targets/a")
	require.NoError(t, repo1.Publish())
	// key has been cached now
	rec1.requireAsked(t, nil)
	rec1.clear()

	// both repos should be able to see all targets
	for _, repo := range []*NotaryRepository{repo1, repo2} {
		targets, err := repo.ListTargets()
		require.NoError(t, err)
		require.Len(t, targets, 3)

		found := make(map[string]bool)
		for _, t := range targets {
			found[t.Name] = true
		}

		for _, targetName := range []string{"first", "second", "third"} {
			_, ok := found[targetName]
			require.True(t, ok)
		}
	}
}

// A client who could publish before can no longer publish once the owner
// removes their delegation key from the delegation role.
func TestPublishRemoveDelegationKeyFromDelegationRole(t *testing.T) {
	gun := "docker.com/notary"
	ts := fullTestServer(t)
	defer ts.Close()

	// this is the original repo - it owns the root/targets keys and creates
	// the delegation to which it doesn't have the key (so server snapshot
	// signing would be required)
	ownerRepo, _ := initializeRepo(t, data.ECDSAKey, gun, ts.URL, true)
	defer os.RemoveAll(ownerRepo.baseDir)

	// this is a user, or otherwise a repo that only has access to the delegation
	// key so it can publish targets to the delegated role
	delgRepo, _ := newRepoToTestRepo(t, ownerRepo, true)
	defer os.RemoveAll(delgRepo.baseDir)

	// create a key on the delegated repo
	aKey, err := delgRepo.CryptoService.Create("targets/a", delgRepo.gun, data.ECDSAKey)
	require.NoError(t, err, "error creating delegation key")

	// owner creates delegation, adds the delegated key to it, and publishes it
	require.NoError(t,
		ownerRepo.AddDelegation("targets/a", []data.PublicKey{aKey}, []string{""}),
		"error creating delegation")
	require.NoError(t, ownerRepo.Publish())

	// delegated repo can now publish to delegated role
	addTarget(t, delgRepo, "v1", "../fixtures/root-ca.crt", "targets/a")
	require.NoError(t, delgRepo.Publish())

	// owner revokes delegation
	// note there is no removekeyfromdelegation yet, so here's a hack to do so
	newKey, err := ownerRepo.CryptoService.Create("targets/a", ownerRepo.gun, data.ECDSAKey)
	require.NoError(t, err)
	tdJSON, err := json.Marshal(&changelist.TUFDelegation{
		NewThreshold: 1,
		AddKeys:      data.KeyList([]data.PublicKey{newKey}),
		RemoveKeys:   []string{aKey.ID()},
	})
	require.NoError(t, err)

	cl, err := changelist.NewFileChangelist(
		filepath.Join(
			filepath.Join(ownerRepo.baseDir, tufDir, filepath.FromSlash(gun)),
			"changelist",
		),
	)
	require.NoError(t, err)
	require.NoError(t, cl.Add(changelist.NewTUFChange(
		changelist.ActionUpdate,
		"targets/a",
		changelist.TypeTargetsDelegation,
		"",
		tdJSON,
	)))
	cl.Close()
	require.NoError(t, ownerRepo.Publish())

	// delegated repo can now no longer publish to delegated role
	addTarget(t, delgRepo, "v2", "../fixtures/root-ca.crt", "targets/a")
	require.Error(t, delgRepo.Publish())
}

// A client who could publish before can no longer publish once the owner
// deletes the delegation
func TestPublishRemoveDelegation(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// this is the original repo - it owns the root/targets keys and creates
	// the delegation to which it doesn't have the key (so server snapshot
	// signing would be required)
	ownerRepo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, true)
	defer os.RemoveAll(ownerRepo.baseDir)

	// this is a user, or otherwise a repo that only has access to the delegation
	// key so it can publish targets to the delegated role
	delgRepo, _ := newRepoToTestRepo(t, ownerRepo, true)
	defer os.RemoveAll(delgRepo.baseDir)

	// create a key on the delegated repo
	aKey, err := delgRepo.CryptoService.Create("targets/a", delgRepo.gun, data.ECDSAKey)
	require.NoError(t, err, "error creating delegation key")

	// owner creates delegation, adds the delegated key to it, and publishes it
	require.NoError(t,
		ownerRepo.AddDelegation("targets/a", []data.PublicKey{aKey}, []string{""}),
		"error creating delegation")
	require.NoError(t, ownerRepo.Publish())

	// delegated repo can now publish to delegated role
	addTarget(t, delgRepo, "v1", "../fixtures/root-ca.crt", "targets/a")
	require.NoError(t, delgRepo.Publish())

	// owner removes delegation
	aKeyCanonicalID, err := utils.CanonicalKeyID(aKey)
	require.NoError(t, err)
	require.NoError(t, ownerRepo.RemoveDelegationKeys("targets/a", []string{aKeyCanonicalID}))
	require.NoError(t, ownerRepo.Publish())

	// delegated repo can now no longer publish to delegated role
	addTarget(t, delgRepo, "v2", "../fixtures/root-ca.crt", "targets/a")
	require.Error(t, delgRepo.Publish())
}

// If the delegation data is corrupt or unreadable, it doesn't matter because
// all the delegation information is just re-downloaded.  When bootstrapping
// the repository from disk, we just don't load the data from disk because
// there should not be anything there yet.
func TestPublishSucceedsDespiteDelegationCorrupt(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	delgKey, err := repo.CryptoService.Create("targets/a", repo.gun, data.ECDSAKey)
	require.NoError(t, err, "error creating delegation key")

	require.NoError(t,
		repo.AddDelegation("targets/a", []data.PublicKey{delgKey}, []string{""}),
		"error creating delegation")

	testPublishBadMetadata(t, "targets/a", repo, false, true)

	// publish again, now that it has already been published, and again there
	// is no error.
	testPublishBadMetadata(t, "targets/a", repo, true, true)
}

// Rotate invalid roles, or attempt to delegate target signing to the server
func TestRotateKeyInvalidRole(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	// create a delegation
	pubKey, err := repo.CryptoService.Create("targets/releases", data.GUN("docker.com/notary"), data.ECDSAKey)
	require.NoError(t, err)
	require.NoError(t, repo.AddDelegation("targets/releases", []data.PublicKey{pubKey}, []string{""}))
	require.NoError(t, repo.Publish())
	require.NoError(t, repo.Update(false))

	// rotating a root key to the server fails
	require.Error(t, repo.RotateKey(data.CanonicalRootRole, true, nil),
		"Rotating a root key with server-managing the key should fail")

	// rotating a targets key to the server fails
	require.Error(t, repo.RotateKey(data.CanonicalTargetsRole, true, nil),
		"Rotating a targets key with server-managing the key should fail")

	// rotating a timestamp key locally fails
	require.Error(t, repo.RotateKey(data.CanonicalTimestampRole, false, nil),
		"Rotating a timestamp key locally should fail")

	// rotating a delegation key fails
	require.Error(t, repo.RotateKey("targets/releases", false, nil),
		"Rotating a delegation key should fail")

	// rotating a delegation key to the server also fails
	require.Error(t, repo.RotateKey("targets/releases", true, nil),
		"Rotating a delegation key should fail")

	// rotating a not a real role key fails
	require.Error(t, repo.RotateKey("nope", false, nil),
		"Rotating a non-real role key should fail")

	// rotating a not a real role key to the server also fails
	require.Error(t, repo.RotateKey("nope", true, nil),
		"Rotating a non-real role key should fail")
}

// If remotely rotating key fails, the failure is propagated
func TestRemoteRotationError(t *testing.T) {
	ts, _, _ := simpleTestServer(t)

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, true)
	defer os.RemoveAll(repo.baseDir)

	ts.Close()

	// server has died, so this should fail
	for _, role := range []data.RoleName{data.CanonicalSnapshotRole, data.CanonicalTimestampRole} {
		err := repo.RotateKey(role, true, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to rotate remote key")
	}
}

// If remotely rotating key fails for any reason, fail the rotation entirely
func TestRemoteRotationEndpointError(t *testing.T) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()
	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, true)
	defer os.RemoveAll(repo.baseDir)

	// simpleTestServer has no rotate key endpoint, so this should fail
	for _, role := range []data.RoleName{data.CanonicalSnapshotRole, data.CanonicalTimestampRole} {
		err := repo.RotateKey(role, true, nil)
		require.Error(t, err)
		require.IsType(t, store.ErrMetaNotFound{}, err)
	}
}

// The rotator is not the owner of the repository, they cannot rotate the remote
// key
func TestRemoteRotationNoRootKey(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, true)
	defer os.RemoveAll(repo.baseDir)
	require.NoError(t, repo.Publish())

	newRepo, _ := newRepoToTestRepo(t, repo, true)
	defer os.RemoveAll(newRepo.baseDir)
	_, err := newRepo.ListTargets()
	require.NoError(t, err)

	err = newRepo.RotateKey(data.CanonicalSnapshotRole, true, nil)
	require.Error(t, err)
	require.IsType(t, signed.ErrInsufficientSignatures{}, err)
}

// The repo hasn't been initialized, so we can't rotate
func TestRemoteRotationNonexistentRepo(t *testing.T) {
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	err := repo.RotateKey(data.CanonicalTimestampRole, true, nil)
	require.Error(t, err)
	require.IsType(t, ErrRepoNotInitialized{}, err)
}

// Rotates the keys.  After the rotation, downloading the latest metadata
// and require that the keys have changed
func requireRotationSuccessful(t *testing.T, repo1 *NotaryRepository, keysToRotate map[data.RoleName]bool) {
	// Create a new repo that is used to download the data after the rotation
	repo2, _ := newRepoToTestRepo(t, repo1, true)
	defer os.RemoveAll(repo2.baseDir)

	repos := []*NotaryRepository{repo1, repo2}

	oldRoles := make(map[string]data.BaseRole)
	for roleName := range keysToRotate {
		baseRole, err := repo1.tufRepo.GetBaseRole(roleName)
		require.NoError(t, err)
		require.Len(t, baseRole.Keys, 1)

		oldRoles[roleName.String()] = baseRole
	}

	// Confirm no changelists get published
	changesPre := getChanges(t, repo1)

	// Do rotation
	for role, serverManaged := range keysToRotate {
		require.NoError(t, repo1.RotateKey(role, serverManaged, nil))
	}

	changesPost := getChanges(t, repo1)
	require.Equal(t, changesPre, changesPost)

	// Download data from remote and check that keys have changed
	for _, repo := range repos {
		err := repo.Update(true)
		require.NoError(t, err)

		for roleName, isRemoteKey := range keysToRotate {
			baseRole, err := repo1.tufRepo.GetBaseRole(roleName)
			require.NoError(t, err)
			require.Len(t, baseRole.Keys, 1)

			// in the new key is not the same as any of the old keys
			for oldKeyID, oldPubKey := range oldRoles[roleName.String()].Keys {
				_, ok := baseRole.Keys[oldKeyID]
				require.False(t, ok)

				// in the old repo, the old keys have been removed not just from
				// the TUF file, but from the cryptoservice (unless it's a root
				// key, in which case it should NOT be removed)
				if repo == repo1 {
					canonicalID, err := utils.CanonicalKeyID(oldPubKey)
					require.NoError(t, err)

					_, _, err = repo.CryptoService.GetPrivateKey(canonicalID)
					switch roleName {
					case data.CanonicalRootRole:
						require.NoError(t, err)
					default:
						require.Error(t, err)
					}
				}
			}

			// On the old repo, the new key is present in the cryptoservice, or
			// not present if remote.
			if repo == repo1 {
				pubKey := baseRole.ListKeys()[0]
				canonicalID, err := utils.CanonicalKeyID(pubKey)
				require.NoError(t, err)

				key, _, err := repo.CryptoService.GetPrivateKey(canonicalID)
				if isRemoteKey {
					require.Error(t, err)
					require.Nil(t, key)
				} else {
					require.NoError(t, err)
					require.NotNil(t, key)
				}
			}
		}
	}
}

// Initialize repo to have the server sign snapshots (remote snapshot key)
// Without downloading a server-signed snapshot file, rotate keys so that
//    snapshots are locally signed (local snapshot key)
// Assert that we can publish.
func TestRotateBeforePublishFromRemoteKeyToLocalKey(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, true)
	defer os.RemoveAll(repo.baseDir)

	// Adding a target will allow us to confirm the repository is still valid
	// after rotating the keys when we publish (and that rotation doesn't publish
	// non-key-rotation changes)
	addTarget(t, repo, "latest", "../fixtures/intermediate-ca.crt")
	requireRotationSuccessful(t, repo, map[data.RoleName]bool{
		data.CanonicalRootRole:     false,
		data.CanonicalTargetsRole:  false,
		data.CanonicalSnapshotRole: false})

	require.NoError(t, repo.Publish())
	_, err := repo.GetTargetByName("latest")
	require.NoError(t, err)
}

// Initialize a repo, locally signed snapshots
// Publish some content (so that the server has a root.json), and download root.json
// Rotate keys
// Download the latest metadata and require that the keys have changed.
func TestRotateKeyAfterPublishNoServerManagementChange(t *testing.T) {
	testRotateKeySuccess(t, false, map[data.RoleName]bool{data.CanonicalRootRole: false})
	testRotateKeySuccess(t, false, map[data.RoleName]bool{data.CanonicalTargetsRole: false})
	testRotateKeySuccess(t, false, map[data.RoleName]bool{data.CanonicalSnapshotRole: false})
	// rotate multiple keys at once before publishing
	testRotateKeySuccess(t, false, map[data.RoleName]bool{
		data.CanonicalRootRole:     false,
		data.CanonicalSnapshotRole: false,
		data.CanonicalTargetsRole:  false})
}

// Tests rotating keys when there's a change from locally managed keys to
// remotely managed keys and vice versa
// Before rotating, publish some content (so that the server has a root.json),
// and download root.json
func TestRotateKeyAfterPublishServerManagementChange(t *testing.T) {
	// delegate snapshot key management to the server
	testRotateKeySuccess(t, false, map[data.RoleName]bool{
		data.CanonicalSnapshotRole: true,
		data.CanonicalTargetsRole:  false,
		data.CanonicalRootRole:     false,
	})
	// check that the snapshot remote rotation creates new keys
	testRotateKeySuccess(t, true, map[data.RoleName]bool{
		data.CanonicalSnapshotRole: true,
	})
	// check that the timestamp remote rotation creates new keys
	testRotateKeySuccess(t, false, map[data.RoleName]bool{
		data.CanonicalTimestampRole: true,
	})
	// reclaim snapshot key management from the server
	testRotateKeySuccess(t, true, map[data.RoleName]bool{
		data.CanonicalSnapshotRole: false,
		data.CanonicalTargetsRole:  false,
		data.CanonicalRootRole:     false,
	})
}

func testRotateKeySuccess(t *testing.T, serverManagesSnapshotInit bool,
	keysToRotate map[data.RoleName]bool) {

	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL,
		serverManagesSnapshotInit)
	defer os.RemoveAll(repo.baseDir)

	// Adding a target will allow us to confirm the repository is still valid after
	// rotating the keys.
	addTarget(t, repo, "latest", "../fixtures/intermediate-ca.crt")

	requireRotationSuccessful(t, repo, keysToRotate)

	// Publish
	require.NoError(t, repo.Publish())
	// Get root.json and capture targets + snapshot key IDs
	_, err := repo.GetTargetByName("latest")
	require.NoError(t, err)
}

func logRepoTrustRoot(t *testing.T, prefix string, repo *NotaryRepository) {
	logrus.Debugf("==== %s", prefix)
	root := repo.tufRepo.Root
	logrus.Debugf("Root signatures:")
	for _, s := range root.Signatures {
		logrus.Debugf("\t%s", s.KeyID)
	}
	logrus.Debugf("Valid root keys:")
	for _, k := range root.Signed.Roles[data.CanonicalRootRole].KeyIDs {
		logrus.Debugf("\t%s", k)
	}
}

// ID of the (only) certificate trusted by the root role metadata
func rootRoleCertID(t *testing.T, repo *NotaryRepository) string {
	rootKeys := repo.tufRepo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs
	require.Len(t, rootKeys, 1)
	return rootKeys[0]
}

func TestRotateRootKey(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// Set up author's view of the repo and publish first version.
	authorRepo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(authorRepo.baseDir)
	err := authorRepo.Publish()
	require.NoError(t, err)
	oldRootCertID := rootRoleCertID(t, authorRepo)
	oldRootRole, err := authorRepo.tufRepo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	oldCanonicalKeyID, err := utils.CanonicalKeyID(oldRootRole.Keys[oldRootCertID])
	require.NoError(t, err)

	// Initialize an user, using the original root cert and key.
	userRepo, _ := newRepoToTestRepo(t, authorRepo, true)
	defer os.RemoveAll(userRepo.baseDir)
	err = userRepo.Update(false)
	require.NoError(t, err)

	// Rotate root certificate and key.
	logRepoTrustRoot(t, "original", authorRepo)
	err = authorRepo.RotateKey(data.CanonicalRootRole, false, nil)
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-rotate", authorRepo)

	require.NoError(t, authorRepo.Update(false))
	newRootRole, err := authorRepo.tufRepo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	require.False(t, newRootRole.Equals(oldRootRole))
	// not only is the root cert different, but the private key is too
	newRootCertID := rootRoleCertID(t, authorRepo)
	require.NotEqual(t, oldRootCertID, newRootCertID)
	newCanonicalKeyID, err := utils.CanonicalKeyID(newRootRole.Keys[newRootCertID])
	require.NoError(t, err)
	require.NotEqual(t, oldCanonicalKeyID, newCanonicalKeyID)

	// Set up a target to verify the repo is actually usable.
	_, err = userRepo.GetTargetByName("current")
	require.Error(t, err)
	addTarget(t, authorRepo, "current", "../fixtures/intermediate-ca.crt")

	// Publish the target, which does an update and pulls down the latest metadata, and
	// should update the trusted root
	logRepoTrustRoot(t, "pre-publish", authorRepo)
	err = authorRepo.Publish()
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-publish", authorRepo)

	// Verify the user can use the rotated repo, and see the added target.
	_, err = userRepo.GetTargetByName("current")
	require.NoError(t, err)
	logRepoTrustRoot(t, "client", userRepo)

	// Verify that clients initialized post-rotation can use the repo, and use
	// the new certificate immediately.
	freshUserRepo, _ := newRepoToTestRepo(t, authorRepo, true)
	defer os.RemoveAll(freshUserRepo.baseDir)
	_, err = freshUserRepo.GetTargetByName("current")
	require.NoError(t, err)
	require.Equal(t, newRootCertID, rootRoleCertID(t, freshUserRepo))
	logRepoTrustRoot(t, "fresh client", freshUserRepo)

	// Verify that the user initialized with the original certificate eventually
	// rotates to the new certificate.
	err = userRepo.Update(false)
	require.NoError(t, err)
	logRepoTrustRoot(t, "user refresh 1", userRepo)
	require.Equal(t, newRootCertID, rootRoleCertID(t, userRepo))
}

func TestRotateRootMultiple(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// Set up author's view of the repo and publish first version.
	authorRepo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(authorRepo.baseDir)
	err := authorRepo.Publish()
	require.NoError(t, err)
	oldRootCertID := rootRoleCertID(t, authorRepo)
	oldRootRole, err := authorRepo.tufRepo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	oldCanonicalKeyID, err := utils.CanonicalKeyID(oldRootRole.Keys[oldRootCertID])
	require.NoError(t, err)

	// Initialize a user, using the original root cert and key.
	userRepo, _ := newRepoToTestRepo(t, authorRepo, true)
	defer os.RemoveAll(userRepo.baseDir)
	err = userRepo.Update(false)
	require.NoError(t, err)

	// Rotate root certificate and key.
	logRepoTrustRoot(t, "original", authorRepo)
	err = authorRepo.RotateKey(data.CanonicalRootRole, false, nil)
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-rotate", authorRepo)

	// Rotate root certificate and key again.
	err = authorRepo.RotateKey(data.CanonicalRootRole, false, nil)
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-rotate-again", authorRepo)

	require.NoError(t, authorRepo.Update(false))
	newRootRole, err := authorRepo.tufRepo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	require.False(t, newRootRole.Equals(oldRootRole))
	// not only is the root cert different, but the private key is too
	newRootCertID := rootRoleCertID(t, authorRepo)
	require.NotEqual(t, oldRootCertID, newRootCertID)
	newCanonicalKeyID, err := utils.CanonicalKeyID(newRootRole.Keys[newRootCertID])
	require.NoError(t, err)
	require.NotEqual(t, oldCanonicalKeyID, newCanonicalKeyID)

	// Set up a target to verify the repo is actually usable.
	_, err = userRepo.GetTargetByName("current")
	require.Error(t, err)
	addTarget(t, authorRepo, "current", "../fixtures/intermediate-ca.crt")

	// Publish the target, which does an update and pulls down the latest metadata, and
	// should update the trusted root
	logRepoTrustRoot(t, "pre-publish", authorRepo)
	err = authorRepo.Publish()
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-publish", authorRepo)

	// Verify the user can use the rotated repo, and see the added target.
	err = userRepo.Update(false)
	require.NoError(t, err)
	_, err = userRepo.GetTargetByName("current")
	require.NoError(t, err)
	logRepoTrustRoot(t, "client", userRepo)

	// Verify that clients initialized post-rotation can use the repo, and use
	// the new certificate immediately.
	freshUserRepo, _ := newRepoToTestRepo(t, authorRepo, true)
	defer os.RemoveAll(freshUserRepo.baseDir)
	_, err = freshUserRepo.GetTargetByName("current")
	require.NoError(t, err)
	require.Equal(t, newRootCertID, rootRoleCertID(t, freshUserRepo))
	logRepoTrustRoot(t, "fresh client", freshUserRepo)

	// Verify that the user initialized with the original certificate eventually
	// rotates to the new certificate.
	err = userRepo.Update(false)
	require.NoError(t, err)
	logRepoTrustRoot(t, "user refresh 1", userRepo)
	require.Equal(t, newRootCertID, rootRoleCertID(t, userRepo))
}

func TestRotateRootKeyProvided(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// Set up author's view of the repo and publish first version.
	authorRepo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(authorRepo.baseDir)
	err := authorRepo.Publish()
	require.NoError(t, err)
	oldRootCertID := rootRoleCertID(t, authorRepo)
	oldRootRole, err := authorRepo.tufRepo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	oldCanonicalKeyID, err := utils.CanonicalKeyID(oldRootRole.Keys[oldRootCertID])
	require.NoError(t, err)

	// Initialize an user, using the original root cert and key.
	userRepo, _ := newRepoToTestRepo(t, authorRepo, true)
	defer os.RemoveAll(userRepo.baseDir)
	err = userRepo.Update(false)
	require.NoError(t, err)

	// Key loaded from file (just generating it here)
	rootPublicKey, err := authorRepo.CryptoService.Create(data.CanonicalRootRole, "", data.ECDSAKey)
	require.NoError(t, err)
	rootPrivateKey, _, err := authorRepo.CryptoService.GetPrivateKey(rootPublicKey.ID())
	require.NoError(t, err)

	// Fail to rotate to bad key
	err = authorRepo.RotateKey(data.CanonicalRootRole, false, []string{"notakey"})
	require.Error(t, err)

	// Rotate root certificate and key.
	logRepoTrustRoot(t, "original", authorRepo)
	err = authorRepo.RotateKey(data.CanonicalRootRole, false, []string{rootPrivateKey.ID()})
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-rotate", authorRepo)

	require.NoError(t, authorRepo.Update(false))
	newRootRole, err := authorRepo.tufRepo.GetBaseRole(data.CanonicalRootRole)
	require.False(t, newRootRole.Equals(oldRootRole))
	require.NoError(t, err)
	// not only is the root cert different, but the private key is too
	newRootCertID := rootRoleCertID(t, authorRepo)
	require.NotEqual(t, oldRootCertID, newRootCertID)
	newCanonicalKeyID, err := utils.CanonicalKeyID(newRootRole.Keys[newRootCertID])
	require.NoError(t, err)
	require.NotEqual(t, oldCanonicalKeyID, newCanonicalKeyID)
	require.Equal(t, rootPrivateKey.ID(), newCanonicalKeyID)

	// Set up a target to verify the repo is actually usable.
	_, err = userRepo.GetTargetByName("current")
	require.Error(t, err)
	addTarget(t, authorRepo, "current", "../fixtures/intermediate-ca.crt")

	// Publish the target, which does an update and pulls down the latest metadata, and
	// should update the trusted root
	logRepoTrustRoot(t, "pre-publish", authorRepo)
	err = authorRepo.Publish()
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-publish", authorRepo)

	// Verify the user can use the rotated repo, and see the added target.
	_, err = userRepo.GetTargetByName("current")
	require.NoError(t, err)
	logRepoTrustRoot(t, "client", userRepo)

	// Verify that clients initialized post-rotation can use the repo, and use
	// the new certificate immediately.
	freshUserRepo, _ := newRepoToTestRepo(t, authorRepo, true)
	defer os.RemoveAll(freshUserRepo.baseDir)
	_, err = freshUserRepo.GetTargetByName("current")
	require.NoError(t, err)
	require.Equal(t, newRootCertID, rootRoleCertID(t, freshUserRepo))
	logRepoTrustRoot(t, "fresh client", freshUserRepo)

	// Verify that the user initialized with the original certificate eventually
	// rotates to the new certificate.
	err = userRepo.Update(false)
	require.NoError(t, err)
	logRepoTrustRoot(t, "user refresh 1", userRepo)
	require.Equal(t, newRootCertID, rootRoleCertID(t, userRepo))
}

func TestRotateRootKeyLegacySupport(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	// Set up author's view of the repo and publish first version.
	authorRepo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(authorRepo.baseDir)
	err := authorRepo.Publish()
	require.NoError(t, err)
	oldRootCertID := rootRoleCertID(t, authorRepo)
	oldRootRole, err := authorRepo.tufRepo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	oldCanonicalKeyID, err := utils.CanonicalKeyID(oldRootRole.Keys[oldRootCertID])
	require.NoError(t, err)

	// Initialize a user, using the original root cert and key.
	userRepo, _ := newRepoToTestRepo(t, authorRepo, true)
	defer os.RemoveAll(userRepo.baseDir)
	err = userRepo.Update(false)
	require.NoError(t, err)

	// Rotate root certificate and key.
	logRepoTrustRoot(t, "original", authorRepo)
	err = authorRepo.RotateKey(data.CanonicalRootRole, false, nil)
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-rotate", authorRepo)

	// Rotate root certificate and key again, this time with legacy support
	authorRepo.LegacyVersions = SignWithAllOldVersions
	err = authorRepo.RotateKey(data.CanonicalRootRole, false, nil)
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-rotate-again", authorRepo)

	require.NoError(t, authorRepo.Update(false))
	newRootRole, err := authorRepo.tufRepo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	require.False(t, newRootRole.Equals(oldRootRole))
	// not only is the root cert different, but the private key is too
	newRootCertID := rootRoleCertID(t, authorRepo)
	require.NotEqual(t, oldRootCertID, newRootCertID)
	newCanonicalKeyID, err := utils.CanonicalKeyID(newRootRole.Keys[newRootCertID])
	require.NoError(t, err)
	require.NotEqual(t, oldCanonicalKeyID, newCanonicalKeyID)

	// Set up a target to verify the repo is actually usable.
	_, err = userRepo.GetTargetByName("current")
	require.Error(t, err)
	addTarget(t, authorRepo, "current", "../fixtures/intermediate-ca.crt")

	// Publish the target, which does an update and pulls down the latest metadata, and
	// should update the trusted root
	logRepoTrustRoot(t, "pre-publish", authorRepo)
	err = authorRepo.Publish()
	require.NoError(t, err)
	logRepoTrustRoot(t, "post-publish", authorRepo)

	// Verify the user can use the rotated repo, and see the added target.
	err = userRepo.Update(false)
	require.NoError(t, err)
	_, err = userRepo.GetTargetByName("current")
	require.NoError(t, err)
	logRepoTrustRoot(t, "client", userRepo)

	// Verify that the user's rotated root is signed with all available old keys
	require.NoError(t, err)
	require.Equal(t, 3, len(userRepo.tufRepo.Root.Signatures))

	// Verify that clients initialized post-rotation can use the repo, and use
	// the new certificate immediately.
	freshUserRepo, _ := newRepoToTestRepo(t, authorRepo, true)
	defer os.RemoveAll(freshUserRepo.baseDir)
	_, err = freshUserRepo.GetTargetByName("current")
	require.NoError(t, err)
	require.Equal(t, newRootCertID, rootRoleCertID(t, freshUserRepo))
	logRepoTrustRoot(t, "fresh client", freshUserRepo)

	// Verify that the user initialized with the original certificate eventually
	// rotates to the new certificate.
	err = userRepo.Update(false)
	require.NoError(t, err)
	logRepoTrustRoot(t, "user refresh 1", userRepo)
	require.Equal(t, newRootCertID, rootRoleCertID(t, userRepo))
}

// If there is no local cache, notary operations return the remote error code
func TestRemoteServerUnavailableNoLocalCache(t *testing.T) {
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	defer os.RemoveAll(tempBaseDir)

	ts := errorTestServer(t, 500)
	defer ts.Close()

	repo, err := NewFileCachedNotaryRepository(tempBaseDir, "docker.com/notary",
		ts.URL, http.DefaultTransport, passphraseRetriever, trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	_, err = repo.ListTargets(data.CanonicalTargetsRole)
	require.Error(t, err)
	require.IsType(t, store.ErrServerUnavailable{}, err)

	_, err = repo.GetTargetByName("targetName")
	require.Error(t, err)
	require.IsType(t, store.ErrServerUnavailable{}, err)

	err = repo.Publish()
	require.Error(t, err)
	require.IsType(t, store.ErrServerUnavailable{}, err)
}

// AddDelegation creates a valid changefile (rejects invalid delegation names,
// but does not check the delegation hierarchy).  When applied, the change adds
// a new delegation role with the correct keys.
func TestAddDelegationChangefileValid(t *testing.T) {
	gun := "docker.com/notary"
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, gun, ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	targetKeyIds := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)
	require.NotEmpty(t, targetKeyIds)
	targetPubKey := repo.CryptoService.GetKey(targetKeyIds[0])
	require.NotNil(t, targetPubKey)

	err := repo.AddDelegation(data.CanonicalRootRole, []data.PublicKey{targetPubKey}, []string{""})
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
	require.Empty(t, getChanges(t, repo))

	// to show that adding does not care about the hierarchy
	err = repo.AddDelegation("targets/a/b/c", []data.PublicKey{targetPubKey}, []string{""})
	require.NoError(t, err)

	// ensure that the changefiles is correct
	changes := getChanges(t, repo)
	require.Len(t, changes, 2)
	require.Equal(t, changelist.ActionCreate, changes[0].Action())
	require.EqualValues(t, "targets/a/b/c", changes[0].Scope())
	require.Equal(t, changelist.TypeTargetsDelegation, changes[0].Type())
	require.Equal(t, changelist.ActionCreate, changes[1].Action())
	require.EqualValues(t, "targets/a/b/c", changes[1].Scope())
	require.Equal(t, changelist.TypeTargetsDelegation, changes[1].Type())
	require.EqualValues(t, "", changes[1].Path())
	require.NotEmpty(t, changes[0].Content())
}

// The changefile produced by AddDelegation, when applied, actually adds
// the delegation to the repo (assuming the delegation hierarchy is correct -
// tests for change application validation are in helpers_test.go)
func TestAddDelegationChangefileApplicable(t *testing.T) {
	gun := "docker.com/notary"
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, gun, ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	targetKeyIds := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)
	require.NotEmpty(t, targetKeyIds)
	targetPubKey := repo.CryptoService.GetKey(targetKeyIds[0])
	require.NotNil(t, targetPubKey)

	// this hierarchy has to be right to be applied
	err := repo.AddDelegation("targets/a", []data.PublicKey{targetPubKey}, []string{""})
	require.NoError(t, err)
	changes := getChanges(t, repo)
	require.Len(t, changes, 2)

	// ensure that it can be applied correctly
	err = applyTargetsChange(repo.tufRepo, nil, changes[0])
	require.NoError(t, err)

	targetRole := repo.tufRepo.Targets[data.CanonicalTargetsRole]
	require.Len(t, targetRole.Signed.Delegations.Roles, 1)
	require.Len(t, targetRole.Signed.Delegations.Keys, 1)

	_, ok := targetRole.Signed.Delegations.Keys[targetPubKey.ID()]
	require.True(t, ok)

	newDelegationRole := targetRole.Signed.Delegations.Roles[0]
	require.Len(t, newDelegationRole.KeyIDs, 1)
	require.Equal(t, targetPubKey.ID(), newDelegationRole.KeyIDs[0])
	require.EqualValues(t, "targets/a", newDelegationRole.Name)
}

// TestAddDelegationErrorWritingChanges expects errors writing a change to file
// to be propagated.
func TestAddDelegationErrorWritingChanges(t *testing.T) {
	testErrorWritingChangefiles(t, func(repo *NotaryRepository) error {
		targetKeyIds := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)
		require.NotEmpty(t, targetKeyIds)
		targetPubKey := repo.CryptoService.GetKey(targetKeyIds[0])
		require.NotNil(t, targetPubKey)

		return repo.AddDelegation("targets/a", []data.PublicKey{targetPubKey}, []string{""})
	})
}

// RemoveDelegation rejects attempts to remove invalidly-named delegations,
// but otherwise does not validate the name of the delegation to remove.  This
// test ensures that the changefile generated by RemoveDelegation is correct.
func TestRemoveDelegationChangefileValid(t *testing.T) {
	gun := "docker.com/notary"
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rootKeyID := initializeRepo(t, data.ECDSAKey, gun, ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	rootPubKey := repo.CryptoService.GetKey(rootKeyID)
	require.NotNil(t, rootPubKey)

	err := repo.RemoveDelegationKeys(data.CanonicalRootRole, []string{rootKeyID})
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
	require.Empty(t, getChanges(t, repo))

	// to demonstrate that so long as the delegation name is valid, the
	// existence of the delegation doesn't matter
	require.NoError(t, repo.RemoveDelegationKeys("targets/a/b/c", []string{rootKeyID}))

	// ensure that the changefile is correct
	changes := getChanges(t, repo)
	require.Len(t, changes, 1)
	require.Equal(t, changelist.ActionUpdate, changes[0].Action())
	require.EqualValues(t, "targets/a/b/c", changes[0].Scope())
	require.Equal(t, changelist.TypeTargetsDelegation, changes[0].Type())
	require.Equal(t, "", changes[0].Path())
}

// The changefile produced by RemoveDelegationKeys, when applied, actually removes
// the delegation from the repo (assuming the repo exists - tests for
// change application validation are in helpers_test.go)
func TestRemoveDelegationChangefileApplicable(t *testing.T) {
	gun := "docker.com/notary"
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rootKeyID := initializeRepo(t, data.ECDSAKey, gun, ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	rootPubKey := repo.CryptoService.GetKey(rootKeyID)
	require.NotNil(t, rootPubKey)

	// add a delegation first so it can be removed
	require.NoError(t, repo.AddDelegation("targets/a", []data.PublicKey{rootPubKey}, []string{""}))
	changes := getChanges(t, repo)
	require.Len(t, changes, 2)
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[0]))
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[1]))

	targetRole := repo.tufRepo.Targets[data.CanonicalTargetsRole]
	require.Len(t, targetRole.Signed.Delegations.Roles, 1)
	require.Len(t, targetRole.Signed.Delegations.Keys, 1)

	// now remove it
	rootKeyCanonicalID, err := utils.CanonicalKeyID(rootPubKey)
	require.NoError(t, err)
	require.NoError(t, repo.RemoveDelegationKeys("targets/a", []string{rootKeyCanonicalID}))
	changes = getChanges(t, repo)
	require.Len(t, changes, 3)
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[2]))

	targetRole = repo.tufRepo.Targets[data.CanonicalTargetsRole]
	require.Len(t, targetRole.Signed.Delegations.Roles, 1)
	require.Empty(t, targetRole.Signed.Delegations.Keys)
}

// The changefile with the ClearAllPaths key set, when applied, actually removes
// all paths from the specified delegation in the repo (assuming the repo and delegation exist)
func TestClearAllPathsDelegationChangefileApplicable(t *testing.T) {
	gun := "docker.com/notary"
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rootKeyID := initializeRepo(t, data.ECDSAKey, gun, ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	rootPubKey := repo.CryptoService.GetKey(rootKeyID)
	require.NotNil(t, rootPubKey)

	// add a delegation first so it can be removed
	require.NoError(t, repo.AddDelegation("targets/a", []data.PublicKey{rootPubKey}, []string{"abc,123,xyz,path"}))
	changes := getChanges(t, repo)
	require.Len(t, changes, 2)
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[0]))
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[1]))

	// now clear paths it
	require.NoError(t, repo.ClearDelegationPaths("targets/a"))
	changes = getChanges(t, repo)
	require.Len(t, changes, 3)
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[2]))

	delgRoles := repo.tufRepo.Targets[data.CanonicalTargetsRole].Signed.Delegations.Roles
	require.Len(t, delgRoles, 1)
	require.Len(t, delgRoles[0].Paths, 0)
}

// TestFullAddDelegationChangefileApplicable generates a single changelist with AddKeys and AddPaths set,
// (in the old style of AddDelegation) and tests that all of its changes are reflected on publish
func TestFullAddDelegationChangefileApplicable(t *testing.T) {
	gun := "docker.com/notary"
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rootKeyID := initializeRepo(t, data.ECDSAKey, gun, ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	rootPubKey := repo.CryptoService.GetKey(rootKeyID)
	require.NotNil(t, rootPubKey)

	key2, err := repo.CryptoService.Create("user", repo.gun, data.ECDSAKey)
	require.NoError(t, err)

	var delegationName data.RoleName = "targets/a"

	// manually create the changelist object to load multiple keys
	tdJSON, err := json.Marshal(&changelist.TUFDelegation{
		NewThreshold: notary.MinThreshold,
		AddKeys:      data.KeyList([]data.PublicKey{rootPubKey, key2}),
		AddPaths:     []string{"abc", "123", "xyz"},
	})
	require.NoError(t, err)
	change := newCreateDelegationChange(delegationName, tdJSON)
	cl, err := changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, tufDir, filepath.FromSlash(gun), "changelist"),
	)
	require.NoError(t, err)
	addChange(cl, change, delegationName)

	changes := getChanges(t, repo)
	require.Len(t, changes, 1)
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[0]))

	delgRoles := repo.tufRepo.Targets[data.CanonicalTargetsRole].Signed.Delegations.Roles
	require.Len(t, delgRoles, 1)
	require.Len(t, delgRoles[0].Paths, 3)
	require.Len(t, delgRoles[0].KeyIDs, 2)
	require.Equal(t, delgRoles[0].Name, delegationName)
}

// TestFullRemoveDelegationChangefileApplicable generates a single changelist with RemoveKeys and RemovePaths set,
// (in the old style of RemoveDelegation) and tests that all of its changes are reflected on publish
func TestFullRemoveDelegationChangefileApplicable(t *testing.T) {
	gun := "docker.com/notary"
	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rootKeyID := initializeRepo(t, data.ECDSAKey, gun, ts.URL, false)
	defer os.RemoveAll(repo.baseDir)
	rootPubKey := repo.CryptoService.GetKey(rootKeyID)
	require.NotNil(t, rootPubKey)

	key2, err := repo.CryptoService.Create("user", repo.gun, data.ECDSAKey)
	require.NoError(t, err)
	key2CanonicalID, err := utils.CanonicalKeyID(key2)
	require.NoError(t, err)

	var delegationName data.RoleName = "targets/a"

	require.NoError(t, repo.AddDelegation(delegationName, []data.PublicKey{rootPubKey, key2}, []string{"abc", "123"}))
	changes := getChanges(t, repo)
	require.Len(t, changes, 2)
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[0]))
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[1]))

	targetRole := repo.tufRepo.Targets[data.CanonicalTargetsRole]
	require.Len(t, targetRole.Signed.Delegations.Roles, 1)
	require.Len(t, targetRole.Signed.Delegations.Keys, 2)

	// manually create the changelist object to load multiple keys
	tdJSON, err := json.Marshal(&changelist.TUFDelegation{
		RemoveKeys:  []string{key2CanonicalID},
		RemovePaths: []string{"abc", "123"},
	})
	require.NoError(t, err)
	change := newUpdateDelegationChange(delegationName, tdJSON)
	cl, err := changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, tufDir, filepath.FromSlash(gun), "changelist"),
	)
	require.NoError(t, err)
	addChange(cl, change, delegationName)

	changes = getChanges(t, repo)
	require.Len(t, changes, 3)
	require.NoError(t, applyTargetsChange(repo.tufRepo, nil, changes[2]))

	delgRoles := repo.tufRepo.Targets[data.CanonicalTargetsRole].Signed.Delegations.Roles
	require.Len(t, delgRoles, 1)
	require.Len(t, delgRoles[0].Paths, 0)
	require.Len(t, delgRoles[0].KeyIDs, 1)
}

// TestRemoveDelegationErrorWritingChanges expects errors writing a change to
// file to be propagated.
func TestRemoveDelegationErrorWritingChanges(t *testing.T) {
	testErrorWritingChangefiles(t, func(repo *NotaryRepository) error {
		return repo.RemoveDelegationKeysAndPaths("targets/a", []string{""}, []string{})
	})
}

// TestBootstrapClientBadURL checks that bootstrapClient correctly
// returns an error when the URL is valid but does not point to
// a TUF server
func TestBootstrapClientBadURL(t *testing.T) {
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	repo, err := NewFileCachedNotaryRepository(
		tempBaseDir,
		"testGun",
		"http://localhost:9998",
		http.DefaultTransport,
		passphraseRetriever,
		trustpinning.TrustPinConfig{},
	)
	require.NoError(t, err, "error creating repo: %s", err)

	c, err := repo.bootstrapClient(false)
	require.Nil(t, c)
	require.Error(t, err)

	c, err2 := repo.bootstrapClient(true)
	require.Nil(t, c)
	require.Error(t, err2)

	// same error should be returned because we don't have local data
	// and are requesting remote root regardless of checkInitialized
	// value
	require.EqualError(t, err, err2.Error())
}

// TestClientInvalidURL checks that instantiating a new NotaryRepository
// correctly returns an error when the URL is valid but does not point to
// a TUF server
func TestClientInvalidURL(t *testing.T) {
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	repo, err := NewFileCachedNotaryRepository(
		tempBaseDir,
		"testGun",
		"#!*)&!)#*^%!#)%^!#",
		http.DefaultTransport,
		passphraseRetriever,
		trustpinning.TrustPinConfig{},
	)
	// NewFileCachedNotaryRepository should fail and return an error
	// since it initializes the cache but also the remote repository
	// from the baseURL and the GUN
	require.Nil(t, repo)
	require.Error(t, err)
}

func TestPublishTargetsDelegationCanUseUserKeyWithArbitraryRole(t *testing.T) {
	testPublishTargetsDelegationCanUseUserKeyWithArbitraryRole(t, false)
	testPublishTargetsDelegationCanUseUserKeyWithArbitraryRole(t, true)
}

func testPublishTargetsDelegationCanUseUserKeyWithArbitraryRole(t *testing.T, x509 bool) {
	gun := "docker.com/notary"
	ts := fullTestServer(t)
	defer ts.Close()

	// this is the original repo - it owns the root/targets keys and creates
	// the delegation to which it doesn't have the key (so server snapshot
	// signing would be required)
	ownerRepo, _ := initializeRepo(t, data.ECDSAKey, gun, ts.URL, true)
	defer os.RemoveAll(ownerRepo.baseDir)

	// this is a user, or otherwise a repo that only has access to the delegation
	// key so it can publish targets to the delegated role
	delgRepo, _ := newRepoToTestRepo(t, ownerRepo, true)
	defer os.RemoveAll(delgRepo.baseDir)

	// create a key on the owner repo
	aKey := createKey(t, ownerRepo, "user", x509)

	_, err := utils.CanonicalKeyID(aKey)
	require.NoError(t, err)

	// create a key on the delegated repo
	bKey := createKey(t, delgRepo, "notARealRoleName", x509)
	_, err = utils.CanonicalKeyID(bKey)
	require.NoError(t, err)

	// clear metadata and unencrypted private key cache
	var ownerRec, delgRec *passRoleRecorder
	ownerRepo, ownerRec = newRepoToTestRepo(t, ownerRepo, false)
	delgRepo, delgRec = newRepoToTestRepo(t, delgRepo, false)

	// owner creates delegations, adds the delegated key to them, and publishes them
	require.NoError(t,
		ownerRepo.AddDelegation("targets/a", []data.PublicKey{aKey}, []string{""}),
		"error creating delegation")
	require.NoError(t,
		ownerRepo.AddDelegation("targets/a/b", []data.PublicKey{bKey}, []string{""}),
		"error creating delegation")

	require.NoError(t, ownerRepo.Publish())
	// delegation parents all get signed
	ownerRec.requireAsked(t, []string{data.CanonicalTargetsRole.String(), "targets/a"})

	// delegated repo now publishes to delegated roles, but it will need
	// to download those roles first, since it doesn't know about them
	requirePublishToRolesSucceeds(t, delgRepo, []data.RoleName{"targets/a/b"}, []data.RoleName{"targets/a/b"})

	delgRec.requireAsked(t, []string{"targets/a/b"})
}

// TestDeleteRepo tests that local repo data is deleted from the client library call
func TestDeleteRepo(t *testing.T) {
	var gun data.GUN = "docker.com/notary"

	ts, _, _ := simpleTestServer(t)
	defer ts.Close()

	repo, rootKeyID := initializeRepo(t, data.ECDSAKey, gun.String(), ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	// Assert initialization was successful before we delete
	requireRepoHasExpectedKeys(t, repo, rootKeyID, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalRootRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTargetsRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, true)

	// Stage a change on the changelist
	addTarget(t, repo, "someTarget", "../fixtures/intermediate-ca.crt", data.CanonicalTargetsRole)
	// load the changelist for this repo and check that we have one staged change
	cl, err := changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")
	require.Len(t, cl.List(), 1)

	// Delete all local trust data for repo
	err = DeleteTrustData(repo.baseDir, gun, "", nil, false)
	require.NoError(t, err)

	// Assert no metadata for this repo exists locally
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalRootRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTargetsRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTimestampRole, false)

	// Assert the changelist is cleared of staged changes
	require.Len(t, cl.List(), 0)

	// Check that the tuf/<GUN> directory itself is gone
	_, err = os.Stat(filepath.Join(repo.baseDir, tufDir, filepath.FromSlash(gun.String())))
	require.Error(t, err)

	// Assert keys for this repo exist locally
	requireRepoHasExpectedKeys(t, repo, rootKeyID, true)
}

// TestDeleteRemoteRepo tests that local and remote repo data is deleted from the client library call
func TestDeleteRemoteRepo(t *testing.T) {
	var gun data.GUN = "docker.com/notary"

	ts := fullTestServer(t)
	defer ts.Close()

	// Create and publish a repo to delete
	repo, rootKeyID := initializeRepo(t, data.ECDSAKey, gun.String(), ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	require.NoError(t, repo.Publish())

	// Stage a change on this repo's changelist
	addTarget(t, repo, "someTarget", "../fixtures/intermediate-ca.crt", data.CanonicalTargetsRole)
	// load the changelist for this repo and check that we have one staged change
	repoCl, err := changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")
	require.Len(t, repoCl.List(), 1)

	// Create another repo to ensure it stays intact
	livingGun := "stayingAlive"
	longLivingRepo, _ := initializeRepo(t, data.ECDSAKey, livingGun, ts.URL, false)
	defer os.RemoveAll(longLivingRepo.baseDir)

	require.NoError(t, longLivingRepo.Publish())

	// Stage a change on the long living repo
	addTarget(t, longLivingRepo, "someLivingTarget", "../fixtures/intermediate-ca.crt", data.CanonicalTargetsRole)
	// load the changelist for this repo and check that we have one staged change
	longLivingCl, err := changelist.NewFileChangelist(
		filepath.Join(longLivingRepo.baseDir, "tuf", filepath.FromSlash(longLivingRepo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")
	require.Len(t, longLivingCl.List(), 1)

	// Assert initialization was successful before we delete
	requireRepoHasExpectedKeys(t, repo, rootKeyID, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalRootRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTargetsRole, true)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, true)
	require.Len(t, repoCl.List(), 1)

	// Delete all local and remote trust data for one repo
	err = DeleteTrustData(repo.baseDir, gun, ts.URL, http.DefaultTransport, true)
	require.NoError(t, err)

	// Assert no metadata for that repo exists locally
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalRootRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTargetsRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalSnapshotRole, false)
	requireRepoHasExpectedMetadata(t, repo, data.CanonicalTimestampRole, false)

	// Assert the changelist is cleared of staged changes
	require.Len(t, repoCl.List(), 0)

	// Check that the tuf/<GUN> directory itself is gone
	_, err = os.Stat(filepath.Join(repo.baseDir, tufDir, filepath.FromSlash(gun.String())))
	require.Error(t, err)

	// Assert keys for this repo still exist locally
	requireRepoHasExpectedKeys(t, repo, rootKeyID, true)

	// Try connecting to the remote store directly and make sure that no metadata exists for this gun
	remoteStore := repo.getRemoteStore()
	require.NotNil(t, remoteStore)
	meta, err := remoteStore.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
	require.Error(t, err)
	require.IsType(t, store.ErrMetaNotFound{}, err)
	require.Nil(t, meta)
	meta, err = remoteStore.GetSized(data.CanonicalTargetsRole.String(), store.NoSizeLimit)
	require.Error(t, err)
	require.IsType(t, store.ErrMetaNotFound{}, err)
	require.Nil(t, meta)
	meta, err = remoteStore.GetSized(data.CanonicalSnapshotRole.String(), store.NoSizeLimit)
	require.Error(t, err)
	require.IsType(t, store.ErrMetaNotFound{}, err)
	require.Nil(t, meta)
	meta, err = remoteStore.GetSized(data.CanonicalTimestampRole.String(), store.NoSizeLimit)
	require.Error(t, err)
	require.IsType(t, store.ErrMetaNotFound{}, err)
	require.Nil(t, meta)

	// Check that the other repo was unaffected: first check local metadata and changelist
	requireRepoHasExpectedMetadata(t, longLivingRepo, data.CanonicalRootRole, true)
	requireRepoHasExpectedMetadata(t, longLivingRepo, data.CanonicalTargetsRole, true)
	requireRepoHasExpectedMetadata(t, longLivingRepo, data.CanonicalSnapshotRole, true)
	require.Len(t, longLivingCl.List(), 1)

	// Check that the other repo's remote data is unaffected
	remoteStore = longLivingRepo.getRemoteStore()
	require.NotNil(t, remoteStore)
	meta, err = remoteStore.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	require.NotNil(t, meta)
	meta, err = remoteStore.GetSized(data.CanonicalTargetsRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	require.NotNil(t, meta)
	meta, err = remoteStore.GetSized(data.CanonicalSnapshotRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	require.NotNil(t, meta)
	meta, err = remoteStore.GetSized(data.CanonicalTimestampRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	require.NotNil(t, meta)
}

// Test that we get a correct list of roles with keys and signatures
func TestListRoles(t *testing.T) {
	ts := fullTestServer(t)
	defer ts.Close()

	repo, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	require.NoError(t, repo.Publish())

	rolesWithSigs, err := repo.ListRoles()
	require.NoError(t, err)

	// Should only have base roles at this point
	require.Len(t, rolesWithSigs, len(data.BaseRoles))
	// Each base role should only have one key, one signature, and its key should match the signature's key
	for _, role := range rolesWithSigs {
		require.Len(t, role.Signatures, 1)
		require.Len(t, role.KeyIDs, 1)
		require.Equal(t, role.Signatures[0].KeyID, role.KeyIDs[0])
	}

	// Create a delegation on the top level
	aKey := createKey(t, repo, "user", true)
	require.NoError(t,
		repo.AddDelegation("targets/a", []data.PublicKey{aKey}, []string{""}),
		"error creating delegation")

	require.NoError(t, repo.Publish())

	rolesWithSigs, err = repo.ListRoles()
	require.NoError(t, err)

	require.Len(t, rolesWithSigs, len(data.BaseRoles)+1)
	// The delegation hasn't published any targets or metadata so it won't have a signature yet
	for _, role := range rolesWithSigs {
		if role.Name == "targets/a" {
			require.Nil(t, role.Signatures)
		} else {
			require.Len(t, role.Signatures, 1)
			require.Equal(t, role.Signatures[0].KeyID, role.KeyIDs[0])
		}
		require.Len(t, role.KeyIDs, 1)
	}

	addTarget(t, repo, "current", "../fixtures/intermediate-ca.crt", "targets/a")
	require.NoError(t, repo.Publish())

	rolesWithSigs, err = repo.ListRoles()
	require.NoError(t, err)

	require.Len(t, rolesWithSigs, len(data.BaseRoles)+1)
	// The delegation should have a signature now
	for _, role := range rolesWithSigs {
		require.Len(t, role.Signatures, 1)
		require.Equal(t, role.Signatures[0].KeyID, role.KeyIDs[0])
		require.Len(t, role.KeyIDs, 1)
	}

	// Create another delegation, one level further
	bKey := createKey(t, repo, "user", true)
	require.NoError(t,
		repo.AddDelegation("targets/a/b", []data.PublicKey{bKey}, []string{""}),
		"error creating delegation")

	require.NoError(t, repo.Publish())

	rolesWithSigs, err = repo.ListRoles()
	require.NoError(t, err)

	require.Len(t, rolesWithSigs, len(data.BaseRoles)+2)
	// The nested delegation hasn't published any targets or metadata so it won't have a signature yet
	for _, role := range rolesWithSigs {
		if role.Name == "targets/a/b" {
			require.Nil(t, role.Signatures)
		} else {
			require.Len(t, role.Signatures, 1)
			require.Equal(t, role.Signatures[0].KeyID, role.KeyIDs[0])
		}
		require.Len(t, role.KeyIDs, 1)
	}

	// Now make another repo and check that we don't pick up its roles
	repo2, _ := initializeRepo(t, data.ECDSAKey, "docker.com/notary2", ts.URL, false)
	defer os.RemoveAll(repo2.baseDir)

	require.NoError(t, repo2.Publish())

	// repo2 only has the base roles
	rolesWithSigs2, err := repo2.ListRoles()
	require.NoError(t, err)
	require.Len(t, rolesWithSigs2, len(data.BaseRoles))

	// original repo stays in same state (base roles + 2 delegations)
	rolesWithSigs, err = repo.ListRoles()
	require.NoError(t, err)
	require.Len(t, rolesWithSigs, len(data.BaseRoles)+2)
}

func TestGetAllTargetInfo(t *testing.T) {
	ts, mux, keys := simpleTestServer(t)
	defer ts.Close()

	rootType := data.ECDSAKey

	repo, _ := initializeRepo(t, rootType, "docker.com/notary", ts.URL, false)
	defer os.RemoveAll(repo.baseDir)

	// tests need to manually bootstrap timestamp as client doesn't generate it
	err := repo.tufRepo.InitTimestamp()
	require.NoError(t, err, "error creating repository: %s", err)

	// add latest and current to targets role
	targetsLatestTarget := addTarget(t, repo, "latest", "../fixtures/intermediate-ca.crt")
	targetsCurrentTarget := addTarget(t, repo, "current", "../fixtures/intermediate-ca.crt")

	// setup delegated targets/level1 role with targets current and other
	k, err := repo.CryptoService.Create("targets/level1", repo.gun, rootType)
	require.NoError(t, err)
	key1 := k
	err = repo.tufRepo.UpdateDelegationKeys("targets/level1", []data.PublicKey{k}, []string{}, 1)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1", []string{""}, []string{}, false)
	require.NoError(t, err)
	level1CurrentTarget := addTarget(t, repo, "current", "../fixtures/root-ca.crt", "targets/level1")
	level1OtherTarget := addTarget(t, repo, "other", "../fixtures/root-ca.crt", "targets/level1")

	// setup delegated targets/level2 role with targets current and level2
	k, err = repo.CryptoService.Create("targets/level2", repo.gun, rootType)
	require.NoError(t, err)
	key2 := k
	err = repo.tufRepo.UpdateDelegationKeys("targets/level2", []data.PublicKey{k}, []string{}, 1)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level2", []string{""}, []string{}, false)
	require.NoError(t, err)
	level2CurrentTarget := addTarget(t, repo, "current", "../fixtures/notary-server.crt", "targets/level2")
	level2Level2Target := addTarget(t, repo, "level2", "../fixtures/notary-server.crt", "targets/level2")

	// Apply the changelist. Normally, this would be done by Publish

	// load the changelist for this repo
	cl, err := changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")

	// apply the changelist to the repo, then clear it
	err = applyChangelist(repo.tufRepo, nil, cl)
	require.NoError(t, err, "could not apply changelist")
	require.NoError(t, cl.Clear(""))

	_, ok := repo.tufRepo.Targets["targets/level1"].Signed.Targets["current"]
	require.True(t, ok)
	_, ok = repo.tufRepo.Targets["targets/level1"].Signed.Targets["other"]
	require.True(t, ok)
	_, ok = repo.tufRepo.Targets["targets/level2"].Signed.Targets["level2"]
	require.True(t, ok)

	// setup delegated targets/level1/level2 role separately, which can only modify paths prefixed with "level2"
	// add level2 to targets/level1/level2
	k, err = repo.CryptoService.Create("targets/level1/level2", repo.gun, rootType)
	require.NoError(t, err)
	key3 := k
	err = repo.tufRepo.UpdateDelegationKeys("targets/level1/level2", []data.PublicKey{k}, []string{}, 1)
	require.NoError(t, err)
	err = repo.tufRepo.UpdateDelegationPaths("targets/level1/level2", []string{"level2"}, []string{}, false)
	require.NoError(t, err)
	level1Level2Level2Target := addTarget(t, repo, "level2", "../fixtures/notary-server.crt", "targets/level1/level2")
	// load the changelist for this repo
	cl, err = changelist.NewFileChangelist(
		filepath.Join(repo.baseDir, "tuf", filepath.FromSlash(repo.gun.String()), "changelist"))
	require.NoError(t, err, "could not open changelist")
	// apply the changelist to the repo
	err = applyChangelist(repo.tufRepo, nil, cl)
	require.NoError(t, err, "could not apply changelist")
	// check the changelist was applied
	_, ok = repo.tufRepo.Targets["targets/level1/level2"].Signed.Targets["level2"]
	require.True(t, ok)

	fakeServerData(t, repo, mux, keys)

	var (
		targetCurrent      = expectation{role: data.CanonicalTargetsRole.String(), target: "current"}
		targetLatest       = expectation{role: data.CanonicalTargetsRole.String(), target: "latest"}
		level1Current      = expectation{role: "targets/level1", target: "current"}
		level1Other        = expectation{role: "targets/level1", target: "other"}
		level2Current      = expectation{role: "targets/level2", target: "current"}
		level2Level2       = expectation{role: "targets/level2", target: "level2"}
		level1Level2Level2 = expectation{role: "targets/level1/level2", target: "level2"}
	)
	targetsKey := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)[0]
	allExpected := map[expectation]TargetSignedStruct{
		targetCurrent: {
			Target: *targetsCurrentTarget,
			Signatures: []data.Signature{
				{KeyID: targetsKey},
			},
		},
		targetLatest: {
			Target: *targetsLatestTarget,
			Signatures: []data.Signature{
				{KeyID: targetsKey},
			},
		},
		level1Current: {
			Target: *level1CurrentTarget,
			Signatures: []data.Signature{
				{KeyID: key1.ID()},
			},
		},
		level1Other: {
			Target: *level1OtherTarget,
			Signatures: []data.Signature{
				{KeyID: key1.ID()},
			},
		},
		level2Current: {
			Target: *level2CurrentTarget,
			Signatures: []data.Signature{
				{KeyID: key2.ID()},
			},
		},
		level2Level2: {
			Target: *level2Level2Target,
			Signatures: []data.Signature{
				{KeyID: key2.ID()},
			},
		},
		level1Level2Level2: {
			Target: *level1Level2Level2Target,
			Signatures: []data.Signature{
				{KeyID: key3.ID()},
			},
		},
	}

	// At this point, we have the following view of targets:
	// current - signed by targets, targets/level1, and targets/level2, all with different hashes
	// other - signed by targets/level1
	// latest - signed by targets
	// level2 - signed by targets/level2 and targets/level1/level2, with the same hash

	// Positive cases
	targetSignatureData, err := repo.GetAllTargetMetadataByName("current")
	require.NoError(t, err)

	checkSignatures(t, targetSignatureData, []expectation{targetCurrent, level1Current, level2Current}, allExpected)

	targetSignatureData, err = repo.GetAllTargetMetadataByName("other")
	require.NoError(t, err)

	checkSignatures(t, targetSignatureData, []expectation{level1Other}, allExpected)

	targetSignatureData, err = repo.GetAllTargetMetadataByName("latest")
	require.NoError(t, err)

	checkSignatures(t, targetSignatureData, []expectation{targetLatest}, allExpected)

	targetSignatureData, err = repo.GetAllTargetMetadataByName("level2")
	require.NoError(t, err)

	checkSignatures(t, targetSignatureData, []expectation{level2Level2, level1Level2Level2}, allExpected)

	// calling with the empty string "" name will get us back all targets signed in all roles
	targetSignatureData, err = repo.GetAllTargetMetadataByName("")
	require.NoError(t, err)
	require.Len(t, targetSignatureData, 7)

	checkSignatures(
		t,
		targetSignatureData,
		[]expectation{
			targetCurrent, targetLatest, level1Current, level1Other, level2Current, level2Level2, level1Level2Level2,
		},
		allExpected,
	)

	// nonexistent targets
	targetSignatureData, err = repo.GetAllTargetMetadataByName("level23")
	require.Error(t, err)
	require.Nil(t, targetSignatureData)
	targetSignatureData, err = repo.GetAllTargetMetadataByName("invalid")
	require.Error(t, err)
	require.Nil(t, targetSignatureData)
}

func checkSignatures(t *testing.T, targetSignatureData []TargetSignedStruct, expected []expectation, allExpected map[expectation]TargetSignedStruct) {
	makeSureWeHitEachCase := make(map[expectation]struct{})

	for _, tarSigStr := range targetSignatureData {
		dataPoint := expectation{role: tarSigStr.Role.Name.String(), target: tarSigStr.Target.Name}
		exp, ok := allExpected[dataPoint]
		require.True(t, ok)
		require.Equal(t, exp.Target, tarSigStr.Target)
		require.Len(t, tarSigStr.Signatures, 1)
		require.Equal(t, exp.Signatures[0].KeyID, tarSigStr.Signatures[0].KeyID)
		makeSureWeHitEachCase[dataPoint] = struct{}{}
	}
	for _, e := range expected {
		_, ok := makeSureWeHitEachCase[e]
		require.True(t, ok)
	}

}

type expectation struct {
	role, target string
}
