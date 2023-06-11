//
// Copyright 2022 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tuf

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/theupdateframework/go-tuf/client"
	tuf_leveldbstore "github.com/theupdateframework/go-tuf/client/leveldbstore"
	"github.com/theupdateframework/go-tuf/data"
	_ "github.com/theupdateframework/go-tuf/pkg/deprecated/set_ecdsa"
	"github.com/theupdateframework/go-tuf/util"
)

const (
	// DefaultRemoteRoot is the default remote TUF root location.
	DefaultRemoteRoot = "https://sigstore-tuf-root.storage.googleapis.com"

	// TufRootEnv is the name of the environment variable that locates an alternate local TUF root location.
	TufRootEnv = "TUF_ROOT"

	// SigstoreNoCache is the name of the environment variable that, if set, configures this code to only store root data in memory.
	SigstoreNoCache = "SIGSTORE_NO_CACHE"
)

var (
	// singletonTUF holds a single instance of TUF that will get reused on
	// subsequent invocations of initializeTUF
	singletonTUF     *TUF
	singletonTUFOnce = new(sync.Once)
	singletonTUFErr  error
)

// getRemoteRoot is a var for testing.
var getRemoteRoot = func() string { return DefaultRemoteRoot }

type TUF struct {
	sync.Mutex
	client   *client.Client
	targets  targetImpl
	local    client.LocalStore
	remote   client.RemoteStore
	embedded fs.FS
	mirror   string // location of mirror
}

// JSON output representing the configured root status
type RootStatus struct {
	Local    string                    `json:"local"`
	Remote   string                    `json:"remote"`
	Metadata map[string]MetadataStatus `json:"metadata"`
	Targets  []string                  `json:"targets"`
}

type MetadataStatus struct {
	Version    int    `json:"version"`
	Size       int    `json:"len"`
	Expiration string `json:"expiration"`
	Error      string `json:"error"`
}

type TargetFile struct {
	Target []byte
	Status StatusKind
}

type customMetadata struct {
	Usage  UsageKind  `json:"usage"`
	Status StatusKind `json:"status"`
}

type sigstoreCustomMetadata struct {
	Sigstore customMetadata `json:"sigstore"`
}

type signedMeta struct {
	Type    string    `json:"_type"`
	Expires time.Time `json:"expires"`
	Version int64     `json:"version"`
}

// RemoteCache contains information to cache on the location of the remote
// repository.
type remoteCache struct {
	Mirror string `json:"mirror"`
}

func resetForTests() {
	singletonTUFOnce = new(sync.Once)
}

func getExpiration(metadata []byte) (*time.Time, error) {
	s := &data.Signed{}
	if err := json.Unmarshal(metadata, s); err != nil {
		return nil, err
	}
	sm := &signedMeta{}
	if err := json.Unmarshal(s.Signed, sm); err != nil {
		return nil, err
	}
	return &sm.Expires, nil
}

func getVersion(metadata []byte) (int64, error) {
	s := &data.Signed{}
	if err := json.Unmarshal(metadata, s); err != nil {
		return 0, err
	}
	sm := &signedMeta{}
	if err := json.Unmarshal(s.Signed, sm); err != nil {
		return 0, err
	}
	return sm.Version, nil
}

var isExpiredTimestamp = func(metadata []byte) bool {
	expiration, err := getExpiration(metadata)
	if err != nil {
		return true
	}
	return time.Until(*expiration) <= 0
}

func getMetadataStatus(b []byte) (*MetadataStatus, error) {
	expires, err := getExpiration(b)
	if err != nil {
		return nil, err
	}
	version, err := getVersion(b)
	if err != nil {
		return nil, err
	}
	return &MetadataStatus{
		Size:       len(b),
		Expiration: expires.Format(time.RFC822),
		Version:    int(version),
	}, nil
}

