package notification

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHttpHelper(t *testing.T) {
	c1 := httpHelper.clients[insecure]
	assert.NotNil(t, c1)
	assert.Equal(t, 3*time.Second, c1.Timeout)

	c2 := httpHelper.clients[secure]
	assert.NotNil(t, c2)
	assert.Equal(t, 3*time.Second, c1.Timeout)

	_, ok := httpHelper.clients["notExists"]
	assert.False(t, ok)
}
