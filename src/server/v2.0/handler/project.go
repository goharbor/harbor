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
	"strconv"
	"strings"
	"sync"

	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	robotSec "github.com/goharbor/harbor/src/common/security/robot"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/p2p/preheat"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/retention"
	"github.com/goharbor/harbor/src/controller/scanner"
	"github.com/goharbor/harbor/src/controller/user"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/pkg/auditext"
	"github.com/goharbor/harbor/src/pkg/member"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	pkgModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/robot"
	userModels "github.com/goharbor/harbor/src/pkg/user/models"
	"github.com/goharbor/harbor/src/server/v2.0/handler/assembler"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project"
)

// for the proxy cache type project, we will create a 7 days retention policy for it by default
const defaultDaysToRetentionForProxyCacheProject = 7

func newProjectAPI() *projectAPI {
	return &projectAPI{
		auditMgr:      audit.Mgr,
		artCtl:        artifact.Ctl,
		auditextMgr:   auditext.Mgr,
		metadataMgr:   pkg.ProjectMetaMgr,
		userCtl:       user.Ctl,
		repositoryCtl: repository.Ctl,
		projectCtl:    project.Ctl,
		memberMgr:     member.Mgr,
		quotaCtl:      quota.Ctl,
		robotMgr:      robot.Mgr,
		preheatCtl:    preheat.Ctl,
		retentionCtl:  retention.Ctl,
		scannerCtl:    scanner.DefaultController,
	}
}

type projectAPI struct {
	BaseAPI
	auditMgr      audit.Manager
	auditextMgr   auditext.Manager
	artCtl        artifact.Controller
	metadataMgr   metadata.Manager
	userCtl       user.Controller
	repositoryCtl repository.Controller
	projectCtl    project.Controller
	memberMgr     member.Manager
	quotaCtl      quota.Controller
	robotMgr      robot.Manager
	preheatCtl    preheat.Controller
	retentionCtl  retention.Controller
	scannerCtl    scanner.Controller
}

