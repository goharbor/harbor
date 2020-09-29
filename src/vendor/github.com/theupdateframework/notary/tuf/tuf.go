// Package tuf defines the core TUF logic around manipulating a repo.
package tuf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/signed"
	"github.com/theupdateframework/notary/tuf/utils"
)

// ErrSigVerifyFail - signature verification failed
type ErrSigVerifyFail struct{}

func (e ErrSigVerifyFail) Error() string {
	return "Error: Signature verification failed"
}

// ErrMetaExpired - metadata file has expired
type ErrMetaExpired struct{}

func (e ErrMetaExpired) Error() string {
	return "Error: Metadata has expired"
}

// ErrLocalRootExpired - the local root file is out of date
type ErrLocalRootExpired struct{}

func (e ErrLocalRootExpired) Error() string {
	return "Error: Local Root Has Expired"
}

// ErrNotLoaded - attempted to access data that has not been loaded into
// the repo. This means specifically that the relevant JSON file has not
// been loaded.
type ErrNotLoaded struct {
	Role data.RoleName
}

func (err ErrNotLoaded) Error() string {
	return fmt.Sprintf("%s role has not been loaded", err.Role)
}

// StopWalk - used by visitor functions to signal WalkTargets to stop walking
type StopWalk struct{}

// Repo is an in memory representation of the TUF Repo.
// It operates at the data.Signed level, accepting and producing
// data.Signed objects. Users of a Repo are responsible for
// fetching raw JSON and using the Set* functions to populate
// the Repo instance.
type Repo struct {
	Root          *data.SignedRoot
	Targets       map[data.RoleName]*data.SignedTargets
	Snapshot      *data.SignedSnapshot
	Timestamp     *data.SignedTimestamp
	cryptoService signed.CryptoService

	// Because Repo is a mutable structure, these keep track of what the root
	// role was when a root is set on the repo (as opposed to what it might be
	// after things like AddBaseKeys and RemoveBaseKeys have been called on it).
	// If we know what the original was, we'll if and how to handle root
	// rotations.
	originalRootRole data.BaseRole
}

// NewRepo initializes a Repo instance with a CryptoService.
// If the Repo will only be used for reading, the CryptoService
// can be nil.
func NewRepo(cryptoService signed.CryptoService) *Repo {
	return &Repo{
		Targets:       make(map[data.RoleName]*data.SignedTargets),
		cryptoService: cryptoService,
	}
}

// AddBaseKeys is used to add keys to the role in root.json
func (tr *Repo) AddBaseKeys(role data.RoleName, keys ...data.PublicKey) error {
	if tr.Root == nil {
		return ErrNotLoaded{Role: data.CanonicalRootRole}
	}
	ids := []string{}
	for _, k := range keys {
		// Store only the public portion
		tr.Root.Signed.Keys[k.ID()] = k
		tr.Root.Signed.Roles[role].KeyIDs = append(tr.Root.Signed.Roles[role].KeyIDs, k.ID())
		ids = append(ids, k.ID())
	}
	tr.Root.Dirty = true

	// also, whichever role was switched out needs to be re-signed
	// root has already been marked dirty.
	switch role {
	case data.CanonicalSnapshotRole:
		if tr.Snapshot != nil {
			tr.Snapshot.Dirty = true
		}
	case data.CanonicalTargetsRole:
		if target, ok := tr.Targets[data.CanonicalTargetsRole]; ok {
			target.Dirty = true
		}
	case data.CanonicalTimestampRole:
		if tr.Timestamp != nil {
			tr.Timestamp.Dirty = true
		}
	}
	return nil
}

// ReplaceBaseKeys is used to replace all keys for the given role with the new keys
func (tr *Repo) ReplaceBaseKeys(role data.RoleName, keys ...data.PublicKey) error {
	r, err := tr.GetBaseRole(role)
	if err != nil {
		return err
	}
	err = tr.RemoveBaseKeys(role, r.ListKeyIDs()...)
	if err != nil {
		return err
	}
	return tr.AddBaseKeys(role, keys...)
}

