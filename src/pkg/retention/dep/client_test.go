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
	"testing"

	"github.com/goharbor/harbor/src/chartserver"
	jmodels "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/goharbor/harbor/src/testing/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"k8s.io/helm/pkg/proto/hapi/chart"
	"k8s.io/helm/pkg/repo"
)

type fakeCoreClient struct {
	clients.DumbCoreClient
}

func (f *fakeCoreClient) ListAllImages(project, repository string) ([]*models.TagResp, error) {
	image := &models.TagResp{}
	image.Name = "latest"
	return []*models.TagResp{image}, nil
}

func (f *fakeCoreClient) ListAllCharts(project, repository string) ([]*chartserver.ChartVersion, error) {
	metadata := &chart.Metadata{
		Name: "1.0",
	}
	chart := &chartserver.ChartVersion{}
	chart.ChartVersion = repo.ChartVersion{
		Metadata: metadata,
	}
	return []*chartserver.ChartVersion{chart}, nil
}

type fakeJobserviceClient struct{}

func (f *fakeJobserviceClient) SubmitJob(*jmodels.JobData) (string, error) {
	return "1", nil
}
func (f *fakeJobserviceClient) GetJobLog(uuid string) ([]byte, error) {
	return nil, nil
}
func (f *fakeJobserviceClient) PostAction(uuid, action string) error {
	return nil
}
func (f *fakeJobserviceClient) GetExecutions(uuid string) ([]job.Stats, error) {
	return nil, nil
}

type clientTestSuite struct {
	suite.Suite
}

func (c *clientTestSuite) TestGetCandidates() {
	client := &basicClient{}
	client.coreClient = &fakeCoreClient{}
	var repository *res.Repository
	// nil repository
	candidates, err := client.GetCandidates(repository)
	require.NotNil(c.T(), err)

	// image repository
	repository = &res.Repository{}
	repository.Kind = res.Image
	repository.Namespace = "library"
	repository.Name = "hello-world"
	candidates, err = client.GetCandidates(repository)
	require.Nil(c.T(), err)
	assert.Equal(c.T(), 1, len(candidates))
	assert.Equal(c.T(), res.Image, candidates[0].Kind)
	assert.Equal(c.T(), "library", candidates[0].Namespace)
	assert.Equal(c.T(), "hello-world", candidates[0].Repository)
	assert.Equal(c.T(), "latest", candidates[0].Tag)

	// chart repository
	repository.Kind = res.Chart
	repository.Namespace = "goharbor"
	repository.Name = "harbor"
	candidates, err = client.GetCandidates(repository)
	require.Nil(c.T(), err)
	assert.Equal(c.T(), 1, len(candidates))
	assert.Equal(c.T(), res.Chart, candidates[0].Kind)
	assert.Equal(c.T(), "goharbor", candidates[0].Namespace)
	assert.Equal(c.T(), "1.0", candidates[0].Tag)
}

func (c *clientTestSuite) TestDelete() {
	client := &basicClient{}
	client.coreClient = &fakeCoreClient{}

	var candidate *res.Candidate
	// nil candidate
	err := client.Delete(candidate)
	require.NotNil(c.T(), err)

	// image
	candidate = &res.Candidate{}
	candidate.Kind = res.Image
	err = client.Delete(candidate)
	require.Nil(c.T(), err)

	// chart
	candidate.Kind = res.Chart
	err = client.Delete(candidate)
	require.Nil(c.T(), err)

	// unsupported type
	candidate.Kind = "unsupported"
	err = client.Delete(candidate)
	require.NotNil(c.T(), err)
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}
