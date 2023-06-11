package tuf

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/secure-systems-lab/go-securesystemslib/cjson"
	"github.com/theupdateframework/go-tuf/data"
	"github.com/theupdateframework/go-tuf/internal/roles"
	"github.com/theupdateframework/go-tuf/internal/sets"
	"github.com/theupdateframework/go-tuf/internal/signer"
	"github.com/theupdateframework/go-tuf/pkg/keys"
	"github.com/theupdateframework/go-tuf/pkg/targets"
	"github.com/theupdateframework/go-tuf/sign"
	"github.com/theupdateframework/go-tuf/util"
	"github.com/theupdateframework/go-tuf/verify"
)

const (
	// The maximum number of delegations to visit while traversing the delegations graph.
	defaultMaxDelegations = 32
)

// topLevelMetadata determines the order signatures are verified when committing.
var topLevelMetadata = []string{
	"root.json",
	"targets.json",
	"snapshot.json",
	"timestamp.json",
}

// TargetsWalkFunc is a function of a target path name and a target payload used to
// execute some function on each staged target file. For example, it may normalize path
// names and generate target file metadata with additional custom metadata.
type TargetsWalkFunc func(path string, target io.Reader) error

type Repo struct {
	local          LocalStore
	hashAlgorithms []string
	meta           map[string]json.RawMessage
	prefix         string
	indent         string
	logger         *log.Logger
}

type RepoOpts func(r *Repo)

func WithLogger(logger *log.Logger) RepoOpts {
	return func(r *Repo) {
		r.logger = logger
	}
}

func WithHashAlgorithms(hashAlgorithms ...string) RepoOpts {
	return func(r *Repo) {
		r.hashAlgorithms = hashAlgorithms
	}
}

func WithPrefix(prefix string) RepoOpts {
	return func(r *Repo) {
		r.prefix = prefix
	}
}

func WithIndex(indent string) RepoOpts {
	return func(r *Repo) {
		r.indent = indent
	}
}

func NewRepo(local LocalStore, hashAlgorithms ...string) (*Repo, error) {
	return NewRepoIndent(local, "", "", hashAlgorithms...)
}

func NewRepoIndent(local LocalStore, prefix string, indent string,
	hashAlgorithms ...string) (*Repo, error) {
	r := &Repo{
		local:          local,
		hashAlgorithms: hashAlgorithms,
		prefix:         prefix,
		indent:         indent,
		logger:         log.New(io.Discard, "", 0),
	}

	var err error
	r.meta, err = local.GetMeta()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func NewRepoWithOpts(local LocalStore, opts ...RepoOpts) (*Repo, error) {
	r, err := NewRepo(local)
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		opt(r)
	}
	return r, nil
}

func (r *Repo) Init(consistentSnapshot bool) error {
	t, err := r.topLevelTargets()
	if err != nil {
		return err
	}
	if len(t.Targets) > 0 {
		return ErrInitNotAllowed
	}
	root := data.NewRoot()
	root.ConsistentSnapshot = consistentSnapshot
	// Set root version to 1 for a new root.
	root.Version = 1
	if err = r.setMeta("root.json", root); err != nil {
		return err
	}

	t.Version = 1
	if err = r.setMeta("targets.json", t); err != nil {
		return err
	}

	r.logger.Println("Repository initialized")
	return nil
}

func (r *Repo) topLevelKeysDB() (*verify.DB, error) {
	db := verify.NewDB()
	root, err := r.root()
	if err != nil {
		return nil, err
	}
	for id, k := range root.Keys {
		if err := db.AddKey(id, k); err != nil {
			return nil, err
		}
	}
	for name, role := range root.Roles {
		if err := db.AddRole(name, role); err != nil {
			return nil, err
		}
	}
	return db, nil
}

func (r *Repo) root() (*data.Root, error) {
	rootJSON, ok := r.meta["root.json"]
	if !ok {
		return data.NewRoot(), nil
	}
	s := &data.Signed{}
	if err := json.Unmarshal(rootJSON, s); err != nil {
		return nil, err
	}
	root := &data.Root{}
	if err := json.Unmarshal(s.Signed, root); err != nil {
		return nil, err
	}
	return root, nil
}

func (r *Repo) snapshot() (*data.Snapshot, error) {
	snapshotJSON, ok := r.meta["snapshot.json"]
	if !ok {
		return data.NewSnapshot(), nil
	}
	s := &data.Signed{}
	if err := json.Unmarshal(snapshotJSON, s); err != nil {
		return nil, err
	}
	snapshot := &data.Snapshot{}
	if err := json.Unmarshal(s.Signed, snapshot); err != nil {
		return nil, err
	}
	return snapshot, nil
}

func (r *Repo) RootVersion() (int64, error) {
	root, err := r.root()
	if err != nil {
		return -1, err
	}
	return root.Version, nil
}

func (r *Repo) GetThreshold(keyRole string) (int, error) {
	if roles.IsDelegatedTargetsRole(keyRole) {
		// The signature threshold for a delegated targets role
		// depends on the incoming delegation edge.
		return -1, ErrInvalidRole{keyRole, "only thresholds for top-level roles supported"}
	}
	root, err := r.root()
	if err != nil {
		return -1, err
	}
	role, ok := root.Roles[keyRole]
	if !ok {
		return -1, ErrInvalidRole{keyRole, "role missing from root metadata"}
	}

	return role.Threshold, nil
}

func (r *Repo) SetThreshold(keyRole string, t int) error {
	if roles.IsDelegatedTargetsRole(keyRole) {
		// The signature threshold for a delegated targets role
		// depends on the incoming delegation edge.
		return ErrInvalidRole{keyRole, "only thresholds for top-level roles supported"}
	}
	root, err := r.root()
	if err != nil {
		return err
	}
	role, ok := root.Roles[keyRole]
	if !ok {
		return ErrInvalidRole{keyRole, "role missing from root metadata"}
	}
	if role.Threshold == t {
		// Change was a no-op.
		return nil
	}
	role.Threshold = t
	if !r.local.FileIsStaged("root.json") {
		root.Version++
	}
	return r.setMeta("root.json", root)
}