// RemoveBaseKeys is used to remove keys from the roles in root.json
func (tr *Repo) RemoveBaseKeys(role data.RoleName, keyIDs ...string) error {
	if tr.Root == nil {
		return ErrNotLoaded{Role: data.CanonicalRootRole}
	}
	var keep []string
	toDelete := make(map[string]struct{})
	emptyStruct := struct{}{}
	// remove keys from specified role
	for _, k := range keyIDs {
		toDelete[k] = emptyStruct
	}

	oldKeyIDs := tr.Root.Signed.Roles[role].KeyIDs
	for _, rk := range oldKeyIDs {
		if _, ok := toDelete[rk]; !ok {
			keep = append(keep, rk)
		}
	}

	tr.Root.Signed.Roles[role].KeyIDs = keep

	// also, whichever role had keys removed needs to be re-signed
	// root has already been marked dirty.
	tr.markRoleDirty(role)

	// determine which keys are no longer in use by any roles
	for roleName, r := range tr.Root.Signed.Roles {
		if roleName == role {
			continue
		}
		for _, rk := range r.KeyIDs {
			if _, ok := toDelete[rk]; ok {
				delete(toDelete, rk)
			}
		}
	}

	// Remove keys no longer in use by any roles, except for root keys.
	// Root private keys must be kept in tr.cryptoService to be able to sign
	// for rotation, and root certificates must be kept in tr.Root.SignedKeys
	// because we are not necessarily storing them elsewhere (tuf.Repo does not
	// depend on certs.Manager, that is an upper layer), and without storing
	// the certificates in their x509 form we are not able to do the
	// util.CanonicalKeyID conversion.
	if role != data.CanonicalRootRole {
		for k := range toDelete {
			delete(tr.Root.Signed.Keys, k)
			tr.cryptoService.RemoveKey(k)
		}
	}

	tr.Root.Dirty = true
	return nil
}

func (tr *Repo) markRoleDirty(role data.RoleName) {
	switch role {
	case data.CanonicalSnapshotRole:
		if tr.Snapshot != nil {
			tr.Snapshot.Dirty = true
		}
	case data.CanonicalTargetsRole:
		if target, ok := tr.Targets[data.CanonicalTargetsRole]; ok {
			target.Dirty = true
		}
	case data.CanonicalTimestampRole:
		if tr.Timestamp != nil {
			tr.Timestamp.Dirty = true
		}
	}
}

// GetBaseRole gets a base role from this repo's metadata
func (tr *Repo) GetBaseRole(name data.RoleName) (data.BaseRole, error) {
	if !data.ValidRole(name) {
		return data.BaseRole{}, data.ErrInvalidRole{Role: name, Reason: "invalid base role name"}
	}
	if tr.Root == nil {
		return data.BaseRole{}, ErrNotLoaded{data.CanonicalRootRole}
	}
	// Find the role data public keys for the base role from TUF metadata
	baseRole, err := tr.Root.BuildBaseRole(name)
	if err != nil {
		return data.BaseRole{}, err
	}

	return baseRole, nil
}

// GetDelegationRole gets a delegation role from this repo's metadata, walking from the targets role down to the delegation itself
func (tr *Repo) GetDelegationRole(name data.RoleName) (data.DelegationRole, error) {
	if !data.IsDelegation(name) {
		return data.DelegationRole{}, data.ErrInvalidRole{Role: name, Reason: "invalid delegation name"}
	}
	if tr.Root == nil {
		return data.DelegationRole{}, ErrNotLoaded{data.CanonicalRootRole}
	}
	_, ok := tr.Root.Signed.Roles[data.CanonicalTargetsRole]
	if !ok {
		return data.DelegationRole{}, ErrNotLoaded{data.CanonicalTargetsRole}
	}
	// Traverse target metadata, down to delegation itself
	// Get all public keys for the base role from TUF metadata
	_, ok = tr.Targets[data.CanonicalTargetsRole]
	if !ok {
		return data.DelegationRole{}, ErrNotLoaded{data.CanonicalTargetsRole}
	}

	// Start with top level roles in targets. Walk the chain of ancestors
	// until finding the desired role, or we run out of targets files to search.
	var foundRole *data.DelegationRole
	buildDelegationRoleVisitor := func(tgt *data.SignedTargets, validRole data.DelegationRole) interface{} {
		// Try to find the delegation and build a DelegationRole structure
		for _, role := range tgt.Signed.Delegations.Roles {
			if role.Name == name {
				delgRole, err := tgt.BuildDelegationRole(name)
				if err != nil {
					return err
				}
				// Check all public key certificates in the role for expiry
				// Currently we do not reject expired delegation keys but warn if they might expire soon or have already
				for _, pubKey := range delgRole.Keys {
					certFromKey, err := utils.LoadCertFromPEM(pubKey.Public())
					if err != nil {
						continue
					}
					//Don't check the delegation certificate expiry once added, use the TUF role expiry instead
					if err := utils.ValidateCertificate(certFromKey, false); err != nil {
						return err
					}
				}
				foundRole = &delgRole
				return StopWalk{}
			}
		}
		return nil
	}

	// Walk to the parent of this delegation, since that is where its role metadata exists
	err := tr.WalkTargets("", name.Parent(), buildDelegationRoleVisitor)
	if err != nil {
		return data.DelegationRole{}, err
	}

	// We never found the delegation. In the context of this repo it is considered
	// invalid. N.B. it may be that it existed at one point but an ancestor has since
	// been modified/removed.
	if foundRole == nil {
		return data.DelegationRole{}, data.ErrInvalidRole{Role: name, Reason: "delegation does not exist"}
	}

	return *foundRole, nil
}

