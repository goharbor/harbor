//Package client implements everything required for interacting with a Notary repository.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"

	canonicaljson "github.com/docker/go/canonical/json"
	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/client/changelist"
	"github.com/theupdateframework/notary/cryptoservice"
	store "github.com/theupdateframework/notary/storage"
	"github.com/theupdateframework/notary/trustpinning"
	"github.com/theupdateframework/notary/tuf"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/signed"
	"github.com/theupdateframework/notary/tuf/utils"
)

const (
	tufDir = "tuf"

	// SignWithAllOldVersions is a sentinel constant for LegacyVersions flag
	SignWithAllOldVersions = -1
)

func init() {
	data.SetDefaultExpiryTimes(data.NotaryDefaultExpiries)
}

// repository stores all the information needed to operate on a notary repository.
type repository struct {
	baseDir        string
	gun            data.GUN
	baseURL        string
	changelist     changelist.Changelist
	cache          store.MetadataStore
	remoteStore    store.RemoteStore
	cryptoService  signed.CryptoService
	tufRepo        *tuf.Repo
	invalid        *tuf.Repo // known data that was parsable but deemed invalid
	roundTrip      http.RoundTripper
	trustPinning   trustpinning.TrustPinConfig
	LegacyVersions int // number of versions back to fetch roots to sign with
}

// NewFileCachedRepository is a wrapper for NewRepository that initializes
// a file cache from the provided repository, local config information and a crypto service.
// It also retrieves the remote store associated to the base directory under where all the
// trust files will be stored and the specified GUN.
//
// In case of a nil RoundTripper, a default offline store is used instead.
func NewFileCachedRepository(baseDir string, gun data.GUN, baseURL string, rt http.RoundTripper,
	retriever notary.PassRetriever, trustPinning trustpinning.TrustPinConfig) (Repository, error) {

	cache, err := store.NewFileStore(
		filepath.Join(baseDir, tufDir, filepath.FromSlash(gun.String()), "metadata"),
		"json",
	)
	if err != nil {
		return nil, err
	}

	keyStores, err := getKeyStores(baseDir, retriever)
	if err != nil {
		return nil, err
	}

	cryptoService := cryptoservice.NewCryptoService(keyStores...)

	remoteStore, err := getRemoteStore(baseURL, gun, rt)
	if err != nil {
		// baseURL is syntactically invalid
		return nil, err
	}

	cl, err := changelist.NewFileChangelist(filepath.Join(
		filepath.Join(baseDir, tufDir, filepath.FromSlash(gun.String()), "changelist"),
	))
	if err != nil {
		return nil, err
	}

	return NewRepository(baseDir, gun, baseURL, remoteStore, cache, trustPinning, cryptoService, cl)
}

// NewRepository is the base method that returns a new notary repository.
// It takes the base directory under where all the trust files will be stored
// (This is normally defaults to "~/.notary" or "~/.docker/trust" when enabling
// docker content trust).
// It expects an initialized cache. In case of a nil remote store, a default
// offline store is used.
func NewRepository(baseDir string, gun data.GUN, baseURL string, remoteStore store.RemoteStore, cache store.MetadataStore,
	trustPinning trustpinning.TrustPinConfig, cryptoService signed.CryptoService, cl changelist.Changelist) (Repository, error) {

	// Repo's remote store is either a valid remote store or an OfflineStore
	if remoteStore == nil {
		remoteStore = store.OfflineStore{}
	}

	if cache == nil {
		return nil, fmt.Errorf("got an invalid cache (nil metadata store)")
	}

	nRepo := &repository{
		gun:            gun,
		baseURL:        baseURL,
		baseDir:        baseDir,
		changelist:     cl,
		cache:          cache,
		remoteStore:    remoteStore,
		cryptoService:  cryptoService,
		trustPinning:   trustPinning,
		LegacyVersions: 0, // By default, don't sign with legacy roles
	}

	return nRepo, nil
}

// GetGUN is a getter for the GUN object from a Repository
func (r *repository) GetGUN() data.GUN {
	return r.gun
}

// Target represents a simplified version of the data TUF operates on, so external
// applications don't have to depend on TUF data types.
type Target struct {
	Name   string                    // the name of the target
	Hashes data.Hashes               // the hash of the target
	Length int64                     // the size in bytes of the target
	Custom *canonicaljson.RawMessage // the custom data provided to describe the file at TARGETPATH
}

// TargetWithRole represents a Target that exists in a particular role - this is
// produced by ListTargets and GetTargetByName
type TargetWithRole struct {
	Target
	Role data.RoleName
}

// NewTarget is a helper method that returns a Target
func NewTarget(targetName, targetPath string, targetCustom *canonicaljson.RawMessage) (*Target, error) {
	b, err := ioutil.ReadFile(targetPath)
	if err != nil {
		return nil, err
	}

	meta, err := data.NewFileMeta(bytes.NewBuffer(b), data.NotaryDefaultHashes...)
	if err != nil {
		return nil, err
	}

	return &Target{Name: targetName, Hashes: meta.Hashes, Length: meta.Length, Custom: targetCustom}, nil
}

// rootCertKey generates the corresponding certificate for the private key given the privKey and repo's GUN
func rootCertKey(gun data.GUN, privKey data.PrivateKey) (data.PublicKey, error) {
	// Hard-coded policy: the generated certificate expires in 10 years.
	startTime := time.Now()
	cert, err := cryptoservice.GenerateCertificate(
		privKey, gun, startTime, startTime.Add(notary.Year*10))
	if err != nil {
		return nil, err
	}

	x509PublicKey := utils.CertToKey(cert)
	if x509PublicKey == nil {
		return nil, fmt.Errorf("cannot generate public key from private key with id: %v and algorithm: %v", privKey.ID(), privKey.Algorithm())
	}

	return x509PublicKey, nil
}

