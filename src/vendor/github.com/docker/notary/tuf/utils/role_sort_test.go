package utils

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoleListLen(t *testing.T) {
	rl := RoleList{"foo", "bar"}
	require.Equal(t, 2, rl.Len())
}

func TestRoleListLess(t *testing.T) {
	rl := RoleList{"foo", "foo/bar", "bar/foo"}
	require.True(t, rl.Less(0, 1))
	require.False(t, rl.Less(1, 2))
	require.True(t, rl.Less(2, 1))
}

func TestRoleListSwap(t *testing.T) {
	rl := RoleList{"foo", "bar"}
	rl.Swap(0, 1)
	require.Equal(t, "bar", rl[0])
	require.Equal(t, "foo", rl[1])
}

func TestRoleListSort(t *testing.T) {
	rl := RoleList{"foo/bar", "foo", "bar", "bar/foo/baz", "bar/foo"}
	sort.Sort(rl)
	for i, s := range rl {
		if i == 0 || i == 1 {
			segs := strings.Split(s, "/")
			require.Len(t, segs, 1)
		} else if i == 2 || i == 3 {
			segs := strings.Split(s, "/")
			require.Len(t, segs, 2)
		} else if i == 4 {
			segs := strings.Split(s, "/")
			require.Len(t, segs, 3)
		} else {
			// there are elements present that shouldn't be there
			t.Fail()
		}
	}
}