// GetAllLoadedRoles returns a list of all role entries loaded in this TUF repo, could be empty
func (tr *Repo) GetAllLoadedRoles() []*data.Role {
	var res []*data.Role
	if tr.Root == nil {
		// if root isn't loaded, we should consider we have no loaded roles because we can't
		// trust any other state that might be present
		return res
	}
	for name, rr := range tr.Root.Signed.Roles {
		res = append(res, &data.Role{
			RootRole: *rr,
			Name:     name,
		})
	}
	for _, delegate := range tr.Targets {
		for _, r := range delegate.Signed.Delegations.Roles {
			res = append(res, r)
		}
	}
	return res
}

// Walk to parent, and either create or update this delegation.  We can only create a new delegation if we're given keys
// Ensure all updates are valid, by checking against parent ancestor paths and ensuring the keys meet the role threshold.
func delegationUpdateVisitor(roleName data.RoleName, addKeys data.KeyList, removeKeys, addPaths, removePaths []string, clearAllPaths bool, newThreshold int) walkVisitorFunc {
	return func(tgt *data.SignedTargets, validRole data.DelegationRole) interface{} {
		var err error
		// Validate the changes underneath this restricted validRole for adding paths, reject invalid path additions
		if len(addPaths) != len(data.RestrictDelegationPathPrefixes(validRole.Paths, addPaths)) {
			return data.ErrInvalidRole{Role: roleName, Reason: "invalid paths to add to role"}
		}
		// Try to find the delegation and amend it using our changelist
		var delgRole *data.Role
		for _, role := range tgt.Signed.Delegations.Roles {
			if role.Name == roleName {
				// Make a copy and operate on this role until we validate the changes
				keyIDCopy := make([]string, len(role.KeyIDs))
				copy(keyIDCopy, role.KeyIDs)
				pathsCopy := make([]string, len(role.Paths))
				copy(pathsCopy, role.Paths)
				delgRole = &data.Role{
					RootRole: data.RootRole{
						KeyIDs:    keyIDCopy,
						Threshold: role.Threshold,
					},
					Name:  role.Name,
					Paths: pathsCopy,
				}
				delgRole.RemovePaths(removePaths)
				if clearAllPaths {
					delgRole.Paths = []string{}
				}
				delgRole.AddPaths(addPaths)
				delgRole.RemoveKeys(removeKeys)
				break
			}
		}
		// We didn't find the role earlier, so create it.
		if addKeys == nil {
			addKeys = data.KeyList{} // initialize to empty list if necessary so calling .IDs() below won't panic
		}
		if delgRole == nil {
			delgRole, err = data.NewRole(roleName, newThreshold, addKeys.IDs(), addPaths)
			if err != nil {
				return err
			}

		}
		// Add the key IDs to the role and the keys themselves to the parent
		for _, k := range addKeys {
			if !utils.StrSliceContains(delgRole.KeyIDs, k.ID()) {
				delgRole.KeyIDs = append(delgRole.KeyIDs, k.ID())
			}
		}
		// Make sure we have a valid role still
		if len(delgRole.KeyIDs) < delgRole.Threshold {
			logrus.Warnf("role %s has fewer keys than its threshold of %d; it will not be usable until keys are added to it", delgRole.Name, delgRole.Threshold)
		}
		// NOTE: this closure CANNOT error after this point, as we've committed to editing the SignedTargets metadata in the repo object.
		// Any errors related to updating this delegation must occur before this point.
		// If all of our changes were valid, we should edit the actual SignedTargets to match our copy
		for _, k := range addKeys {
			tgt.Signed.Delegations.Keys[k.ID()] = k
		}
		foundAt := utils.FindRoleIndex(tgt.Signed.Delegations.Roles, delgRole.Name)
		if foundAt < 0 {
			tgt.Signed.Delegations.Roles = append(tgt.Signed.Delegations.Roles, delgRole)
		} else {
			tgt.Signed.Delegations.Roles[foundAt] = delgRole
		}
		tgt.Dirty = true
		utils.RemoveUnusedKeys(tgt)
		return StopWalk{}
	}
}

// UpdateDelegationKeys updates the appropriate delegations, either adding
// a new delegation or updating an existing one. If keys are
// provided, the IDs will be added to the role (if they do not exist
// there already), and the keys will be added to the targets file.
func (tr *Repo) UpdateDelegationKeys(roleName data.RoleName, addKeys data.KeyList, removeKeys []string, newThreshold int) error {
	if !data.IsDelegation(roleName) {
		return data.ErrInvalidRole{Role: roleName, Reason: "not a valid delegated role"}
	}
	parent := roleName.Parent()

	if err := tr.VerifyCanSign(parent); err != nil {
		return err
	}

	// check the parent role's metadata
	_, ok := tr.Targets[parent]
	if !ok { // the parent targetfile may not exist yet - if not, then create it
		var err error
		_, err = tr.InitTargets(parent)
		if err != nil {
			return err
		}
	}

	// Walk to the parent of this delegation, since that is where its role metadata exists
	// We do not have to verify that the walker reached its desired role in this scenario
	// since we've already done another walk to the parent role in VerifyCanSign, and potentially made a targets file
	return tr.WalkTargets("", roleName.Parent(), delegationUpdateVisitor(roleName, addKeys, removeKeys, []string{}, []string{}, false, newThreshold))
}