func (r *Repo) Targets() (data.TargetFiles, error) {
	targets, err := r.topLevelTargets()
	if err != nil {
		return nil, err
	}
	return targets.Targets, nil
}

func (r *Repo) SetTargetsVersion(v int64) error {
	t, err := r.topLevelTargets()
	if err != nil {
		return err
	}
	t.Version = v
	return r.setMeta("targets.json", t)
}

func (r *Repo) TargetsVersion() (int64, error) {
	t, err := r.topLevelTargets()
	if err != nil {
		return -1, err
	}
	return t.Version, nil
}

func (r *Repo) SetTimestampVersion(v int64) error {
	ts, err := r.timestamp()
	if err != nil {
		return err
	}
	ts.Version = v
	return r.setMeta("timestamp.json", ts)
}

func (r *Repo) TimestampVersion() (int64, error) {
	ts, err := r.timestamp()
	if err != nil {
		return -1, err
	}
	return ts.Version, nil
}

func (r *Repo) SetSnapshotVersion(v int64) error {
	s, err := r.snapshot()
	if err != nil {
		return err
	}

	s.Version = v
	return r.setMeta("snapshot.json", s)
}

func (r *Repo) SnapshotVersion() (int64, error) {
	s, err := r.snapshot()
	if err != nil {
		return -1, err
	}
	return s.Version, nil
}

func (r *Repo) topLevelTargets() (*data.Targets, error) {
	return r.targets("targets")
}

func (r *Repo) targets(metaName string) (*data.Targets, error) {
	targetsJSON, ok := r.meta[metaName+".json"]
	if !ok {
		return data.NewTargets(), nil
	}
	s := &data.Signed{}
	if err := json.Unmarshal(targetsJSON, s); err != nil {
		return nil, fmt.Errorf("error unmarshalling for targets %q: %w", metaName, err)
	}
	targets := &data.Targets{}
	if err := json.Unmarshal(s.Signed, targets); err != nil {
		return nil, fmt.Errorf("error unmarshalling signed data for targets %q: %w", metaName, err)
	}
	return targets, nil
}

func (r *Repo) timestamp() (*data.Timestamp, error) {
	timestampJSON, ok := r.meta["timestamp.json"]
	if !ok {
		return data.NewTimestamp(), nil
	}
	s := &data.Signed{}
	if err := json.Unmarshal(timestampJSON, s); err != nil {
		return nil, err
	}
	timestamp := &data.Timestamp{}
	if err := json.Unmarshal(s.Signed, timestamp); err != nil {
		return nil, err
	}
	return timestamp, nil
}

func (r *Repo) ChangePassphrase(keyRole string) error {
	if p, ok := r.local.(PassphraseChanger); ok {
		return p.ChangePassphrase(keyRole)
	}

	return ErrChangePassphraseNotSupported
}

func (r *Repo) GenKey(role string) ([]string, error) {
	// Not compatible with delegated targets roles, since delegated targets keys
	// are associated with a delegation (edge), not a role (node).

	return r.GenKeyWithExpires(role, data.DefaultExpires(role))
}

func (r *Repo) GenKeyWithExpires(keyRole string, expires time.Time) (keyids []string, err error) {
	return r.GenKeyWithSchemeAndExpires(keyRole, expires, data.KeySchemeEd25519)
}

func (r *Repo) GenKeyWithSchemeAndExpires(role string, expires time.Time, keyScheme data.KeyScheme) ([]string, error) {
	var signer keys.Signer
	var err error
	switch keyScheme {
	case data.KeySchemeEd25519:
		signer, err = keys.GenerateEd25519Key()
	case data.KeySchemeECDSA_SHA2_P256:
		signer, err = keys.GenerateEcdsaKey()
	case data.KeySchemeRSASSA_PSS_SHA256:
		signer, err = keys.GenerateRsaKey()
	default:
		return nil, errors.New("unknown key type")
	}
	if err != nil {
		return nil, err
	}

	// Not compatible with delegated targets roles, since delegated targets keys
	// are associated with a delegation (edge), not a role (node).

	if err = r.AddPrivateKeyWithExpires(role, signer, expires); err != nil {
		return nil, err
	}
	return signer.PublicData().IDs(), nil
}

func (r *Repo) AddPrivateKey(role string, signer keys.Signer) error {
	// Not compatible with delegated targets roles, since delegated targets keys
	// are associated with a delegation (edge), not a role (node).

	return r.AddPrivateKeyWithExpires(role, signer, data.DefaultExpires(role))
}

func (r *Repo) AddPrivateKeyWithExpires(keyRole string, signer keys.Signer, expires time.Time) error {
	// Not compatible with delegated targets roles, since delegated targets keys
	// are associated with a delegation (edge), not a role (node).

	if roles.IsDelegatedTargetsRole(keyRole) {
		return ErrInvalidRole{keyRole, "only support adding keys for top-level roles"}
	}

	if !validExpires(expires) {
		return ErrInvalidExpires{expires}
	}

	// Must add signer before adding verification key, so
	// root.json can be signed when a new root key is added.
	if err := r.local.SaveSigner(keyRole, signer); err != nil {
		return err
	}

	if err := r.AddVerificationKeyWithExpiration(keyRole, signer.PublicData(), expires); err != nil {
		return err
	}

	return nil
}

func (r *Repo) AddVerificationKey(keyRole string, pk *data.PublicKey) error {
	// Not compatible with delegated targets roles, since delegated targets keys
	// are associated with a delegation (edge), not a role (node).

	return r.AddVerificationKeyWithExpiration(keyRole, pk, data.DefaultExpires(keyRole))
}

