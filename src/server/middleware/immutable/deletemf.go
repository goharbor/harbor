package immutable

import (
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/api/artifact"
	common_util "github.com/goharbor/harbor/src/common/utils"
	internal_errors "github.com/goharbor/harbor/src/internal/error"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

// MiddlewareDelete ...
func MiddlewareDelete() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if err := handleDelete(req); err != nil {
				var e *ErrImmutable
				if errors.As(err, &e) {
					pkgE := internal_errors.New(e).WithCode(internal_errors.PreconditionCode)
					serror.SendError(rw, pkgE)
					return
				}
				pkgE := internal_errors.New(fmt.Errorf("error occurred when to handle request in immutable handler: %v", err)).WithCode(internal_errors.GeneralCode)
				serror.SendError(rw, pkgE)
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}

// handleDelete ...
func handleDelete(req *http.Request) error {
	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	if !ok {
		return errors.New("cannot get the manifest information from request context")
	}

	af, err := artifact.Ctl.GetByReference(req.Context(), mf.Repository, mf.Digest, &artifact.Option{
		WithTag:   true,
		TagOption: &artifact.TagOption{WithImmutableStatus: true},
	})
	if err != nil {
		if internal_errors.IsErr(err, internal_errors.NotFoundCode) {
			return nil
		}
		return err
	}

	_, repoName := common_util.ParseRepository(mf.Repository)
	for _, tag := range af.Tags {
		if tag.Immutable {
			return NewErrImmutable(repoName, tag.Name)
		}
	}

	return nil
}