// PurgeDelegationKeys removes the provided canonical key IDs from all delegations
// present in the subtree rooted at role. The role argument must be provided in a wildcard
// format, i.e. targets/* would remove the key from all delegations in the repo
func (tr *Repo) PurgeDelegationKeys(role data.RoleName, removeKeys []string) error {
	if !data.IsWildDelegation(role) {
		return data.ErrInvalidRole{
			Role:   role,
			Reason: "only wildcard roles can be used in a purge",
		}
	}

	removeIDs := make(map[string]struct{})
	for _, id := range removeKeys {
		removeIDs[id] = struct{}{}
	}

	start := role.Parent()
	tufIDToCanon := make(map[string]string)

	purgeKeys := func(tgt *data.SignedTargets, validRole data.DelegationRole) interface{} {
		var (
			deleteCandidates []string
			err              error
		)
		for id, key := range tgt.Signed.Delegations.Keys {
			var (
				canonID string
				ok      bool
			)
			if canonID, ok = tufIDToCanon[id]; !ok {
				canonID, err = utils.CanonicalKeyID(key)
				if err != nil {
					return err
				}
				tufIDToCanon[id] = canonID
			}
			if _, ok := removeIDs[canonID]; ok {
				deleteCandidates = append(deleteCandidates, id)
			}
		}
		if len(deleteCandidates) == 0 {
			// none of the interesting keys were present. We're done with this role
			return nil
		}
		// now we know there are changes, check if we'll be able to sign them in
		if err := tr.VerifyCanSign(validRole.Name); err != nil {
			logrus.Warnf(
				"role %s contains keys being purged but you do not have the necessary keys present to sign it; keys will not be purged from %s or its immediate children",
				validRole.Name,
				validRole.Name,
			)
			return nil
		}
		// we know we can sign in the changes, delete the keys
		for _, id := range deleteCandidates {
			delete(tgt.Signed.Delegations.Keys, id)
		}
		// delete candidate keys from all roles.
		for _, role := range tgt.Signed.Delegations.Roles {
			role.RemoveKeys(deleteCandidates)
			if len(role.KeyIDs) < role.Threshold {
				logrus.Warnf("role %s has fewer keys than its threshold of %d; it will not be usable until keys are added to it", role.Name, role.Threshold)
			}
		}
		tgt.Dirty = true
		return nil
	}
	return tr.WalkTargets("", start, purgeKeys)
}

// UpdateDelegationPaths updates the appropriate delegation's paths.
// It is not allowed to create a new delegation.
func (tr *Repo) UpdateDelegationPaths(roleName data.RoleName, addPaths, removePaths []string, clearPaths bool) error {
	if !data.IsDelegation(roleName) {
		return data.ErrInvalidRole{Role: roleName, Reason: "not a valid delegated role"}
	}
	parent := roleName.Parent()

	if err := tr.VerifyCanSign(parent); err != nil {
		return err
	}

	// check the parent role's metadata
	_, ok := tr.Targets[parent]
	if !ok { // the parent targetfile may not exist yet
		// if not, this is an error because a delegation must exist to edit only paths
		return data.ErrInvalidRole{Role: roleName, Reason: "no valid delegated role exists"}
	}

	// Walk to the parent of this delegation, since that is where its role metadata exists
	// We do not have to verify that the walker reached its desired role in this scenario
	// since we've already done another walk to the parent role in VerifyCanSign
	err := tr.WalkTargets("", parent, delegationUpdateVisitor(roleName, data.KeyList{}, []string{}, addPaths, removePaths, clearPaths, notary.MinThreshold))
	if err != nil {
		return err
	}
	return nil
}