// GetCryptoService is the getter for the repository's CryptoService
func (r *repository) GetCryptoService() signed.CryptoService {
	return r.cryptoService
}

// initialize initializes the notary repository with a set of rootkeys, root certificates and roles.
func (r *repository) initialize(rootKeyIDs []string, rootCerts []data.PublicKey, serverManagedRoles ...data.RoleName) error {

	// currently we only support server managing timestamps and snapshots, and
	// nothing else - timestamps are always managed by the server, and implicit
	// (do not have to be passed in as part of `serverManagedRoles`, so that
	// the API of Initialize doesn't change).
	var serverManagesSnapshot bool
	locallyManagedKeys := []data.RoleName{
		data.CanonicalTargetsRole,
		data.CanonicalSnapshotRole,
		// root is also locally managed, but that should have been created
		// already
	}
	remotelyManagedKeys := []data.RoleName{data.CanonicalTimestampRole}
	for _, role := range serverManagedRoles {
		switch role {
		case data.CanonicalTimestampRole:
			continue // timestamp is already in the right place
		case data.CanonicalSnapshotRole:
			// because we put Snapshot last
			locallyManagedKeys = []data.RoleName{data.CanonicalTargetsRole}
			remotelyManagedKeys = append(
				remotelyManagedKeys, data.CanonicalSnapshotRole)
			serverManagesSnapshot = true
		default:
			return ErrInvalidRemoteRole{Role: role}
		}
	}

	// gets valid public keys corresponding to the rootKeyIDs or generate if necessary
	var publicKeys []data.PublicKey
	var err error
	if len(rootCerts) == 0 {
		publicKeys, err = r.createNewPublicKeyFromKeyIDs(rootKeyIDs)
	} else {
		publicKeys, err = r.publicKeysOfKeyIDs(rootKeyIDs, rootCerts)
	}
	if err != nil {
		return err
	}

	//initialize repo with public keys
	rootRole, targetsRole, snapshotRole, timestampRole, err := r.initializeRoles(
		publicKeys,
		locallyManagedKeys,
		remotelyManagedKeys,
	)
	if err != nil {
		return err
	}

	r.tufRepo = tuf.NewRepo(r.GetCryptoService())

	if err := r.tufRepo.InitRoot(
		rootRole,
		timestampRole,
		snapshotRole,
		targetsRole,
		false,
	); err != nil {
		logrus.Debug("Error on InitRoot: ", err.Error())
		return err
	}
	if _, err := r.tufRepo.InitTargets(data.CanonicalTargetsRole); err != nil {
		logrus.Debug("Error on InitTargets: ", err.Error())
		return err
	}
	if err := r.tufRepo.InitSnapshot(); err != nil {
		logrus.Debug("Error on InitSnapshot: ", err.Error())
		return err
	}

	return r.saveMetadata(serverManagesSnapshot)
}

// createNewPublicKeyFromKeyIDs generates a set of public keys corresponding to the given list of
// key IDs existing in the repository's CryptoService.
// the public keys returned are ordered to correspond to the keyIDs
func (r *repository) createNewPublicKeyFromKeyIDs(keyIDs []string) ([]data.PublicKey, error) {
	publicKeys := []data.PublicKey{}

	privKeys, err := getAllPrivKeys(keyIDs, r.GetCryptoService())
	if err != nil {
		return nil, err
	}

	for _, privKey := range privKeys {
		rootKey, err := rootCertKey(r.gun, privKey)
		if err != nil {
			return nil, err
		}
		publicKeys = append(publicKeys, rootKey)
	}
	return publicKeys, nil
}

// publicKeysOfKeyIDs confirms that the public key and private keys (by Key IDs) forms valid, strictly ordered key pairs
// (eg. keyIDs[0] must match pubKeys[0] and keyIDs[1] must match certs[1] and so on).
// Or throw error when they mismatch.
func (r *repository) publicKeysOfKeyIDs(keyIDs []string, pubKeys []data.PublicKey) ([]data.PublicKey, error) {
	if len(keyIDs) != len(pubKeys) {
		err := fmt.Errorf("require matching number of keyIDs and public keys but got %d IDs and %d public keys", len(keyIDs), len(pubKeys))
		return nil, err
	}

	if err := matchKeyIdsWithPubKeys(r, keyIDs, pubKeys); err != nil {
		return nil, fmt.Errorf("could not obtain public key from IDs: %v", err)
	}
	return pubKeys, nil
}

// matchKeyIdsWithPubKeys validates that the private keys (represented by their IDs) and the public keys
// forms matching key pairs
func matchKeyIdsWithPubKeys(r *repository, ids []string, pubKeys []data.PublicKey) error {
	for i := 0; i < len(ids); i++ {
		privKey, _, err := r.GetCryptoService().GetPrivateKey(ids[i])
		if err != nil {
			return fmt.Errorf("could not get the private key matching id %v: %v", ids[i], err)
		}

		pubKey := pubKeys[i]
		err = signed.VerifyPublicKeyMatchesPrivateKey(privKey, pubKey)
		if err != nil {
			return err
		}
	}
	return nil
}

// Initialize creates a new repository by using rootKey as the root Key for the
// TUF repository. The server must be reachable (and is asked to generate a
// timestamp key and possibly other serverManagedRoles), but the created repository
// result is only stored on local disk, not published to the server. To do that,
// use r.Publish() eventually.
func (r *repository) Initialize(rootKeyIDs []string, serverManagedRoles ...data.RoleName) error {
	return r.initialize(rootKeyIDs, nil, serverManagedRoles...)
}

type errKeyNotFound struct{}

func (errKeyNotFound) Error() string {
	return fmt.Sprintf("cannot find matching private key id")
}

