package logger

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// TestFileGetterFactory
func TestFileGetterFactory(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"other_key1", 11})
	ois = append(ois, OptionItem{"base_dir", "/tmp"})
	ois = append(ois, OptionItem{"other_key2", ""})

	_, err := FileGetterFactory(ois...)
	require.Nil(t, err)
}

// TestFileGetterFactoryErr1
func TestFileGetterFactoryErr1(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"other_key1", 11})

	_, err := FileGetterFactory(ois...)
	require.NotNil(t, err)
}

// TestDBGetterFactory
func TestDBGetterFactory(t *testing.T) {
	ois := make([]OptionItem, 0)

	_, err := DBGetterFactory(ois...)
	require.Nil(t, err)
}
