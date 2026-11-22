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

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/optimizer"
	"github.com/goharbor/harbor/src/controller/project"
	liberrors "github.com/goharbor/harbor/src/lib/errors"
	dockerfileoptdao "github.com/goharbor/harbor/src/pkg/dockerfileoptimization/dao"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/dockerfile"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
)

// stubOptimizerCtl is a simple in-test stub for optimizer.Controller. Only the
// Optimize and CheckAvailability methods are used by the dockerfile handler; the
// embedded interface panics on any other call.
type stubOptimizerCtl struct {
	optimizer.Controller
	executionID int64
	err         error
	// captured arguments
	optimizedArtifact *artifact.Artifact
	optimizedTag      string

	availabilityErr      error
	availabilityArtifact *artifact.Artifact
}

func (s *stubOptimizerCtl) Optimize(_ context.Context, art *artifact.Artifact, tag string) (int64, error) {
	s.optimizedArtifact = art
	s.optimizedTag = tag
	return s.executionID, s.err
}

func (s *stubOptimizerCtl) CheckAvailability(_ context.Context, art *artifact.Artifact) error {
	s.availabilityArtifact = art
	return s.availabilityErr
}

// stubOptDAO is a simple in-test stub for dockerfileoptdao.DAO.
type stubOptDAO struct {
	upsertErr   error
	upsertedRec *dockerfileoptdao.DockerfileOptimization // captured by Upsert
	getResult   *dockerfileoptdao.DockerfileOptimization
	getErr      error
}

func (s *stubOptDAO) Upsert(_ context.Context, rec *dockerfileoptdao.DockerfileOptimization) error {
	s.upsertedRec = rec
	return s.upsertErr
}

func (s *stubOptDAO) GetByArtifact(_ context.Context, _, _ string) (*dockerfileoptdao.DockerfileOptimization, error) {
	return s.getResult, s.getErr
}

func (s *stubOptDAO) UpdateStatus(_ context.Context, _, _, _, _ string) error {
	return nil
}

// ctxWithSecurity returns a context with a mock security context that grants
// access (IsAuthenticated=true, Can=true).
func ctxWithSecurity() context.Context {
	sec := &securitytesting.Context{}
	sec.On("IsAuthenticated").Return(true)
	mock.OnAnything(sec, "Can").Return(true)

	// We also need baseProjectCtl to resolve project name → ID.
	// Use the package-level mock set up by TestMain.
	mock.OnAnything(projectCtlMock, "GetByName").Return(&project.Project{ProjectID: 1}, nil)

	return security.NewContext(context.Background(), sec)
}

// newTestDockerfileAPI creates a dockerfileAPI with stubbed dependencies.
func newTestDockerfileAPI(
	artCtl artifact.Controller,
	optimizerCtl optimizer.Controller,
	optDAO dockerfileoptdao.DAO,
) *dockerfileAPI {
	return &dockerfileAPI{
		artCtl:       artCtl,
		optimizerCtl: optimizerCtl,
		optDAO:       optDAO,
	}
}

func testArtifact() *artifact.Artifact {
	art := &artifact.Artifact{}
	art.RepositoryName = "library/myrepo"
	art.Digest = "sha256:abc123"
	return art
}