// keyExistsInList returns the id of the private key in ids that matches the public key
// otherwise return empty string
func keyExistsInList(cert data.PublicKey, ids map[string]bool) error {
	pubKeyID, err := utils.CanonicalKeyID(cert)
	if err != nil {
		return fmt.Errorf("failed to obtain the public key id from the given certificate: %v", err)
	}
	if _, ok := ids[pubKeyID]; ok {
		return nil
	}
	return errKeyNotFound{}
}

// InitializeWithCertificate initializes the repository with root keys and their corresponding certificates
func (r *repository) InitializeWithCertificate(rootKeyIDs []string, rootCerts []data.PublicKey,
	serverManagedRoles ...data.RoleName) error {

	// If we explicitly pass in certificate(s) but not key, then look keys up using certificate
	if len(rootKeyIDs) == 0 && len(rootCerts) != 0 {
		rootKeyIDs = []string{}
		availableRootKeyIDs := make(map[string]bool)
		for _, k := range r.GetCryptoService().ListKeys(data.CanonicalRootRole) {
			availableRootKeyIDs[k] = true
		}

		for _, cert := range rootCerts {
			if err := keyExistsInList(cert, availableRootKeyIDs); err != nil {
				return fmt.Errorf("error initializing repository with certificate: %v", err)
			}
			keyID, _ := utils.CanonicalKeyID(cert)
			rootKeyIDs = append(rootKeyIDs, keyID)
		}
	}
	return r.initialize(rootKeyIDs, rootCerts, serverManagedRoles...)
}

func (r *repository) initializeRoles(rootKeys []data.PublicKey, localRoles, remoteRoles []data.RoleName) (
	root, targets, snapshot, timestamp data.BaseRole, err error) {
	root = data.NewBaseRole(
		data.CanonicalRootRole,
		notary.MinThreshold,
		rootKeys...,
	)

	// we want to create all the local keys first so we don't have to
	// make unnecessary network calls
	for _, role := range localRoles {
		// This is currently hardcoding the keys to ECDSA.
		var key data.PublicKey
		key, err = r.GetCryptoService().Create(role, r.gun, data.ECDSAKey)
		if err != nil {
			return
		}
		switch role {
		case data.CanonicalSnapshotRole:
			snapshot = data.NewBaseRole(
				role,
				notary.MinThreshold,
				key,
			)
		case data.CanonicalTargetsRole:
			targets = data.NewBaseRole(
				role,
				notary.MinThreshold,
				key,
			)
		}
	}

	remote := r.getRemoteStore()

	for _, role := range remoteRoles {
		// This key is generated by the remote server.
		var key data.PublicKey
		key, err = getRemoteKey(role, remote)
		if err != nil {
			return
		}
		logrus.Debugf("got remote %s %s key with keyID: %s",
			role, key.Algorithm(), key.ID())
		switch role {
		case data.CanonicalSnapshotRole:
			snapshot = data.NewBaseRole(
				role,
				notary.MinThreshold,
				key,
			)
		case data.CanonicalTimestampRole:
			timestamp = data.NewBaseRole(
				role,
				notary.MinThreshold,
				key,
			)
		}
	}
	return root, targets, snapshot, timestamp, nil
}

// adds a TUF Change template to the given roles
func addChange(cl changelist.Changelist, c changelist.Change, roles ...data.RoleName) error {
	if len(roles) == 0 {
		roles = []data.RoleName{data.CanonicalTargetsRole}
	}

	var changes []changelist.Change
	for _, role := range roles {
		// Ensure we can only add targets to the CanonicalTargetsRole,
		// or a Delegation role (which is <CanonicalTargetsRole>/something else)
		if role != data.CanonicalTargetsRole && !data.IsDelegation(role) && !data.IsWildDelegation(role) {
			return data.ErrInvalidRole{
				Role:   role,
				Reason: "cannot add targets to this role",
			}
		}

		changes = append(changes, changelist.NewTUFChange(
			c.Action(),
			role,
			c.Type(),
			c.Path(),
			c.Content(),
		))
	}

	for _, c := range changes {
		if err := cl.Add(c); err != nil {
			return err
		}
	}
	return nil
}

// AddTarget creates new changelist entries to add a target to the given roles
// in the repository when the changelist gets applied at publish time.
// If roles are unspecified, the default role is "targets"
func (r *repository) AddTarget(target *Target, roles ...data.RoleName) error {
	if len(target.Hashes) == 0 {
		return fmt.Errorf("no hashes specified for target \"%s\"", target.Name)
	}
	logrus.Debugf("Adding target \"%s\" with sha256 \"%x\" and size %d bytes.\n", target.Name, target.Hashes["sha256"], target.Length)

	meta := data.FileMeta{Length: target.Length, Hashes: target.Hashes, Custom: target.Custom}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	template := changelist.NewTUFChange(
		changelist.ActionCreate, "", changelist.TypeTargetsTarget,
		target.Name, metaJSON)
	return addChange(r.changelist, template, roles...)
}

// RemoveTarget creates new changelist entries to remove a target from the given
// roles in the repository when the changelist gets applied at publish time.
// If roles are unspecified, the default role is "target".
func (r *repository) RemoveTarget(targetName string, roles ...data.RoleName) error {
	logrus.Debugf("Removing target \"%s\"", targetName)
	template := changelist.NewTUFChange(changelist.ActionDelete, "",
		changelist.TypeTargetsTarget, targetName, nil)
	return addChange(r.changelist, template, roles...)
}

