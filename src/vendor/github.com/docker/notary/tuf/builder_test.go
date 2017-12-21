package tuf_test

// tests for builder that live in an external package, tuf_test, so that we can use
// the testutils without causing an import cycle

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/docker/notary"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/testutils"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

var _cachedMeta map[data.RoleName][]byte

// we just want sample metadata for a role - so we can build cached metadata
// and use it once.
func getSampleMeta(t *testing.T) (map[data.RoleName][]byte, data.GUN) {
	var gun data.GUN = "docker.com/notary"
	delgNames := []data.RoleName{"targets/a", "targets/a/b", "targets/a/b/force_parent_metadata"}
	if _cachedMeta == nil {
		meta, _, err := testutils.NewRepoMetadata(gun, delgNames...)
		require.NoError(t, err)

		_cachedMeta = meta
	}
	return _cachedMeta, gun
}

// We load only if the rolename is a valid rolename - even if the metadata we provided is valid
func TestBuilderLoadsValidRolesOnly(t *testing.T) {
	meta, gun := getSampleMeta(t)
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	err := builder.Load("NotRoot", meta[data.CanonicalRootRole], 1, false)
	require.Error(t, err)
	require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
	require.Contains(t, err.Error(), "is an invalid role")
}

func TestBuilderOnlyAcceptsRootFirstWhenLoading(t *testing.T) {
	meta, gun := getSampleMeta(t)
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})

	for roleName, content := range meta {
		if roleName != data.CanonicalRootRole {
			err := builder.Load(roleName, content, 1, true)
			require.Error(t, err)
			require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
			require.Contains(t, err.Error(), "root must be loaded first")
			require.False(t, builder.IsLoaded(roleName))
			require.Equal(t, 1, builder.GetLoadedVersion(roleName))
		}
	}

	// we can load the root
	require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
	require.True(t, builder.IsLoaded(data.CanonicalRootRole))
}

func TestBuilderOnlyAcceptsDelegationsAfterParent(t *testing.T) {
	meta, gun := getSampleMeta(t)
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})

	// load the root
	require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))

	// delegations can't be loaded without target
	for _, delgName := range []data.RoleName{"targets/a", "targets/a/b"} {
		err := builder.Load(delgName, meta[delgName], 1, false)
		require.Error(t, err)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "targets must be loaded first")
		require.False(t, builder.IsLoaded(delgName))
		require.Equal(t, 1, builder.GetLoadedVersion(delgName))
	}

	// load the targets
	require.NoError(t, builder.Load(data.CanonicalTargetsRole, meta[data.CanonicalTargetsRole], 1, false))

	// targets/a/b can't be loaded because targets/a isn't loaded
	err := builder.Load("targets/a/b", meta["targets/a/b"], 1, false)
	require.Error(t, err)
	require.IsType(t, data.ErrInvalidRole{}, err)

	// targets/a can be loaded now though because targets is loaded
	require.NoError(t, builder.Load("targets/a", meta["targets/a"], 1, false))

	// and now targets/a/b can be loaded because targets/a is loaded
	require.NoError(t, builder.Load("targets/a/b", meta["targets/a/b"], 1, false))
}

func TestMarkingIsValid(t *testing.T) {
	meta, gun := getSampleMeta(t)
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})

	// testing that the signed objects have a false isValid value confirming
	// that verify signatures has not been called on them yet
	// now when we check that isValid is true after calling load which calls
	// verify signatures- we can be sure that verify signatures is actually
	// setting the isValid fields for our data.Signed objects
	for _, meta := range meta {
		signedObj := &data.Signed{}
		if err := json.Unmarshal(meta, signedObj); err != nil {
			require.NoError(t, err)
		}
		require.Len(t, signedObj.Signatures, 1)
		require.False(t, signedObj.Signatures[0].IsValid)
	}

	// load the root
	require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))

	// load a timestamp
	require.NoError(t, builder.Load(data.CanonicalTimestampRole, meta[data.CanonicalTimestampRole], 1, false))

	// load a snapshot
	require.NoError(t, builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false))

	// load the targets
	require.NoError(t, builder.Load(data.CanonicalTargetsRole, meta[data.CanonicalTargetsRole], 1, false))

	// targets/a can be loaded now though because targets is loaded
	require.NoError(t, builder.Load("targets/a", meta["targets/a"], 1, false))

	// and now targets/a/b can be loaded because targets/a is loaded
	require.NoError(t, builder.Load("targets/a/b", meta["targets/a/b"], 1, false))

	valid, _, err := builder.Finish()
	require.True(t, valid.Root.Signatures[0].IsValid)
	require.True(t, valid.Timestamp.Signatures[0].IsValid)
	require.True(t, valid.Snapshot.Signatures[0].IsValid)
	require.True(t, valid.Targets[data.CanonicalTargetsRole].Signatures[0].IsValid)
	require.True(t, valid.Targets["targets/a"].Signatures[0].IsValid)
	require.True(t, valid.Targets["targets/a/b"].Signatures[0].IsValid)
	require.NoError(t, err)
}