// DeleteDelegation removes a delegated targets role from its parent
// targets object. It also deletes the delegation from the snapshot.
// DeleteDelegation will only make use of the role Name field.
func (tr *Repo) DeleteDelegation(roleName data.RoleName) error {
	if !data.IsDelegation(roleName) {
		return data.ErrInvalidRole{Role: roleName, Reason: "not a valid delegated role"}
	}

	parent := roleName.Parent()
	if err := tr.VerifyCanSign(parent); err != nil {
		return err
	}

	// delete delegated data from Targets map and Snapshot - if they don't
	// exist, these are no-op
	delete(tr.Targets, roleName)
	tr.Snapshot.DeleteMeta(roleName)

	p, ok := tr.Targets[parent]
	if !ok {
		// if there is no parent metadata (the role exists though), then this
		// is as good as done.
		return nil
	}

	foundAt := utils.FindRoleIndex(p.Signed.Delegations.Roles, roleName)

	if foundAt >= 0 {
		var roles []*data.Role
		// slice out deleted role
		roles = append(roles, p.Signed.Delegations.Roles[:foundAt]...)
		if foundAt+1 < len(p.Signed.Delegations.Roles) {
			roles = append(roles, p.Signed.Delegations.Roles[foundAt+1:]...)
		}
		p.Signed.Delegations.Roles = roles

		utils.RemoveUnusedKeys(p)

		p.Dirty = true
	} // if the role wasn't found, it's a good as deleted

	return nil
}

// InitRoot initializes an empty root file with the 4 core roles passed to the
// method, and the consistent flag.
func (tr *Repo) InitRoot(root, timestamp, snapshot, targets data.BaseRole, consistent bool) error {
	rootRoles := make(map[data.RoleName]*data.RootRole)
	rootKeys := make(map[string]data.PublicKey)

	for _, r := range []data.BaseRole{root, timestamp, snapshot, targets} {
		rootRoles[r.Name] = &data.RootRole{
			Threshold: r.Threshold,
			KeyIDs:    r.ListKeyIDs(),
		}
		for kid, k := range r.Keys {
			rootKeys[kid] = k
		}
	}
	r, err := data.NewRoot(rootKeys, rootRoles, consistent)
	if err != nil {
		return err
	}
	tr.Root = r
	tr.originalRootRole = root
	return nil
}

// InitTargets initializes an empty targets, and returns the new empty target
func (tr *Repo) InitTargets(role data.RoleName) (*data.SignedTargets, error) {
	if !data.IsDelegation(role) && role != data.CanonicalTargetsRole {
		return nil, data.ErrInvalidRole{
			Role:   role,
			Reason: fmt.Sprintf("role is not a valid targets role name: %s", role.String()),
		}
	}
	targets := data.NewTargets()
	tr.Targets[role] = targets
	return targets, nil
}

// InitSnapshot initializes a snapshot based on the current root and targets
func (tr *Repo) InitSnapshot() error {
	if tr.Root == nil {
		return ErrNotLoaded{Role: data.CanonicalRootRole}
	}
	root, err := tr.Root.ToSigned()
	if err != nil {
		return err
	}

	if _, ok := tr.Targets[data.CanonicalTargetsRole]; !ok {
		return ErrNotLoaded{Role: data.CanonicalTargetsRole}
	}
	targets, err := tr.Targets[data.CanonicalTargetsRole].ToSigned()
	if err != nil {
		return err
	}
	snapshot, err := data.NewSnapshot(root, targets)
	if err != nil {
		return err
	}
	tr.Snapshot = snapshot
	return nil
}

// InitTimestamp initializes a timestamp based on the current snapshot
func (tr *Repo) InitTimestamp() error {
	snap, err := tr.Snapshot.ToSigned()
	if err != nil {
		return err
	}
	timestamp, err := data.NewTimestamp(snap)
	if err != nil {
		return err
	}

	tr.Timestamp = timestamp
	return nil
}

// TargetMeta returns the FileMeta entry for the given path in the
// targets file associated with the given role. This may be nil if
// the target isn't found in the targets file.
func (tr Repo) TargetMeta(role data.RoleName, path string) *data.FileMeta {
	if t, ok := tr.Targets[role]; ok {
		if m, ok := t.Signed.Targets[path]; ok {
			return &m
		}
	}
	return nil
}

// TargetDelegations returns a slice of Roles that are valid publishers
// for the target path provided.
func (tr Repo) TargetDelegations(role data.RoleName, path string) []*data.Role {
	var roles []*data.Role
	if t, ok := tr.Targets[role]; ok {
		for _, r := range t.Signed.Delegations.Roles {
			if r.CheckPaths(path) {
				roles = append(roles, r)
			}
		}
	}
	return roles
}

// VerifyCanSign returns nil if the role exists and we have at least one
// signing key for the role, false otherwise.  This does not check that we have
// enough signing keys to meet the threshold, since we want to support the use
// case of multiple signers for a role.  It returns an error if the role doesn't
// exist or if there are no signing keys.
func (tr *Repo) VerifyCanSign(roleName data.RoleName) error {
	var (
		role            data.BaseRole
		err             error
		canonicalKeyIDs []string
	)
	// we only need the BaseRole part of a delegation because we're just
	// checking KeyIDs
	if data.IsDelegation(roleName) {
		r, err := tr.GetDelegationRole(roleName)
		if err != nil {
			return err
		}
		role = r.BaseRole
	} else {
		role, err = tr.GetBaseRole(roleName)
	}
	if err != nil {
		return data.ErrInvalidRole{Role: roleName, Reason: "does not exist"}
	}

	for keyID, k := range role.Keys {
		check := []string{keyID}
		if canonicalID, err := utils.CanonicalKeyID(k); err == nil {
			check = append(check, canonicalID)
			canonicalKeyIDs = append(canonicalKeyIDs, canonicalID)
		}
		for _, id := range check {
			p, _, err := tr.cryptoService.GetPrivateKey(id)
			if err == nil && p != nil {
				return nil
			}
		}
	}
	return signed.ErrNoKeys{KeyIDs: canonicalKeyIDs}
}

