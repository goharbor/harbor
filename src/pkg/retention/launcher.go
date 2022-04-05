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
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/task"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/selector/selectors/index"

	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	pq "github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"
)

const (
	// ParamRepo ...
	ParamRepo = "repository"
	// ParamMeta ...
	ParamMeta = "liteMeta"
	// ParamDryRun ...
	ParamDryRun = "dryRun"
)

// Launcher provides function to launch the async jobs to run retentions based on the provided policy.
type Launcher interface {
	// Launch async jobs for the retention policy
	// A separate job will be launched for each repository
	//
	//  Arguments:
	//   policy *policy.Metadata: the policy info
	//   executionID int64      : the execution ID
	//   isDryRun bool          : indicate if it is a dry run
	//
	//  Returns:
	//   int64               : the count of tasks
	//   error               : common error if any errors occurred
	Launch(ctx context.Context, policy *policy.Metadata, executionID int64, isDryRun bool) (int64, error)
	// Stop the jobs for one execution
	//
	//  Arguments:
	//   executionID int64 : the execution ID
	//
	//  Returns:
	//   error : common error if any errors occurred
	Stop(ctx context.Context, executionID int64) error
}

// NewLauncher returns an instance of Launcher
func NewLauncher(projectMgr project.Manager, repositoryMgr repository.Manager,
	retentionMgr Manager, execMgr task.ExecutionManager, taskMgr task.Manager) Launcher {
	return &launcher{
		projectMgr:       projectMgr,
		repositoryMgr:    repositoryMgr,
		retentionMgr:     retentionMgr,
		execMgr:          execMgr,
		taskMgr:          taskMgr,
		jobserviceClient: cjob.GlobalClient,
	}
}

type jobData struct {
	TaskID     int64
	Repository selector.Repository
	JobName    string
	JobParams  map[string]interface{}
}

type launcher struct {
	retentionMgr     Manager
	taskMgr          task.Manager
	execMgr          task.ExecutionManager
	projectMgr       project.Manager
	repositoryMgr    repository.Manager
	jobserviceClient cjob.Client
}

func (l *launcher) Launch(ctx context.Context, ply *policy.Metadata, executionID int64, isDryRun bool) (int64, error) {
	if ply == nil {
		return 0, launcherError(fmt.Errorf("the policy is nil"))
	}
	// no rules, return directly
	if len(ply.Rules) == 0 {
		log.Debugf("no rules for policy %d, skip", ply.ID)
		return 0, nil
	}
	scope := ply.Scope
	if scope == nil {
		return 0, launcherError(fmt.Errorf("the scope of policy is nil"))
	}
	repositoryRules := make(map[selector.Repository]*lwp.Metadata, 0)
	level := scope.Level
	var allProjects []*selector.Candidate
	var err error
	if level == "system" {
		// get projects
		allProjects, err = getProjects(ctx, l.projectMgr)
		if err != nil {
			return 0, launcherError(err)
		}
	}

	for _, rule := range ply.Rules {
		if rule.Disabled {
			log.Infof("Policy %d rule %d %s is deactivated", ply.ID, rule.ID, rule.Template)
			continue
		}
		projectCandidates := allProjects
		switch level {
		case "system":
			// filter projects according to the project selectors
			for _, projectSelector := range rule.ScopeSelectors["project"] {
				selector, err := index.Get(projectSelector.Kind, projectSelector.Decoration,
					projectSelector.Pattern, "")
				if err != nil {
					return 0, launcherError(err)
				}
				projectCandidates, err = selector.Select(projectCandidates)
				if err != nil {
					return 0, launcherError(err)
				}
			}
		case "project":
			projectCandidates = append(projectCandidates, &selector.Candidate{
				NamespaceID: scope.Reference,
			})
		}

		var repositoryCandidates []*selector.Candidate
		// get repositories of projects
		for _, projectCandidate := range projectCandidates {
			repositories, err := getRepositories(ctx, l.projectMgr, l.repositoryMgr, projectCandidate.NamespaceID)
			if err != nil {
				return 0, launcherError(err)
			}
			for _, repository := range repositories {
				repositoryCandidates = append(repositoryCandidates, repository)
			}
		}
		// filter repositories according to the repository selectors
		for _, repositorySelector := range rule.ScopeSelectors["repository"] {
			selector, err := index.Get(repositorySelector.Kind, repositorySelector.Decoration,
				repositorySelector.Pattern, repositorySelector.Extras)
			if err != nil {
				return 0, launcherError(err)
			}
			repositoryCandidates, err = selector.Select(repositoryCandidates)
			if err != nil {
				return 0, launcherError(err)
			}
		}

		for _, repositoryCandidate := range repositoryCandidates {
			reposit := selector.Repository{
				NamespaceID: repositoryCandidate.NamespaceID,
				Namespace:   repositoryCandidate.Namespace,
				Name:        repositoryCandidate.Repository,
				Kind:        repositoryCandidate.Kind,
			}
			if repositoryRules[reposit] == nil {
				repositoryRules[reposit] = &lwp.Metadata{
					Algorithm: ply.Algorithm,
				}
			}
			r := rule
			repositoryRules[reposit].Rules = append(repositoryRules[reposit].Rules, &r)
		}
	}

	// create job data list
	jobDatas, err := createJobs(repositoryRules, isDryRun)
	if err != nil {
		return 0, launcherError(err)
	}

	// no jobs, return directly
	if len(jobDatas) == 0 {
		log.Debugf("no candidates for policy %d, skip", ply.ID)
		return 0, nil
	}

	// submit tasks to jobservice
	if err = l.submitTasks(ctx, executionID, jobDatas); err != nil {
		return 0, launcherError(err)
	}

	return int64(len(jobDatas)), nil
}