func TestBuilderLoadInvalidDelegations(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	tufRepo, _, err := testutils.EmptyRepo(gun, "targets/a", "targets/a/b", "targets/b")
	require.NoError(t, err)

	meta, err := testutils.SignAndSerialize(tufRepo)
	require.NoError(t, err)

	builder := tuf.NewBuilderFromRepo(gun, tufRepo, trustpinning.TrustPinConfig{})

	// modify targets/a to remove the signature and update the snapshot
	// (we're not going to load the timestamp so no need to modify)
	targetsAJSON := meta["targets/a"]
	targetsA := data.Signed{}
	err = json.Unmarshal(targetsAJSON, &targetsA)
	require.NoError(t, err)
	targetsA.Signatures = make([]data.Signature, 0)
	targetsAJSON, err = json.Marshal(&targetsA)
	require.NoError(t, err)
	meta["targets/a"] = targetsAJSON
	delete(tufRepo.Targets, "targets/a")

	snap := tufRepo.Snapshot
	m, err := data.NewFileMeta(
		bytes.NewReader(targetsAJSON),
		"sha256", "sha512",
	)
	require.NoError(t, err)
	snap.AddMeta("targets/a", m)

	// load snapshot directly into repo to bypass signature check (we've invalidated
	// the signature by modifying it)
	tufRepo.Snapshot = snap

	// load targets/a
	require.Error(
		t,
		builder.Load(
			"targets/a",
			meta["targets/a"],
			1,
			false,
		),
	)

	_, invalid, err := builder.Finish()
	require.NoError(t, err)
	_, ok := invalid.Targets["targets/a"]
	require.True(t, ok)
}

func TestBuilderLoadInvalidDelegationsOldVersion(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	tufRepo, _, err := testutils.EmptyRepo(gun, "targets/a", "targets/a/b", "targets/b")
	require.NoError(t, err)

	meta, err := testutils.SignAndSerialize(tufRepo)
	require.NoError(t, err)

	builder := tuf.NewBuilderFromRepo(gun, tufRepo, trustpinning.TrustPinConfig{})
	delete(tufRepo.Targets, "targets/a")

	// load targets/a with high min-version so this one is too old
	err = builder.Load(
		"targets/a",
		meta["targets/a"],
		10,
		false,
	)
	require.Error(t, err)
	require.IsType(t, signed.ErrLowVersion{}, err)

	_, invalid, err := builder.Finish()
	require.NoError(t, err)
	_, ok := invalid.Targets["targets/a"]
	require.False(t, ok)
}

func TestBuilderAcceptRoleOnce(t *testing.T) {
	meta, gun := getSampleMeta(t)
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})

	for _, roleName := range append(data.BaseRoles, "targets/a", "targets/a/b") {
		// first time loading is ok
		require.NoError(t, builder.Load(roleName, meta[roleName], 1, false))
		require.True(t, builder.IsLoaded(roleName))
		require.Equal(t, 1, builder.GetLoadedVersion(roleName))

		// second time loading is not
		err := builder.Load(roleName, meta[roleName], 1, false)
		require.Error(t, err)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "has already been loaded")

		// still loaded
		require.True(t, builder.IsLoaded(roleName))
	}
}