func (t *TUF) getRootStatus() (*RootStatus, error) {
	local := rootCacheDir()
	if noCache() {
		local = "in-memory"
	}
	status := &RootStatus{
		Local:    local,
		Remote:   t.mirror,
		Metadata: make(map[string]MetadataStatus),
		Targets:  []string{},
	}

	// Get targets
	targets, err := t.client.Targets()
	if err != nil {
		return nil, err
	}
	for t := range targets {
		status.Targets = append(status.Targets, t)
	}

	// Get metadata expiration
	trustedMeta, err := t.local.GetMeta()
	if err != nil {
		return nil, fmt.Errorf("getting trusted meta: %w", err)
	}
	for role, md := range trustedMeta {
		mdStatus, err := getMetadataStatus(md)
		if err != nil {
			status.Metadata[role] = MetadataStatus{Error: err.Error()}
			continue
		}
		status.Metadata[role] = *mdStatus
	}

	return status, nil
}

func getRoot(meta map[string]json.RawMessage, fallback fs.FS) (json.RawMessage, error) {
	if trustedRoot, ok := meta["root.json"]; ok {
		return trustedRoot, nil
	}
	// On first initialize, there will be no root in the TUF DB, so read from embedded.
	rd, ok := fallback.(fs.ReadFileFS)
	if !ok {
		return nil, errors.New("fs.ReadFileFS unimplemented for embedded repo")
	}
	trustedRoot, err := rd.ReadFile(path.Join("repository", "root.json"))
	if err != nil {
		return nil, err
	}
	return trustedRoot, nil
}

// GetRootStatus gets the current root status for info logging
func GetRootStatus(ctx context.Context) (*RootStatus, error) {
	t, err := NewFromEnv(ctx)
	if err != nil {
		return nil, err
	}
	return t.getRootStatus()
}

// initializeTUF creates a TUF client using the following params:
// * embed: indicates using the embedded metadata and in-memory file updates.
// When this is false, this uses a filesystem cache.
// * mirror: provides a reference to a remote GCS or HTTP mirror.
// * root: provides an external initial root.json. When this is not provided, this
// defaults to the embedded root.json.
// * embedded: An embedded filesystem that provides a trusted root and pre-downloaded
// targets in a targets/ subfolder.
// * forceUpdate: indicates checking the remote for an update, even when the local
// timestamp.json is up to date.
func initializeTUF(mirror string, root []byte, embedded fs.FS, forceUpdate bool) (*TUF, error) {
	singletonTUFOnce.Do(func() {
		t := &TUF{
			mirror:   mirror,
			embedded: embedded,
		}

		t.targets = newFileImpl()
		t.local, singletonTUFErr = newLocalStore()
		if singletonTUFErr != nil {
			return
		}

		t.remote, singletonTUFErr = remoteFromMirror(t.mirror)
		if singletonTUFErr != nil {
			return
		}

		t.client = client.NewClient(t.local, t.remote)

		trustedMeta, err := t.local.GetMeta()
		if err != nil {
			singletonTUFErr = fmt.Errorf("getting trusted meta: %w", err)
			return
		}

		// If the caller does not supply a root, then either use the root in the local store
		// or default to the embedded one.
		if root == nil {
			root, err = getRoot(trustedMeta, t.embedded)
			if err != nil {
				singletonTUFErr = fmt.Errorf("getting trusted root: %w", err)
				return
			}
		}

		if err := t.client.Init(root); err != nil {
			singletonTUFErr = fmt.Errorf("unable to initialize client, local cache may be corrupt: %w", err)
			return
		}

		// We may already have an up-to-date local store! Check to see if it needs to be updated.
		trustedTimestamp, ok := trustedMeta["timestamp.json"]
		if ok && !isExpiredTimestamp(trustedTimestamp) && !forceUpdate {
			// We're golden so stash the TUF object for later use
			singletonTUF = t
			return
		}

		// Update if local is not populated or out of date.
		if err := t.updateMetadataAndDownloadTargets(); err != nil {
			singletonTUFErr = fmt.Errorf("updating local metadata and targets: %w", err)
			return
		}

		// We're golden so stash the TUF object for later use
		singletonTUF = t
	})
	return singletonTUF, singletonTUFErr
}

// TODO: Remove ctx arg.
func NewFromEnv(_ context.Context) (*TUF, error) {
	// Check for the current remote mirror.
	mirror := getRemoteRoot()
	b, err := os.ReadFile(cachedRemote(rootCacheDir()))
	if err == nil {
		remoteInfo := remoteCache{}
		if err := json.Unmarshal(b, &remoteInfo); err == nil {
			mirror = remoteInfo.Mirror
		}
	}

	// Initializes a new TUF object from the local cache or defaults.
	return initializeTUF(mirror, nil, getEmbedded(), false)
}

