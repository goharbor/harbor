// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"strings"
)

// TODO update the value of role when admiral API is ready
const (
	// AuthTokenHeader is the key of auth token header
	AuthTokenHeader  = "x-xenon-auth-token"
	sysAdminRole     = "CLOUD_ADMIN"
	projectAdminRole = "PROJECT_ADMIN"
	developerRole    = "PROJECT_MEMBER"
	guestRole        = "PROJECT_VIEWER"
)

type project struct {
	DocumentSelfLink string   `json:"documentSelfLink"`
	Name             string   `json:"name"`
	Roles            []string `json:"roles"`
}

// AuthContext ...
type AuthContext struct {
	PrincipalID string     `json:"id"`
	Name        string     `json:"name"`
	Email       string     `json:"email"`
	Roles       []string   `json:"roles"`
	Projects    []*project `json:"projects"`
}

// GetUsername ...
func (a *AuthContext) GetUsername() string {
	return a.PrincipalID
}

// IsSysAdmin ...
func (a *AuthContext) IsSysAdmin() bool {
	isSysAdmin := false
	for _, role := range a.Roles {
		if role == sysAdminRole {
			isSysAdmin = true
			break
		}
	}
	return isSysAdmin
}

// HasReadPerm ...
func (a *AuthContext) HasReadPerm(projectName string) bool {
	roles := a.getRoles(projectName)
	return len(roles) > 0
}

// HasWritePerm ...
func (a *AuthContext) HasWritePerm(projectName string) bool {
	roles := a.getRoles(projectName)
	for _, role := range roles {
		if role == projectAdminRole || role == developerRole {
			return true
		}
	}
	return false
}

// HasAllPerm ...
func (a *AuthContext) HasAllPerm(projectName string) bool {
	roles := a.getRoles(projectName)
	for _, role := range roles {
		if role == projectAdminRole {
			return true
		}
	}
	return false
}

func (a *AuthContext) getRoles(projectName string) []string {
	for _, project := range a.Projects {
		if project.Name == projectName {
			return project.Roles
		}
	}

	return []string{}
}

// GetMyProjects returns all projects which the user is a member of
func (a *AuthContext) GetMyProjects() []string {
	projects := []string{}
	for _, project := range a.Projects {
		projects = append(projects, project.Name)
	}
	return projects
}

// GetAuthCtx returns the auth context of the current user
func GetAuthCtx(client *http.Client, url, token string) (*AuthContext, error) {
	return get(client, url, token)
}

// GetAuthCtxOfUser returns the auth context of the specific user
func GetAuthCtxOfUser(client *http.Client, url, token string, username string) (*AuthContext, error) {
	return get(client, url, token, username)
}

// get the user's auth context, if the username is not provided
// get the default auth context of the token
func get(client *http.Client, url, token string, username ...string) (*AuthContext, error) {
	endpoint := ""
	if len(username) > 0 && len(username[0]) > 0 {
		endpoint = buildSpecificUserAuthCtxURL(url, username[0])
	} else {
		endpoint = buildCurrentUserAuthCtxURL(url)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add(AuthTokenHeader, token)

	return send(client, req)
}

// Login with credential and returns auth context and error
func Login(client *http.Client, url, username, password string) (*AuthContext, error) {
	data, err := json.Marshal(&struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, buildLoginURL(url), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	return send(client, req)
}

func send(client *http.Client, req *http.Request) (*AuthContext, error) {
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, string(data))
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

func buildSpecificUserAuthCtxURL(url, principalID string) string {
	return fmt.Sprintf("%s/auth/idm/principals/%s/security-context",
		strings.TrimRight(url, "/"), principalID)
}

// TODO update the url
func buildLoginURL(url string) string {
	return strings.TrimRight(url, "/") + "/sso/login"
}
