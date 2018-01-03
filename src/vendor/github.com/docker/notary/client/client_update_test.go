package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary"
	"github.com/docker/notary/passphrase"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/testutils"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func newBlankRepo(t *testing.T, url string) *NotaryRepository {
	// Temporary directory where test files will be created
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)

	repo, err := NewFileCachedNotaryRepository(tempBaseDir, "docker.com/notary", url,
		http.DefaultTransport, passphrase.ConstantRetriever("pass"), trustpinning.TrustPinConfig{})
	require.NoError(t, err)
	return repo
}

var metadataDelegations = []data.RoleName{"targets/a", "targets/a/b", "targets/b", "targets/a/b/c", "targets/b/c"}
var delegationsWithNonEmptyMetadata = []data.RoleName{"targets/a", "targets/a/b", "targets/b"}

func newServerSwizzler(t *testing.T) (map[data.RoleName][]byte, *testutils.MetadataSwizzler) {
	serverMeta, cs, err := testutils.NewRepoMetadata("docker.com/notary", metadataDelegations...)
	require.NoError(t, err)

	serverSwizzler := testutils.NewMetadataSwizzler("docker.com/notary", serverMeta, cs)
	require.NoError(t, err)

	return serverMeta, serverSwizzler
}

// bumps the versions of everything in the metadata cache, to force local cache
// to update
func bumpVersions(t *testing.T, s *testutils.MetadataSwizzler, offset int) {
	// bump versions of everything on the server, to force everything to update
	for _, r := range s.Roles {
		require.NoError(t, s.OffsetMetadataVersion(r, offset))
	}
	require.NoError(t, s.UpdateSnapshotHashes())
	require.NoError(t, s.UpdateTimestampHash())
}

// create a server that just serves static metadata files from a metaStore
func readOnlyServer(t *testing.T, cache store.MetadataStore, notFoundStatus int, gun data.GUN) *httptest.Server {
	m := mux.NewRouter()
	handler := func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var role string
		if vars["version"] != "" {
			role = fmt.Sprintf("%s.%s", vars["version"], vars["role"])
		} else {
			role = vars["role"]
		}
		metaBytes, err := cache.GetSized(role, store.NoSizeLimit)
		if _, ok := err.(store.ErrMetaNotFound); ok {
			w.WriteHeader(notFoundStatus)
		} else {
			require.NoError(t, err)
			w.Write(metaBytes)
		}
	}
	m.HandleFunc(fmt.Sprintf("/v2/%s/_trust/tuf/{version:[0-9]+}.{role:.*}.json", gun), handler)
	m.HandleFunc(fmt.Sprintf("/v2/%s/_trust/tuf/{role:.*}.{checksum:.*}.json", gun), handler)
	m.HandleFunc(fmt.Sprintf("/v2/%s/_trust/tuf/{role:.*}.json", gun), handler)
	return httptest.NewServer(m)
}

type unwritableStore struct {
	store.MetadataStore
	roleToNotWrite data.RoleName
}

func (u *unwritableStore) Set(role string, serverMeta []byte) error {
	if role == u.roleToNotWrite.String() {
		return fmt.Errorf("Non-writable")
	}
	return u.MetadataStore.Set(role, serverMeta)
}

// Update can succeed even if we cannot write any metadata to the repo (assuming
// no data in the repo)
func TestUpdateSucceedsEvenIfCannotWriteNewRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	serverMeta, _, err := testutils.NewRepoMetadata("docker.com/notary", metadataDelegations...)
	require.NoError(t, err)

	ts := readOnlyServer(t, store.NewMemoryStore(serverMeta), http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	for role := range serverMeta {
		repo := newBlankRepo(t, ts.URL)
		repo.cache = &unwritableStore{MetadataStore: repo.cache, roleToNotWrite: role}
		err := repo.Update(false)
		require.NoError(t, err)

		for r, expected := range serverMeta {
			actual, err := repo.cache.GetSized(r.String(), store.NoSizeLimit)
			if r == role {
				require.Error(t, err)
				require.IsType(t, store.ErrMetaNotFound{}, err,
					"expected no data because unable to write for %s", role)
			} else {
				require.NoError(t, err, "problem getting repo metadata for %s", r)
				require.True(t, bytes.Equal(expected, actual),
					"%s: expected to update since only %s was unwritable", r, role)
			}
		}

		os.RemoveAll(repo.baseDir)
	}
}

// Update can succeed even if we cannot write any metadata to the repo (assuming
// existing data in the repo)
func TestUpdateSucceedsEvenIfCannotWriteExistingRepo(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	serverMeta, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	// download existing metadata
	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	err := repo.Update(false)
	require.NoError(t, err)

	origFileStore := repo.cache

	for role := range serverMeta {
		for _, forWrite := range []bool{true, false} {
			// bump versions of everything on the server, to force everything to update
			bumpVersions(t, serverSwizzler, 1)

			// update fileStore
			repo.cache = &unwritableStore{MetadataStore: origFileStore, roleToNotWrite: role}
			err := repo.Update(forWrite)

			require.NoError(t, err)

			for r := range serverMeta {
				expected, err := serverSwizzler.MetadataCache.GetSized(r.String(), store.NoSizeLimit)
				require.NoError(t, err)
				if r != data.CanonicalRootRole && strings.Contains(r.String(), "root") {
					// don't fetch versioned root roles here
					continue
				}
				if strings.ContainsAny(r.String(), "123456789") {
					continue
				}
				actual, err := repo.cache.GetSized(r.String(), store.NoSizeLimit)
				require.NoError(t, err, "problem getting repo metadata for %s", r.String())
				if role == r {
					require.False(t, bytes.Equal(expected, actual),
						"%s: expected to not update because %s was unwritable", r.String(), role)
				} else {
					require.True(t, bytes.Equal(expected, actual),
						"%s: expected to update since only %s was unwritable", r.String(), role)
				}
			}
		}
	}
}