func Initialize(ctx context.Context, mirror string, root []byte) error {
	// Initialize the client. Force an update with remote.
	if _, err := initializeTUF(mirror, root, getEmbedded(), true); err != nil {
		return err
	}

	// Store the remote for later if we are caching.
	if !noCache() {
		remoteInfo := &remoteCache{Mirror: mirror}
		b, err := json.Marshal(remoteInfo)
		if err != nil {
			return err
		}
		if err := os.WriteFile(cachedRemote(rootCacheDir()), b, 0o600); err != nil {
			return fmt.Errorf("storing remote: %w", err)
		}
	}
	return nil
}

// Checks if the testTarget matches the valid target file metadata.
func isValidTarget(testTarget []byte, validMeta data.TargetFileMeta) bool {
	localMeta, err := util.GenerateTargetFileMeta(bytes.NewReader(testTarget))
	if err != nil {
		return false
	}
	if err := util.TargetFileMetaEqual(localMeta, validMeta); err != nil {
		return false
	}
	return true
}

func (t *TUF) GetTarget(name string) ([]byte, error) {
	t.Lock()
	defer t.Unlock()
	// Get valid target metadata. Does a local verification.
	validMeta, err := t.client.Target(name)
	if err != nil {
		return nil, fmt.Errorf("error verifying local metadata; local cache may be corrupt: %w", err)
	}
	targetBytes, err := t.targets.Get(name)
	if err != nil {
		return nil, err
	}

	if !isValidTarget(targetBytes, validMeta) {
		return nil, fmt.Errorf("cache contains invalid target; local cache may be corrupt")
	}

	return targetBytes, nil
}

// Get target files by a custom usage metadata tag. If there are no files found,
// use the fallback target names to fetch the targets by name.
func (t *TUF) GetTargetsByMeta(usage UsageKind, fallbacks []string) ([]TargetFile, error) {
	t.Lock()
	targets, err := t.client.Targets()
	t.Unlock()
	if err != nil {
		return nil, fmt.Errorf("error getting targets: %w", err)
	}
	var matchedTargets []TargetFile
	for name, targetMeta := range targets {
		// Skip any targets that do not include custom metadata.
		if targetMeta.Custom == nil {
			continue
		}
		var scm sigstoreCustomMetadata
		err := json.Unmarshal(*targetMeta.Custom, &scm)
		if err != nil {
			fmt.Fprintf(os.Stderr, "**Warning** Custom metadata not configured properly for target %s, skipping target\n", name)
			continue
		}
		if scm.Sigstore.Usage == usage {
			target, err := t.GetTarget(name)
			if err != nil {
				return nil, fmt.Errorf("error getting target %s by usage: %w", name, err)
			}
			matchedTargets = append(matchedTargets, TargetFile{Target: target, Status: scm.Sigstore.Status})
		}
	}
	if len(matchedTargets) == 0 {
		for _, fallback := range fallbacks {
			target, err := t.GetTarget(fallback)
			if err != nil {
				fmt.Fprintf(os.Stderr, "**Warning** Missing fallback target %s, skipping\n", fallback)
				continue
			}
			matchedTargets = append(matchedTargets, TargetFile{Target: target, Status: Active})
		}
	}
	if len(matchedTargets) == 0 {
		return matchedTargets, fmt.Errorf("no matching targets by custom metadata, fallbacks not found: %s", strings.Join(fallbacks, ", "))
	}
	return matchedTargets, nil
}

