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

// TODO move this file out of v2.0 folder as this is common for all versions of API

import (
	"context"
	"errors"
	"github.com/goharbor/harbor/src/internal"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"net/url"
	"strconv"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/project"
	errs "github.com/goharbor/harbor/src/server/error"
)

// BaseAPI base API handler
type BaseAPI struct{}

// Prepare default prepare for operation
func (*BaseAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

// SendError returns response for the err
func (*BaseAPI) SendError(ctx context.Context, err error) middleware.Responder {
	return errs.NewErrResponder(err)
}

// HasPermission returns true when the request has action permission on resource
func (*BaseAPI) HasPermission(ctx context.Context, action rbac.Action, resource rbac.Resource) bool {
	s, ok := security.FromContext(ctx)
	if !ok {
		log.Warningf("security not found in the context")
		return false
	}

	return s.Can(action, resource)
}

// HasProjectPermission returns true when the request has action permission on project subresource
func (b *BaseAPI) HasProjectPermission(ctx context.Context, projectIDOrName interface{}, action rbac.Action, subresource ...rbac.Resource) bool {
	projectID, projectName, err := utils.ParseProjectIDOrName(projectIDOrName)
	if err != nil {
		return false
	}

	if projectName != "" {
		// TODO: use the project controller to replace the project manager
		p, err := project.Mgr.Get(projectName)
		if err != nil {
			log.Errorf("failed to get project %s: %v", projectName, err)
			return false
		}
		if p == nil {
			log.Warningf("project %s not found", projectName)
			return false
		}

		projectID = p.ProjectID
	}

	resource := rbac.NewProjectNamespace(projectID).Resource(subresource...)
	return b.HasPermission(ctx, action, resource)
}

// RequireProjectAccess checks the permission against the resources according to the context
// An error will be returned if it doesn't meet the requirement
func (b *BaseAPI) RequireProjectAccess(ctx context.Context, projectIDOrName interface{}, action rbac.Action, subresource ...rbac.Resource) error {
	if b.HasProjectPermission(ctx, projectIDOrName, action, subresource...) {
		return nil
	}
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return ierror.UnauthorizedError(errors.New("security context not found"))
	}
	if !secCtx.IsAuthenticated() {
		return ierror.UnauthorizedError(nil)
	}
	return ierror.ForbiddenError(nil)
}

// Links return Links based on the provided pagination information
func (b *BaseAPI) Links(ctx context.Context, u *url.URL, total, pageNumber, pageSize int64) internal.Links {
	url := *u
	var links internal.Links
	// prev
	if pageNumber > 1 && (pageNumber-1)*pageSize < total {
		q := url.Query()
		q.Set("page", strconv.FormatInt(pageNumber-1, 10))
		url.RawQuery = q.Encode()
		link := &internal.Link{
			URL: url.String(),
			Rel: "prev",
		}
		links = append(links, link)
	}
	// next
	if pageSize*pageNumber < total {
		q := url.Query()
		q.Set("page", strconv.FormatInt(pageNumber+1, 10))
		url.RawQuery = q.Encode()
		link := &internal.Link{
			URL: url.String(),
			Rel: "next",
		}
		links = append(links, link)
	}
	return links
}
