// The client can read and operate on older repository formats

package client

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/notary"
	"github.com/docker/notary/passphrase"
	store "github.com/docker/notary/storage"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

// Once a fixture is read in, ensure that it's valid by making sure the expiry
// times of all the metadata and certificates is > 10 years ahead
func requireValidFixture(t *testing.T, notaryRepo *NotaryRepository) {
	tenYearsInFuture := time.Now().AddDate(10, 0, 0)
	require.True(t, notaryRepo.tufRepo.Root.Signed.Expires.After(tenYearsInFuture))
	require.True(t, notaryRepo.tufRepo.Snapshot.Signed.Expires.After(tenYearsInFuture))
	require.True(t, notaryRepo.tufRepo.Timestamp.Signed.Expires.After(tenYearsInFuture))
	for _, targetObj := range notaryRepo.tufRepo.Targets {
		require.True(t, targetObj.Signed.Expires.After(tenYearsInFuture))
	}
}

// recursively copies the contents of one directory into another - ignores
// symlinks
func recursiveCopy(sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		targetFP := filepath.Join(targetDir, strings.TrimPrefix(fp, sourceDir+"/"))

		if fi.IsDir() {
			return os.MkdirAll(targetFP, fi.Mode())
		}

		// Ignore symlinks
		if fi.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil
		}

		// copy the file
		in, err := os.Open(fp)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(targetFP)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}
		return nil
	})
}

