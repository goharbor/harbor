package tuf

import (
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

var testGUN data.GUN = "gun"

func initRepo(t *testing.T, cryptoService signed.CryptoService) *Repo {
	rootKey, err := cryptoService.Create("root", testGUN, data.ED25519Key)
	require.NoError(t, err)
	return initRepoWithRoot(t, cryptoService, rootKey)
}

func initRepoWithRoot(t *testing.T, cryptoService signed.CryptoService, rootKey data.PublicKey) *Repo {
	targetsKey, err := cryptoService.Create("targets", testGUN, data.ED25519Key)
	require.NoError(t, err)
	snapshotKey, err := cryptoService.Create("snapshot", testGUN, data.ED25519Key)
	require.NoError(t, err)
	timestampKey, err := cryptoService.Create("timestamp", testGUN, data.ED25519Key)
	require.NoError(t, err)

	rootRole := data.NewBaseRole(
		data.CanonicalRootRole,
		1,
		rootKey,
	)
	targetsRole := data.NewBaseRole(
		data.CanonicalTargetsRole,
		1,
		targetsKey,
	)
	snapshotRole := data.NewBaseRole(
		data.CanonicalSnapshotRole,
		1,
		snapshotKey,
	)
	timestampRole := data.NewBaseRole(
		data.CanonicalTimestampRole,
		1,
		timestampKey,
	)

	repo := NewRepo(cryptoService)
	err = repo.InitRoot(rootRole, timestampRole, snapshotRole, targetsRole, false)
	require.NoError(t, err)
	_, err = repo.InitTargets(data.CanonicalTargetsRole)
	require.NoError(t, err)
	err = repo.InitSnapshot()
	require.NoError(t, err)
	err = repo.InitTimestamp()
	require.NoError(t, err)
	return repo
}

func TestInitSnapshotNoTargets(t *testing.T) {
	cs := signed.NewEd25519()
	repo := initRepo(t, cs)

	repo.Targets = make(map[data.RoleName]*data.SignedTargets)

	err := repo.InitSnapshot()
	require.Error(t, err)
	require.IsType(t, ErrNotLoaded{}, err)
}

func writeRepo(t *testing.T, dir string, repo *Repo) {
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)
	signedRoot, err := repo.SignRoot(data.DefaultExpires("root"), nil)
	require.NoError(t, err)
	rootJSON, _ := json.Marshal(signedRoot)
	ioutil.WriteFile(dir+"/root.json", rootJSON, 0755)

	for r := range repo.Targets {
		signedTargets, err := repo.SignTargets(r, data.DefaultExpires("targets"))
		require.NoError(t, err)
		targetsJSON, _ := json.Marshal(signedTargets)
		p := path.Join(dir, r.String()+".json")
		parentDir := filepath.Dir(p)
		os.MkdirAll(parentDir, 0755)
		ioutil.WriteFile(p, targetsJSON, 0755)
	}

	signedSnapshot, err := repo.SignSnapshot(data.DefaultExpires("snapshot"))
	require.NoError(t, err)
	snapshotJSON, _ := json.Marshal(signedSnapshot)
	ioutil.WriteFile(dir+"/snapshot.json", snapshotJSON, 0755)

	signedTimestamp, err := repo.SignTimestamp(data.DefaultExpires("timestamp"))
	require.NoError(t, err)
	timestampJSON, _ := json.Marshal(signedTimestamp)
	ioutil.WriteFile(dir+"/timestamp.json", timestampJSON, 0755)
}

func TestInitRepo(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)
	writeRepo(t, testDir, repo)
	// after signing a new repo, there are only 4 roles: the 4 base roles
	require.Len(t, repo.Root.Signed.Roles, 4)

	// can't use getBaseRole because it's not a valid real role
	_, err = repo.Root.BuildBaseRole("root.1")
	require.Error(t, err)
}

func TestUpdateDelegations(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{"test"}, []string{}, false)
	require.NoError(t, err)

	// no empty metadata is created for this role
	_, ok := repo.Targets["targets/test"]
	require.False(t, ok, "no empty targets file should be created for deepest delegation")

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 1)
	require.Len(t, r.Signed.Delegations.Keys, 1)
	keyIDs := r.Signed.Delegations.Roles[0].KeyIDs
	require.Len(t, keyIDs, 1)
	require.Equal(t, testKey.ID(), keyIDs[0])

	testDeepKey, err := ed25519.Create("targets/test/deep", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test/deep", []data.PublicKey{testDeepKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test/deep", []string{"test/deep"}, []string{}, false)
	require.NoError(t, err)

	// this metadata didn't exist before, but creating targets/test/deep created
	// the targets/test metadata
	r, ok = repo.Targets["targets/test"]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 1)
	require.Len(t, r.Signed.Delegations.Keys, 1)
	keyIDs = r.Signed.Delegations.Roles[0].KeyIDs
	require.Len(t, keyIDs, 1)
	require.Equal(t, testDeepKey.ID(), keyIDs[0])
	require.True(t, r.Dirty)

	// no empty delegation metadata is created for targets/test/deep
	_, ok = repo.Targets["targets/test/deep"]
	require.False(t, ok, "no empty targets file should be created for deepest delegation")
}

