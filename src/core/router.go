// Copyright 2018 Project Harbor Authors
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

package main

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/controllers"
	"github.com/goharbor/harbor/src/core/service/notifications/admin"
	"github.com/goharbor/harbor/src/core/service/notifications/jobs"
	"github.com/goharbor/harbor/src/core/service/notifications/registry"
	"github.com/goharbor/harbor/src/core/service/notifications/scheduler"
	"github.com/goharbor/harbor/src/core/service/token"

	"github.com/astaxie/beego"
)

func initRouters() {

	// standalone
	if !config.WithAdmiral() {
		// Controller API:
		beego.Router("/c/login", &controllers.CommonController{}, "post:Login")
		beego.Router("/c/log_out", &controllers.CommonController{}, "get:LogOut")
		beego.Router("/c/reset", &controllers.CommonController{}, "post:ResetPassword")
		beego.Router("/c/userExists", &controllers.CommonController{}, "post:UserExists")
		beego.Router("/c/sendEmail", &controllers.CommonController{}, "get:SendResetEmail")
		beego.Router(common.OIDCLoginPath, &controllers.OIDCController{}, "get:RedirectLogin")
		beego.Router("/c/oidc/onboard", &controllers.OIDCController{}, "post:Onboard")
		beego.Router(common.OIDCCallbackPath, &controllers.OIDCController{}, "get:Callback")

		// API:
		beego.Router("/api/projects/:pid([0-9]+)/members/?:pmid([0-9]+)", &api.ProjectMemberAPI{})
		beego.Router("/api/projects/", &api.ProjectAPI{}, "head:Head")
		beego.Router("/api/projects/:id([0-9]+)", &api.ProjectAPI{})

		beego.Router("/api/users/:id", &api.UserAPI{}, "get:Get;delete:Delete;put:Put")
		beego.Router("/api/users", &api.UserAPI{}, "get:List;post:Post")
		beego.Router("/api/users/search", &api.UserAPI{}, "get:Search")
		beego.Router("/api/users/:id([0-9]+)/password", &api.UserAPI{}, "put:ChangePassword")
		beego.Router("/api/users/:id/permissions", &api.UserAPI{}, "get:ListUserPermissions")
		beego.Router("/api/users/:id/sysadmin", &api.UserAPI{}, "put:ToggleUserAdminRole")
		beego.Router("/api/users/:id/gen_cli_secret", &api.UserAPI{}, "post:GenCLISecret")
		beego.Router("/api/usergroups/?:ugid([0-9]+)", &api.UserGroupAPI{})
		beego.Router("/api/ldap/ping", &api.LdapAPI{}, "post:Ping")
		beego.Router("/api/ldap/users/search", &api.LdapAPI{}, "get:Search")
		beego.Router("/api/ldap/groups/search", &api.LdapAPI{}, "get:SearchGroup")
		beego.Router("/api/ldap/users/import", &api.LdapAPI{}, "post:ImportUser")
		beego.Router("/api/email/ping", &api.EmailAPI{}, "post:Ping")
	}

	// API
	beego.Router("/api/health", &api.HealthAPI{}, "get:CheckHealth")
	beego.Router("/api/ping", &api.SystemInfoAPI{}, "get:Ping")
	beego.Router("/api/search", &api.SearchAPI{})
	beego.Router("/api/projects/", &api.ProjectAPI{}, "get:List;post:Post")
	beego.Router("/api/projects/:id([0-9]+)/summary", &api.ProjectAPI{}, "get:Summary")
	beego.Router("/api/projects/:id([0-9]+)/logs", &api.ProjectAPI{}, "get:Logs")
	beego.Router("/api/projects/:id([0-9]+)/_deletable", &api.ProjectAPI{}, "get:Deletable")
	beego.Router("/api/projects/:id([0-9]+)/metadatas/?:name", &api.MetadataAPI{}, "get:Get")
	beego.Router("/api/projects/:id([0-9]+)/metadatas/", &api.MetadataAPI{}, "post:Post")
	beego.Router("/api/projects/:id([0-9]+)/metadatas/:name", &api.MetadataAPI{}, "put:Put;delete:Delete")

	beego.Router("/api/projects/:pid([0-9]+)/robots", &api.RobotAPI{}, "post:Post;get:List")
	beego.Router("/api/projects/:pid([0-9]+)/robots/:id([0-9]+)", &api.RobotAPI{}, "get:Get;put:Put;delete:Delete")

	beego.Router("/api/quotas", &api.QuotaAPI{}, "get:List")
	beego.Router("/api/quotas/:id([0-9]+)", &api.QuotaAPI{}, "get:Get;put:Put")

	beego.Router("/api/repositories", &api.RepositoryAPI{}, "get:Get")
	beego.Router("/api/repositories/*", &api.RepositoryAPI{}, "delete:Delete;put:Put")
	beego.Router("/api/repositories/*/labels", &api.RepositoryLabelAPI{}, "get:GetOfRepository;post:AddToRepository")
	beego.Router("/api/repositories/*/labels/:id([0-9]+)", &api.RepositoryLabelAPI{}, "delete:RemoveFromRepository")
	beego.Router("/api/repositories/*/tags/:tag", &api.RepositoryAPI{}, "delete:Delete;get:GetTag")
	beego.Router("/api/repositories/*/tags/:tag/labels", &api.RepositoryLabelAPI{}, "get:GetOfImage;post:AddToImage")
	beego.Router("/api/repositories/*/tags/:tag/labels/:id([0-9]+)", &api.RepositoryLabelAPI{}, "delete:RemoveFromImage")
	beego.Router("/api/repositories/*/tags", &api.RepositoryAPI{}, "get:GetTags;post:Retag")
	beego.Router("/api/repositories/*/tags/:tag/scan", &api.RepositoryAPI{}, "post:ScanImage")
	beego.Router("/api/repositories/*/tags/:tag/vulnerability/details", &api.RepositoryAPI{}, "Get:VulnerabilityDetails")
	beego.Router("/api/repositories/*/tags/:tag/manifest", &api.RepositoryAPI{}, "get:GetManifests")
	beego.Router("/api/repositories/*/signatures", &api.RepositoryAPI{}, "get:GetSignatures")
	beego.Router("/api/repositories/top", &api.RepositoryAPI{}, "get:GetTopRepos")
	beego.Router("/api/jobs/scan/:id([0-9]+)/log", &api.ScanJobAPI{}, "get:GetLog")

	beego.Router("/api/system/gc", &api.GCAPI{}, "get:List")
	beego.Router("/api/system/gc/:id", &api.GCAPI{}, "get:GetGC")
	beego.Router("/api/system/gc/:id([0-9]+)/log", &api.GCAPI{}, "get:GetLog")
	beego.Router("/api/system/gc/schedule", &api.GCAPI{}, "get:Get;put:Put;post:Post")
	beego.Router("/api/system/scanAll/schedule", &api.ScanAllAPI{}, "get:Get;put:Put;post:Post")
	beego.Router("/api/system/CVEWhitelist", &api.SysCVEWhitelistAPI{}, "get:Get;put:Put")
	beego.Router("/api/system/oidc/ping", &api.OIDCAPI{}, "post:Ping")

	beego.Router("/api/logs", &api.LogAPI{})

	beego.Router("/api/replication/adapters", &api.ReplicationAdapterAPI{}, "get:List")
	beego.Router("/api/replication/executions", &api.ReplicationOperationAPI{}, "get:ListExecutions;post:CreateExecution")
	beego.Router("/api/replication/executions/:id([0-9]+)", &api.ReplicationOperationAPI{}, "get:GetExecution;put:StopExecution")
	beego.Router("/api/replication/executions/:id([0-9]+)/tasks", &api.ReplicationOperationAPI{}, "get:ListTasks")
	beego.Router("/api/replication/executions/:id([0-9]+)/tasks/:tid([0-9]+)/log", &api.ReplicationOperationAPI{}, "get:GetTaskLog")

	beego.Router("/api/replication/policies", &api.ReplicationPolicyAPI{}, "get:List;post:Create")
	beego.Router("/api/replication/policies/:id([0-9]+)", &api.ReplicationPolicyAPI{}, "get:Get;put:Update;delete:Delete")

	beego.Router("/api/internal/configurations", &api.ConfigAPI{}, "get:GetInternalConfig;put:Put")
	beego.Router("/api/configurations", &api.ConfigAPI{}, "get:Get;put:Put")
	beego.Router("/api/statistics", &api.StatisticAPI{})
	beego.Router("/api/labels", &api.LabelAPI{}, "post:Post;get:List")
	beego.Router("/api/labels/:id([0-9]+)", &api.LabelAPI{}, "get:Get;put:Put;delete:Delete")
	beego.Router("/api/labels/:id([0-9]+)/resources", &api.LabelAPI{}, "get:ListResources")

	beego.Router("/api/systeminfo", &api.SystemInfoAPI{}, "get:GetGeneralInfo")
	beego.Router("/api/systeminfo/volumes", &api.SystemInfoAPI{}, "get:GetVolumeInfo")
	beego.Router("/api/systeminfo/getcert", &api.SystemInfoAPI{}, "get:GetCert")

	beego.Router("/api/internal/syncregistry", &api.InternalAPI{}, "post:SyncRegistry")
	beego.Router("/api/internal/renameadmin", &api.InternalAPI{}, "post:RenameAdmin")

	// external service that hosted on harbor process:
	beego.Router("/service/notifications", &registry.NotificationHandler{})
	beego.Router("/service/notifications/jobs/scan/:id([0-9]+)", &jobs.Handler{}, "post:HandleScan")
	beego.Router("/service/notifications/jobs/adminjob/:id([0-9]+)", &admin.Handler{}, "post:HandleAdminJob")
	beego.Router("/service/notifications/jobs/replication/:id([0-9]+)", &jobs.Handler{}, "post:HandleReplicationScheduleJob")
	beego.Router("/service/notifications/jobs/replication/task/:id([0-9]+)", &jobs.Handler{}, "post:HandleReplicationTask")
	beego.Router("/service/notifications/jobs/retention/task/:id([0-9]+)", &jobs.Handler{}, "post:HandleRetentionTask")
	beego.Router("/service/notifications/schedules/:id([0-9]+)", &scheduler.Handler{}, "post:Handle")
	beego.Router("/service/token", &token.Handler{})

	beego.Router("/api/registries", &api.RegistryAPI{}, "get:List;post:Post")
	beego.Router("/api/registries/:id([0-9]+)", &api.RegistryAPI{}, "get:Get;put:Put;delete:Delete")
	beego.Router("/api/registries/ping", &api.RegistryAPI{}, "post:Ping")
	// we use "0" as the ID of the local Harbor registry, so don't add "([0-9]+)" in the path
	beego.Router("/api/registries/:id/info", &api.RegistryAPI{}, "get:GetInfo")
	beego.Router("/api/registries/:id/namespace", &api.RegistryAPI{}, "get:GetNamespace")

	beego.Router("/api/retentions/metadatas", &api.RetentionAPI{}, "get:GetMetadatas")
	beego.Router("/api/retentions/:id", &api.RetentionAPI{}, "get:GetRetention")
	beego.Router("/api/retentions", &api.RetentionAPI{}, "post:CreateRetention")
	beego.Router("/api/retentions/:id", &api.RetentionAPI{}, "put:UpdateRetention")
	beego.Router("/api/retentions/:id/executions", &api.RetentionAPI{}, "post:TriggerRetentionExec")
	beego.Router("/api/retentions/:id/executions/:eid", &api.RetentionAPI{}, "patch:OperateRetentionExec")
	beego.Router("/api/retentions/:id/executions", &api.RetentionAPI{}, "get:ListRetentionExecs")
	beego.Router("/api/retentions/:id/executions/:eid/tasks", &api.RetentionAPI{}, "get:ListRetentionExecTasks")
	beego.Router("/api/retentions/:id/executions/:eid/tasks/:tid", &api.RetentionAPI{}, "get:GetRetentionExecTaskLog")

	beego.Router("/v2/*", &controllers.RegistryProxy{}, "*:Handle")

	// APIs for chart repository
	if config.WithChartMuseum() {
		// Charts are controlled under projects
		chartRepositoryAPIType := &api.ChartRepositoryAPI{}
		beego.Router("/api/chartrepo/health", chartRepositoryAPIType, "get:GetHealthStatus")
		beego.Router("/api/chartrepo/:repo/charts", chartRepositoryAPIType, "get:ListCharts")
		beego.Router("/api/chartrepo/:repo/charts/:name", chartRepositoryAPIType, "get:ListChartVersions")
		beego.Router("/api/chartrepo/:repo/charts/:name", chartRepositoryAPIType, "delete:DeleteChart")
		beego.Router("/api/chartrepo/:repo/charts/:name/:version", chartRepositoryAPIType, "get:GetChartVersion")
		beego.Router("/api/chartrepo/:repo/charts/:name/:version", chartRepositoryAPIType, "delete:DeleteChartVersion")
		beego.Router("/api/chartrepo/:repo/charts", chartRepositoryAPIType, "post:UploadChartVersion")
		beego.Router("/api/chartrepo/:repo/prov", chartRepositoryAPIType, "post:UploadChartProvFile")
		beego.Router("/api/chartrepo/charts", chartRepositoryAPIType, "post:UploadChartVersion")

		// Repository services
		beego.Router("/chartrepo/:repo/index.yaml", chartRepositoryAPIType, "get:GetIndexByRepo")
		beego.Router("/chartrepo/index.yaml", chartRepositoryAPIType, "get:GetIndex")
		beego.Router("/chartrepo/:repo/charts/:filename", chartRepositoryAPIType, "get:DownloadChart")

		// Labels for chart
		chartLabelAPIType := &api.ChartLabelAPI{}
		beego.Router("/api/chartrepo/:repo/charts/:name/:version/labels", chartLabelAPIType, "get:GetLabels;post:MarkLabel")
		beego.Router("/api/chartrepo/:repo/charts/:name/:version/labels/:id([0-9]+)", chartLabelAPIType, "delete:RemoveLabel")
	}

	// Error pages
	beego.ErrorController(&controllers.ErrorController{})

}
