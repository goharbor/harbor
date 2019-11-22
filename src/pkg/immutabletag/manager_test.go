package immutabletag

import (
	dao_model "github.com/goharbor/harbor/src/pkg/immutabletag/dao/model"
	"github.com/goharbor/harbor/src/pkg/immutabletag/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"os"
	"testing"
)

type mockImmutableDao struct {
	mock.Mock
}

func (m *mockImmutableDao) CreateImmutableRule(ir *dao_model.ImmutableRule) (int64, error) {
	args := m.Called(ir)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockImmutableDao) UpdateImmutableRule(projectID int64, ir *dao_model.ImmutableRule) (int64, error) {
	args := m.Called(ir)
	return int64(0), args.Error(1)
}

func (m *mockImmutableDao) QueryImmutableRuleByProjectID(projectID int64) ([]dao_model.ImmutableRule, error) {
	args := m.Called()
	var irs []dao_model.ImmutableRule
	if args.Get(0) != nil {
		irs = args.Get(0).([]dao_model.ImmutableRule)
	}
	return irs, args.Error(1)
}

func (m *mockImmutableDao) QueryEnabledImmutableRuleByProjectID(projectID int64) ([]dao_model.ImmutableRule, error) {
	args := m.Called()
	var irs []dao_model.ImmutableRule
	if args.Get(0) != nil {
		irs = args.Get(0).([]dao_model.ImmutableRule)
	}
	return irs, args.Error(1)
}

func (m *mockImmutableDao) DeleteImmutableRule(id int64) (int64, error) {
	args := m.Called(id)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockImmutableDao) ToggleImmutableRule(id int64, enabled bool) (int64, error) {
	args := m.Called(id)
	return int64(args.Int(0)), args.Error(1)
}

func (m *mockImmutableDao) GetImmutableRule(id int64) (*dao_model.ImmutableRule, error) {
	args := m.Called(id)
	var ir *dao_model.ImmutableRule
	if args.Get(0) != nil {
		ir = args.Get(0).(*dao_model.ImmutableRule)
	}
	return ir, args.Error(1)

}

type managerTestingSuite struct {
	suite.Suite
	t                *testing.T
	assert           *assert.Assertions
	require          *require.Assertions
	mockImmutableDao *mockImmutableDao
}

func (m *managerTestingSuite) SetupSuite() {
	m.t = m.T()
	m.assert = assert.New(m.t)
	m.require = require.New(m.t)

	err := os.Setenv("RUN_MODE", "TEST")
	m.require.Nil(err)
}

func (m *managerTestingSuite) TearDownSuite() {
	err := os.Unsetenv("RUN_MODE")
	m.require.Nil(err)
}

func (m *managerTestingSuite) SetupTest() {
	m.mockImmutableDao = &mockImmutableDao{}
	Mgr = &defaultRuleManager{
		dao: m.mockImmutableDao,
	}
}

func TestManagerTestingSuite(t *testing.T) {
	suite.Run(t, &managerTestingSuite{})
}

