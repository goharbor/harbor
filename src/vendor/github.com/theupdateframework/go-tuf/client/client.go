package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"

	"github.com/theupdateframework/go-tuf/data"
	"github.com/theupdateframework/go-tuf/util"
	"github.com/theupdateframework/go-tuf/verify"
)

const (
	// This is the upper limit in bytes we will use to limit the download
	// size of the root/timestamp roles, since we might not don't know how
	// big it is.
	defaultRootDownloadLimit      = 512000
	defaultTimestampDownloadLimit = 16384
	defaultMaxDelegations         = 32
	defaultMaxRootRotations       = 1e3
)

// LocalStore is local storage for downloaded top-level metadata.
type LocalStore interface {
	io.Closer

	// GetMeta returns top-level metadata from local storage. The keys are
	// in the form `ROLE.json`, with ROLE being a valid top-level role.
	GetMeta() (map[string]json.RawMessage, error)

	// SetMeta persists the given top-level metadata in local storage, the
	// name taking the same format as the keys returned by GetMeta.
	SetMeta(name string, meta json.RawMessage) error

	// DeleteMeta deletes a given metadata.
	DeleteMeta(name string) error
}

// RemoteStore downloads top-level metadata and target files from a remote
// repository.
type RemoteStore interface {
	// GetMeta downloads the given metadata from remote storage.
	//
	// `name` is the filename of the metadata (e.g. "root.json")
	//
	// `err` is ErrNotFound if the given file does not exist.
	//
	// `size` is the size of the stream, -1 indicating an unknown length.
	GetMeta(name string) (stream io.ReadCloser, size int64, err error)

	// GetTarget downloads the given target file from remote storage.
	//
	// `path` is the path of the file relative to the root of the remote
	//        targets directory (e.g. "/path/to/file.txt").
	//
	// `err` is ErrNotFound if the given file does not exist.
	//
	// `size` is the size of the stream, -1 indicating an unknown length.
	GetTarget(path string) (stream io.ReadCloser, size int64, err error)
}

// Client provides methods for fetching updates from a remote repository and
// downloading remote target files.
type Client struct {
	local  LocalStore
	remote RemoteStore

	// The following four fields represent the versions of metatdata either
	// from local storage or from recently downloaded metadata
	rootVer      int64
	targetsVer   int64
	snapshotVer  int64
	timestampVer int64

	// targets is the list of available targets, either from local storage
	// or from recently downloaded targets metadata
	targets data.TargetFiles

	// localMeta is the raw metadata from local storage and is used to
	// check whether remote metadata is present locally
	localMeta map[string]json.RawMessage

	// db is a key DB used for verifying metadata
	db *verify.DB

	// consistentSnapshot indicates whether the remote storage is using
	// consistent snapshots (as specified in root.json)
	consistentSnapshot bool

	// MaxDelegations limits by default the number of delegations visited for any
	// target
	MaxDelegations int

	// MaxRootRotations limits the number of downloaded roots in 1.0.19 root updater
	MaxRootRotations int
}

func NewClient(local LocalStore, remote RemoteStore) *Client {
	return &Client{
		local:            local,
		remote:           remote,
		MaxDelegations:   defaultMaxDelegations,
		MaxRootRotations: defaultMaxRootRotations,
	}
}

// Init initializes a local repository from root metadata.
//
// The root's keys are extracted from the root and saved in local storage.
// Root expiration is not checked.
// It is expected that rootJSON was securely distributed with the software
// being updated.
func (c *Client) Init(rootJSON []byte) error {
	err := c.loadAndVerifyRootMeta(rootJSON, true /*ignoreExpiredCheck*/)
	if err != nil {
		return err
	}
	return c.local.SetMeta("root.json", rootJSON)
}

