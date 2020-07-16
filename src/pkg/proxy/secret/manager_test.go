package secret

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManger(t *testing.T) {
	manager := GetManager()
	rn1 := "project1/golang"
	assert.False(t, manager.Verify("whatever", rn1))
	s1 := manager.Generate(rn1)
	s2 := manager.Generate(rn1)
	assert.False(t, s1 == s2)

	assert.False(t, manager.Verify(s1, "project1/donotexist"))
	assert.True(t, manager.Verify(s1, rn1))
	// A secret can be used only once.
	assert.False(t, manager.Verify(s1, rn1))
	manager2 := GetManager()
	assert.Equal(t, manager2, manager)
}

func TestExpiration(t *testing.T) {
	manager := createManager(1 * time.Second)
	rn1 := "project1/golang"
	s := manager.Generate(rn1)
	// Sleep till the secret expires
	time.Sleep(2 * time.Second)
	assert.False(t, manager.Verify(s, rn1))
}