func TestPurgeDelegationsKeyFromTop(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	vetinari := data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "vetinari"))
	sybil := data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "sybil"))
	vimes := data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "vimes"))
	carrot := data.RoleName(path.Join(vimes.String(), "carrot"))
	targetsWild := data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "*"))

	// create 2 keys, we'll purge one of them
	testKey1, err := ed25519.Create(vetinari, testGUN, data.ED25519Key)
	require.NoError(t, err)
	testKey2, err := ed25519.Create(vetinari, testGUN, data.ED25519Key)
	require.NoError(t, err)

	// create some delegations
	err = repo.UpdateDelegationKeys(vetinari, []data.PublicKey{testKey1, testKey2}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths(vetinari, []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys(sybil, []data.PublicKey{testKey1}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths(sybil, []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys(vimes, []data.PublicKey{testKey2}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths(vimes, []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys(carrot, []data.PublicKey{testKey1}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths(carrot, []string{""}, []string{}, false)
	require.NoError(t, err)

	id1, err := utils.CanonicalKeyID(testKey1)
	require.NoError(t, err)
	err = repo.PurgeDelegationKeys(targetsWild, []string{id1})
	require.NoError(t, err)

	role, err := repo.GetDelegationRole(vetinari)
	require.NoError(t, err)
	require.Len(t, role.Keys, 1)
	_, ok := role.Keys[testKey2.ID()]
	require.True(t, ok)

	role, err = repo.GetDelegationRole(sybil)
	require.NoError(t, err)
	require.Len(t, role.Keys, 0)

	role, err = repo.GetDelegationRole(vimes)
	require.NoError(t, err)
	require.Len(t, role.Keys, 1)
	_, ok = role.Keys[testKey2.ID()]
	require.True(t, ok)

	role, err = repo.GetDelegationRole(carrot)
	require.NoError(t, err)
	require.Len(t, role.Keys, 0)

	// we know id1 was successfully purged, try purging again and make sure it doesn't error
	err = repo.PurgeDelegationKeys(targetsWild, []string{id1})
	require.NoError(t, err)
}

func TestPurgeDelegationsKeyFromDeep(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	vetinari := data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "vetinari"))
	sybil := data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "sybil"))
	vimes := data.RoleName(path.Join(data.CanonicalTargetsRole.String(), "vimes"))
	carrot := data.RoleName(path.Join(vimes.String(), "carrot"))
	vimesWild := data.RoleName(path.Join(vimes.String(), "*"))

	// create 2 keys, we'll purge one of them
	testKey1, err := ed25519.Create(vetinari, testGUN, data.ED25519Key)
	require.NoError(t, err)
	testKey2, err := ed25519.Create(vetinari, testGUN, data.ED25519Key)
	require.NoError(t, err)

	// create some delegations
	err = repo.UpdateDelegationKeys(vetinari, []data.PublicKey{testKey1, testKey2}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths(vetinari, []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys(sybil, []data.PublicKey{testKey1}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths(sybil, []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys(vimes, []data.PublicKey{testKey2}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths(vimes, []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys(carrot, []data.PublicKey{testKey1}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths(carrot, []string{""}, []string{}, false)
	require.NoError(t, err)

	id1, err := utils.CanonicalKeyID(testKey1)
	require.NoError(t, err)
	err = repo.PurgeDelegationKeys(vimesWild, []string{id1})
	require.NoError(t, err)

	role, err := repo.GetDelegationRole(vetinari)
	require.NoError(t, err)
	require.Len(t, role.Keys, 2)
	_, ok := role.Keys[testKey1.ID()]
	require.True(t, ok)
	_, ok = role.Keys[testKey2.ID()]
	require.True(t, ok)

	role, err = repo.GetDelegationRole(sybil)
	require.NoError(t, err)
	require.Len(t, role.Keys, 1)
	_, ok = role.Keys[testKey1.ID()]
	require.True(t, ok)

	role, err = repo.GetDelegationRole(vimes)
	require.NoError(t, err)
	require.Len(t, role.Keys, 1)
	_, ok = role.Keys[testKey2.ID()]
	require.True(t, ok)

	role, err = repo.GetDelegationRole(carrot)
	require.NoError(t, err)
	require.Len(t, role.Keys, 0)
}

func TestPurgeDelegationsKeyBadWildRole(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	err := repo.PurgeDelegationKeys("targets/foo", nil)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}
func TestUpdateDelegationsParentMissing(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testDeepKey, err := ed25519.Create("targets/test/deep", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test/deep", []data.PublicKey{testDeepKey}, []string{}, 1)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 0)

	// no delegation metadata created for non-existent parent
	_, ok = repo.Targets["targets/test"]
	require.False(t, ok, "no targets file should be created for nonexistent parent delegation")
}

// Updating delegations needs to modify the parent of the role being updated.
// If there is no signing key for that parent, the delegation cannot be added.
func TestUpdateDelegationsMissingParentKey(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	// remove the target key (all keys)
	repo.cryptoService = signed.NewEd25519()

	roleKey, err := ed25519.Create("Invalid Role", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/role", []data.PublicKey{roleKey}, []string{}, 1)
	require.Error(t, err)
	require.IsType(t, signed.ErrNoKeys{}, err)

	// no empty delegation metadata created for new delegation
	_, ok := repo.Targets["targets/role"]
	require.False(t, ok, "no targets file should be created for empty delegation")
}

func TestUpdateDelegationsInvalidRole(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	roleKey, err := ed25519.Create("Invalid Role", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys(data.CanonicalRootRole, []data.PublicKey{roleKey}, []string{}, 1)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 0)

	// no delegation metadata created for invalid delegation
	_, ok = repo.Targets["root"]
	require.False(t, ok, "no targets file should be created since delegation failed")
}

// A delegation can be created with a role that is missing a signing key, so
// long as UpdateDelegations is called with the key
func TestUpdateDelegationsRoleThatIsMissingDelegationKey(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	roleKey, err := ed25519.Create("Invalid Role", testGUN, data.ED25519Key)
	require.NoError(t, err)

	// key should get added to role as part of updating the delegation
	err = repo.UpdateDelegationKeys("targets/role", []data.PublicKey{roleKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/role", []string{""}, []string{}, false)
	require.NoError(t, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 1)
	require.Len(t, r.Signed.Delegations.Keys, 1)
	keyIDs := r.Signed.Delegations.Roles[0].KeyIDs
	require.Len(t, keyIDs, 1)
	require.Equal(t, roleKey.ID(), keyIDs[0])
	require.True(t, r.Dirty)

	// no empty delegation metadata created for new delegation
	_, ok = repo.Targets["targets/role"]
	require.False(t, ok, "no targets file should be created for empty delegation")
}

func TestUpdateDelegationsNotEnoughKeys(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	roleKey, err := ed25519.Create("Invalid Role", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/role", []data.PublicKey{roleKey}, []string{}, 2)
	require.NoError(t, err)

	// no delegation metadata created for failed delegation
	_, ok := repo.Targets["targets/role"]
	require.False(t, ok, "no targets file should be created since delegation failed")
}

func TestUpdateDelegationsAddKeyToRole(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{"test"}, []string{}, false)
	require.NoError(t, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 1)
	require.Len(t, r.Signed.Delegations.Keys, 1)
	keyIDs := r.Signed.Delegations.Roles[0].KeyIDs
	require.Len(t, keyIDs, 1)
	require.Equal(t, testKey.ID(), keyIDs[0])

	testKey2, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey2}, []string{}, 1)
	require.NoError(t, err)

	r, ok = repo.Targets["targets"]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 1)
	require.Len(t, r.Signed.Delegations.Keys, 2)
	keyIDs = r.Signed.Delegations.Roles[0].KeyIDs
	require.Len(t, keyIDs, 2)
	// it does an append so the order is deterministic (but not meaningful to TUF)
	require.Equal(t, testKey.ID(), keyIDs[0])
	require.Equal(t, testKey2.ID(), keyIDs[1])
	require.True(t, r.Dirty)
}

func TestDeleteDelegations(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{"test"}, []string{}, false)
	require.NoError(t, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 1)
	require.Len(t, r.Signed.Delegations.Keys, 1)
	keyIDs := r.Signed.Delegations.Roles[0].KeyIDs
	require.Len(t, keyIDs, 1)
	require.Equal(t, testKey.ID(), keyIDs[0])

	// ensure that the metadata is there and snapshot is there
	targets, err := repo.InitTargets("targets/test")
	require.NoError(t, err)
	targetsSigned, err := targets.ToSigned()
	require.NoError(t, err)
	require.NoError(t, repo.UpdateSnapshot("targets/test", targetsSigned))
	_, ok = repo.Snapshot.Signed.Meta["targets/test"]
	require.True(t, ok)

	require.NoError(t, repo.DeleteDelegation("targets/test"))
	require.Len(t, r.Signed.Delegations.Roles, 0)
	require.Len(t, r.Signed.Delegations.Keys, 0)
	require.True(t, r.Dirty)

	// metadata should be deleted
	_, ok = repo.Targets["targets/test"]
	require.False(t, ok)
	_, ok = repo.Snapshot.Signed.Meta["targets/test"]
	require.False(t, ok)
}

func TestDeleteDelegationsRoleNotExistBecauseNoParentMeta(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{"test"}, []string{}, false)
	require.NoError(t, err)

	// no empty delegation metadata created for new delegation
	_, ok := repo.Targets["targets/test"]
	require.False(t, ok, "no targets file should be created for empty delegation")

	delRole, err := data.NewRole("targets/test/a", 1, []string{testKey.ID()}, []string{"test"})
	require.NoError(t, err)

	err = repo.DeleteDelegation(delRole.Name)
	require.NoError(t, err)
	// still no metadata
	_, ok = repo.Targets["targets/test"]
	require.False(t, ok)
}

func TestDeleteDelegationsRoleNotExist(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	// initRepo leaves all the roles as Dirty. Set to false
	// to test removing a non-existent role doesn't mark
	// a role as dirty
	repo.Targets[data.CanonicalTargetsRole].Dirty = false

	role, err := data.NewRole("targets/test", 1, []string{}, []string{""})
	require.NoError(t, err)

	err = repo.DeleteDelegation(role.Name)
	require.NoError(t, err)
	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 0)
	require.Len(t, r.Signed.Delegations.Keys, 0)
	require.False(t, r.Dirty)
}

func TestDeleteDelegationsInvalidRole(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	// data.NewRole errors if the role isn't a valid TUF role so use one of the non-delegation
	// valid roles
	invalidRole, err := data.NewRole("root", 1, []string{}, []string{""})
	require.NoError(t, err)

	err = repo.DeleteDelegation(invalidRole.Name)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 0)
}

func TestDeleteDelegationsParentMissing(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testRole, err := data.NewRole("targets/test/deep", 1, []string{}, []string{""})
	require.NoError(t, err)

	err = repo.DeleteDelegation(testRole.Name)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 0)
}

// Can't delete a delegation if we don't have the parent's signing key
func TestDeleteDelegationsMissingParentSigningKey(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{"test"}, []string{}, false)
	require.NoError(t, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 1)
	require.Len(t, r.Signed.Delegations.Keys, 1)
	keyIDs := r.Signed.Delegations.Roles[0].KeyIDs
	require.Len(t, keyIDs, 1)
	require.Equal(t, testKey.ID(), keyIDs[0])

	// ensure that the metadata is there and snapshot is there
	targets, err := repo.InitTargets("targets/test")
	require.NoError(t, err)
	targetsSigned, err := targets.ToSigned()
	require.NoError(t, err)
	require.NoError(t, repo.UpdateSnapshot("targets/test", targetsSigned))
	_, ok = repo.Snapshot.Signed.Meta["targets/test"]
	require.True(t, ok)

	// delete all signing keys
	repo.cryptoService = signed.NewEd25519()
	err = repo.DeleteDelegation("targets/test")
	require.Error(t, err)
	require.IsType(t, signed.ErrNoKeys{}, err)

	require.Len(t, r.Signed.Delegations.Roles, 1)
	require.Len(t, r.Signed.Delegations.Keys, 1)
	require.True(t, r.Dirty)

	// metadata should be here still
	_, ok = repo.Targets["targets/test"]
	require.True(t, ok)
	_, ok = repo.Snapshot.Signed.Meta["targets/test"]
	require.True(t, ok)
}

func TestDeleteDelegationsMidSliceRole(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/test2", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test2", []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/test3", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test3", []string{"test"}, []string{}, false)
	require.NoError(t, err)

	err = repo.DeleteDelegation("targets/test2")
	require.NoError(t, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	require.Len(t, r.Signed.Delegations.Roles, 2)
	require.Len(t, r.Signed.Delegations.Keys, 1)
	require.True(t, r.Dirty)
}

// If the parent exists, the metadata exists, and the delegation is in it,
// returns the role that was found
func TestGetDelegationRoleAndMetadataExistDelegationExists(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("meh", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/level1", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/level1", []string{""}, []string{}, false)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/level1/level2", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/level1/level2", []string{""}, []string{}, false)
	require.NoError(t, err)

	gottenRole, err := repo.GetDelegationRole("targets/level1/level2")
	require.NoError(t, err)
	require.EqualValues(t, "targets/level1/level2", gottenRole.Name)
	require.Equal(t, 1, gottenRole.Threshold)
	require.Equal(t, []string{""}, gottenRole.Paths)
	_, ok := gottenRole.Keys[testKey.ID()]
	require.True(t, ok)
}

// If the parent exists, the metadata exists, and the delegation isn't in it,
// returns an ErrNoSuchRole
func TestGetDelegationRoleAndMetadataExistDelegationDoesntExists(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("meh", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/level1", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/level1", []string{""}, []string{}, false)
	require.NoError(t, err)

	// ensure metadata exists
	repo.InitTargets("targets/level1")

	_, err = repo.GetDelegationRole("targets/level1/level2")
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}

// If the parent exists but the metadata doesn't exist, returns an ErrNoSuchRole
func TestGetDelegationRoleAndMetadataDoesntExists(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("meh", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/level1", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/level1", []string{""}, []string{}, false)
	require.NoError(t, err)

	// no empty delegation metadata created for new delegation
	_, ok := repo.Targets["targets/test"]
	require.False(t, ok, "no targets file should be created for empty delegation")

	_, err = repo.GetDelegationRole("targets/level1/level2")
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}

// If the parent role doesn't exist, GetDelegation fails with an ErrInvalidRole
func TestGetDelegationParentMissing(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	_, err := repo.GetDelegationRole("targets/level1/level2")
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}

// Adding targets to a role that exists and has metadata (like targets)
// correctly adds the target
func TestAddTargetsRoleAndMetadataExist(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	hash := sha256.Sum256([]byte{})
	f := data.FileMeta{
		Length: 1,
		Hashes: map[string][]byte{
			"sha256": hash[:],
		},
	}

	_, err := repo.AddTargets(data.CanonicalTargetsRole, data.Files{"f": f})
	require.NoError(t, err)

	r, ok := repo.Targets[data.CanonicalTargetsRole]
	require.True(t, ok)
	targetsF, ok := r.Signed.Targets["f"]
	require.True(t, ok)
	require.Equal(t, f, targetsF)
}

// Adding targets to a role that exists and has not metadata first creates the
// metadata and then correctly adds the target
func TestAddTargetsRoleExistsAndMetadataDoesntExist(t *testing.T) {
	hash := sha256.Sum256([]byte{})
	f := data.FileMeta{
		Length: 1,
		Hashes: map[string][]byte{
			"sha256": hash[:],
		},
	}

	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{""}, []string{}, false)
	require.NoError(t, err)

	// no empty metadata is created for this role
	_, ok := repo.Targets["targets/test"]
	require.False(t, ok, "no empty targets file should be created")

	// adding the targets to the role should create the metadata though
	_, err = repo.AddTargets("targets/test", data.Files{"f": f})
	require.NoError(t, err)

	r, ok := repo.Targets["targets/test"]
	require.True(t, ok)
	targetsF, ok := r.Signed.Targets["f"]
	require.True(t, ok)
	require.Equal(t, f, targetsF)
	require.True(t, r.Dirty)

	// set it to not dirty so we can assert that if we add the exact same data, it won't be dirty
	r.Dirty = false
	_, err = repo.AddTargets("targets/test", data.Files{"f": f})
	require.NoError(t, err)
	require.False(t, r.Dirty)

	// If we add the same target but with different metadata, it's dirty again
	f2 := f
	f2.Length = 2
	_, err = repo.AddTargets("targets/test", data.Files{"f": f2})
	require.NoError(t, err)
	require.True(t, r.Dirty)
}

// Adding targets to a role that doesn't exist fails only if a target was actually added or updated
func TestAddTargetsRoleDoesntExist(t *testing.T) {
	hash := sha256.Sum256([]byte{})
	f := data.FileMeta{
		Length: 1,
		Hashes: map[string][]byte{
			"sha256": hash[:],
		},
	}

	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	_, err := repo.AddTargets("targets/test", data.Files{"f": f})
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}

// Adding targets to a role that we don't have signing keys for fails
func TestAddTargetsNoSigningKeys(t *testing.T) {
	hash := sha256.Sum256([]byte{})
	f := data.FileMeta{
		Length: 1,
		Hashes: map[string][]byte{
			"sha256": hash[:],
		},
	}

	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{""}, []string{}, false)
	require.NoError(t, err)

	_, err = repo.AddTargets("targets/test", data.Files{"f": f})
	require.NoError(t, err)

	// now delete the signing key (all keys)
	repo.cryptoService = signed.NewEd25519()

	// adding the same exact target to the role should succeed even though the key is missing
	_, err = repo.AddTargets("targets/test", data.Files{"f": f})
	require.NoError(t, err)

	// adding a different target to the role should fail because the keys is missing
	_, err = repo.AddTargets("targets/test", data.Files{"t": f})
	require.Error(t, err)
	require.IsType(t, signed.ErrNoKeys{}, err)
}

// Removing targets from a role that exists, has targets, and is signable
// should succeed, even if we also want to remove targets that don't exist.
func TestRemoveExistingAndNonexistingTargets(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{"test"}, []string{}, false)
	require.NoError(t, err)

	// no empty metadata is created for this role
	_, ok := repo.Targets["targets/test"]
	require.False(t, ok, "no empty targets file should be created")

	// now remove a target
	require.NoError(t, repo.RemoveTargets("targets/test", "f"))

	// still no metadata
	_, ok = repo.Targets["targets/test"]
	require.False(t, ok)

	// add a target to remove
	hash := sha256.Sum256([]byte{})
	_, err = repo.AddTargets("targets/test", data.Files{"test": data.FileMeta{
		Length: 1,
		Hashes: map[string][]byte{
			"sha256": hash[:],
		},
	}})
	require.NoError(t, err)
	tgt, ok := repo.Targets["targets/test"]
	require.True(t, ok)
	require.True(t, tgt.Dirty)
	// set this to false so we can prove that removing a non-existing target does not mark as dirty
	tgt.Dirty = false

	require.NoError(t, repo.RemoveTargets("targets/test", "not_real"))
	require.False(t, tgt.Dirty)
	require.NotEmpty(t, tgt.Signed.Targets)

	require.NoError(t, repo.RemoveTargets("targets/test", "test"))
	require.True(t, tgt.Dirty)
	require.Empty(t, tgt.Signed.Targets)
}

