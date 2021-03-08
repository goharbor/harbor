package immutable

import (
	"context"
	"github.com/goharbor/harbor/src/lib/q"
	dao_model "github.com/goharbor/harbor/src/pkg/immutable/dao/model"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
	"github.com/goharbor/harbor/src/testing/pkg/immutable/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"os"
	"testing"
)

type managerTestingSuite struct {
	suite.Suite
	t                *testing.T
	assert           *assert.Assertions
	require          *require.Assertions
	mockImmutableDao *dao.DAO
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
	m.mockImmutableDao = &dao.DAO{}
	Mgr = &defaultRuleManager{
		dao: m.mockImmutableDao,
	}
}

func TestManagerTestingSuite(t *testing.T) {
	suite.Run(t, &managerTestingSuite{})
}

func (m *managerTestingSuite) TestCreateImmutableRule() {
	m.mockImmutableDao.On("CreateImmutableRule", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := Mgr.CreateImmutableRule(context.Background(), &model.Metadata{})
	m.mockImmutableDao.AssertCalled(m.t, "CreateImmutableRule", mock.Anything, mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(int64(1), id)
}

func (m *managerTestingSuite) TestQueryImmutableRuleByProjectID() {
	m.mockImmutableDao.On("ListImmutableRules", mock.Anything, mock.Anything).Return([]*dao_model.ImmutableRule{
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
	irs, err := Mgr.ListImmutableRules(context.Background(), &q.Query{})
	m.mockImmutableDao.AssertCalled(m.t, "ListImmutableRules", mock.Anything, mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(len(irs), 2)
	m.assert.Equal(irs[1].Disabled, false)
}

func (m *managerTestingSuite) TestQueryEnabledImmutableRuleByProjectID() {
	m.mockImmutableDao.On("ListImmutableRules", mock.Anything, mock.Anything).Return([]*dao_model.ImmutableRule{
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
	irs, err := Mgr.ListImmutableRules(context.Background(), &q.Query{})
	m.mockImmutableDao.AssertCalled(m.t, "ListImmutableRules", mock.Anything, mock.Anything)
	m.require.Nil(err)
	m.assert.Equal(len(irs), 2)
	m.assert.Equal(irs[0].Disabled, true)
}

func (m *managerTestingSuite) TestGetImmutableRule() {
	m.mockImmutableDao.On("GetImmutableRule", mock.Anything, mock.Anything).Return(&dao_model.ImmutableRule{
		ID:        1,
		ProjectID: 1,
		Disabled:  true,
		TagFilter: "{\"id\":1, \"project_id\":1,\"priority\":0,\"disabled\":false,\"action\":\"immutable\"," +
			"\"template\":\"immutable_template\"," +
			"\"tag_selectors\":[{\"kind\":\"doublestar\",\"decoration\":\"matches\",\"pattern\":\"**\"}]," +
			"\"scope_selectors\":{\"repository\":[{\"kind\":\"doublestar\",\"decoration\":\"repoMatches\",\"pattern\":\"**\"}]}}",
	}, nil)
	ir, err := Mgr.GetImmutableRule(context.Background(), 1)
	m.mockImmutableDao.AssertCalled(m.t, "GetImmutableRule", mock.Anything, mock.Anything)
	m.require.Nil(err)
	m.require.NotNil(ir)
	m.assert.Equal(int64(1), ir.ID)
}

func (m *managerTestingSuite) TestUpdateImmutableRule() {
	m.mockImmutableDao.On("UpdateImmutableRule", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := Mgr.UpdateImmutableRule(context.Background(), int64(1), &model.Metadata{})
	m.mockImmutableDao.AssertCalled(m.t, "UpdateImmutableRule", mock.Anything, mock.Anything, mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestEnableImmutableRule() {
	m.mockImmutableDao.On("ToggleImmutableRule", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := Mgr.EnableImmutableRule(context.Background(), int64(1), true)
	m.mockImmutableDao.AssertCalled(m.t, "ToggleImmutableRule", mock.Anything, mock.Anything, mock.Anything)
	m.require.Nil(err)
}

func (m *managerTestingSuite) TestDeleteImmutableRule() {
	m.mockImmutableDao.On("DeleteImmutableRule", mock.Anything, mock.Anything).Return(nil)
	err := Mgr.DeleteImmutableRule(context.Background(), int64(1))
	m.mockImmutableDao.AssertCalled(m.t, "DeleteImmutableRule", mock.Anything, mock.Anything)
	m.require.Nil(err)
}