// ListTargets lists all targets for the current repository. The list of
// roles should be passed in order from highest to lowest priority.
//
// IMPORTANT: if you pass a set of roles such as [ "targets/a", "targets/x"
// "targets/a/b" ], even though "targets/a/b" is part of the "targets/a" subtree
// its entries will be strictly shadowed by those in other parts of the "targets/a"
// subtree and also the "targets/x" subtree, as we will defer parsing it until
// we explicitly reach it in our iteration of the provided list of roles.
func (r *repository) ListTargets(roles ...data.RoleName) ([]*TargetWithRole, error) {
	if err := r.Update(false); err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		roles = []data.RoleName{data.CanonicalTargetsRole}
	}
	targets := make(map[string]*TargetWithRole)
	for _, role := range roles {
		// Define an array of roles to skip for this walk (see IMPORTANT comment above)
		skipRoles := utils.RoleNameSliceRemove(roles, role)

		// Define a visitor function to populate the targets map in priority order
		listVisitorFunc := func(tgt *data.SignedTargets, validRole data.DelegationRole) interface{} {
			// We found targets so we should try to add them to our targets map
			for targetName, targetMeta := range tgt.Signed.Targets {
				// Follow the priority by not overriding previously set targets
				// and check that this path is valid with this role
				if _, ok := targets[targetName]; ok || !validRole.CheckPaths(targetName) {
					continue
				}
				targets[targetName] = &TargetWithRole{
					Target: Target{
						Name:   targetName,
						Hashes: targetMeta.Hashes,
						Length: targetMeta.Length,
						Custom: targetMeta.Custom,
					},
					Role: validRole.Name,
				}
			}
			return nil
		}

		r.tufRepo.WalkTargets("", role, listVisitorFunc, skipRoles...)
	}

	var targetList []*TargetWithRole
	for _, v := range targets {
		targetList = append(targetList, v)
	}

	return targetList, nil
}

// GetTargetByName returns a target by the given name. If no roles are passed
// it uses the targets role and does a search of the entire delegation
// graph, finding the first entry in a breadth first search of the delegations.
// If roles are passed, they should be passed in descending priority and
// the target entry found in the subtree of the highest priority role
// will be returned.
// See the IMPORTANT section on ListTargets above. Those roles also apply here.
func (r *repository) GetTargetByName(name string, roles ...data.RoleName) (*TargetWithRole, error) {
	if err := r.Update(false); err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		roles = append(roles, data.CanonicalTargetsRole)
	}
	var resultMeta data.FileMeta
	var resultRoleName data.RoleName
	var foundTarget bool
	for _, role := range roles {
		// Define an array of roles to skip for this walk (see IMPORTANT comment above)
		skipRoles := utils.RoleNameSliceRemove(roles, role)

		// Define a visitor function to find the specified target
		getTargetVisitorFunc := func(tgt *data.SignedTargets, validRole data.DelegationRole) interface{} {
			if tgt == nil {
				return nil
			}
			// We found the target and validated path compatibility in our walk,
			// so we should stop our walk and set the resultMeta and resultRoleName variables
			if resultMeta, foundTarget = tgt.Signed.Targets[name]; foundTarget {
				resultRoleName = validRole.Name
				return tuf.StopWalk{}
			}
			return nil
		}
		// Check that we didn't error, and that we assigned to our target
		if err := r.tufRepo.WalkTargets(name, role, getTargetVisitorFunc, skipRoles...); err == nil && foundTarget {
			return &TargetWithRole{Target: Target{Name: name, Hashes: resultMeta.Hashes, Length: resultMeta.Length, Custom: resultMeta.Custom}, Role: resultRoleName}, nil
		}
	}
	return nil, ErrNoSuchTarget(name)

}

// TargetSignedStruct is a struct that contains a Target, the role it was found in, and the list of signatures for that role
type TargetSignedStruct struct {
	Role       data.DelegationRole
	Target     Target
	Signatures []data.Signature
}

//ErrNoSuchTarget is returned when no valid trust data is found.
type ErrNoSuchTarget string

func (f ErrNoSuchTarget) Error() string {
	return fmt.Sprintf("No valid trust data for %s", string(f))
}

// GetAllTargetMetadataByName searches the entire delegation role tree to find the specified target by name for all
// roles, and returns a list of TargetSignedStructs for each time it finds the specified target.
// If given an empty string for a target name, it will return back all targets signed into the repository in every role
func (r *repository) GetAllTargetMetadataByName(name string) ([]TargetSignedStruct, error) {
	if err := r.Update(false); err != nil {
		return nil, err
	}

	var targetInfoList []TargetSignedStruct

	// Define a visitor function to find the specified target
	getAllTargetInfoByNameVisitorFunc := func(tgt *data.SignedTargets, validRole data.DelegationRole) interface{} {
		if tgt == nil {
			return nil
		}
		// We found a target and validated path compatibility in our walk,
		// so add it to our list if we have a match
		// if we have an empty name, add all targets, else check if we have it
		var targetMetaToAdd data.Files
		if name == "" {
			targetMetaToAdd = tgt.Signed.Targets
		} else {
			if meta, ok := tgt.Signed.Targets[name]; ok {
				targetMetaToAdd = data.Files{name: meta}
			}
		}

		for targetName, resultMeta := range targetMetaToAdd {
			targetInfo := TargetSignedStruct{
				Role:       validRole,
				Target:     Target{Name: targetName, Hashes: resultMeta.Hashes, Length: resultMeta.Length, Custom: resultMeta.Custom},
				Signatures: tgt.Signatures,
			}
			targetInfoList = append(targetInfoList, targetInfo)
		}
		// continue walking to all child roles
		return nil
	}

	// Check that we didn't error, and that we found the target at least once
	if err := r.tufRepo.WalkTargets(name, "", getAllTargetInfoByNameVisitorFunc); err != nil {
		return nil, err
	}
	if len(targetInfoList) == 0 {
		return nil, ErrNoSuchTarget(name)
	}
	return targetInfoList, nil
}

// GetChangelist returns the list of the repository's unpublished changes
func (r *repository) GetChangelist() (changelist.Changelist, error) {
	return r.changelist, nil
}