func (a *projectAPI) CreateProject(ctx context.Context, params operation.CreateProjectParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	onlyAdmin, err := config.OnlyAdminCreateProject(ctx)
	if err != nil {
		return a.SendError(ctx, fmt.Errorf("failed to determine whether only admin can create projects: %v", err))
	}

	secCtx, _ := security.FromContext(ctx)
	if r, ok := secCtx.(*robotSec.SecurityContext); ok && !r.User().IsSysLevel() {
		log.Errorf("Only system level robot can create project")
		return a.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Only system level robot can create project"))
	}
	if onlyAdmin && !(a.isSysAdmin(ctx, rbac.ActionCreate) || secCtx.IsSolutionUser()) {
		log.Errorf("Only sys admin can create project")
		return a.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Only system admin can create project"))
	}

	req := params.Project

	if req.RegistryID != nil && !a.isSysAdmin(ctx, rbac.ActionCreate) {
		return a.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Only system admin can create proxy cache project"))
	}

	// populate storage limit
	if config.QuotaPerProjectEnable(ctx) {
		// the security context is not sys admin, set the StorageLimit the global StoragePerProject
		if req.StorageLimit == nil || *req.StorageLimit == 0 || !a.isSysAdmin(ctx, rbac.ActionCreate) {
			setting, err := config.QuotaSetting(ctx)
			if err != nil {
				log.Errorf("failed to get quota setting: %v", err)
				return a.SendError(ctx, fmt.Errorf("failed to get quota setting: %v", err))
			}
			defaultStorageLimit := setting.StoragePerProject
			req.StorageLimit = &defaultStorageLimit
		}
	} else {
		// ignore storage limit when quota per project disabled
		req.StorageLimit = nil
	}

	if req.Metadata == nil {
		req.Metadata = &models.ProjectMetadata{}
	}

	// accept the "public" property to make replication work well with old versions(<=1.2.0)
	if req.Public != nil && req.Metadata.Public == "" {
		req.Metadata.Public = strconv.FormatBool(*req.Public)
	}

	// populate public metadata as false if it isn't set
	if req.Metadata.Public == "" {
		req.Metadata.Public = strconv.FormatBool(false)
	}

	// validate metadata.public value, should only be "true" or "false"
	if p := req.Metadata.Public; p != "" {
		if p != "true" && p != "false" {
			return a.SendError(ctx, errors.BadRequestError(nil).WithMessagef("metadata.public should only be 'true' or 'false', but got: '%s'", p))
		}
	}

	// ignore metadata.proxy_speed_kb for non-proxy-cache project
	if req.RegistryID == nil {
		req.Metadata.ProxySpeedKb = nil
	}

	// ignore enable_content_trust metadata for proxy cache project
	// see https://github.com/goharbor/harbor/issues/12940 to get more info
	if req.RegistryID != nil {
		req.Metadata.EnableContentTrust = nil
	}

	// validate the RetentionID, RegistryID and StorageLimit in the body of the request
	if err := a.validateProjectReq(ctx, req); err != nil {
		return a.SendError(ctx, err)
	}

	var ownerID int
	// TODO: revise the ownerID in project model.
	// set the owner as the system admin when the API being called by replication
	// it's a solution to workaround the restriction of project creation API:
	// only normal users can create projects
	// Remove the assumption of user id 1 is the system admin. And use the minimum system admin id as the owner ID,
	// in most case, it's 1
	if _, ok := secCtx.(*robotSec.SecurityContext); ok || secCtx.IsSolutionUser() {
		q := &q.Query{
			Keywords: map[string]interface{}{
				"sysadmin_flag": true,
			},
			Sorts: []*q.Sort{
				q.NewSort("user_id", false),
			},
		}
		admins, err := a.userCtl.List(ctx, q, userModels.WithDefaultAdmin())
		if err != nil {
			return a.SendError(ctx, err)
		}
		if len(admins) == 0 {
			return a.SendError(ctx, errors.New(nil).WithMessage("cannot create project as no system admin found"))
		}
		ownerID = admins[0].UserID
	} else {
		ownerName := secCtx.GetUsername()
		user, err := a.userCtl.GetByName(ctx, ownerName)
		if err != nil {
			return a.SendError(ctx, err)
		}
		ownerID = user.UserID
	}

	p := &project.Project{
		Name:       req.ProjectName,
		OwnerID:    ownerID,
		RegistryID: lib.Int64Value(req.RegistryID),
	}
	if err := lib.JSONCopy(&p.Metadata, req.Metadata); err != nil {
		log.Warningf("failed to call JSONCopy on project metadata when CreateProject, error: %v", err)
	}
	delete(p.Metadata, "retention_id")

	projectID, err := a.projectCtl.Create(ctx, p)
	if err != nil {
		return a.SendError(ctx, err)
	}

	// StorageLimit is provided in the request body and it's valid,
	// create the quota for the project
	if req.StorageLimit != nil {
		referenceID := quota.ReferenceID(projectID)
		hardLimits := types.ResourceList{types.ResourceStorage: *req.StorageLimit}
		if _, err := a.quotaCtl.Create(ctx, quota.ProjectReference, referenceID, hardLimits); err != nil {
			return a.SendError(ctx, fmt.Errorf("failed to create quota for project: %v", err))
		}
	}

	// RegistryID is provided in the request body and it's valid,
	// create a default retention policy for proxy project
	if req.RegistryID != nil {
		plc := policy.WithNDaysSinceLastPull(projectID, defaultDaysToRetentionForProxyCacheProject)
		retentionID, err := a.retentionCtl.CreateRetention(ctx, plc)
		if err != nil {
			return a.SendError(ctx, err)
		}
		md := map[string]string{"retention_id": strconv.FormatInt(retentionID, 10)}
		if err := a.metadataMgr.Add(ctx, projectID, md); err != nil {
			return a.SendError(ctx, err)
		}
	}

	var location string
	if lib.BoolValue(params.XResourceNameInLocation) {
		location = fmt.Sprintf("%s/%s", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), req.ProjectName)
	} else {
		location = fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), projectID)
	}

	return operation.NewCreateProjectCreated().WithLocation(location)
}

