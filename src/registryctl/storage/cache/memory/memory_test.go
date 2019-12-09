package memory

import (
	"testing"

	"github.com/goharbor/harbor/src/registryctl/storage/cache/cachecheck"
)

// TestInMemoryBlobInfoCache checks the in memory implementation is working
// correctly.
func TestInMemoryBlobInfoCache(t *testing.T) {
	cachecheck.CheckBlobDescriptorCache(t, NewInMemoryBlobDescriptorCacheProvider())
}
