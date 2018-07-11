package match

import (
	"reflect"
	"testing"
)

func TestPrefixAnyIndex(t *testing.T) {
	for id, test := range []struct {
		prefix     string
		separators []rune
		fixture    string
		index      int
		segments   []int
	}{
		{
			"ab",
			[]rune{'.'},
			"ab",
			0,
			[]int{2},
		},
		{
			"ab",
			[]rune{'.'},
			"abc",
			0,
			[]int{2, 3},
		},
		{
			"ab",
			[]rune{'.'},
			"qw.abcd.efg",
			3,
			[]int{2, 3, 4},
		},
	} {
		p := NewPrefixAny(test.prefix, test.separators)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}