// If there is no local cache, update will error if it can't connect to the server.  Otherwise
// it uses the local cache.
func TestUpdateInOfflineMode(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// invalid URL, no cache - errors
	invalidURLRepo := newBlankRepo(t, "https://nothisdoesnotexist.com")
	defer os.RemoveAll(invalidURLRepo.baseDir)
	err := invalidURLRepo.Update(false)
	require.Error(t, err)
	require.IsType(t, store.NetworkError{}, err)

	// offline client: no cache - errors
	tempBaseDir, err := ioutil.TempDir("", "notary-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	defer os.RemoveAll(tempBaseDir)

	offlineRepo, err := NewFileCachedNotaryRepository(tempBaseDir, "docker.com/notary", "https://nope",
		nil, passphrase.ConstantRetriever("pass"), trustpinning.TrustPinConfig{})
	require.NoError(t, err)
	err = offlineRepo.Update(false)
	require.Error(t, err)
	require.IsType(t, store.ErrOffline{}, err)

	// set existing metadata on the repo
	serverMeta, _, err := testutils.NewRepoMetadata("docker.com/notary", metadataDelegations...)
	require.NoError(t, err)
	for name, metaBytes := range serverMeta {
		require.NoError(t, invalidURLRepo.cache.Set(name.String(), metaBytes))
		require.NoError(t, offlineRepo.cache.Set(name.String(), metaBytes))
	}

	// both of these can read from cache and load repo
	require.NoError(t, invalidURLRepo.Update(false))
	require.NoError(t, offlineRepo.Update(false))
}

type swizzleFunc func(*testutils.MetadataSwizzler, data.RoleName) error
type swizzleExpectations struct {
	desc       string
	swizzle    swizzleFunc
	expectErrs []interface{}
}

// the errors here are only relevant for root - we bail if the root is corrupt, but
// other metadata will be replaced
var waysToMessUpLocalMetadata = []swizzleExpectations{
	// for instance if the metadata got truncated or otherwise block corrupted
	{desc: "invalid JSON", swizzle: (*testutils.MetadataSwizzler).SetInvalidJSON,
		expectErrs: []interface{}{&json.SyntaxError{}}},
	// if the metadata was accidentally deleted
	{desc: "missing metadata", swizzle: (*testutils.MetadataSwizzler).RemoveMetadata,
		expectErrs: []interface{}{store.ErrMetaNotFound{}, ErrRepoNotInitialized{}, ErrRepositoryNotExist{}}},
	// if the signature was invalid - maybe the user tried to modify something manually
	// that they forgot (add a key, or something)
	{desc: "signed with right key but wrong hash",
		swizzle:    (*testutils.MetadataSwizzler).InvalidateMetadataSignatures,
		expectErrs: []interface{}{&trustpinning.ErrValidationFail{}, signed.ErrRoleThreshold{}}},
	// if the user copied the wrong root.json over it by accident or something
	{desc: "signed with wrong key", swizzle: (*testutils.MetadataSwizzler).SignMetadataWithInvalidKey,
		expectErrs: []interface{}{&trustpinning.ErrValidationFail{}, signed.ErrRoleThreshold{}}},
	// self explanatory
	{desc: "expired metadata", swizzle: (*testutils.MetadataSwizzler).ExpireMetadata,
		expectErrs: []interface{}{signed.ErrExpired{}}},

	// Not trying any of the other repoSwizzler methods, because those involve modifying
	// and re-serializing, and that means a user has the root and other keys and was trying to
	// actively sabotage and break their own local repo (particularly the root.json)
}

// If a repo has missing metadata, an update will replace all of it
// If a repo has corrupt metadata for root, the update will fail
// For other roles, corrupt metadata will be replaced
func TestUpdateReplacesCorruptOrMissingMetadata(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	serverMeta, cs, err := testutils.NewRepoMetadata("docker.com/notary", metadataDelegations...)
	require.NoError(t, err)

	ts := readOnlyServer(t, store.NewMemoryStore(serverMeta), http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	err = repo.Update(false) // ensure we have all metadata to start with
	require.NoError(t, err)

	// we want to swizzle the local cache, not the server, so create a new one
	repoSwizzler := testutils.NewMetadataSwizzler("docker.com/notary", serverMeta, cs)
	repoSwizzler.MetadataCache = repo.cache

	origMeta := testutils.CopyRepoMetadata(serverMeta)

	for _, role := range repoSwizzler.Roles {
		for _, expt := range waysToMessUpLocalMetadata {
			text, messItUp := expt.desc, expt.swizzle
			for _, forWrite := range []bool{true, false} {
				require.NoError(t, messItUp(repoSwizzler, role), "could not fuzz %s (%s)", role, text)
				err := repo.Update(forWrite)
				// If this is a root role, we should error if it's corrupted or invalid data;
				// missing metadata is ok.
				if role == data.CanonicalRootRole && expt.desc != "missing metadata" &&
					expt.desc != "expired metadata" {

					require.Error(t, err, "%s for %s: expected to error when bootstrapping root", text, role)
					// revert our original metadata
					for role := range origMeta {
						require.NoError(t, repo.cache.Set(role.String(), origMeta[role]))
					}
				} else {
					require.NoError(t, err)
					for r, expected := range serverMeta {
						actual, err := repo.cache.GetSized(r.String(), store.NoSizeLimit)
						require.NoError(t, err, "problem getting repo metadata for %s", role)
						require.True(t, bytes.Equal(expected, actual),
							"%s for %s: expected to recover after update", text, role)
					}
				}
			}
		}
	}
}

// If a repo has an invalid root (signed by wrong key, expired, invalid version,
// invalid number of signatures, etc.), the repo will just get the new root from
// the server, whether or not the update is for writing (forced update), but
// it will refuse to update if the root key has changed and the new root is
// not signed by the old and new key
func TestUpdateFailsIfServerRootKeyChangedWithoutMultiSign(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	serverMeta, serverSwizzler := newServerSwizzler(t)
	origMeta := testutils.CopyRepoMetadata(serverMeta)

	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	err := repo.Update(false) // ensure we have all metadata to start with
	require.NoError(t, err)

	// rotate the server's root.json root key so that they no longer match trust anchors
	require.NoError(t, serverSwizzler.ChangeRootKey())
	// bump versions, update snapshot and timestamp too so it will not fail on a hash
	bumpVersions(t, serverSwizzler, 1)

	// we want to swizzle the local cache, not the server, so create a new one
	repoSwizzler := &testutils.MetadataSwizzler{
		MetadataCache: repo.cache,
		CryptoService: serverSwizzler.CryptoService,
		Roles:         serverSwizzler.Roles,
	}

	for _, expt := range waysToMessUpLocalMetadata {
		text, messItUp := expt.desc, expt.swizzle
		for _, forWrite := range []bool{true, false} {
			require.NoError(t, messItUp(repoSwizzler, data.CanonicalRootRole), "could not fuzz root (%s)", text)
			messedUpMeta, err := repo.cache.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)

			if _, ok := err.(store.ErrMetaNotFound); ok { // one of the ways to mess up is to delete metadata

				err = repo.Update(forWrite)
				// the new server has a different root key, but we don't have any way of pinning against a previous root
				require.NoError(t, err)
				// revert our original metadata
				for role := range origMeta {
					require.NoError(t, repo.cache.Set(role.String(), origMeta[role]))
				}
			} else {

				require.NoError(t, err)

				err = repo.Update(forWrite)
				require.Error(t, err) // the new server has a different root, update fails

				// we can't test that all the metadata is the same, because we probably would
				// have downloaded a new timestamp and maybe snapshot.  But the root should be the
				// same because it has failed to update.
				for role, expected := range origMeta {
					if role != data.CanonicalTimestampRole && role != data.CanonicalSnapshotRole {
						actual, err := repo.cache.GetSized(role.String(), store.NoSizeLimit)
						require.NoError(t, err, "problem getting repo metadata for %s", role.String())

						if role == data.CanonicalRootRole {
							expected = messedUpMeta
						}
						require.True(t, bytes.Equal(expected, actual),
							"%s for %s: expected to not have updated -- swizzle method %s", text, role, expt.desc)
					}
				}

			}

			// revert our original root metadata
			require.NoError(t,
				repo.cache.Set(data.CanonicalRootRole.String(), origMeta[data.CanonicalRootRole]))
		}
	}
}

type updateOpts struct {
	notFoundCode     int           // what code to return when the cache doesn't have the metadata
	serverHasNewData bool          // whether the server should have the same or new version than the local cache
	localCache       bool          // whether the repo should have a local cache before updating
	forWrite         bool          // whether the update is for writing or not (force check remote root.json)
	role             data.RoleName // the role to mess up on the server

	checkRepo func(*NotaryRepository, *testutils.MetadataSwizzler) // a callback that can examine the repo at the end
}

// If there's no local cache, we go immediately to check the remote server for
// root, and if it doesn't exist, we return ErrRepositoryNotExist. This happens
// with or without a force check (update for write).
func TestUpdateRemoteRootNotExistNoLocalCache(t *testing.T) {
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode: http.StatusNotFound,
		forWrite:     false,
		role:         data.CanonicalRootRole,
	}, ErrRepositoryNotExist{})
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode: http.StatusNotFound,
		forWrite:     true,
		role:         data.CanonicalRootRole,
	}, ErrRepositoryNotExist{})
}

// If there is a local cache, we use the local root as the trust anchor and we
// then an update. If the server has no root.json, and we don't need to force
// check (update for write), we can used the cached root because the timestamp
// has not changed.
// If we force check (update for write), then it hits the server first, and
// still returns an ErrRepositoryNotExist.  This is the
// case where the server has the same data as the client, in which case we might
// be able to just used the cached data and not have to download.
func TestUpdateRemoteRootNotExistCanUseLocalCache(t *testing.T) {
	// if for-write is false, then we don't need to check the root.json on bootstrap,
	// and hence we can just use the cached version on update
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode: http.StatusNotFound,
		localCache:   true,
		forWrite:     false,
		role:         data.CanonicalRootRole,
	}, nil)
	// fails because bootstrap requires a check to remote root.json and fails if
	// the check fails
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode: http.StatusNotFound,
		localCache:   true,
		forWrite:     true,
		role:         data.CanonicalRootRole,
	}, ErrRepositoryNotExist{})
}

// If there is a local cache, we use the local root as the trust anchor and we
// then an update. If the server has no root.json, we return an ErrRepositoryNotExist.
// If we force check (update for write), then it hits the server first, and
// still returns an ErrRepositoryNotExist. This is the case where the server
// has new updated data, so we cannot default to cached data.
func TestUpdateRemoteRootNotExistCannotUseLocalCache(t *testing.T) {
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode:     http.StatusNotFound,
		serverHasNewData: true,
		localCache:       true,
		forWrite:         false,
		role:             data.CanonicalRootRole,
	}, ErrRepositoryNotExist{})
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode:     http.StatusNotFound,
		serverHasNewData: true,
		localCache:       true,
		forWrite:         true,
		role:             data.CanonicalRootRole,
	}, ErrRepositoryNotExist{})
}

// If there's no local cache, we go immediately to check the remote server for
// root, and if it 50X's, we return ErrServerUnavailable. This happens
// with or without a force check (update for write).
func TestUpdateRemoteRoot50XNoLocalCache(t *testing.T) {
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode: http.StatusServiceUnavailable,
		forWrite:     false,
		role:         data.CanonicalRootRole,
	}, store.ErrServerUnavailable{})
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode: http.StatusServiceUnavailable,
		forWrite:     true,
		role:         data.CanonicalRootRole,
	}, store.ErrServerUnavailable{})
}