func (a *projectAPI) DeleteProject(ctx context.Context, params operation.DeleteProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete); err != nil {
		return a.SendError(ctx, err)
	}

	p, result, err := a.deletable(ctx, projectNameOrID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if !result.Deletable {
		return a.SendError(ctx, errors.PreconditionFailedError(errors.New(result.Message)))
	}

	if err := a.projectCtl.Delete(ctx, p.ProjectID); err != nil {
		return a.SendError(ctx, err)
	}

	// remove the robot associated with the project
	if err := a.robotMgr.DeleteByProjectID(ctx, p.ProjectID); err != nil {
		return a.SendError(ctx, err)
	}

	referenceID := quota.ReferenceID(p.ProjectID)
	q, err := a.quotaCtl.GetByRef(ctx, quota.ProjectReference, referenceID)
	if err != nil {
		log.Warningf("failed to get quota for project %s, error: %v", projectNameOrID, err)
	} else {
		if err := a.quotaCtl.Delete(ctx, q.ID); err != nil {
			return a.SendError(ctx, fmt.Errorf("failed to delete quota for project: %v", err))
		}
	}

	// preheat policies under the project should be deleted after deleting the project
	if err = a.preheatCtl.DeletePoliciesOfProject(ctx, p.ProjectID); err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewDeleteProjectOK()
}

func (a *projectAPI) GetLogs(ctx context.Context, params operation.GetLogsParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceLog); err != nil {
		return a.SendError(ctx, err)
	}
	pro, err := a.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["ProjectID"] = pro.ProjectID

	total, err := a.auditMgr.Count(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}
	logs, err := a.auditMgr.List(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}

	var auditLogs []*models.AuditLog
	for _, log := range logs {
		auditLogs = append(auditLogs, &models.AuditLog{
			ID:           log.ID,
			Resource:     log.Resource,
			ResourceType: log.ResourceType,
			Username:     log.Username,
			Operation:    log.Operation,
			OpTime:       strfmt.DateTime(log.OpTime),
		})
	}
	return operation.NewGetLogsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(auditLogs)
}

func (a *projectAPI) GetProject(ctx context.Context, params operation.GetProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead); err != nil {
		return a.SendError(ctx, err)
	}

	p, err := a.getProject(ctx, projectNameOrID, project.WithCVEAllowlist(), project.WithOwner())
	if err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewGetProjectOK().WithPayload(model.NewProject(p).ToSwagger())
}

func (a *projectAPI) GetProjectDeletable(ctx context.Context, params operation.GetProjectDeletableParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete); err != nil {
		return a.SendError(ctx, err)
	}

	_, result, err := a.deletable(ctx, projectNameOrID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewGetProjectDeletableOK().WithPayload(result)
}

func (a *projectAPI) GetProjectSummary(ctx context.Context, params operation.GetProjectSummaryParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead); err != nil {
		return a.SendError(ctx, err)
	}

	p, err := a.getProject(ctx, projectNameOrID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	summary := &models.ProjectSummary{
		RepoCount: p.RepoCount,
	}

	var fetchSummaries []func(context.Context, *project.Project, *models.ProjectSummary)

	if hasPerm := a.HasProjectPermission(ctx, p.ProjectID, rbac.ActionRead, rbac.ResourceQuota); hasPerm {
		fetchSummaries = append(fetchSummaries, getProjectQuotaSummary)
	}

	if hasPerm := a.HasProjectPermission(ctx, p.ProjectID, rbac.ActionList, rbac.ResourceMember); hasPerm {
		fetchSummaries = append(fetchSummaries, a.getProjectMemberSummary)
	}

	if p.IsProxy() {
		fetchSummaries = append(fetchSummaries, getProjectRegistrySummary)
	}

	var wg sync.WaitGroup
	for _, fn := range fetchSummaries {
		fn := fn

		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(ctx, p, summary)
		}()
	}
	wg.Wait()

	return operation.NewGetProjectSummaryOK().WithPayload(summary)
}

func (a *projectAPI) HeadProject(ctx context.Context, params operation.HeadProjectParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	if _, err := a.projectCtl.GetByName(ctx, params.ProjectName); err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewHeadProjectOK()
}

