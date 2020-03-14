package notification

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/api/event"
	pkg_art "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type NotificatoinMiddlewareTestSuite struct {
	suite.Suite
}

func (suite *NotificatoinMiddlewareTestSuite) TestMiddleware() {
	next := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			notification.AddEvent(r.Context(), &event.DeleteArtifactEventMetadata{
				Ctx: context.Background(),
				Artifact: &pkg_art.Artifact{
					ProjectID:      1,
					RepositoryID:   2,
					RepositoryName: "library/hello-world",
				},
				Tags: []string{"latest"},
			})
		})
	}
	path := fmt.Sprintf("/v2/library/photon/manifests/latest")
	req := httptest.NewRequest(http.MethodPatch, path, nil)
	res := httptest.NewRecorder()
	Middleware()(next()).ServeHTTP(res, req)
	suite.Equal(http.StatusAccepted, res.Code)
}

func (suite *NotificatoinMiddlewareTestSuite) TestMiddlewareMustNotify() {
	next := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			notification.AddEvent(r.Context(), &event.DeleteArtifactEventMetadata{
				Ctx: context.Background(),
				Artifact: &pkg_art.Artifact{
					ProjectID:      1,
					RepositoryID:   2,
					RepositoryName: "library/hello-world",
				},
				Tags: []string{"latest"},
			}, true)
		})
	}
	path := fmt.Sprintf("/v2/library/photon/manifests/latest")
	req := httptest.NewRequest(http.MethodPatch, path, nil)
	res := httptest.NewRecorder()
	Middleware()(next()).ServeHTTP(res, req)
	suite.Equal(http.StatusInternalServerError, res.Code)
}

func TestNotificatoinMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &NotificatoinMiddlewareTestSuite{})
}