// getRemoteStore returns the remoteStore of a repository if valid or
// or an OfflineStore otherwise
func (r *repository) getRemoteStore() store.RemoteStore {
	if r.remoteStore != nil {
		return r.remoteStore
	}

	r.remoteStore = &store.OfflineStore{}

	return r.remoteStore
}

// RoleWithSignatures is a Role with its associated signatures
type RoleWithSignatures struct {
	Signatures []data.Signature
	data.Role
}

// ListRoles returns a list of RoleWithSignatures objects for this repo
// This represents the latest metadata for each role in this repo
func (r *repository) ListRoles() ([]RoleWithSignatures, error) {
	// Update to latest repo state
	if err := r.Update(false); err != nil {
		return nil, err
	}

	// Get all role info from our updated keysDB, can be empty
	roles := r.tufRepo.GetAllLoadedRoles()

	var roleWithSigs []RoleWithSignatures

	// Populate RoleWithSignatures with Role from keysDB and signatures from TUF metadata
	for _, role := range roles {
		roleWithSig := RoleWithSignatures{Role: *role, Signatures: nil}
		switch role.Name {
		case data.CanonicalRootRole:
			roleWithSig.Signatures = r.tufRepo.Root.Signatures
		case data.CanonicalTargetsRole:
			roleWithSig.Signatures = r.tufRepo.Targets[data.CanonicalTargetsRole].Signatures
		case data.CanonicalSnapshotRole:
			roleWithSig.Signatures = r.tufRepo.Snapshot.Signatures
		case data.CanonicalTimestampRole:
			roleWithSig.Signatures = r.tufRepo.Timestamp.Signatures
		default:
			if !data.IsDelegation(role.Name) {
				continue
			}
			if _, ok := r.tufRepo.Targets[role.Name]; ok {
				// We'll only find a signature if we've published any targets with this delegation
				roleWithSig.Signatures = r.tufRepo.Targets[role.Name].Signatures
			}
		}
		roleWithSigs = append(roleWithSigs, roleWithSig)
	}
	return roleWithSigs, nil
}

// Publish pushes the local changes in signed material to the remote notary-server
// Conceptually it performs an operation similar to a `git rebase`
func (r *repository) Publish() error {
	if err := r.publish(r.changelist); err != nil {
		return err
	}
	if err := r.changelist.Clear(""); err != nil {
		// This is not a critical problem when only a single host is pushing
		// but will cause weird behaviour if changelist cleanup is failing
		// and there are multiple hosts writing to the repo.
		logrus.Warn("Unable to clear changelist. You may want to manually delete the folder ", r.changelist.Location())
	}
	return nil
}

// publish pushes the changes in the given changelist to the remote notary-server
// Conceptually it performs an operation similar to a `git rebase`
func (r *repository) publish(cl changelist.Changelist) error {
	var initialPublish bool
	// update first before publishing
	if err := r.Update(true); err != nil {
		// If the remote is not aware of the repo, then this is being published
		// for the first time.  Try to initialize the repository before publishing.
		if _, ok := err.(ErrRepositoryNotExist); ok {
			err := r.bootstrapRepo()
			if _, ok := err.(store.ErrMetaNotFound); ok {
				logrus.Infof("No TUF data found locally or remotely - initializing repository %s for the first time", r.gun.String())
				err = r.Initialize(nil)
			}

			if err != nil {
				logrus.WithError(err).Debugf("Unable to load or initialize repository during first publish: %s", err.Error())
				return err
			}

			// Ensure we will push the initial root and targets file.  Either or
			// both of the root and targets may not be marked as Dirty, since
			// there may not be any changes that update them, so use a
			// different boolean.
			initialPublish = true
		} else {
			// We could not update, so we cannot publish.
			logrus.Error("Could not publish Repository since we could not update: ", err.Error())
			return err
		}
	}
	// apply the changelist to the repo
	if err := applyChangelist(r.tufRepo, r.invalid, cl); err != nil {
		logrus.Debug("Error applying changelist")
		return err
	}

	// these are the TUF files we will need to update, serialized as JSON before
	// we send anything to remote
	updatedFiles := make(map[data.RoleName][]byte)

	// Fetch old keys to support old clients
	legacyKeys, err := r.oldKeysForLegacyClientSupport(r.LegacyVersions, initialPublish)
	if err != nil {
		return err
	}

	// check if our root file is nearing expiry or dirty. Resign if it is.  If
	// root is not dirty but we are publishing for the first time, then just
	// publish the existing root we have.
	if err := signRootIfNecessary(updatedFiles, r.tufRepo, legacyKeys, initialPublish); err != nil {
		return err
	}

	if err := signTargets(updatedFiles, r.tufRepo, initialPublish); err != nil {
		return err
	}

	// if we initialized the repo while designating the server as the snapshot
	// signer, then there won't be a snapshots file.  However, we might now
	// have a local key (if there was a rotation), so initialize one.
	if r.tufRepo.Snapshot == nil {
		if err := r.tufRepo.InitSnapshot(); err != nil {
			return err
		}
	}

	if snapshotJSON, err := serializeCanonicalRole(
		r.tufRepo, data.CanonicalSnapshotRole, nil); err == nil {
		// Only update the snapshot if we've successfully signed it.
		updatedFiles[data.CanonicalSnapshotRole] = snapshotJSON
	} else if signErr, ok := err.(signed.ErrInsufficientSignatures); ok && signErr.FoundKeys == 0 {
		// If signing fails due to us not having the snapshot key, then
		// assume the server is going to sign, and do not include any snapshot
		// data.
		logrus.Debugf("Client does not have the key to sign snapshot. " +
			"Assuming that server should sign the snapshot.")
	} else {
		logrus.Debugf("Client was unable to sign the snapshot: %s", err.Error())
		return err
	}

	remote := r.getRemoteStore()

	return remote.SetMulti(data.MetadataRoleMapToStringMap(updatedFiles))
}