// Removing targets from a role that doesn't exist fails
func TestRemoveTargetsRoleDoesntExist(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	err := repo.RemoveTargets("targets/test", "f")
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}

// Removing targets from a role that we don't have signing keys for fails only if
// a target was actually removed
func TestRemoveTargetsNoSigningKeys(t *testing.T) {
	hash := sha256.Sum256([]byte{})
	f := data.FileMeta{
		Length: 1,
		Hashes: map[string][]byte{
			"sha256": hash[:],
		},
	}

	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{""}, []string{}, false)
	require.NoError(t, err)

	// adding the targets to the role should create the metadata though
	_, err = repo.AddTargets("targets/test", data.Files{"f": f})
	require.NoError(t, err)

	r, ok := repo.Targets["targets/test"]
	require.True(t, ok)
	_, ok = r.Signed.Targets["f"]
	require.True(t, ok)

	// now delete the signing key (all keys)
	repo.cryptoService = signed.NewEd25519()

	// remove a nonexistent target - it should not fail
	err = repo.RemoveTargets("targets/test", "t")
	require.NoError(t, err)

	// now remove a target that does exist - it should fail
	err = repo.RemoveTargets("targets/test", "t", "f", "g")
	require.Error(t, err)
	require.IsType(t, signed.ErrNoKeys{}, err)
}