// Update downloads and verifies remote metadata and returns updated targets.
// It always performs root update (5.2 and 5.3) section of the v1.0.19 spec.
//
// https://theupdateframework.github.io/specification/v1.0.19/index.html#load-trusted-root
func (c *Client) Update() (data.TargetFiles, error) {
	if err := c.UpdateRoots(); err != nil {
		if _, ok := err.(verify.ErrExpired); ok {
			// For backward compatibility, we wrap the ErrExpired inside
			// ErrDecodeFailed.
			return nil, ErrDecodeFailed{"root.json", err}
		}
		return nil, err
	}

	// Load trusted metadata files, if any, and verify them against the latest root
	c.getLocalMeta()

	// 5.4.1 - Download the timestamp metadata
	timestampJSON, err := c.downloadMetaUnsafe("timestamp.json", defaultTimestampDownloadLimit)
	if err != nil {
		return nil, err
	}
	// 5.4.(2,3 and 4) - Verify timestamp against various attacks
	// Returns the extracted snapshot metadata
	snapshotMeta, sameTimestampVersion, err := c.decodeTimestamp(timestampJSON)
	if sameTimestampVersion {
		// The new timestamp.json file had the same version; we don't need to
		// update, so bail early.
		return c.targets, nil
	}

	if err != nil {
		return nil, err
	}
	// 5.4.5 - Persist the timestamp metadata
	if err := c.local.SetMeta("timestamp.json", timestampJSON); err != nil {
		return nil, err
	}

	// 5.5.1 - Download snapshot metadata
	// 5.5.2 and 5.5.4 - Check against timestamp role's snapshot hash and version
	snapshotJSON, err := c.downloadMetaFromTimestamp("snapshot.json", snapshotMeta)
	if err != nil {
		return nil, err
	}
	// 5.5.(3,5 and 6) - Verify snapshot against various attacks
	// Returns the extracted metadata files
	snapshotMetas, err := c.decodeSnapshot(snapshotJSON)
	if err != nil {
		return nil, err
	}
	// 5.5.7 - Persist snapshot metadata
	if err := c.local.SetMeta("snapshot.json", snapshotJSON); err != nil {
		return nil, err
	}

	// If we don't have the targets.json, download it, determine updated
	// targets and save targets.json in local storage
	var updatedTargets data.TargetFiles
	targetsMeta := snapshotMetas["targets.json"]
	if !c.hasMetaFromSnapshot("targets.json", targetsMeta) {
		// 5.6.1 - Download the top-level targets metadata file
		// 5.6.2 and 5.6.4 - Check against snapshot role's targets hash and version
		targetsJSON, err := c.downloadMetaFromSnapshot("targets.json", targetsMeta)
		if err != nil {
			return nil, err
		}
		// 5.6.(3 and 5) - Verify signatures and check against freeze attack
		updatedTargets, err = c.decodeTargets(targetsJSON)
		if err != nil {
			return nil, err
		}
		// 5.6.6 - Persist targets metadata
		if err := c.local.SetMeta("targets.json", targetsJSON); err != nil {
			return nil, err
		}
	}

	return updatedTargets, nil
}

