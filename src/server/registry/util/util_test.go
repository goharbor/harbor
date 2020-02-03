package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateLinkEntry(t *testing.T) {
	u1, err := SetLinkHeader("/v2/hello-wrold/tags/list", 10, "v10")
	assert.Nil(t, err)
	assert.Equal(t, u1, "</v2/hello-wrold/tags/list?last=v10&n=10>; rel=\"next\"")

	u2, err := SetLinkHeader("/v2/hello-wrold/tags/list", 5, "v5")
	assert.Nil(t, err)
	assert.Equal(t, u2, "</v2/hello-wrold/tags/list?last=v5&n=5>; rel=\"next\"")

}

func TestIndexString(t *testing.T) {
	a := []string{"B", "A", "C", "E"}

	assert.True(t, IndexString(a, "E") == 3)
	assert.True(t, IndexString(a, "B") == 1)
	assert.True(t, IndexString(a, "A") == 0)
	assert.True(t, IndexString(a, "C") == 2)

	assert.True(t, IndexString(a, "Z") == -1)
	assert.True(t, IndexString(a, "") == -1)
}
