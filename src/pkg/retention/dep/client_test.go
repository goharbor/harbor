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
	"net/http"
	"testing"

	jmodels "github.com/goharbor/harbor/src/common/job/models"
	modelsv2 "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/selector"
	model_tag "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/testing/clients"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type fakeCoreClient struct {
	clients.DumbCoreClient
}

func (f *fakeCoreClient) ListAllArtifacts(project, repository string) ([]*modelsv2.Artifact, error) {
	image := &modelsv2.Artifact{}
	image.Digest = "sha256:123456"
	image.Tags = []*tag.Tag{
		{
			Tag: model_tag.Tag{
				Name: "latest",
			},
		},
	}
	return []*modelsv2.Artifact{image}, nil
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
	var repository *selector.Repository
	// nil repository
	candidates, err := client.GetCandidates(repository)
	require.NotNil(c.T(), err)

	// image repository
	repository = &selector.Repository{}
	repository.Kind = selector.Image
	repository.Namespace = "library"
	repository.Name = "hello-world"
	candidates, err = client.GetCandidates(repository)
	require.Nil(c.T(), err)
	assert.Equal(c.T(), 1, len(candidates))
	assert.Equal(c.T(), selector.Image, candidates[0].Kind)
	assert.Equal(c.T(), "library", candidates[0].Namespace)
	assert.Equal(c.T(), "hello-world", candidates[0].Repository)
	assert.Equal(c.T(), "latest", candidates[0].Tags[0])

	/*
		// chart repository
		repository.Kind = art.Chart
		repository.Namespace = "goharbor"
		repository.Name = "harbor"
		candidates, err = client.GetCandidates(repository)
		require.Nil(c.T(), err)
		assert.Equal(c.T(), 1, len(candidates))
		assert.Equal(c.T(), art.Chart, candidates[0].Kind)
		assert.Equal(c.T(), "goharbor", candidates[0].Namespace)
		assert.Equal(c.T(), "1.0", candidates[0].Tag)
	*/
}

func (c *clientTestSuite) TestDelete() {
	client := &basicClient{}
	client.coreClient = &fakeCoreClient{}

	var candidate *selector.Candidate
	// nil candidate
	err := client.Delete(candidate)
	require.NotNil(c.T(), err)

	// image
	candidate = &selector.Candidate{}
	candidate.Kind = selector.Image
	err = client.Delete(candidate)
	require.Nil(c.T(), err)

	/*
		// chart
		candidate.Kind = art.Chart
		err = client.Delete(candidate)
		require.Nil(c.T(), err)
	*/

	// unsupported type
	candidate.Kind = "unsupported"
	err = client.Delete(candidate)
	require.NotNil(c.T(), err)
}

func (c *clientTestSuite) TestInjectVendorType() {
	injector := &injectVendorType{}
	req, err := http.NewRequest("GET", "http://localhost:8080/api", nil)
	assert.NoError(c.T(), err)
	assert.Equal(c.T(), "", req.Header.Get("VendorType"))
	// after injecting should appear vendor type in header
	err = injector.Modify(req)
	assert.NoError(c.T(), err)
	assert.Equal(c.T(), "RETENTION", req.Header.Get("VendorType"))
}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(clientTestSuite))
}