// If there is a local cache, we use the local root as the trust anchor and we
// then an update. If the server 50X's on root.json, and we don't force check,
// then because the timestamp is the same we can just use our cached root.json
// and don't have to download another.
// If we force check (update for write), we return an ErrServerUnavailable.
// This is the case where the server has the same data as the client
func TestUpdateRemoteRoot50XCanUseLocalCache(t *testing.T) {
	// if for-write is false, then we don't need to check the root.json on bootstrap,
	// and hence we can just use the cached version on update.
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode: http.StatusServiceUnavailable,
		localCache:   true,
		forWrite:     false,
		role:         data.CanonicalRootRole,
	}, nil)
	// fails because bootstrap requires a check to remote root.json and fails if
	// the check fails
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode: http.StatusServiceUnavailable,
		localCache:   true,
		forWrite:     true,
		role:         data.CanonicalRootRole,
	}, store.ErrServerUnavailable{})
}

// If there is a local cache, we use the local root as the trust anchor and we
// then an update. If the server 50X's on root.json, we return an ErrServerUnavailable.
// This happens with or without a force check (update for write)
func TestUpdateRemoteRoot50XCannotUseLocalCache(t *testing.T) {
	// if for-write is false, then we don't need to check the root.json on bootstrap,
	// and hence we can just use the cached version on update
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode:     http.StatusServiceUnavailable,
		serverHasNewData: true,
		localCache:       true,
		forWrite:         false,
		role:             data.CanonicalRootRole,
	}, store.ErrServerUnavailable{})
	// fails because of bootstrap
	testUpdateRemoteNon200Error(t, updateOpts{
		notFoundCode:     http.StatusServiceUnavailable,
		serverHasNewData: true,
		localCache:       true,
		forWrite:         true,
		role:             data.CanonicalRootRole,
	}, store.ErrServerUnavailable{})
}

// If there is no local cache, we just update. If the server has a root.json,
// but is missing other data, then we propagate the ErrMetaNotFound.  Skipping
// force check, because that only matters for root.
func TestUpdateNonRootRemoteMissingMetadataNoLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		if role == data.CanonicalRootRole {
			continue
		}
		testUpdateRemoteNon200Error(t, updateOpts{
			notFoundCode: http.StatusNotFound,
			role:         role,
		}, store.ErrMetaNotFound{})
	}
}

// If there is a local cache, we update anyway and see if anything's different
// (assuming remote has a root.json).  If the timestamp is missing, we use the
// local timestamp and already have all data, so nothing needs to be downloaded.
// If the timestamp is present, but the same, we already have all the data, so
// nothing needs to be downloaded.
// Skipping force check, because that only matters for root.
func TestUpdateNonRootRemoteMissingMetadataCanUseLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	// really we can delete everything at once except for the timestamp, but
	// it's better to check one by one in case we change the download code
	// somewhat.
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		if role == data.CanonicalRootRole {
			continue
		}
		testUpdateRemoteNon200Error(t, updateOpts{
			notFoundCode: http.StatusNotFound,
			localCache:   true,
			role:         role,
		}, nil)
	}
}

// If there is a local cache, we update anyway and see if anything's different
// (assuming remote has a root.json).  If the server has new data, we cannot
// use the local cache so if the server is missing any metadata we cannot update.
// Skipping force check, because that only matters for root.
func TestUpdateNonRootRemoteMissingMetadataCannotUseLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		if role == data.CanonicalRootRole {
			continue
		}
		var errExpected interface{} = store.ErrMetaNotFound{}
		if role == data.CanonicalTimestampRole {
			// if we can't download the timestamp, we use the cached timestamp.
			// it says that we have all the local data already, so we download
			// nothing.  So the update won't error, it will just fail to update
			// to the latest version.  We log a warning in this case.
			errExpected = nil
		}

		testUpdateRemoteNon200Error(t, updateOpts{
			notFoundCode:     http.StatusNotFound,
			serverHasNewData: true,
			localCache:       true,
			role:             role,
		}, errExpected)
	}
}

// If there is no local cache, we just update. If the server 50X's when getting
// metadata, we propagate ErrServerUnavailable.
func TestUpdateNonRootRemote50XNoLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		if role == data.CanonicalRootRole {
			continue
		}
		testUpdateRemoteNon200Error(t, updateOpts{
			notFoundCode: http.StatusServiceUnavailable,
			role:         role,
		}, store.ErrServerUnavailable{})
	}
}

// If there is a local cache, we update anyway and see if anything's different
// (assuming remote has a root.json).  If the timestamp is 50X's, we use the
// local timestamp and already have all data, so nothing needs to be downloaded.
// If the timestamp is present, but the same, we already have all the data, so
// nothing needs to be downloaded.
func TestUpdateNonRootRemote50XCanUseLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	// actually everything can error at once, but it's better to check one by
	// one in case we change the download code somewhat.
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		if role == data.CanonicalRootRole {
			continue
		}
		testUpdateRemoteNon200Error(t, updateOpts{
			notFoundCode: http.StatusServiceUnavailable,
			localCache:   true,
			role:         role,
		}, nil)
	}
}

// If there is a local cache, we update anyway and see if anything's different
// (assuming remote has a root.json).  If the server has new data, we cannot
// use the local cache so if the server 50X's on any metadata we cannot update.
// This happens whether or not we force a remote check (because that's on the
// root)
func TestUpdateNonRootRemote50XCannotUseLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		if role == data.CanonicalRootRole {
			continue
		}

		var errExpected interface{} = store.ErrServerUnavailable{}
		if role == data.CanonicalTimestampRole {
			// if we can't download the timestamp, we use the cached timestamp.
			// it says that we have all the local data already, so we download
			// nothing.  So the update won't error, it will just fail to update
			// to the latest version.  We log a warning in this case.
			errExpected = nil
		}

		testUpdateRemoteNon200Error(t, updateOpts{
			notFoundCode:     http.StatusServiceUnavailable,
			serverHasNewData: true,
			localCache:       true,
			role:             role,
		}, errExpected)
	}
}

func testUpdateRemoteNon200Error(t *testing.T, opts updateOpts, errExpected interface{}) {
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, opts.notFoundCode, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	if opts.localCache {
		err := repo.Update(false) // acquire local cache
		require.NoError(t, err)
	}

	if opts.serverHasNewData {
		bumpVersions(t, serverSwizzler, 1)
	}

	require.NoError(t, serverSwizzler.RemoveMetadata(opts.role), "failed to remove %s", opts.role)

	err := repo.Update(opts.forWrite)
	if errExpected == nil {
		require.NoError(t, err, "expected no failure updating when %s is %v (forWrite: %v)",
			opts.role, opts.notFoundCode, opts.forWrite)
	} else {
		require.Error(t, err, "expected failure updating when %s is %v (forWrite: %v)",
			opts.role, opts.notFoundCode, opts.forWrite)
		require.IsType(t, errExpected, err, "wrong update error when %s is %v (forWrite: %v)",
			opts.role, opts.notFoundCode, opts.forWrite)
		if notFound, ok := err.(store.ErrMetaNotFound); ok {
			require.True(t, strings.HasPrefix(notFound.Resource, opts.role.String()), "wrong resource missing (forWrite: %v)", opts.forWrite)
		}
	}
}

// If there's no local cache, we go immediately to check the remote server for
// root. If the root is corrupted in transit in such a way that the signature is
// wrong, but it is correct in all other ways, then it validates during bootstrap,
// but it will fail validation during update. So it will fail with or without
// a force check (update for write).  If any of the other roles (except
// timestamp, because there is no checksum for that) are corrupted in the same
// way, they will also fail during update with the same error.
func TestUpdateRemoteChecksumWrongNoLocalCache(t *testing.T) {
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		testUpdateRemoteFileChecksumWrong(t, updateOpts{
			serverHasNewData: false,
			localCache:       false,
			forWrite:         false,
			role:             role,
		}, role != data.CanonicalTimestampRole) // timestamp role should not fail

		if role == data.CanonicalRootRole {
			testUpdateRemoteFileChecksumWrong(t, updateOpts{
				serverHasNewData: false,
				localCache:       false,
				forWrite:         true,
				role:             role,
			}, true)
		}
	}
}

// If there's is a local cache, and the remote server has the same data (except
// corrupted), then we can just use our local cache.  So update succeeds (
// with or without a force check (update for write))
func TestUpdateRemoteChecksumWrongCanUseLocalCache(t *testing.T) {
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		testUpdateRemoteFileChecksumWrong(t, updateOpts{
			serverHasNewData: false,
			localCache:       true,
			forWrite:         false,
			role:             role,
		}, false)

		if role == data.CanonicalRootRole {
			testUpdateRemoteFileChecksumWrong(t, updateOpts{
				serverHasNewData: false,
				localCache:       true,
				forWrite:         true,
				role:             role,
			}, false)
		}
	}
}

