package dao

import (
	"fmt"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/rbac/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO

	permissionID1 int64
	permissionID2 int64
	permissionID3 int64
	permissionID4 int64

	rbacPolicyID1 int64
	rbacPolicyID2 int64
	rbacPolicyID3 int64
	rbacPolicyID4 int64
}

func (suite *DaoTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.dao = New()
	suite.Suite.ClearTables = []string{"rbac_policy", "role_permission"}

	suite.prepareRolePermission()
	suite.preparePermissionPolicy()
}

func (suite *DaoTestSuite) prepareRolePermission() {
	rp := &model.RolePermission{
		RoleType:           "robot",
		RoleID:             1,
		PermissionPolicyID: 2,
	}
	id, err := suite.dao.CreatePermission(orm.Context(), rp)
	suite.permissionID1 = id
	suite.Nil(err)

	rp2 := &model.RolePermission{
		RoleType:           "robot",
		RoleID:             1,
		PermissionPolicyID: 3,
	}
	id2, err := suite.dao.CreatePermission(orm.Context(), rp2)
	suite.permissionID2 = id2
	suite.Nil(err)

	rp3 := &model.RolePermission{
		RoleType:           "robot",
		RoleID:             1,
		PermissionPolicyID: 4,
	}
	id3, err := suite.dao.CreatePermission(orm.Context(), rp3)
	suite.permissionID3 = id3
	suite.Nil(err)

	rp4 := &model.RolePermission{
		RoleType:           "serviceaccount",
		RoleID:             2,
		PermissionPolicyID: 1,
	}
	id4, err := suite.dao.CreatePermission(orm.Context(), rp4)
	suite.permissionID4 = id4
	suite.Nil(err)
}

func (suite *DaoTestSuite) preparePermissionPolicy() {
	rp := &model.PermissionPolicy{
		Scope:    "/system",
		Resource: "label",
		Action:   "create",
	}
	id, err := suite.dao.CreateRbacPolicy(orm.Context(), rp)
	suite.rbacPolicyID1 = id
	suite.Nil(err)

	rp2 := &model.PermissionPolicy{
		Scope:    "/project/1",
		Resource: "repository",
		Action:   "push",
	}
	id2, err := suite.dao.CreateRbacPolicy(orm.Context(), rp2)
	suite.rbacPolicyID2 = id2
	suite.Nil(err)

	rp3 := &model.PermissionPolicy{
		Scope:    "/project/1",
		Resource: "repository",
		Action:   "pull",
	}
	id3, err := suite.dao.CreateRbacPolicy(orm.Context(), rp3)
	suite.rbacPolicyID3 = id3
	suite.Nil(err)

	rp4 := &model.PermissionPolicy{
		Scope:    "/project/2",
		Resource: "helm-chart",
		Action:   "create",
	}
	id4, err := suite.dao.CreateRbacPolicy(orm.Context(), rp4)
	suite.rbacPolicyID4 = id4
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestCreatePermission() {
	rp := &model.RolePermission{
		RoleType:           "robot",
		RoleID:             1,
		PermissionPolicyID: 2,
	}
	_, err := suite.dao.CreatePermission(orm.Context(), rp)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestDeletePermission() {
	err := suite.dao.DeletePermission(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	err = suite.dao.DeletePermission(orm.Context(), suite.permissionID2)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestListPermissions() {
	rps, err := suite.dao.ListPermissions(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"role_type":            "robot",
			"role_id":              1,
			"permission_policy_id": 4,
		},
	})
	suite.Require().Nil(err)
	suite.Equal(int64(4), rps[0].PermissionPolicyID)
}

func (suite *DaoTestSuite) TestDeletePermissionsByRole() {
	err := suite.dao.DeletePermissionsByRole(orm.Context(), "serviceaccount", 2)
	suite.Require().Nil(err)

	rps, err := suite.dao.ListPermissions(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"role_type": "serviceaccount",
			"role_id":   2,
		},
	})
	suite.Require().Nil(err)
	suite.Equal(0, len(rps))

}

func (suite *DaoTestSuite) TestCreateRbacPolicy() {
	rp := &model.PermissionPolicy{
		Scope:    "/system",
		Resource: "label",
		Action:   "create",
	}
	_, err := suite.dao.CreateRbacPolicy(orm.Context(), rp)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestDeleteRbacPolicy() {
	err := suite.dao.DeleteRbacPolicy(orm.Context(), 1234)
	suite.Require().NotNil(err)
	suite.True(errors.IsErr(err, errors.NotFoundCode))

	err = suite.dao.DeleteRbacPolicy(orm.Context(), suite.rbacPolicyID2)
	suite.Nil(err)
}

func (suite *DaoTestSuite) TestListRbacPolicies() {
	rps, err := suite.dao.ListRbacPolicies(orm.Context(), &q.Query{
		Keywords: map[string]interface{}{
			"scope":    "/project/1",
			"resource": "repository",
			"action":   "pull",
		},
	})
	suite.Require().Nil(err)
	suite.Equal(suite.rbacPolicyID3, rps[0].ID)
}

func (suite *DaoTestSuite) TestGetPermissionsByRole() {
	rp := &model.PermissionPolicy{
		Scope:    "/system",
		Resource: "label",
		Action:   "delete",
	}
	id, err := suite.dao.CreateRbacPolicy(orm.Context(), rp)
	suite.Nil(err)

	rpe := &model.RolePermission{
		RoleType:           "TestGetPermissionsByRole",
		RoleID:             1,
		PermissionPolicyID: id,
	}
	_, err = suite.dao.CreatePermission(orm.Context(), rpe)
	suite.Nil(err)

	rpes, err := suite.dao.GetPermissionsByRole(orm.Context(), "TestGetPermissionsByRole", 1)
	suite.Nil(err)
	fmt.Println(rpes[0])
	suite.Equal("/system", rpes[0].Scope)
}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
