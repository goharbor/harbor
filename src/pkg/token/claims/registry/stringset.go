package registry

// StringSet is a useful type for looking up strings.
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
