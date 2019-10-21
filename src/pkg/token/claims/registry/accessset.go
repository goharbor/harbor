package registry

import (
	"github.com/docker/distribution/registry/auth"
)

// AccessSet ...
type AccessSet map[auth.Resource]actionSet

// Contains ...
func (s AccessSet) Contains(access auth.Access) bool {
	actionSet, ok := s[access.Resource]
	if ok {
		return actionSet.contains(access.Action)
	}

	return false
}
