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

package models

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/orm"
	allowlist "github.com/goharbor/harbor/src/pkg/allowlist/models"
)

const (
	// ProjectTable is the table name for project
	ProjectTable = "project"
	// ProjectPublic means project is public
	ProjectPublic = "public"
	// ProjectPrivate means project is private
	ProjectPrivate = "private"
)

func init() {
	orm.RegisterModel(&Project{})
}

// Project holds the details of a project.
type Project struct {
	ProjectID    int64                  `orm:"pk;auto;column(project_id)" json:"project_id"`
	OwnerID      int                    `orm:"column(owner_id)" json:"owner_id"`
	Name         string                 `orm:"column(name)" json:"name" sort:"default"`
	CreationTime time.Time              `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time              `orm:"column(update_time);auto_now" json:"update_time"`
	Deleted      bool                   `orm:"column(deleted)" json:"deleted"`
	OwnerName    string                 `orm:"-" json:"owner_name"`
	Role         int                    `orm:"-" json:"current_user_role_id"`
	RoleList     []int                  `orm:"-" json:"current_user_role_ids"`
	RepoCount    int64                  `orm:"-" json:"repo_count"`
	ChartCount   uint64                 `orm:"-" json:"chart_count"`
	Metadata     map[string]string      `orm:"-" json:"metadata"`
	CVEAllowlist allowlist.CVEAllowlist `orm:"-" json:"cve_allowlist"`
	RegistryID   int64                  `orm:"column(registry_id)" json:"registry_id"`
}

// NamesQuery ...
type NamesQuery struct {
	Names      []string // the names of project
	WithPublic bool     // include the public projects
}

// GetMetadata ...
func (p *Project) GetMetadata(key string) (string, bool) {
	if len(p.Metadata) == 0 {
		return "", false
	}
	value, exist := p.Metadata[key]
	return value, exist
}

// SetMetadata ...
func (p *Project) SetMetadata(key, value string) {
	if p.Metadata == nil {
		p.Metadata = map[string]string{}
	}
	p.Metadata[key] = value
}

// IsPublic ...
func (p *Project) IsPublic() bool {
	public, exist := p.GetMetadata(ProMetaPublic)
	if !exist {
		return false
	}

	return isTrue(public)
}

// IsProxy returns true when the project type is proxy cache
func (p *Project) IsProxy() bool {
	return p.RegistryID > 0
}

// ContentTrustEnabled ...
func (p *Project) ContentTrustEnabled() bool {
	enabled, exist := p.GetMetadata(ProMetaEnableContentTrust)
	if !exist {
		return false
	}
	return isTrue(enabled)
}

// VulPrevented ...
func (p *Project) ContentTrustCosignEnabled() bool {
	enabled, exist := p.GetMetadata(ProMetaEnableContentTrustCosign)
	if !exist {
		return false
	}
	return isTrue(enabled)
}

// VulPrevented ...
func (p *Project) VulPrevented() bool {
	prevent, exist := p.GetMetadata(ProMetaPreventVul)
	if !exist {
		return false
	}
	return isTrue(prevent)
}

// ReuseSysCVEAllowlist ...
func (p *Project) ReuseSysCVEAllowlist() bool {
	r, ok := p.GetMetadata(ProMetaReuseSysCVEAllowlist)
	if !ok {
		return true
	}
	return isTrue(r)
}

// Severity ...
func (p *Project) Severity() string {
	severity, exist := p.GetMetadata(ProMetaSeverity)
	if !exist {
		return ""
	}
	return severity
}

// AutoScan ...
func (p *Project) AutoScan() bool {
	auto, exist := p.GetMetadata(ProMetaAutoScan)
	if !exist {
		return false
	}
	return isTrue(auto)
}

// FilterByPublic returns orm.QuerySeter with public filter
func (p *Project) FilterByPublic(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	subQuery := `SELECT project_id FROM project_metadata WHERE name = 'public' AND value = '%s'`
	if isTrue(value) {
		subQuery = fmt.Sprintf(subQuery, "true")
	} else {
		subQuery = fmt.Sprintf(subQuery, "false")
	}
	return qs.FilterRaw("project_id", fmt.Sprintf("IN (%s)", subQuery))
}

// FilterByOwner returns orm.QuerySeter with owner filter
func (p *Project) FilterByOwner(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	username, ok := value.(string)
	if !ok {
		return qs
	}

	return qs.FilterRaw("owner_id", fmt.Sprintf("IN (SELECT user_id FROM harbor_user WHERE username = %s)", orm.QuoteLiteral(username)))
}

// FilterByMember returns orm.QuerySeter with member filter
func (p *Project) FilterByMember(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	query, ok := value.(*MemberQuery)
	if !ok {
		return qs
	}
	subQuery := fmt.Sprintf(`SELECT project_id FROM project_member WHERE entity_id = %d AND entity_type = 'u'`, query.UserID)
	if query.Role > 0 {
		subQuery = fmt.Sprintf("%s AND role = %d", subQuery, query.Role)
	}

	if query.WithPublic {
		subQuery = fmt.Sprintf("(%s) UNION (SELECT project_id FROM project_metadata WHERE name = 'public' AND value = 'true')", subQuery)
	}

	if len(query.GroupIDs) > 0 {
		var elems []string
		for _, groupID := range query.GroupIDs {
			elems = append(elems, strconv.Itoa(groupID))
		}

		tpl := "(%s) UNION (SELECT project_id FROM project_member pm, user_group ug WHERE pm.entity_id = ug.id AND pm.entity_type = 'g' AND ug.id IN (%s))"
		subQuery = fmt.Sprintf(tpl, subQuery, strings.TrimSpace(strings.Join(elems, ", ")))
	}

	return qs.FilterRaw("project_id", fmt.Sprintf("IN (%s)", subQuery))
}

// FilterByNames returns orm.QuerySeter with name filter
func (p *Project) FilterByNames(ctx context.Context, qs orm.QuerySeter, key string, value interface{}) orm.QuerySeter {
	query, ok := value.(*NamesQuery)
	if !ok {
		return qs
	}

	if len(query.Names) == 0 {
		return qs
	}

	var names []string
	for _, v := range query.Names {
		names = append(names, `'`+v+`'`)
	}
	subQuery := fmt.Sprintf("SELECT project_id FROM project where name IN (%s)", strings.Join(names, ","))

	if query.WithPublic {
		subQuery = fmt.Sprintf("(%s) UNION (SELECT project_id FROM project_metadata WHERE name = 'public' AND value = 'true')", subQuery)
	}

	return qs.FilterRaw("project_id", fmt.Sprintf("IN (%s)", subQuery))
}

func isTrue(i interface{}) bool {
	switch value := i.(type) {
	case bool:
		return value
	case string:
		v := strings.ToLower(value)
		return v == "true" || v == "1"
	default:
		return false
	}
}

// TableName is required by beego orm to map Project to table project
func (p *Project) TableName() string {
	return ProjectTable
}

// Projects the connection for Project
type Projects []*Project

// OwnerIDs returns all the owner ids from the projects
func (projects Projects) OwnerIDs() []int {
	var ownerIDs []int
	for _, project := range projects {
		ownerIDs = append(ownerIDs, project.OwnerID)
	}
	return ownerIDs
}
