// Copyright Project Harbor Authors
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
	"fmt"
	"testing"

	commonmodels "github.com/goharbor/harbor/src/common/models"
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
	suite.ClearSQLs = []string{}
	suite.dao = New()
}

func (suite *DaoTestSuite) TestCount() {
	ctx := orm.Context()
	{
		n, err := suite.dao.Count(ctx, nil)
		suite.Nil(err)
		users, err := suite.dao.List(orm.Context(), nil)
		suite.Nil(err)
		suite.Equal(len(users), int(n))
	}
	{

		n, err := suite.dao.Count(ctx, nil)
		suite.Nil(err)
		id, err := suite.dao.Create(ctx, &commonmodels.User{
			Username:        "testuser2",
			Realname:        "user test",
			Email:           "testuser@test.com",
			Password:        "somepassword",
			PasswordVersion: "sha256",
		})
		suite.Nil(err)
		defer suite.appendClearSQL(id)
		n2, err := suite.dao.Count(ctx, nil)
		suite.Nil(err)
		suite.Equal(n+1, n2)
		err2 := suite.dao.Update(ctx, &commonmodels.User{
			UserID:  id,
			Deleted: true,
		})
		suite.Nil(err2)
		n3, err := suite.dao.Count(ctx, nil)
		suite.Nil(err)
		suite.Equal(n, n3)

	}

}

func (suite *DaoTestSuite) TestList() {
	ctx := orm.Context()
	{
		users, err := suite.dao.List(ctx, q.New(q.KeyWords{"user_id": 1}))
		suite.Nil(err)
		suite.Len(users, 1)
	}

	{
		users, err := suite.dao.List(ctx, q.New(q.KeyWords{"username": "admin"}))
		suite.Nil(err)
		suite.Len(users, 1)
	}
	id, err := suite.dao.Create(ctx, &commonmodels.User{
		Username:        "list_test",
		Realname:        "list test",
		Email:           "list_test@test.com",
		Password:        "somepassword",
		PasswordVersion: "sha256",
	})
	suite.appendClearSQL(id)
	suite.Nil(err)
	{
		users, err := suite.dao.List(ctx, q.New(q.KeyWords{"username_or_email": "list_test"}))
		suite.Nil(err)
		suite.Len(users, 1)
	}
	{
		users, err := suite.dao.List(ctx, q.New(q.KeyWords{"username_or_email": "list_test@test.com"}))
		suite.Nil(err)
		suite.Len(users, 1)
	}
	{
		users, err := suite.dao.List(ctx, q.New(q.KeyWords{"username_or_email": "noremail_norusername"}))
		suite.Nil(err)
		suite.Len(users, 0)
	}

}

func (suite *DaoTestSuite) TestCreate() {
	cases := []struct {
		name     string
		input    *commonmodels.User
		hasError bool
	}{
		{
			name: "create with user ID",
			input: &commonmodels.User{
				UserID:          3,
				Username:        "testuser",
				Realname:        "user test",
				Email:           "testuser@test.com",
				Password:        "somepassword",
				PasswordVersion: "sha256",
			},
			hasError: true,
		},
		{
			name: "create without user ID",
			input: &commonmodels.User{
				Username:        "testuser",
				Realname:        "user test",
				Email:           "testuser@test.com",
				Password:        "somepassword",
				PasswordVersion: "sha256",
			},
			hasError: false,
		},
		{
			name: "create with empty email_1",
			input: &commonmodels.User{
				Username:        "emptyemail1",
				Realname:        "empty test",
				Email:           "",
				Password:        "somepassword",
				PasswordVersion: "sha256",
			},
			hasError: false,
		},
		{
			name: "create with empty email_2",
			input: &commonmodels.User{
				Username:        "emptyemail2",
				Realname:        "empty test2",
				Email:           "",
				Password:        "somepassword",
				PasswordVersion: "sha256",
			},
			hasError: false,
		},
	}
	for _, c := range cases {
		suite.Run(c.name, func() {
			ctx := orm.Context()
			id, err := suite.dao.Create(ctx, c.input)
			defer suite.appendClearSQL(id)
			if c.hasError {
				suite.NotNil(err)
			} else {
				suite.Nil(err)
				l, err2 := suite.dao.List(ctx, q.New(q.KeyWords{"user_id": id}))
				suite.Nil(err2)
				suite.Equal(c.input.Username, l[0].Username)
				suite.Equal(c.input.Password, l[0].Password)
				suite.Equal(c.input.Email, l[0].Email)
				suite.Equal(c.input.Realname, l[0].Realname)
				suite.Equal(c.input.PasswordVersion, l[0].PasswordVersion)
			}
		})
	}
}

func (suite *DaoTestSuite) appendClearSQL(uid int) {
	suite.ClearSQLs = append(suite.ClearSQLs, fmt.Sprintf("DELETE FROM harbor_user WHERE user_id = %d", uid))
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
