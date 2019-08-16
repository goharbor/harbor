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

package dep

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/pkg/clients/core"
	"github.com/goharbor/harbor/src/pkg/retention/res"
)

// DefaultClient for the retention
var DefaultClient = NewClient()

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

	// Delete the given repository
	//
	//  Arguments:
	//    repo *res.Repository : repository info
	//
	//  Returns:
	//    error            : common error if any errors occurred
	DeleteRepository(repo *res.Repository) error

	// Delete the specified candidate
	//
	//  Arguments:
	//    candidate *res.Candidate : the deleting candidate
	//
	//  Returns:
	//    error : common error if any errors occurred
	Delete(candidate *res.Candidate) error
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
	internalCoreURL := config.GetCoreURL()
	jobserviceSecret := config.GetAuthSecret()
	authorizer := auth.NewSecretAuthorizer(jobserviceSecret)
	coreClient := core.New(internalCoreURL, c, authorizer)

	return &basicClient{
		internalCoreURL: internalCoreURL,
		coreClient:      coreClient,
	}
}

// basicClient is a default
type basicClient struct {
	internalCoreURL string
	coreClient      core.Client
}

// GetCandidates gets the tag candidates under the repository
func (bc *basicClient) GetCandidates(repository *res.Repository) ([]*res.Candidate, error) {
	if repository == nil {
		return nil, errors.New("repository is nil")
	}
	candidates := make([]*res.Candidate, 0)
	switch repository.Kind {
	case res.Image:
		images, err := bc.coreClient.ListAllImages(repository.Namespace, repository.Name)
		if err != nil {
			return nil, err
		}
		for _, image := range images {
			labels := make([]string, 0)
			for _, label := range image.Labels {
				labels = append(labels, label.Name)
			}
			candidate := &res.Candidate{
				Kind:         res.Image,
				Namespace:    repository.Namespace,
				Repository:   repository.Name,
				Tag:          image.Name,
				Labels:       labels,
				CreationTime: image.Created.Unix(),
				PulledTime:   image.PullTime.Unix(),
				PushedTime:   image.PushTime.Unix(),
			}
			candidates = append(candidates, candidate)
		}
	/*
		case res.Chart:
			charts, err := bc.coreClient.ListAllCharts(repository.Namespace, repository.Name)
			if err != nil {
				return nil, err
			}
			for _, chart := range charts {
				labels := make([]string, 0)
				for _, label := range chart.Labels {
					labels = append(labels, label.Name)
				}
				candidate := &res.Candidate{
					Kind:         res.Chart,
					Namespace:    repository.Namespace,
					Repository:   repository.Name,
					Tag:          chart.Name,
					Labels:       labels,
					CreationTime: chart.Created.Unix(),
					PushedTime:   ,
					PulledTime:   ,
				}
				candidates = append(candidates, candidate)
			}
	*/
	default:
		return nil, fmt.Errorf("unsupported repository kind: %s", repository.Kind)
	}
	return candidates, nil
}

// DeleteRepository deletes the specified repository
func (bc *basicClient) DeleteRepository(repo *res.Repository) error {
	if repo == nil {
		return errors.New("repository is nil")
	}
	switch repo.Kind {
	case res.Image:
		return bc.coreClient.DeleteImageRepository(repo.Namespace, repo.Name)
	/*
		case res.Chart:
			return bc.coreClient.DeleteChartRepository(repo.Namespace, repo.Name)
	*/
	default:
		return fmt.Errorf("unsupported repository kind: %s", repo.Kind)
	}
}

// Deletes the specified candidate
func (bc *basicClient) Delete(candidate *res.Candidate) error {
	if candidate == nil {
		return errors.New("candidate is nil")
	}
	switch candidate.Kind {
	case res.Image:
		return bc.coreClient.DeleteImage(candidate.Namespace, candidate.Repository, candidate.Tag)
	/*
		case res.Chart:
			return bc.coreClient.DeleteChart(candidate.Namespace, candidate.Repository, candidate.Tag)
	*/
	default:
		return fmt.Errorf("unsupported candidate kind: %s", candidate.Kind)
	}
}