func createJobs(repositoryRules map[selector.Repository]*lwp.Metadata, isDryRun bool) ([]*jobData, error) {
	jobDatas := []*jobData{}
	for repository, policy := range repositoryRules {
		jobData := &jobData{
			Repository: repository,
			JobName:    job.Retention,
			JobParams:  make(map[string]interface{}, 3),
		}
		// set dry run
		jobData.JobParams[ParamDryRun] = isDryRun
		// set repository
		repoJSON, err := repository.ToJSON()
		if err != nil {
			return nil, err
		}
		jobData.JobParams[ParamRepo] = repoJSON
		// set retention policy
		policyJSON, err := policy.ToJSON()
		if err != nil {
			return nil, err
		}
		jobData.JobParams[ParamMeta] = policyJSON
		jobDatas = append(jobDatas, jobData)
	}
	return jobDatas, nil
}

func (l *launcher) submitTasks(ctx context.Context, executionID int64, jobDatas []*jobData) error {
	for _, jobData := range jobDatas {
		_, err := l.taskMgr.Create(ctx, executionID, &task.Job{
			Name:       jobData.JobName,
			Parameters: jobData.JobParams,
			Metadata: &job.Metadata{
				JobKind: job.KindGeneric,
			},
		},
			map[string]interface{}{
				"repository": jobData.Repository.Name,
				"dry_run":    jobData.JobParams[ParamDryRun],
			})
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *launcher) Stop(ctx context.Context, executionID int64) error {
	if executionID <= 0 {
		return launcherError(fmt.Errorf("invalid execution ID: %d", executionID))
	}
	return l.execMgr.Stop(ctx, executionID)
}

func launcherError(err error) error {
	return errors.Wrap(err, "launcher")
}

func getProjects(ctx context.Context, projectMgr project.Manager) ([]*selector.Candidate, error) {
	projects, err := projectMgr.List(ctx, nil)
	if err != nil {
		return nil, err
	}
	var candidates []*selector.Candidate
	for _, pro := range projects {
		candidates = append(candidates, &selector.Candidate{
			NamespaceID: pro.ProjectID,
			Namespace:   pro.Name,
		})
	}
	return candidates, nil
}

func getRepositories(ctx context.Context, projectMgr project.Manager, repositoryMgr repository.Manager, projectID int64) ([]*selector.Candidate, error) {
	var candidates []*selector.Candidate
	/*
		pro, err := projectMgr.Get(projectID)
		if err != nil {
			return nil, err
		}
	*/
	// get image repositories
	imageRepositories, err := repositoryMgr.List(ctx, &pq.Query{
		Keywords: map[string]interface{}{
			"ProjectID": projectID,
		},
	})
	if err != nil {
		return nil, err
	}
	for _, r := range imageRepositories {
		namespace, repo := utils.ParseRepository(r.Name)
		candidates = append(candidates, &selector.Candidate{
			NamespaceID: projectID,
			Namespace:   namespace,
			Repository:  repo,
			Kind:        "image",
		})
	}

	return candidates, nil
}
