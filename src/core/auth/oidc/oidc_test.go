package oidc

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	retCode := m.Run()
	os.Exit(retCode)
}

func TestAuth_SearchGroup(t *testing.T) {
	a := Auth{}
	res, err := a.SearchGroup("grp")
	assert.Nil(t, err)
	assert.Equal(t, models.UserGroup{GroupName: "grp", GroupType: common.OIDCGroupType}, *res)
}

func TestAuth_OnBoardGroup(t *testing.T) {
	a := Auth{}
	g1 := &models.UserGroup{GroupName: "", GroupType: common.OIDCGroupType}
	err1 := a.OnBoardGroup(g1, "")
	assert.NotNil(t, err1)
	g2 := &models.UserGroup{GroupName: "group", GroupType: common.LDAPGroupType}
	err2 := a.OnBoardGroup(g2, "")
	assert.NotNil(t, err2)
}
