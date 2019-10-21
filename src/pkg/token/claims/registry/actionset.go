package registry

// actionSet is a special type of stringSet.
type actionSet struct {
	stringSet
}

func newActionSet(actions ...string) actionSet {
	return actionSet{newStringSet(actions...)}
}

func (s actionSet) contains(action string) bool {
	return s.stringSet.contains(action)
}