func (r *Repo) AddVerificationKeyWithExpiration(keyRole string, pk *data.PublicKey, expires time.Time) error {
	// Not compatible with delegated targets roles, since delegated targets keys
	// are associated with a delegation (edge), not a role (node).

	if roles.IsDelegatedTargetsRole(keyRole) {
		return ErrInvalidRole{
			Role:   keyRole,
			Reason: "only top-level targets roles are supported",
		}
	}

	if !validExpires(expires) {
		return ErrInvalidExpires{expires}
	}

	root, err := r.root()
	if err != nil {
		return err
	}

	role, ok := root.Roles[keyRole]
	if !ok {
		role = &data.Role{KeyIDs: []string{}, Threshold: 1}
		root.Roles[keyRole] = role
	}
	changed := false
	if role.AddKeyIDs(pk.IDs()) {
		changed = true
	}

	if root.AddKey(pk) {
		changed = true
	}

	if !changed {
		return nil
	}

	root.Expires = expires.Round(time.Second)
	if !r.local.FileIsStaged("root.json") {
		root.Version++
	}

	return r.setMeta("root.json", root)
}

func validExpires(expires time.Time) bool {
	return time.Until(expires) > 0
}

func (r *Repo) RootKeys() ([]*data.PublicKey, error) {
	root, err := r.root()
	if err != nil {
		return nil, err
	}
	role, ok := root.Roles["root"]
	if !ok {
		return nil, nil
	}

	// We might have multiple key ids that correspond to the same key, so
	// make sure we only return unique keys.
	seen := make(map[string]struct{})
	rootKeys := []*data.PublicKey{}
	for _, id := range role.KeyIDs {
		key, ok := root.Keys[id]
		if !ok {
			return nil, fmt.Errorf("tuf: invalid root metadata")
		}
		found := false
		if _, ok := seen[id]; ok {
			found = true
			break
		}
		if !found {
			for _, id := range key.IDs() {
				seen[id] = struct{}{}
			}
			rootKeys = append(rootKeys, key)
		}
	}
	return rootKeys, nil
}

func (r *Repo) RevokeKey(role, id string) error {
	// Not compatible with delegated targets roles, since delegated targets keys
	// are associated with a delegation (edge), not a role (node).

	return r.RevokeKeyWithExpires(role, id, data.DefaultExpires("root"))
}

func (r *Repo) RevokeKeyWithExpires(keyRole, id string, expires time.Time) error {
	// Not compatible with delegated targets roles, since delegated targets keys
	// are associated with a delegation (edge), not a role (node).

	if roles.IsDelegatedTargetsRole(keyRole) {
		return ErrInvalidRole{keyRole, "only revocations for top-level roles supported"}
	}

	if !validExpires(expires) {
		return ErrInvalidExpires{expires}
	}

	root, err := r.root()
	if err != nil {
		return err
	}

	key, ok := root.Keys[id]
	if !ok {
		return ErrKeyNotFound{keyRole, id}
	}

	role, ok := root.Roles[keyRole]
	if !ok {
		return ErrKeyNotFound{keyRole, id}
	}

	// Create a list of filtered key IDs that do not contain the revoked key IDs.
	filteredKeyIDs := make([]string, 0, len(role.KeyIDs))

	// There may be multiple keyids that correspond to this key, so
	// filter all of them out.
	for _, keyID := range role.KeyIDs {
		if !key.ContainsID(keyID) {
			filteredKeyIDs = append(filteredKeyIDs, keyID)
		}
	}
	if len(filteredKeyIDs) == len(role.KeyIDs) {
		return ErrKeyNotFound{keyRole, id}
	}
	role.KeyIDs = filteredKeyIDs
	root.Roles[keyRole] = role

	// Only delete the key from root.Keys if the key is no longer in use by
	// any other role.
	key_in_use := false
	for _, role := range root.Roles {
		for _, keyID := range role.KeyIDs {
			if key.ContainsID(keyID) {
				key_in_use = true
			}
		}
	}
	if !key_in_use {
		for _, keyID := range key.IDs() {
			delete(root.Keys, keyID)
		}
	}
	root.Expires = expires.Round(time.Second)
	if !r.local.FileIsStaged("root.json") {
		root.Version++
	}

	err = r.setMeta("root.json", root)
	if err == nil {
		r.logger.Println("Revoked", keyRole, "key with ID", id, "in root metadata")
	}
	return err
}

// AddDelegatedRole is equivalent to AddDelegatedRoleWithExpires, but
// with a default expiration time.
func (r *Repo) AddDelegatedRole(delegator string, delegatedRole data.DelegatedRole, keys []*data.PublicKey) error {
	return r.AddDelegatedRoleWithExpires(delegator, delegatedRole, keys, data.DefaultExpires("targets"))
}

// AddDelegatedRoleWithExpires adds a delegation from the delegator to the
// role specified in the role argument. Key IDs referenced in role.KeyIDs
// should have corresponding Key entries in the keys argument. New metadata is
// written with the given expiration time.
func (r *Repo) AddDelegatedRoleWithExpires(delegator string, delegatedRole data.DelegatedRole, keys []*data.PublicKey, expires time.Time) error {
	expires = expires.Round(time.Second)

	t, err := r.targets(delegator)
	if err != nil {
		return fmt.Errorf("error getting delegator (%q) metadata: %w", delegator, err)
	}

	if t.Delegations == nil {
		t.Delegations = &data.Delegations{}
		t.Delegations.Keys = make(map[string]*data.PublicKey)
	}

	for _, keyID := range delegatedRole.KeyIDs {
		for _, key := range keys {
			if key.ContainsID(keyID) {
				t.Delegations.Keys[keyID] = key
				break
			}
		}
	}

	for _, r := range t.Delegations.Roles {
		if r.Name == delegatedRole.Name {
			return fmt.Errorf("role: %s is already delegated to by %s", delegatedRole.Name, r.Name)
		}
	}
	t.Delegations.Roles = append(t.Delegations.Roles, delegatedRole)
	t.Expires = expires

	delegatorFile := delegator + ".json"
	if !r.local.FileIsStaged(delegatorFile) {
		t.Version++
	}

	err = r.setMeta(delegatorFile, t)
	if err != nil {
		return fmt.Errorf("error setting metadata for %q: %w", delegatorFile, err)
	}

	delegatee := delegatedRole.Name
	dt, err := r.targets(delegatee)
	if err != nil {
		return fmt.Errorf("error getting delegatee (%q) metadata: %w", delegatee, err)
	}
	dt.Expires = expires

	delegateeFile := delegatee + ".json"
	if !r.local.FileIsStaged(delegateeFile) {
		dt.Version++
	}

	err = r.setMeta(delegateeFile, dt)
	if err != nil {
		return fmt.Errorf("error setting metadata for %q: %w", delegateeFile, err)
	}

	return nil
}

