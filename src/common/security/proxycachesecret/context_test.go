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

package proxycachesecret

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/stretchr/testify/suite"
)

type proxyCacheSecretTestSuite struct {
	suite.Suite
	sc  *SecurityContext
	mgr *project.Manager
}

func (p *proxyCacheSecretTestSuite) SetupTest() {
	p.mgr = &project.Manager{}
	p.sc = &SecurityContext{
		repository: "library/hello-world",
		getProject: func(i interface{}) (*models.Project, error) {
			return p.mgr.Get(context.TODO(), i)
		},
	}
}

func (p *proxyCacheSecretTestSuite) TestName() {
	p.Equal("proxy_cache_secret", p.sc.Name())
}

func (p *proxyCacheSecretTestSuite) TestIsAuthenticated() {
	p.True(p.sc.IsAuthenticated())
}

func (p *proxyCacheSecretTestSuite) TestGetUsername() {
	p.Equal(ProxyCacheService, p.sc.GetUsername())
}

func (p *proxyCacheSecretTestSuite) TestIsSysAdmin() {
	p.False(p.sc.IsSysAdmin())
}

func (p *proxyCacheSecretTestSuite) TestIsSolutionUser() {
	p.False(p.sc.IsSolutionUser())
}

func (p *proxyCacheSecretTestSuite) TestCan() {
	// the action isn't pull/push
	action := rbac.ActionDelete
	resource := rbac.NewProjectNamespace(1).Resource(rbac.ResourceRepository)
	p.False(p.sc.Can(action, resource))

	// the resource isn't repository
	action = rbac.ActionPull
	resource = rbac.ResourceConfiguration
	p.False(p.sc.Can(action, resource))

	// the requested project not found
	action = rbac.ActionPull
	resource = rbac.NewProjectNamespace(2).Resource(rbac.ResourceRepository)
	p.mgr.On("Get", mock.Anything, mock.Anything).Return(nil, nil)
	p.False(p.sc.Can(action, resource))
	p.mgr.AssertExpectations(p.T())

	// reset the mock
	p.SetupTest()

	// pass for action pull
	action = rbac.ActionPull
	resource = rbac.NewProjectNamespace(1).Resource(rbac.ResourceRepository)
	p.mgr.On("Get", mock.Anything, mock.Anything).Return(&models.Project{
		ProjectID: 1,
		Name:      "library",
	}, nil)
	p.True(p.sc.Can(action, resource))
	p.mgr.AssertExpectations(p.T())

	// reset the mock
	p.SetupTest()

	// pass for action push
	action = rbac.ActionPush
	resource = rbac.NewProjectNamespace(1).Resource(rbac.ResourceRepository)
	p.mgr.On("Get", mock.Anything, mock.Anything).Return(&models.Project{
		ProjectID: 1,
		Name:      "library",
	}, nil)
	p.True(p.sc.Can(action, resource))
	p.mgr.AssertExpectations(p.T())
}

func TestProxyCacheSecretTestSuite(t *testing.T) {
	suite.Run(t, &proxyCacheSecretTestSuite{})
}