// If there's is a local cache, but the remote server has new data (some
// corrupted), we go immediately to check the remote server for root.  If the
// root is corrupted in transit in such a way that the signature is wrong, but
// it is correct in all other ways, it from validates during bootstrap,
// but it will fail validation during update. So it will fail with or without
// a force check (update for write).  If any of the other roles (except
// timestamp, because there is no checksum for that) is corrupted in the same
// way, they will also fail during update with the same error.
func TestUpdateRemoteChecksumWrongCannotUseLocalCache(t *testing.T) {
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		testUpdateRemoteFileChecksumWrong(t, updateOpts{
			serverHasNewData: true,
			localCache:       true,
			forWrite:         false,
			role:             role,
		}, role != data.CanonicalTimestampRole) // timestamp role should not fail

		if role == data.CanonicalRootRole {
			testUpdateRemoteFileChecksumWrong(t, updateOpts{
				serverHasNewData: true,
				localCache:       true,
				forWrite:         true,
				role:             role,
			}, true)
		}
	}
}

func testUpdateRemoteFileChecksumWrong(t *testing.T, opts updateOpts, errExpected bool) {
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	if opts.localCache {
		err := repo.Update(false) // acquire local cache
		require.NoError(t, err)
	}

	if opts.serverHasNewData {
		bumpVersions(t, serverSwizzler, 1)
	}

	require.NoError(t, serverSwizzler.AddExtraSpace(opts.role), "failed to checksum-corrupt to %s", opts.role)

	err := repo.Update(opts.forWrite)
	if !errExpected {
		require.NoError(t, err, "expected no failure updating when %s has the wrong checksum (forWrite: %v)",
			opts.role, opts.forWrite)
	} else {
		require.Error(t, err, "expected failure updating when %s has the wrong checksum (forWrite: %v)",
			opts.role, opts.forWrite)

		// it could be ErrMaliciousServer (if the server sent the metadata with a content length)
		// or a checksum error (if the server didn't set content length because transfer-encoding
		// was specified).  For the timestamp, which is really short, it should be the content-length.

		var rightError bool
		if opts.role == data.CanonicalTimestampRole {
			_, rightError = err.(store.ErrMaliciousServer)
		} else {
			_, isErrChecksum := err.(data.ErrMismatchedChecksum)
			_, isErrMaliciousServer := err.(store.ErrMaliciousServer)
			rightError = isErrChecksum || isErrMaliciousServer
		}
		require.True(t, rightError,
			"wrong update error (%v) when %s has the wrong checksum (forWrite: %v)",
			reflect.TypeOf(err), opts.role, opts.forWrite)
	}
}

// --- these tests below assume the checksums are correct (since the server can sign snapshots and
// timestamps, so can be malicious) ---

var waysToMessUpServerBadMeta = []swizzleExpectations{
	{desc: "invalid JSON", expectErrs: []interface{}{&trustpinning.ErrValidationFail{}, &json.SyntaxError{}},
		swizzle: (*testutils.MetadataSwizzler).SetInvalidJSON},

	{desc: "an invalid Signed", expectErrs: []interface{}{&trustpinning.ErrValidationFail{}, &json.UnmarshalTypeError{}},
		swizzle: (*testutils.MetadataSwizzler).SetInvalidSigned},

	{desc: "an invalid SignedMeta",
		// it depends which field gets unmarshalled first
		expectErrs: []interface{}{&trustpinning.ErrValidationFail{}, &json.UnmarshalTypeError{}, &time.ParseError{}},
		swizzle:    (*testutils.MetadataSwizzler).SetInvalidSignedMeta},

	// for everything else, the errors come from tuf/signed

	{desc: "invalid SignedMeta Type", expectErrs: []interface{}{
		&trustpinning.ErrValidationFail{}, signed.ErrWrongType, data.ErrInvalidMetadata{}},
		swizzle: (*testutils.MetadataSwizzler).SetInvalidMetadataType},

	{desc: "lower metadata version", expectErrs: []interface{}{
		&trustpinning.ErrValidationFail{}, signed.ErrLowVersion{}, data.ErrInvalidMetadata{}},
		swizzle: func(s *testutils.MetadataSwizzler, role data.RoleName) error {
			return s.OffsetMetadataVersion(role, -3)
		}},
}

var waysToMessUpServerBadSigs = []swizzleExpectations{
	{desc: "invalid signatures", expectErrs: []interface{}{
		&trustpinning.ErrValidationFail{}, signed.ErrRoleThreshold{}, &trustpinning.ErrRootRotationFail{}},
		swizzle: (*testutils.MetadataSwizzler).InvalidateMetadataSignatures},

	{desc: "meta signed by wrong key", expectErrs: []interface{}{
		&trustpinning.ErrValidationFail{}, signed.ErrRoleThreshold{}, &trustpinning.ErrRootRotationFail{}},
		swizzle: (*testutils.MetadataSwizzler).SignMetadataWithInvalidKey},

	{desc: "insufficient signatures", expectErrs: []interface{}{
		&trustpinning.ErrValidationFail{}, signed.ErrRoleThreshold{}},
		swizzle: func(s *testutils.MetadataSwizzler, role data.RoleName) error {
			return s.SetThreshold(role, 2)
		}},
}

var wayToMessUpServerBadExpiry = swizzleExpectations{
	desc: "expired metadata", expectErrs: []interface{}{
		&trustpinning.ErrValidationFail{}, signed.ErrExpired{}},
	swizzle: (*testutils.MetadataSwizzler).ExpireMetadata,
}

// this does not include delete, which is tested separately so we can try to get
// 404s and 503s
var waysToMessUpServer = append(waysToMessUpServerBadMeta, append(waysToMessUpServerBadSigs, wayToMessUpServerBadExpiry)...)

var _waysToMessUpServerRoot []swizzleExpectations

// We also want to remove a every role from root once, or remove the role's keys.
// This function generates once and caches the result for later re-use.
func waysToMessUpServerRoot() []swizzleExpectations {
	if _waysToMessUpServerRoot == nil {
		_waysToMessUpServerRoot = waysToMessUpServer
		for _, roleName := range data.BaseRoles {
			_waysToMessUpServerRoot = append(_waysToMessUpServerRoot,
				swizzleExpectations{
					desc: fmt.Sprintf("no %s keys", roleName),
					expectErrs: []interface{}{
						&trustpinning.ErrValidationFail{}, signed.ErrRoleThreshold{}},
					swizzle: func(s *testutils.MetadataSwizzler, role data.RoleName) error {
						return s.MutateRoot(func(r *data.Root) {
							r.Roles[roleName].KeyIDs = []string{}
						})
					}},
				swizzleExpectations{
					desc:       fmt.Sprintf("no %s role", roleName),
					expectErrs: []interface{}{data.ErrInvalidMetadata{}},
					swizzle: func(s *testutils.MetadataSwizzler, role data.RoleName) error {
						return s.MutateRoot(func(r *data.Root) { delete(r.Roles, roleName) })
					}},
			)
		}
	}
	return _waysToMessUpServerRoot
}

// If there's no local cache, we go immediately to check the remote server for
// root, and if it invalid (corrupted), we cannot update.  This happens
// with and without a force check (update for write).
func TestUpdateRootRemoteCorruptedNoLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	for _, testData := range waysToMessUpServerRoot() {
		testUpdateRemoteCorruptValidChecksum(t, updateOpts{
			forWrite: false,
			role:     data.CanonicalRootRole,
		}, testData, true)
		testUpdateRemoteCorruptValidChecksum(t, updateOpts{
			forWrite: true,
			role:     data.CanonicalRootRole,
		}, testData, true)
	}
}

// Having a local cache, if the server has the same data (timestamp has not changed),
// should succeed in all cases if whether forWrite (force check) is true or not
// because the fact that the timestamp hasn't changed should mean that we don't
// have to re-download the root.
func TestUpdateRootRemoteCorruptedCanUseLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	for _, testData := range waysToMessUpServerRoot() {
		testUpdateRemoteCorruptValidChecksum(t, updateOpts{
			localCache: true,
			forWrite:   false,
			role:       data.CanonicalRootRole,
		}, testData, false)
		testUpdateRemoteCorruptValidChecksum(t, updateOpts{
			localCache: true,
			forWrite:   true,
			role:       data.CanonicalRootRole,
		}, testData, false)
	}
}

// Having a local cache, if the server has new same data should fail in all cases
// because the metadata is re-downloaded.
func TestUpdateRootRemoteCorruptedCannotUseLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, testData := range waysToMessUpServerRoot() {
		testUpdateRemoteCorruptValidChecksum(t, updateOpts{
			serverHasNewData: true,
			localCache:       true,
			forWrite:         false,
			role:             data.CanonicalRootRole,
		}, testData, true)
		testUpdateRemoteCorruptValidChecksum(t, updateOpts{
			serverHasNewData: true,
			localCache:       true,
			forWrite:         true,
			role:             data.CanonicalRootRole,
		}, testData, true)
	}
}

