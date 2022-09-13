package repository

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	repo := "library/ns1/busybox"
	assert.Equal(t, "library%252Fns1%252Fbusybox", Encode(repo))
}
