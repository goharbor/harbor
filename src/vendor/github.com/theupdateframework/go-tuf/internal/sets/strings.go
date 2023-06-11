package sets

func StringSliceToSet(items []string) map[string]struct{} {
	s := map[string]struct{}{}
	for _, item := range items {
		s[item] = struct{}{}
	}
	return s
}

func StringSetToSlice(items map[string]struct{}) []string {
	ret := []string{}

	for k := range items {
		ret = append(ret, k)
	}

	return ret
}

func DeduplicateStrings(items []string) []string {
	s := StringSliceToSet(items)
	return StringSetToSlice(s)
}