// adding a key to a role marks root as dirty as well as the role
func TestAddBaseKeysToRoot(t *testing.T) {
	for _, role := range data.BaseRoles {
		ed25519 := signed.NewEd25519()
		repo := initRepo(t, ed25519)

		origKeyIDs := ed25519.ListKeys(role)
		require.Len(t, origKeyIDs, 1)

		key, err := ed25519.Create(role, testGUN, data.ED25519Key)
		require.NoError(t, err)

		require.Len(t, repo.Root.Signed.Roles[role].KeyIDs, 1)

		require.NoError(t, repo.AddBaseKeys(role, key))

		_, ok := repo.Root.Signed.Keys[key.ID()]
		require.True(t, ok)
		require.Len(t, repo.Root.Signed.Roles[role].KeyIDs, 2)
		require.True(t, repo.Root.Dirty)

		switch role {
		case data.CanonicalSnapshotRole:
			require.True(t, repo.Snapshot.Dirty)
		case data.CanonicalTargetsRole:
			require.True(t, repo.Targets[data.CanonicalTargetsRole].Dirty)
		case data.CanonicalTimestampRole:
			require.True(t, repo.Timestamp.Dirty)
		case data.CanonicalRootRole:
			require.NoError(t, err)
			require.Len(t, repo.originalRootRole.Keys, 1)
			require.Contains(t, repo.originalRootRole.ListKeyIDs(), origKeyIDs[0])
		}
	}
}

