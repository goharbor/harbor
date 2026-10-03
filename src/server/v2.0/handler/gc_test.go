package handler

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
)

func TestValidateWorkers(t *testing.T) {
	// no limit configured
	assert.False(t, validateWorkers(0, 0))
	assert.False(t, validateWorkers(-1, 0))
	assert.True(t, validateWorkers(1, 0))
	assert.True(t, validateWorkers(5, 0))
	assert.True(t, validateWorkers(15, 0))
	assert.True(t, validateWorkers(100, 0))

	// limit configured
	assert.False(t, validateWorkers(0, 8))
	assert.False(t, validateWorkers(-1, 8))
	assert.True(t, validateWorkers(1, 8))
	assert.True(t, validateWorkers(8, 8))
	assert.False(t, validateWorkers(9, 8))
	assert.False(t, validateWorkers(24, 8))
}

func TestParseWorkers(t *testing.T) {
	t.Run("no limit configured", func(t *testing.T) {
		workers, err := parseWorkers(json.Number("24"))
		require.NoError(t, err)
		assert.Equal(t, 24, workers)
	})

	t.Run("within the configured limit", func(t *testing.T) {
		t.Setenv("GC_MAX_WORKERS", "8")
		workers, err := parseWorkers(json.Number("8"))
		require.NoError(t, err)
		assert.Equal(t, 8, workers)
	})

	t.Run("above the configured limit", func(t *testing.T) {
		t.Setenv("GC_MAX_WORKERS", "8")
		_, err := parseWorkers(json.Number("24"))
		require.Error(t, err)
		assert.Equal(t, errors.BadRequestCode, errors.ErrCode(err))
		// the configured maximum is part of the message so that it surfaces in the UI alert
		assert.Contains(t, err.Error(), "at most 8")
	})

	t.Run("invalid limit is ignored", func(t *testing.T) {
		t.Setenv("GC_MAX_WORKERS", "not-a-number")
		workers, err := parseWorkers(json.Number("24"))
		require.NoError(t, err)
		assert.Equal(t, 24, workers)
	})

	t.Run("non positive workers", func(t *testing.T) {
		_, err := parseWorkers(json.Number("0"))
		require.Error(t, err)
		assert.Equal(t, errors.BadRequestCode, errors.ErrCode(err))
		assert.Contains(t, err.Error(), "greater than 0")
	})

	t.Run("not an integer", func(t *testing.T) {
		_, err := parseWorkers(json.Number("1.5"))
		require.Error(t, err)
		assert.Equal(t, errors.BadRequestCode, errors.ErrCode(err))
	})
}
