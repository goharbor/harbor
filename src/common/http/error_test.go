package http

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test case for error wrapping function.
func TestWrapError(t *testing.T) {
	err := Error{
		Code:    1,
		Message: "test",
	}

	assert.Equal(t, err.String(), "{\"code\":1,\"message\":\"test\"}")

}
