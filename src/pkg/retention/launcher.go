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
	"time"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/retention/res/selectors/index"

	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	cmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/lwp"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/pkg/errors"
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
	Launch(policy *policy.Metadata, executionID int64, isDryRun bool) (int64, error)
	// Stop the jobs for one execution
	//
	//  Arguments:
	//   executionID int64 : the execution ID
	//
	//  Returns:
	//   error : common error if any errors occurred
	Stop(executionID int64) error
}

// NewLauncher returns an instance of Launcher
func NewLauncher(projectMgr project.Manager, repositoryMgr repository.Manager,
	retentionMgr Manager) Launcher {
	return &launcher{
		projectMgr:         projectMgr,
		repositoryMgr:      repositoryMgr,
		retentionMgr:       retentionMgr,
		jobserviceClient:   cjob.GlobalClient,
		internalCoreURL:    config.InternalCoreURL(),
		chartServerEnabled: config.WithChartMuseum(),
	}
}

type jobData struct {
	TaskID     int64
	Repository res.Repository
	JobName    string
	JobParams  map[string]interface{}
}

type launcher struct {
	retentionMgr       Manager
	projectMgr         project.Manager
	repositoryMgr      repository.Manager
	jobserviceClient   cjob.Client
	internalCoreURL    string
	chartServerEnabled bool
}

func (l *launcher) Launch(ply *policy.Metadata, executionID int64, isDryRun bool) (int64, error) {
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
	allRepositories := make(map[res.Repository]struct{}, 0)
	repositoryRules := make(map[res.Repository]*lwp.Metadata, 0)
	level := scope.Level
	var allProjects []*res.Candidate
	var err error
	if level == "system" {
		// get projects
		allProjects, err = getProjects(l.projectMgr)
		if err != nil {
			return 0, launcherError(err)
		}
	}

	for _, rule := range ply.Rules {
		if rule.Disabled {
			log.Infof("Policy %d rule %d %s is disabled", ply.ID, rule.ID, rule.Template)
			continue
		}
		projectCandidates := allProjects
		switch level {
		case "system":
			// filter projects according to the project selectors
			for _, projectSelector := range rule.ScopeSelectors["project"] {
				selector, err := index.Get(projectSelector.Kind, projectSelector.Decoration,
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
			repositories, err := getRepositories(l.projectMgr, l.repositoryMgr, projectCandidate.NamespaceID, l.chartServerEnabled)
			if err != nil {
				return 0, launcherError(err)
			}
			for _, repository := range repositories {
				repo := res.Repository{
					Namespace: repository.Namespace,
					Name:      repository.Repository,
					Kind:      repository.Kind,
				}
				allRepositories[repo] = struct{}{}
				repositoryCandidates = append(repositoryCandidates, repository)
			}
		}
		// filter repositories according to the repository selectors
		for _, repositorySelector := range rule.ScopeSelectors["repository"] {
			selector, err := index.Get(repositorySelector.Kind, repositorySelector.Decoration,
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
			r := rule
			repositoryRules[reposit].Rules = append(repositoryRules[reposit].Rules, &r)
		}
	}

	// create job data list
	jobDatas, err := createJobs(allRepositories, repositoryRules, isDryRun)
	if err != nil {
		return 0, launcherError(err)
	}

	// no jobs, return directly
	if len(jobDatas) == 0 {
		log.Debugf("no candidates for policy %d, skip", ply.ID)
		return 0, nil
	}

	// create task records in database
	if err = l.createTasks(executionID, jobDatas); err != nil {
		return 0, launcherError(err)
	}

	// submit jobs to jobservice
	if err = l.submitJobs(jobDatas); err != nil {
		return 0, launcherError(err)
	}

	return int64(len(jobDatas)), nil
}

func createJobs(allRepositories map[res.Repository]struct{},
	repositoryRules map[res.Repository]*lwp.Metadata, isDryRun bool) ([]*jobData, error) {
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
	for repository := range allRepositories {
		if _, exist := repositoryRules[repository]; exist {
			continue
		}
		jobData := &jobData{
			Repository: repository,
			JobName:    job.RetentionDel,
			JobParams:  make(map[string]interface{}, 2),
		}
		// set dry run
		jobData.JobParams[ParamDryRun] = isDryRun
		// set repository
		repoJSON, err := repository.ToJSON()
		if err != nil {
			return nil, err
		}
		jobData.JobParams[ParamRepo] = repoJSON
		jobDatas = append(jobDatas, jobData)
	}
	return jobDatas, nil
}

// create task records in database
func (l *launcher) createTasks(executionID int64, jobDatas []*jobData) error {
	now := time.Now()
	for _, jobData := range jobDatas {
		taskID, err := l.retentionMgr.CreateTask(&Task{
			ExecutionID: executionID,
			Repository:  jobData.Repository.Name,
			StartTime:   now,
		})
		if err != nil {
			return err
		}
		jobData.TaskID = taskID
	}
	return nil
}

// create task records in database
func (l *launcher) submitJobs(jobDatas []*jobData) error {
	allFailed := true
	for _, jobData := range jobDatas {
		task := &Task{
			ID: jobData.TaskID,
		}
		props := []string{"Status"}
		j := &models.JobData{
			Name: jobData.JobName,
			Metadata: &models.JobMetadata{
				JobKind: job.KindGeneric,
			},
			StatusHook: fmt.Sprintf("%s/service/notifications/jobs/retention/task/%d", l.internalCoreURL, jobData.TaskID),
			Parameters: jobData.JobParams,
		}
		// Submit job
		jobID, err := l.jobserviceClient.SubmitJob(j)
		if err != nil {
			log.Error(launcherError(fmt.Errorf("failed to submit task %d: %v", jobData.TaskID, err)))
			task.Status = cmodels.JobError
			task.EndTime = time.Now()
			props = append(props, "EndTime")
		} else {
			allFailed = false
			task.JobID = jobID
			task.Status = cmodels.JobPending
			props = append(props, "JobID")
		}
		if err = l.retentionMgr.UpdateTask(task, props...); err != nil {
			log.Errorf("failed to update the status of task %d: %v", task.ID, err)
		}
	}
	if allFailed {
		return launcherError(fmt.Errorf("all tasks failed"))
	}
	return nil
}

func (l *launcher) Stop(executionID int64) error {
	if executionID <= 0 {
		return launcherError(fmt.Errorf("invalid execution ID: %d", executionID))
	}
	tasks, err := l.retentionMgr.ListTasks(&q.TaskQuery{
		ExecutionID: executionID,
	})
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if err = l.jobserviceClient.PostAction(task.JobID, cjob.JobActionStop); err != nil {
			log.Errorf("failed to stop task %d, job ID: %s : %v", task.ID, task.JobID, err)
			continue
		}
	}
	return nil
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

func getRepositories(projectMgr project.Manager, repositoryMgr repository.Manager,
	projectID int64, chartServerEnabled bool) ([]*res.Candidate, error) {
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
	if chartServerEnabled {
		// get chart repositories when chart server is enabled
		chartRepositories, err := repositoryMgr.ListChartRepositories(projectID)
		if err != nil {
			return nil, err
		}
		for _, r := range chartRepositories {
			candidates = append(candidates, &res.Candidate{
				Namespace:  pro.Name,
				Repository: r.Name,
				Kind:       "chart",
			})
		}
	}

	return candidates, nil
}
