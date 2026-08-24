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
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/server/middleware/requestid"
)

// probeBlob handles config/layer and manifest status in the PUT Blob & Manifest middleware, and update the status before it passed into proxy(distribution).
func probeBlob(r *http.Request, digest string) error {
	logger := log.G(r.Context())

	// digest empty is handled by the blob controller GET method
	bb, err := blobController.Get(r.Context(), digest)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return nil
		}
		return err
	}

	switch bb.Status {
	case models.StatusNone:
		if err := touchIfDue(r, bb); err != nil {
			return err
		}
	case models.StatusDelete, models.StatusDeleteFailed:
		// Rescue the blob from GC candidacy: this status transition must
		// always be written.
		if err := blobController.Touch(r.Context(), bb); err != nil {
			logger.Errorf("failed to update blob: %s status to StatusNone, error:%v", bb.Digest, err)
			return errors.Wrapf(err, "the request id is: %s", r.Header.Get(requestid.HeaderXRequestID))
		}
	case models.StatusDeleting:
		now := time.Now().UTC()
		// if the deleting exceed 2 hours, marks the blob as StatusDeleteFailed
		if now.Sub(bb.UpdateTime) > time.Duration(config.GetGCTimeWindow())*time.Hour {
			if err := blob.Ctl.Fail(r.Context(), bb); err != nil {
				log.Errorf("failed to update blob: %s status to StatusDeleteFailed, error:%v", bb.Digest, err)
				return errors.Wrapf(err, "the request id is: %s", r.Header.Get(requestid.HeaderXRequestID))
			}
			// StatusDeleteFailed => StatusNone, and then let the proxy to handle manifest upload
			return probeBlob(r, digest)
		}
		return errors.New(nil).WithMessagef("the asking blob is in GC, mark it as non existing, request id: %s", r.Header.Get(requestid.HeaderXRequestID)).WithCode(errors.NotFoundCode)
	default:
		return nil
	}
	return nil
}

// touchIsDue reports whether a healthy (StatusNone) blob's update_time is
// stale enough that Touch must rewrite it. For StatusNone blobs the status
// transition is a no-op: the write exists only to refresh update_time so
// that the GC useless-blob sweep (update_time <= now() - GC time window)
// cannot select a blob that is still in use. If the row was written within
// the last half window, the stored timestamp is still at least half a
// window away from the GC boundary and rewriting it changes nothing.
// Under concurrent pushes of images sharing base layers, this redundant
// write is the hottest row contention on the database (every HEAD and PUT
// of every shared layer), so it is skipped while fresh.
func touchIsDue(bb *models.Blob) bool {
	return time.Since(bb.UpdateTime) > time.Duration(config.GetGCTimeWindow())*time.Hour/2
}

// touchIfDue is the shared conditional-Touch path for healthy (StatusNone)
// blobs: the blob is already healthy, Touch only refreshes update_time to
// keep it out of GC's candidate window, so the write is skipped while the
// timestamp is still fresh (see touchIsDue).
func touchIfDue(r *http.Request, bb *models.Blob) error {
	if !touchIsDue(bb) {
		return nil
	}
	if err := blobController.Touch(r.Context(), bb); err != nil {
		log.G(r.Context()).Errorf("failed to update blob: %s status to StatusNone, error:%v", bb.Digest, err)
		return errors.Wrapf(err, "the request id is: %s", r.Header.Get(requestid.HeaderXRequestID))
	}
	return nil
}
