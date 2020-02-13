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
	"github.com/docker/distribution/reference"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/repository"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/artifact"
	"github.com/opencontainers/go-digest"
	"net/http"
	"strings"
	"time"
)

func newArtifactAPI() *artifactAPI {
	return &artifactAPI{
		artCtl:  artifact.Ctl,
		proMgr:  project.Mgr,
		repoCtl: repository.Ctl,
	}
}

type artifactAPI struct {
	BaseAPI
	artCtl  artifact.Controller
	proMgr  project.Manager
	repoCtl repository.Controller
}

func (a *artifactAPI) ListArtifacts(ctx context.Context, params operation.ListArtifactsParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	// set query
	query := &q.Query{
		Keywords: map[string]interface{}{},
	}
	if params.Type != nil {
		query.Keywords["Type"] = *(params.Type)
	}
	if params.Tags != nil {
		query.Keywords["Tags"] = *(params.Tags)
	}
	if params.Page != nil {
		query.PageNumber = *(params.Page)
	}
	if params.PageSize != nil {
		query.PageSize = *(params.PageSize)
	}
	repository, err := a.repoCtl.GetByName(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName))
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["RepositoryID"] = repository.RepositoryID

	// set option
	option := option(params.WithTag, params.WithImmutableStatus,
		params.WithLabel, params.WithSignature)

	// get the total count of artifacts
	total, err := a.artCtl.Count(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}
	// list artifacts according to the query and option
	arts, err := a.artCtl.List(ctx, query, option)
	if err != nil {
		return a.SendError(ctx, err)
	}

	var artifacts []*model.Artifact
	for _, art := range arts {
		artifact := &model.Artifact{}
		artifact.Artifact = *art
		a.assembleArtifact(ctx, artifact, params.WithScanOverview)
		artifacts = append(artifacts, artifact)
	}

	// TODO add link header
	return operation.NewListArtifactsOK().WithXTotalCount(total).WithLink("").WithPayload(artifacts)
}

func (a *artifactAPI) GetArtifact(ctx context.Context, params operation.GetArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	// set option
	option := option(params.WithTag, params.WithImmutableStatus,
		params.WithLabel, params.WithSignature)

	// get the artifact
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, option)
	if err != nil {
		return a.SendError(ctx, err)
	}
	art := &model.Artifact{}
	art.Artifact = *artifact
	a.assembleArtifact(ctx, art, params.WithScanOverview)
	return operation.NewGetArtifactOK().WithPayload(art)
}

func (a *artifactAPI) DeleteArtifact(ctx context.Context, params operation.DeleteArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.Delete(ctx, artifact.ID); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewDeleteArtifactOK()
}

// TODO immutable, quota, readonly middlewares should cover this API
func (a *artifactAPI) CopyArtifact(ctx context.Context, params operation.CopyArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	srcRepo, srcRef, err := parse(params.From)
	if err != nil {
		return a.SendError(ctx, err)
	}
	srcPro, _ := utils.ParseRepository(srcRepo)
	if err = a.RequireProjectAccess(ctx, srcPro, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	srcArt, err := a.artCtl.GetByReference(ctx, srcRepo, srcRef, &artifact.Option{WithTag: true})
	if err != nil {
		return a.SendError(ctx, err)
	}
	_, id, err := a.repoCtl.Ensure(ctx, params.ProjectName+"/"+params.RepositoryName)
	if err != nil {
		return a.SendError(ctx, err)
	}
	id, err = a.artCtl.Copy(ctx, srcArt.ID, id)
	if err != nil {
		return a.SendError(ctx, err)
	}
	// TODO set location header
	_ = id
	return operation.NewCopyArtifactCreated()
}

// parse "repository:tag" or "repository@digest" into repository and reference parts
func parse(s string) (string, string, error) {
	matches := reference.ReferenceRegexp.FindStringSubmatch(s)
	if matches == nil {
		return "", "", ierror.New(nil).WithCode(ierror.BadRequestCode).
			WithMessage("invalid input: %s", s)
	}
	repository := matches[1]
	reference := matches[2]
	if matches[3] != "" {
		_, err := digest.Parse(matches[3])
		if err != nil {
			return "", "", ierror.New(nil).WithCode(ierror.BadRequestCode).
				WithMessage("invalid input: %s", s)
		}
		reference = matches[3]
	}
	return repository, reference, nil
}

func (a *artifactAPI) CreateTag(ctx context.Context, params operation.CreateTagParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceTag); err != nil {
		return a.SendError(ctx, err)
	}
	art, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName),
		params.Reference, &artifact.Option{
			WithTag: true,
		})
	if err != nil {
		return a.SendError(ctx, err)
	}
	tag := &artifact.Tag{}
	tag.RepositoryID = art.RepositoryID
	tag.ArtifactID = art.ID
	tag.Name = params.Tag.Name
	tag.PushTime = time.Now()
	if _, err = a.artCtl.CreateTag(ctx, tag); err != nil {
		return a.SendError(ctx, err)
	}
	// TODO set location header?
	return operation.NewCreateTagCreated()
}

