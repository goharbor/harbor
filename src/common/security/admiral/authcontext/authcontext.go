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

package authcontext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/common"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
)

const (
	// AuthTokenHeader is the key of auth token header
	AuthTokenHeader  = "x-xenon-auth-token"
	sysAdminRole     = "CLOUD_ADMIN"
	projectAdminRole = "PROJECT_ADMIN"
	developerRole    = "PROJECT_MEMBER"
	guestRole        = "PROJECT_VIEWER"
)

type project struct {
	SelfLink   string            `json:"documentSelfLink"`
	Name       string            `json:"name"`
	Roles      []string          `json:"roles"`
	Properties map[string]string `json:"customProperties"`
}

// AuthContext ...
type AuthContext struct {
	PrincipalID string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Roles       []string   `json:"roles"`
	Projects    []*project `json:"projects"`
}

// IsSysAdmin ...
func (a *AuthContext) IsSysAdmin() bool {
	for _, role := range a.Roles {
		if role == sysAdminRole {
			return true
		}
	}
	return false
}

// GetProjectRoles ...
func (a *AuthContext) GetProjectRoles(projectIDOrName interface{}) []int {
	id, name, err := utils.ParseProjectIDOrName(projectIDOrName)
	if err != nil {
		log.Errorf("failed to parse project ID or name: %v", err)
		return []int{}
	}

	roles := []string{}
	for _, project := range a.Projects {
		p := convertProject(project)
		if id != 0 && p.ProjectID == id || len(name) > 0 && p.Name == name {
			roles = append(roles, project.Roles...)
			break
		}
	}

	return convertRoles(roles)
}

// GetMyProjects returns all projects which the user is a member of
func (a *AuthContext) GetMyProjects() []*models.Project {
	projects := []*models.Project{}
	for _, project := range a.Projects {
		projects = append(projects, convertProject(project))
	}
	return projects
}

// convert project returned by Admiral to project used in Harbor
func convertProject(p *project) *models.Project {
	project := &models.Project{
		Name: p.Name,
	}

	index := ""
	if p.Properties != nil {
		index = p.Properties["__projectIndex"]
	}

	if len(index) == 0 {
		log.Errorf("property __projectIndex not found when parsing project")
		return project
	}

	id, err := strconv.ParseInt(index, 10, 64)
	if err != nil {
		log.Errorf("failed to parse __projectIndex %s: %v", index, err)
		return project
	}

	project.ProjectID = id
	return project
}

// convert roles defined by Admiral to roles used in Harbor
func convertRoles(roles []string) []int {
	list := []int{}
	for _, role := range roles {
		switch role {
		case projectAdminRole:
			list = append(list, common.RoleProjectAdmin)
		case developerRole:
			list = append(list, common.RoleDeveloper)
		case guestRole:
			list = append(list, common.RoleGuest)
		default:
			log.Warningf("unknown role: %s", role)
		}
	}

	return list
}

// GetAuthCtx returns the auth context of the current user
func GetAuthCtx(client *http.Client, url, token string) (*AuthContext, error) {
	req, err := http.NewRequest(http.MethodGet, buildCurrentUserAuthCtxURL(url), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(AuthTokenHeader, token)

	return send(client, req)
}

// Login with credential and returns auth context and error
func Login(client *http.Client, url, username, password, token string) (*AuthContext, error) {
	data, err := json.Marshal(&struct {
		Password string `json:"password"`
	}{
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, buildLoginURL(url, username), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	req.Header.Add(AuthTokenHeader, token)

	return send(client, req)
}

func send(client *http.Client, req *http.Request) (*AuthContext, error) {
	resp, err := client.Do(req)
	if err != nil {
		log.Debugf("\"%s %s\" failed", req.Method, req.URL.String())
		return nil, err
	}
	defer resp.Body.Close()
	log.Debugf("\"%s %s\" %d", req.Method, req.URL.String(), resp.StatusCode)

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &commonhttp.Error{
			Code:    resp.StatusCode,
			Message: string(data),
		}
	}

	ctx := &AuthContext{}
	if err = json.Unmarshal(data, ctx); err != nil {
		return nil, err
	}

	return ctx, nil
}

func buildCurrentUserAuthCtxURL(url string) string {
	return strings.TrimRight(url, "/") + "/auth/session"
}

func buildLoginURL(url, principalID string) string {
	return fmt.Sprintf("%s/auth/idm/principals/%s/security-context",
		strings.TrimRight(url, "/"), principalID)
}
