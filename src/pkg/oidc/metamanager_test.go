package oidc

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/testing/mock"
	tdao "github.com/goharbor/harbor/src/testing/pkg/oidc/dao"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// encrypt "secret1" using key "naa4JtarA1Zsc3uY" (set in helper_test)
var encSecret = "<enc-v1>6FvOrx1O9TKBdalX4gMQrrKNZ99KIyg="

type metaMgrTestSuite struct {
	suite.Suite
	mgr MetaManager
	dao *tdao.MetaDAO
}

func (m *metaMgrTestSuite) SetupTest() {
	m.dao = &tdao.MetaDAO{}
	m.mgr = &metaManager{
		dao: m.dao,
	}
}

func (m *metaMgrTestSuite) TestGetByUserID() {
	{
		m.dao.On("List", mock.Anything, testifymock.MatchedBy(
			func(query *q.Query) bool {
				return query.Keywords["user_id"] == 8
			})).Return([]*models.OIDCUser{}, nil)
		_, err := m.mgr.GetByUserID(context.Background(), 8)
		m.NotNil(err)
	}
	{
		m.dao.On("List", mock.Anything, testifymock.MatchedBy(
			func(query *q.Query) bool {
				return query.Keywords["user_id"] == 9
			})).Return([]*models.OIDCUser{

			{ID: 1, UserID: 9, Secret: encSecret, Token: "token1"},
			{ID: 2, UserID: 9, Secret: "secret", Token: "token2"},
		}, nil)
		ou, err := m.mgr.GetByUserID(context.Background(), 9)
		m.Nil(err)
		m.Equal(encSecret, ou.Secret)
		m.Equal("secret1", ou.PlainSecret)
	}
}

func (m *metaMgrTestSuite) TestUpdateSecret() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*models.OIDCUser{
		{ID: 1, UserID: 9, Secret: encSecret, Token: "token1"},
	}, nil)
	m.dao.On("Update", mock.Anything, mock.Anything, "secret").Return(nil)
	err := m.mgr.SetCliSecretByUserID(context.Background(), 9, "new")
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &metaMgrTestSuite{})
}
