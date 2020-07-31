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

package dao

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = New()
}

func (suite *DaoTestSuite) TestCreate() {
	id, err := suite.dao.Create(orm.Context(), 1, "foo", "bar")
	if suite.Nil(err) {
		defer func() {
			suite.Nil(suite.dao.Delete(orm.Context(), q.New(q.KeyWords{"id": id})))
		}()
	}

	mds, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"id": id}))
	suite.Nil(err)
	suite.Len(mds, 1)
	suite.Equal("foo", mds[0].Name)
	suite.Equal("bar", mds[0].Value)

}

func (suite *DaoTestSuite) TestUpdate() {
	id, err := suite.dao.Create(orm.Context(), 1, "foo", "bar")
	if suite.Nil(err) {
		defer func() {
			suite.Nil(suite.dao.Delete(orm.Context(), q.New(q.KeyWords{"id": id})))
		}()
	}

	suite.Nil(suite.dao.Update(orm.Context(), 1, "foo", "Bar"))
	mds, err := suite.dao.List(orm.Context(), q.New(q.KeyWords{"id": id}))
	suite.Nil(err)
	suite.Len(mds, 1)
	suite.Equal("foo", mds[0].Name)
	suite.Equal("Bar", mds[0].Value)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
