package secret

import (
	"fmt"
	"sync/atomic"
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
	manager := createManager(50*time.Millisecond, defaultCap, defaultGCInterval)
	rn1 := "project1/golang"
	s := manager.Generate(rn1)
	assert.Eventually(t, func() bool {
		return !manager.Verify(s, rn1)
	}, 1*time.Second, 10*time.Millisecond)
}

func TestGC(t *testing.T) {
	manager := createManager(50*time.Millisecond, 10, 50*time.Millisecond).(*mgr)
	for i := range 10 {
		rn := fmt.Sprintf("project%d/golang", i)
		manager.Generate(rn)
	}
	assert.Equal(t, uint64(10), manager.size)
	// Wait for the initial 10 items to expire before triggering GC with new items
	time.Sleep(60 * time.Millisecond)
	for i := range 1000 {
		rn := fmt.Sprintf("project%d/redis", i)
		manager.Generate(rn)
	}
	assert.Equal(t, uint64(1000), atomic.LoadUint64(&manager.size))
	assert.Eventually(t, func() bool {
		return atomic.LoadUint64(&manager.size) == 0
	}, 2*time.Second, 10*time.Millisecond)
}
