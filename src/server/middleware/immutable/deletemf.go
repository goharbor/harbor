package immutable

import (
	"errors"
	"fmt"
	common_util "github.com/goharbor/harbor/src/common/utils"
	internal_errors "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/tag"
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

// handleDelete ...
func handleDelete(req *http.Request) error {
	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	if !ok {
		return errors.New("cannot get the manifest information from request context")
	}

	_, repoName := common_util.ParseRepository(mf.Repository)
	total, repos, err := repository.Mgr.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"Name": mf.Repository,
		},
	})
	if err != nil {
		return err
	}
	if total == 0 {
		return nil
	}

	total, afs, err := artifact.Mgr.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"ProjectID":    mf.ProjectID,
			"RepositoryID": repos[0].RepositoryID,
			"Digest":       mf.Digest,
		},
	})
	if err != nil {
		return err
	}
	if total == 0 {
		return nil
	}

	total, tags, err := tag.Mgr.List(req.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"ArtifactID": afs[0].ID,
		},
	})
	if err != nil {
		return err
	}
	if total == 0 {
		return nil
	}

	for _, tag := range tags {
		var matched bool
		matched, err = rule.NewRuleMatcher(mf.ProjectID).Match(art.Candidate{
			Repository:  repoName,
			Tag:         tag.Name,
			NamespaceID: mf.ProjectID,
		})
		if err != nil {
			return err
		}
		if matched {
			return NewErrImmutable(repoName, tag.Name)
		}
	}

	return nil
}
