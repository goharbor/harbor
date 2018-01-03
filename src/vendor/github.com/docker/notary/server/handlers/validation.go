package handlers

import (
	"fmt"
	"sort"

	"github.com/Sirupsen/logrus"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/trustpinning"
	"github.com/docker/notary/tuf"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/utils"
	"github.com/docker/notary/tuf/validation"
)

// validateUpload checks that the updates being pushed
// are semantically correct and the signatures are correct
// A list of possibly modified updates are returned if all
// validation was successful. This allows the snapshot to be
// created and added if snapshotting has been delegated to the
// server
func validateUpdate(cs signed.CryptoService, gun data.GUN, updates []storage.MetaUpdate, store storage.MetaStore) ([]storage.MetaUpdate, error) {

	// some delegated targets role may be invalid based on other updates
	// that have been made by other clients. We'll rebuild the slice of
	// updates with only the things we should actually update
	updatesToApply := make([]storage.MetaUpdate, 0, len(updates))

	roles := make(map[data.RoleName]storage.MetaUpdate)
	for _, v := range updates {
		roles[v.Role] = v
	}

	builder := tuf.NewRepoBuilder(gun, cs, trustpinning.TrustPinConfig{})
	if err := loadFromStore(gun, data.CanonicalRootRole, builder, store); err != nil {
		if _, ok := err.(storage.ErrNotFound); !ok {
			return nil, err
		}
	}

	if rootUpdate, ok := roles[data.CanonicalRootRole]; ok {
		currentRootVersion := builder.GetLoadedVersion(data.CanonicalRootRole)
		if rootUpdate.Version != currentRootVersion && rootUpdate.Version != currentRootVersion+1 {
			msg := fmt.Sprintf("Root modifications must increment the version. Current %d, new %d", currentRootVersion, rootUpdate.Version)
			return nil, validation.ErrBadRoot{Msg: msg}
		}
		builder = builder.BootstrapNewBuilder()
		if err := builder.Load(data.CanonicalRootRole, rootUpdate.Data, currentRootVersion, false); err != nil {
			return nil, validation.ErrBadRoot{Msg: err.Error()}
		}

		logrus.Debug("Successfully validated root")
		updatesToApply = append(updatesToApply, rootUpdate)
	} else if !builder.IsLoaded(data.CanonicalRootRole) {
		return nil, validation.ErrValidation{Msg: "no pre-existing root and no root provided in update."}
	}

	targetsToUpdate, err := loadAndValidateTargets(gun, builder, roles, store)
	if err != nil {
		return nil, err
	}
	updatesToApply = append(updatesToApply, targetsToUpdate...)
	// there's no need to load files from the database if no targets etc...
	// were uploaded because that means they haven't been updated and
	// the snapshot will already contain the correct hashes and sizes for
	// those targets (incl. delegated targets)
	logrus.Debug("Successfully validated targets")

	// At this point, root and targets must have been loaded into the repo
	if snapshotUpdate, ok := roles[data.CanonicalSnapshotRole]; ok {
		if err := builder.Load(data.CanonicalSnapshotRole, snapshotUpdate.Data, 1, false); err != nil {
			return nil, validation.ErrBadSnapshot{Msg: err.Error()}
		}
		logrus.Debug("Successfully validated snapshot")
		updatesToApply = append(updatesToApply, roles[data.CanonicalSnapshotRole])
	} else {
		// Check:
		//   - we have a snapshot key
		//   - it matches a snapshot key signed into the root.json
		// Then:
		//   - generate a new snapshot
		//   - add it to the updates
		update, err := generateSnapshot(gun, builder, store)
		if err != nil {
			return nil, err
		}
		updatesToApply = append(updatesToApply, *update)
	}

	// generate a timestamp immediately
	update, err := generateTimestamp(gun, builder, store)
	if err != nil {
		return nil, err
	}
	return append(updatesToApply, *update), nil
}