func (c *Client) UpdateRoots() error {
	// https://theupdateframework.github.io/specification/v1.0.19/index.html#load-trusted-root
	// 5.2 Load the trusted root metadata file. We assume that a good,
	// trusted copy of this file was shipped with the package manager
	// or software updater using an out-of-band process.
	if err := c.loadAndVerifyLocalRootMeta( /*ignoreExpiredCheck=*/ true); err != nil {
		return err
	}
	m, err := c.local.GetMeta()
	if err != nil {
		return err
	}

	type KeyInfo struct {
		KeyIDs    map[string]bool
		Threshold int
	}

	// Prepare for 5.3.11: If the timestamp and / or snapshot keys have been rotated,
	// then delete the trusted timestamp and snapshot metadata files.
	getKeyInfo := func(role string) KeyInfo {
		keyIDs := make(map[string]bool)
		for k := range c.db.GetRole(role).KeyIDs {
			keyIDs[k] = true
		}
		return KeyInfo{keyIDs, c.db.GetRole(role).Threshold}
	}

	// The nonRootKeyInfo looks like this:
	// {
	//	"timestamp": {KeyIDs={"KEYID1": true, "KEYID2": true}, Threshold=2},
	//	"snapshot": {KeyIDs={"KEYID3": true}, Threshold=1},
	//	"targets": {KeyIDs={"KEYID4": true, "KEYID5": true, "KEYID6": true}, Threshold=1}
	// }

	nonRootKeyInfo := map[string]KeyInfo{"timestamp": {}, "snapshot": {}, "targets": {}}
	for k := range nonRootKeyInfo {
		nonRootKeyInfo[k] = getKeyInfo(k)
	}

	// 5.3.1 Temorarily turn on the consistent snapshots in order to download
	// versioned root metadata files as described next.
	consistentSnapshot := c.consistentSnapshot
	c.consistentSnapshot = true

	nRootMetadata := m["root.json"]

	// https://theupdateframework.github.io/specification/v1.0.19/index.html#update-root

	// 5.3.1 Since it may now be signed using entirely different keys,
	// the client MUST somehow be able to establish a trusted line of
	// continuity to the latest set of keys (see § 6.1 Key
	// management and migration). To do so, the client MUST
	// download intermediate root metadata files, until the
	// latest available one is reached. Therefore, it MUST
	// temporarily turn on consistent snapshots in order to
	// download versioned root metadata files as described next.

	// This loop returns on error or breaks after downloading the lastest root metadata.
	// 5.3.2 Let N denote the version number of the trusted root metadata file.
	for i := 0; i < c.MaxRootRotations; i++ {
		// 5.3.3 Try downloading version nPlusOne of the root metadata file.
		// NOTE: as a side effect, we do update c.rootVer to nPlusOne between iterations.
		nPlusOne := c.rootVer + 1
		nPlusOneRootPath := util.VersionedPath("root.json", nPlusOne)
		nPlusOneRootMetadata, err := c.downloadMetaUnsafe(nPlusOneRootPath, defaultRootDownloadLimit)

		if err != nil {
			if _, ok := err.(ErrMissingRemoteMetadata); ok {
				// stop when the next root can't be downloaded
				break
			}
			return err
		}

		// 5.3.4 Check for an arbitrary software attack.
		// 5.3.4.1 Check that N signed N+1
		nPlusOneRootMetadataSigned, err := c.verifyRoot(nRootMetadata, nPlusOneRootMetadata)
		if err != nil {
			return err
		}

		// 5.3.4.2 check that N+1 signed itself.
		if _, err := c.verifyRoot(nPlusOneRootMetadata, nPlusOneRootMetadata); err != nil {
			// 5.3.6 Note that the expiration of the new (intermediate) root
			// metadata file does not matter yet, because we will check for
			// it in step 5.3.10.
			return err
		}

		// 5.3.5 Check for a rollback attack. Here, we check that nPlusOneRootMetadataSigned.version == nPlusOne.
		if nPlusOneRootMetadataSigned.Version != nPlusOne {
			return verify.ErrWrongVersion{
				Given:    nPlusOneRootMetadataSigned.Version,
				Expected: nPlusOne,
			}
		}

		// 5.3.7 Set the trusted root metadata file to the new root metadata file.
		c.rootVer = nPlusOneRootMetadataSigned.Version
		// NOTE: following up on 5.3.1, we want to always have consistent snapshots on for the duration
		// of root rotation. AFTER the rotation is over, we will set it to the value of the last root.
		consistentSnapshot = nPlusOneRootMetadataSigned.ConsistentSnapshot
		// 5.3.8 Persist root metadata. The client MUST write the file to non-volatile storage as FILENAME.EXT (e.g. root.json).
		// NOTE: Internally, setMeta stores metadata in LevelDB in a persistent manner.
		if err := c.local.SetMeta("root.json", nPlusOneRootMetadata); err != nil {
			return err
		}
		nRootMetadata = nPlusOneRootMetadata
		// 5.3.9 Repeat steps 5.3.2 to 5.3.9

	} // End of the for loop.

	// 5.3.10 Check for a freeze attack.
	// NOTE: This will check for any, including freeze, attack.
	if err := c.loadAndVerifyLocalRootMeta( /*ignoreExpiredCheck=*/ false); err != nil {
		return err
	}

	countDeleted := func(s1 map[string]bool, s2 map[string]bool) int {
		c := 0
		for k := range s1 {
			if _, ok := s2[k]; !ok {
				c++
			}
		}
		return c
	}

	// 5.3.11 To recover from fast-forward attack, certain metadata files need
	// to be deleted if a threshold of keys are revoked.
	// List of metadata that should be deleted per role if a threshold of keys
	// are revoked:
	// (based on the ongoing PR: https://github.com/mnm678/specification/tree/e50151d9df632299ddea364c4f44fe8ca9c10184)
	// timestamp -> delete timestamp.json
	// snapshot ->  delete timestamp.json and snapshot.json
	// targets ->   delete snapshot.json and targets.json
	//
	// nonRootKeyInfo contains the keys and thresholds from root.json
	// that were on disk before the root update process begins.
	for topLevelRolename := range nonRootKeyInfo {
		// ki contains the keys and thresholds from the latest downloaded root.json.
		ki := getKeyInfo(topLevelRolename)
		if countDeleted(nonRootKeyInfo[topLevelRolename].KeyIDs, ki.KeyIDs) >= nonRootKeyInfo[topLevelRolename].Threshold {
			deleteMeta := map[string][]string{
				"timestamp": {"timestamp.json"},
				"snapshot":  {"timestamp.json", "snapshot.json"},
				"targets":   {"snapshot.json", "targets.json"},
			}

			for _, r := range deleteMeta[topLevelRolename] {
				c.local.DeleteMeta(r)
			}
		}
	}

	// 5.3.12 Set whether consistent snapshots are used as per the trusted root metadata file.
	c.consistentSnapshot = consistentSnapshot
	return nil
}