func TestBuilderStopsAcceptingOrProducingDataOnceDone(t *testing.T) {
	meta, gun := getSampleMeta(t)
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})

	for _, roleName := range data.BaseRoles {
		require.NoError(t, builder.Load(roleName, meta[roleName], 1, false))
		require.True(t, builder.IsLoaded(roleName))
	}

	_, _, err := builder.Finish()
	require.NoError(t, err)

	err = builder.Load("targets/a", meta["targets/a"], 1, false)
	require.Error(t, err)
	require.Equal(t, tuf.ErrBuildDone, err)

	err = builder.LoadRootForUpdate(meta["root"], 1, true)
	require.Error(t, err)
	require.Equal(t, tuf.ErrBuildDone, err)

	// a new bootstrapped builder can also not have any more input output
	bootstrapped := builder.BootstrapNewBuilder()

	err = bootstrapped.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 0, false)
	require.Error(t, err)
	require.Equal(t, tuf.ErrBuildDone, err)

	for _, b := range []tuf.RepoBuilder{builder, bootstrapped} {
		_, _, err = b.Finish()
		require.Error(t, err)
		require.Equal(t, tuf.ErrBuildDone, err)

		_, _, err = b.GenerateSnapshot(nil)
		require.Error(t, err)
		require.Equal(t, tuf.ErrBuildDone, err)

		_, _, err = b.GenerateTimestamp(nil)
		require.Error(t, err)
		require.Equal(t, tuf.ErrBuildDone, err)

		for roleName := range meta {
			// a finished builder thinks nothing is loaded
			require.False(t, b.IsLoaded(roleName))
			// checksums are all empty, versions are all zero
			require.Equal(t, 0, b.GetLoadedVersion(roleName))
			require.Equal(t, tuf.ConsistentInfo{RoleName: roleName}, b.GetConsistentInfo(roleName))
		}

	}
}

// Test the cases in which GenerateSnapshot fails
func TestGenerateSnapshotInvalidOperations(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	// make snapshot have 2 keys and a threshold of 2
	snapKeys := make([]data.PublicKey, 2)
	for i := 0; i < 2; i++ {
		snapKeys[i], err = cs.Create(data.CanonicalSnapshotRole, gun, data.ECDSAKey)
		require.NoError(t, err)
	}

	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalSnapshotRole, snapKeys...))
	repo.Root.Signed.Roles[data.CanonicalSnapshotRole].Threshold = 2

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)

	for _, prevSnapshot := range []*data.SignedSnapshot{nil, repo.Snapshot} {
		// copy keys, since we expect one of these generation attempts to succeed and we do
		// some key deletion tests later
		newCS := testutils.CopyKeys(t, cs, data.CanonicalSnapshotRole)

		// --- we can't generate a snapshot if the root isn't loaded
		builder := tuf.NewRepoBuilder(gun, newCS, trustpinning.TrustPinConfig{})
		_, _, err := builder.GenerateSnapshot(prevSnapshot)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "root must be loaded first")
		require.False(t, builder.IsLoaded(data.CanonicalSnapshotRole))

		// --- we can't generate a snapshot if the targets isn't loaded and we have no previous snapshot,
		// --- but if we have a previous snapshot with a valid targets, we're good even if no snapshot
		// --- is loaded
		require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
		_, _, err = builder.GenerateSnapshot(prevSnapshot)
		if prevSnapshot == nil {
			require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
			require.Contains(t, err.Error(), "targets must be loaded first")
			require.False(t, builder.IsLoaded(data.CanonicalSnapshotRole))
		} else {
			require.NoError(t, err)
		}

		// --- we can't generate a snapshot if we've loaded the timestamp already
		builder = tuf.NewRepoBuilder(gun, newCS, trustpinning.TrustPinConfig{})
		require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
		if prevSnapshot == nil {
			require.NoError(t, builder.Load(data.CanonicalTargetsRole, meta[data.CanonicalTargetsRole], 1, false))
		}
		require.NoError(t, builder.Load(data.CanonicalTimestampRole, meta[data.CanonicalTimestampRole], 1, false))

		_, _, err = builder.GenerateSnapshot(prevSnapshot)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "cannot generate snapshot if timestamp has already been loaded")
		require.False(t, builder.IsLoaded(data.CanonicalSnapshotRole))

		// --- we cannot generate a snapshot if we've already loaded a snapshot
		builder = tuf.NewRepoBuilder(gun, newCS, trustpinning.TrustPinConfig{})
		require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
		if prevSnapshot == nil {
			require.NoError(t, builder.Load(data.CanonicalTargetsRole, meta[data.CanonicalTargetsRole], 1, false))
		}
		require.NoError(t, builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false))

		_, _, err = builder.GenerateSnapshot(prevSnapshot)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "snapshot has already been loaded")

		// --- we cannot generate a snapshot if we can't satisfy the role threshold
		for i := 0; i < len(snapKeys); i++ {
			require.NoError(t, newCS.RemoveKey(snapKeys[i].ID()))
			builder = tuf.NewRepoBuilder(gun, newCS, trustpinning.TrustPinConfig{})
			require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
			if prevSnapshot == nil {
				require.NoError(t, builder.Load(data.CanonicalTargetsRole, meta[data.CanonicalTargetsRole], 1, false))
			}

			_, _, err = builder.GenerateSnapshot(prevSnapshot)
			require.IsType(t, signed.ErrInsufficientSignatures{}, err)
			require.False(t, builder.IsLoaded(data.CanonicalSnapshotRole))
		}

		// --- we cannot generate a snapshot if we don't have a cryptoservice
		builder = tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
		require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
		if prevSnapshot == nil {
			require.NoError(t, builder.Load(data.CanonicalTargetsRole, meta[data.CanonicalTargetsRole], 1, false))
		}

		_, _, err = builder.GenerateSnapshot(prevSnapshot)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "cannot generate snapshot without a cryptoservice")
		require.False(t, builder.IsLoaded(data.CanonicalSnapshotRole))
	}

	// --- we can't generate a snapshot if we're given an invalid previous snapshot (for instance, an empty one),
	// --- even if we have a targets loaded
	builder := tuf.NewRepoBuilder(gun, cs, trustpinning.TrustPinConfig{})
	require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
	require.NoError(t, builder.Load(data.CanonicalTargetsRole, meta[data.CanonicalTargetsRole], 1, false))

	_, _, err = builder.GenerateSnapshot(&data.SignedSnapshot{})
	require.IsType(t, data.ErrInvalidMetadata{}, err)
	require.False(t, builder.IsLoaded(data.CanonicalSnapshotRole))
}

