package exporter

import (
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
