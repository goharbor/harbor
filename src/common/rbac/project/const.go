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
	ActionAll                 = rbac.Action("*")
	ActionPull                = rbac.Action("pull")
	ActionPush                = rbac.Action("push")
	ActionPushPull            = rbac.Action("push+pull")
	ActionCreate              = rbac.Action("create")
	ActionRead                = rbac.Action("read")
	ActionUpdate              = rbac.Action("update")
	ActionDelete              = rbac.Action("delete")
	ActionList                = rbac.Action("list")
	ActionExecute             = rbac.Action("execute")
	ActionScan                = rbac.Action("scan")
	ActionReTag               = rbac.Action("retag")
	ActionListVulnerabilities = rbac.Action("list-vulnerabilities")
	ActionReadManifest        = rbac.Action("read-manifest")
	ActionAddLabel            = rbac.Action("add-label")
	ActionRemoveLabel         = rbac.Action("remove-label")
	ActionUpload              = rbac.Action("upload")
	ActionDownload            = rbac.Action("download")
)

// const resource variables
const (
	ResourceAll              = rbac.Resource("*")
	ResourceSelf             = rbac.Resource("") // subresource for project self
	ResourceMember           = rbac.Resource("member")
	ResourceLog              = rbac.Resource("log")
	ResourceReplication      = rbac.Resource("replication")
	ResourceLabel            = rbac.Resource("label")
	ResourceRepository       = rbac.Resource("repository")
	ResourceRepositoryTag    = rbac.Resource("repository-tag")
	ResourceHelmChart        = rbac.Resource("helm-chart")
	ResourceHelmChartVersion = rbac.Resource("helm-chart-version")
	ResourceConfiguration    = rbac.Resource("configuration") // compatible for portal only
)