func (m *managerTestingSuite) TestCreateImmutableRule() {
	m.mockImmutableDao.On("CreateImmutableRule", mock.Anything).Return(1, nil)
	id, err := Mgr.CreateImmutableRule(&model.Metadata{})
	m.mockImmutableDao.AssertCalled(m.t, "CreateImmutableRule", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}

func (m *managerTestingSuite) TestQueryImmutableRuleByProjectID() {
	m.mockImmutableDao.On("QueryImmutableRuleByProjectID", mock.Anything).Return([]dao_model.ImmutableRule{
		{
			ID:        1,
			ProjectID: 1,
			Disabled:  false,
			TagFilter: "{\"id\":1, \"project_id\":1,\"priority\":0,\"disabled\":false,\"action\":\"immutable\"," +
				"\"template\":\"immutable_template\"," +
				"\"tag_selectors\":[{\"kind\":\"doublestar\",\"decoration\":\"matches\",\"pattern\":\"**\"}]," +
				"\"scope_selectors\":{\"repository\":[{\"kind\":\"doublestar\",\"decoration\":\"repoMatches\",\"pattern\":\"**\"}]}}",
		},
		{
			ID:        2,
			ProjectID: 1,
			Disabled:  false,
			TagFilter: "{\"id\":2, \"project_id\":1,\"priority\":0,\"disabled\":false,\"action\":\"immutable\"," +
				"\"template\":\"immutable_template\"," +
				"\"tag_selectors\":[{\"kind\":\"doublestar\",\"decoration\":\"matches\",\"pattern\":\"**\"}]," +
				"\"scope_selectors\":{\"repository\":[{\"kind\":\"doublestar\",\"decoration\":\"repoMatches\",\"pattern\":\"**\"}]}}",
		}}, nil)
	irs, err := Mgr.QueryImmutableRuleByProjectID(int64(1))
	m.mockImmutableDao.AssertCalled(m.t, "QueryImmutableRuleByProjectID", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(len(irs), 2)
	m.assert.Equal(irs[1].Disabled, false)
}

func (m *managerTestingSuite) TestQueryEnabledImmutableRuleByProjectID() {
	m.mockImmutableDao.On("QueryEnabledImmutableRuleByProjectID", mock.Anything).Return([]dao_model.ImmutableRule{
		{
			ID:        1,
			ProjectID: 1,
			Disabled:  true,
			TagFilter: "{\"id\":1, \"project_id\":1,\"priority\":0,\"disabled\":false,\"action\":\"immutable\"," +
				"\"template\":\"immutable_template\"," +
				"\"tag_selectors\":[{\"kind\":\"doublestar\",\"decoration\":\"matches\",\"pattern\":\"**\"}]," +
				"\"scope_selectors\":{\"repository\":[{\"kind\":\"doublestar\",\"decoration\":\"repoMatches\",\"pattern\":\"**\"}]}}",
		},
		{
			ID:        2,
			ProjectID: 1,
			Disabled:  true,
			TagFilter: "{\"id\":2, \"project_id\":1,\"priority\":0,\"disabled\":false,\"action\":\"immutable\"," +
				"\"template\":\"immutable_template\"," +
				"\"tag_selectors\":[{\"kind\":\"doublestar\",\"decoration\":\"matches\",\"pattern\":\"**\"}]," +
				"\"scope_selectors\":{\"repository\":[{\"kind\":\"doublestar\",\"decoration\":\"repoMatches\",\"pattern\":\"**\"}]}}",
		}}, nil)
	irs, err := Mgr.QueryEnabledImmutableRuleByProjectID(int64(1))
	m.mockImmutableDao.AssertCalled(m.t, "QueryEnabledImmutableRuleByProjectID", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(len(irs), 2)
	m.assert.Equal(irs[0].Disabled, false)
}

func (m *managerTestingSuite) TestGetImmutableRule() {
	m.mockImmutableDao.On("GetImmutableRule", mock.Anything).Return(&dao_model.ImmutableRule{
		ID:        1,
		ProjectID: 1,
		Disabled:  true,
		TagFilter: "{\"id\":1, \"project_id\":1,\"priority\":0,\"disabled\":false,\"action\":\"immutable\"," +
			"\"template\":\"immutable_template\"," +
			"\"tag_selectors\":[{\"kind\":\"doublestar\",\"decoration\":\"matches\",\"pattern\":\"**\"}]," +
			"\"scope_selectors\":{\"repository\":[{\"kind\":\"doublestar\",\"decoration\":\"repoMatches\",\"pattern\":\"**\"}]}}",
	}, nil)
	ir, err := Mgr.GetImmutableRule(1)
	m.mockImmutableDao.AssertCalled(m.t, "GetImmutableRule", mock.Anything)
	m.require.Nil(err)
	m.require.NotNil(ir)
	m.assert.Equal(int64(1), ir.ID)
}

func (m *managerTestingSuite) TestUpdateImmutableRule() {
	m.mockImmutableDao.On("UpdateImmutableRule", mock.Anything).Return(1, nil)
	id, err := Mgr.UpdateImmutableRule(int64(1), &model.Metadata{})
	m.mockImmutableDao.AssertCalled(m.t, "UpdateImmutableRule", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(0), id)
}

func (m *managerTestingSuite) TestEnableImmutableRule() {
	m.mockImmutableDao.On("ToggleImmutableRule", mock.Anything).Return(1, nil)
	id, err := Mgr.EnableImmutableRule(int64(1), true)
	m.mockImmutableDao.AssertCalled(m.t, "ToggleImmutableRule", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}

func (m *managerTestingSuite) TestDeleteImmutableRule() {
	m.mockImmutableDao.On("DeleteImmutableRule", mock.Anything).Return(1, nil)
	id, err := Mgr.DeleteImmutableRule(int64(1))
	m.mockImmutableDao.AssertCalled(m.t, "DeleteImmutableRule", mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}
