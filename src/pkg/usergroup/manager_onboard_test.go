//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package usergroup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
	daotesting "github.com/goharbor/harbor/src/testing/pkg/usergroup/dao"
)

const (
	testGroupName = "harbor_dev"
	testGroupDN   = "cn=harbor_dev,ou=groups,dc=example,dc=com"
)

func ldapGroup() *model.UserGroup {
	return &model.UserGroup{GroupName: testGroupName, GroupType: common.LDAPGroupType, LdapGroupDN: testGroupDN}
}

// nameOnlyQuery matches a query on the group name with no group type filter,
// because user_group is unique on group_name across all group types.
func nameOnlyQuery(query *q.Query) bool {
	_, hasType := query.Keywords["GroupType"]
	return query.Keywords["GroupName"] == testGroupName && !hasType
}

func TestOnboardLdapGroupNameUsedByOtherGroupType(t *testing.T) {
	for _, tc := range []struct {
		groupType int
		typeName  string
	}{
		{common.HTTPGroupType, "HTTP"},
		{common.OIDCGroupType, "OIDC"},
	} {
		t.Run(tc.typeName, func(t *testing.T) {
			d := daotesting.NewDAO(t)
			d.On("Query", mock.Anything, mock.MatchedBy(nameOnlyQuery)).
				Return([]*model.UserGroup{{ID: 7, GroupName: testGroupName, GroupType: tc.groupType}}, nil)

			err := (&manager{dao: d}).Onboard(context.TODO(), ldapGroup())

			require.Error(t, err)
			assert.True(t, errors.IsConflictErr(err))
			assert.Equal(t, `cannot onboard LDAP group "harbor_dev" (`+testGroupDN+`): the name is already used by `+tc.typeName+` user group 7`, err.Error())
			d.AssertNotCalled(t, "ReadOrCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestOnboardLdapGroupAlreadyOnboarded(t *testing.T) {
	d := daotesting.NewDAO(t)
	d.On("Query", mock.Anything, mock.MatchedBy(nameOnlyQuery)).
		Return([]*model.UserGroup{{ID: 3, GroupName: testGroupName, GroupType: common.LDAPGroupType, LdapGroupDN: testGroupDN}}, nil)
	g := ldapGroup()

	require.NoError(t, (&manager{dao: d}).Onboard(context.TODO(), g))
	assert.Equal(t, 3, g.ID)
}

func TestOnboardLdapGroupNameUsedByOtherLdapGroupFallsBackToDN(t *testing.T) {
	d := daotesting.NewDAO(t)
	d.On("Query", mock.Anything, mock.MatchedBy(nameOnlyQuery)).
		Return([]*model.UserGroup{{ID: 3, GroupName: testGroupName, GroupType: common.LDAPGroupType, LdapGroupDN: "cn=harbor_dev,ou=other,dc=example,dc=com"}}, nil)
	d.On("ReadOrCreate", mock.Anything, mock.MatchedBy(func(g *model.UserGroup) bool { return g.GroupName == testGroupDN }), "LdapGroupDN", "GroupType").
		Return(true, int64(9), nil)
	g := ldapGroup()

	require.NoError(t, (&manager{dao: d}).Onboard(context.TODO(), g))
	assert.Equal(t, 9, g.ID)
	assert.Equal(t, testGroupDN, g.GroupName)
}

func TestOnboardLdapGroupInsertConflictOnName(t *testing.T) {
	d := daotesting.NewDAO(t)
	d.On("Query", mock.Anything, mock.MatchedBy(nameOnlyQuery)).Return([]*model.UserGroup{}, nil)
	// orm.ReadOrCreate re-reads by key after a duplicate key and finds nothing
	// when the insert collided on group_name, e.g. a concurrent onboard
	d.On("ReadOrCreate", mock.Anything, mock.Anything, "LdapGroupDN", "GroupType").Return(false, int64(0), orm.ErrNoRows)

	err := (&manager{dao: d}).Onboard(context.TODO(), ldapGroup())

	require.Error(t, err)
	assert.True(t, errors.IsConflictErr(err))
	assert.Equal(t, `cannot onboard user group "harbor_dev": it conflicts with an existing user group of the same name`, err.Error())
}

func TestOnboardLdapGroupCreated(t *testing.T) {
	d := daotesting.NewDAO(t)
	d.On("Query", mock.Anything, mock.MatchedBy(nameOnlyQuery)).Return([]*model.UserGroup{}, nil)
	d.On("ReadOrCreate", mock.Anything, mock.Anything, "LdapGroupDN", "GroupType").Return(true, int64(5), nil)
	g := ldapGroup()

	require.NoError(t, (&manager{dao: d}).Onboard(context.TODO(), g))
	assert.Equal(t, 5, g.ID)
	assert.Equal(t, testGroupName, g.GroupName)
}