// AddDelegatedRolesForPathHashBins is equivalent to
// AddDelegatedRolesForPathHashBinsWithExpires, but with a default
// expiration time.
func (r *Repo) AddDelegatedRolesForPathHashBins(delegator string, bins *targets.HashBins, keys []*data.PublicKey, threshold int) error {
	return r.AddDelegatedRolesForPathHashBinsWithExpires(delegator, bins, keys, threshold, data.DefaultExpires("targets"))
}

// AddDelegatedRolesForPathHashBinsWithExpires adds delegations to the
// delegator role for the given hash bins configuration. New metadata is
// written with the given expiration time.
func (r *Repo) AddDelegatedRolesForPathHashBinsWithExpires(delegator string, bins *targets.HashBins, keys []*data.PublicKey, threshold int, expires time.Time) error {
	keyIDs := []string{}
	for _, key := range keys {
		keyIDs = append(keyIDs, key.IDs()...)
	}

	n := bins.NumBins()
	for i := uint64(0); i < n; i += 1 {
		bin := bins.GetBin(i)
		name := bin.RoleName()
		err := r.AddDelegatedRoleWithExpires(delegator, data.DelegatedRole{
			Name:             name,
			KeyIDs:           sets.DeduplicateStrings(keyIDs),
			PathHashPrefixes: bin.HashPrefixes(),
			Threshold:        threshold,
		}, keys, expires)
		if err != nil {
			return fmt.Errorf("error adding delegation from %v to %v: %w", delegator, name, err)
		}
	}

	return nil
}

// ResetTargetsDelegation is equivalent to ResetTargetsDelegationsWithExpires
// with a default expiry time.
func (r *Repo) ResetTargetsDelegations(delegator string) error {
	return r.ResetTargetsDelegationsWithExpires(delegator, data.DefaultExpires("targets"))
}

// ResetTargetsDelegationsWithExpires removes all targets delegations from the
// given delegator role. New metadata is written with the given expiration
// time.
func (r *Repo) ResetTargetsDelegationsWithExpires(delegator string, expires time.Time) error {
	t, err := r.targets(delegator)
	if err != nil {
		return fmt.Errorf("error getting delegator (%q) metadata: %w", delegator, err)
	}

	t.Delegations = &data.Delegations{}
	t.Delegations.Keys = make(map[string]*data.PublicKey)

	t.Expires = expires.Round(time.Second)

	delegatorFile := delegator + ".json"
	if !r.local.FileIsStaged(delegatorFile) {
		t.Version++
	}

	err = r.setMeta(delegatorFile, t)
	if err != nil {
		return fmt.Errorf("error setting metadata for %q: %w", delegatorFile, err)
	}

	return nil
}

func (r *Repo) jsonMarshal(v interface{}) ([]byte, error) {
	if r.prefix == "" && r.indent == "" {
		return json.Marshal(v)
	}
	return json.MarshalIndent(v, r.prefix, r.indent)
}

func (r *Repo) dbsForRole(role string) ([]*verify.DB, error) {
	dbs := []*verify.DB{}

	if roles.IsTopLevelRole(role) {
		db, err := r.topLevelKeysDB()
		if err != nil {
			return nil, err
		}
		dbs = append(dbs, db)
	} else {
		ddbs, err := r.delegatorDBs(role)
		if err != nil {
			return nil, err
		}

		dbs = append(dbs, ddbs...)
	}

	return dbs, nil
}

func (r *Repo) signersForRole(role string) ([]keys.Signer, error) {
	dbs, err := r.dbsForRole(role)
	if err != nil {
		return nil, err
	}

	signers := []keys.Signer{}
	for _, db := range dbs {
		ss, err := r.getSignersInDB(role, db)
		if err != nil {
			return nil, err
		}

		signers = append(signers, ss...)
	}

	return signers, nil
}

func (r *Repo) setMeta(roleFilename string, meta interface{}) error {
	role := strings.TrimSuffix(roleFilename, ".json")

	signers, err := r.signersForRole(role)
	if err != nil {
		return err
	}

	s, err := sign.Marshal(meta, signers...)
	if err != nil {
		return err
	}
	b, err := r.jsonMarshal(s)
	if err != nil {
		return err
	}
	r.meta[roleFilename] = b
	return r.local.SetMeta(roleFilename, b)
}

// SignPayload signs the given payload using the key(s) associated with role.
//
// It returns the total number of keys used for signing, 0 (along with
// ErrNoKeys) if no keys were found, or -1 (along with an error) in error cases.
func (r *Repo) SignPayload(role string, payload *data.Signed) (int, error) {
	keys, err := r.signersForRole(role)
	if err != nil {
		return -1, err
	}
	if len(keys) == 0 {
		return 0, ErrNoKeys{role}
	}
	for _, k := range keys {
		if err = sign.Sign(payload, k); err != nil {
			return -1, err
		}
	}
	return len(keys), nil
}

func (r *Repo) Sign(roleFilename string) error {
	signed, err := r.SignedMeta(roleFilename)
	if err != nil {
		return err
	}

	role := strings.TrimSuffix(roleFilename, ".json")
	numKeys, err := r.SignPayload(role, signed)
	if errors.Is(err, ErrNoKeys{role}) {
		return ErrNoKeys{roleFilename}
	} else if err != nil {
		return err
	}

	b, err := r.jsonMarshal(signed)
	if err != nil {
		return err
	}
	r.meta[roleFilename] = b
	err = r.local.SetMeta(roleFilename, b)
	if err == nil {
		r.logger.Println("Signed", roleFilename, "with", numKeys, "key(s)")
	}
	return err
}

