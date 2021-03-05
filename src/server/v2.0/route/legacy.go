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

package route

import (
	"github.com/astaxie/beego"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/config"
)

// RegisterRoutes for Harbor legacy APIs
func registerLegacyRoutes() {
	version := APIVersion
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/members/?:pmid([0-9]+)", &api.ProjectMemberAPI{})
	beego.Router("/api/"+version+"/users/:id", &api.UserAPI{}, "get:Get;delete:Delete;put:Put")
	beego.Router("/api/"+version+"/users", &api.UserAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/users/search", &api.UserAPI{}, "get:Search")
	beego.Router("/api/"+version+"/users/:id([0-9]+)/password", &api.UserAPI{}, "put:ChangePassword")
	beego.Router("/api/"+version+"/users/:id/permissions", &api.UserAPI{}, "get:ListUserPermissions")
	beego.Router("/api/"+version+"/users/:id/sysadmin", &api.UserAPI{}, "put:ToggleUserAdminRole")
	beego.Router("/api/"+version+"/users/:id/cli_secret", &api.UserAPI{}, "put:SetCLISecret")
	beego.Router("/api/"+version+"/usergroups/?:ugid([0-9]+)", &api.UserGroupAPI{})
	beego.Router("/api/"+version+"/ldap/ping", &api.LdapAPI{}, "post:Ping")
	beego.Router("/api/"+version+"/ldap/users/search", &api.LdapAPI{}, "get:Search")
	beego.Router("/api/"+version+"/ldap/groups/search", &api.LdapAPI{}, "get:SearchGroup")
	beego.Router("/api/"+version+"/ldap/users/import", &api.LdapAPI{}, "post:ImportUser")
	beego.Router("/api/"+version+"/email/ping", &api.EmailAPI{}, "post:Ping")
	beego.Router("/api/"+version+"/health", &api.HealthAPI{}, "get:CheckHealth")
	beego.Router("/api/"+version+"/search", &api.SearchAPI{})
	beego.Router("/api/"+version+"/projects/:id([0-9]+)/metadatas/?:name", &api.MetadataAPI{}, "get:Get")
	beego.Router("/api/"+version+"/projects/:id([0-9]+)/metadatas/", &api.MetadataAPI{}, "post:Post")

	beego.Router("/api/"+version+"/system/CVEAllowlist", &api.SysCVEAllowlistAPI{}, "get:Get;put:Put")

	beego.Router("/api/"+version+"/replication/adapters", &api.ReplicationAdapterAPI{}, "get:List")
	beego.Router("/api/"+version+"/replication/adapterinfos", &api.ReplicationAdapterAPI{}, "get:ListAdapterInfos")
	beego.Router("/api/"+version+"/replication/policies", &api.ReplicationPolicyAPI{}, "get:List;post:Create")
	beego.Router("/api/"+version+"/replication/policies/:id([0-9]+)", &api.ReplicationPolicyAPI{}, "get:Get;put:Update;delete:Delete")

	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/policies", &api.NotificationPolicyAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/policies/:id([0-9]+)", &api.NotificationPolicyAPI{})
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/lasttrigger", &api.NotificationPolicyAPI{}, "get:ListGroupByEventType")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/events", &api.NotificationPolicyAPI{}, "get:GetSupportedEventTypes")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/webhook/jobs/", &api.NotificationJobAPI{}, "get:List")

	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/immutabletagrules", &api.ImmutableTagRuleAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/immutabletagrules/:id([0-9]+)", &api.ImmutableTagRuleAPI{})

	beego.Router("/api/"+version+"/configurations", &api.ConfigAPI{}, "get:Get;put:Put")
	beego.Router("/api/"+version+"/statistics", &api.StatisticAPI{})
	beego.Router("/api/"+version+"/labels", &api.LabelAPI{}, "post:Post;get:List")
	beego.Router("/api/"+version+"/labels/:id([0-9]+)", &api.LabelAPI{}, "get:Get;put:Put;delete:Delete")

	beego.Router("/api/"+version+"/registries", &api.RegistryAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/registries/:id([0-9]+)", &api.RegistryAPI{}, "get:Get;put:Put;delete:Delete")
	beego.Router("/api/"+version+"/registries/ping", &api.RegistryAPI{}, "post:Ping")
	// we use "0" as the ID of the local Harbor registry, so don't add "([0-9]+)" in the path
	beego.Router("/api/"+version+"/registries/:id/info", &api.RegistryAPI{}, "get:GetInfo")
	beego.Router("/api/"+version+"/registries/:id/namespace", &api.RegistryAPI{}, "get:GetNamespace")

	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/immutabletagrules", &api.ImmutableTagRuleAPI{}, "get:List;post:Post")
	beego.Router("/api/"+version+"/projects/:pid([0-9]+)/immutabletagrules/:id([0-9]+)", &api.ImmutableTagRuleAPI{})

	// APIs for chart repository
	if config.WithChartMuseum() {
		// Labels for chart
		chartLabelAPIType := &api.ChartLabelAPI{}
		beego.Router("/api/"+version+"/chartrepo/:repo/charts/:name/:version/labels", chartLabelAPIType, "get:GetLabels;post:MarkLabel")
		beego.Router("/api/"+version+"/chartrepo/:repo/charts/:name/:version/labels/:id([0-9]+)", chartLabelAPIType, "delete:RemoveLabel")
	}
}