func signRootIfNecessary(updates map[data.RoleName][]byte, repo *tuf.Repo, extraSigningKeys data.KeyList, initialPublish bool) error {
	if len(extraSigningKeys) > 0 {
		repo.Root.Dirty = true
	}
	if nearExpiry(repo.Root.Signed.SignedCommon) || repo.Root.Dirty {
		rootJSON, err := serializeCanonicalRole(repo, data.CanonicalRootRole, extraSigningKeys)
		if err != nil {
			return err
		}
		updates[data.CanonicalRootRole] = rootJSON
	} else if initialPublish {
		rootJSON, err := repo.Root.MarshalJSON()
		if err != nil {
			return err
		}
		updates[data.CanonicalRootRole] = rootJSON
	}
	return nil
}

// Fetch back a `legacyVersions` number of roots files, collect the root public keys
// This includes old `root` roles as well as legacy versioned root roles, e.g. `1.root`
func (r *repository) oldKeysForLegacyClientSupport(legacyVersions int, initialPublish bool) (data.KeyList, error) {
	if initialPublish {
		return nil, nil
	}

	var oldestVersion int
	prevVersion := r.tufRepo.Root.Signed.Version

	if legacyVersions == SignWithAllOldVersions {
		oldestVersion = 1
	} else {
		oldestVersion = r.tufRepo.Root.Signed.Version - legacyVersions
	}

	if oldestVersion < 1 {
		oldestVersion = 1
	}

	if prevVersion <= 1 || oldestVersion == prevVersion {
		return nil, nil
	}
	oldKeys := make(map[string]data.PublicKey)

	c, err := r.bootstrapClient(true)
	// require a server connection to fetch old roots
	if err != nil {
		return nil, err
	}

	for v := prevVersion; v >= oldestVersion; v-- {
		logrus.Debugf("fetching old keys from version %d", v)
		// fetch old root version
		versionedRole := fmt.Sprintf("%d.%s", v, data.CanonicalRootRole.String())

		raw, err := c.remote.GetSized(versionedRole, -1)
		if err != nil {
			logrus.Debugf("error downloading %s: %s", versionedRole, err)
			continue
		}

		signedOldRoot := &data.Signed{}
		if err := json.Unmarshal(raw, signedOldRoot); err != nil {
			return nil, err
		}
		oldRootVersion, err := data.RootFromSigned(signedOldRoot)
		if err != nil {
			return nil, err
		}

		// extract legacy versioned root keys
		oldRootVersionKeys := getOldRootPublicKeys(oldRootVersion)
		for _, oldKey := range oldRootVersionKeys {
			oldKeys[oldKey.ID()] = oldKey
		}
	}
	oldKeyList := make(data.KeyList, 0, len(oldKeys))
	for _, key := range oldKeys {
		oldKeyList = append(oldKeyList, key)
	}
	return oldKeyList, nil
}

// get all the saved previous roles keys < the current root version
func getOldRootPublicKeys(root *data.SignedRoot) data.KeyList {
	rootRole, err := root.BuildBaseRole(data.CanonicalRootRole)
	if err != nil {
		return nil
	}
	return rootRole.ListKeys()
}

func signTargets(updates map[data.RoleName][]byte, repo *tuf.Repo, initialPublish bool) error {
	// iterate through all the targets files - if they are dirty, sign and update
	for roleName, roleObj := range repo.Targets {
		if roleObj.Dirty || (roleName == data.CanonicalTargetsRole && initialPublish) {
			targetsJSON, err := serializeCanonicalRole(repo, roleName, nil)
			if err != nil {
				return err
			}
			updates[roleName] = targetsJSON
		}
	}
	return nil
}

// bootstrapRepo loads the repository from the local file system (i.e.
// a not yet published repo or a possibly obsolete local copy) into
// r.tufRepo.  This attempts to load metadata for all roles.  Since server
// snapshots are supported, if the snapshot metadata fails to load, that's ok.
// This assumes that bootstrapRepo is only used by Publish() or RotateKey()
func (r *repository) bootstrapRepo() error {
	b := tuf.NewRepoBuilder(r.gun, r.GetCryptoService(), r.trustPinning)

	logrus.Debugf("Loading trusted collection.")

	for _, role := range data.BaseRoles {
		jsonBytes, err := r.cache.GetSized(role.String(), store.NoSizeLimit)
		if err != nil {
			if _, ok := err.(store.ErrMetaNotFound); ok &&
				// server snapshots are supported, and server timestamp management
				// is required, so if either of these fail to load that's ok - especially
				// if the repo is new
				role == data.CanonicalSnapshotRole || role == data.CanonicalTimestampRole {
				continue
			}
			return err
		}
		if err := b.Load(role, jsonBytes, 1, true); err != nil {
			return err
		}
	}

	tufRepo, _, err := b.Finish()
	if err == nil {
		r.tufRepo = tufRepo
	}
	return nil
}

// saveMetadata saves contents of r.tufRepo onto the local disk, creating
// signatures as necessary, possibly prompting for passphrases.
func (r *repository) saveMetadata(ignoreSnapshot bool) error {
	logrus.Debugf("Saving changes to Trusted Collection.")

	rootJSON, err := serializeCanonicalRole(r.tufRepo, data.CanonicalRootRole, nil)
	if err != nil {
		return err
	}
	err = r.cache.Set(data.CanonicalRootRole.String(), rootJSON)
	if err != nil {
		return err
	}

	targetsToSave := make(map[data.RoleName][]byte)
	for t := range r.tufRepo.Targets {
		signedTargets, err := r.tufRepo.SignTargets(t, data.DefaultExpires(data.CanonicalTargetsRole))
		if err != nil {
			return err
		}
		targetsJSON, err := json.Marshal(signedTargets)
		if err != nil {
			return err
		}
		targetsToSave[t] = targetsJSON
	}

	for role, blob := range targetsToSave {
		// If the parent directory does not exist, the cache.Set will create it
		r.cache.Set(role.String(), blob)
	}

	if ignoreSnapshot {
		return nil
	}

	snapshotJSON, err := serializeCanonicalRole(r.tufRepo, data.CanonicalSnapshotRole, nil)
	if err != nil {
		return err
	}

	return r.cache.Set(data.CanonicalSnapshotRole.String(), snapshotJSON)
}