// getLocalMeta decodes and verifies metadata from local storage.
// The verification of local files is purely for consistency, if an attacker
// has compromised the local storage, there is no guarantee it can be trusted.
// Before trying to load the metadata files, it clears the in-memory copy of the local metadata.
// This is to insure that all of the loaded metadata files at the end are indeed verified by the latest root.
// If some of the metadata files fail to load it will proceed with trying to load the rest,
// but still return an error at the end, if such occurred. Otherwise returns nil.
func (c *Client) getLocalMeta() error {
	var retErr error
	loadFailed := false
	// Clear the in-memory copy of the local metadata. The goal is to reload and take into account
	// only the metadata files that are verified by the latest root. Otherwise, their content should
	// be ignored.
	c.localMeta = make(map[string]json.RawMessage)

	// Load the latest root meta
	if err := c.loadAndVerifyLocalRootMeta( /*ignoreExpiredCheck=*/ false); err != nil {
		return err
	}

	// Load into memory the existing meta, if any, from the local storage
	meta, err := c.local.GetMeta()
	if err != nil {
		return nil
	}

	// Verify the top-level metadata (timestamp, snapshot and targets) against the latest root and load it, if okay
	if timestampJSON, ok := meta["timestamp.json"]; ok {
		timestamp := &data.Timestamp{}
		if err := c.db.UnmarshalTrusted(timestampJSON, timestamp, "timestamp"); err != nil {
			loadFailed = true
			retErr = err
		} else {
			c.localMeta["timestamp.json"] = meta["timestamp.json"]
			c.timestampVer = timestamp.Version
		}
	}

	if snapshotJSON, ok := meta["snapshot.json"]; ok {
		snapshot := &data.Snapshot{}
		if err := c.db.UnmarshalTrusted(snapshotJSON, snapshot, "snapshot"); err != nil {
			loadFailed = true
			retErr = err
		} else {
			c.localMeta["snapshot.json"] = meta["snapshot.json"]
			c.snapshotVer = snapshot.Version
		}
	}

	if targetsJSON, ok := meta["targets.json"]; ok {
		targets := &data.Targets{}
		if err := c.db.UnmarshalTrusted(targetsJSON, targets, "targets"); err != nil {
			loadFailed = true
			retErr = err
		} else {
			c.localMeta["targets.json"] = meta["targets.json"]
			c.targetsVer = targets.Version
			// FIXME(TUF-0.9) temporarily support files with leading path separators.
			// c.targets = targets.Targets
			c.loadTargets(targets.Targets)
		}
	}
	if loadFailed {
		// If any of the metadata failed to be verified, return the reason for that failure
		return retErr
	}
	return nil
}

// loadAndVerifyLocalRootMeta decodes and verifies root metadata from
// local storage and loads the top-level keys. This method first clears
// the DB for top-level keys and then loads the new keys.
func (c *Client) loadAndVerifyLocalRootMeta(ignoreExpiredCheck bool) error {
	meta, err := c.local.GetMeta()
	if err != nil {
		return err
	}
	rootJSON, ok := meta["root.json"]
	if !ok {
		return ErrNoRootKeys
	}
	return c.loadAndVerifyRootMeta(rootJSON, ignoreExpiredCheck)
}