// removing one or more keys from a role marks root as dirty as well as the role
func TestRemoveBaseKeysFromRoot(t *testing.T) {
	for _, role := range data.BaseRoles {
		ed25519 := signed.NewEd25519()
		repo := initRepo(t, ed25519)

		origKeyIDs := ed25519.ListKeys(role)
		require.Len(t, origKeyIDs, 1)

		require.Len(t, repo.Root.Signed.Roles[role].KeyIDs, 1)

		require.NoError(t, repo.RemoveBaseKeys(role, origKeyIDs...))

		require.Len(t, repo.Root.Signed.Roles[role].KeyIDs, 0)
		require.True(t, repo.Root.Dirty)

		switch role {
		case data.CanonicalSnapshotRole:
			require.True(t, repo.Snapshot.Dirty)
		case data.CanonicalTargetsRole:
			require.True(t, repo.Targets[data.CanonicalTargetsRole].Dirty)
		case data.CanonicalTimestampRole:
			require.True(t, repo.Timestamp.Dirty)
		case data.CanonicalRootRole:
			require.Len(t, repo.originalRootRole.Keys, 1)
			require.Contains(t, repo.originalRootRole.ListKeyIDs(), origKeyIDs[0])
		}
	}
}

// replacing keys in a role marks root as dirty as well as the role
func TestReplaceBaseKeysInRoot(t *testing.T) {
	for _, role := range data.BaseRoles {
		ed25519 := signed.NewEd25519()
		repo := initRepo(t, ed25519)

		origKeyIDs := ed25519.ListKeys(role)
		require.Len(t, origKeyIDs, 1)

		key, err := ed25519.Create(role, testGUN, data.ED25519Key)
		require.NoError(t, err)

		require.Len(t, repo.Root.Signed.Roles[role].KeyIDs, 1)

		require.NoError(t, repo.ReplaceBaseKeys(role, key))

		_, ok := repo.Root.Signed.Keys[key.ID()]
		require.True(t, ok)
		require.Len(t, repo.Root.Signed.Roles[role].KeyIDs, 1)
		require.True(t, repo.Root.Dirty)

		switch role {
		case data.CanonicalSnapshotRole:
			require.True(t, repo.Snapshot.Dirty)
		case data.CanonicalTargetsRole:
			require.True(t, repo.Targets[data.CanonicalTargetsRole].Dirty)
		case data.CanonicalTimestampRole:
			require.True(t, repo.Timestamp.Dirty)
		case data.CanonicalRootRole:
			require.Len(t, repo.originalRootRole.Keys, 1)
			require.Contains(t, repo.originalRootRole.ListKeyIDs(), origKeyIDs[0])
		}

		origNumRoles := len(repo.Root.Signed.Roles)
		// sign the root and assert the number of roles after
		_, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
		require.NoError(t, err)
		// number of roles should not have changed
		require.Len(t, repo.Root.Signed.Roles, origNumRoles)
	}
}

