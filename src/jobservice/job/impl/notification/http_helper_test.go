package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHttpHelper(t *testing.T) {
	c1 := httpHelper.clients[insecure]
	assert.NotNil(t, c1)

	c2 := httpHelper.clients[secure]
	assert.NotNil(t, c2)

	_, ok := httpHelper.clients["notExists"]
	assert.False(t, ok)
}