// loadAndVerifyRootMeta decodes and verifies root metadata and loads the top-level keys.
// This method first clears the DB for top-level keys and then loads the new keys.
func (c *Client) loadAndVerifyRootMeta(rootJSON []byte, ignoreExpiredCheck bool) error {
	// unmarshal root.json without verifying as we need the root
	// keys first
	s := &data.Signed{}
	if err := json.Unmarshal(rootJSON, s); err != nil {
		return err
	}
	root := &data.Root{}
	if err := json.Unmarshal(s.Signed, root); err != nil {
		return err
	}
	ndb := verify.NewDB()
	for id, k := range root.Keys {
		if err := ndb.AddKey(id, k); err != nil {
			return err
		}
	}
	for name, role := range root.Roles {
		if err := ndb.AddRole(name, role); err != nil {
			return err
		}
	}
	// Any trusted local root metadata version must be greater than 0.
	if ignoreExpiredCheck {
		if err := ndb.VerifyIgnoreExpiredCheck(s, "root", 0); err != nil {
			return err
		}
	} else {
		if err := ndb.Verify(s, "root", 0); err != nil {
			return err
		}
	}
	c.consistentSnapshot = root.ConsistentSnapshot
	c.rootVer = root.Version
	c.db = ndb
	return nil
}

// verifyRoot verifies Signed section of the bJSON
// using verification keys in aJSON.
func (c *Client) verifyRoot(aJSON []byte, bJSON []byte) (*data.Root, error) {
	aSigned := &data.Signed{}
	if err := json.Unmarshal(aJSON, aSigned); err != nil {
		return nil, err
	}
	aRoot := &data.Root{}
	if err := json.Unmarshal(aSigned.Signed, aRoot); err != nil {
		return nil, err
	}

	bSigned := &data.Signed{}
	if err := json.Unmarshal(bJSON, bSigned); err != nil {
		return nil, err
	}
	bRoot := &data.Root{}
	if err := json.Unmarshal(bSigned.Signed, bRoot); err != nil {
		return nil, err
	}

	ndb := verify.NewDB()
	for id, k := range aRoot.Keys {
		if err := ndb.AddKey(id, k); err != nil {
			return nil, err
		}
	}
	for name, role := range aRoot.Roles {
		if err := ndb.AddRole(name, role); err != nil {
			return nil, err
		}
	}

	if err := ndb.VerifySignatures(bSigned, "root"); err != nil {
		return nil, err
	}
	return bRoot, nil
}

// FIXME(TUF-0.9) TUF is considering removing support for target files starting
// with a leading path separator. In order to be backwards compatible, we'll
// just remove leading separators for now.
func (c *Client) loadTargets(targets data.TargetFiles) {
	c.targets = make(data.TargetFiles)
	for name, meta := range targets {
		c.targets[name] = meta
		c.targets[util.NormalizeTarget(name)] = meta
	}
}

// downloadMetaUnsafe downloads top-level metadata from remote storage without
// verifying it's length and hashes (used for example to download timestamp.json
// which has unknown size). It will download at most maxMetaSize bytes.
func (c *Client) downloadMetaUnsafe(name string, maxMetaSize int64) ([]byte, error) {
	r, size, err := c.remote.GetMeta(name)
	if err != nil {
		if IsNotFound(err) {
			return nil, ErrMissingRemoteMetadata{name}
		}
		return nil, ErrDownloadFailed{name, err}
	}
	defer r.Close()

	// return ErrMetaTooLarge if the reported size is greater than maxMetaSize
	if size > maxMetaSize {
		return nil, ErrMetaTooLarge{name, size, maxMetaSize}
	}

	// although the size has been checked above, use a LimitReader in case
	// the reported size is inaccurate, or size is -1 which indicates an
	// unknown length
	return io.ReadAll(io.LimitReader(r, maxMetaSize))
}

// remoteGetFunc is the type of function the download method uses to download
// remote files
type remoteGetFunc func(string) (io.ReadCloser, int64, error)

// downloadHashed tries to download the hashed prefixed version of the file.
func (c *Client) downloadHashed(file string, get remoteGetFunc, hashes data.Hashes) (io.ReadCloser, int64, error) {
	// try each hashed path in turn, and either return the contents,
	// try the next one if a 404 is returned, or return an error
	for _, path := range util.HashedPaths(file, hashes) {
		r, size, err := get(path)
		if err != nil {
			if IsNotFound(err) {
				continue
			}
			return nil, 0, err
		}
		return r, size, nil
	}
	return nil, 0, ErrNotFound{file}
}