func TestGetAllRoles(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	// After we init, we get the base roles
	roles := repo.GetAllLoadedRoles()
	require.Len(t, roles, len(data.BaseRoles))
}

func TestGetBaseRoles(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	// After we init, we get the base roles
	for _, role := range data.BaseRoles {
		baseRole, err := repo.GetBaseRole(role)
		require.NoError(t, err)

		require.Equal(t, role, baseRole.Name)
		keyIDs := repo.cryptoService.ListKeys(role)
		for _, keyID := range keyIDs {
			_, ok := baseRole.Keys[keyID]
			require.True(t, ok)
			require.Contains(t, baseRole.ListKeyIDs(), keyID)
		}
		// initRepo should set all key thresholds to 1
		require.Equal(t, 1, baseRole.Threshold)
	}
}

func TestGetBaseRolesInvalidName(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	_, err := repo.GetBaseRole("invalid")
	require.Error(t, err)

	_, err = repo.GetBaseRole("targets/delegation")
	require.Error(t, err)
}

func TestGetDelegationValidRoles(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey1, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey1}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{"path", "anotherpath"}, []string{}, false)
	require.NoError(t, err)

	delgRole, err := repo.GetDelegationRole("targets/test")
	require.NoError(t, err)
	require.EqualValues(t, "targets/test", delgRole.Name)
	require.Equal(t, 1, delgRole.Threshold)
	require.Equal(t, []string{testKey1.ID()}, delgRole.ListKeyIDs())
	require.Equal(t, []string{"path", "anotherpath"}, delgRole.Paths)
	require.Equal(t, testKey1, delgRole.Keys[testKey1.ID()])

	testKey2, err := ed25519.Create("targets/a", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/a", []data.PublicKey{testKey2}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/a", []string{""}, []string{}, false)
	require.NoError(t, err)

	delgRole, err = repo.GetDelegationRole("targets/a")
	require.NoError(t, err)
	require.EqualValues(t, "targets/a", delgRole.Name)
	require.Equal(t, 1, delgRole.Threshold)
	require.Equal(t, []string{testKey2.ID()}, delgRole.ListKeyIDs())
	require.Equal(t, []string{""}, delgRole.Paths)
	require.Equal(t, testKey2, delgRole.Keys[testKey2.ID()])

	testKey3, err := ed25519.Create("targets/test/b", testGUN, data.ED25519Key)
	require.NoError(t, err)
	err = repo.UpdateDelegationKeys("targets/test/b", []data.PublicKey{testKey3}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test/b", []string{"path/subpath", "anotherpath"}, []string{}, false)
	require.NoError(t, err)

	delgRole, err = repo.GetDelegationRole("targets/test/b")
	require.NoError(t, err)
	require.EqualValues(t, "targets/test/b", delgRole.Name)
	require.Equal(t, 1, delgRole.Threshold)
	require.Equal(t, []string{testKey3.ID()}, delgRole.ListKeyIDs())
	require.Equal(t, []string{"path/subpath", "anotherpath"}, delgRole.Paths)
	require.Equal(t, testKey3, delgRole.Keys[testKey3.ID()])

	testKey4, err := ed25519.Create("targets/test/c", testGUN, data.ED25519Key)
	require.NoError(t, err)
	// Try adding empty paths, ensure this is valid
	err = repo.UpdateDelegationKeys("targets/test/c", []data.PublicKey{testKey4}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test/c", []string{}, []string{}, false)
	require.NoError(t, err)
}

func TestGetDelegationRolesInvalidName(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	_, err := repo.GetDelegationRole("invalid")
	require.Error(t, err)

	for _, role := range data.BaseRoles {
		_, err = repo.GetDelegationRole(role)
		require.Error(t, err)
		require.IsType(t, data.ErrInvalidRole{}, err)
	}
	_, err = repo.GetDelegationRole("targets/does_not_exist")
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}

func TestGetDelegationRolesInvalidPaths(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	testKey1, err := ed25519.Create("targets/test", testGUN, data.ED25519Key)
	require.NoError(t, err)

	err = repo.UpdateDelegationKeys("targets/test", []data.PublicKey{testKey1}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test", []string{"path", "anotherpath"}, []string{}, false)
	require.NoError(t, err)

	testKey2, err := ed25519.Create("targets/test/b", testGUN, data.ED25519Key)
	require.NoError(t, err)
	// Now we add a delegation with a path that is not prefixed by its parent delegation, the invalid path can't be added so there is an error
	err = repo.UpdateDelegationKeys("targets/test/b", []data.PublicKey{testKey2}, []string{}, 1)
	require.NoError(t, err)
	err = repo.UpdateDelegationPaths("targets/test/b", []string{"invalidpath"}, []string{}, false)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)

	delgRole, err := repo.GetDelegationRole("targets/test")
	require.NoError(t, err)
	require.Contains(t, delgRole.Paths, "path")
	require.Contains(t, delgRole.Paths, "anotherpath")
}

func TestDelegationRolesParent(t *testing.T) {
	delgA := data.DelegationRole{
		BaseRole: data.BaseRole{
			Keys:      nil,
			Name:      "targets/a",
			Threshold: 1,
		},
		Paths: []string{"path", "anotherpath"},
	}

	delgB := data.DelegationRole{
		BaseRole: data.BaseRole{
			Keys:      nil,
			Name:      "targets/a/b",
			Threshold: 1,
		},
		Paths: []string{"path/b", "anotherpath/b", "b/invalidpath"},
	}

	// Assert direct parent relationship
	require.True(t, delgA.IsParentOf(delgB))
	require.False(t, delgB.IsParentOf(delgA))
	require.False(t, delgA.IsParentOf(delgA))

	delgC := data.DelegationRole{
		BaseRole: data.BaseRole{
			Keys:      nil,
			Name:      "targets/a/b/c",
			Threshold: 1,
		},
		Paths: []string{"path/b", "anotherpath/b/c", "c/invalidpath"},
	}

	// Assert direct parent relationship
	require.True(t, delgB.IsParentOf(delgC))
	require.False(t, delgB.IsParentOf(delgB))
	require.False(t, delgA.IsParentOf(delgC))
	require.False(t, delgC.IsParentOf(delgB))
	require.False(t, delgC.IsParentOf(delgA))
	require.False(t, delgC.IsParentOf(delgC))

	// Check that parents correctly restrict paths
	restrictedDelgB, err := delgA.Restrict(delgB)
	require.NoError(t, err)
	require.Contains(t, restrictedDelgB.Paths, "path/b")
	require.Contains(t, restrictedDelgB.Paths, "anotherpath/b")
	require.NotContains(t, restrictedDelgB.Paths, "b/invalidpath")

	_, err = delgB.Restrict(delgA)
	require.Error(t, err)
	_, err = delgA.Restrict(delgC)
	require.Error(t, err)
	_, err = delgC.Restrict(delgB)
	require.Error(t, err)
	_, err = delgC.Restrict(delgA)
	require.Error(t, err)

	// Make delgA have no paths and check that it changes delgB and delgC accordingly when chained
	delgA.Paths = []string{}
	restrictedDelgB, err = delgA.Restrict(delgB)
	require.NoError(t, err)
	require.Empty(t, restrictedDelgB.Paths)
	restrictedDelgC, err := restrictedDelgB.Restrict(delgC)
	require.NoError(t, err)
	require.Empty(t, restrictedDelgC.Paths)
}

func TestGetBaseRoleEmptyRepo(t *testing.T) {
	repo := NewRepo(nil)
	_, err := repo.GetBaseRole(data.CanonicalRootRole)
	require.Error(t, err)
	require.IsType(t, ErrNotLoaded{}, err)
}

func TestGetBaseRoleKeyMissing(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	// change root role to have a KeyID that doesn't exist
	repo.Root.Signed.Roles[data.CanonicalRootRole].KeyIDs = []string{"abc"}

	_, err := repo.GetBaseRole(data.CanonicalRootRole)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}

func TestGetDelegationRoleKeyMissing(t *testing.T) {
	ed25519 := signed.NewEd25519()
	repo := initRepo(t, ed25519)

	// add a delegation that has a KeyID that doesn't exist
	// in the relevant key map
	tar := repo.Targets[data.CanonicalTargetsRole]
	tar.Signed.Delegations.Roles = []*data.Role{
		{
			RootRole: data.RootRole{
				KeyIDs:    []string{"abc"},
				Threshold: 1,
			},
			Name:  "targets/missing_key",
			Paths: []string{""},
		},
	}

	_, err := repo.GetDelegationRole("targets/missing_key")
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)
}

