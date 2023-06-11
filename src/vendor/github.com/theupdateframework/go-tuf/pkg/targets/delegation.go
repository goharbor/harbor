package targets

import (
	"errors"

	"github.com/theupdateframework/go-tuf/data"
	"github.com/theupdateframework/go-tuf/internal/sets"
	"github.com/theupdateframework/go-tuf/verify"
)

type Delegation struct {
	Delegator string
	Delegatee data.DelegatedRole
	DB        *verify.DB
}

type delegationsIterator struct {
	stack        []Delegation
	target       string
	visitedRoles map[string]struct{}
}

var ErrTopLevelTargetsRoleMissing = errors.New("tuf: top level targets role missing from top level keys DB")

// NewDelegationsIterator initialises an iterator with a first step
// on top level targets.
func NewDelegationsIterator(target string, topLevelKeysDB *verify.DB) (*delegationsIterator, error) {
	targetsRole := topLevelKeysDB.GetRole("targets")
	if targetsRole == nil {
		return nil, ErrTopLevelTargetsRoleMissing
	}

	i := &delegationsIterator{
		target: target,
		stack: []Delegation{
			{
				Delegatee: data.DelegatedRole{
					Name:      "targets",
					KeyIDs:    sets.StringSetToSlice(targetsRole.KeyIDs),
					Threshold: targetsRole.Threshold,
				},
				DB: topLevelKeysDB,
			},
		},
		visitedRoles: make(map[string]struct{}),
	}
	return i, nil
}

func (d *delegationsIterator) Next() (value Delegation, ok bool) {
	if len(d.stack) == 0 {
		return Delegation{}, false
	}
	delegation := d.stack[len(d.stack)-1]
	d.stack = d.stack[:len(d.stack)-1]

	// 5.6.7.1: If this role has been visited before, then skip this role (so
	// that cycles in the delegation graph are avoided).
	roleName := delegation.Delegatee.Name
	if _, ok := d.visitedRoles[roleName]; ok {
		return d.Next()
	}
	d.visitedRoles[roleName] = struct{}{}

	// 5.6.7.2 trim delegations to visit, only the current role and its delegations
	// will be considered
	// https://github.com/theupdateframework/specification/issues/168
	if delegation.Delegatee.Terminating {
		// Empty the stack.
		d.stack = d.stack[0:0]
	}
	return delegation, true
}

func (d *delegationsIterator) Add(roles []data.DelegatedRole, delegator string, db *verify.DB) error {
	for i := len(roles) - 1; i >= 0; i-- {
		// Push the roles onto the stack in reverse so we get an preorder traversal
		// of the delegations graph.
		r := roles[i]
		matchesPath, err := r.MatchesPath(d.target)
		if err != nil {
			return err
		}
		if matchesPath {
			delegation := Delegation{
				Delegator: delegator,
				Delegatee: r,
				DB:        db,
			}
			d.stack = append(d.stack, delegation)
		}
	}

	return nil
}