// updateClient() updates the TUF client and also caches new metadata, if needed.
func (t *TUF) updateClient() (data.TargetFiles, error) {
	targets, err := t.client.Update()
	if err != nil {
		// Get some extra information for debugging. What was the state of the top-level
		// metadata on the remote?
		status := struct {
			Mirror   string                    `json:"mirror"`
			Metadata map[string]MetadataStatus `json:"metadata"`
		}{
			Mirror:   t.mirror,
			Metadata: make(map[string]MetadataStatus),
		}
		for _, md := range []string{"root.json", "targets.json", "snapshot.json", "timestamp.json"} {
			r, _, err := t.remote.GetMeta(md)
			if err != nil {
				// May be missing, or failed download.
				continue
			}
			defer r.Close()
			b, err := io.ReadAll(r)
			if err != nil {
				continue
			}
			mdStatus, err := getMetadataStatus(b)
			if err != nil {
				continue
			}
			status.Metadata[md] = *mdStatus
		}
		b, innerErr := json.MarshalIndent(status, "", "\t")
		if innerErr != nil {
			return nil, innerErr
		}
		return nil, fmt.Errorf("error updating to TUF remote mirror: %w\nremote status:%s", err, string(b))
	}
	// Success! Cache new metadata, if needed.
	if noCache() {
		return targets, nil
	}
	// Sync the on-disk cache with the metadata from the in-memory store.
	tufDB := filepath.FromSlash(filepath.Join(rootCacheDir(), "tuf.db"))
	diskLocal, err := tuf_leveldbstore.FileLocalStore(tufDB)
	defer func() {
		if diskLocal != nil {
			diskLocal.Close()
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("creating cached local store: %w", err)
	}
	if err := syncLocalMeta(t.local, diskLocal); err != nil {
		return nil, err
	}
	// Return updated targets.
	return targets, nil
}

func (t *TUF) updateMetadataAndDownloadTargets() error {
	// Download updated targets and cache new metadata and targets in ${TUF_ROOT}.
	// NOTE: This only returns *updated* targets.
	targetFiles, err := t.updateClient()
	if err != nil {
		return err
	}

	// Download **newly** updated targets.
	// TODO: Consider lazily downloading these -- be careful with embedded targets if so.
	for name, targetMeta := range targetFiles {
		if err := maybeDownloadRemoteTarget(name, targetMeta, t); err != nil {
			return err
		}
	}

	return nil
}

type targetDestination struct {
	buf *bytes.Buffer
}

func (t *targetDestination) Write(b []byte) (int, error) {
	return t.buf.Write(b)
}

func (t *targetDestination) Delete() error {
	t.buf = &bytes.Buffer{}
	return nil
}

func maybeDownloadRemoteTarget(name string, meta data.TargetFileMeta, t *TUF) error {
	// If we already have the target locally, don't bother downloading from remote storage.
	if cachedTarget, err := t.targets.Get(name); err == nil {
		// If the target we have stored matches the meta, use that.
		if isValidTarget(cachedTarget, meta) {
			return nil
		}
	}

	// Check if we already have the target in the embedded store.
	w := bytes.Buffer{}
	rd, ok := t.embedded.(fs.ReadFileFS)
	if !ok {
		return errors.New("fs.ReadFileFS unimplemented for embedded repo")
	}
	b, err := rd.ReadFile(path.Join("repository", "targets", name))

	if err == nil {
		// Unfortunately go:embed appears to somehow replace our line endings on windows, we need to switch them back.
		// It should theoretically be safe to do this everywhere - but the files only seem to get mutated on Windows so
		// let's only change them back there.
		if runtime.GOOS == "windows" {
			b = bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
		}

		if isValidTarget(b, meta) {
			if _, err := io.Copy(&w, bytes.NewReader(b)); err != nil {
				return fmt.Errorf("using embedded target: %w", err)
			}
		}
	}

	// Nope -- no local matching target, go download it.
	if w.Len() == 0 {
		dest := targetDestination{buf: &w}
		if err := t.client.Download(name, &dest); err != nil {
			return fmt.Errorf("downloading target: %w", err)
		}
	}

	// Set the target in the cache.
	if err := t.targets.Set(name, w.Bytes()); err != nil {
		return err
	}
	return nil
}

func rootCacheDir() string {
	rootDir := os.Getenv(TufRootEnv)
	if rootDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = ""
		}
		return filepath.FromSlash(filepath.Join(home, ".sigstore", "root"))
	}
	return rootDir
}

func cachedRemote(cacheRoot string) string {
	return filepath.FromSlash(filepath.Join(cacheRoot, "remote.json"))
}