func verifySignatureList(t *testing.T, signed *data.Signed, expectedKeys ...data.PublicKey) {
	require.Equal(t, len(expectedKeys), len(signed.Signatures))
	usedKeys := make(map[string]struct{}, len(signed.Signatures))
	for _, sig := range signed.Signatures {
		usedKeys[sig.KeyID] = struct{}{}
	}
	for _, key := range expectedKeys {
		_, ok := usedKeys[key.ID()]
		require.True(t, ok)
		verifyRootSignatureAgainstKey(t, signed, key)
	}
}

func verifyRootSignatureAgainstKey(t *testing.T, signedRoot *data.Signed, key data.PublicKey) error {
	roleWithKeys := data.BaseRole{Name: data.CanonicalRootRole, Keys: data.Keys{key.ID(): key}, Threshold: 1}
	return signed.VerifySignatures(signedRoot, roleWithKeys)
}

func TestSignRootOldKeyCertExists(t *testing.T) {
	var gun data.GUN = "docker/test-sign-root"
	referenceTime := time.Now()

	cs := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(
		passphrase.ConstantRetriever("password")))

	rootPublicKey, err := cs.Create(data.CanonicalRootRole, gun, data.ECDSAKey)
	require.NoError(t, err)
	rootPrivateKey, _, err := cs.GetPrivateKey(rootPublicKey.ID())
	require.NoError(t, err)
	oldRootCert, err := cryptoservice.GenerateCertificate(rootPrivateKey, gun, referenceTime.AddDate(-9, 0, 0),
		referenceTime.AddDate(1, 0, 0))
	require.NoError(t, err)
	oldRootCertKey := utils.CertToKey(oldRootCert)

	repo := initRepoWithRoot(t, cs, oldRootCertKey)

	// Create a first signature, using the old key.
	signedRoot, err := repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	verifySignatureList(t, signedRoot, oldRootCertKey)
	err = verifyRootSignatureAgainstKey(t, signedRoot, oldRootCertKey)
	require.NoError(t, err)

	// Create a new certificate
	newRootCert, err := cryptoservice.GenerateCertificate(rootPrivateKey, gun, referenceTime, referenceTime.AddDate(10, 0, 0))
	require.NoError(t, err)
	newRootCertKey := utils.CertToKey(newRootCert)
	require.NotEqual(t, oldRootCertKey.ID(), newRootCertKey.ID())

	// Only trust the new certificate
	err = repo.ReplaceBaseKeys(data.CanonicalRootRole, newRootCertKey)
	require.NoError(t, err)
	updatedRootRole, err := repo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	updatedRootKeyIDs := updatedRootRole.ListKeyIDs()
	require.Equal(t, 1, len(updatedRootKeyIDs))
	require.Equal(t, newRootCertKey.ID(), updatedRootKeyIDs[0])

	// Create a second signature
	signedRoot, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	verifySignatureList(t, signedRoot, oldRootCertKey, newRootCertKey)

	// Verify that the signature can be verified when trusting the old certificate
	err = verifyRootSignatureAgainstKey(t, signedRoot, oldRootCertKey)
	require.NoError(t, err)
	// Verify that the signature can be verified when trusting the new certificate
	err = verifyRootSignatureAgainstKey(t, signedRoot, newRootCertKey)
	require.NoError(t, err)
}

