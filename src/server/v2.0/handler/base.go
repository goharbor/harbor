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
	"net/http"
	"net/url"
	"strconv"

	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/rbac/system"

	"github.com/go-openapi/runtime"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/q"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/log"
)

var (
	baseProjectCtl = project.Ctl
)

// BaseAPI base API handler
type BaseAPI struct{}

// Prepare default prepare for operation
func (*BaseAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	return nil
}

// SendError returns response for the err
func (*BaseAPI) SendError(ctx context.Context, err error) middleware.Responder {
	return NewErrResponder(err)
}

// GetSecurityContext from the provided context
func (*BaseAPI) GetSecurityContext(ctx context.Context) (security.Context, error) {
	sc, ok := security.FromContext(ctx)
	if !ok {
		return nil, errors.UnauthorizedError(errors.New("security context not found"))
	}
	return sc, nil
}

// HasPermission returns true when the request has action permission on resource
func (b *BaseAPI) HasPermission(ctx context.Context, action rbac.Action, resource rbac.Resource) bool {
	s, err := b.GetSecurityContext(ctx)
	if err != nil {
		log.Warningf("security context not found")
		return false
	}
	return s.Can(ctx, action, resource)
}

// HasProjectPermission returns true when the request has action permission on project subresource
func (b *BaseAPI) HasProjectPermission(ctx context.Context, projectIDOrName interface{}, action rbac.Action, subresource ...rbac.Resource) bool {
	projectID, projectName, err := utils.ParseProjectIDOrName(projectIDOrName)
	if err != nil {
		return false
	}

	if projectName != "" {
		p, err := baseProjectCtl.GetByName(ctx, projectName)
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

	resource := rbac_project.NewNamespace(projectID).Resource(subresource...)
	return b.HasPermission(ctx, action, resource)
}

// RequireProjectAccess checks the permission against the resources according to the context
// An error will be returned if it doesn't meet the requirement
func (b *BaseAPI) RequireProjectAccess(ctx context.Context, projectIDOrName interface{}, action rbac.Action, subresource ...rbac.Resource) error {
	if b.HasProjectPermission(ctx, projectIDOrName, action, subresource...) {
		return nil
	}
	secCtx, err := b.GetSecurityContext(ctx)
	if err != nil {
		return err
	}
	if !secCtx.IsAuthenticated() {
		return errors.UnauthorizedError(nil)
	}
	return errors.ForbiddenError(nil)
}

// RequireSystemAccess checks the system admin permission according to the security context
func (b *BaseAPI) RequireSystemAccess(ctx context.Context, action rbac.Action, subresource ...rbac.Resource) error {
	secCtx, err := b.GetSecurityContext(ctx)
	if err != nil {
		return err
	}
	if !secCtx.IsAuthenticated() {
		return errors.UnauthorizedError(nil)
	}
	resource := system.NewNamespace().Resource(subresource...)
	if !secCtx.Can(ctx, action, resource) {
		return errors.ForbiddenError(nil).WithMessage(secCtx.GetUsername())
	}
	return nil
}

// RequireAuthenticated checks it's authenticated according to the security context
func (b *BaseAPI) RequireAuthenticated(ctx context.Context) error {
	secCtx, err := b.GetSecurityContext(ctx)
	if err != nil {
		return err
	}
	if !secCtx.IsAuthenticated() {
		return errors.UnauthorizedError(nil)
	}
	return nil
}

// RequireSolutionUserAccess check if current user is internal service
func (b *BaseAPI) RequireSolutionUserAccess(ctx context.Context) error {
	sec, exist := security.FromContext(ctx)
	if !exist || !sec.IsSolutionUser() {
		return errors.UnauthorizedError(nil).WithMessage("only internal service is allowed to call this API")
	}
	return nil
}

// BuildQuery builds the query model according to the query string
func (b *BaseAPI) BuildQuery(ctx context.Context, query, sort *string, pageNumber, pageSize *int64) (*q.Query, error) {
	var (
		qs string
		st string
		pn int64
		ps int64
	)
	if query != nil {
		qs = *query
	}
	if sort != nil {
		st = *sort
	}
	if pageNumber != nil {
		pn = *pageNumber
	}
	if pageSize != nil {
		ps = *pageSize
	}
	return q.Build(qs, st, pn, ps)
}

// Links return Links based on the provided pagination information
func (b *BaseAPI) Links(ctx context.Context, u *url.URL, total, pageNumber, pageSize int64) lib.Links {
	var links lib.Links
	if pageSize == 0 {
		return links
	}
	ul := *u
	// prev
	if pageNumber > 1 && (pageNumber-1)*pageSize < total {
		q := ul.Query()
		q.Set("page", strconv.FormatInt(pageNumber-1, 10))
		// the URL may contain no "page_size", in this case the pageSize in the query is set by
		// the go-swagger automatically
		q.Set("page_size", strconv.FormatInt(pageSize, 10))
		ul.RawQuery = q.Encode()
		// try to unescape the query
		if escapedQuery, err := url.QueryUnescape(ul.RawQuery); err == nil {
			ul.RawQuery = escapedQuery
		} else {
			log.Errorf("failed to unescape the query %s: %v", ul.RawQuery, err)
		}
		link := &lib.Link{
			URL: ul.String(),
			Rel: "prev",
		}
		links = append(links, link)
	}
	// next
	if pageSize*pageNumber < total {
		q := ul.Query()
		q.Set("page", strconv.FormatInt(pageNumber+1, 10))
		q.Set("page_size", strconv.FormatInt(pageSize, 10))
		ul.RawQuery = q.Encode()
		// try to unescape the query
		if escapedQuery, err := url.QueryUnescape(ul.RawQuery); err == nil {
			ul.RawQuery = escapedQuery
		} else {
			log.Errorf("failed to unescape the query %s: %v", ul.RawQuery, err)
		}
		link := &lib.Link{
			URL: ul.String(),
			Rel: "next",
		}
		links = append(links, link)
	}
	return links
}

var _ middleware.Responder = &ErrResponder{}

// ErrResponder error responder
type ErrResponder struct {
	err error
}

// WriteResponse ...
func (r *ErrResponder) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	lib_http.SendError(rw, r.err)
}

// NewErrResponder returns responder for err
func NewErrResponder(err error) *ErrResponder {
	return &ErrResponder{err: err}
}
