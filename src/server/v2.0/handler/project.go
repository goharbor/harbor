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
	pro "github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/p2p/preheat"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/audit"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/quota/types"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/robot"
	"github.com/goharbor/harbor/src/pkg/user"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project"
)

// for the proxy cache type project, we will create a 7 days retention policy for it by default
const defaultDaysToRetentionForProxyCacheProject = 7

func newProjectAPI() *projectAPI {
	return &projectAPI{
		auditMgr:      audit.Mgr,
		metadataMgr:   metadata.Mgr,
		userMgr:       user.Mgr,
		repositoryCtl: repository.Ctl,
		projectCtl:    project.Ctl,
		quotaCtl:      quota.Ctl,
		robotMgr:      robot.Mgr,
		preheatCtl:    preheat.Ctl,
	}
}

type projectAPI struct {
	BaseAPI
	auditMgr      audit.Manager
	metadataMgr   metadata.Manager
	userMgr       user.Manager
	repositoryCtl repository.Controller
	projectCtl    project.Controller
	quotaCtl      quota.Controller
	robotMgr      robot.Manager
	preheatCtl    preheat.Controller
}

func (a *projectAPI) CreateProject(ctx context.Context, params operation.CreateProjectParams) middleware.Responder {
	if err := a.RequireAuthenticated(ctx); err != nil {
		return a.SendError(ctx, err)
	}

	onlyAdmin, err := config.OnlyAdminCreateProject()
	if err != nil {
		return a.SendError(ctx, fmt.Errorf("failed to determine whether only admin can create projects: %v", err))
	}

	secCtx, _ := security.FromContext(ctx)
	if onlyAdmin && !(secCtx.IsSysAdmin() || secCtx.IsSolutionUser()) {
		log.Errorf("Only sys admin can create project")
		return a.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Only system admin can create project"))
	}

	req := params.Project

	if req.RegistryID != nil && !secCtx.IsSysAdmin() {
		// only system admin can create the proxy cache project
		return a.SendError(ctx, errors.ForbiddenError(nil).WithMessage("Only system admin can create proxy cache project"))
	}

	// populate storage limit
	if config.QuotaPerProjectEnable() {
		// the security context is not sys admin, set the StorageLimit the global StoragePerProject
		if req.StorageLimit == nil || *req.StorageLimit == 0 || !secCtx.IsSysAdmin() {
			setting, err := config.QuotaSetting()
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

	// ignore enable_content_trust metadata for proxy cache project
	// see https://github.com/goharbor/harbor/issues/12940 to get more info
	if req.RegistryID != nil {
		req.Metadata.EnableContentTrust = nil
	}

	// validate the RegistryID and StorageLimit in the body of the request
	if err := a.validateProjectReq(ctx, req); err != nil {
		return a.SendError(ctx, err)
	}

	var ownerID int
	// set the owner as the system admin when the API being called by replication
	// it's a solution to workaround the restriction of project creation API:
	// only normal users can create projects
	if secCtx.IsSolutionUser() {
		ownerID = 1
	} else {
		ownerName := secCtx.GetUsername()
		user, err := a.userMgr.GetByName(ctx, ownerName)
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
	lib.JSONCopy(&p.Metadata, req.Metadata)

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
		// TODO: move the retention controller to `src/controller/retention` and
		// change to use the default retention controller in `src/controller/retention`
		retentionID, err := api.GetRetentionController().CreateRetention(plc)
		if err != nil {
			return a.SendError(ctx, err)
		}
		md := map[string]string{"retention_id": strconv.FormatInt(retentionID, 10)}
		if err := a.metadataMgr.Add(ctx, projectID, md); err != nil {
			return a.SendError(ctx, err)
		}
		return nil
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), projectID)
	return operation.NewCreateProjectCreated().WithLocation(location)
}

func (a *projectAPI) DeleteProject(ctx context.Context, params operation.DeleteProjectParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectID, rbac.ActionDelete); err != nil {
		return a.SendError(ctx, err)
	}

	result, err := a.deletable(ctx, params.ProjectID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	if !result.Deletable {
		return a.SendError(ctx, errors.PreconditionFailedError(errors.New(result.Message)))
	}

	if err := a.projectCtl.Delete(ctx, params.ProjectID); err != nil {
		return a.SendError(ctx, err)
	}

	// remove the robot associated with the project
	if err := a.robotMgr.DeleteByProjectID(ctx, params.ProjectID); err != nil {
		return a.SendError(ctx, err)
	}

	referenceID := quota.ReferenceID(params.ProjectID)
	q, err := a.quotaCtl.GetByRef(ctx, quota.ProjectReference, referenceID)
	if err != nil {
		log.Warningf("failed to get quota for project %d, error: %v", params.ProjectID, err)
	} else {
		if err := a.quotaCtl.Delete(ctx, q.ID); err != nil {
			return a.SendError(ctx, fmt.Errorf("failed to delete quota for project: %v", err))
		}
	}

	// preheat policies under the project should be deleted after deleting the project
	if err = a.preheatCtl.DeletePoliciesOfProject(ctx, params.ProjectID); err != nil {
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
	query, err := a.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
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
	if err := a.RequireProjectAccess(ctx, params.ProjectID, rbac.ActionRead); err != nil {
		return a.SendError(ctx, err)
	}

	p, err := a.getProject(ctx, params.ProjectID, project.WithCVEAllowlist(), project.WithOwner())
	if err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewGetProjectOK().WithPayload(model.NewProject(p).ToSwagger())
}

func (a *projectAPI) GetProjectDeletable(ctx context.Context, params operation.GetProjectDeletableParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectID, rbac.ActionDelete); err != nil {
		return a.SendError(ctx, err)
	}

	result, err := a.deletable(ctx, params.ProjectID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewGetProjectDeletableOK().WithPayload(result)
}

func (a *projectAPI) GetProjectSummary(ctx context.Context, params operation.GetProjectSummaryParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectID, rbac.ActionRead); err != nil {
		return a.SendError(ctx, err)
	}

	p, err := a.getProject(ctx, params.ProjectID)
	if err != nil {
		return a.SendError(ctx, err)
	}

	summary := &models.ProjectSummary{
		ChartCount: int64(p.ChartCount),
		RepoCount:  p.RepoCount,
	}

	var fetchSummaries []func(context.Context, *project.Project, *models.ProjectSummary)

	if hasPerm := a.HasProjectPermission(ctx, p.ProjectID, rbac.ActionRead, rbac.ResourceQuota); hasPerm {
		fetchSummaries = append(fetchSummaries, getProjectQuotaSummary)
	}

	if hasPerm := a.HasProjectPermission(ctx, p.ProjectID, rbac.ActionList, rbac.ResourceMember); hasPerm {
		fetchSummaries = append(fetchSummaries, getProjectMemberSummary)
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
	query := q.New(q.KeyWords{})
	query.Sorting = "name"
	query.PageNumber = *params.Page
	query.PageSize = *params.PageSize

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
		if !secCtx.IsSysAdmin() && !secCtx.IsSolutionUser() {
			// authenticated but not system admin or solution user,
			// return public projects and projects that the user is member of
			if l, ok := secCtx.(*local.SecurityContext); ok {
				currentUser := l.User()
				member := &project.MemberQuery{
					Name:     currentUser.Username,
					GroupIDs: currentUser.GroupIDs,
				}

				// not filter by public or filter by the public with true,
				// so also return public projects for the member
				if public, ok := query.Keywords["public"]; !ok || lib.ToBool(public) {
					member.WithPublic = true
				}

				query.Keywords["member"] = member
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

	projects, err := a.projectCtl.List(ctx, query, project.WithCVEAllowlist(), project.WithOwner())
	if err != nil {
		return a.SendError(ctx, err)
	}

	var wg sync.WaitGroup
	for _, p := range projects {
		wg.Add(1)
		go func(p *project.Project) {
			defer wg.Done()
			// due to the issue https://github.com/lib/pq/issues/81 of lib/pg or postgres,
			// simultaneous queries in transaction may failed, so clone a ctx with new ormer here
			if err := a.populateProperties(orm.Clone(ctx), p); err != nil {
				log.G(ctx).Errorf("failed to populate propertites for project %s, error: %v", p.Name, err)
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
	if err := a.RequireProjectAccess(ctx, params.ProjectID, rbac.ActionUpdate); err != nil {
		return a.SendError(ctx, err)
	}

	p, err := a.projectCtl.Get(ctx, params.ProjectID, project.Metadata(false))
	if err != nil {
		return a.SendError(ctx, err)
	}

	if params.Project.CVEAllowlist != nil {
		if params.Project.CVEAllowlist.ProjectID == 0 {
			// project_id in cve_allowlist not provided or provided as 0, let it to be the id of the project which will be updating
			params.Project.CVEAllowlist.ProjectID = params.ProjectID
		} else if params.Project.CVEAllowlist.ProjectID != params.ProjectID {
			return a.SendError(ctx, errors.BadRequestError(nil).
				WithMessage("project_id in cve_allowlist must be %d but it's %d", params.ProjectID, params.Project.CVEAllowlist.ProjectID))
		}

		if err := lib.JSONCopy(&p.CVEAllowlist, params.Project.CVEAllowlist); err != nil {
			return a.SendError(ctx, errors.UnknownError(nil).WithMessage("failed to process cve_allowlist, error: %v", err))
		}
	}

	// ignore enable_content_trust metadata for proxy cache project
	// see https://github.com/goharbor/harbor/issues/12940 to get more info
	if params.Project.Metadata != nil && p.IsProxy() {
		params.Project.Metadata.EnableContentTrust = nil
	}
	lib.JSONCopy(&p.Metadata, params.Project.Metadata)

	if err := a.projectCtl.Update(ctx, p); err != nil {
		return a.SendError(ctx, err)
	}

	return operation.NewUpdateProjectOK()
}

func (a *projectAPI) deletable(ctx context.Context, projectID int64) (*models.ProjectDeletable, error) {
	proj, err := a.getProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	result := &models.ProjectDeletable{Deletable: true}
	if proj.RepoCount > 0 {
		result.Deletable = false
		result.Message = "the project contains repositories, can not be deleted"
	} else if proj.ChartCount > 0 {
		result.Deletable = false
		result.Message = "the project contains helm charts, can not be deleted"
	}

	return result, nil
}

func (a *projectAPI) getProject(ctx context.Context, projectID int64, options ...project.Option) (*project.Project, error) {
	p, err := a.projectCtl.Get(ctx, projectID, options...)
	if err != nil {
		return nil, err
	}

	if err := a.populateProperties(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (a *projectAPI) validateProjectReq(ctx context.Context, req *models.ProjectReq) error {
	if req.RegistryID != nil {
		if *req.RegistryID <= 0 {
			return errors.BadRequestError(fmt.Errorf("%d is invalid value of registry_id, it should be geater than 0", *req.RegistryID))
		}

		registry, err := replication.RegistryMgr.Get(*req.RegistryID)
		if err != nil {
			return fmt.Errorf("failed to get the registry %d: %v", *req.RegistryID, err)
		}
		if registry == nil {
			return errors.NotFoundError(fmt.Errorf("registry %d not found", *req.RegistryID))
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
			roles, err := pro.ListRoles(sc.User(), p.ProjectID)
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

	// Populate chart count property
	if config.WithChartMuseum() {
		count, err := api.GetChartController().GetCountOfCharts([]string{p.Name})
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("get chart count of project %d failed", p.ProjectID))
			return err
		}

		p.ChartCount = count
	}
	return nil
}

func getProjectQuotaSummary(ctx context.Context, p *project.Project, summary *models.ProjectSummary) {
	if !config.QuotaPerProjectEnable() {
		log.Debug("Quota per project disabled")
		return
	}

	q, err := quota.Ctl.GetByRef(ctx, quota.ProjectReference, quota.ReferenceID(p.ProjectID))
	if err != nil {
		log.Warningf("failed to get quota for project: %d", p.ProjectID)
		return
	}

	summary.Quota = &models.ProjectSummaryQuota{}
	if hard, err := q.GetHard(); err == nil {
		lib.JSONCopy(&summary.Quota.Hard, hard)
	}
	if used, err := q.GetUsed(); err == nil {
		lib.JSONCopy(&summary.Quota.Used, used)
	}
}

func getProjectMemberSummary(ctx context.Context, p *project.Project, summary *models.ProjectSummary) {
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

			total, err := pro.GetTotalOfProjectMembers(p.ProjectID, role)
			if err != nil {
				log.Warningf("failed to get total of project members of role %d", role)
				return
			}

			*count = total
		}(e.role, e.count)
	}

	wg.Wait()
}

func getProjectRegistrySummary(ctx context.Context, p *project.Project, summary *models.ProjectSummary) {
	if p.RegistryID <= 0 {
		return
	}

	registry, err := replication.RegistryMgr.Get(p.RegistryID)
	if err != nil {
		log.Warningf("failed to get registry %d: %v", p.RegistryID, err)
	} else if registry != nil {
		registry.Credential = nil
		lib.JSONCopy(&summary.Registry, registry)
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