func waysToMessUpServerNonRootPerRole(t *testing.T) map[string][]swizzleExpectations {
	perRoleSwizzling := make(map[string][]swizzleExpectations)
	for _, missing := range []data.RoleName{data.CanonicalRootRole, data.CanonicalTargetsRole} {
		perRoleSwizzling[data.CanonicalSnapshotRole.String()] = append(
			perRoleSwizzling[data.CanonicalSnapshotRole.String()],
			swizzleExpectations{
				desc:       fmt.Sprintf("snapshot missing root meta checksum"),
				expectErrs: []interface{}{data.ErrInvalidMetadata{}},
				swizzle: func(s *testutils.MetadataSwizzler, role data.RoleName) error {
					return s.MutateSnapshot(func(sn *data.Snapshot) {
						delete(sn.Meta, missing.String())
					})
				},
			})
	}
	perRoleSwizzling[data.CanonicalTargetsRole.String()] = []swizzleExpectations{{
		desc:       fmt.Sprintf("target missing delegations data"),
		expectErrs: []interface{}{data.ErrMismatchedChecksum{}},
		swizzle: func(s *testutils.MetadataSwizzler, role data.RoleName) error {
			return s.MutateTargets(func(tg *data.Targets) {
				tg.Delegations.Roles = tg.Delegations.Roles[1:]
			})
		},
	}}
	perRoleSwizzling[data.CanonicalTimestampRole.String()] = []swizzleExpectations{{
		desc:       fmt.Sprintf("timestamp missing snapshot meta checksum"),
		expectErrs: []interface{}{data.ErrInvalidMetadata{}},
		swizzle: func(s *testutils.MetadataSwizzler, role data.RoleName) error {
			return s.MutateTimestamp(func(ts *data.Timestamp) {
				delete(ts.Meta, data.CanonicalSnapshotRole.String())
			})
		},
	}}
	perRoleSwizzling["targets/a"] = []swizzleExpectations{{
		desc:       fmt.Sprintf("delegation has invalid role"),
		expectErrs: []interface{}{data.ErrInvalidMetadata{}},
		swizzle: func(s *testutils.MetadataSwizzler, role data.RoleName) error {
			return s.MutateTargets(func(tg *data.Targets) {
				var keyIDs []string
				for k := range tg.Delegations.Keys {
					keyIDs = append(keyIDs, k)
				}
				// add the keys from root too
				rootMeta, err := s.MetadataCache.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
				require.NoError(t, err)

				signedRoot := &data.SignedRoot{}
				require.NoError(t, json.Unmarshal(rootMeta, signedRoot))

				for k := range signedRoot.Signed.Keys {
					keyIDs = append(keyIDs, k)
				}

				// add an invalid role (root) to delegation
				tg.Delegations.Roles = append(tg.Delegations.Roles,
					&data.Role{RootRole: data.RootRole{KeyIDs: keyIDs, Threshold: 1},
						Name: data.CanonicalRootRole})
			})
		},
	}}
	return perRoleSwizzling
}

// If there's no local cache, we just download from the server and if anything
// is corrupt, we cannot update.
func TestUpdateNonRootRemoteCorruptedNoLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	for _, role := range append(data.BaseRoles) {
		switch role {
		case data.CanonicalRootRole:
			break
		default:
			for _, testData := range waysToMessUpServer {
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					role: role,
				}, testData, true)
			}
		}
	}

	for _, role := range delegationsWithNonEmptyMetadata {
		for _, testData := range waysToMessUpServerBadMeta {
			testUpdateRemoteCorruptValidChecksum(t, updateOpts{
				role: role,
			}, testData, true)
		}

		for _, testData := range append(waysToMessUpServerBadSigs, wayToMessUpServerBadExpiry) {
			testUpdateRemoteCorruptValidChecksum(t, updateOpts{
				role:      role,
				checkRepo: checkBadDelegationRoleSkipped(t, role.String()),
			}, testData, false)
		}
	}

	for role, expectations := range waysToMessUpServerNonRootPerRole(t) {
		for _, testData := range expectations {
			roleName := data.RoleName(role)
			switch roleName {
			case data.CanonicalSnapshotRole:
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					role: roleName,
				}, testData, true)
			case data.CanonicalTargetsRole:
				// if there are no delegation target roles, we're fine, we just don't
				// download them
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					role: roleName,
				}, testData, false)
			case data.CanonicalTimestampRole:
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					role: roleName,
				}, testData, true)
			case data.RoleName("targets/a"):
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					role: roleName,
				}, testData, true)
			}
		}
	}
}

// Having a local cache, if the server has the same data (timestamp has not changed),
// should succeed in all cases if whether forWrite (force check) is true or not.
// If the timestamp is fine, it hasn't changed and we don't have to download
// anything. If it's broken, we used the cached timestamp only if the error on
// downloading the new one was not validation related
func TestUpdateNonRootRemoteCorruptedCanUseLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		if role == data.CanonicalRootRole {
			continue
		}
		for _, testData := range waysToMessUpServer {
			// remote timestamp swizzling will fail the update
			if role == data.CanonicalTimestampRole {
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					localCache: true,
					role:       role,
				}, testData, testData.desc != "insufficient signatures")
			} else {
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					localCache: true,
					role:       role,
				}, testData, false)
			}
		}
	}
	for role, expectations := range waysToMessUpServerNonRootPerRole(t) {
		for _, testData := range expectations {
			roleName := data.RoleName(role)
			switch roleName {
			case data.CanonicalSnapshotRole:
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					localCache: true,
					role:       roleName,
				}, testData, false)
			case data.CanonicalTargetsRole:
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					localCache: true,
					role:       roleName,
				}, testData, false)
			case data.CanonicalTimestampRole:
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					localCache: true,
					role:       roleName,
				}, testData, true)
			case data.RoleName("targets/a"):
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					localCache: true,
					role:       roleName,
				}, testData, false)
			}
		}
	}
}

// requires that a delegation role and its descendants were not accepted as a valid part of the
// TUF repo, but everything else was
func checkBadDelegationRoleSkipped(t *testing.T, delgRoleName string) func(*NotaryRepository, *testutils.MetadataSwizzler) {
	return func(repo *NotaryRepository, s *testutils.MetadataSwizzler) {
		for _, roleName := range s.Roles {
			if roleName != data.CanonicalTargetsRole && !data.IsDelegation(roleName) {
				continue
			}
			_, hasTarget := repo.tufRepo.Targets[roleName]
			require.Equal(t, !strings.HasPrefix(roleName.String(), delgRoleName), hasTarget)
		}

		require.NotNil(t, repo.tufRepo.Root)
		require.NotNil(t, repo.tufRepo.Snapshot)
		require.NotNil(t, repo.tufRepo.Timestamp)
	}
}

// Having a local cache, if the server has all new data (some being corrupt),
// and update should fail in all cases (except if we modify the timestamp)
// because the metadata is re-downloaded.
// In the case of the timestamp, we'd default to our cached timestamp, and
// not have to redownload anything (usually)
func TestUpdateNonRootRemoteCorruptedCannotUseLocalCache(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	for _, role := range data.BaseRoles {
		switch role {
		case data.CanonicalRootRole:
			break
		default:
			for _, testData := range waysToMessUpServer {
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					serverHasNewData: true,
					localCache:       true,
					role:             role,
				}, testData, true)
			}
		}
	}

	for _, role := range delegationsWithNonEmptyMetadata {
		for _, testData := range waysToMessUpServerBadMeta {
			testUpdateRemoteCorruptValidChecksum(t, updateOpts{
				serverHasNewData: true,
				localCache:       true,
				role:             role,
			}, testData, true)
		}

		for _, testData := range append(waysToMessUpServerBadSigs, wayToMessUpServerBadExpiry) {
			testUpdateRemoteCorruptValidChecksum(t, updateOpts{
				serverHasNewData: true,
				localCache:       true,
				role:             role,
				checkRepo:        checkBadDelegationRoleSkipped(t, role.String()),
			}, testData, false)
		}
	}

	for role, expectations := range waysToMessUpServerNonRootPerRole(t) {
		for _, testData := range expectations {
			roleName := data.RoleName(role)
			switch roleName {
			case data.CanonicalSnapshotRole:
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					serverHasNewData: true,
					localCache:       true,
					role:             roleName,
				}, testData, true)
			case data.CanonicalTargetsRole:
				// if there are no delegation target roles, we're fine, we just don't
				// download them
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					serverHasNewData: true,
					localCache:       true,
					role:             roleName,
				}, testData, false)
			case data.CanonicalTimestampRole:
				// we only default to the previous cached version of the timestamp if
				// there is a network/storage error, so swizzling will fail the update
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					serverHasNewData: true,
					localCache:       true,
					role:             roleName,
				}, testData, true)
			case data.RoleName("targets/a"):
				testUpdateRemoteCorruptValidChecksum(t, updateOpts{
					serverHasNewData: true,
					localCache:       true,
					role:             roleName,
				}, testData, true)
			}
		}
	}
}

