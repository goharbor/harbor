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
	"strconv"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
)

// PutBlobUploadMiddleware middleware to create Blob and ProjectBlob after PUT /v2/<name>/blobs/uploads/<session_id> success
func PutBlobUploadMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusCreated {
			return nil
		}

		logPrefix := fmt.Sprintf("[middleware][%s][blob]", r.URL.Path)

		size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if err != nil || size == 0 {
			size, err = blobController.GetAcceptedBlobSize(distribution.ParseSessionID(r.URL.Path))
		}
		if err != nil {
			log.Errorf("%s: get blob size failed, error: %v", logPrefix, err)
			return err
		}

		ctx := r.Context()

		p, err := projectController.GetByName(ctx, distribution.ParseProjectName(r.URL.Path))
		if err != nil {
			log.Errorf("%s: get project failed, error: %v", logPrefix, err)
			return err
		}

		digest := w.Header().Get("Docker-Content-Digest")
		blobID, err := blobController.Ensure(ctx, digest, "application/octet-stream", size)
		if err != nil {
			log.Errorf("%s: ensure blob %s failed, error: %v", logPrefix, digest, err)
			return err
		}

		if err := blobController.AssociateWithProjectByID(ctx, blobID, p.ProjectID); err != nil {
			log.Errorf("%s: associate blob %s with project %s failed, error: %v", logPrefix, digest, p.Name, err)
			return err
		}

		return nil
	})
}
