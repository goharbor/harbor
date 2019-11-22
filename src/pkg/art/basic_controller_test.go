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

package art

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// TestControllerSuite is a test suite for testing controller.
type TestControllerSuite struct {
	suite.Suite

	c *basicController
	m *MockManager
}

// TestController is the entry point of TestControllerSuite.
func TestController(t *testing.T) {
	suite.Run(t, &TestControllerSuite{})
}

// SetupSuite prepares env for test suite.
func (suite *TestControllerSuite) SetupSuite() {
	suite.m = &MockManager{}
	suite.c = &basicController{
		m: suite.m,
	}
}

// TestControllerList ...
func (suite *TestControllerSuite) TestControllerList() {
	kws := make(map[string]interface{})
	kws["digest"] = "digest-code"
	query := &q.Query{
		Keywords: kws,
	}

	artifacts := []*models.Artifact{
		{
			ID:     1000,
			PID:    1,
			Repo:   "library/busybox",
			Tag:    "dev",
			Digest: "digest-code",
			Kind:   "image",
		},
	}

	suite.m.On("List", query).Return(artifacts, nil)

	arts, err := suite.c.List(query)
	require.NoError(suite.T(), err)
	suite.Equal(1, len(arts))
}

// MockManager ...
type MockManager struct {
	mock.Mock
}

// List ...
func (mm *MockManager) List(query *q.Query) ([]*models.Artifact, error) {
	args := mm.Called(query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).([]*models.Artifact), args.Error(1)
}
