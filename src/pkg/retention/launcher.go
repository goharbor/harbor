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

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors"
	"github.com/pkg/errors"
)

// TODO init the client
var client Client

// Launcher provides function to launch the async jobs to run retentions based on the provided policy.
type Launcher interface {
	// Launch async jobs for the retention policy
	// A separate job will be launched for each repository
	//
	//  Arguments:
	//   policy *policy.Metadata: the policy info
	//
	//  Returns:
	//   []*TaskSubmitResult : the submit results of tasks
	//   error               : common error if any errors occurred
	Launch(policy *policy.Metadata) ([]*TaskSubmitResult, error)
}

// NewLauncher returns an instance of Launcher
func NewLauncher() Launcher {
	return &launcher{}
}

type launcher struct {
}

func (l *launcher) Launch(ply *policy.Metadata) ([]*TaskSubmitResult, error) {
	if ply == nil {
		return nil, launcherError(fmt.Errorf("the policy is nil"))
	}
	// no rules, return directly
	if len(ply.Rules) == 0 {
		return nil, nil
	}
	scope := ply.Scope
	if scope == nil {
		return nil, launcherError(fmt.Errorf("the scope of policy is nil"))
	}

	repositoryRules := make(map[res.Repository]*policy.LiteMeta, 0)
	level := scope.Level
	var projectCandidates []*res.Candidate
	var err error
	if level == "system" {
		// get projects
		projectCandidates, err = getProjects()
		if err != nil {
			return nil, launcherError(err)
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
					return nil, launcherError(err)
				}
				projectCandidates, err = selector.Select(projectCandidates)
				if err != nil {
					return nil, launcherError(err)
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
			repositories, err := getRepositories(projectCandidate.NamespaceID)
			if err != nil {
				return nil, launcherError(err)
			}
			repositoryCandidates = append(repositoryCandidates, repositories...)
		}
		// filter repositories according to the repository selectors
		for _, repositorySelector := range rule.ScopeSelectors["repository"] {
			selector, err := selectors.Get(repositorySelector.Kind, repositorySelector.Decoration,
				repositorySelector.Pattern)
			if err != nil {
				return nil, launcherError(err)
			}
			repositoryCandidates, err = selector.Select(repositoryCandidates)
			if err != nil {
				return nil, launcherError(err)
			}
		}

		for _, repositoryCandidate := range repositoryCandidates {
			repository := res.Repository{
				Namespace: repositoryCandidate.Namespace,
				Name:      repositoryCandidate.Repository,
				Kind:      repositoryCandidate.Kind,
			}
			if repositoryRules[repository] == nil {
				repositoryRules[repository] = &policy.LiteMeta{
					Algorithm: ply.Algorithm,
				}
			}
			repositoryRules[repository].Rules = append(repositoryRules[repository].Rules, &rule)
		}
	}

	var result []*TaskSubmitResult
	for repository, rule := range repositoryRules {
		jobID, err := client.SubmitTask(&repository, rule)
		result = append(result, &TaskSubmitResult{
			JobID: jobID,
			Error: err,
		})
		if err != nil {
			log.Error(launcherError(fmt.Errorf("failed to submit task: %v", err)))
		}
	}
	return result, nil
}

func launcherError(err error) error {
	return errors.Wrap(err, "launcher")
}

func getProjects() ([]*res.Candidate, error) {
	projects, err := project.Mgr.List()
	if err != nil {
		return nil, err
	}
	var candidates []*res.Candidate
	for _, project := range projects {
		candidates = append(candidates, &res.Candidate{
			NamespaceID: project.ProjectID,
			Namespace:   project.Name,
		})
	}
	return candidates, nil
}

func getRepositories(projectID int64) ([]*res.Candidate, error) {
	var candidates []*res.Candidate
	project, err := project.Mgr.Get(projectID)
	if err != nil {
		return nil, err
	}
	// get image repositories
	imageRepositories, err := repository.Mgr.ListImageRepositories(projectID)
	if err != nil {
		return nil, err
	}
	for _, repository := range imageRepositories {
		namespace, repo := utils.ParseRepository(repository.Name)
		candidates = append(candidates, &res.Candidate{
			Namespace:  namespace,
			Repository: repo,
			Kind:       "image",
		})
	}
	// get chart repositories
	chartRepositories, err := repository.Mgr.ListChartRepositories(projectID)
	for _, repository := range chartRepositories {
		candidates = append(candidates, &res.Candidate{
			Namespace:  project.Name,
			Repository: repository.Name,
			Kind:       "chart",
		})
	}
	return candidates, nil
}
