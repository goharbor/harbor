package job

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/notification/job/model"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/notification/job/dao"
	"github.com/stretchr/testify/suite"
	"testing"
)

type managerTestSuite struct {
	suite.Suite
	mgr *manager
	dao *dao.DAO
}

func (m *managerTestSuite) SetupTest() {
	m.dao = &dao.DAO{}
	m.mgr = &manager{
		dao: m.dao,
	}
}

func (m *managerTestSuite) TestCreate() {
	m.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	_, err := m.mgr.Create(context.Background(), &model.Job{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestUpdate() {
	m.dao.On("Update", mock.Anything, mock.Anything).Return(nil)
	err := m.mgr.Update(context.Background(), &model.Job{})
	m.Nil(err)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestCount() {
	m.dao.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	n, err := m.mgr.Count(context.Background(), nil)
	m.Nil(err)
	m.Equal(int64(1), n)
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestList() {
	m.dao.On("List", mock.Anything, mock.Anything).Return([]*model.Job{
		{
			ID:        1,
			EventType: "test_job",
		},
	}, nil)
	rpers, err := m.mgr.List(context.Background(), nil)
	m.Nil(err)
	m.Equal(1, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func (m *managerTestSuite) TestListJobsGroupByEventType() {
	m.dao.On("GetLastTriggerJobsGroupByEventType", mock.Anything, mock.Anything).Return([]*model.Job{
		{
			ID:        1,
			EventType: "test_job",
			PolicyID:  1,
		},
		{
			ID:        2,
			EventType: "test_job",
			PolicyID:  1,
		},
	}, nil)
	rpers, err := m.mgr.ListJobsGroupByEventType(context.Background(), 1)
	m.Nil(err)
	m.Equal(2, len(rpers))
	m.dao.AssertExpectations(m.T())
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
