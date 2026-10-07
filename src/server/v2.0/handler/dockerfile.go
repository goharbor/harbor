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
	"fmt"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/optimizer"
	dockerfileoptdao "github.com/goharbor/harbor/src/pkg/dockerfileoptimization/dao"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/dockerfile"
)

func newDockerfileAPI() *dockerfileAPI {
	return &dockerfileAPI{
		artCtl:       artifact.Ctl,
		optimizerCtl: optimizer.DefaultController,
		optDAO:       dockerfileoptdao.New(),
	}
}

type dockerfileAPI struct {
	BaseAPI
	artCtl       artifact.Controller
	optimizerCtl optimizer.Controller
	optDAO       dockerfileoptdao.DAO
}

func (d *dockerfileAPI) Prepare(ctx context.Context, _ string, params any) middleware.Responder {
	if err := unescapePathParams(params, "RepositoryName"); err != nil {
		d.SendError(ctx, err)
	}
	return nil
}

// OptimizeDockerfile starts an asynchronous optimization through the registered
// optimizer adapter and returns 202 with the pending record. The extraction and LLM
// work happens in the adapter service; poll the GET endpoint for the result.
func (d *dockerfileAPI) OptimizeDockerfile(ctx context.Context, params operation.OptimizeDockerfileParams) middleware.Responder {
	if err := d.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return d.SendError(ctx, err)
	}

	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	art, err := d.artCtl.GetByReference(ctx, repository, params.Reference, nil)
	if err != nil {
		return d.SendError(ctx, err)
	}

	var tag string
	if art.Digest != params.Reference {
		tag = params.Reference
	}

	if _, err := d.optimizerCtl.Optimize(ctx, art, tag); err != nil {
		return d.SendError(ctx, err)
	}

	rec, err := d.optDAO.GetByArtifact(ctx, art.RepositoryName, art.Digest)
	if err != nil {
		return d.SendError(ctx, err)
	}

	return operation.NewOptimizeDockerfileAccepted().WithPayload(toDockerfileOptimizationModel(rec))
}

// GetDockerfileOptimizeAvailability reports whether the default optimizer adapter is
// currently configured, enabled, reachable, and capable of optimizing this artifact —
// the same checks OptimizeDockerfile performs before starting a job, without starting one.
func (d *dockerfileAPI) GetDockerfileOptimizeAvailability(ctx context.Context, params operation.GetDockerfileOptimizeAvailabilityParams) middleware.Responder {
	if err := d.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return d.SendError(ctx, err)
	}

	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	art, err := d.artCtl.GetByReference(ctx, repository, params.Reference, nil)
	if err != nil {
		return d.SendError(ctx, err)
	}

	availability := &models.OptimizerAvailability{Available: true}
	if err := d.optimizerCtl.CheckAvailability(ctx, art); err != nil {
		availability.Available = false
		availability.Reason = err.Error()
	}

	return operation.NewGetDockerfileOptimizeAvailabilityOK().WithPayload(availability)
}

func (d *dockerfileAPI) GetDockerfileOptimization(ctx context.Context, params operation.GetDockerfileOptimizationParams) middleware.Responder {
	if err := d.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return d.SendError(ctx, err)
	}

	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	art, err := d.artCtl.GetByReference(ctx, repository, params.Reference, nil)
	if err != nil {
		return d.SendError(ctx, err)
	}

	rec, err := d.optDAO.GetByArtifact(ctx, art.RepositoryName, art.Digest)
	if err != nil {
		return d.SendError(ctx, err)
	}

	return operation.NewGetDockerfileOptimizationOK().WithPayload(toDockerfileOptimizationModel(rec))
}

func toDockerfileOptimizationModel(rec *dockerfileoptdao.DockerfileOptimization) *models.DockerfileOptimization {
	return &models.DockerfileOptimization{
		Dockerfile:                rec.Dockerfile,
		OptimizedDockerfile:       rec.OptimizedDockerfile,
		AttestationManifestDigest: rec.AttestationManifestDigest,
		StatementDigest:           rec.StatementDigest,
		Generated:                 rec.Generated,
		Status:                    rec.Status,
		Error:                     rec.Error,
		CreatedAt:                 strfmt.DateTime(rec.CreatedAt),
		UpdateTime:                strfmt.DateTime(rec.UpdateTime),
	}
}