func TestGetDockerfileOptimizeAvailability_Available(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	art := testArtifact()
	mock.OnAnything(artCtl, "GetByReference").Return(art, nil).Once()

	optimizerCtl := &stubOptimizerCtl{}
	api := newTestDockerfileAPI(artCtl, optimizerCtl, &stubOptDAO{})
	params := operation.GetDockerfileOptimizeAvailabilityParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.GetDockerfileOptimizeAvailability(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 200, rw.Code)

	var body struct {
		Available bool   `json:"available"`
		Reason    string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	require.True(t, body.Available)
	require.Empty(t, body.Reason)
	require.Same(t, art, optimizerCtl.availabilityArtifact)
}

func TestGetDockerfileOptimizeAvailability_Unavailable(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	mock.OnAnything(artCtl, "GetByReference").Return(testArtifact(), nil).Once()

	optimizerCtl := &stubOptimizerCtl{
		availabilityErr: liberrors.PreconditionFailedError(nil).WithMessage("no optimizer adapter is configured"),
	}
	api := newTestDockerfileAPI(artCtl, optimizerCtl, &stubOptDAO{})
	params := operation.GetDockerfileOptimizeAvailabilityParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.GetDockerfileOptimizeAvailability(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	// The availability check itself always succeeds (200); unavailability is
	// reported in the payload, not as an HTTP error.
	require.Equal(t, 200, rw.Code)

	var body struct {
		Available bool   `json:"available"`
		Reason    string `json:"reason"`
	}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	require.False(t, body.Available)
	require.Equal(t, "no optimizer adapter is configured", body.Reason)
}

func TestGetDockerfileOptimization_HappyPath(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	mock.OnAnything(artCtl, "GetByReference").Return(testArtifact(), nil).Once()

	ts := time.Now().Truncate(time.Second)
	optDAO := &stubOptDAO{
		getResult: &dockerfileoptdao.DockerfileOptimization{
			Dockerfile:          "FROM alpine",
			OptimizedDockerfile: "FROM alpine:3.21",
			Status:              dockerfileoptdao.StatusSuccess,
			CreatedAt:           ts,
		},
	}

	api := newTestDockerfileAPI(artCtl, &stubOptimizerCtl{}, optDAO)
	params := operation.GetDockerfileOptimizationParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.GetDockerfileOptimization(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 200, rw.Code)

	var body struct {
		Dockerfile          string `json:"dockerfile"`
		OptimizedDockerfile string `json:"optimized_dockerfile"`
		Status              string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	require.Equal(t, "FROM alpine", body.Dockerfile)
	require.Equal(t, "FROM alpine:3.21", body.OptimizedDockerfile)
	require.Equal(t, dockerfileoptdao.StatusSuccess, body.Status)
}

func TestGetDockerfileOptimization_NotFound(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	mock.OnAnything(artCtl, "GetByReference").Return(testArtifact(), nil).Once()

	optDAO := &stubOptDAO{
		getErr: liberrors.New(nil).WithCode(liberrors.NotFoundCode).WithMessage("not found"),
	}

	api := newTestDockerfileAPI(artCtl, &stubOptimizerCtl{}, optDAO)
	params := operation.GetDockerfileOptimizationParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.GetDockerfileOptimization(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 404, rw.Code)
}

func TestGetDockerfileOptimization_ErrorState(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	mock.OnAnything(artCtl, "GetByReference").Return(testArtifact(), nil).Once()

	optDAO := &stubOptDAO{
		getResult: &dockerfileoptdao.DockerfileOptimization{
			Status: dockerfileoptdao.StatusError,
			Error:  "NO_ATTESTATION: attestation manifest not found",
		},
	}

	api := newTestDockerfileAPI(artCtl, &stubOptimizerCtl{}, optDAO)
	params := operation.GetDockerfileOptimizationParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.GetDockerfileOptimization(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 200, rw.Code)

	var body struct {
		Status string `json:"status"`
		Error  string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	require.Equal(t, dockerfileoptdao.StatusError, body.Status)
	require.Contains(t, body.Error, "NO_ATTESTATION")
}

func TestOptimizeDockerfile_ArtifactNotFound(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	mock.OnAnything(artCtl, "GetByReference").
		Return(nil, liberrors.NotFoundError(nil).WithMessage("artifact not found")).Once()

	api := newTestDockerfileAPI(artCtl, &stubOptimizerCtl{}, &stubOptDAO{})
	params := operation.OptimizeDockerfileParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.OptimizeDockerfile(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 404, rw.Code)
}

func TestOptimizeDockerfile_NoOptimizerConfigured(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	mock.OnAnything(artCtl, "GetByReference").Return(testArtifact(), nil).Once()

	optimizerCtl := &stubOptimizerCtl{
		err: liberrors.PreconditionFailedError(nil).WithMessage("no optimizer adapter is configured"),
	}

	api := newTestDockerfileAPI(artCtl, optimizerCtl, &stubOptDAO{})
	params := operation.OptimizeDockerfileParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.OptimizeDockerfile(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 412, rw.Code)
}

func TestOptimizeDockerfile_Accepted(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	mock.OnAnything(artCtl, "GetByReference").Return(testArtifact(), nil).Once()

	optimizerCtl := &stubOptimizerCtl{executionID: 7}
	optDAO := &stubOptDAO{
		getResult: &dockerfileoptdao.DockerfileOptimization{
			Status:      dockerfileoptdao.StatusPending,
			ExecutionID: 7,
		},
	}

	api := newTestDockerfileAPI(artCtl, optimizerCtl, optDAO)
	params := operation.OptimizeDockerfileParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.OptimizeDockerfile(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 202, rw.Code)

	var body struct {
		Status string `json:"status"`
	}
	require.NoError(t, json.Unmarshal(rw.Body.Bytes(), &body))
	require.Equal(t, dockerfileoptdao.StatusPending, body.Status)

	// The tag reference is forwarded to the optimize controller.
	require.NotNil(t, optimizerCtl.optimizedArtifact)
	require.Equal(t, "library/myrepo", optimizerCtl.optimizedArtifact.RepositoryName)
	require.Equal(t, "latest", optimizerCtl.optimizedTag)
}

func TestOptimizeDockerfile_DigestReferenceHasNoTag(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	art := testArtifact()
	mock.OnAnything(artCtl, "GetByReference").Return(art, nil).Once()

	optimizerCtl := &stubOptimizerCtl{executionID: 8}
	optDAO := &stubOptDAO{
		getResult: &dockerfileoptdao.DockerfileOptimization{Status: dockerfileoptdao.StatusPending},
	}

	api := newTestDockerfileAPI(artCtl, optimizerCtl, optDAO)
	params := operation.OptimizeDockerfileParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      art.Digest, // referenced by digest, not tag
	}

	responder := api.OptimizeDockerfile(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 202, rw.Code)
	require.Equal(t, "", optimizerCtl.optimizedTag)
}

func TestOptimizeDockerfile_RecordReadBackError(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	mock.OnAnything(artCtl, "GetByReference").Return(testArtifact(), nil).Once()

	optimizerCtl := &stubOptimizerCtl{executionID: 9}
	optDAO := &stubOptDAO{
		getErr: liberrors.New(nil).WithMessage("db read failed"),
	}

	api := newTestDockerfileAPI(artCtl, optimizerCtl, optDAO)
	params := operation.OptimizeDockerfileParams{
		HTTPRequest:    &http.Request{},
		ProjectName:    "library",
		RepositoryName: "myrepo",
		Reference:      "latest",
	}

	responder := api.OptimizeDockerfile(ctxWithSecurity(), params)

	rw := httptest.NewRecorder()
	responder.WriteResponse(rw, runtime.JSONProducer())
	require.Equal(t, 500, rw.Code)
}