func testUpdateRemoteCorruptValidChecksum(t *testing.T, opts updateOpts, expt swizzleExpectations, shouldErr bool) {
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	if opts.localCache {
		err := repo.Update(false)
		require.NoError(t, err)
	}

	if opts.serverHasNewData {
		bumpVersions(t, serverSwizzler, 1)
	}

	msg := fmt.Sprintf("swizzling %s to return: %v (forWrite: %v)", opts.role, expt.desc, opts.forWrite)

	require.NoError(t, expt.swizzle(serverSwizzler, opts.role),
		"failed %s", msg)

	// update the snapshot and timestamp hashes to make sure it's not an involuntary checksum failure
	// unless we want the server to not actually have any new data
	if !opts.localCache || opts.serverHasNewData {
		// we don't want to sign if we are trying to swizzle one of these roles to
		// have a different signature - updating hashes would be pointless (because
		// nothing else has changed) and would just overwrite the signature.
		isSignatureSwizzle := expt.desc == "invalid signatures" || expt.desc == "meta signed by wrong key"
		// just try to do these - if they fail (probably because they've been swizzled), that's fine
		if opts.role != data.CanonicalSnapshotRole || !isSignatureSwizzle {
			// if we are purposely editing out some snapshot metadata, don't re-generate
			if !strings.HasPrefix(expt.desc, "snapshot missing") {
				serverSwizzler.UpdateSnapshotHashes()
			}
		}
		if opts.role != data.CanonicalTimestampRole || !isSignatureSwizzle {
			// if we are purposely editing out some timestamp metadata, don't re-generate
			if !strings.HasPrefix(expt.desc, "timestamp missing") {
				serverSwizzler.UpdateTimestampHash()
			}
		}
	}
	err := repo.Update(opts.forWrite)
	checkErrors(t, err, shouldErr, expt.expectErrs, msg)

	if opts.checkRepo != nil {
		opts.checkRepo(repo, serverSwizzler)
	}
}

func checkErrors(t *testing.T, err error, shouldErr bool, expectedErrs []interface{}, msg string) {
	if shouldErr {
		require.Error(t, err, "expected failure updating when %s", msg)

		errType := reflect.TypeOf(err)
		isExpectedType := false
		var expectedTypes []string
		for _, expectErr := range expectedErrs {
			expectedType := reflect.TypeOf(expectErr)
			isExpectedType = isExpectedType || reflect.DeepEqual(errType, expectedType)
			expectedTypes = append(expectedTypes, expectedType.String())
		}
		require.True(t, isExpectedType, "expected one of %v when %s: got %s",
			expectedTypes, msg, errType)

	} else {
		require.NoError(t, err, "expected no failure updating when %s", msg)
	}
}

// If the local root is corrupt, and the remote root is corrupt, we should fail
// to update.  Note - this one is really slow.
func TestUpdateLocalAndRemoteRootCorrupt(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	for _, localExpt := range waysToMessUpLocalMetadata {
		for _, serverExpt := range waysToMessUpServer {
			testUpdateLocalAndRemoteRootCorrupt(t, true, localExpt, serverExpt)
			testUpdateLocalAndRemoteRootCorrupt(t, false, localExpt, serverExpt)
		}
	}
}

func testUpdateLocalAndRemoteRootCorrupt(t *testing.T, forWrite bool, localExpt, serverExpt swizzleExpectations) {
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	// get local cache
	err := repo.Update(false)
	require.NoError(t, err)
	repoSwizzler := &testutils.MetadataSwizzler{
		Gun:           serverSwizzler.Gun,
		MetadataCache: repo.cache,
		CryptoService: serverSwizzler.CryptoService,
		Roles:         serverSwizzler.Roles,
	}

	bumpVersions(t, serverSwizzler, 1)

	require.NoError(t, localExpt.swizzle(repoSwizzler, data.CanonicalRootRole),
		"failed to swizzle local root to %s", localExpt.desc)
	require.NoError(t, serverExpt.swizzle(serverSwizzler, data.CanonicalRootRole),
		"failed to swizzle remote root to %s", serverExpt.desc)

	// update the hashes on both
	require.NoError(t, serverSwizzler.UpdateSnapshotHashes())
	require.NoError(t, serverSwizzler.UpdateTimestampHash())

	msg := fmt.Sprintf("swizzling root locally to return <%v> and remotely to return: <%v> (forWrite: %v)",
		localExpt.desc, serverExpt.desc, forWrite)

	err = repo.Update(forWrite)
	require.Error(t, err, "expected failure updating when %s", msg)

	expectedErrs := serverExpt.expectErrs
	// If the local root is corrupt or invalid, we won't even try to update and
	// will fail with the local metadata error.  Missing or expired metadata is ok.
	if localExpt.desc != "missing metadata" && localExpt.desc != "expired metadata" {
		expectedErrs = localExpt.expectErrs
	}

	errType := reflect.TypeOf(err)
	isExpectedType := false
	var expectedTypes []string

	for _, expectErr := range expectedErrs {
		expectedType := reflect.TypeOf(expectErr)
		isExpectedType = isExpectedType || errType == expectedType
		expectedTypes = append(expectedTypes, expectedType.String())
	}
	require.True(t, isExpectedType, "expected one of %v when %s: got %s",
		expectedTypes, msg, errType)
}

// Update when we have a local cache.  This differs from
// TestUpdateNonRootRemoteCorruptedCannotUseLocalCache in that
// in this case, the ONLY change upstream is that a key is rotated.
// Therefore the only metadata that needs to change is the root
// (or the targets file), the snapshot, and the timestamp.  All other data has
// the same checksum as the data already cached (whereas in
// TestUpdateNonRootRemoteCorruptedCannotUseLocalCache all the metadata has their
// versions bumped and are re-signed).  The update should still fail, because
// the local cache no longer matches.
func TestUpdateRemoteKeyRotated(t *testing.T) {
	for _, role := range append(data.BaseRoles, delegationsWithNonEmptyMetadata...) {
		testUpdateRemoteKeyRotated(t, role)
	}
}

func testUpdateRemoteKeyRotated(t *testing.T, role data.RoleName) {
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	// get local cache
	err := repo.Update(false)
	require.NoError(t, err)

	cs := signed.NewEd25519()
	pubKey, err := cs.Create(role, repo.gun, data.ED25519Key)
	require.NoError(t, err)

	// bump the version
	bumpRole := role.Parent()
	if !data.IsDelegation(role) {
		bumpRole = data.CanonicalRootRole
	}
	require.NoError(t, serverSwizzler.OffsetMetadataVersion(bumpRole, 1),
		"failed to swizzle remote %s to bump version", bumpRole)
	// now change the key
	require.NoError(t, serverSwizzler.RotateKey(role, pubKey),
		"failed to swizzle remote %s to rotate key", role)

	// update the hashes on both snapshot and timestamp
	require.NoError(t, serverSwizzler.UpdateSnapshotHashes())
	require.NoError(t, serverSwizzler.UpdateTimestampHash())

	msg := fmt.Sprintf("swizzling %s remotely to rotate key (forWrite: false)", role)

	err = repo.Update(false)
	// invalid signatures are ok - the delegation is just skipped
	if data.IsDelegation(role) {
		require.NoError(t, err)
		checkBadDelegationRoleSkipped(t, role.String())(repo, serverSwizzler)
		return
	}
	require.Error(t, err, "expected failure updating when %s", msg)
	switch role {
	case data.CanonicalRootRole:
		require.IsType(t, &trustpinning.ErrValidationFail{}, err,
			"expected trustpinning.ErrValidationFail when %s: got %s",
			msg, reflect.TypeOf(err))
	default:
		require.IsType(t, signed.ErrRoleThreshold{}, err,
			"expected ErrRoleThreshold when %s: got %s",
			msg, reflect.TypeOf(err))
	}
}

// Helper function that takes a signedRoot, and signs it with the provided keys and only these keys.
// Then serializes this to bytes and updates the swizzler with it, and updates the snapshot and
// timestamp too so that the update won't fail due to a checksum mismatch.
func signSerializeAndUpdateRoot(t *testing.T, signedRoot data.SignedRoot,
	serverSwizzler *testutils.MetadataSwizzler, keys []data.PublicKey) {

	signedObj, err := signedRoot.ToSigned()
	require.NoError(t, err)

	// sign with the provided keys, and require all the keys have signed
	require.NoError(t, signed.Sign(serverSwizzler.CryptoService, signedObj, keys, len(keys), nil))
	rootBytes, err := json.Marshal(signedObj)
	require.NoError(t, err)
	require.NoError(t, serverSwizzler.MetadataCache.Set(data.CanonicalRootRole.String(), rootBytes))

	// update the hashes on both snapshot and timestamp
	require.NoError(t, serverSwizzler.UpdateSnapshotHashes())
	require.NoError(t, serverSwizzler.UpdateTimestampHash())
}