func loadAndValidateTargets(gun data.GUN, builder tuf.RepoBuilder, roles map[data.RoleName]storage.MetaUpdate, store storage.MetaStore) ([]storage.MetaUpdate, error) {
	targetsRoles := make(utils.RoleList, 0)
	for role := range roles {
		if role == data.CanonicalTargetsRole || data.IsDelegation(role) {
			targetsRoles = append(targetsRoles, role.String())
		}
	}

	// N.B. RoleList sorts paths with fewer segments first.
	// By sorting, we'll always process shallower targets updates before deeper
	// ones (i.e. we'll load and validate targets before targets/foo). This
	// helps ensure we only load from storage when necessary in a cleaner way.
	sort.Sort(targetsRoles)

	updatesToApply := make([]storage.MetaUpdate, 0, len(targetsRoles))
	for _, role := range targetsRoles {
		// don't load parent if current role is "targets",
		// we must load all ancestor roles, starting from `targets` and working down,
		// for delegations to validate the full parent chain
		var parentsToLoad []data.RoleName
		roleName := data.RoleName(role)
		ancestorRole := roleName
		for ancestorRole != data.CanonicalTargetsRole {
			ancestorRole = ancestorRole.Parent()
			if !builder.IsLoaded(ancestorRole) {
				parentsToLoad = append(parentsToLoad, ancestorRole)
			}
		}
		for i := len(parentsToLoad) - 1; i >= 0; i-- {
			if err := loadFromStore(gun, parentsToLoad[i], builder, store); err != nil {
				// if the parent doesn't exist, just keep going - loading the role will eventually fail
				// due to it being an invalid role
				if _, ok := err.(storage.ErrNotFound); !ok {
					return nil, err
				}
			}
		}

		if err := builder.Load(roleName, roles[roleName].Data, 1, false); err != nil {
			logrus.Error("ErrBadTargets: ", err.Error())
			return nil, validation.ErrBadTargets{Msg: err.Error()}
		}
		updatesToApply = append(updatesToApply, roles[roleName])
	}

	return updatesToApply, nil
}

// generateSnapshot generates a new snapshot from the previous one in the store - this assumes all
// the other roles except timestamp have already been set on the repo, and will set the generated
// snapshot on the repo as well
func generateSnapshot(gun data.GUN, builder tuf.RepoBuilder, store storage.MetaStore) (*storage.MetaUpdate, error) {
	var prev *data.SignedSnapshot
	_, currentJSON, err := store.GetCurrent(gun, data.CanonicalSnapshotRole)
	if err == nil {
		prev = new(data.SignedSnapshot)
		if err = json.Unmarshal(currentJSON, prev); err != nil {
			logrus.Error("Failed to unmarshal existing snapshot for GUN ", gun)
			return nil, err
		}
	}

	if _, ok := err.(storage.ErrNotFound); !ok && err != nil {
		return nil, err
	}

	meta, ver, err := builder.GenerateSnapshot(prev)

	switch err.(type) {
	case nil:
		return &storage.MetaUpdate{
			Role:    data.CanonicalSnapshotRole,
			Version: ver,
			Data:    meta,
		}, nil
	case signed.ErrInsufficientSignatures, signed.ErrNoKeys, signed.ErrRoleThreshold:
		// If we cannot sign the snapshot, then we don't have keys for the snapshot,
		// and the client should have submitted a snapshot
		return nil, validation.ErrBadHierarchy{
			Missing: data.CanonicalSnapshotRole.String(),
			Msg:     "no snapshot was included in update and server does not hold current snapshot key for repository"}
	default:
		return nil, validation.ErrValidation{Msg: err.Error()}
	}
}

// generateTimestamp generates a new timestamp from the previous one in the store - this assumes all
// the other roles have already been set on the repo, and will set the generated timestamp on the repo as well
func generateTimestamp(gun data.GUN, builder tuf.RepoBuilder, store storage.MetaStore) (*storage.MetaUpdate, error) {
	var prev *data.SignedTimestamp
	_, currentJSON, err := store.GetCurrent(gun, data.CanonicalTimestampRole)

	switch err.(type) {
	case nil:
		prev = new(data.SignedTimestamp)
		if err := json.Unmarshal(currentJSON, prev); err != nil {
			logrus.Error("Failed to unmarshal existing timestamp for GUN ", gun)
			return nil, err
		}
	case storage.ErrNotFound:
		break // this is the first timestamp ever for the repo
	default:
		return nil, err
	}

	meta, ver, err := builder.GenerateTimestamp(prev)

	switch err.(type) {
	case nil:
		return &storage.MetaUpdate{
			Role:    data.CanonicalTimestampRole,
			Version: ver,
			Data:    meta,
		}, nil
	case signed.ErrInsufficientSignatures, signed.ErrNoKeys:
		// If we cannot sign the timestamp, then we don't have keys for the timestamp,
		// and the client screwed up their root
		return nil, validation.ErrBadRoot{
			Msg: fmt.Sprintf("no  timestamp keys exist on the server"),
		}
	default:
		return nil, validation.ErrValidation{Msg: err.Error()}
	}
}

func loadFromStore(gun data.GUN, roleName data.RoleName, builder tuf.RepoBuilder, store storage.MetaStore) error {
	_, metaJSON, err := store.GetCurrent(gun, roleName)
	if err != nil {
		return err
	}
	if err := builder.Load(roleName, metaJSON, 1, true); err != nil {
		return err
	}
	return nil
}
