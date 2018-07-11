package match

import (
	"reflect"
	"testing"
)

func TestSuffixAnyIndex(t *testing.T) {
	for id, test := range []struct {
		suffix     string
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
			"cab",
			0,
			[]int{3},
		},
		{
			"ab",
			[]rune{'.'},
			"qw.cdab.efg",
			3,
			[]int{4},
		},
	} {
		p := NewSuffixAny(test.suffix, test.separators)
		index, segments := p.Index(test.fixture)
		if index != test.index {
			t.Errorf("#%d unexpected index: exp: %d, act: %d", id, test.index, index)
		}
		if !reflect.DeepEqual(segments, test.segments) {
			t.Errorf("#%d unexpected segments: exp: %v, act: %v", id, test.segments, segments)
		}
	}
}