func (a *projectAPI) ListProjects(ctx context.Context, params operation.ListProjectsParams) middleware.Responder {
	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if name := lib.StringValue(params.Name); name != "" {
		query.Keywords["name"] = &q.FuzzyMatchValue{Value: name}
	}
	if owner := lib.StringValue(params.Owner); owner != "" {
		query.Keywords["owner"] = owner
	}
	if params.Public != nil {
		query.Keywords["public"] = lib.BoolValue(params.Public)
	}

	secCtx, ok := security.FromContext(ctx)
	if ok && secCtx.IsAuthenticated() {
		if !a.isSysAdmin(ctx, rbac.ActionList) && !secCtx.IsSolutionUser() {
			// authenticated but not system admin or solution user,
			// return public projects and projects that the user is member of
			if l, ok := secCtx.(*local.SecurityContext); ok {
				currentUser := l.User()
				member := &project.MemberQuery{
					UserID:   currentUser.UserID,
					GroupIDs: currentUser.GroupIDs,
				}

				// not filter by public or filter by the public with true,
				// so also return public projects for the member
				if public, ok := query.Keywords["public"]; !ok || lib.ToBool(public) {
					member.WithPublic = true
				}

				query.Keywords["member"] = member
			} else if r, ok := secCtx.(*robotSec.SecurityContext); ok {
				// for the system level robot that covers all the project, see it as the system admin.
				var coverAll bool
				var names []string
				for _, p := range r.User().Permissions {
					if p.IsCoverAll() {
						coverAll = true
						break
					}
					names = append(names, p.Namespace)
				}
				if !coverAll {
					namesQuery := &pkgModels.NamesQuery{
						Names: names,
					}
					if public, ok := query.Keywords["public"]; !ok || lib.ToBool(public) {
						namesQuery.WithPublic = true
					}
					query.Keywords["names"] = namesQuery
				}
			} else {
				// can't get the user info, force to return public projects
				query.Keywords["public"] = true
			}
		}
	} else {
		if params.Public != nil && !*params.Public {
			// anonymous want to query private projects return empty projects directly
			return operation.NewListProjectsOK().WithXTotalCount(0).WithPayload([]*models.Project{})
		}
		// force to return public projects for anonymous
		query.Keywords["public"] = true
	}

	total, err := a.projectCtl.Count(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if total == 0 {
		// no projects found for the query return directly
		return operation.NewListProjectsOK().WithXTotalCount(0).WithPayload([]*models.Project{})
	}

	projects, err := a.projectCtl.List(ctx, query, project.Detail(lib.BoolValue(params.WithDetail)), project.WithCVEAllowlist(), project.WithOwner())
	if err != nil {
		return a.SendError(ctx, err)
	}

	var wg sync.WaitGroup
	for _, p := range projects {
		wg.Add(1)
		go func(p *project.Project) {
			defer wg.Done()
			// simultaneous queries in transaction will fail, so clone a ctx with new ormer here
			if err := a.populateProperties(orm.Clone(ctx), p); err != nil {
				log.G(ctx).Errorf("failed to populate properties for project %s, error: %v", p.Name, err)
			}
		}(p)
	}
	wg.Wait()

	var payload []*models.Project
	for _, p := range projects {
		payload = append(payload, model.NewProject(p).ToSwagger())
	}

	return operation.NewListProjectsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(payload)
}

func (a *projectAPI) UpdateProject(ctx context.Context, params operation.UpdateProjectParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate); err != nil {
		return a.SendError(ctx, err)
	}

	p, err := a.projectCtl.Get(ctx, projectNameOrID, project.Metadata(false))
	if err != nil {
		return a.SendError(ctx, err)
	}

	if params.Project.CVEAllowlist != nil {
		if params.Project.CVEAllowlist.ProjectID == 0 {
			// project_id in cve_allowlist not provided or provided as 0, let it to be the id of the project which will be updating
			params.Project.CVEAllowlist.ProjectID = p.ProjectID
		} else if params.Project.CVEAllowlist.ProjectID != p.ProjectID {
			return a.SendError(ctx, errors.BadRequestError(nil).
				WithMessagef("project_id in cve_allowlist must be %d but it's %d", p.ProjectID, params.Project.CVEAllowlist.ProjectID))
		}

		if err := lib.JSONCopy(&p.CVEAllowlist, params.Project.CVEAllowlist); err != nil {
			return a.SendError(ctx, errors.UnknownError(nil).WithMessagef("failed to process cve_allowlist, error: %v", err))
		}
	}

	// ignore metadata.proxy_speed_kb for non-proxy-cache project
	if params.Project.Metadata != nil && !p.IsProxy() {
		params.Project.Metadata.ProxySpeedKb = nil
	}

	// ignore enable_content_trust metadata for proxy cache project
	// see https://github.com/goharbor/harbor/issues/12940 to get more info
	if params.Project.Metadata != nil && p.IsProxy() {
		params.Project.Metadata.EnableContentTrust = nil
	}
	if err := lib.JSONCopy(&p.Metadata, params.Project.Metadata); err != nil {
		log.Warningf("failed to call JSONCopy on project metadata when UpdateProject, error: %v", err)
	}

	// validate retention_id
	if ridParam, ok := p.Metadata["retention_id"]; ok {
		md, err := a.metadataMgr.Get(ctx, p.ProjectID)
		if err != nil {
			return a.SendError(ctx, err)
		}
		if rid, ok := md["retention_id"]; !ok || rid != ridParam {
			errMsg := "the retention_id in the request's payload when updating a project should be omitted, alternatively passing the one that has already been associated to this project"
			return a.SendError(ctx, errors.New(nil).WithMessage(errMsg).WithCode(errors.BadRequestCode))
		}
	}

	if err := a.projectCtl.Update(ctx, p); err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewUpdateProjectOK()
}

