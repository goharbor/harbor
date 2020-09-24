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

package model

import (
	"strings"

	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// Project model
type Project struct {
	*project.Project
}

// ToSwagger converts the project to the swagger model
func (p *Project) ToSwagger() *models.Project {
	var currentUserRoleIds []int32
	for _, role := range p.RoleList {
		currentUserRoleIds = append(currentUserRoleIds, int32(role))
	}

	var md *models.ProjectMetadata
	if p.Metadata != nil {
		var m models.ProjectMetadata
		lib.JSONCopy(&m, p.Metadata)

		// Transform the severity to severity of CVSS v3.0 Ratings
		if m.Severity != nil {
			severity := strings.ToLower(vuln.ParseSeverityVersion3(*m.Severity).String())
			m.Severity = &severity
		}

		md = &m
	}

	var allowlist models.CVEAllowlist
	if err := lib.JSONCopy(&allowlist, p.CVEAllowlist); err != nil {
		log.Warningf("failed to copy CVEAllowlist form %T", p.CVEAllowlist)
	}

	return &models.Project{
		ChartCount:         int64(p.ChartCount),
		CreationTime:       strfmt.DateTime(p.CreationTime),
		CurrentUserRoleID:  int64(p.Role),
		CurrentUserRoleIds: currentUserRoleIds,
		CVEAllowlist:       &allowlist,
		Metadata:           md,
		Name:               p.Name,
		OwnerID:            int32(p.OwnerID),
		OwnerName:          p.OwnerName,
		ProjectID:          int32(p.ProjectID),
		RegistryID:         p.RegistryID,
		RepoCount:          p.RepoCount,
		UpdateTime:         strfmt.DateTime(p.UpdateTime),
	}
}

// NewProject ...
func NewProject(p *project.Project) *Project {
	return &Project{p}
}