// download downloads the given target file from remote storage using the get
// function, adding hashes to the path if consistent snapshots are in use
func (c *Client) downloadTarget(file string, get remoteGetFunc, hashes data.Hashes) (io.ReadCloser, int64, error) {
	if c.consistentSnapshot {
		return c.downloadHashed(file, get, hashes)
	} else {
		return get(file)
	}
}

// downloadVersionedMeta downloads top-level metadata from remote storage and
// verifies it using the given file metadata.
func (c *Client) downloadMeta(name string, version int64, m data.FileMeta) ([]byte, error) {
	r, size, err := func() (io.ReadCloser, int64, error) {
		if c.consistentSnapshot {
			path := util.VersionedPath(name, version)
			r, size, err := c.remote.GetMeta(path)
			if err == nil {
				return r, size, nil
			}

			return nil, 0, err
		} else {
			return c.remote.GetMeta(name)
		}
	}()
	if err != nil {
		if IsNotFound(err) {
			return nil, ErrMissingRemoteMetadata{name}
		}
		return nil, err
	}
	defer r.Close()

	// return ErrWrongSize if the reported size is known and incorrect
	var stream io.Reader
	if m.Length != 0 {
		if size >= 0 && size != m.Length {
			return nil, ErrWrongSize{name, size, m.Length}
		}

		// wrap the data in a LimitReader so we download at most m.Length bytes
		stream = io.LimitReader(r, m.Length)
	} else {
		stream = r
	}

	return io.ReadAll(stream)
}

func (c *Client) downloadMetaFromSnapshot(name string, m data.SnapshotFileMeta) ([]byte, error) {
	b, err := c.downloadMeta(name, m.Version, data.FileMeta{Length: m.Length, Hashes: m.Hashes})
	if err != nil {
		return nil, err
	}

	// 5.6.2 – Check length and hashes of fetched bytes *before* parsing metadata
	if err := util.BytesMatchLenAndHashes(b, m.Length, m.Hashes); err != nil {
		return nil, ErrDownloadFailed{name, err}
	}

	meta, err := util.GenerateSnapshotFileMeta(bytes.NewReader(b), m.Hashes.HashAlgorithms()...)
	if err != nil {
		return nil, err
	}

	// 5.6.4 - Check against snapshot role's version
	if err := util.VersionEqual(meta.Version, m.Version); err != nil {
		return nil, ErrDownloadFailed{name, err}
	}

	return b, nil
}

func (c *Client) downloadMetaFromTimestamp(name string, m data.TimestampFileMeta) ([]byte, error) {
	b, err := c.downloadMeta(name, m.Version, data.FileMeta{Length: m.Length, Hashes: m.Hashes})
	if err != nil {
		return nil, err
	}

	// 5.2.2. – Check length and hashes of fetched bytes *before* parsing metadata
	if err := util.BytesMatchLenAndHashes(b, m.Length, m.Hashes); err != nil {
		return nil, ErrDownloadFailed{name, err}
	}

	meta, err := util.GenerateTimestampFileMeta(bytes.NewReader(b), m.Hashes.HashAlgorithms()...)
	if err != nil {
		return nil, err
	}

	// 5.5.4 - Check against timestamp role's version
	if err := util.VersionEqual(meta.Version, m.Version); err != nil {
		return nil, ErrDownloadFailed{name, err}
	}

	return b, nil
}