// Test the cases in which GenerateTimestamp fails
func TestGenerateTimestampInvalidOperations(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, cs, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	// make timsetamp have 2 keys and a threshold of 2
	tsKeys := make([]data.PublicKey, 2)
	for i := 0; i < 2; i++ {
		tsKeys[i], err = cs.Create(data.CanonicalTimestampRole, gun, data.ECDSAKey)
		require.NoError(t, err)
	}

	require.NoError(t, repo.ReplaceBaseKeys(data.CanonicalTimestampRole, tsKeys...))
	repo.Root.Signed.Roles[data.CanonicalTimestampRole].Threshold = 2

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)

	for _, prevTimestamp := range []*data.SignedTimestamp{nil, repo.Timestamp} {
		// --- we can't generate a timestamp if the root isn't loaded
		builder := tuf.NewRepoBuilder(gun, cs, trustpinning.TrustPinConfig{})
		_, _, err := builder.GenerateTimestamp(prevTimestamp)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "root must be loaded first")
		require.False(t, builder.IsLoaded(data.CanonicalTimestampRole))

		// --- we can't generate a timestamp if the snapshot isn't loaded, no matter if we have a previous
		// --- timestamp or not
		require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
		_, _, err = builder.GenerateTimestamp(prevTimestamp)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "snapshot must be loaded first")
		require.False(t, builder.IsLoaded(data.CanonicalTimestampRole))

		// --- we can't generate a timestamp if we've loaded the timestamp already
		builder = tuf.NewRepoBuilder(gun, cs, trustpinning.TrustPinConfig{})
		require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
		require.NoError(t, builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false))
		require.NoError(t, builder.Load(data.CanonicalTimestampRole, meta[data.CanonicalTimestampRole], 1, false))

		_, _, err = builder.GenerateTimestamp(prevTimestamp)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "timestamp has already been loaded")

		// --- we cannot generate a timestamp if we can't satisfy the role threshold
		for i := 0; i < len(tsKeys); i++ {
			require.NoError(t, cs.RemoveKey(tsKeys[i].ID()))
			builder = tuf.NewRepoBuilder(gun, cs, trustpinning.TrustPinConfig{})
			require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
			require.NoError(t, builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false))

			_, _, err = builder.GenerateTimestamp(prevTimestamp)
			require.IsType(t, signed.ErrInsufficientSignatures{}, err)
			require.False(t, builder.IsLoaded(data.CanonicalTimestampRole))
		}

		// --- we cannot generate a timestamp if we don't have a cryptoservice
		builder = tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
		require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
		require.NoError(t, builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false))

		_, _, err = builder.GenerateTimestamp(prevTimestamp)
		require.IsType(t, tuf.ErrInvalidBuilderInput{}, err)
		require.Contains(t, err.Error(), "cannot generate timestamp without a cryptoservice")
		require.False(t, builder.IsLoaded(data.CanonicalTimestampRole))
	}

	// --- we can't generate a timsetamp if we're given an invalid previous timestamp (for instance, an empty one),
	// --- even if we have a snapshot loaded
	builder := tuf.NewRepoBuilder(gun, cs, trustpinning.TrustPinConfig{})
	require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
	require.NoError(t, builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false))

	_, _, err = builder.GenerateTimestamp(&data.SignedTimestamp{})
	require.IsType(t, data.ErrInvalidMetadata{}, err)
	require.False(t, builder.IsLoaded(data.CanonicalTimestampRole))
}

