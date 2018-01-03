package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateStack(t *testing.T) {
	s := NewStack()
	require.NotNil(t, s)
}

func TestPush(t *testing.T) {
	s := NewStack()
	s.Push("foo")
	require.Len(t, s.s, 1)
	require.Equal(t, "foo", s.s[0])
}

func TestPop(t *testing.T) {
	s := NewStack()
	s.Push("foo")
	i, err := s.Pop()
	require.NoError(t, err)
	require.Len(t, s.s, 0)
	require.IsType(t, "", i)
	require.Equal(t, "foo", i)
}

func TestPopEmpty(t *testing.T) {
	s := NewStack()
	_, err := s.Pop()
	require.Error(t, err)
	require.IsType(t, ErrEmptyStack{}, err)
}

func TestPopString(t *testing.T) {
	s := NewStack()
	s.Push("foo")
	i, err := s.PopString()
	require.NoError(t, err)
	require.Len(t, s.s, 0)
	require.Equal(t, "foo", i)
}

func TestPopStringWrongType(t *testing.T) {
	s := NewStack()
	s.Push(123)
	_, err := s.PopString()
	require.Error(t, err)
	require.IsType(t, ErrBadTypeCast{}, err)
	require.Len(t, s.s, 1)
}

func TestPopStringEmpty(t *testing.T) {
	s := NewStack()
	_, err := s.PopString()
	require.Error(t, err)
	require.IsType(t, ErrEmptyStack{}, err)
}

func TestEmpty(t *testing.T) {
	s := NewStack()
	require.True(t, s.Empty())
	s.Push("foo")
	require.False(t, s.Empty())
	s.Pop()
	require.True(t, s.Empty())
}