func (a *projectAPI) GetScannerOfProject(ctx context.Context, params operation.GetScannerOfProjectParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceScanner); err != nil {
		return a.SendError(ctx, err)
	}

	p, err := a.projectCtl.Get(ctx, projectNameOrID, project.Metadata(false))
	if err != nil {
		return a.SendError(ctx, err)
	}

	s, err := a.scannerCtl.GetRegistrationByProject(ctx, p.ProjectID)
	if err != nil {
		return a.SendError(ctx, err)
	}
	if s != nil {
		if err := a.scannerCtl.RetrieveCap(ctx, s); err != nil {
			log.Warningf(scanner.RetrieveCapFailMsg, err)
		}
	}
	return operation.NewGetScannerOfProjectOK().WithPayload(model.NewScannerRegistration(s).ToSwagger(ctx))
}

func (a *projectAPI) ListScannerCandidatesOfProject(ctx context.Context, params operation.ListScannerCandidatesOfProjectParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceScanner); err != nil {
		return a.SendError(ctx, err)
	}

	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}

	total, err := a.scannerCtl.GetTotalOfRegistrations(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}

	scanners, err := a.scannerCtl.ListRegistrations(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}

	payload := make([]*models.ScannerRegistration, len(scanners))
	for i, scanner := range scanners {
		payload[i] = model.NewScannerRegistration(scanner).ToSwagger(ctx)
	}

	return operation.NewListScannerCandidatesOfProjectOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(payload)
}

func (a *projectAPI) SetScannerOfProject(ctx context.Context, params operation.SetScannerOfProjectParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceScanner); err != nil {
		return a.SendError(ctx, err)
	}

	p, err := a.projectCtl.Get(ctx, projectNameOrID, project.Metadata(false))
	if err != nil {
		return a.SendError(ctx, err)
	}

	if err := a.scannerCtl.SetRegistrationByProject(ctx, p.ProjectID, *params.Payload.UUID); err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewSetScannerOfProjectOK()
}

