package logger

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// TestFileSweeperFactory
func TestFileSweeperFactory(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"work_dir", "/tmp"})
	ois = append(ois, OptionItem{"duration", 2})

	_, err := FileSweeperFactory(ois...)
	require.Nil(t, err)
}

// TestFileSweeperFactoryErr
func TestFileSweeperFactoryErr(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"duration", 2})

	_, err := FileSweeperFactory(ois...)
	require.NotNil(t, err)
}

// TestDBSweeperFactory
func TestDBSweeperFactory(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"duration", 2})

	_, err := DBSweeperFactory(ois...)
	require.Nil(t, err)
}
