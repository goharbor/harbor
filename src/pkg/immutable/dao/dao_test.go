package dao

import (
	"github.com/goharbor/harbor/src/lib/q"
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/immutable/dao/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type immutableRuleDaoTestSuite struct {
	htesting.Suite
	require *require.Assertions
	assert  *assert.Assertions
	dao     DAO
	id      int64
}

func (t *immutableRuleDaoTestSuite) SetupSuite() {
	t.require = require.New(t.T())
	t.assert = assert.New(t.T())
	dao.PrepareTestForPostgresSQL()
	t.dao = New()
}

func (t *immutableRuleDaoTestSuite) TestCreateImmutableRule() {
	ir := &model.ImmutableRule{TagFilter: "**", ProjectID: 1}
	id, err := t.dao.CreateImmutableRule(t.Context(), ir)
	t.require.Nil(err)
	t.require.True(id > 0, "Can not create immutable tag rule")

	// insert duplicate rows
	ir2 := &model.ImmutableRule{TagFilter: "**", ProjectID: 1}
	id2, err := t.dao.CreateImmutableRule(t.Context(), ir2)
	t.require.True(strings.Contains(err.Error(), "immutable rule already exist"))
	t.require.Equal(int64(0), id2)

	err = t.dao.DeleteImmutableRule(t.Context(), id)
	t.require.Nil(err)
}

func (t *immutableRuleDaoTestSuite) TestUpdateImmutableRule() {
	ir := &model.ImmutableRule{TagFilter: "**", ProjectID: 1}
	id, err := t.dao.CreateImmutableRule(t.Context(), ir)
	t.require.Nil(err)
	t.require.True(id > 0, "Can not create immutable tag rule")

	updatedIR := &model.ImmutableRule{ID: id, TagFilter: "1.2.0", ProjectID: 1}
	err = t.dao.UpdateImmutableRule(t.Context(), 1, updatedIR)
	t.require.Nil(err)

	newIr, err := t.dao.GetImmutableRule(t.Context(), id)
	t.require.Nil(err)
	t.require.True(newIr.TagFilter == "1.2.0", "Failed to update immutable tag")

	defer t.dao.DeleteImmutableRule(t.Context(), id)

}

func (t *immutableRuleDaoTestSuite) TestEnableImmutableRule() {
	ir := &model.ImmutableRule{TagFilter: "**", ProjectID: 1}
	id, err := t.dao.CreateImmutableRule(t.Context(), ir)
	t.require.Nil(err)
	t.require.True(id > 0, "Can not create immutable tag rule")

	t.dao.ToggleImmutableRule(t.Context(), id, true)
	newIr, err := t.dao.GetImmutableRule(t.Context(), id)

	t.require.Nil(err)
	t.require.True(newIr.Disabled, "Failed to disable the immutable rule")

	defer t.dao.DeleteImmutableRule(t.Context(), id)
}

func (t *immutableRuleDaoTestSuite) TestGetImmutableRuleByProject() {
	irs := []*model.ImmutableRule{
		{TagFilter: "version1", ProjectID: 99},
		{TagFilter: "version2", ProjectID: 99},
		{TagFilter: "version3", ProjectID: 99},
		{TagFilter: "version4", ProjectID: 99},
	}
	for _, ir := range irs {
		t.dao.CreateImmutableRule(t.Context(), ir)
	}

	qrs, err := t.dao.ListImmutableRules(t.Context(), q.New(q.KeyWords{"ProjectID": 99}))
	t.require.Nil(err)
	t.require.True(len(qrs) == 4, "Failed to query 4 rows!")

	defer dao.ExecuteBatchSQL([]string{"delete from immutable_tag_rule where project_id = 99 "})

}
func (t *immutableRuleDaoTestSuite) TestGetEnabledImmutableRuleByProject() {
	irs := []*model.ImmutableRule{
		{TagFilter: "version1", ProjectID: 99},
		{TagFilter: "version2", ProjectID: 99},
		{TagFilter: "version3", ProjectID: 99},
		{TagFilter: "version4", ProjectID: 99},
	}
	for i, ir := range irs {
		id, _ := t.dao.CreateImmutableRule(t.Context(), ir)
		if i == 1 {
			t.dao.ToggleImmutableRule(t.Context(), id, true)
		}
	}

	qrs, err := t.dao.ListImmutableRules(t.Context(), q.New(q.KeyWords{"ProjectID": 99, "Disabled": "false"}))
	t.require.Nil(err)
	t.require.True(len(qrs) == 3, "Failed to query 3 rows!, got %v", len(qrs))

	defer dao.ExecuteBatchSQL([]string{"delete from immutable_tag_rule where project_id = 99 "})

}

func TestImmutableRuleDaoTestSuite(t *testing.T) {
	suite.Run(t, &immutableRuleDaoTestSuite{})
}
