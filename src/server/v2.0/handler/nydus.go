package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/nydus"
	"github.com/goharbor/harbor/src/pkg/task"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/nydus"
)

func newNydusAPI() *nydusAPI {
	return &nydusAPI{
		artCtl: artifact.Ctl,
		nyCtl:  nydus.DefaultController,
	}
}

type nydusAPI struct {
	BaseAPI
	artCtl artifact.Controller
	nyCtl  nydus.Controller
}

func (s *nydusAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	if err := unescapePathParams(params, "RepositoryName"); err != nil {
		s.SendError(ctx, err)
	}

	return nil
}

func (s *nydusAPI) ConvertArtifact(ctx context.Context, params operation.ConvertArtifactParams) middleware.Responder {
	if err := s.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceConvert); err != nil {
		return s.SendError(ctx, err)
	}

	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	artifact, err := s.artCtl.GetByReference(ctx, repository, params.Reference, &artifact.Option{WithTag: true})
	if err != nil {
		return s.SendError(ctx, err)
	}

	if err := s.nyCtl.Convert(ctx, artifact, task.ExecutionTriggerManual); err != nil {
		return s.SendError(ctx, err)
	}

	return operation.NewConvertArtifactAccepted()
}
