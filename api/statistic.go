/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package api

import (
	"net/http"
	"os"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry"
	"github.com/vmware/harbor/utils/registry/auth"
)

type StatisticAPI struct {
	BaseAPI
	userID   int
	username string
	registry *registry.Registry
}

//Prepare validates the URL and the user
func (s *StatisticAPI) Prepare() {
	userID, ok := s.GetSession("userId").(int)
	if !ok {
		s.userID = dao.NonExistUserID
	} else {
		s.userID = userID
		log.Debug("userID is xxx", userID)
	}
	username, ok := s.GetSession("username").(string)
	if !ok {
		log.Warning("failed to get username from session")
		s.username = ""
	} else {
		s.username = username
		log.Debug("username is xxx", username)
	}

	var client *http.Client

	//no session, initialize a standard auth handler
	if s.userID == dao.NonExistUserID && len(s.username) == 0 {
		username, password, _ := s.Ctx.Request.BasicAuth()

		credential := auth.NewBasicAuthCredential(username, password)
		client = registry.NewClientStandardAuthHandlerEmbeded(credential)
		log.Debug("initializing standard auth handler")

	} else {
		// session works, initialize a username auth handler
		username := s.username
		if len(username) == 0 {
			user, err := dao.GetUser(models.User{
				UserID: s.userID,
			})
			if err != nil {
				log.Errorf("error occurred whiling geting user for initializing a username auth handler: %v", err)
				return
			}

			username = user.Username
		}

		client = registry.NewClientUsernameAuthHandlerEmbeded(username)
		log.Debug("initializing username auth handler: %s", username)
	}

	endpoint := os.Getenv("REGISTRY_URL")
	r, err := registry.New(endpoint, client)
	if err != nil {
		log.Fatalf("error occurred while initializing auth handler for repository API: %v", err)
	}

	s.registry = r
}

// Get total projects and repos of the user
func (s *StatisticAPI) Get() {
	queryProject := models.Project{UserID: s.userID}
	projectList, err := dao.QueryProject(queryProject)
	var projectArr [6]int
	if err != nil {
		log.Errorf("Error occured in QueryProject, error: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	log.Debug("projectList xxx ", projectList)
	isAdmin, _ := dao.IsAdminRole(s.userID)
	for i := 0; i < len(projectList); i++ {
		if isProjectAdmin(s.userID, projectList[i].ProjectID) {
			projectArr[0] += 1
			projectArr[1] += s.GetRepos(projectList[i].ProjectID)
		}
		if projectList[i].Public == 1 {
			projectArr[2] += 1
			projectArr[3] += s.GetRepos(projectList[i].ProjectID)
		}
		if isAdmin {
			projectArr[5] += s.GetRepos(projectList[i].ProjectID)
		}
	}
	if isAdmin {
		projectArr[4] = len(projectList)
	}
	s.Data["json"] = projectArr
	s.ServeJSON()
}

func (s *StatisticAPI) GetRepos(projectID int64) int {
	p, err := dao.GetProjectByID(projectID)
	if err != nil {
		log.Errorf("Error occurred in GetProjectById, error: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	if p == nil {
		log.Warningf("Project with Id: %d does not exist", projectID)
		s.RenderError(http.StatusNotFound, "")
		return 0
	}
	if p.Public == 0 && !checkProjectPermission(s.userID, projectID) {
		s.RenderError(http.StatusForbidden, "")
		return 0
	}

	repoList, err := svc_utils.GetRepoFromCache()
	if err != nil {
		log.Errorf("Failed to get repo from cache, error: %v", err)
		s.RenderError(http.StatusInternalServerError, "internal sever error")
	}
	return len(repoList)
}
