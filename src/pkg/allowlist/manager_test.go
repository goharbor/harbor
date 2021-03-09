package allowlist

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/pkg/allowlist/models"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/allowlist/dao"
	"github.com/stretchr/testify/suite"
)

type mgrTestSuite struct {
	suite.Suite
	mgr Manager
	dao *dao.DAO
}

func (mt *mgrTestSuite) SetupTest() {
	mt.dao = &dao.DAO{}
	mt.mgr = &defaultManager{
		dao: mt.dao,
	}
}

func (mt *mgrTestSuite) TestSet() {
	mt.dao.On("Set", mock.Anything, models.CVEAllowlist{
		ProjectID: 9,
		Items: []models.CVEAllowlistItem{
			{
				CVEID: "testcve-1-1-1-1",
			},
		},
	}).Return(int64(9), nil)
	err := mt.mgr.Set(context.Background(), 9, models.CVEAllowlist{
		Items: []models.CVEAllowlistItem{
			{
				CVEID: "testcve-1-1-1-1",
			},
		},
	})
	mt.Nil(err)
	mt.dao.AssertExpectations(mt.T())
}

func (mt *mgrTestSuite) TestSetSys() {
	mt.dao.On("Set", mock.Anything, models.CVEAllowlist{
		ProjectID: 9,
		Items: []models.CVEAllowlistItem{
			{
				CVEID: "testcve-1-1-1-1",
			},
		},
	}).Return(int64(0), nil)
	err := mt.mgr.Set(context.Background(), 9, models.CVEAllowlist{
		Items: []models.CVEAllowlistItem{
			{
				CVEID: "testcve-1-1-1-1",
			},
		},
	})
	mt.Nil(err)
	mt.dao.AssertExpectations(mt.T())
}

func (mt *mgrTestSuite) TestGet() {
	mt.dao.On("QueryByProjectID", mock.Anything, int64(3)).Return(nil, nil)
	l, err := mt.mgr.Get(context.Background(), 3)
	mt.Nil(err)
	mt.Equal(models.CVEAllowlist{
		ProjectID: 3,
		Items:     []models.CVEAllowlistItem{},
	}, *l)
}

func (mt *mgrTestSuite) TestGetSys() {
	mt.dao.On("QueryByProjectID", mock.Anything, int64(0)).Return(&models.CVEAllowlist{
		ProjectID: 0,
		Items: []models.CVEAllowlistItem{
			{
				CVEID: "testcve-1-1-1-1",
			},
		},
	}, nil)
	l, err := mt.mgr.GetSys(context.Background())
	mt.Nil(err)
	mt.Equal(models.CVEAllowlist{
		ProjectID: 0,
		Items: []models.CVEAllowlistItem{
			{
				CVEID: "testcve-1-1-1-1",
			},
		},
	}, *l)
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &mgrTestSuite{})
}