// AddOrUpdateSignature allows users to add or update a signature generated with an external tool.
// The name must be a valid metadata file name, like root.json.
func (r *Repo) AddOrUpdateSignature(roleFilename string, signature data.Signature) error {
	role := strings.TrimSuffix(roleFilename, ".json")

	// Check key ID is in valid for the role.
	dbs, err := r.dbsForRole(role)
	if err != nil {
		return err
	}

	if len(dbs) == 0 {
		return ErrInvalidRole{role, "no trusted keys for role"}
	}

	keyInDB := false
	for _, db := range dbs {
		roleData := db.GetRole(role)
		if roleData == nil {
			return ErrInvalidRole{role, "role is not in verifier DB"}
		}
		if roleData.ValidKey(signature.KeyID) {
			keyInDB = true
		}
	}
	if !keyInDB {
		return verify.ErrInvalidKey
	}

	s, err := r.SignedMeta(roleFilename)
	if err != nil {
		return err
	}

	// Add or update signature.
	signatures := make([]data.Signature, 0, len(s.Signatures)+1)
	for _, sig := range s.Signatures {
		if sig.KeyID != signature.KeyID {
			signatures = append(signatures, sig)
		}
	}
	signatures = append(signatures, signature)
	s.Signatures = signatures

	// Check signature on signed meta. Ignore threshold errors as this may not be fully
	// signed.
	for _, db := range dbs {
		if err := db.VerifySignatures(s, role); err != nil {
			if _, ok := err.(verify.ErrRoleThreshold); !ok {
				return err
			}
		}
	}

	b, err := r.jsonMarshal(s)
	if err != nil {
		return err
	}
	r.meta[roleFilename] = b

	return r.local.SetMeta(roleFilename, b)
}

// getSignersInDB returns available signing interfaces, sorted by key ID.
//
// Only keys contained in the keys db are returned (i.e. local keys which have
// been revoked are omitted), except for the root role in which case all local
// keys are returned (revoked root keys still need to sign new root metadata so
// clients can verify the new root.json and update their keys db accordingly).
func (r *Repo) getSignersInDB(roleName string, db *verify.DB) ([]keys.Signer, error) {
	signers, err := r.local.GetSigners(roleName)
	if err != nil {
		return nil, err
	}

	if roleName == "root" {
		sorted := make([]keys.Signer, len(signers))
		copy(sorted, signers)
		sort.Sort(signer.ByIDs(sorted))
		return sorted, nil
	}

	role := db.GetRole(roleName)
	if role == nil {
		return nil, nil
	}
	if len(role.KeyIDs) == 0 {
		return nil, nil
	}

	signersInDB := []keys.Signer{}
	for _, s := range signers {
		for _, id := range s.PublicData().IDs() {
			if _, ok := role.KeyIDs[id]; ok {
				signersInDB = append(signersInDB, s)
			}
		}
	}

	sort.Sort(signer.ByIDs(signersInDB))

	return signersInDB, nil
}

// Used to retrieve the signable portion of the metadata when using an external signing tool.
func (r *Repo) SignedMeta(roleFilename string) (*data.Signed, error) {
	b, ok := r.meta[roleFilename]
	if !ok {
		return nil, ErrMissingMetadata{roleFilename}
	}
	s := &data.Signed{}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, err
	}
	return s, nil
}

// delegatorDBs returns a list of key DBs for all incoming delegations.
func (r *Repo) delegatorDBs(delegateeRole string) ([]*verify.DB, error) {
	delegatorDBs := []*verify.DB{}
	for metaName := range r.meta {
		if roles.IsTopLevelManifest(metaName) && metaName != "targets.json" {
			continue
		}
		roleName := strings.TrimSuffix(metaName, ".json")

		t, err := r.targets(roleName)
		if err != nil {
			return nil, err
		}

		if t.Delegations == nil {
			continue
		}

		delegatesToRole := false
		for _, d := range t.Delegations.Roles {
			if d.Name == delegateeRole {
				delegatesToRole = true
				break
			}
		}
		if !delegatesToRole {
			continue
		}

		db, err := verify.NewDBFromDelegations(t.Delegations)
		if err != nil {
			return nil, err
		}

		delegatorDBs = append(delegatorDBs, db)
	}

	return delegatorDBs, nil
}

// targetDelegationForPath finds the targets metadata for the role that should
// sign the given path. The final delegation that led to the returned target
// metadata is also returned.
//
// Since there may be multiple targets roles that are able to sign a specific
// path, we must choose which roles's metadata to return. If preferredRole is
// specified (non-empty string) and eligible to sign the given path by way of
// some delegation chain, targets metadata for that role is returned. If
// preferredRole is not specified (""), we return targets metadata for the
// final role visited in the depth-first delegation traversal.
func (r *Repo) targetDelegationForPath(path string, preferredRole string) (*data.Targets, *targets.Delegation, error) {
	topLevelKeysDB, err := r.topLevelKeysDB()
	if err != nil {
		return nil, nil, err
	}

	iterator, err := targets.NewDelegationsIterator(path, topLevelKeysDB)
	if err != nil {
		return nil, nil, err
	}
	d, ok := iterator.Next()
	if !ok {
		return nil, nil, ErrNoDelegatedTarget{Path: path}
	}

	for i := 0; i < defaultMaxDelegations; i++ {
		targetsMeta, err := r.targets(d.Delegatee.Name)
		if err != nil {
			return nil, nil, err
		}

		if preferredRole != "" && d.Delegatee.Name == preferredRole {
			// The preferredRole is eligible to sign for the given path, and we've
			// found its metadata. Return it.
			return targetsMeta, &d, nil
		}

		if targetsMeta.Delegations != nil && len(targetsMeta.Delegations.Roles) > 0 {
			db, err := verify.NewDBFromDelegations(targetsMeta.Delegations)
			if err != nil {
				return nil, nil, err
			}

			// Add delegations to the iterator that are eligible to sign for the
			// given path (there may be none).
			iterator.Add(targetsMeta.Delegations.Roles, d.Delegatee.Name, db)
		}

		next, ok := iterator.Next()
		if !ok { // No more roles to traverse.
			if preferredRole == "" {
				// No preferredRole was given, so return metadata for the final role in the traversal.
				return targetsMeta, &d, nil
			} else {
				// There are no more roles to traverse, so preferredRole is either an
				// invalid role, or not eligible to sign the given path.
				return nil, nil, ErrNoDelegatedTarget{Path: path}
			}
		}

		d = next
	}

	return nil, nil, ErrNoDelegatedTarget{Path: path}
}