func cachedTargetsDir(cacheRoot string) string {
	return filepath.FromSlash(filepath.Join(cacheRoot, "targets"))
}

func syncLocalMeta(from, to client.LocalStore) error {
	// Copy trusted metadata in the from LocalStore into the to LocalStore.
	tufLocalStoreMeta, err := from.GetMeta()
	if err != nil {
		return fmt.Errorf("getting metadata to sync: %w", err)
	}
	for k, v := range tufLocalStoreMeta {
		if err := to.SetMeta(k, v); err != nil {
			return fmt.Errorf("syncing local store for metadata %s", k)
		}
	}
	return nil
}

// Local store implementations
func newLocalStore() (client.LocalStore, error) {
	local := client.MemoryLocalStore()
	if noCache() {
		return local, nil
	}
	// Otherwise populate the in-memory local store with data fetched from the cache.
	tufDB := filepath.FromSlash(filepath.Join(rootCacheDir(), "tuf.db"))
	diskLocal, err := tuf_leveldbstore.FileLocalStore(tufDB)
	defer func() {
		if diskLocal != nil {
			diskLocal.Close()
		}
	}()
	if err != nil {
		return nil, fmt.Errorf("creating cached local store: %w", err)
	}
	// Populate the in-memory local store with data fetched from the cache.
	if err := syncLocalMeta(diskLocal, local); err != nil {
		return nil, err
	}
	return local, nil
}

//go:embed repository
var embeddedRootRepo embed.FS

// getEmbedded is a var for testing.
var getEmbedded = func() fs.FS { return embeddedRootRepo }

// Target Implementations
type targetImpl interface {
	Set(string, []byte) error
	Get(string) ([]byte, error)
}

func newFileImpl() targetImpl {
	memTargets := &memoryCache{}
	if noCache() {
		return memTargets
	}
	// Otherwise use a disk-cache with in-memory cached targets.
	return &diskCache{
		base:   cachedTargetsDir(rootCacheDir()),
		memory: memTargets,
	}
}

// In-memory cache for targets
type memoryCache struct {
	targets map[string][]byte
}

func (m *memoryCache) Set(p string, b []byte) error {
	if m.targets == nil {
		m.targets = map[string][]byte{}
	}
	m.targets[p] = b
	return nil
}

func (m *memoryCache) Get(p string) ([]byte, error) {
	if m.targets == nil {
		return nil, fmt.Errorf("no cached targets available, cannot retrieve %s", p)
	}
	b, ok := m.targets[p]
	if !ok {
		return nil, fmt.Errorf("missing cached target %s", p)
	}
	return b, nil
}

// On-disk cache for targets
type diskCache struct {
	// Base directory for accessing targets.
	base string
	// An in-memory map of targets that are kept in sync.
	memory *memoryCache
}

func (d *diskCache) Get(p string) ([]byte, error) {
	// Read from the in-memory cache first.
	if b, err := d.memory.Get(p); err == nil {
		return b, nil
	}
	fp := filepath.FromSlash(filepath.Join(d.base, p))
	return os.ReadFile(fp)
}

func (d *diskCache) Set(p string, b []byte) error {
	if err := d.memory.Set(p, b); err != nil {
		return err
	}

	fp := filepath.FromSlash(filepath.Join(d.base, p))
	if err := os.MkdirAll(filepath.Dir(fp), 0o700); err != nil {
		return fmt.Errorf("creating targets dir: %w", err)
	}

	return os.WriteFile(fp, b, 0o600)
}

func noCache() bool {
	b, err := strconv.ParseBool(os.Getenv(SigstoreNoCache))
	if err != nil {
		return false
	}
	return b
}

func remoteFromMirror(mirror string) (client.RemoteStore, error) {
	// This is for compatibility with specifying a GCS bucket remote.
	u, parseErr := url.ParseRequestURI(mirror)
	if parseErr != nil {
		return client.HTTPRemoteStore(fmt.Sprintf("https://%s.storage.googleapis.com", mirror), nil, nil)
	}
	if u.Scheme != "file" {
		return client.HTTPRemoteStore(mirror, nil, nil)
	}
	// Use local filesystem for remote.
	return client.NewFileRemoteStore(os.DirFS(u.Path), "")
}