func (a *artifactAPI) DeleteTag(ctx context.Context, params operation.DeleteTagParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceTag); err != nil {
		return a.SendError(ctx, err)
	}
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName),
		params.Reference, &artifact.Option{
			WithTag: true,
		})
	if err != nil {
		return a.SendError(ctx, err)
	}
	var id int64
	for _, tag := range artifact.Tags {
		if tag.Name == params.TagName {
			id = tag.ID
			break
		}
	}
	// the tag not found
	if id == 0 {
		err = ierror.New(nil).WithCode(ierror.NotFoundCode).WithMessage(
			"tag %s attached to artifact %d not found", params.TagName, artifact.ID)
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.DeleteTag(ctx, id); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewDeleteTagOK()
}

func (a *artifactAPI) GetAddition(ctx context.Context, params operation.GetAdditionParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifactAddition); err != nil {
		return a.SendError(ctx, err)
	}
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	addition, err := a.artCtl.GetAddition(ctx, artifact.ID, strings.ToUpper(params.Addition))
	if err != nil {
		return a.SendError(ctx, err)
	}
	return middleware.ResponderFunc(func(w http.ResponseWriter, p runtime.Producer) {
		w.Header().Set("Content-Type", addition.ContentType)
		w.Write(addition.Content)
	})
}

func (a *artifactAPI) AddLabel(ctx context.Context, params operation.AddLabelParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceArtifactLabel); err != nil {
		return a.SendError(ctx, err)
	}
	art, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.AddLabel(ctx, art.ID, params.Label.ID); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewAddLabelOK()
}

func (a *artifactAPI) RemoveLabel(ctx context.Context, params operation.RemoveLabelParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceArtifactLabel); err != nil {
		return a.SendError(ctx, err)
	}
	art, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.RemoveLabel(ctx, art.ID, params.LabelID); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewRemoveLabelOK()
}

func (a *artifactAPI) assembleArtifact(ctx context.Context, artifact *model.Artifact, withScanOverview *bool) {
	if withScanOverview != nil && *withScanOverview {
		// TODO populate scan result
	}
	// TODO populate vulnerability link
}

func option(withTag, withImmutableStatus, withLabel, withSignature *bool) *artifact.Option {
	option := &artifact.Option{
		WithTag: true, // return the tag by default
	}
	if withTag != nil {
		option.WithTag = *(withTag)
	}
	if option.WithTag {
		option.TagOption = &artifact.TagOption{}
		if withImmutableStatus != nil {
			option.TagOption.WithImmutableStatus = *(withImmutableStatus)
		}
		if withSignature != nil {
			option.TagOption.WithSignature = *withSignature
		}
	}
	if withLabel != nil {
		option.WithLabel = *(withLabel)
	}
	return option
}