func TestGetConsistentInfo(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, _, err := testutils.EmptyRepo(gun)
	require.NoError(t, err)

	// add some hashes for items in the snapshot that don't correspond to real metadata, but that
	// will cause ConsistentInfo to behave differently
	realSha512Sum := sha512.Sum512([]byte("stuff"))
	repo.Snapshot.Signed.Meta["only512"] = data.FileMeta{Hashes: data.Hashes{notary.SHA512: realSha512Sum[:]}}
	repo.Snapshot.Signed.Meta["targets/random"] = data.FileMeta{Hashes: data.Hashes{"randomsha": []byte("12345")}}
	repo.Snapshot.Signed.Meta["targets/nohashes"] = data.FileMeta{Length: 1}

	extraMeta := []data.RoleName{"only512", "targets/random", "targets/nohashes"}

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)
	metadata := data.MetadataRoleMapToStringMap(meta)

	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	checkTimestampSnapshotRequired(t, metadata, extraMeta, builder)
	checkOnlySnapshotConsistentAfterTimestamp(t, repo, metadata, extraMeta, builder)
	checkOtherRolesConsistentAfterSnapshot(t, repo, metadata, builder)

	// the fake roles have invalid-ish checksums: the ConsistentInfos for those will return
	// non-consistent names but non -1 sizes
	for _, checkName := range extraMeta {
		ci := builder.GetConsistentInfo(checkName)
		require.EqualValues(t, checkName.String(), ci.ConsistentName()) // because no sha256 hash
		require.True(t, ci.ChecksumKnown())
		require.True(t, ci.Length() > -1)
	}

	// a non-existent role's ConsistentInfo is empty
	ci := builder.GetConsistentInfo("nonExistent")
	require.EqualValues(t, "nonExistent", ci.ConsistentName())
	require.False(t, ci.ChecksumKnown())
	require.Equal(t, int64(-1), ci.Length())

	// when we bootstrap a new builder, the root has consistent info because the checksum is provided,
	// but nothing else does
	builder = builder.BootstrapNewBuilder()
	for _, checkName := range append(data.BaseRoles, extraMeta...) {
		ci := builder.GetConsistentInfo(checkName)

		switch checkName {
		case data.CanonicalTimestampRole:
			// timestamp's size is always the max timestamp size
			require.EqualValues(t, checkName.String(), ci.ConsistentName())
			require.True(t, ci.ChecksumKnown())
			require.Equal(t, notary.MaxTimestampSize, ci.Length())

		case data.CanonicalRootRole:
			cName := utils.ConsistentName(data.CanonicalRootRole.String(),
				repo.Snapshot.Signed.Meta[data.CanonicalRootRole.String()].Hashes[notary.SHA256])

			require.EqualValues(t, cName, ci.ConsistentName())
			require.True(t, ci.ChecksumKnown())
			require.True(t, ci.Length() > -1)

		default:
			require.EqualValues(t, checkName.String(), ci.ConsistentName())
			require.False(t, ci.ChecksumKnown())
			require.Equal(t, int64(-1), ci.Length())
		}
	}
}