func (r *Repo) AddTarget(path string, custom json.RawMessage) error {
	return r.AddTargets([]string{path}, custom)
}

func (r *Repo) AddTargetToPreferredRole(path string, custom json.RawMessage, preferredRole string) error {
	return r.AddTargetsToPreferredRole([]string{path}, custom, preferredRole)
}

func (r *Repo) AddTargets(paths []string, custom json.RawMessage) error {
	return r.AddTargetsToPreferredRole(paths, custom, "")
}

func (r *Repo) AddTargetsToPreferredRole(paths []string, custom json.RawMessage, preferredRole string) error {
	return r.AddTargetsWithExpiresToPreferredRole(paths, custom, data.DefaultExpires("targets"), preferredRole)
}

func (r *Repo) AddTargetsWithDigest(digest string, digestAlg string, length int64, path string, custom json.RawMessage) error {
	// TODO: Rename this to AddTargetWithDigest
	// https://github.com/theupdateframework/go-tuf/issues/242

	expires := data.DefaultExpires("targets")
	path = util.NormalizeTarget(path)

	targetsMeta, delegation, err := r.targetDelegationForPath(path, "")
	if err != nil {
		return err
	}
	// This is the targets role that needs to sign the target file.
	targetsRoleName := delegation.Delegatee.Name

	meta := data.TargetFileMeta{FileMeta: data.FileMeta{Length: length, Hashes: make(data.Hashes, 1)}}
	meta.Hashes[digestAlg], err = hex.DecodeString(digest)
	if err != nil {
		return err
	}

	// If custom is provided, set custom, otherwise maintain existing custom
	// metadata
	if len(custom) > 0 {
		meta.Custom = &custom
	} else if t, ok := targetsMeta.Targets[path]; ok {
		meta.Custom = t.Custom
	}

	// What does G2 mean? Copying and pasting this comment from elsewhere in this file.
	// G2 -> we no longer desire any readers to ever observe non-prefix targets.
	delete(targetsMeta.Targets, "/"+path)
	targetsMeta.Targets[path] = meta

	targetsMeta.Expires = expires.Round(time.Second)

	manifestName := targetsRoleName + ".json"
	if !r.local.FileIsStaged(manifestName) {
		targetsMeta.Version++
	}

	err = r.setMeta(manifestName, targetsMeta)
	if err != nil {
		return fmt.Errorf("error setting metadata for %q: %w", manifestName, err)
	}

	return nil
}

func (r *Repo) AddTargetWithExpires(path string, custom json.RawMessage, expires time.Time) error {
	return r.AddTargetsWithExpires([]string{path}, custom, expires)
}

func (r *Repo) AddTargetsWithExpires(paths []string, custom json.RawMessage, expires time.Time) error {
	return r.AddTargetsWithExpiresToPreferredRole(paths, custom, expires, "")
}

func (r *Repo) AddTargetWithExpiresToPreferredRole(path string, custom json.RawMessage, expires time.Time, preferredRole string) error {
	return r.AddTargetsWithExpiresToPreferredRole([]string{path}, custom, expires, preferredRole)
}

// AddTargetsWithExpiresToPreferredRole signs the staged targets at `paths`.
//
// If preferredRole is not the empty string, the target is added to the given
// role's manifest if delegations allow it. If delegations do not allow the
// preferredRole to sign the given path, an error is returned.
func (r *Repo) AddTargetsWithExpiresToPreferredRole(paths []string, custom json.RawMessage, expires time.Time, preferredRole string) error {
	if !validExpires(expires) {
		return ErrInvalidExpires{expires}
	}

	normalizedPaths := make([]string, len(paths))
	for i, path := range paths {
		normalizedPaths[i] = util.NormalizeTarget(path)
	}

	// As we iterate through staged targets files, we accumulate changes to their
	// corresponding targets metadata.
	updatedTargetsMeta := map[string]*data.Targets{}

	if err := r.local.WalkStagedTargets(normalizedPaths, func(path string, target io.Reader) (err error) {
		originalMeta, delegation, err := r.targetDelegationForPath(path, preferredRole)
		if err != nil {
			return err
		}

		// This is the targets role that needs to sign the target file.
		targetsRoleName := delegation.Delegatee.Name

		targetsMeta := originalMeta
		if tm, ok := updatedTargetsMeta[targetsRoleName]; ok {
			// Metadata in updatedTargetsMeta overrides staged/commited metadata.
			targetsMeta = tm
		}

		fileMeta, err := util.GenerateTargetFileMeta(target, r.hashAlgorithms...)
		if err != nil {
			return err
		}

		// If we have custom metadata, set it, otherwise maintain
		// existing metadata if present
		if len(custom) > 0 {
			fileMeta.Custom = &custom
		} else if tf, ok := targetsMeta.Targets[path]; ok {
			fileMeta.Custom = tf.Custom
		}

		// G2 -> we no longer desire any readers to ever observe non-prefix targets.
		delete(targetsMeta.Targets, "/"+path)
		targetsMeta.Targets[path] = fileMeta

		updatedTargetsMeta[targetsRoleName] = targetsMeta

		return nil
	}); err != nil {
		return err
	}

	if len(updatedTargetsMeta) == 0 {
		// This is potentially unexpected behavior kept for backwards compatibility.
		// See https://github.com/theupdateframework/go-tuf/issues/243
		t, err := r.topLevelTargets()
		if err != nil {
			return err
		}

		updatedTargetsMeta["targets"] = t
	}

	exp := expires.Round(time.Second)
	for roleName, targetsMeta := range updatedTargetsMeta {
		targetsMeta.Expires = exp

		manifestName := roleName + ".json"
		if !r.local.FileIsStaged(manifestName) {
			targetsMeta.Version++
		}

		err := r.setMeta(manifestName, targetsMeta)
		if err != nil {
			return fmt.Errorf("error setting metadata for %q: %w", manifestName, err)
		}
	}

	return nil
}

