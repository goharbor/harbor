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
	"net/http"
	"strings"
	"time"

	"github.com/docker/distribution/reference"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/api/repository"
	"github.com/goharbor/harbor/src/api/scan"
	"github.com/goharbor/harbor/src/api/tag"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	ierror "github.com/goharbor/harbor/src/internal/error"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/server/v2.0/handler/assembler"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/artifact"
	"github.com/opencontainers/go-digest"
)

const (
	vulnerabilitiesAddition = "vulnerabilities"
)

func newArtifactAPI() *artifactAPI {
	return &artifactAPI{
		artCtl:  artifact.Ctl,
		repoCtl: repository.Ctl,
		scanCtl: scan.DefaultController,
		tagCtl:  tag.Ctl,
	}
}

type artifactAPI struct {
	BaseAPI
	artCtl  artifact.Controller
	repoCtl repository.Controller
	scanCtl scan.Controller
	tagCtl  tag.Controller
}

func (a *artifactAPI) ListArtifacts(ctx context.Context, params operation.ListArtifactsParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}

	// set query
	query, err := a.BuildQuery(ctx, params.Q)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["RepositoryName"] = fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)

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
		artifacts = append(artifacts, artifact)
	}

	assembler.NewVulAssembler(boolValue(params.WithScanOverview)).WithArtifacts(artifacts...).Assemble(ctx)

	return operation.NewListArtifactsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(artifacts)
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

	assembler.NewVulAssembler(boolValue(params.WithScanOverview)).WithArtifacts(art).Assemble(ctx)

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

	srcRepo, ref, err := parse(params.From)
	if err != nil {
		return a.SendError(ctx, err)
	}

	srcPro, _ := utils.ParseRepository(srcRepo)
	if err = a.RequireProjectAccess(ctx, srcPro, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}

	dstRepo := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	_, _, err = a.repoCtl.Ensure(ctx, dstRepo)
	if err != nil {
		return a.SendError(ctx, err)
	}

	_, err = a.artCtl.Copy(ctx, srcRepo, ref, dstRepo)
	if err != nil {
		return a.SendError(ctx, err)
	}
	location := strings.TrimSuffix(params.HTTPRequest.URL.Path, "/") + "/" + ref
	return operation.NewCopyArtifactCreated().WithLocation(location)
}

func (a *artifactAPI) ScanArtifact(ctx context.Context, params operation.ScanArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceScan); err != nil {
		return a.SendError(ctx, err)
	}

	repository := fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName)
	artifact, err := a.artCtl.GetByReference(ctx, repository, params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}

	art := &v1.Artifact{
		NamespaceID: artifact.ProjectID,
		Repository:  repository,
		Digest:      artifact.Digest,
		MimeType:    artifact.ManifestMediaType,
	}
	if err := a.scanCtl.Scan(art); err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewScanArtifactAccepted()
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
	tag := &tag.Tag{}
	tag.RepositoryID = art.RepositoryID
	tag.ArtifactID = art.ID
	tag.Name = params.Tag.Name
	tag.PushTime = time.Now()
	if _, err = a.tagCtl.Create(ctx, tag); err != nil {
		return a.SendError(ctx, err)
	}
	// TODO as we provide no API for get the single tag, ignore setting the location header here
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
	if err = a.tagCtl.Delete(ctx, id); err != nil {
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

	var addition *resolver.Addition

	if params.Addition == vulnerabilitiesAddition {
		addition, err = resolveVulnerabilitiesAddition(ctx, artifact)
	} else {
		addition, err = a.artCtl.GetAddition(ctx, artifact.ID, strings.ToUpper(params.Addition))
	}
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

func option(withTag, withImmutableStatus, withLabel, withSignature *bool) *artifact.Option {
	option := &artifact.Option{
		WithTag:   true, // return the tag by default
		WithLabel: boolValue(withLabel),
	}

	if withTag != nil {
		option.WithTag = *(withTag)
	}

	if option.WithTag {
		option.TagOption = &tag.Option{
			WithImmutableStatus: boolValue(withImmutableStatus),
			WithSignature:       boolValue(withSignature),
		}
	}

	return option
}