// returns a properly constructed ErrRepositoryNotExist error based on this
// repo's information
func (r *repository) errRepositoryNotExist() error {
	host := r.baseURL
	parsed, err := url.Parse(r.baseURL)
	if err == nil {
		host = parsed.Host // try to exclude the scheme and any paths
	}
	return ErrRepositoryNotExist{remote: host, gun: r.gun}
}

// Update bootstraps a trust anchor (root.json) before updating all the
// metadata from the repo.
func (r *repository) Update(forWrite bool) error {
	c, err := r.bootstrapClient(forWrite)
	if err != nil {
		if _, ok := err.(store.ErrMetaNotFound); ok {
			return r.errRepositoryNotExist()
		}
		return err
	}
	repo, invalid, err := c.Update()
	if err != nil {
		// notFound.Resource may include a version or checksum so when the role is root,
		// it will be root, <version>.root or root.<checksum>.
		notFound, ok := err.(store.ErrMetaNotFound)
		isRoot, _ := regexp.MatchString(`\.?`+data.CanonicalRootRole.String()+`\.?`, notFound.Resource)
		if ok && isRoot {
			return r.errRepositoryNotExist()
		}
		return err
	}
	// we can be assured if we are at this stage that the repo we built is good
	// no need to test the following function call for an error as it will always be fine should the repo be good- it is!
	r.tufRepo = repo
	r.invalid = invalid
	warnRolesNearExpiry(repo)
	return nil
}

// bootstrapClient attempts to bootstrap a root.json to be used as the trust
// anchor for a repository. The checkInitialized argument indicates whether
// we should always attempt to contact the server to determine if the repository
// is initialized or not. If set to true, we will always attempt to download
// and return an error if the remote repository errors.
//
// Populates a tuf.RepoBuilder with this root metadata. If the root metadata
// downloaded is a newer version than what is on disk, then intermediate
// versions will be downloaded and verified in order to rotate trusted keys
// properly. Newer root metadata must always be signed with the previous
// threshold and keys.
//
// Fails if the remote server is reachable and does not know the repo
// (i.e. before the first r.Publish()), in which case the error is
// store.ErrMetaNotFound, or if the root metadata (from whichever source is used)
// is not trusted.
//
// Returns a TUFClient for the remote server, which may not be actually
// operational (if the URL is invalid but a root.json is cached).
func (r *repository) bootstrapClient(checkInitialized bool) (*tufClient, error) {
	minVersion := 1
	// the old root on disk should not be validated against any trust pinning configuration
	// because if we have an old root, it itself is the thing that pins trust
	oldBuilder := tuf.NewRepoBuilder(r.gun, r.GetCryptoService(), trustpinning.TrustPinConfig{})

	// by default, we want to use the trust pinning configuration on any new root that we download
	newBuilder := tuf.NewRepoBuilder(r.gun, r.GetCryptoService(), r.trustPinning)

	// Try to read root from cache first. We will trust this root until we detect a problem
	// during update which will cause us to download a new root and perform a rotation.
	// If we have an old root, and it's valid, then we overwrite the newBuilder to be one
	// preloaded with the old root or one which uses the old root for trust bootstrapping.
	if rootJSON, err := r.cache.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit); err == nil {
		// if we can't load the cached root, fail hard because that is how we pin trust
		if err := oldBuilder.Load(data.CanonicalRootRole, rootJSON, minVersion, true); err != nil {
			return nil, err
		}

		// again, the root on disk is the source of trust pinning, so use an empty trust
		// pinning configuration
		newBuilder = tuf.NewRepoBuilder(r.gun, r.GetCryptoService(), trustpinning.TrustPinConfig{})

		if err := newBuilder.Load(data.CanonicalRootRole, rootJSON, minVersion, false); err != nil {
			// Ok, the old root is expired - we want to download a new one.  But we want to use the
			// old root to verify the new root, so bootstrap a new builder with the old builder
			// but use the trustpinning to validate the new root
			minVersion = oldBuilder.GetLoadedVersion(data.CanonicalRootRole)
			newBuilder = oldBuilder.BootstrapNewBuilderWithNewTrustpin(r.trustPinning)
		}
	}

	remote := r.getRemoteStore()

	if !newBuilder.IsLoaded(data.CanonicalRootRole) || checkInitialized {
		// remoteErr was nil and we were not able to load a root from cache or
		// are specifically checking for initialization of the repo.

		// if remote store successfully set up, try and get root from remote
		// We don't have any local data to determine the size of root, so try the maximum (though it is restricted at 100MB)
		tmpJSON, err := remote.GetSized(data.CanonicalRootRole.String(), store.NoSizeLimit)
		if err != nil {
			// we didn't have a root in cache and were unable to load one from
			// the server. Nothing we can do but error.
			return nil, err
		}

		if !newBuilder.IsLoaded(data.CanonicalRootRole) {
			// we always want to use the downloaded root if we couldn't load from cache
			if err := newBuilder.Load(data.CanonicalRootRole, tmpJSON, minVersion, false); err != nil {
				return nil, err
			}

			err = r.cache.Set(data.CanonicalRootRole.String(), tmpJSON)
			if err != nil {
				// if we can't write cache we should still continue, just log error
				logrus.Errorf("could not save root to cache: %s", err.Error())
			}
		}
	}

	// We can only get here if remoteErr != nil (hence we don't download any new root),
	// and there was no root on disk
	if !newBuilder.IsLoaded(data.CanonicalRootRole) {
		return nil, ErrRepoNotInitialized{}
	}

	return newTufClient(oldBuilder, newBuilder, remote, r.cache), nil
}