func checkTimestampSnapshotRequired(t *testing.T, meta map[string][]byte, extraMeta []data.RoleName, builder tuf.RepoBuilder) {
	// if neither snapshot nor timestamp are loaded, no matter how much other data is loaded, consistent info
	// is empty except for timestamp: timestamps have no checksums, and the length is always -1
	for _, roleToLoad := range []data.RoleName{data.CanonicalRootRole, data.CanonicalTargetsRole} {
		require.NoError(t, builder.Load(roleToLoad, meta[roleToLoad.String()], 1, false))
		for _, checkName := range append(data.BaseRoles, extraMeta...) {
			ci := builder.GetConsistentInfo(checkName)
			require.EqualValues(t, checkName, ci.ConsistentName())

			switch checkName {
			case data.CanonicalTimestampRole:
				// timestamp's size is always the max timestamp size
				require.True(t, ci.ChecksumKnown())
				require.Equal(t, notary.MaxTimestampSize, ci.Length())
			default:
				require.False(t, ci.ChecksumKnown())
				require.Equal(t, int64(-1), ci.Length())
			}
		}
	}
}

func checkOnlySnapshotConsistentAfterTimestamp(t *testing.T, repo *tuf.Repo, meta map[string][]byte, extraMeta []data.RoleName, builder tuf.RepoBuilder) {
	// once timestamp is loaded, we can get the consistent info for snapshot but nothing else
	require.NoError(t, builder.Load(data.CanonicalTimestampRole, meta[data.CanonicalTimestampRole.String()], 1, false))
	for _, checkName := range append(data.BaseRoles, extraMeta...) {
		ci := builder.GetConsistentInfo(checkName)

		switch checkName {
		case data.CanonicalSnapshotRole:
			cName := utils.ConsistentName(data.CanonicalSnapshotRole.String(),
				repo.Timestamp.Signed.Meta[data.CanonicalSnapshotRole.String()].Hashes[notary.SHA256])
			require.EqualValues(t, cName, ci.ConsistentName())
			require.True(t, ci.ChecksumKnown())
			require.True(t, ci.Length() > -1)
		case data.CanonicalTimestampRole:
			// timestamp's canonical name is always "timestamp" and its size is always the max
			// timestamp size
			require.EqualValues(t, data.CanonicalTimestampRole, ci.ConsistentName())
			require.True(t, ci.ChecksumKnown())
			require.Equal(t, notary.MaxTimestampSize, ci.Length())
		default:
			require.EqualValues(t, checkName, ci.ConsistentName())
			require.False(t, ci.ChecksumKnown())
			require.Equal(t, int64(-1), ci.Length())
		}
	}
}

func checkOtherRolesConsistentAfterSnapshot(t *testing.T, repo *tuf.Repo, meta map[string][]byte, builder tuf.RepoBuilder) {
	// once the snapshot is loaded, we can get real consistent info for all loaded roles
	require.NoError(t, builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole.String()], 1, false))
	for _, checkName := range data.BaseRoles {
		ci := builder.GetConsistentInfo(checkName)
		require.True(t, ci.ChecksumKnown(), "%s's checksum is not known", checkName)

		switch checkName {
		case data.CanonicalTimestampRole:
			// timestamp's canonical name is always "timestamp" and its size is always -1
			require.EqualValues(t, data.CanonicalTimestampRole, ci.ConsistentName())
			require.Equal(t, notary.MaxTimestampSize, ci.Length())
		default:
			fileInfo := repo.Snapshot.Signed.Meta
			if checkName == data.CanonicalSnapshotRole {
				fileInfo = repo.Timestamp.Signed.Meta
			}

			cName := utils.ConsistentName(checkName.String(), fileInfo[checkName.String()].Hashes[notary.SHA256])
			require.EqualValues(t, cName, ci.ConsistentName())
			require.True(t, ci.Length() > -1)
		}
	}
}

