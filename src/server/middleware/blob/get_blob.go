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

package blob

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	libredis "github.com/goharbor/harbor/src/lib/redis"
	"github.com/goharbor/harbor/src/server/middleware"
)

// GetBlobMiddleware cleans up zero-sized blob keys from Redis before serving blob
func GetBlobMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		log.Infof("GetBlobMiddleware: processing request %s %s", r.Method, r.URL.Path)

		// Get blob digest from request context
		blobInfo := lib.GetArtifactInfo(r.Context())
		log.Infof("GetBlobMiddleware: artifact info: %+v", blobInfo)

		if blobInfo.Digest == "" {
			log.Infof("GetBlobMiddleware: no digest found, skipping cleanup")
			next.ServeHTTP(w, r) // No digest, skip cleanup
			return
		}

		// Clean up zero-sized blob key in Redis
		key := fmt.Sprintf("blobs::%s", blobInfo.Digest)
		log.Infof("GetBlobMiddleware: checking Redis key: %s", key)

		rc, err := libredis.GetRegistryClient()
		if err != nil {
			log.Errorf("GetBlobMiddleware: failed to get Redis client for blob cleanup: %v", err)
			next.ServeHTTP(w, r) // Don't fail the request, just skip cleanup
			return
		}

		// Check if key exists and has zero size
		size, err := rc.HGet(r.Context(), key, "size").Result()
		if err != nil {
			log.Infof("GetBlobMiddleware: Redis key %s doesn't exist or error: %v", key, err)
			next.ServeHTTP(w, r)
			return
		}

		log.Infof("GetBlobMiddleware: Redis key %s has size: %s", key, size)

		if size == "0" {
			// Delete the zero-sized key
			log.Warningf("found zero-sized blob key %s for digest %s, removing to prevent pull errors", key, blobInfo.Digest)
			if err := rc.Del(r.Context(), key).Err(); err != nil {
				log.Errorf("failed to delete zero-sized blob key %s for digest %s: %v", key, blobInfo.Digest, err)
			} else {
				log.Infof("successfully cleaned up zero-sized blob key %s for digest %s", key, blobInfo.Digest)
			}
		} else {
			log.Debugf("blob key %s has valid size %s, no cleanup needed", key, size)
		}

		next.ServeHTTP(w, r)
	})
}
