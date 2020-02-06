package immutable

import (
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/api/artifact"
	common_util "github.com/goharbor/harbor/src/common/utils"
	internal_errors "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

// MiddlewarePush ...
func MiddlewarePush() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			if err := handlePush(req); err != nil {
				var e *ErrImmutable
				if errors.As(err, &e) {
					pkgE := internal_errors.New(e).WithCode(internal_errors.PreconditionCode)
					msg := internal_errors.NewErrs(pkgE).Error()
					http.Error(rw, msg, http.StatusPreconditionFailed)
					return
				}
				pkgE := internal_errors.New(fmt.Errorf("error occurred when to handle request in immutable handler: %v", err)).WithCode(internal_errors.GeneralCode)
				msg := internal_errors.NewErrs(pkgE).Error()
				http.Error(rw, msg, http.StatusInternalServerError)
			}
			next.ServeHTTP(rw, req)
		})
	}
}

// handlePush ...
// If the pushing image matched by any of immutable rule, will have to whether it is the first time to push it,
// as the immutable rule only impacts the existing tag.
func handlePush(req *http.Request) error {
	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	if !ok {
		return errors.New("cannot get the manifest information from request context")
	}

	af, err := artifact.Ctl.GetByReference(req.Context(), mf.Repository, mf.Tag, &artifact.Option{
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
		// push a existing immutable tag, reject th e request
		if tag.Name == mf.Tag && tag.Immutable {
			return NewErrImmutable(repoName, mf.Tag)
		}
	}

	return nil
}
