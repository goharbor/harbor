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

package retention

import (
	"fmt"

	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors"
	"github.com/pkg/errors"
)

// Launcher provides function to launch the async jobs to run retentions based on the provided policy.
type Launcher interface {
	// Launch async jobs for the retention policy
	// A separate job will be launched for each repository
	//
	//  Arguments:
	//   policy *policy.Metadata: the policy info
	//   executionID int64      : the execution ID
	//
	//  Returns:
	//   int64               : the count of tasks
	//   error               : common error if any errors occurred
	Launch(policy *policy.Metadata, executionID int64) (int64, error)
}

// NewLauncher returns an instance of Launcher
func NewLauncher(projectMgr project.Manager, repositoryMgr repository.Manager,
	retentionMgr Manager, retentionClient dep.Client) Launcher {
	return &launcher{
		projectMgr:      projectMgr,
		repositoryMgr:   repositoryMgr,
		retentionMgr:    retentionMgr,
		retentionClient: retentionClient,
	}
}

type launcher struct {
	retentionMgr    Manager
	retentionClient dep.Client
	projectMgr      project.Manager
	repositoryMgr   repository.Manager
}

type jobData struct {
	repository *res.Repository
	policy     *lwp.Metadata
	taskID     int64
}

func (l *launcher) Launch(ply *policy.Metadata, executionID int64) (int64, error) {
	if ply == nil {
		return 0, launcherError(fmt.Errorf("the policy is nil"))
	}
	// no rules, return directly
	if len(ply.Rules) == 0 {
		return 0, nil
	}
	scope := ply.Scope
	if scope == nil {
		return 0, launcherError(fmt.Errorf("the scope of policy is nil"))
	}

	repositoryRules := make(map[res.Repository]*lwp.Metadata, 0)
	level := scope.Level
	var projectCandidates []*res.Candidate
	var err error
	if level == "system" {
		// get projects
		projectCandidates, err = getProjects(l.projectMgr)
		if err != nil {
			return 0, launcherError(err)
		}
	}

	for _, rule := range ply.Rules {
		switch level {
		case "system":
			// filter projects according to the project selectors
			for _, projectSelector := range rule.ScopeSelectors["project"] {
				selector, err := selectors.Get(projectSelector.Kind, projectSelector.Decoration,
					projectSelector.Pattern)
				if err != nil {
					return 0, launcherError(err)
				}
				projectCandidates, err = selector.Select(projectCandidates)
				if err != nil {
					return 0, launcherError(err)
				}
			}
		case "project":
			projectCandidates = append(projectCandidates, &res.Candidate{
				NamespaceID: scope.Reference,
			})
		}

		var repositoryCandidates []*res.Candidate
		// get repositories of projects
		for _, projectCandidate := range projectCandidates {
			repositories, err := getRepositories(l.projectMgr, l.repositoryMgr, projectCandidate.NamespaceID)
			if err != nil {
				return 0, launcherError(err)
			}
			repositoryCandidates = append(repositoryCandidates, repositories...)
		}
		// filter repositories according to the repository selectors
		for _, repositorySelector := range rule.ScopeSelectors["repository"] {
			selector, err := selectors.Get(repositorySelector.Kind, repositorySelector.Decoration,
				repositorySelector.Pattern)
			if err != nil {
				return 0, launcherError(err)
			}
			repositoryCandidates, err = selector.Select(repositoryCandidates)
			if err != nil {
				return 0, launcherError(err)
			}
		}

		for _, repositoryCandidate := range repositoryCandidates {
			reposit := res.Repository{
				Namespace: repositoryCandidate.Namespace,
				Name:      repositoryCandidate.Repository,
				Kind:      repositoryCandidate.Kind,
			}
			if repositoryRules[reposit] == nil {
				repositoryRules[reposit] = &lwp.Metadata{
					Algorithm: ply.Algorithm,
				}
			}
			repositoryRules[reposit].Rules = append(repositoryRules[reposit].Rules, &rule)
		}
	}
	// no tasks need to be submitted
	if len(repositoryRules) == 0 {
		return 0, nil
	}

	// create task records
	jobDatas := make([]*jobData, 0)
	for repo, p := range repositoryRules {
		taskID, err := l.retentionMgr.CreateTask(&Task{
			ExecutionID: executionID,
		})
		if err != nil {
			return 0, launcherError(err)
		}
		jobDatas = append(jobDatas, &jobData{
			repository: &repo,
			policy:     p,
			taskID:     taskID,
		})
	}

	allFailed := true
	for _, jobData := range jobDatas {
		_, err := l.retentionClient.SubmitTask(jobData.taskID, jobData.repository, jobData.policy)
		if err != nil {
			log.Error(launcherError(fmt.Errorf("failed to submit task %d: %v", jobData.taskID, err)))
			continue
		}
		allFailed = false
	}
	if allFailed {
		return 0, launcherError(fmt.Errorf("all tasks failed"))
	}
	return int64(len(jobDatas)), nil
}

func launcherError(err error) error {
	return errors.Wrap(err, "launcher")
}

func getProjects(projectMgr project.Manager) ([]*res.Candidate, error) {
	projects, err := projectMgr.List()
	if err != nil {
		return nil, err
	}
	var candidates []*res.Candidate
	for _, pro := range projects {
		candidates = append(candidates, &res.Candidate{
			NamespaceID: pro.ProjectID,
			Namespace:   pro.Name,
		})
	}
	return candidates, nil
}

func getRepositories(projectMgr project.Manager, repositoryMgr repository.Manager, projectID int64) ([]*res.Candidate, error) {
	var candidates []*res.Candidate
	pro, err := projectMgr.Get(projectID)
	if err != nil {
		return nil, err
	}
	// get image repositories
	imageRepositories, err := repositoryMgr.ListImageRepositories(projectID)
	if err != nil {
		return nil, err
	}
	for _, r := range imageRepositories {
		namespace, repo := utils.ParseRepository(r.Name)
		candidates = append(candidates, &res.Candidate{
			Namespace:  namespace,
			Repository: repo,
			Kind:       "image",
		})
	}
	// get chart repositories
	chartRepositories, err := repositoryMgr.ListChartRepositories(projectID)
	for _, r := range chartRepositories {
		candidates = append(candidates, &res.Candidate{
			Namespace:  pro.Name,
			Repository: r.Name,
			Kind:       "chart",
		})
	}
	return candidates, nil
}
