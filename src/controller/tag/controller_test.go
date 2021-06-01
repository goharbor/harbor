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

package tag

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	pkg_artifact "github.com/goharbor/harbor/src/pkg/artifact"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/goharbor/harbor/src/pkg/tag/model/tag"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/pkg/artifact"
	"github.com/goharbor/harbor/src/testing/pkg/immutable"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	tagtesting "github.com/goharbor/harbor/src/testing/pkg/tag"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type controllerTestSuite struct {
	suite.Suite
	ctl          *controller
	repoMgr      *repository.Manager
	artMgr       *artifact.Manager
	tagMgr       *tagtesting.FakeManager
	immutableMtr *immutable.FakeMatcher
}

func (c *controllerTestSuite) SetupTest() {
	c.repoMgr = &repository.Manager{}
	c.artMgr = &artifact.Manager{}
	c.tagMgr = &tagtesting.FakeManager{}
	c.immutableMtr = &immutable.FakeMatcher{}
	c.ctl = &controller{
		tagMgr:       c.tagMgr,
		artMgr:       c.artMgr,
		immutableMtr: c.immutableMtr,
	}

	var tagCtlTestConfig = map[string]interface{}{
		common.WithNotary: false,
	}
	config.InitWithSettings(tagCtlTestConfig)
}

func (c *controllerTestSuite) TestEnsureTag() {
	// the tag already exists under the repository and is attached to the artifact
	c.tagMgr.On("List").Return([]*tag.Tag{
		{
			ID:           1,
			RepositoryID: 1,
			ArtifactID:   1,
			Name:         "latest",
		},
	}, nil)
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&pkg_artifact.Artifact{
		ID: 1,
	}, nil)
	c.immutableMtr.On("Match").Return(false, nil)
	err := c.ctl.Ensure(orm.NewContext(nil, &ormtesting.FakeOrmer{}), 1, 1, "latest")
	c.Require().Nil(err)
	c.tagMgr.AssertExpectations(c.T())

	// reset the mock
	c.SetupTest()

	// the tag exists under the repository, but it is attached to other artifact
	c.tagMgr.On("List").Return([]*tag.Tag{
		{
			ID:           1,
			RepositoryID: 1,
			ArtifactID:   2,
			Name:         "latest",
		},
	}, nil)
	c.tagMgr.On("Update").Return(nil)
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&pkg_artifact.Artifact{
		ID: 1,
	}, nil)
	c.immutableMtr.On("Match").Return(false, nil)
	err = c.ctl.Ensure(orm.NewContext(nil, &ormtesting.FakeOrmer{}), 1, 1, "latest")
	c.Require().Nil(err)
	c.tagMgr.AssertExpectations(c.T())

	// reset the mock
	c.SetupTest()

	// the tag doesn't exist under the repository, create it
	c.tagMgr.On("List").Return([]*tag.Tag{}, nil)
	c.tagMgr.On("Create").Return(1, nil)
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&pkg_artifact.Artifact{
		ID: 1,
	}, nil)
	c.immutableMtr.On("Match").Return(false, nil)
	err = c.ctl.Ensure(orm.NewContext(nil, &ormtesting.FakeOrmer{}), 1, 1, "latest")
	c.Require().Nil(err)
	c.tagMgr.AssertExpectations(c.T())
}

func (c *controllerTestSuite) TestCount() {
	c.tagMgr.On("Count").Return(1, nil)
	total, err := c.ctl.Count(nil, nil)
	c.Require().Nil(err)
	c.Equal(int64(1), total)
}

func (c *controllerTestSuite) TestList() {
	c.tagMgr.On("List").Return([]*tag.Tag{
		{
			RepositoryID: 1,
			Name:         "testlist",
		},
	}, nil)
	tags, err := c.ctl.List(nil, nil, nil)
	c.Require().Nil(err)
	c.Require().Len(tags, 1)
	c.Equal(int64(1), tags[0].RepositoryID)
	c.Equal("testlist", tags[0].Name)
}

func (c *controllerTestSuite) TestGet() {
	getTest := &tag.Tag{}
	getTest.RepositoryID = 1
	getTest.Name = "testget"

	c.tagMgr.On("Get").Return(getTest, nil)
	tag, err := c.ctl.Get(nil, 1, nil)
	c.Require().Nil(err)
	c.tagMgr.AssertExpectations(c.T())
	c.Equal(int64(1), tag.RepositoryID)
	c.Equal(false, tag.Immutable)
}

func (c *controllerTestSuite) TestDelete() {
	c.tagMgr.On("Get").Return(&tag.Tag{
		RepositoryID: 1,
		Name:         "test",
	}, nil)
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&pkg_artifact.Artifact{
		ID: 1,
	}, nil)
	c.immutableMtr.On("Match").Return(false, nil)
	c.tagMgr.On("Delete").Return(nil)
	err := c.ctl.Delete(nil, 1)
	c.Require().Nil(err)
}

func (c *controllerTestSuite) TestDeleteImmutable() {
	c.tagMgr.On("Get").Return(&tag.Tag{
		RepositoryID: 1,
		Name:         "test",
	}, nil)
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&pkg_artifact.Artifact{
		ID: 1,
	}, nil)
	c.immutableMtr.On("Match").Return(true, nil)
	c.tagMgr.On("Delete").Return(nil)
	err := c.ctl.Delete(nil, 1)
	c.Require().NotNil(err)
	c.True(errors.IsErr(err, errors.PreconditionCode))
}

func (c *controllerTestSuite) TestUpdate() {
	c.tagMgr.On("Update").Return(nil)
	err := c.ctl.Update(nil, &Tag{
		Tag: tag.Tag{
			RepositoryID: 1,
			Name:         "test",
		},
		Immutable: true,
	}, "ArtifactID")
	c.Require().Nil(err)
}

func (c *controllerTestSuite) TestDeleteTags() {
	c.tagMgr.On("Get").Return(&tag.Tag{
		RepositoryID: 1,
	}, nil)
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&pkg_artifact.Artifact{
		ID: 1,
	}, nil)
	c.immutableMtr.On("Match").Return(false, nil)
	c.tagMgr.On("Delete").Return(nil)
	ids := []int64{1, 2, 3, 4}
	err := c.ctl.DeleteTags(nil, ids)
	c.Require().Nil(err)
}

func (c *controllerTestSuite) TestAssembleTag() {
	art := &pkg_artifact.Artifact{
		ID:             1,
		ProjectID:      1,
		RepositoryID:   1,
		RepositoryName: "library/hello-world",
		Digest:         "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
	}
	tg := &tag.Tag{
		ID:           1,
		RepositoryID: 1,
		ArtifactID:   1,
		Name:         "latest",
		PushTime:     time.Now(),
		PullTime:     time.Now(),
	}
	option := &Option{
		WithImmutableStatus: true,
	}

	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(art, nil)
	c.immutableMtr.On("Match").Return(true, nil)
	tag := c.ctl.assembleTag(nil, tg, option)
	c.Require().NotNil(tag)
	c.Equal(tag.ID, tg.ID)
	c.Equal(true, tag.Immutable)
	// TODO check signature
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}
