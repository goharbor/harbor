package registry

import (
	"github.com/docker/distribution/registry/auth"
)

// Accesses ...
type Accesses map[auth.Resource]actions

// Contains ...
func (s Accesses) Contains(access auth.Access) bool {
	actionSet, ok := s[access.Resource]
	if ok {
		return actionSet.contains(access.Action)
	}

	return false
}

type actions struct {
	stringSet
}

func newActions(set ...string) actions {
	return actions{newStringSet(set...)}
}

func (s actions) contains(action string) bool {
	return s.stringSet.contains(action)
}

type stringSet map[string]struct{}

func newStringSet(keys ...string) stringSet {
	ss := make(stringSet, len(keys))
	ss.add(keys...)
	return ss
}

func (ss stringSet) add(keys ...string) {
	for _, key := range keys {
		ss[key] = struct{}{}
	}
}

func (ss stringSet) contains(key string) bool {
	_, ok := ss[key]
	return ok
}
