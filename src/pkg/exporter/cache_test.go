package exporter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type CacheTestSuite struct {
	suite.Suite
}

func (c *CacheTestSuite) SetupTest() {
	CacheInit(&Opt{
		CacheDuration: 1,
	})
}

func (c *CacheTestSuite) TestCacheFunction() {
	// Empty cache should not contain key1
	_, ok := CacheGet("key1")
	c.False(ok)
	// Put keys to CacheGet
	CachePut("key1", "value1")
	CachePut("key2", "value2")
	// Get key1 should return value1
	v, ok := CacheGet("key1")
	c.True(ok)
	c.Equal("value1", v)
	// Delete key1, it should not exist anymore
	CacheDelete("key1")
	_, ok = CacheGet("key1")
	c.False(ok)
	// timeout 1 second
	time.Sleep(2 * time.Second)
	_, ok = CacheGet("key2")
	c.False(ok)
}

func TestCleanExpired(t *testing.T) {
	CacheInit(&Opt{
		CacheDuration: 60,
	})

	now := time.Now().Unix()
	c.Lock()
	c.store["valid"] = cachedValue{
		Value:      "valid",
		Expiration: now + 60,
	}
	c.store["expired"] = cachedValue{
		Value:      "expired",
		Expiration: now - 1,
	}
	c.Unlock()

	cleanExpired(now)

	c.RLock()
	_, validExists := c.store["valid"]
	_, expiredExists := c.store["expired"]
	c.RUnlock()
	if !validExists {
		t.Fatal("cleanExpired removed a valid cache entry")
	}
	if expiredExists {
		t.Fatal("cleanExpired retained an expired cache entry")
	}
}
