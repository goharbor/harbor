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

package api

import (
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

const (
	// PriPC : count of private projects
	PriPC = "private_project_count"
	// PriRC : count of private repositories
	PriRC = "private_repo_count"
	// PubPC : count of public projects
	PubPC = "public_project_count"
	// PubRC : count of public repositories
	PubRC = "public_repo_count"
	// TPC : total count of projects
	TPC = "total_project_count"
	// TRC : total count of repositories
	TRC = "total_repo_count"
)

// StatisticAPI handles request to /api/statistics/
type StatisticAPI struct {
	BaseController
	username string
}

// Prepare validates the URL and the user
func (s *StatisticAPI) Prepare() {
	s.BaseController.Prepare()
	if !s.SecurityCtx.IsAuthenticated() {
		s.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	s.username = s.SecurityCtx.GetUsername()
}

// Get total projects and repos of the user
func (s *StatisticAPI) Get() {
	statistic := map[string]int64{}
	pubProjs, err := s.ProjectMgr.GetPublic()
	if err != nil {
		s.ParseAndHandleError("failed to get public projects", err)
		return
	}

	statistic[PubPC] = (int64)(len(pubProjs))
	if len(pubProjs) == 0 {
		statistic[PubRC] = 0
	} else {
		ids := make([]int64, 0)
		for _, p := range pubProjs {
			ids = append(ids, p.ProjectID)
		}
		n, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
			ProjectIDs: ids,
		})
		if err != nil {
			log.Errorf("failed to get total of public repositories: %v", err)
			s.SendInternalServerError(fmt.Errorf("failed to get total of public repositories: %v", err))
			return
		}
		statistic[PubRC] = n
	}

	if s.SecurityCtx.IsSysAdmin() {
		result, err := s.ProjectMgr.List(nil)
		if err != nil {
			log.Errorf("failed to get total of projects: %v", err)
			s.SendInternalServerError(fmt.Errorf("failed to get total of projects: %v", err))
			return
		}
		statistic[TPC] = result.Total
		statistic[PriPC] = result.Total - statistic[PubPC]

		n, err := dao.GetTotalOfRepositories()
		if err != nil {
			log.Errorf("failed to get total of repositories: %v", err)
			s.SendInternalServerError(fmt.Errorf("failed to get total of repositories: %v", err))
			return
		}
		statistic[TRC] = n
		statistic[PriRC] = n - statistic[PubRC]
	} else {
		// including the public ones
		myProjects, err := s.SecurityCtx.GetMyProjects()
		privProjectIDs := make([]int64, 0)
		if err != nil {
			s.ParseAndHandleError(fmt.Sprintf(
				"failed to get projects of user %s", s.username), err)
			return
		}
		for _, p := range myProjects {
			if !p.IsPublic() {
				privProjectIDs = append(privProjectIDs, p.ProjectID)
			}
		}

		statistic[PriPC] = int64(len(privProjectIDs))
		if statistic[PriPC] == 0 {
			statistic[PriRC] = 0
		} else {
			n, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
				ProjectIDs: privProjectIDs,
			})
			if err != nil {
				s.SendInternalServerError(fmt.Errorf(
					"failed to get total of repositories for user %s: %v",
					s.username, err))
				return
			}
			statistic[PriRC] = n
		}
	}

	s.Data["json"] = statistic
	s.ServeJSON()
}
