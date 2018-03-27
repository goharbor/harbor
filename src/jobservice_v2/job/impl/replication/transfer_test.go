package replication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxFailsOfTransfer(t *testing.T) {
	r := &Transfer{}
	assert.Equal(t, uint(3), r.MaxFails())
}

func TestValidateOfTransfer(t *testing.T) {
	r := &Transfer{}
	require.Nil(t, r.Validate(nil))
}

func TestShouldRetryOfTransfer(t *testing.T) {
	r := &Transfer{}
	assert.False(t, r.ShouldRetry())
	r.retry = true
	assert.True(t, r.ShouldRetry())
}
