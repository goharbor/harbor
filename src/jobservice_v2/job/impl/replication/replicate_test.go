package replication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxFails(t *testing.T) {
	r := &Replicator{}
	assert.Equal(t, uint(3), r.MaxFails())
}

func TestValidate(t *testing.T) {
	r := &Replicator{}
	require.Nil(t, r.Validate(nil))
}

func TestShouldRetry(t *testing.T) {
	r := &Replicator{}
	assert.False(t, r.retry)
	r.retry = true
	assert.True(t, r.retry)
}
