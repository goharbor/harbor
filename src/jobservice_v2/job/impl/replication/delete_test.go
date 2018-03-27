package replication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxFailsOfDeleter(t *testing.T) {
	d := &Deleter{}
	assert.Equal(t, uint(3), d.MaxFails())
}

func TestValidateOfDeleter(t *testing.T) {
	d := &Deleter{}
	require.Nil(t, d.Validate(nil))
}

func TestShouldRetryOfDeleter(t *testing.T) {
	d := &Deleter{}
	assert.False(t, d.ShouldRetry())
	d.retry = true
	assert.True(t, d.ShouldRetry())
}