// RotateKey removes all existing keys associated with the role. If no keys are
// specified in keyList, then this creates and adds one new key or delegates
// managing the key to the server. If key(s) are specified by keyList, then they are
// used for signing the role.
// These changes are staged in a changelist until publish is called.
func (r *repository) RotateKey(role data.RoleName, serverManagesKey bool, keyList []string) error {
	if err := checkRotationInput(role, serverManagesKey); err != nil {
		return err
	}

	pubKeyList, err := r.pubKeyListForRotation(role, serverManagesKey, keyList)
	if err != nil {
		return err
	}

	cl := changelist.NewMemChangelist()
	if err := r.rootFileKeyChange(cl, role, changelist.ActionCreate, pubKeyList); err != nil {
		return err
	}
	return r.publish(cl)
}

// Given a set of new keys to rotate to and a set of keys to drop, returns the list of current keys to use
func (r *repository) pubKeyListForRotation(role data.RoleName, serverManaged bool, newKeys []string) (pubKeyList data.KeyList, err error) {
	var pubKey data.PublicKey

	// If server manages the key being rotated, request a rotation and return the new key
	if serverManaged {
		remote := r.getRemoteStore()
		pubKey, err = rotateRemoteKey(role, remote)
		pubKeyList = make(data.KeyList, 0, 1)
		pubKeyList = append(pubKeyList, pubKey)
		if err != nil {
			return nil, fmt.Errorf("unable to rotate remote key: %s", err)
		}
		return pubKeyList, nil
	}

	// If no new keys are passed in, we generate one
	if len(newKeys) == 0 {
		pubKeyList = make(data.KeyList, 0, 1)
		pubKey, err = r.GetCryptoService().Create(role, r.gun, data.ECDSAKey)
		pubKeyList = append(pubKeyList, pubKey)
	}
	if err != nil {
		return nil, fmt.Errorf("unable to generate key: %s", err)
	}

	// If a list of keys to rotate to are provided, we add those
	if len(newKeys) > 0 {
		pubKeyList = make(data.KeyList, 0, len(newKeys))
		for _, keyID := range newKeys {
			pubKey = r.GetCryptoService().GetKey(keyID)
			if pubKey == nil {
				return nil, fmt.Errorf("unable to find key: %s", keyID)
			}
			pubKeyList = append(pubKeyList, pubKey)
		}
	}

	// Convert to certs (for root keys)
	if pubKeyList, err = r.pubKeysToCerts(role, pubKeyList); err != nil {
		return nil, err
	}

	return pubKeyList, nil
}

func (r *repository) pubKeysToCerts(role data.RoleName, pubKeyList data.KeyList) (data.KeyList, error) {
	// only generate certs for root keys
	if role != data.CanonicalRootRole {
		return pubKeyList, nil
	}

	for i, pubKey := range pubKeyList {
		privKey, loadedRole, err := r.GetCryptoService().GetPrivateKey(pubKey.ID())
		if err != nil {
			return nil, err
		}
		if loadedRole != role {
			return nil, fmt.Errorf("attempted to load root key but given %s key instead", loadedRole)
		}
		pubKey, err = rootCertKey(r.gun, privKey)
		if err != nil {
			return nil, err
		}
		pubKeyList[i] = pubKey
	}
	return pubKeyList, nil
}

func checkRotationInput(role data.RoleName, serverManaged bool) error {
	// We currently support remotely managing timestamp and snapshot keys
	canBeRemoteKey := role == data.CanonicalTimestampRole || role == data.CanonicalSnapshotRole
	// And locally managing root, targets, and snapshot keys
	canBeLocalKey := role == data.CanonicalSnapshotRole || role == data.CanonicalTargetsRole ||
		role == data.CanonicalRootRole

	switch {
	case !data.ValidRole(role) || data.IsDelegation(role):
		return fmt.Errorf("notary does not currently permit rotating the %s key", role)
	case serverManaged && !canBeRemoteKey:
		return ErrInvalidRemoteRole{Role: role}
	case !serverManaged && !canBeLocalKey:
		return ErrInvalidLocalRole{Role: role}
	}
	return nil
}

func (r *repository) rootFileKeyChange(cl changelist.Changelist, role data.RoleName, action string, keyList []data.PublicKey) error {
	meta := changelist.TUFRootData{
		RoleName: role,
		Keys:     keyList,
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	c := changelist.NewTUFChange(
		action,
		changelist.ScopeRoot,
		changelist.TypeBaseRole,
		role.String(),
		metaJSON,
	)
	return cl.Add(c)
}

// DeleteTrustData removes the trust data stored for this repo in the TUF cache on the client side
// Note that we will not delete any private key material from local storage
func DeleteTrustData(baseDir string, gun data.GUN, URL string, rt http.RoundTripper, deleteRemote bool) error {
	localRepo := filepath.Join(baseDir, tufDir, filepath.FromSlash(gun.String()))
	// Remove the tufRepoPath directory, which includes local TUF metadata files and changelist information
	if err := os.RemoveAll(localRepo); err != nil {
		return fmt.Errorf("error clearing TUF repo data: %v", err)
	}
	// Note that this will require admin permission for the gun in the roundtripper
	if deleteRemote {
		remote, err := getRemoteStore(URL, gun, rt)
		if err != nil {
			logrus.Error("unable to instantiate a remote store: %v", err)
			return err
		}
		if err := remote.RemoveAll(); err != nil {
			return err
		}
	}
	return nil
}

// SetLegacyVersions allows the number of legacy versions of the root
// to be inspected for old signing keys to be configured.
func (r *repository) SetLegacyVersions(n int) {
	r.LegacyVersions = n
}
