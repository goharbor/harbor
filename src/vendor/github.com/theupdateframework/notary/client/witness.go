package client

import (
	"github.com/theupdateframework/notary/client/changelist"
	"github.com/theupdateframework/notary/tuf"
	"github.com/theupdateframework/notary/tuf/data"
)

// Witness creates change objects to witness (i.e. re-sign) the given
// roles on the next publish. One change is created per role
func (r *repository) Witness(roles ...data.RoleName) ([]data.RoleName, error) {
	var err error
	successful := make([]data.RoleName, 0, len(roles))
	for _, role := range roles {
		// scope is role
		c := changelist.NewTUFChange(
			changelist.ActionUpdate,
			role,
			changelist.TypeWitness,
			"",
			nil,
		)
		err = r.changelist.Add(c)
		if err != nil {
			break
		}
		successful = append(successful, role)
	}
	return successful, err
}

func witnessTargets(repo *tuf.Repo, invalid *tuf.Repo, role data.RoleName) error {
	if r, ok := repo.Targets[role]; ok {
		// role is already valid, mark for re-signing/updating
		r.Dirty = true
		return nil
	}

	if roleObj, err := repo.GetDelegationRole(role); err == nil && invalid != nil {
		// A role with a threshold > len(keys) is technically invalid, but we let it build in the builder because
		// we want to be able to download the role (which may still have targets on it), add more keys, and then
		// witness the role, thus bringing it back to valid.  However, if no keys have been added before witnessing,
		// then it is still an invalid role, and can't be witnessed because nothing can bring it back to valid.
		if roleObj.Threshold > len(roleObj.Keys) {
			return data.ErrInvalidRole{
				Role:   role,
				Reason: "role does not specify enough valid signing keys to meet its required threshold",
			}
		}
		if r, ok := invalid.Targets[role]; ok {
			// role is recognized but invalid, move to valid data and mark for re-signing
			repo.Targets[role] = r
			r.Dirty = true
			return nil
		}
	}
	// role isn't recognized, even as invalid
	return data.ErrInvalidRole{
		Role:   role,
		Reason: "this role is not known",
	}
}
