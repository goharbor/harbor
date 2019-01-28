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

package project

import (
	"github.com/goharbor/harbor/src/common/rbac"
)

// const action variables
const (
	ActionAll = rbac.Action("*") // action match any other actions

	ActionPull     = rbac.Action("pull")      // pull repository tag
	ActionPush     = rbac.Action("push")      // push repository tag
	ActionPushPull = rbac.Action("push+pull") // compatible with security all perm of project

	// create, read, update, delete, list actions compatible with restful api methods
	ActionCreate = rbac.Action("create")
	ActionRead   = rbac.Action("read")
	ActionUpdate = rbac.Action("update")
	ActionDelete = rbac.Action("delete")
	ActionList   = rbac.Action("list")

	// execute replication for the replication policy (replication rule)
	ActionExecute = rbac.Action("execute")

	// vulnerabilities scan for repository tag (aka, image tag)
	ActionScan = rbac.Action("scan")
)

// const resource variables
const (
	ResourceAll                        = rbac.Resource("*") // resource match any other resources
	ResourceSelf                       = rbac.Resource("")  // subresource for project self
	ResourceMember                     = rbac.Resource("member")
	ResourceLog                        = rbac.Resource("log")
	ResourceReplication                = rbac.Resource("replication")
	ResourceLabel                      = rbac.Resource("label")
	ResourceRepository                 = rbac.Resource("repository")
	ResourceRepositoryTag              = rbac.Resource("repository-tag")
	ResourceRepositoryTagManifest      = rbac.Resource("repository-tag-manifest")
	ResourceRepositoryTagVulnerability = rbac.Resource("repository-tag-vulnerability")
	ResourceRepositoryTagLabel         = rbac.Resource("repository-tag-label")
	ResourceHelmChart                  = rbac.Resource("helm-chart")
	ResourceHelmChartVersion           = rbac.Resource("helm-chart-version")
	ResourceHelmChartVersionLabel      = rbac.Resource("helm-chart-version-label")
	ResourceConfiguration              = rbac.Resource("configuration") // compatible for portal only
	ResourceRobot                      = rbac.Resource("robot")
)
