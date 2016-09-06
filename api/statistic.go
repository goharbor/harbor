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

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

const (
	// MPC : count of my projects
	MPC = "my_project_count"
	// MRC : count of my repositories
	MRC = "my_repo_count"
	// PPC : count of public projects
	PPC = "public_project_count"
	// PRC : count of public repositories
	PRC = "public_repo_count"
	// TPC : total count of projects
	TPC = "total_project_count"
	// TRC : total count of repositories
	TRC = "total_repo_count"
)

// StatisticAPI handles request to /api/statistics/
type StatisticAPI struct {
	BaseAPI
	userID int
}

//Prepare validates the URL and the user
func (s *StatisticAPI) Prepare() {
	s.userID = s.ValidateUser()
}

// Get total projects and repos of the user
func (s *StatisticAPI) Get() {
	statistic := map[string]int64{}

	n, err := dao.GetTotalOfProjects("", 1)
	if err != nil {
		log.Errorf("failed to get total of public projects: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "")
	}
	statistic[PPC] = n

	n, err = dao.GetTotalOfPublicRepositories("")
	if err != nil {
		log.Errorf("failed to get total of public repositories: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "")
	}
	statistic[PRC] = n

	isAdmin, err := dao.IsAdminRole(s.userID)
	if err != nil {
		log.Errorf("Error occured in check admin, error: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}

	if isAdmin {
		n, err := dao.GetTotalOfProjects("")
		if err != nil {
			log.Errorf("failed to get total of projects: %v", err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}
		statistic[MPC] = n
		statistic[TPC] = n

		n, err = dao.GetTotalOfRepositories("")
		if err != nil {
			log.Errorf("failed to get total of repositories: %v", err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}
		statistic[MRC] = n
		statistic[TRC] = n
	} else {
		n, err := dao.GetTotalOfUserRelevantProjects(s.userID, "")
		if err != nil {
			log.Errorf("failed to get total of projects for user %d: %v", s.userID, err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}
		statistic[MPC] = n

		n, err = dao.GetTotalOfUserRelevantRepositories(s.userID, "")
		if err != nil {
			log.Errorf("failed to get total of repositories for user %d: %v", s.userID, err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}
		statistic[MRC] = n
	}

	s.Data["json"] = statistic
	s.ServeJSON()
}
