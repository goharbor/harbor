package blob

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/server/middleware/requestid"
	"net/http"
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
	case models.StatusNone, models.StatusDelete, models.StatusDeleteFailed:
		err := blobController.Touch(r.Context(), bb)
		if err != nil {
			logger.Errorf("failed to update blob: %s status to StatusNone, error:%v", bb.Digest, err)
			return errors.Wrapf(err, fmt.Sprintf("the request id is: %s", r.Header.Get(requestid.HeaderXRequestID)))
		}
	case models.StatusDeleting:
		logger.Warningf(fmt.Sprintf("the asking blob is in GC, mark it as non existing, request id: %s", r.Header.Get(requestid.HeaderXRequestID)))
		return errors.New(nil).WithMessage(fmt.Sprintf("the asking blob is in GC, mark it as non existing, request id: %s", r.Header.Get(requestid.HeaderXRequestID))).WithCode(errors.NotFoundCode)
	default:
		return nil
	}
	return nil
}
