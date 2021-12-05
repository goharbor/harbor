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

package project

import (
	"context"
	"fmt"
	commonmodels "github.com/goharbor/harbor/src/common/models"
	"testing"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	models2 "github.com/goharbor/harbor/src/pkg/allowlist/models"
	"github.com/goharbor/harbor/src/pkg/project/models"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	allowlisttesting "github.com/goharbor/harbor/src/testing/pkg/allowlist"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/project/metadata"
	"github.com/goharbor/harbor/src/testing/pkg/user"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
}

func (suite *ControllerTestSuite) TestCreate() {
	ctx := orm.NewContext(context.TODO(), &ormtesting.FakeOrmer{})
	mgr := &project.Manager{}

	allowlistMgr := &allowlisttesting.Manager{}
	allowlistMgr.On("CreateEmpty", mock.Anything, mock.Anything).Return(nil)

	metadataMgr := &metadata.Manager{}

	c := controller{projectMgr: mgr, allowlistMgr: allowlistMgr, metaMgr: metadataMgr}

	{
		metadataMgr.On("Add", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
		mgr.On("Create", mock.Anything, mock.Anything).Return(int64(2), nil).Once()
		projectID, err := c.Create(ctx, &models.Project{OwnerID: 1, Metadata: map[string]string{"public": "true"}})
		suite.Nil(err)
		suite.Equal(int64(2), projectID)
	}

	{
		metadataMgr.On("Add", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("oops")).Once()
		mgr.On("Create", mock.Anything, mock.Anything).Return(int64(2), nil).Once()
		projectID, err := c.Create(ctx, &models.Project{OwnerID: 1, Metadata: map[string]string{"public": "true"}})
		suite.Error(err)
		suite.Equal(int64(0), projectID)
	}
}

func (suite *ControllerTestSuite) TestGetByName() {
	ctx := context.TODO()

	mgr := &project.Manager{}
	mgr.On("Get", ctx, "library").Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	mgr.On("Get", ctx, "test").Return(nil, errors.NotFoundError(nil))
	mgr.On("Get", ctx, "oops").Return(nil, fmt.Errorf("oops"))

	allowlistMgr := &allowlisttesting.Manager{}

	metadataMgr := &metadata.Manager{}
	metadataMgr.On("Get", ctx, mock.Anything).Return(map[string]string{"public": "true"}, nil)

	c := controller{projectMgr: mgr, allowlistMgr: allowlistMgr, metaMgr: metadataMgr}

	{
		p, err := c.GetByName(ctx, "library")
		suite.Nil(err)
		suite.Equal("library", p.Name)
		suite.Equal(int64(1), p.ProjectID)
	}

	{
		p, err := c.GetByName(ctx, "test")
		suite.Error(err)
		suite.True(errors.IsNotFoundErr(err))
		suite.Nil(p)
	}

	{
		p, err := c.GetByName(ctx, "oops")
		suite.Error(err)
		suite.False(errors.IsNotFoundErr(err))
		suite.Nil(p)
	}

	{
		allowlistMgr.On("Get", mock.Anything, mock.Anything).Return(&models2.CVEAllowlist{ProjectID: 1}, nil)
		p, err := c.GetByName(ctx, "library", WithCVEAllowlist())
		suite.Nil(err)
		suite.Equal("library", p.Name)
		suite.Equal(p.ProjectID, p.CVEAllowlist.ProjectID)
	}
}

func (suite *ControllerTestSuite) TestWithOwner() {
	ctx := context.TODO()

	mgr := &project.Manager{}
	mgr.On("Get", ctx, int64(1)).Return(&models.Project{ProjectID: 1, OwnerID: 1, Name: "library"}, nil)
	mgr.On("Get", ctx, "library").Return(&models.Project{ProjectID: 1, OwnerID: 1, Name: "library"}, nil)
	mgr.On("List", ctx, mock.Anything).Return([]*models.Project{
		{ProjectID: 1, OwnerID: 1, Name: "library"},
	}, nil)

	userMgr := &user.Manager{}
	userMgr.On("List", ctx, mock.Anything).Return(commonmodels.Users{
		&commonmodels.User{UserID: 1, Username: "admin"},
	}, nil)

	c := controller{projectMgr: mgr, userMgr: userMgr}

	{
		project, err := c.Get(ctx, int64(1), Metadata(false), WithOwner())
		suite.Nil(err)
		suite.Equal("admin", project.OwnerName)
	}

	{
		project, err := c.GetByName(ctx, "library", Metadata(false), WithOwner())
		suite.Nil(err)
		suite.Equal("admin", project.OwnerName)
	}

	{
		projects, err := c.List(ctx, q.New(q.KeyWords{"project_id__in": []int64{1}}), Metadata(false), WithOwner())
		suite.Nil(err)
		suite.Len(projects, 1)
		suite.Equal("admin", projects[0].OwnerName)
	}
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
