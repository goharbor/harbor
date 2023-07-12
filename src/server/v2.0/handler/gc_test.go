package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateWorkers(t *testing.T) {
	assert.False(t, validateWorkers(0))
	assert.False(t, validateWorkers(10))
	assert.False(t, validateWorkers(-1))
	assert.True(t, validateWorkers(1))
	assert.True(t, validateWorkers(5))
}