// No matter what order timestamp and snapshot is loaded, if the snapshot's checksum doesn't match
// what's in the timestamp, the builder will error and refuse to load the latest piece of metadata
// whether that is snapshot (because it was loaded after timestamp) or timestamp (because builder
// retroactive checks the loaded snapshot's checksum).  Timestamp ONLY checks the snapshot checksum.
func TestTimestampPreAndPostChecksumming(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	repo, _, err := testutils.EmptyRepo(gun, "targets/other", "targets/other/other")
	require.NoError(t, err)

	// add invalid checkums for all the other roles to timestamp too, and show that
	// cached items aren't checksummed against this
	fakeChecksum, err := data.NewFileMeta(bytes.NewBuffer([]byte("fake")), notary.SHA256, notary.SHA512)
	require.NoError(t, err)
	for _, roleName := range append(data.BaseRoles, "targets/other") {
		// add a wrong checksum for every role, including timestamp itself
		repo.Timestamp.Signed.Meta[roleName.String()] = fakeChecksum
	}
	// this will overwrite the snapshot checksum with the right one
	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)
	// ensure that the fake meta for other roles weren't destroyed by signing the timestamp
	require.Len(t, repo.Timestamp.Signed.Meta, 5)

	snapJSON := append(meta[data.CanonicalSnapshotRole], ' ')

	// --- load timestamp first
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
	// timestamp doesn't fail, even though its checksum for root is wrong according to timestamp
	require.NoError(t, builder.Load(data.CanonicalTimestampRole, meta[data.CanonicalTimestampRole], 1, false))
	// loading the snapshot in fails, because of the checksum the timestamp has
	err = builder.Load(data.CanonicalSnapshotRole, snapJSON, 1, false)
	require.Error(t, err)
	require.IsType(t, data.ErrMismatchedChecksum{}, err)
	require.True(t, builder.IsLoaded(data.CanonicalTimestampRole))
	require.False(t, builder.IsLoaded(data.CanonicalSnapshotRole))
	// all the other metadata can be loaded in, even though the checksums are wrong according to timestamp
	for _, roleName := range []data.RoleName{data.CanonicalTargetsRole, "targets/other"} {
		require.NoError(t, builder.Load(roleName, meta[roleName], 1, false))
	}

	// --- load snapshot first
	builder = tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	for _, roleName := range append(data.BaseRoles, "targets/other") {
		switch roleName {
		case data.CanonicalTimestampRole:
			continue
		case data.CanonicalSnapshotRole:
			require.NoError(t, builder.Load(roleName, snapJSON, 1, false))
		default:
			require.NoError(t, builder.Load(roleName, meta[roleName], 1, false))
		}
	}
	// timestamp fails because the snapshot checksum is wrong
	err = builder.Load(data.CanonicalTimestampRole, meta[data.CanonicalTimestampRole], 1, false)
	require.Error(t, err)
	checksumErr, ok := err.(data.ErrMismatchedChecksum)
	require.True(t, ok)
	require.Contains(t, checksumErr.Error(), "checksum for snapshot did not match")
	require.False(t, builder.IsLoaded(data.CanonicalTimestampRole))
	require.True(t, builder.IsLoaded(data.CanonicalSnapshotRole))
}

// Creates metadata in the following manner:
// - the snapshot has bad checksums for itself and for timestamp, to show that those aren't checked
// - snapshot has valid checksums for root, targets, and targets/other
// - snapshot doesn't have a checksum for targets/other/other, but targets/other/other is a valid
//   delegation role in targets/other and there is metadata for targets/other/other that is correctly
//   signed
func setupSnapshotChecksumming(t *testing.T, gun data.GUN) map[data.RoleName][]byte {
	repo, _, err := testutils.EmptyRepo(gun, "targets/other", "targets/other/other")
	require.NoError(t, err)

	// add invalid checkums for all the other roles to timestamp too, and show that
	// cached items aren't checksummed against this
	fakeChecksum, err := data.NewFileMeta(bytes.NewBuffer([]byte("fake")), notary.SHA256, notary.SHA512)
	require.NoError(t, err)
	// fake the snapshot and timestamp checksums
	repo.Snapshot.Signed.Meta[data.CanonicalSnapshotRole.String()] = fakeChecksum
	repo.Snapshot.Signed.Meta[data.CanonicalTimestampRole.String()] = fakeChecksum

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)
	// ensure that the fake metadata for other roles wasn't destroyed by signing
	require.Len(t, repo.Snapshot.Signed.Meta, 5)

	// create delegation metadata that should not be in snapshot, but has a valid role and signature
	_, err = repo.InitTargets("targets/other/other")
	require.NoError(t, err)
	s, err := repo.SignTargets("targets/other/other", data.DefaultExpires(data.CanonicalTargetsRole))
	require.NoError(t, err)
	meta["targets/other/other"], err = json.Marshal(s)
	require.NoError(t, err)

	return meta
}

