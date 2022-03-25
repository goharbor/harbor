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

package pkg

import (
	"sync"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/cached/artifact/redis"
)

// Define global resource manager.
var (
	// once for only init one time.
	once sync.Once
	// ArtifactMgr is the manager for artifact.
	ArtifactMgr artifact.Manager
)

// init initialize mananger for resources
func init() {
	once.Do(func() {
		cacheEnabled := config.CacheEnabled()
		initArtifactManager(cacheEnabled)
	})
}

func initArtifactManager(cacheEnabled bool) {
	artMgr := artifact.NewManager()
	// check cache enable
	if cacheEnabled {
		ArtifactMgr = redis.NewManager(artMgr)
	} else {
		ArtifactMgr = artMgr
	}
}