func TestSignRootOldKeyCertMissing(t *testing.T) {
	var gun data.GUN = "docker/test-sign-root"
	referenceTime := time.Now()

	cs := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(
		passphrase.ConstantRetriever("password")))

	rootPublicKey, err := cs.Create(data.CanonicalRootRole, gun, data.ECDSAKey)
	require.NoError(t, err)
	rootPrivateKey, _, err := cs.GetPrivateKey(rootPublicKey.ID())
	require.NoError(t, err)
	oldRootCert, err := cryptoservice.GenerateCertificate(rootPrivateKey, gun, referenceTime.AddDate(-9, 0, 0),
		referenceTime.AddDate(1, 0, 0))
	require.NoError(t, err)
	oldRootCertKey := utils.CertToKey(oldRootCert)

	repo := initRepoWithRoot(t, cs, oldRootCertKey)

	// Create a first signature, using the old key.
	signedRoot, err := repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	verifySignatureList(t, signedRoot, oldRootCertKey)
	err = verifyRootSignatureAgainstKey(t, signedRoot, oldRootCertKey)
	require.NoError(t, err)

	// Create a new certificate
	newRootCert, err := cryptoservice.GenerateCertificate(rootPrivateKey, gun, referenceTime, referenceTime.AddDate(10, 0, 0))
	require.NoError(t, err)
	newRootCertKey := utils.CertToKey(newRootCert)
	require.NotEqual(t, oldRootCertKey.ID(), newRootCertKey.ID())

	// Only trust the new certificate
	err = repo.ReplaceBaseKeys(data.CanonicalRootRole, newRootCertKey)
	require.NoError(t, err)
	updatedRootRole, err := repo.GetBaseRole(data.CanonicalRootRole)
	require.NoError(t, err)
	updatedRootKeyIDs := updatedRootRole.ListKeyIDs()
	require.Equal(t, 1, len(updatedRootKeyIDs))
	require.Equal(t, newRootCertKey.ID(), updatedRootKeyIDs[0])

	// Now forget all about the old certificate: drop it from the Root carried keys
	delete(repo.Root.Signed.Keys, oldRootCertKey.ID())
	repo2 := NewRepo(cs)
	repo2.Root = repo.Root
	repo2.originalRootRole = updatedRootRole

	// Create a second signature
	signedRoot, err = repo2.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	verifySignatureList(t, signedRoot, newRootCertKey) // Without oldRootCertKey

	// Verify that the signature can be verified when trusting the new certificate
	err = verifyRootSignatureAgainstKey(t, signedRoot, newRootCertKey)
	require.NoError(t, err)
	err = verifyRootSignatureAgainstKey(t, signedRoot, oldRootCertKey)
	require.Error(t, err)
}

// SignRoot signs with the current root and the previous, to allow root key
// rotation. After signing with the previous keys, they can be discarded from
// the root role.
func TestRootKeyRotation(t *testing.T) {
	var gun data.GUN = "docker/test-sign-root"
	referenceTime := time.Now()

	cs := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(
		passphrase.ConstantRetriever("password")))

	rootCertKeys := make([]data.PublicKey, 7)
	rootPrivKeys := make([]data.PrivateKey, cap(rootCertKeys))
	for i := 0; i < cap(rootCertKeys); i++ {
		rootPublicKey, err := cs.Create(data.CanonicalRootRole, gun, data.ECDSAKey)
		require.NoError(t, err)
		rootPrivateKey, _, err := cs.GetPrivateKey(rootPublicKey.ID())
		require.NoError(t, err)
		rootCert, err := cryptoservice.GenerateCertificate(rootPrivateKey, gun, referenceTime.AddDate(-9, 0, 0),
			referenceTime.AddDate(1, 0, 0))
		require.NoError(t, err)
		rootCertKeys[i] = utils.CertToKey(rootCert)
		rootPrivKeys[i] = rootPrivateKey
	}

	// Initialize and sign with one key
	repo := initRepoWithRoot(t, cs, rootCertKeys[0])
	signedObj, err := repo.Root.ToSigned()
	require.NoError(t, err)
	signedObj, err = repo.sign(signedObj, nil, []data.PublicKey{rootCertKeys[0]})
	require.NoError(t, err)
	verifySignatureList(t, signedObj, rootCertKeys[0])
	repo.Root.Signatures = signedObj.Signatures

	// Add new root key, should sign with previous and new
	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, rootCertKeys[1]))
	signedObj, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	expectedSigningKeys := []data.PublicKey{
		rootCertKeys[0],
		rootCertKeys[1],
	}
	verifySignatureList(t, signedObj, expectedSigningKeys...)

	// Add new root key, should sign with previous and new, not with old
	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, rootCertKeys[2]))
	signedObj, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	expectedSigningKeys = []data.PublicKey{
		rootCertKeys[1],
		rootCertKeys[2],
	}
	verifySignatureList(t, signedObj, expectedSigningKeys...)

	// Rotate to two new keys, should be signed with previous and current (3 total)
	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, rootCertKeys[3], rootCertKeys[4]))
	signedObj, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	expectedSigningKeys = []data.PublicKey{
		rootCertKeys[2],
		rootCertKeys[3],
		rootCertKeys[4],
	}
	verifySignatureList(t, signedObj, expectedSigningKeys...)

	// Rotate to two new keys, should be signed with previous set and current set (4 total)
	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalRootRole, rootCertKeys[5], rootCertKeys[6]))
	signedObj, err = repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	expectedSigningKeys = []data.PublicKey{
		rootCertKeys[3],
		rootCertKeys[4],
		rootCertKeys[5],
		rootCertKeys[6],
	}
	verifySignatureList(t, signedObj, expectedSigningKeys...)
}