// decodeSnapshot decodes and verifies snapshot metadata, and returns the new
// root and targets file meta.
func (c *Client) decodeSnapshot(b json.RawMessage) (data.SnapshotFiles, error) {
	snapshot := &data.Snapshot{}
	// 5.5.(3 and 6) - Verify it's signed correctly and it's not expired
	if err := c.db.Unmarshal(b, snapshot, "snapshot", c.snapshotVer); err != nil {
		return data.SnapshotFiles{}, ErrDecodeFailed{"snapshot.json", err}
	}
	// 5.5.5 - Check for top-level targets rollback attack
	// Verify explicitly that current targets meta version is less than or equal to the new one
	if snapshot.Meta["targets.json"].Version < c.targetsVer {
		return data.SnapshotFiles{}, verify.ErrLowVersion{Actual: snapshot.Meta["targets.json"].Version, Current: c.targetsVer}
	}

	// 5.5.5 - Get the local/trusted snapshot metadata, if any, and check all target metafiles against rollback attack
	// In case the local snapshot metadata was not verified by the keys in the latest root during getLocalMeta(),
	// snapshot.json won't be present in c.localMeta and thus this check will not be processed.
	if snapshotJSON, ok := c.localMeta["snapshot.json"]; ok {
		currentSnapshot := &data.Snapshot{}
		if err := c.db.UnmarshalTrusted(snapshotJSON, currentSnapshot, "snapshot"); err != nil {
			return data.SnapshotFiles{}, err
		}
		// 5.5.5 - Check for rollback attacks in both top-level and delegated targets roles (note that the Meta object includes both)
		for path, local := range currentSnapshot.Meta {
			if newMeta, ok := snapshot.Meta[path]; ok {
				// 5.5.5 - Check for rollback attack
				if newMeta.Version < local.Version {
					return data.SnapshotFiles{}, verify.ErrLowVersion{Actual: newMeta.Version, Current: local.Version}
				}
			} else {
				// 5.5.5 - Abort the update if a target file has been removed from the new snapshot file
				return data.SnapshotFiles{}, verify.ErrMissingTargetFile
			}
		}
	}
	// At this point we can trust the new snapshot, the top-level targets, and any delegated targets versions it refers to
	// so we can update the client's trusted versions and proceed with persisting the new snapshot metadata
	// c.snapshotVer was already set when we verified the timestamp metadata
	c.targetsVer = snapshot.Meta["targets.json"].Version
	return snapshot.Meta, nil
}

// decodeTargets decodes and verifies targets metadata, sets c.targets and
// returns updated targets.
func (c *Client) decodeTargets(b json.RawMessage) (data.TargetFiles, error) {
	targets := &data.Targets{}
	// 5.6.(3 and 5) - Verify signatures and check against freeze attack
	if err := c.db.Unmarshal(b, targets, "targets", c.targetsVer); err != nil {
		return nil, ErrDecodeFailed{"targets.json", err}
	}
	// Generate a list with the updated targets
	updatedTargets := make(data.TargetFiles)
	for path, meta := range targets.Targets {
		if local, ok := c.targets[path]; ok {
			if err := util.TargetFileMetaEqual(local, meta); err == nil {
				continue
			}
		}
		updatedTargets[path] = meta
	}
	// c.targetsVer was already updated when we verified the snapshot metadata
	// FIXME(TUF-0.9) temporarily support files with leading path separators.
	// c.targets = targets.Targets
	c.loadTargets(targets.Targets)
	return updatedTargets, nil
}

// decodeTimestamp decodes and verifies timestamp metadata, and returns the
// new snapshot file meta and signals whether the update should be aborted early
// (the new timestamp has the same version as the old one, so there's no need to
// complete the update).
func (c *Client) decodeTimestamp(b json.RawMessage) (data.TimestampFileMeta, bool, error) {
	timestamp := &data.Timestamp{}

	if err := c.db.Unmarshal(b, timestamp, "timestamp", c.timestampVer); err != nil {
		return data.TimestampFileMeta{}, false, ErrDecodeFailed{"timestamp.json", err}
	}
	// 5.4.3.1  - Check for timestamp rollback attack
	// We already checked for timestamp.Version < c.timestampVer in the Unmarshal call above.
	// Here, we're checking for version equality, which indicates that we can abandon this update.
	if timestamp.Version == c.timestampVer {
		return data.TimestampFileMeta{}, true, nil
	}
	// 5.4.3.2 - Check for snapshot rollback attack
	// Verify that the current snapshot meta version is less than or equal to the new one
	if timestamp.Meta["snapshot.json"].Version < c.snapshotVer {
		return data.TimestampFileMeta{}, false, verify.ErrLowVersion{Actual: timestamp.Meta["snapshot.json"].Version, Current: c.snapshotVer}
	}
	// At this point we can trust the new timestamp and the snapshot version it refers to
	// so we can update the client's trusted versions and proceed with persisting the new timestamp
	c.timestampVer = timestamp.Version
	c.snapshotVer = timestamp.Meta["snapshot.json"].Version
	return timestamp.Meta["snapshot.json"], false, nil
}

// hasMetaFromSnapshot checks whether local metadata has the given meta
func (c *Client) hasMetaFromSnapshot(name string, m data.SnapshotFileMeta) bool {
	_, ok := c.localMetaFromSnapshot(name, m)
	return ok
}