// used for walking the targets/delegations tree, potentially modifying the underlying SignedTargets for the repo
type walkVisitorFunc func(*data.SignedTargets, data.DelegationRole) interface{}

// WalkTargets will apply the specified visitor function to iteratively walk the targets/delegation metadata tree,
// until receiving a StopWalk.  The walk starts from the base "targets" role, and searches for the correct targetPath and/or rolePath
// to call the visitor function on.  Any roles passed into skipRoles will be excluded from the walk, as well as roles in those subtrees
func (tr *Repo) WalkTargets(targetPath string, rolePath data.RoleName, visitTargets walkVisitorFunc, skipRoles ...data.RoleName) error {
	// Start with the base targets role, which implicitly has the "" targets path
	targetsRole, err := tr.GetBaseRole(data.CanonicalTargetsRole)
	if err != nil {
		return err
	}
	// Make the targets role have the empty path, when we treat it as a delegation role
	roles := []data.DelegationRole{
		{
			BaseRole: targetsRole,
			Paths:    []string{""},
		},
	}

	for len(roles) > 0 {
		role := roles[0]
		roles = roles[1:]

		// Check the role metadata
		signedTgt, ok := tr.Targets[role.Name]
		if !ok {
			// The role meta doesn't exist in the repo so continue onward
			continue
		}

		// We're at a prefix of the desired role subtree, so add its delegation role children and continue walking
		if strings.HasPrefix(rolePath.String(), role.Name.String()+"/") {
			roles = append(roles, signedTgt.GetValidDelegations(role)...)
			continue
		}

		// Determine whether to visit this role or not:
		// If the paths validate against the specified targetPath and the role is empty or is a path in the subtree.
		// Also check if we are choosing to skip visiting this role on this walk (see ListTargets and GetTargetByName priority)
		if isValidPath(targetPath, role) && isAncestorRole(role.Name, rolePath) && !utils.RoleNameSliceContains(skipRoles, role.Name) {
			// If we had matching path or role name, visit this target and determine whether or not to keep walking
			res := visitTargets(signedTgt, role)
			switch typedRes := res.(type) {
			case StopWalk:
				// If the visitor function signalled a stop, return nil to finish the walk
				return nil
			case nil:
				// If the visitor function signalled to continue, add this role's delegation to the walk
				roles = append(roles, signedTgt.GetValidDelegations(role)...)
			case error:
				// Propagate any errors from the visitor
				return typedRes
			default:
				// Return out with an error if we got a different result
				return fmt.Errorf("unexpected return while walking: %v", res)
			}

		}
	}
	return nil
}

// helper function that returns whether the candidateChild role name is an ancestor or equal to the candidateAncestor role name
// Will return true if given an empty candidateAncestor role name
// The HasPrefix check is for determining whether the role name for candidateChild is a child (direct or further down the chain)
// of candidateAncestor, for ex: candidateAncestor targets/a and candidateChild targets/a/b/c
func isAncestorRole(candidateChild data.RoleName, candidateAncestor data.RoleName) bool {
	return candidateAncestor.String() == "" || candidateAncestor == candidateChild || strings.HasPrefix(candidateChild.String(), candidateAncestor.String()+"/")
}

// helper function that returns whether the delegation Role is valid against the given path
// Will return true if given an empty candidatePath
func isValidPath(candidatePath string, delgRole data.DelegationRole) bool {
	return candidatePath == "" || delgRole.CheckPaths(candidatePath)
}