func (a *projectAPI) ListArtifactsOfProject(ctx context.Context, params operation.ListArtifactsOfProjectParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := a.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	// set query
	pro, err := a.projectCtl.Get(ctx, projectNameOrID)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["ProjectID"] = pro.ProjectID

	// set option
	option := option(params.WithTag, params.WithImmutableStatus,
		params.WithLabel, params.WithAccessory, params.LatestInRepository)

	var total int64
	// list artifacts according to the query and option
	var arts []*artifact.Artifact
	if option.LatestInRepository {
		// ignore page & page_size
		_, hasMediaType := query.Keywords["media_type"]
		_, hasArtifactType := query.Keywords["artifact_type"]
		if hasMediaType == hasArtifactType {
			return a.SendError(ctx, errors.BadRequestError(fmt.Errorf("either 'media_type' or 'artifact_type' must be specified, but not both, when querying with latest_in_repository")))
		}

		getCount := func() (int64, error) {
			var countQ *q.Query
			if query != nil {
				countQ = q.New(query.Keywords)
			}
			allArts, err := a.artCtl.ListWithLatest(ctx, countQ, nil)
			if err != nil {
				return int64(0), err
			}
			return int64(len(allArts)), nil
		}
		total, err = getCount()
		if err != nil {
			return a.SendError(ctx, err)
		}
		arts, err = a.artCtl.ListWithLatest(ctx, query, option)
	} else {
		total, err = a.artCtl.Count(ctx, query)
		if err != nil {
			return a.SendError(ctx, err)
		}
		arts, err = a.artCtl.List(ctx, query, option)
	}
	if err != nil {
		return a.SendError(ctx, err)
	}
	overviewOpts := model.NewOverviewOptions(model.WithSBOM(lib.BoolValue(params.WithSbomOverview)), model.WithVuln(lib.BoolValue(params.WithScanOverview)))
	assembler := assembler.NewScanReportAssembler(overviewOpts, parseScanReportMimeTypes(params.XAcceptVulnerabilities))
	var artifacts []*models.Artifact
	for _, art := range arts {
		artifact := &model.Artifact{}
		artifact.Artifact = *art
		_ = assembler.WithArtifacts(artifact).Assemble(ctx)
		artifacts = append(artifacts, artifact.ToSwagger())
	}

	return operation.NewListArtifactsOfProjectOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(artifacts)
}

func (a *projectAPI) deletable(ctx context.Context, projectNameOrID interface{}) (*project.Project, *models.ProjectDeletable, error) {
	p, err := a.getProject(ctx, projectNameOrID)
	if err != nil {
		return nil, nil, err
	}

	result := &models.ProjectDeletable{Deletable: true}
	if p.RepoCount > 0 {
		result.Deletable = false
		result.Message = "the project contains repositories, can not be deleted"
	}

	return p, result, nil
}