// If the snapshot is loaded first (-ish, because really root has to be loaded first)
// it will be used to validate the checksums of all other metadata that gets loaded.
// If the checksum doesn't match, or if there is no checksum, then the other metadata
// cannot be loaded.
func TestSnapshotLoadedFirstChecksumsOthers(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	meta := setupSnapshotChecksumming(t, gun)
	// --- load root then snapshot
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	require.NoError(t, builder.Load(data.CanonicalRootRole, meta[data.CanonicalRootRole], 1, false))
	require.NoError(t, builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false))

	// loading timestamp is fine, even though the timestamp metadata has the wrong checksum because
	// we don't check timestamp checksums
	require.NoError(t, builder.Load(data.CanonicalTimestampRole, meta[data.CanonicalTimestampRole], 1, false))

	// loading the other roles' metadata with a space will fail because of a checksum failure (builder
	// checks right away if the snapshot is loaded) - in the case of targets/other/other, which should
	// not be in snapshot at all, loading should fail even without a space because there is no checksum
	// for it
	for _, roleNameToLoad := range []data.RoleName{data.CanonicalTargetsRole, "targets/other"} {
		err := builder.Load(roleNameToLoad, append(meta[roleNameToLoad], ' '), 0, false)
		require.Error(t, err)
		checksumErr, ok := err.(data.ErrMismatchedChecksum)
		require.True(t, ok)
		require.Contains(t, checksumErr.Error(), fmt.Sprintf("checksum for %s did not match", roleNameToLoad))
		require.False(t, builder.IsLoaded(roleNameToLoad))

		// now load it for real (since we need targets loaded before trying to load "targets/other")
		require.NoError(t, builder.Load(roleNameToLoad, meta[roleNameToLoad], 1, false))
	}
	// loading the non-existent role wil fail
	err := builder.Load("targets/other/other", meta["targets/other/other"], 1, false)
	require.Error(t, err)
	require.IsType(t, data.ErrMissingMeta{}, err)
	require.False(t, builder.IsLoaded("targets/other/other"))
}

// If any other metadata is loaded first, when the snapshot is loaded it will retroactively go back
// and validate that metadata.  If anything fails to validate, or there is metadata for which this
// snapshot has no checksums for, the snapshot will fail to validate.
func TestSnapshotLoadedAfterChecksumsOthersRetroactively(t *testing.T) {
	var gun data.GUN = "docker.com/notary"
	meta := setupSnapshotChecksumming(t, gun)

	// --- load all the other metadata first, but with an extra space at the end which should
	// --- validate fine, except for the checksum.
	for _, roleNameToPermute := range append(data.BaseRoles, "targets/other") {
		builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
		if roleNameToPermute == data.CanonicalSnapshotRole {
			continue
		}
		// load all the roles normally, except for roleToPermute, which has one space added
		// to the end, thus changing the checksum
		for _, roleNameToLoad := range append(data.BaseRoles, "targets/other") {
			switch roleNameToLoad {
			case data.CanonicalSnapshotRole:
				continue // we load this later
			case roleNameToPermute:
				// having a space added at the end should not affect any validity check except checksum
				require.NoError(t, builder.Load(roleNameToLoad, append(meta[roleNameToLoad], ' '), 0, false))
			default:
				require.NoError(t, builder.Load(roleNameToLoad, meta[roleNameToLoad], 1, false))
			}
			require.True(t, builder.IsLoaded(roleNameToLoad))
		}
		// now load the snapshot - it should fail with the checksum failure for the permuted role
		err := builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false)
		switch roleNameToPermute {
		case data.CanonicalTimestampRole:
			require.NoError(t, err) // we don't check the timestamp's checksum
		default:
			require.Error(t, err)
			checksumErr, ok := err.(data.ErrMismatchedChecksum)
			require.True(t, ok)
			require.Contains(t, checksumErr.Error(), fmt.Sprintf("checksum for %s did not match", roleNameToPermute))
			require.False(t, builder.IsLoaded(data.CanonicalSnapshotRole))
		}
	}

	// load all the metadata as is without alteration (so they should validate all checksums)
	// but also load the metadata that is not contained in the snapshot.  Then when the snapshot
	// is loaded it will fail validation, because it doesn't have target/other/other's checksum
	builder := tuf.NewRepoBuilder(gun, nil, trustpinning.TrustPinConfig{})
	for _, roleNameToLoad := range append(data.BaseRoles, "targets/other", "targets/other/other") {
		if roleNameToLoad == data.CanonicalSnapshotRole {
			continue
		}
		require.NoError(t, builder.Load(roleNameToLoad, meta[roleNameToLoad], 1, false))
	}
	err := builder.Load(data.CanonicalSnapshotRole, meta[data.CanonicalSnapshotRole], 1, false)
	require.Error(t, err)
	require.IsType(t, data.ErrMissingMeta{}, err)
}