// AddTargets will attempt to add the given targets specifically to
// the directed role. If the metadata for the role doesn't exist yet,
// AddTargets will create one.
func (tr *Repo) AddTargets(role data.RoleName, targets data.Files) (data.Files, error) {
	cantSignErr := tr.VerifyCanSign(role)
	if _, ok := cantSignErr.(data.ErrInvalidRole); ok {
		return nil, cantSignErr
	}
	var needSign bool

	// check existence of the role's metadata
	_, ok := tr.Targets[role]
	if !ok { // the targetfile may not exist yet - if not, then create it
		var err error
		_, err = tr.InitTargets(role)
		if err != nil {
			return nil, err
		}
	}

	addedTargets := make(data.Files)
	addTargetVisitor := func(targetPath string, targetMeta data.FileMeta) func(*data.SignedTargets, data.DelegationRole) interface{} {
		return func(tgt *data.SignedTargets, validRole data.DelegationRole) interface{} {
			// We've already validated the role's target path in our walk, so just modify the metadata
			if targetMeta.Equals(tgt.Signed.Targets[targetPath]) {
				// Also add to our new addedTargets map because this target was "added" successfully
				addedTargets[targetPath] = targetMeta
				return StopWalk{}
			}
			needSign = true
			if cantSignErr == nil {
				tgt.Signed.Targets[targetPath] = targetMeta
				tgt.Dirty = true
				// Also add to our new addedTargets map to keep track of every target we've added successfully
				addedTargets[targetPath] = targetMeta
			}
			return StopWalk{}
		}
	}

	// Walk the role tree while validating the target paths, and add all of our targets
	for path, target := range targets {
		tr.WalkTargets(path, role, addTargetVisitor(path, target))
		if needSign && cantSignErr != nil {
			return nil, cantSignErr
		}
	}
	if len(addedTargets) != len(targets) {
		return nil, fmt.Errorf("Could not add all targets")
	}
	return nil, nil
}

// RemoveTargets removes the given target (paths) from the given target role (delegation)
func (tr *Repo) RemoveTargets(role data.RoleName, targets ...string) error {
	cantSignErr := tr.VerifyCanSign(role)
	if _, ok := cantSignErr.(data.ErrInvalidRole); ok {
		return cantSignErr
	}
	var needSign bool
	removeTargetVisitor := func(targetPath string) func(*data.SignedTargets, data.DelegationRole) interface{} {
		return func(tgt *data.SignedTargets, validRole data.DelegationRole) interface{} {
			// We've already validated the role path in our walk, so just modify the metadata
			// We don't check against the target path against the valid role paths because it's
			// possible we got into an invalid state and are trying to fix it
			if _, needSign = tgt.Signed.Targets[targetPath]; needSign && cantSignErr == nil {
				delete(tgt.Signed.Targets, targetPath)
				tgt.Dirty = true
			}
			return StopWalk{}
		}
	}

	// if the role exists but metadata does not yet, then our work is done
	_, ok := tr.Targets[role]
	if ok {
		for _, path := range targets {
			tr.WalkTargets("", role, removeTargetVisitor(path))
			if needSign && cantSignErr != nil {
				return cantSignErr
			}
		}
	}

	return nil
}

// UpdateSnapshot updates the FileMeta for the given role based on the Signed object
func (tr *Repo) UpdateSnapshot(role data.RoleName, s *data.Signed) error {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return err
	}
	meta, err := data.NewFileMeta(bytes.NewReader(jsonData), data.NotaryDefaultHashes...)
	if err != nil {
		return err
	}
	tr.Snapshot.Signed.Meta[role.String()] = meta
	tr.Snapshot.Dirty = true
	return nil
}

// UpdateTimestamp updates the snapshot meta in the timestamp based on the Signed object
func (tr *Repo) UpdateTimestamp(s *data.Signed) error {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return err
	}
	meta, err := data.NewFileMeta(bytes.NewReader(jsonData), data.NotaryDefaultHashes...)
	if err != nil {
		return err
	}
	tr.Timestamp.Signed.Meta[data.CanonicalSnapshotRole.String()] = meta
	tr.Timestamp.Dirty = true
	return nil
}

// SignRoot signs the root, using all keys from the "root" role (i.e. currently trusted)
// as well as available keys used to sign the previous version, if the public part is
// carried in tr.Root.Keys and the private key is available (i.e. probably previously
// trusted keys, to allow rollover).  If there are any errors, attempt to put root
// back to the way it was (so version won't be incremented, for instance).
// Extra signing keys can be added to support older clients
func (tr *Repo) SignRoot(expires time.Time, extraSigningKeys data.KeyList) (*data.Signed, error) {
	logrus.Debug("signing root...")

	// duplicate root and attempt to modify it rather than the existing root
	rootBytes, err := tr.Root.MarshalJSON()
	if err != nil {
		return nil, err
	}
	tempRoot := data.SignedRoot{}
	if err := json.Unmarshal(rootBytes, &tempRoot); err != nil {
		return nil, err
	}

	currRoot, err := tr.GetBaseRole(data.CanonicalRootRole)
	if err != nil {
		return nil, err
	}

	var rolesToSignWith []data.BaseRole

	// If the root role (root keys or root threshold) has changed, sign with the
	// previous root role keys
	if !tr.originalRootRole.Equals(currRoot) {
		rolesToSignWith = append(rolesToSignWith, tr.originalRootRole)
	}

	tempRoot.Signed.Expires = expires
	tempRoot.Signed.Version++
	rolesToSignWith = append(rolesToSignWith, currRoot)

	signed, err := tempRoot.ToSigned()
	if err != nil {
		return nil, err
	}

	signed, err = tr.sign(signed, rolesToSignWith, extraSigningKeys)
	if err != nil {
		return nil, err
	}

	tr.Root = &tempRoot
	tr.Root.Signatures = signed.Signatures
	tr.originalRootRole = currRoot
	return signed, nil
}

