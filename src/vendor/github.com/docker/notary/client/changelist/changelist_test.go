package changelist

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMemChangelist(t *testing.T) {
	cl := memChangelist{}

	c := NewTUFChange(ActionCreate, "targets", "target", "test/targ", []byte{1})

	err := cl.Add(c)
	require.Nil(t, err, "Non-nil error while adding change")

	cs := cl.List()

	require.Equal(t, 1, len(cs), "List should have returned exactly one item")
	require.Equal(t, c.Action(), cs[0].Action(), "Action mismatch")
	require.Equal(t, c.Scope(), cs[0].Scope(), "Scope mismatch")
	require.Equal(t, c.Type(), cs[0].Type(), "Type mismatch")
	require.Equal(t, c.Path(), cs[0].Path(), "Path mismatch")
	require.Equal(t, c.Content(), cs[0].Content(), "Content mismatch")

	err = cl.Clear("")
	require.Nil(t, err, "Non-nil error while clearing")

	cs = cl.List()
	require.Equal(t, 0, len(cs), "List should be empty")
}

func TestMemChangeIterator(t *testing.T) {
	cl := memChangelist{}
	it, err := cl.NewIterator()
	require.Nil(t, err, "Non-nil error from NewIterator")
	require.False(t, it.HasNext(), "HasNext returns false for empty ChangeList")

	c1 := NewTUFChange(ActionCreate, "t1", "target1", "test/targ1", []byte{1})
	cl.Add(c1)

	c2 := NewTUFChange(ActionUpdate, "t2", "target2", "test/targ2", []byte{2})
	cl.Add(c2)

	c3 := NewTUFChange(ActionUpdate, "t3", "target3", "test/targ3", []byte{3})
	cl.Add(c3)

	cs := cl.List()
	index := 0
	it, _ = cl.NewIterator()
	for it.HasNext() {
		c, err := it.Next()
		require.Nil(t, err, "Next err should be false")
		require.Equal(t, c.Action(), cs[index].Action(), "Action mismatch")
		require.Equal(t, c.Scope(), cs[index].Scope(), "Scope mismatch")
		require.Equal(t, c.Type(), cs[index].Type(), "Type mismatch")
		require.Equal(t, c.Path(), cs[index].Path(), "Path mismatch")
		require.Equal(t, c.Content(), cs[index].Content(), "Content mismatch")
		index++
	}
	require.Equal(t, index, len(cs), "Iterator produced all data in ChangeList")
	_, err = it.Next()
	require.NotNil(t, err, "Next errors gracefully when exhausted")
	var iterError IteratorBoundsError
	require.IsType(t, iterError, err, "IteratorBoundsError type")
}

func TestMemChangelistRemove(t *testing.T) {
	cl := memChangelist{}
	it, err := cl.NewIterator()
	require.Nil(t, err, "Non-nil error from NewIterator")
	require.False(t, it.HasNext(), "HasNext returns false for empty ChangeList")

	c1 := NewTUFChange(ActionCreate, "t1", "target1", "test/targ1", []byte{1})
	cl.Add(c1)

	c2 := NewTUFChange(ActionUpdate, "t2", "target2", "test/targ2", []byte{2})
	cl.Add(c2)

	c3 := NewTUFChange(ActionUpdate, "t3", "target3", "test/targ3", []byte{3})
	cl.Add(c3)

	err = cl.Remove([]int{0, 1})
	require.NoError(t, err)

	chs := cl.List()
	require.Len(t, chs, 1)
	require.EqualValues(t, "t3", chs[0].Scope())
}
