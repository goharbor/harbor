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
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	blob_models "github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/requestid"
	"net/http"
	"strconv"
)

// PutBlobUploadMiddleware middleware is to update the blob status according to the different situation before the request passed into proxy(distribution).
// And it creates Blob and ProjectBlob after PUT /v2/<name>/blobs/uploads/<session_id>?digest=<digest> success - http.StatusCreated
// Why to use the middleware to handle blob status?
// 1, As Put blob will always happen after head blob gets a 404, but the 404 could be caused by blob status is deleting, which is marked by GC.
// 2, It has to deal with the concurrence blob push.
func PutBlobUploadMiddleware() func(http.Handler) http.Handler {

	before := middleware.BeforeRequest(func(r *http.Request) error {
		ctx := r.Context()
		logger := log.G(ctx)

		v := r.URL.Query()
		digest := v.Get("digest")

		// digest empty is handled by the blob controller GET method
		bb, err := blobController.Get(r.Context(), digest)
		if err != nil {
			if errors.IsNotFoundErr(err) {
				return nil
			}
			return err
		}

		switch bb.Status {
		case blob_models.StatusNone, blob_models.StatusDelete, blob_models.StatusDeleteFailed:
			err := blobController.Touch(r.Context(), bb)
			if err != nil {
				logger.Errorf("failed to update blob: %s status to StatusNone, error:%v", bb.Digest, err)
				return errors.Wrapf(err, fmt.Sprintf("the request id is: %s", r.Header.Get(requestid.HeaderXRequestID)))
			}
		case blob_models.StatusDeleting:
			logger.Warningf(fmt.Sprintf("the asking blob is in GC, mark it as non existing, request id: %s", r.Header.Get(requestid.HeaderXRequestID)))
			return errors.New(nil).WithMessage(fmt.Sprintf("the asking blob is in GC, mark it as non existing, request id: %s", r.Header.Get(requestid.HeaderXRequestID))).WithCode(errors.NotFoundCode)
		default:
			return nil
		}
		return nil
	})

	after := middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusCreated {
			return nil
		}

		ctx := r.Context()

		logger := log.G(ctx).WithFields(log.Fields{"middleware": "blob"})

		size, err := strconv.ParseInt(r.Header.Get("Content-Length"), 10, 64)
		if err != nil || size == 0 {
			size, err = blobController.GetAcceptedBlobSize(distribution.ParseSessionID(r.URL.Path))
		}
		if err != nil {
			logger.Errorf("get blob size failed, error: %v", err)
			return err
		}

		p, err := projectController.GetByName(ctx, distribution.ParseProjectName(r.URL.Path))
		if err != nil {
			logger.Errorf("get project failed, error: %v", err)
			return err
		}

		digest := w.Header().Get("Docker-Content-Digest")
		blobID, err := blobController.Ensure(ctx, digest, "application/octet-stream", size)
		if err != nil {
			logger.Errorf("ensure blob %s failed, error: %v", digest, err)
			return err
		}

		if err := blobController.AssociateWithProjectByID(ctx, blobID, p.ProjectID); err != nil {
			logger.Errorf("associate blob %s with project %s failed, error: %v", digest, p.Name, err)
			return err
		}

		return nil
	})

	return middleware.Chain(before, after)
}
