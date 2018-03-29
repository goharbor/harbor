package replication

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxFailsOfReplicator(t *testing.T) {
	r := &Replicator{}
	assert.Equal(t, uint(0), r.MaxFails())
}

func TestValidateOfReplicator(t *testing.T) {
	r := &Replicator{}
	require.Nil(t, r.Validate(nil))
}

func TestShouldRetryOfReplicator(t *testing.T) {
	r := &Replicator{}
	assert.False(t, r.ShouldRetry())
}
