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
	"errors"
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/clients/core"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/res"
)

// DefaultClient for the retention
var DefaultClient Client

// Client is designed to access core service to get required infos
type Client interface {
	// Get the tag candidates under the repository
	//
	//  Arguments:
	//    repo *res.Repository : repository info
	//
	//  Returns:
	//    []*res.Candidate : candidates returned
	//    error            : common error if any errors occurred
	GetCandidates(repo *res.Repository) ([]*res.Candidate, error)

	// Delete the specified candidate
	//
	//  Arguments:
	//    candidate *res.Candidate : the deleting candidate
	//
	//  Returns:
	//    error : common error if any errors occurred
	Delete(candidate *res.Candidate) error

	// SubmitTask to jobservice
	//
	//  Arguments:
	//    taskID                      : the ID of task
	//    repository *res.Repository  : repository info
	//    meta *policy.LiteMeta       : policy lite metadata
	//
	//  Returns:
	//    string : the job ID
	//    error  : common error if any errors occurred
	SubmitTask(taskID int64, repository *res.Repository, meta *policy.LiteMeta) (string, error)
}

// NewClient new a basic client
func NewClient(client ...*http.Client) Client {
	var c *http.Client
	if len(client) > 0 {
		c = client[0]
	}
	if c == nil {
		c = http.DefaultClient
	}

	// init core client
	internalCoreURL := config.InternalCoreURL()
	jobserviceSecret := config.JobserviceSecret()
	authorizer := auth.NewSecretAuthorizer(jobserviceSecret)
	coreClient := core.New(internalCoreURL, c, authorizer)

	// init jobservice client
	internalJobserviceURL := config.InternalJobServiceURL()
	coreSecret := config.CoreSecret()
	jobserviceClient := cjob.NewDefaultClient(internalJobserviceURL, coreSecret)

	return &basicClient{
		internalCoreURL:  internalCoreURL,
		coreClient:       coreClient,
		jobserviceClient: jobserviceClient,
	}
}

// basicClient is a default
type basicClient struct {
	internalCoreURL  string
	coreClient       core.Client
	jobserviceClient cjob.Client
}

// GetCandidates gets the tag candidates under the repository
func (bc *basicClient) GetCandidates(repository *res.Repository) ([]*res.Candidate, error) {
	if repository == nil {
		return nil, errors.New("repository is nil")
	}
	candidates := make([]*res.Candidate, 0)
	switch repository.Kind {
	case CandidateKindImage:
		images, err := bc.coreClient.ListAllImages(repository.Namespace, repository.Name)
		if err != nil {
			return nil, err
		}
		for _, image := range images {
			labels := []string{}
			for _, label := range image.Labels {
				labels = append(labels, label.Name)
			}
			candidate := &res.Candidate{
				Kind:         CandidateKindImage,
				Namespace:    repository.Namespace,
				Repository:   repository.Name,
				Tag:          image.Name,
				Labels:       labels,
				CreationTime: image.Created.Unix(),
				// TODO: populate the pull/push time
				// PulledTime: ,
				// PushedTime:,
			}
			candidates = append(candidates, candidate)
		}
	case CandidateKindChart:
		charts, err := bc.coreClient.ListAllCharts(repository.Namespace, repository.Name)
		if err != nil {
			return nil, err
		}
		for _, chart := range charts {
			labels := []string{}
			for _, label := range chart.Labels {
				labels = append(labels, label.Name)
			}
			candidate := &res.Candidate{
				Kind:         CandidateKindChart,
				Namespace:    repository.Namespace,
				Repository:   repository.Name,
				Tag:          chart.Name,
				Labels:       labels,
				CreationTime: chart.Created.Unix(),
				// TODO: populate the pull/push time
				// PulledTime: ,
				// PushedTime:,
			}
			candidates = append(candidates, candidate)
		}
	default:
		return nil, fmt.Errorf("unsupported repository kind: %s", repository.Kind)
	}
	return candidates, nil
}

// Deletes the specified candidate
func (bc *basicClient) Delete(candidate *res.Candidate) error {
	if candidate == nil {
		return errors.New("candidate is nil")
	}
	switch candidate.Kind {
	case CandidateKindImage:
		return bc.coreClient.DeleteImage(candidate.Namespace, candidate.Repository, candidate.Tag)
	case CandidateKindChart:
		return bc.coreClient.DeleteChart(candidate.Namespace, candidate.Repository, candidate.Tag)
	default:
		return fmt.Errorf("unsupported candidate kind: %s", candidate.Kind)
	}
}

// SubmitTask to jobservice
func (bc *basicClient) SubmitTask(taskID int64, repository *res.Repository, meta *policy.LiteMeta) (string, error) {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
		StatusHook: fmt.Sprintf("%s/service/notifications/jobs/retention/tasks/%d", bc.internalCoreURL, taskID),
	}
	j.Name = job.Retention
	j.Parameters = map[string]interface{}{
		ParamRepo: repository,
		ParamMeta: meta,
	}
	return bc.jobserviceClient.SubmitJob(j)
}