func (r *Repo) RemoveTarget(path string) error {
	return r.RemoveTargets([]string{path})
}

func (r *Repo) RemoveTargets(paths []string) error {
	return r.RemoveTargetsWithExpires(paths, data.DefaultExpires("targets"))
}

func (r *Repo) RemoveTargetWithExpires(path string, expires time.Time) error {
	return r.RemoveTargetsWithExpires([]string{path}, expires)
}

// If paths is empty, all targets will be removed.
func (r *Repo) RemoveTargetsWithExpires(paths []string, expires time.Time) error {
	if !validExpires(expires) {
		return ErrInvalidExpires{expires}
	}

	for metaName := range r.meta {
		if metaName != "targets.json" && !roles.IsDelegatedTargetsManifest(metaName) {
			continue
		}

		err := r.removeTargetsWithExpiresFromMeta(metaName, paths, expires)
		if err != nil {
			return fmt.Errorf("could not remove %v from %v: %w", paths, metaName, err)
		}
	}

	return nil
}

func (r *Repo) removeTargetsWithExpiresFromMeta(metaName string, paths []string, expires time.Time) error {
	roleName := strings.TrimSuffix(metaName, ".json")
	t, err := r.targets(roleName)
	if err != nil {
		return err
	}
	removed_targets := []string{}
	if len(paths) == 0 {
		for rt := range t.Targets {
			removed_targets = append(removed_targets, rt)
		}
		t.Targets = make(data.TargetFiles)
	} else {
		removed := false
		for _, path := range paths {
			path = util.NormalizeTarget(path)
			if _, ok := t.Targets[path]; !ok {
				r.logger.Printf("[%v] The following target is not present: %v\n", metaName, path)
				continue
			}
			removed = true
			// G2 -> we no longer desire any readers to ever observe non-prefix targets.
			delete(t.Targets, "/"+path)
			delete(t.Targets, path)
			removed_targets = append(removed_targets, path)
		}
		if !removed {
			return nil
		}
	}
	t.Expires = expires.Round(time.Second)
	if !r.local.FileIsStaged(metaName) {
		t.Version++
	}

	err = r.setMeta(metaName, t)
	if err == nil {
		r.logger.Printf("[%v] Removed targets:\n", metaName)
		for _, v := range removed_targets {
			r.logger.Println("*", v)
		}
		if len(t.Targets) != 0 {
			r.logger.Printf("[%v] Added/staged targets:\n", metaName)
			for k := range t.Targets {
				r.logger.Println("*", k)
			}
		} else {
			r.logger.Printf("[%v] There are no added/staged targets\n", metaName)
		}
	}
	return err
}

func (r *Repo) Snapshot() error {
	return r.SnapshotWithExpires(data.DefaultExpires("snapshot"))
}

func (r *Repo) snapshotMetadata() []string {
	ret := []string{"targets.json"}

	for name := range r.meta {
		if !roles.IsVersionedManifest(name) &&
			roles.IsDelegatedTargetsManifest(name) {
			ret = append(ret, name)
		}
	}

	return ret
}

func (r *Repo) SnapshotWithExpires(expires time.Time) error {
	if !validExpires(expires) {
		return ErrInvalidExpires{expires}
	}

	snapshot, err := r.snapshot()
	if err != nil {
		return err
	}

	// Verify root metadata before verifying signatures on role metadata.
	if err := r.verifySignatures("root.json"); err != nil {
		return err
	}

	for _, metaName := range r.snapshotMetadata() {
		if err := r.verifySignatures(metaName); err != nil {
			return err
		}
		var err error
		snapshot.Meta[metaName], err = r.snapshotFileMeta(metaName)
		if err != nil {
			return err
		}
	}
	snapshot.Expires = expires.Round(time.Second)
	if !r.local.FileIsStaged("snapshot.json") {
		snapshot.Version++
	}
	err = r.setMeta("snapshot.json", snapshot)
	if err == nil {
		r.logger.Println("Staged snapshot.json metadata with expiration date:", snapshot.Expires)
	}
	return err
}

func (r *Repo) Timestamp() error {
	return r.TimestampWithExpires(data.DefaultExpires("timestamp"))
}

func (r *Repo) TimestampWithExpires(expires time.Time) error {
	if !validExpires(expires) {
		return ErrInvalidExpires{expires}
	}

	if err := r.verifySignatures("snapshot.json"); err != nil {
		return err
	}
	timestamp, err := r.timestamp()
	if err != nil {
		return err
	}
	timestamp.Meta["snapshot.json"], err = r.timestampFileMeta("snapshot.json")
	if err != nil {
		return err
	}
	timestamp.Expires = expires.Round(time.Second)
	if !r.local.FileIsStaged("timestamp.json") {
		timestamp.Version++
	}

	err = r.setMeta("timestamp.json", timestamp)
	if err == nil {
		r.logger.Println("Staged timestamp.json metadata with expiration date:", timestamp.Expires)
	}
	return err
}

func (r *Repo) fileVersions() (map[string]int64, error) {
	versions := make(map[string]int64)

	for fileName := range r.meta {
		if roles.IsVersionedManifest(fileName) {
			continue
		}

		roleName := strings.TrimSuffix(fileName, ".json")

		var version int64

		switch roleName {
		case "root":
			root, err := r.root()
			if err != nil {
				return nil, err
			}
			version = root.Version
		case "snapshot":
			snapshot, err := r.snapshot()
			if err != nil {
				return nil, err
			}
			version = snapshot.Version
		case "timestamp":
			continue
		default:
			// Targets or delegated targets manifest.
			targets, err := r.targets(roleName)
			if err != nil {
				return nil, err
			}

			version = targets.Version
		}

		versions[fileName] = version
	}

	return versions, nil
}