func (a *projectAPI) getProject(ctx context.Context, projectNameOrID interface{}, options ...project.Option) (*project.Project, error) {
	p, err := a.projectCtl.Get(ctx, projectNameOrID, options...)
	if err != nil {
		return nil, err
	}

	if err := a.populateProperties(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (a *projectAPI) validateProjectReq(ctx context.Context, req *models.ProjectReq) error {
	if req.Metadata.RetentionID != nil && *req.Metadata.RetentionID != "" {
		return errors.BadRequestError(fmt.Errorf("the retention_id in the request's payload when creating a project should be omitted, alternatively passing an empty string"))
	}

	if req.RegistryID != nil {
		if *req.RegistryID <= 0 {
			return errors.BadRequestError(fmt.Errorf("%d is invalid value of registry_id, it should be geater than 0", *req.RegistryID))
		}

		registry, err := registry.Ctl.Get(ctx, *req.RegistryID)
		if err != nil {
			return fmt.Errorf("failed to get the registry %d: %v", *req.RegistryID, err)
		}
		permitted := false
		for _, t := range config.GetPermittedRegistryTypesForProxyCache() {
			if string(registry.Type) == t {
				permitted = true
				break
			}
		}
		if !permitted {
			return errors.BadRequestError(fmt.Errorf("unsupported registry type %s", string(registry.Type)))
		}

		// validate metadata.proxy_speed_kb. It should be an int32
		if ps := req.Metadata.ProxySpeedKb; ps != nil {
			if _, err := strconv.ParseInt(*ps, 10, 32); err != nil {
				return errors.BadRequestError(nil).WithMessagef("metadata.proxy_speed_kb should by an int32, but got: '%s', err: %s", *ps, err)
			}
		}
	}

	if req.StorageLimit != nil {
		hardLimits := types.ResourceList{types.ResourceStorage: *req.StorageLimit}
		if err := quota.Validate(ctx, quota.ProjectReference, hardLimits); err != nil {
			return errors.BadRequestError(err)
		}
	}

	return nil
}

func (a *projectAPI) populateProperties(ctx context.Context, p *project.Project) error {
	if secCtx, ok := security.FromContext(ctx); ok {
		if sc, ok := secCtx.(*local.SecurityContext); ok {
			roles, err := a.projectCtl.ListRoles(ctx, p.ProjectID, sc.User())
			if err != nil {
				return err
			}
			p.RoleList = roles
			p.Role = highestRole(roles)
		}
	}

	total, err := a.repositoryCtl.Count(ctx, q.New(q.KeyWords{"project_id": p.ProjectID}))
	if err != nil {
		return err
	}
	p.RepoCount = total

	return nil
}

func (a *projectAPI) isSysAdmin(ctx context.Context, action rbac.Action) bool {
	if err := a.RequireSystemAccess(ctx, action, rbac.ResourceProject); err != nil {
		return false
	}
	return true
}

func getProjectQuotaSummary(ctx context.Context, p *project.Project, summary *models.ProjectSummary) {
	if !config.QuotaPerProjectEnable(ctx) {
		log.Debug("Quota per project deactivated")
		return
	}

	q, err := quota.Ctl.GetByRef(ctx, quota.ProjectReference, quota.ReferenceID(p.ProjectID))
	if err != nil {
		log.Warningf("failed to get quota for project: %d", p.ProjectID)
		return
	}

	summary.Quota = &models.ProjectSummaryQuota{}
	if hard, err := q.GetHard(); err == nil {
		summary.Quota.Hard = model.NewResourceList(hard).ToSwagger()
	}
	if used, err := q.GetUsed(); err == nil {
		summary.Quota.Used = model.NewResourceList(used).ToSwagger()
	}
}

func (a *projectAPI) getProjectMemberSummary(ctx context.Context, p *project.Project, summary *models.ProjectSummary) {
	var wg sync.WaitGroup

	for _, e := range []struct {
		role  int
		count *int64
	}{
		{common.RoleProjectAdmin, &summary.ProjectAdminCount},
		{common.RoleMaintainer, &summary.MaintainerCount},
		{common.RoleDeveloper, &summary.DeveloperCount},
		{common.RoleGuest, &summary.GuestCount},
		{common.RoleLimitedGuest, &summary.LimitedGuestCount},
	} {
		wg.Add(1)
		go func(role int, count *int64) {
			defer wg.Done()
			total, err := a.memberMgr.GetTotalOfProjectMembers(orm.Clone(ctx), p.ProjectID, nil, role)
			if err != nil {
				log.Warningf("failed to get total of project members of role %d", role)
				return
			}

			*count = int64(total)
		}(e.role, e.count)
	}

	wg.Wait()
}

func getProjectRegistrySummary(ctx context.Context, p *project.Project, summary *models.ProjectSummary) {
	if p.RegistryID <= 0 {
		return
	}

	registry, err := registry.Ctl.Get(ctx, p.RegistryID)
	if err != nil {
		log.Warningf("failed to get registry %d: %v", p.RegistryID, err)
	} else if registry != nil {
		registry.Credential = nil
		if err := lib.JSONCopy(&summary.Registry, registry); err != nil {
			log.Warningf("failed to call JSONCopy on project registry summary, error: %v", err)
		}
	}
}

// Returns the highest role in the role list.
// This func should be removed once we deprecate the "current_user_role_id" in project API
// A user can have multiple roles and they may not have a strict ranking relationship
func highestRole(roles []int) int {
	if roles == nil {
		return 0
	}
	rolePower := map[int]int{
		common.RoleProjectAdmin: 50,
		common.RoleMaintainer:   40,
		common.RoleDeveloper:    30,
		common.RoleGuest:        20,
		common.RoleLimitedGuest: 10,
	}
	var highest, highestPower int
	for _, role := range roles {
		if p, ok := rolePower[role]; ok && p > highestPower {
			highest = role
			highestPower = p
		}
	}
	return highest
}

func (a *projectAPI) GetLogExts(ctx context.Context, params operation.GetLogExtsParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceLog); err != nil {
		return a.SendError(ctx, err)
	}
	pro, err := a.projectCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query, err := a.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["ProjectID"] = pro.ProjectID

	total, err := a.auditextMgr.Count(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}
	logs, err := a.auditextMgr.List(ctx, query)
	if err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewGetLogExtsOK().
		WithXTotalCount(total).
		WithLink(a.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(convertToModelAuditLogExt(logs))
}