// localMetaFromSnapshot returns localmetadata if it matches the snapshot
func (c *Client) localMetaFromSnapshot(name string, m data.SnapshotFileMeta) (json.RawMessage, bool) {
	b, ok := c.localMeta[name]
	if !ok {
		return nil, false
	}
	meta, err := util.GenerateSnapshotFileMeta(bytes.NewReader(b), m.Hashes.HashAlgorithms()...)
	if err != nil {
		return nil, false
	}
	err = util.SnapshotFileMetaEqual(meta, m)
	return b, err == nil
}

type Destination interface {
	io.Writer
	Delete() error
}

// Download downloads the given target file from remote storage into dest.
//
// dest will be deleted and an error returned in the following situations:
//
//   - The target does not exist in the local targets.json
//   - Failed to fetch the chain of delegations accessible from local snapshot.json
//   - The target does not exist in any targets
//   - Metadata cannot be generated for the downloaded data
//   - Generated metadata does not match local metadata for the given file
//   - Size of the download does not match if the reported size is known and
//     incorrect
func (c *Client) Download(name string, dest Destination) (err error) {
	// delete dest if there is an error
	defer func() {
		if err != nil {
			dest.Delete()
		}
	}()

	// populate c.targets from local storage if not set
	if c.targets == nil {
		if err := c.getLocalMeta(); err != nil {
			return err
		}
	}

	normalizedName := util.NormalizeTarget(name)
	localMeta, ok := c.targets[normalizedName]
	if !ok {
		// search in delegations
		localMeta, err = c.getTargetFileMeta(normalizedName)
		if err != nil {
			return err
		}
	}

	// get the data from remote storage
	r, size, err := c.downloadTarget(normalizedName, c.remote.GetTarget, localMeta.Hashes)
	if err != nil {
		return err
	}
	defer r.Close()

	// return ErrWrongSize if the reported size is known and incorrect
	if size >= 0 && size != localMeta.Length {
		return ErrWrongSize{name, size, localMeta.Length}
	}

	// wrap the data in a LimitReader so we download at most localMeta.Length bytes
	stream := io.LimitReader(r, localMeta.Length)

	// read the data, simultaneously writing it to dest and generating metadata
	actual, err := util.GenerateTargetFileMeta(io.TeeReader(stream, dest), localMeta.HashAlgorithms()...)
	if err != nil {
		return ErrDownloadFailed{name, err}
	}

	// check the data has the correct length and hashes
	if err := util.TargetFileMetaEqual(actual, localMeta); err != nil {
		if e, ok := err.(util.ErrWrongLength); ok {
			return ErrWrongSize{name, e.Actual, e.Expected}
		}
		return ErrDownloadFailed{name, err}
	}

	return nil
}

func (c *Client) VerifyDigest(digest string, digestAlg string, length int64, path string) error {
	localMeta, ok := c.targets[path]
	if !ok {
		return ErrUnknownTarget{Name: path, SnapshotVersion: c.snapshotVer}
	}

	actual := data.FileMeta{Length: length, Hashes: make(data.Hashes, 1)}
	var err error
	actual.Hashes[digestAlg], err = hex.DecodeString(digest)
	if err != nil {
		return err
	}

	if err := util.TargetFileMetaEqual(data.TargetFileMeta{FileMeta: actual}, localMeta); err != nil {
		if e, ok := err.(util.ErrWrongLength); ok {
			return ErrWrongSize{path, e.Actual, e.Expected}
		}
		return ErrDownloadFailed{path, err}
	}

	return nil
}

// Target returns the target metadata for a specific target if it
// exists, searching from top-level level targets then through
// all delegations. If it does not, ErrNotFound will be returned.
func (c *Client) Target(name string) (data.TargetFileMeta, error) {
	target, err := c.getTargetFileMeta(util.NormalizeTarget(name))
	if err == nil {
		return target, nil
	}

	if _, ok := err.(ErrUnknownTarget); ok {
		return data.TargetFileMeta{}, ErrNotFound{name}
	}

	return data.TargetFileMeta{}, err
}

// Targets returns the complete list of available top-level targets.
func (c *Client) Targets() (data.TargetFiles, error) {
	// populate c.targets from local storage if not set
	if c.targets == nil {
		if err := c.getLocalMeta(); err != nil {
			return nil, err
		}
	}
	return c.targets, nil
}