func oldRootVersionName(version int) string {
	return fmt.Sprintf("%s.%v", data.CanonicalRootRole, version)
}

// SignTargets signs the targets file for the given top level or delegated targets role
func (tr *Repo) SignTargets(role data.RoleName, expires time.Time) (*data.Signed, error) {
	logrus.Debugf("sign targets called for role %s", role)
	if _, ok := tr.Targets[role]; !ok {
		return nil, data.ErrInvalidRole{
			Role:   role,
			Reason: "SignTargets called with non-existent targets role",
		}
	}
	tr.Targets[role].Signed.Expires = expires
	tr.Targets[role].Signed.Version++
	signed, err := tr.Targets[role].ToSigned()
	if err != nil {
		logrus.Debug("errored getting targets data.Signed object")
		return nil, err
	}

	var targets data.BaseRole
	if role == data.CanonicalTargetsRole {
		targets, err = tr.GetBaseRole(role)
	} else {
		tr, err := tr.GetDelegationRole(role)
		if err != nil {
			return nil, err
		}
		targets = tr.BaseRole
	}
	if err != nil {
		return nil, err
	}

	signed, err = tr.sign(signed, []data.BaseRole{targets}, nil)
	if err != nil {
		logrus.Debug("errored signing ", role)
		return nil, err
	}
	tr.Targets[role].Signatures = signed.Signatures
	return signed, nil
}

// SignSnapshot updates the snapshot based on the current targets and root then signs it
func (tr *Repo) SignSnapshot(expires time.Time) (*data.Signed, error) {
	logrus.Debug("signing snapshot...")
	signedRoot, err := tr.Root.ToSigned()
	if err != nil {
		return nil, err
	}
	err = tr.UpdateSnapshot(data.CanonicalRootRole, signedRoot)
	if err != nil {
		return nil, err
	}
	tr.Root.Dirty = false // root dirty until changes captures in snapshot
	for role, targets := range tr.Targets {
		signedTargets, err := targets.ToSigned()
		if err != nil {
			return nil, err
		}
		err = tr.UpdateSnapshot(role, signedTargets)
		if err != nil {
			return nil, err
		}
		targets.Dirty = false
	}
	tr.Snapshot.Signed.Expires = expires
	tr.Snapshot.Signed.Version++
	signed, err := tr.Snapshot.ToSigned()
	if err != nil {
		return nil, err
	}
	snapshot, err := tr.GetBaseRole(data.CanonicalSnapshotRole)
	if err != nil {
		return nil, err
	}
	signed, err = tr.sign(signed, []data.BaseRole{snapshot}, nil)
	if err != nil {
		return nil, err
	}
	tr.Snapshot.Signatures = signed.Signatures
	return signed, nil
}

// SignTimestamp updates the timestamp based on the current snapshot then signs it
func (tr *Repo) SignTimestamp(expires time.Time) (*data.Signed, error) {
	logrus.Debug("SignTimestamp")
	signedSnapshot, err := tr.Snapshot.ToSigned()
	if err != nil {
		return nil, err
	}
	err = tr.UpdateTimestamp(signedSnapshot)
	if err != nil {
		return nil, err
	}
	tr.Timestamp.Signed.Expires = expires
	tr.Timestamp.Signed.Version++
	signed, err := tr.Timestamp.ToSigned()
	if err != nil {
		return nil, err
	}
	timestamp, err := tr.GetBaseRole(data.CanonicalTimestampRole)
	if err != nil {
		return nil, err
	}
	signed, err = tr.sign(signed, []data.BaseRole{timestamp}, nil)
	if err != nil {
		return nil, err
	}
	tr.Timestamp.Signatures = signed.Signatures
	tr.Snapshot.Dirty = false // snapshot is dirty until changes have been captured in timestamp
	return signed, nil
}

func (tr Repo) sign(signedData *data.Signed, roles []data.BaseRole, optionalKeys []data.PublicKey) (*data.Signed, error) {
	validKeys := optionalKeys
	for _, r := range roles {
		roleKeys := r.ListKeys()
		validKeys = append(roleKeys, validKeys...)
		if err := signed.Sign(tr.cryptoService, signedData, roleKeys, r.Threshold, validKeys); err != nil {
			return nil, err
		}
	}
	// Attempt to sign with the optional keys, but ignore any errors, because these keys are optional
	signed.Sign(tr.cryptoService, signedData, optionalKeys, 0, validKeys)

	return signedData, nil
}