func requireRootSignatures(t *testing.T, serverSwizzler *testutils.MetadataSwizzler, num int) {
	updatedRootBytes, _ := serverSwizzler.MetadataCache.GetSized(data.CanonicalRootRole.String(), -1)
	updatedRoot := &data.SignedRoot{}
	require.NoError(t, json.Unmarshal(updatedRootBytes, updatedRoot))
	require.EqualValues(t, len(updatedRoot.Signatures), num)
}

// A valid root rotation only cares about the immediately previous old root keys,
// whether or not there are old root roles, and cares that the role is satisfied
// (for instance if the old role has 2 keys, either of which can sign, then it
// doesn't matter which key signs the rotation)
func TestValidateRootRotationWithOldRole(t *testing.T) {
	// start with a repo with a root with 2 keys, optionally signing 1
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	// --- setup so that the root starts with a role with 3 keys, and threshold of 2
	// --- signed by the first two keys (also, the original role which has 1 original
	// --- key is saved, but doesn't matter at all for rotation if we're already on
	// --- the root metadata with the 3 keys)

	rootBytes, err := serverSwizzler.MetadataCache.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	signedRoot := data.SignedRoot{}
	require.NoError(t, json.Unmarshal(rootBytes, &signedRoot))

	// save the old role to prove that it is not needed for client updates
	oldVersion := data.RoleName(fmt.Sprintf("%v.%v", data.CanonicalRootRole, signedRoot.Signed.Version))
	signedRoot.Signed.Roles[oldVersion] = &data.RootRole{
		Threshold: 1,
		KeyIDs:    signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs,
	}

	threeKeys := make([]data.PublicKey, 3)
	keyIDs := make([]string, len(threeKeys))
	for i := 0; i < len(threeKeys); i++ {
		threeKeys[i], err = testutils.CreateKey(
			serverSwizzler.CryptoService, "docker.com/notary", data.CanonicalRootRole, data.ECDSAKey)
		require.NoError(t, err)
		keyIDs[i] = threeKeys[i].ID()
		signedRoot.Signed.Keys[keyIDs[i]] = threeKeys[i]
	}
	signedRoot.Signed.Version++
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = keyIDs
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 2
	// sign with the first two keys only
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, threeKeys[:2])

	// Load this root for the first time with 3 keys
	require.NoError(t, repo.Update(false))

	// --- First root rotation: replace the first key with a different key and change the
	// --- threshold back to 1

	replacementKey, err := testutils.CreateKey(
		serverSwizzler.CryptoService, "docker.com/notary", data.CanonicalRootRole, data.ECDSAKey)
	require.NoError(t, err)
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[replacementKey.ID()] = replacementKey
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = append(keyIDs[1:], replacementKey.ID())
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 1

	// --- If the current role is satisfied but the previous one is not, root rotation
	// --- will fail.  (signing with just the second key will not satisfy the first role,
	// --- because that one has a threshold of 2, although it will satisfy the new role)
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, threeKeys[1:2])
	require.Error(t, repo.Update(false))

	// --- If both the current and previous roles are satisfied, then the root rotation
	// --- will succeed (signing with the second and third keys will satisfies both)
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, threeKeys[1:])
	require.NoError(t, repo.Update(false))

	// --- Older roles do not have to be satisfied in order to validate if the update
	// --- does not involve a root rotation (replacing the snapshot key is not a root
	// --- rotation, and signing with just the replacement key will satisfy ONLY the
	// --- latest root role)
	signedRoot.Signed.Version++
	snapKey, err := testutils.CreateKey(
		serverSwizzler.CryptoService, "docker.com/notary", data.CanonicalSnapshotRole, data.ECDSAKey)
	require.NoError(t, err)
	signedRoot.Signed.Keys[snapKey.ID()] = snapKey
	signedRoot.Signed.Roles[data.CanonicalSnapshotRole].KeyIDs = []string{snapKey.ID()}

	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{replacementKey})
	require.NoError(t, repo.Update(false))

	// --- Second root rotation: if only the previous role is satisfied but not the new role,
	// --- then the root rotation will fail (if we rotate to the only valid signing key being
	// --- threeKeys[0], signing with just the replacement key will satisfy ONLY the
	// --- previous root role)
	signedRoot.Signed.Version++
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[0]}

	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{replacementKey})
	require.Error(t, repo.Update(false))

	// again, signing with both will succeed
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{replacementKey, threeKeys[0]})
	require.NoError(t, repo.Update(false))
}

// A valid root role is signed by the current root role keys and the previous root role keys
func TestRootRoleInvariant(t *testing.T) {
	// start with a repo
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	// --- setup so that the root starts with a role with 1 keys, and threshold of 1
	rootBytes, err := serverSwizzler.MetadataCache.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	signedRoot := data.SignedRoot{}
	require.NoError(t, json.Unmarshal(rootBytes, &signedRoot))

	// save the old role to prove that it is not needed for client updates
	oldVersion := data.RoleName(fmt.Sprintf("%v.%v", data.CanonicalRootRole.String(), signedRoot.Signed.Version))
	signedRoot.Signed.Roles[oldVersion] = &data.RootRole{
		Threshold: 1,
		KeyIDs:    signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs,
	}

	threeKeys := make([]data.PublicKey, 3)
	keyIDs := make([]string, len(threeKeys))
	for i := 0; i < len(threeKeys); i++ {
		threeKeys[i], err = testutils.CreateKey(
			serverSwizzler.CryptoService, "docker.com/notary", data.CanonicalRootRole, data.ECDSAKey)
		require.NoError(t, err)
		keyIDs[i] = threeKeys[i].ID()
	}
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[keyIDs[0]] = threeKeys[0]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[0]}
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 1
	// sign with the first key only
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{threeKeys[0]})

	// Load this root for the first time with 1 key
	require.NoError(t, repo.Update(false))

	// --- First root rotation: replace the first key with a different key
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[keyIDs[1]] = threeKeys[1]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[1]}

	// --- If the current role is satisfied but the previous one is not, root rotation
	// --- will fail.  Signing with just the second key will not satisfy the first role.
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{threeKeys[1]})
	require.Error(t, repo.Update(false))
	requireRootSignatures(t, serverSwizzler, 1)

	// --- If both the current and previous roles are satisfied, then the root rotation
	// --- will succeed (signing with the first and second keys will satisfy both)
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, threeKeys[:2])
	require.NoError(t, repo.Update(false))
	requireRootSignatures(t, serverSwizzler, 2)

	// --- Second root rotation: replace the second key with a third
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[keyIDs[2]] = threeKeys[2]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[2]}

	// --- If the current role is satisfied but the previous one is not, root rotation
	// --- will fail.  Signing with just the second key will not satisfy the first role.
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{threeKeys[2]})
	require.Error(t, repo.Update(false))
	requireRootSignatures(t, serverSwizzler, 1)

	// --- If both the current and previous roles are satisfied, then the root rotation
	// --- will succeed (signing with the second and third keys will satisfy both)
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, threeKeys[1:])
	require.NoError(t, repo.Update(false))
	requireRootSignatures(t, serverSwizzler, 2)

	// -- If signed with all previous roles, update will succeed
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, threeKeys)
	require.NoError(t, repo.Update(false))
	requireRootSignatures(t, serverSwizzler, 3)
}

// All intermediate roots must be signed by the previous root role
func TestBadIntermediateTransitions(t *testing.T) {
	// start with a repo
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	// --- setup so that the root starts with a role with 1 keys, and threshold of 1
	rootBytes, err := serverSwizzler.MetadataCache.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	signedRoot := data.SignedRoot{}
	require.NoError(t, json.Unmarshal(rootBytes, &signedRoot))

	// generate keys for testing
	threeKeys := make([]data.PublicKey, 3)
	keyIDs := make([]string, len(threeKeys))
	for i := 0; i < len(threeKeys); i++ {
		threeKeys[i], err = testutils.CreateKey(
			serverSwizzler.CryptoService, "docker.com/notary", data.CanonicalRootRole, data.ECDSAKey)
		require.NoError(t, err)
		keyIDs[i] = threeKeys[i].ID()
	}

	// increment the root version and sign with the first key only
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[keyIDs[0]] = threeKeys[0]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[0]}
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 1
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{threeKeys[0]})

	require.NoError(t, repo.Update(false))

	// increment the root version and sign with the second key only
	signedRoot.Signed.Version++
	delete(signedRoot.Signed.Keys, keyIDs[0])
	signedRoot.Signed.Keys[keyIDs[1]] = threeKeys[1]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[1]}
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 1
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{threeKeys[1]})

	// increment the root version and sign with all three keys
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[keyIDs[0]] = threeKeys[0]
	signedRoot.Signed.Keys[keyIDs[1]] = threeKeys[1]
	signedRoot.Signed.Keys[keyIDs[2]] = threeKeys[2]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[0], keyIDs[1], keyIDs[2]}
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 1
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{threeKeys[1]})
	requireRootSignatures(t, serverSwizzler, 1)

	// Update fails because version 1 -> 2 is invalid.
	require.Error(t, repo.Update(false))
}