func Test0Dot1Migration(t *testing.T) {
	// make a temporary directory and copy the fixture into it, since updating
	// and publishing will modify the files
	tmpDir, err := ioutil.TempDir("", "notary-backwards-compat-test")
	defer os.RemoveAll(tmpDir)
	require.NoError(t, err)
	require.NoError(t, recursiveCopy("../fixtures/compatibility/notary0.1", tmpDir))

	var gun data.GUN = "docker.com/notary0.1/samplerepo"
	passwd := "randompass"

	ts := fullTestServer(t)
	defer ts.Close()

	_, err = NewFileCachedNotaryRepository(tmpDir, gun, ts.URL, http.DefaultTransport,
		passphrase.ConstantRetriever(passwd), trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	// check that root_keys and tuf_keys are gone and that all corect keys are present and have the correct headers
	files, _ := ioutil.ReadDir(filepath.Join(tmpDir, notary.PrivDir))
	require.Equal(t, files[0].Name(), "7fc757801b9bab4ec9e35bfe7a6b61668ff6f4c81b5632af19e6c728ab799599.key")
	targKey, err := os.OpenFile(filepath.Join(tmpDir, notary.PrivDir, "7fc757801b9bab4ec9e35bfe7a6b61668ff6f4c81b5632af19e6c728ab799599.key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer targKey.Close()
	targBytes, _ := ioutil.ReadAll(targKey)
	targString := string(targBytes)
	require.Contains(t, targString, "gun: docker.com/notary0.1/samplerepo")
	require.Contains(t, targString, "role: targets")
	require.Equal(t, files[1].Name(), "a55ccf652b0be4b6c4d356cbb02d9ea432bb84a2571665be3df7c7396af8e8b8.key")
	snapKey, err := os.OpenFile(filepath.Join(tmpDir, notary.PrivDir, "a55ccf652b0be4b6c4d356cbb02d9ea432bb84a2571665be3df7c7396af8e8b8.key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer snapKey.Close()
	snapBytes, _ := ioutil.ReadAll(snapKey)
	snapString := string(snapBytes)
	require.Contains(t, snapString, "gun: docker.com/notary0.1/samplerepo")
	require.Contains(t, snapString, "role: snapshot")
	require.Equal(t, files[2].Name(), "d0c623c8e70c70d42a8a8125c44a8598588b3f6e31d5c21a83cbc338dfde8a68.key")
	rootKey, err := os.OpenFile(filepath.Join(tmpDir, notary.PrivDir, "d0c623c8e70c70d42a8a8125c44a8598588b3f6e31d5c21a83cbc338dfde8a68.key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer rootKey.Close()
	rootBytes, _ := ioutil.ReadAll(rootKey)
	rootString := string(rootBytes)
	require.Contains(t, rootString, "role: root")
	require.NotContains(t, rootString, "gun")
	require.Len(t, files, 3)
}

func Test0Dot3Migration(t *testing.T) {
	// make a temporary directory and copy the fixture into it, since updating
	// and publishing will modify the files
	tmpDir, err := ioutil.TempDir("", "notary-backwards-compat-test")
	defer os.RemoveAll(tmpDir)
	require.NoError(t, err)
	require.NoError(t, recursiveCopy("../fixtures/compatibility/notary0.3", tmpDir))

	var gun data.GUN = "docker.com/notary0.3/samplerepo"
	passwd := "randompass"

	ts := fullTestServer(t)
	defer ts.Close()

	_, err = NewFileCachedNotaryRepository(tmpDir, gun, ts.URL, http.DefaultTransport,
		passphrase.ConstantRetriever(passwd), trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	// check that root_keys and tuf_keys are gone and that all corect keys are present and have the correct headers
	files, _ := ioutil.ReadDir(filepath.Join(tmpDir, notary.PrivDir))
	require.Equal(t, files[0].Name(), "041b64dab281324ef2b62fd2d04f4758269e120ff063b7bc78709272821a0a02.key")
	targKey, err := os.OpenFile(filepath.Join(tmpDir, notary.PrivDir, "041b64dab281324ef2b62fd2d04f4758269e120ff063b7bc78709272821a0a02.key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer targKey.Close()
	targBytes, _ := ioutil.ReadAll(targKey)
	targString := string(targBytes)
	require.Contains(t, targString, "gun: docker.com/notary0.3/tst")
	require.Contains(t, targString, "role: targets")
	require.Equal(t, files[1].Name(), "85559599cf3cf681ff193f432a7ca6d128182bd1cfa8ede2c70761deac8bc2dc.key")
	snapKey, err := os.OpenFile(filepath.Join(tmpDir, notary.PrivDir, "85559599cf3cf681ff193f432a7ca6d128182bd1cfa8ede2c70761deac8bc2dc.key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer snapKey.Close()
	snapBytes, _ := ioutil.ReadAll(snapKey)
	snapString := string(snapBytes)
	require.Contains(t, snapString, "gun: docker.com/notary0.3/tst")
	require.Contains(t, snapString, "role: snapshot")
	require.Equal(t, files[2].Name(), "f4eaf871a74aa3b3a0ff95cef2455a1e4d461639f5625418e76756fc5c948690.key")
	rootKey, err := os.OpenFile(filepath.Join(tmpDir, notary.PrivDir, "f4eaf871a74aa3b3a0ff95cef2455a1e4d461639f5625418e76756fc5c948690.key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer rootKey.Close()
	rootBytes, _ := ioutil.ReadAll(rootKey)
	rootString := string(rootBytes)
	require.Contains(t, rootString, "role: root")
	require.NotContains(t, rootString, "gun")
	require.Equal(t, files[3].Name(), "fa842f66cac2dc898677a8660789dcff0e3b0b93b73f8952491f6493199936d3.key")
	delKey, err := os.OpenFile(filepath.Join(tmpDir, notary.PrivDir, "fa842f66cac2dc898677a8660789dcff0e3b0b93b73f8952491f6493199936d3.key"), os.O_RDONLY, notary.PrivExecPerms)
	require.NoError(t, err)
	defer delKey.Close()
	delBytes, _ := ioutil.ReadAll(delKey)
	delString := string(delBytes)
	require.Contains(t, delString, "role: targets/releases")
	require.NotContains(t, delString, "gun")
	require.Len(t, files, 4)
}

// We can read and publish from notary0.1 repos
func Test0Dot1RepoFormat(t *testing.T) {
	// make a temporary directory and copy the fixture into it, since updating
	// and publishing will modify the files
	tmpDir, err := ioutil.TempDir("", "notary-backwards-compat-test")
	defer os.RemoveAll(tmpDir)
	require.NoError(t, err)
	require.NoError(t, recursiveCopy("../fixtures/compatibility/notary0.1", tmpDir))

	var gun data.GUN = "docker.com/notary0.1/samplerepo"
	passwd := "randompass"

	ts := fullTestServer(t)
	defer ts.Close()

	repo, err := NewFileCachedNotaryRepository(tmpDir, gun, ts.URL, http.DefaultTransport,
		passphrase.ConstantRetriever(passwd), trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	// targets should have 1 target, and it should be readable offline
	targets, err := repo.ListTargets()
	require.NoError(t, err)
	require.Len(t, targets, 1)
	require.Equal(t, "LICENSE", targets[0].Name)

	// ok, now that everything has been loaded, verify that the fixture is valid
	requireValidFixture(t, repo)

	// delete the timestamp metadata, since the server will ignore the uploaded
	// one and try to create a new one from scratch, which will be the wrong version
	require.NoError(t, repo.cache.Remove(data.CanonicalTimestampRole.String()))

	// rotate the timestamp key, since the server doesn't have that one
	err = repo.RotateKey(data.CanonicalTimestampRole, true, nil)
	require.NoError(t, err)

	require.NoError(t, repo.Publish())

	targets, err = repo.ListTargets()
	require.NoError(t, err)
	require.Len(t, targets, 2)

	// Also check that we can add/remove keys by rotating keys
	oldTargetsKeys := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)
	require.NoError(t, repo.RotateKey(data.CanonicalTargetsRole, false, nil))
	require.NoError(t, repo.Publish())
	newTargetsKeys := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)

	require.Len(t, oldTargetsKeys, 1)
	require.Len(t, newTargetsKeys, 1)
	require.NotEqual(t, oldTargetsKeys[0], newTargetsKeys[0])

	// rotate the snapshot key to the server and ensure that the server can re-generate the snapshot
	// and we can download the snapshot
	require.NoError(t, repo.RotateKey(data.CanonicalSnapshotRole, true, nil))
	require.NoError(t, repo.Publish())
	err = repo.Update(false)
	require.NoError(t, err)
}

// We can read and publish from notary0.3 repos
func Test0Dot3RepoFormat(t *testing.T) {
	// make a temporary directory and copy the fixture into it, since updating
	// and publishing will modify the files
	tmpDir, err := ioutil.TempDir("", "notary-backwards-compat-test")
	defer os.RemoveAll(tmpDir)
	require.NoError(t, err)
	require.NoError(t, recursiveCopy("../fixtures/compatibility/notary0.3", tmpDir))

	var gun data.GUN = "docker.com/notary0.3/tst"
	passwd := "password"

	ts := fullTestServer(t)
	defer ts.Close()

	repo, err := NewFileCachedNotaryRepository(tmpDir, gun, ts.URL, http.DefaultTransport,
		passphrase.ConstantRetriever(passwd), trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	// targets should have 1 target, and it should be readable offline
	targets, err := repo.ListTargets()
	require.NoError(t, err)
	require.Len(t, targets, 3)

	// ok, now that everything has been loaded, verify that the fixture is valid
	requireValidFixture(t, repo)

	// delete the timestamp metadata, since the server will ignore the uploaded
	// one and try to create a new one from scratch, which will be the wrong version
	require.NoError(t, repo.cache.Remove(data.CanonicalTimestampRole.String()))

	// rotate the timestamp key, since the server doesn't have that one
	err = repo.RotateKey(data.CanonicalTimestampRole, true, nil)
	require.NoError(t, err)

	require.NoError(t, repo.Publish())

	targets, err = repo.ListTargets()
	require.NoError(t, err)
	require.Len(t, targets, 5)
	// the changelist target/releases delegation will get published with the above publish
	delegations, err := repo.GetDelegationRoles()
	require.NoError(t, err)
	require.Len(t, delegations, 1)
	require.Equal(t, data.RoleName("targets/releases"), delegations[0].Name)

	// Also check that we can add/remove keys by rotating keys
	oldTargetsKeys := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)
	require.NoError(t, repo.RotateKey(data.CanonicalTargetsRole, false, nil))
	require.NoError(t, repo.Publish())
	newTargetsKeys := repo.CryptoService.ListKeys(data.CanonicalTargetsRole)

	require.Len(t, oldTargetsKeys, 1)
	require.Len(t, newTargetsKeys, 1)
	require.NotEqual(t, oldTargetsKeys[0], newTargetsKeys[0])

	// rotate the snapshot key to the server and ensure that the server can re-generate the snapshot
	// and we can download the snapshot
	require.NoError(t, repo.RotateKey(data.CanonicalSnapshotRole, true, nil))
	require.NoError(t, repo.Publish())
	err = repo.Update(false)
	require.NoError(t, err)
}

// Ensures that the current client can download metadata that is published from notary 0.1 repos
func TestDownloading0Dot1RepoFormat(t *testing.T) {
	var gun data.GUN = "docker.com/notary0.1/samplerepo"
	passwd := "randompass"

	metaCache, err := store.NewFileStore(
		filepath.Join("../fixtures/compatibility/notary0.1/tuf", filepath.FromSlash(gun.String()), "metadata"),
		"json")
	require.NoError(t, err)

	ts := readOnlyServer(t, metaCache, http.StatusNotFound, gun)
	defer ts.Close()

	repoDir, err := ioutil.TempDir("", "notary-backwards-compat-test")
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	repo, err := NewFileCachedNotaryRepository(repoDir, gun, ts.URL, http.DefaultTransport,
		passphrase.ConstantRetriever(passwd), trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	err = repo.Update(true)
	require.NoError(t, err, "error updating repo: %s", err)
}

// Ensures that the current client can download metadata that is published from notary 0.3 repos
func TestDownloading0Dot3RepoFormat(t *testing.T) {
	var gun data.GUN = "docker.com/notary0.3/tst"
	passwd := "randompass"

	metaCache, err := store.NewFileStore(
		filepath.Join("../fixtures/compatibility/notary0.3/tuf", filepath.FromSlash(gun.String()), "metadata"),
		"json")
	require.NoError(t, err)

	ts := readOnlyServer(t, metaCache, http.StatusNotFound, gun)
	defer ts.Close()

	repoDir, err := ioutil.TempDir("", "notary-backwards-compat-test")
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	repo, err := NewFileCachedNotaryRepository(repoDir, gun, ts.URL, http.DefaultTransport,
		passphrase.ConstantRetriever(passwd), trustpinning.TrustPinConfig{})
	require.NoError(t, err, "error creating repo: %s", err)

	err = repo.Update(true)
	require.NoError(t, err, "error updating repo: %s", err)
}
