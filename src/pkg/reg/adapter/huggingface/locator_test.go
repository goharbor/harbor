// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package huggingface

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/cache"
)

type scanSpy struct {
	cache.Cache
	matches []string
}

func (s *scanSpy) Scan(ctx context.Context, match string) (cache.Iterator, error) {
	s.matches = append(s.matches, match)
	return s.Cache.Scan(ctx, match)
}

// The memory cache deletes expired entries it matches in Scan (lib/cache/memory Scan), so the
// sweep only has to scan the adapter's keys.
func TestSweepScansAdapterKeys(t *testing.T) {
	spy := &scanSpy{Cache: newMemoryCache()}
	sweep(context.Background(), spy)
	assert.Equal(t, []string{"huggingface:"}, spy.matches)
}

func TestKeys(t *testing.T) {
	l := &locator{registryID: 42}
	assert.Equal(t, "huggingface:42:refs:org/m", l.refsKey("org/m"))
	assert.Equal(t, "huggingface:42:scanned:org/m", l.scannedKey("org/m"))
	assert.Equal(t, "huggingface:42:snapshot:org/m:c", l.snapshotKey("org/m", "c"))
	assert.Equal(t, "huggingface:42:model:org/m", l.modelKey("org/m"))
	assert.Equal(t, "huggingface:blob:abc", l.blobKey("abc"))
}