// All intermediate roots must be signed by the previous root role
func TestExpiredIntermediateTransitions(t *testing.T) {
	// start with a repo
	_, serverSwizzler := newServerSwizzler(t)
	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	// --- setup so that the root starts with a role with 1 keys, and threshold of 1
	rootBytes, err := serverSwizzler.MetadataCache.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
	require.NoError(t, err)
	signedRoot := data.SignedRoot{}
	require.NoError(t, json.Unmarshal(rootBytes, &signedRoot))

	// generate keys for testing
	threeKeys := make([]data.PublicKey, 3)
	keyIDs := make([]string, len(threeKeys))
	for i := 0; i < len(threeKeys); i++ {
		threeKeys[i], err = testutils.CreateKey(
			serverSwizzler.CryptoService, "docker.com/notary", data.CanonicalRootRole, data.ECDSAKey)
		require.NoError(t, err)
		keyIDs[i] = threeKeys[i].ID()
	}

	// increment the root version and sign with the first key only
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[keyIDs[0]] = threeKeys[0]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[0]}
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 1
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{threeKeys[0]})

	require.NoError(t, repo.Update(false))

	// increment the root version and sign with the first and second keys, but set metadata to be expired.
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[keyIDs[1]] = threeKeys[1]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[0], keyIDs[1]}
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 1
	signedRoot.Signed.Expires = time.Now().AddDate(0, -1, 0)
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, []data.PublicKey{threeKeys[0], threeKeys[1]})

	// increment the root version and sign with all three keys
	signedRoot.Signed.Version++
	signedRoot.Signed.Keys[keyIDs[2]] = threeKeys[2]
	signedRoot.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{keyIDs[0], keyIDs[1], keyIDs[2]}
	signedRoot.Signed.Roles[data.CanonicalRootRole].Threshold = 1
	signedRoot.Signed.Expires = time.Now().AddDate(0, 1, 0)
	signSerializeAndUpdateRoot(t, signedRoot, serverSwizzler, threeKeys[:3])
	requireRootSignatures(t, serverSwizzler, 3)

	// Update succeeds despite version 2 being expired.
	require.NoError(t, repo.Update(false))
}

// TestDownloadTargetsLarge: Check that we can download very large targets metadata files,
// which may be caused by adding a large number of targets.
// This test is slow, so it will not run in short mode.
func TestDownloadTargetsLarge(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}
	numTargets := 75000

	tufRepo, cs, err := testutils.EmptyRepo("docker.com/notary")
	require.NoError(t, err)
	fMeta, err := data.NewFileMeta(bytes.NewBuffer([]byte("hello")), notary.SHA256)
	require.NoError(t, err)

	// Add a ton of target files to the targets role to make this targets metadata huge
	// 75,000 targets results in > 5MB (~6.5MB on recent runs)
	for i := 0; i < numTargets; i++ {
		_, err = tufRepo.AddTargets(data.CanonicalTargetsRole, data.Files{strconv.Itoa(i): fMeta})
		require.NoError(t, err)
	}

	serverMeta, err := testutils.SignAndSerialize(tufRepo)
	require.NoError(t, err)

	serverSwizzler := testutils.NewMetadataSwizzler("docker.com/notary", serverMeta, cs)
	require.NoError(t, err)

	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	notaryRepo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(notaryRepo.baseDir)

	tgts, err := notaryRepo.ListTargets()
	require.NoError(t, err)
	require.Len(t, tgts, numTargets)
}

func TestDownloadTargetsDeep(t *testing.T) {
	delegations := []data.RoleName{
		// left subtree
		"targets/level1",
		"targets/level1/a",
		"targets/level1/a/i",
		"targets/level1/a/i/0",
		"targets/level1/a/ii",
		"targets/level1/a/iii",
		// right subtree
		"targets/level2",
		"targets/level2/b",
		"targets/level2/b/i",
		"targets/level2/b/i/0",
		"targets/level2/b/i/1",
	}

	serverMeta, cs, err := testutils.NewRepoMetadata("docker.com/notary", delegations...)
	require.NoError(t, err)

	serverSwizzler := testutils.NewMetadataSwizzler("docker.com/notary", serverMeta, cs)
	require.NoError(t, err)

	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)

	roles, err := repo.ListRoles()
	require.NoError(t, err)

	// 4 base roles + all the delegation roles
	require.Len(t, roles, len(delegations)+4)
	// downloaded all of the delegations except for the two lowest level ones, which
	// should have no metadata because there are no targets
	for _, r := range roles {
		if _, ok := serverMeta[r.Name]; ok {
			require.Len(t, r.Signatures, 1, r.Name, "should have 1 signature because there was metadata")
		} else {
			require.Len(t, r.Signatures, 0, r.Name,
				"should have no signatures because no metadata should have been downloaded")
		}
	}
}

// TestDownloadSnapshotLargeDelegationsMany: Check that we can download very large
// snapshot metadata files, which may be caused by adding a large number of delegations,
// as well as a large number of delegations.
// This test is very slow, so it will not run in short mode.
func TestDownloadSnapshotLargeDelegationsMany(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	numSnapsnotMeta := 75000

	fMeta, err := data.NewFileMeta(bytes.NewBuffer([]byte("hello")), notary.SHA256)
	require.NoError(t, err)

	tufRepo, cs, err := testutils.EmptyRepo("docker.com/notary")
	require.NoError(t, err)

	delgKey, err := cs.Create("docker.com/notary", "delegation", data.ECDSAKey)
	require.NoError(t, err)

	// Add a ton of empty delegation roles to targets to make snapshot data huge
	// This can also be done by adding legitimate delegations but it will be much slower
	// 75,000 delegation roles results in > 5MB (~7.3MB on recent runs)
	for i := 0; i < numSnapsnotMeta; i++ {
		roleName := data.RoleName(fmt.Sprintf("targets/%d", i))
		// for a tiny fraction of the delegations,  make sure role is added, so the meta is downloaded
		if i%1000 == 0 {
			require.NoError(t, tufRepo.UpdateDelegationKeys(roleName, data.KeyList{delgKey}, nil, 1))
			_, err := tufRepo.InitTargets(roleName) // make sure metadata is created
			require.NoError(t, err)
		} else {
			tufRepo.Snapshot.AddMeta(roleName, fMeta)
		}
	}

	serverMeta, err := testutils.SignAndSerialize(tufRepo)
	require.NoError(t, err)

	serverSwizzler := testutils.NewMetadataSwizzler("docker.com/notary", serverMeta, cs)
	require.NoError(t, err)

	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	notaryRepo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(notaryRepo.baseDir)

	roles, err := notaryRepo.ListRoles()
	require.NoError(t, err)

	// all the roles have server meta this time
	require.Len(t, roles, len(serverMeta))

	// downloaded all of the delegations except for the two lowest level ones, which
	// should have no metadata because there are no targets
	for _, r := range roles {
		require.Len(t, r.Signatures, 1, r.Name, "should have 1 signature because there was metadata")
	}

	// the snapshot downloaded has numSnapsnotMeta items + one for root and one for targets
	require.Len(t, notaryRepo.tufRepo.Snapshot.Signed.Meta, numSnapsnotMeta+2)
}

// If we have a root on disk, use it as the source of trust pinning rather than the trust pinning
// config
func TestRootOnDiskTrustPinning(t *testing.T) {
	meta, serverSwizzler := newServerSwizzler(t)

	ts := readOnlyServer(t, serverSwizzler.MetadataCache, http.StatusNotFound, "docker.com/notary")
	defer ts.Close()

	restrictiveTrustPinning := trustpinning.TrustPinConfig{DisableTOFU: true}

	// for sanity, ensure that without a root on disk, we can't download a new root
	repo := newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)
	repo.trustPinning = restrictiveTrustPinning

	err := repo.Update(false)
	require.Error(t, err)
	require.IsType(t, &trustpinning.ErrValidationFail{}, err)

	// show that if we have a root on disk, we can update
	repo = newBlankRepo(t, ts.URL)
	defer os.RemoveAll(repo.baseDir)
	repo.trustPinning = restrictiveTrustPinning
	// put root on disk
	require.NoError(t, repo.cache.Set(data.CanonicalRootRole.String(), meta[data.CanonicalRootRole]))

	require.NoError(t, repo.Update(false))
}