func (r *Repo) fileHashes() (map[string]data.Hashes, error) {
	hashes := make(map[string]data.Hashes)

	for fileName := range r.meta {
		if roles.IsVersionedManifest(fileName) {
			continue
		}

		roleName := strings.TrimSuffix(fileName, ".json")

		switch roleName {
		case "snapshot":
			timestamp, err := r.timestamp()
			if err != nil {
				return nil, err
			}

			if m, ok := timestamp.Meta[fileName]; ok {
				hashes[fileName] = m.Hashes
			}
		case "timestamp":
			continue
		default:
			snapshot, err := r.snapshot()
			if err != nil {
				return nil, err
			}
			if m, ok := snapshot.Meta[fileName]; ok {
				hashes[fileName] = m.Hashes
			}

			if roleName != "root" {
				// Scalability issue: Commit/fileHashes loads all targets metadata into memory
				// https://github.com/theupdateframework/go-tuf/issues/245
				t, err := r.targets(roleName)
				if err != nil {
					return nil, err
				}
				for name, m := range t.Targets {
					hashes[path.Join("targets", name)] = m.Hashes
				}
			}

		}

	}

	return hashes, nil
}

func (r *Repo) Commit() error {
	// check we have all the metadata
	for _, name := range topLevelMetadata {
		if _, ok := r.meta[name]; !ok {
			return ErrMissingMetadata{name}
		}
	}

	// check roles are valid
	root, err := r.root()
	if err != nil {
		return err
	}
	for name, role := range root.Roles {
		if len(role.KeyIDs) < role.Threshold {
			return ErrNotEnoughKeys{name, len(role.KeyIDs), role.Threshold}
		}
	}

	// verify hashes in snapshot.json are up to date
	snapshot, err := r.snapshot()
	if err != nil {
		return err
	}
	for _, name := range r.snapshotMetadata() {
		expected, ok := snapshot.Meta[name]
		if !ok {
			return fmt.Errorf("tuf: snapshot.json missing hash for %s", name)
		}
		actual, err := r.snapshotFileMeta(name)
		if err != nil {
			return err
		}
		if err := util.SnapshotFileMetaEqual(actual, expected); err != nil {
			return fmt.Errorf("tuf: invalid %s in snapshot.json: %s", name, err)
		}
	}

	// verify hashes in timestamp.json are up to date
	timestamp, err := r.timestamp()
	if err != nil {
		return err
	}
	snapshotMeta, err := r.timestampFileMeta("snapshot.json")
	if err != nil {
		return err
	}
	if err := util.TimestampFileMetaEqual(snapshotMeta, timestamp.Meta["snapshot.json"]); err != nil {
		return fmt.Errorf("tuf: invalid snapshot.json in timestamp.json: %s", err)
	}

	for _, name := range topLevelMetadata {
		if err := r.verifySignatures(name); err != nil {
			return err
		}
	}

	versions, err := r.fileVersions()
	if err != nil {
		return err
	}
	hashes, err := r.fileHashes()
	if err != nil {
		return err
	}

	err = r.local.Commit(root.ConsistentSnapshot, versions, hashes)
	if err == nil {
		r.logger.Println("Committed successfully")
	}
	return err
}

func (r *Repo) Clean() error {
	err := r.local.Clean()
	if err == nil {
		r.logger.Println("Removed all staged metadata and target files")
	}
	return err
}

func (r *Repo) verifySignatures(metaFilename string) error {
	s, err := r.SignedMeta(metaFilename)
	if err != nil {
		return err
	}

	role := strings.TrimSuffix(metaFilename, ".json")

	dbs, err := r.dbsForRole(role)
	if err != nil {
		return err
	}

	for _, db := range dbs {
		if err := db.Verify(s, role, 0); err != nil {
			return ErrInsufficientSignatures{metaFilename, err}
		}
	}

	return nil
}

func (r *Repo) snapshotFileMeta(roleFilename string) (data.SnapshotFileMeta, error) {
	b, ok := r.meta[roleFilename]
	if !ok {
		return data.SnapshotFileMeta{}, ErrMissingMetadata{roleFilename}
	}
	return util.GenerateSnapshotFileMeta(bytes.NewReader(b), r.hashAlgorithms...)
}

func (r *Repo) timestampFileMeta(roleFilename string) (data.TimestampFileMeta, error) {
	b, ok := r.meta[roleFilename]
	if !ok {
		return data.TimestampFileMeta{}, ErrMissingMetadata{roleFilename}
	}
	return util.GenerateTimestampFileMeta(bytes.NewReader(b), r.hashAlgorithms...)
}

func (r *Repo) Payload(roleFilename string) ([]byte, error) {
	s, err := r.SignedMeta(roleFilename)
	if err != nil {
		return nil, err
	}

	p, err := cjson.EncodeCanonical(s.Signed)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *Repo) CheckRoleUnexpired(role string, validAt time.Time) error {
	var expires time.Time
	switch role {
	case "root":
		root, err := r.root()
		if err != nil {
			return err
		}
		expires = root.Expires
	case "snapshot":
		snapshot, err := r.snapshot()
		if err != nil {
			return err
		}
		expires = snapshot.Expires
	case "timestamp":
		timestamp, err := r.timestamp()
		if err != nil {
			return err
		}
		expires = timestamp.Expires
	case "targets":
		targets, err := r.topLevelTargets()
		if err != nil {
			return err
		}
		expires = targets.Expires
	default:
		return fmt.Errorf("invalid role: %s", role)
	}
	if expires.Before(validAt) || expires.Equal(validAt) {
		return fmt.Errorf("role expired on: %s", expires)
	}
	return nil
}

// GetMeta returns the underlying meta file map from the store.
func (r *Repo) GetMeta() (map[string]json.RawMessage, error) {
	return r.local.GetMeta()
}
