package blob

import (
	"fmt"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	blob_models "github.com/goharbor/harbor/src/pkg/blob/models"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

// HeadBlobMiddleware intercept the head blob request
func HeadBlobMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(rw http.ResponseWriter, req *http.Request, next http.Handler) {
		if err := handleHead(req); err != nil {
			serror.SendError(rw, err)
			return
		}
		next.ServeHTTP(rw, req)
	})
}

// handleHead ...
func handleHead(req *http.Request) error {
	none := lib.ArtifactInfo{}
	art := lib.GetArtifactInfo(req.Context())
	if art == none {
		return errors.New("cannot get the artifact information from request context").WithCode(errors.NotFoundCode)
	}

	bb, err := blob.Ctl.Get(req.Context(), art.Digest)
	if err != nil {
		return err
	}

	switch bb.Status {
	case blob_models.StatusNone, blob_models.StatusDelete:
		bb.Status = blob_models.StatusNone
		count, err := blob.Ctl.Touch(req.Context(), bb)
		if err != nil {
			log.Errorf("failed to update blob: %s status to None, error:%v", art.Digest, err)
			return err
		}
		if count == 0 {
			return errors.New("the asking blob is in GC, mark it as non existing").WithCode(errors.NotFoundCode)
		}
	case blob_models.StatusDeleting, blob_models.StatusDeleteFailed:
		return errors.New("the asking blob is in GC, mark it as non existing").WithCode(errors.NotFoundCode)
	default:
		return errors.New(nil).WithMessage(fmt.Sprintf("wrong blob status, %s", bb.Status))
	}
	return nil
}
