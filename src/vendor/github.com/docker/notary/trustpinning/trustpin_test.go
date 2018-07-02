package trustpinning

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWildcardMatch(t *testing.T) {
	testCerts := map[string][]string{
		"docker.io/library/ubuntu": {"abc"},
		"docker.io/endophage/b*":   {"def"},
		"docker.io/endophage/*":    {"xyz"},
	}

	// wildcardMatch should ONLY match wildcarded names even if a specific
	// match is present
	res, ok := wildcardMatch("docker.io/library/ubuntu", testCerts)
	require.Nil(t, res)
	require.False(t, ok)

	// wildcard match should match on segment boundaries
	res, ok = wildcardMatch("docker.io/endophage/foo", testCerts)
	require.Len(t, res, 1)
	require.Equal(t, "xyz", res[0])
	require.True(t, ok)

	// wildcardMatch should also match between segment boundaries, and take
	// the longest match it finds as the ONLY match (i.e. there is no merging
	// of key IDs when there are multiple matches).
	res, ok = wildcardMatch("docker.io/endophage/bar", testCerts)
	require.Len(t, res, 1)
	require.Equal(t, "def", res[0])
	require.True(t, ok)
}
