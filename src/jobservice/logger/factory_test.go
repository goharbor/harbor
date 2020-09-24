package logger

import (
	"github.com/stretchr/testify/require"
	"testing"
)

// TestFileFactory
func TestFileFactory(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"level", "DEBUG"})
	ois = append(ois, OptionItem{"base_dir", "/tmp"})
	ois = append(ois, OptionItem{"filename", "test.out"})
	ois = append(ois, OptionItem{"depth", 5})

	ff, err := FileFactory(ois...)
	require.Nil(t, err)

	if closer, ok := ff.(Closer); ok {
		_ = closer.Close()
	}
}

// TestFileFactoryErr1
func TestFileFactoryErr1(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"level", "DEBUG"})
	ois = append(ois, OptionItem{"filename", "test.out"})

	_, err := FileFactory(ois...)
	require.NotNil(t, err)
}

// TestFileFactoryErr2
func TestFileFactoryErr2(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"level", "DEBUG"})
	ois = append(ois, OptionItem{"base_dir", "/tmp"})

	_, err := FileFactory(ois...)
	require.NotNil(t, err)
}

// TestStdFactory
func TestStdFactory(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"level", "DEBUG"})
	ois = append(ois, OptionItem{"output", "std_out"})
	ois = append(ois, OptionItem{"depth", 5})

	_, err := StdFactory(ois...)
	require.Nil(t, err)
}

// TestDBFactory
func TestDBFactory(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"level", "DEBUG"})
	ois = append(ois, OptionItem{"key", "key_db_logger_unit_text"})
	ois = append(ois, OptionItem{"depth", 5})

	_, err := DBFactory(ois...)
	require.Nil(t, err)
}

// TestDBFactoryErr1
func TestDBFactoryErr1(t *testing.T) {
	ois := make([]OptionItem, 0)
	ois = append(ois, OptionItem{"level", "DEBUG"})
	ois = append(ois, OptionItem{"depth", 5})

	_, err := DBFactory(ois...)
	require.NotNil(t, err)
}
