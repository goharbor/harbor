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
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	ierror "github.com/goharbor/harbor/src/lib/error"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite
}

func (suite *ControllerTestSuite) TestGetByName() {
	mgr := &project.FakeManager{}
	mgr.On("Get", "library").Return(&models.Project{ProjectID: 1, Name: "library"}, nil)
	mgr.On("Get", "test").Return(nil, nil)
	mgr.On("Get", "oops").Return(nil, fmt.Errorf("oops"))

	c := controller{projectMgr: mgr}

	{
		p, err := c.GetByName(context.TODO(), "library", Metadata(false))
		suite.Nil(err)
		suite.Equal("library", p.Name)
		suite.Equal(int64(1), p.ProjectID)
	}

	{
		p, err := c.GetByName(context.TODO(), "test", Metadata(false))
		suite.Error(err)
		suite.True(ierror.IsNotFoundErr(err))
		suite.Nil(p)
	}

	{
		p, err := c.GetByName(context.TODO(), "oops", Metadata(false))
		suite.Error(err)
		suite.False(ierror.IsNotFoundErr(err))
		suite.Nil(p)
	}
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
