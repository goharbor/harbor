package dao

import (
	"testing"

	"github.com/magiconair/properties/assert"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	common_models "github.com/goharbor/harbor/src/common/models"

	"github.com/goharbor/harbor/src/retention/filter"

	"github.com/goharbor/harbor/src/common/retention"
	"github.com/goharbor/harbor/src/retention/dao/models"
	"github.com/stretchr/testify/require"
)

var (
	exampleServerPolicy = &models.Policy{
		Name:    "Example Server Policy",
		Enabled: true,

		Scope:             retention.ScopeServer,
		FallThroughAction: retention.KeepExtraTags,

		Filters: []*models.FilterMetadata{
			{Type: filter.TypeKeepMostRecentN, Options: map[string]interface{}{filter.MetaDataKeyN: 3}},
		},
	}

	exampleProjectPolicy = &models.Policy{
		Name:    "Example Project Policy",
		Enabled: true,

		Scope:   retention.ScopeProject,
		Project: &common_models.Project{ProjectID: 1},

		Filters: []*models.FilterMetadata{
			{Type: filter.TypeKeepRegex, Options: map[string]interface{}{filter.MetaDataKeyMatch: `^latest$`}},
			{Type: filter.TypeKeepMostRecentN, Options: map[string]interface{}{filter.MetaDataKeyN: 3}},
			{Type: filter.TypeDeleteEverything},
		},
	}

	exampleRepoPolicy = &models.Policy{
		Name:    "Example Repo Policy",
		Enabled: true,

		Scope:      retention.ScopeRepository,
		Project:    &common_models.Project{ProjectID: 1},
		Repository: &common_models.RepoRecord{RepositoryID: 1},

		Filters: []*models.FilterMetadata{
			{Type: filter.TypeKeepEverything},
		},
	}
)

func TestAddPolicy(t *testing.T) {
	require.NoError(t, common_dao.AddRepository(common_models.RepoRecord{
		ProjectID: 1,
		Name:      "RetentionTest",
	}))

	tests := []struct {
		Name       string
		Policy     *models.Policy
		ExpectedID int64
	}{
		{Name: "Server", ExpectedID: 1, Policy: exampleServerPolicy},
		{Name: "Project", ExpectedID: 2, Policy: exampleProjectPolicy},
		{Name: "Repo", ExpectedID: 3, Policy: exampleRepoPolicy},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			id, err := AddPolicy(tt.Policy)

			require.NoError(t, err)
			require.Equal(t, tt.ExpectedID, id)
		})
	}
}

func TestAddPolicy_SinglePolicyPerScope(t *testing.T) {
	tests := []struct {
		Name   string
		Policy *models.Policy
	}{
		{Name: "Server", Policy: exampleServerPolicy},
		{Name: "Project", Policy: exampleProjectPolicy},
		{Name: "Repo", Policy: exampleRepoPolicy},
	}

	for idx, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.Policy.ID = 0
			_, err := AddPolicy(tt.Policy)

			require.Error(t, err)
			tt.Policy.ID = int64(idx + 1)
		})
	}
}

func TestGetServerPolicy(t *testing.T) {
	p, err := GetServerPolicy()

	require.NoError(t, err)
	assertPolicyEffectivelyEqual(t, exampleServerPolicy, p)
}

func TestGetProjectPolicy(t *testing.T) {
	p, err := GetProjectPolicy(1)

	require.NoError(t, err)
	assertPolicyEffectivelyEqual(t, exampleProjectPolicy, p)
}

func TestGetRepoPolicy(t *testing.T) {
	p, err := GetRepoPolicy(1, 1)

	require.NoError(t, err)
	assertPolicyEffectivelyEqual(t, exampleRepoPolicy, p)
}

func TestUpdatePolicy(t *testing.T) {
	tests := []struct {
		Name       string
		Policy     *models.Policy
		UpdateFunc func(p *models.Policy)
	}{
		{Name: "Server Policy", Policy: exampleServerPolicy, UpdateFunc: func(p *models.Policy) {
			p.FallThroughAction = retention.DeleteExtraTags

			p.Filters[0].Options[filter.MetaDataKeyN] = 5
			p.Filters = append(p.Filters, &models.FilterMetadata{
				Type:    filter.TypeKeepRegex,
				Options: map[string]interface{}{filter.MetaDataKeyMatch: `^foo$`},
			})
		}},
		{Name: "Project Policy", Policy: exampleProjectPolicy, UpdateFunc: func(p *models.Policy) {
			p.FallThroughAction = retention.DeleteExtraTags
			p.Filters[0].Options[filter.MetaDataKeyMatch] = `^bar$`
			p.Filters = append(p.Filters[:1], p.Filters[2:]...)
		}},
		{Name: "Repo Policy", Policy: exampleRepoPolicy, UpdateFunc: func(p *models.Policy) {
			p.FallThroughAction = retention.DeleteExtraTags
		}},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tt.UpdateFunc(tt.Policy)
			require.NoError(t, UpdatePolicy(tt.Policy))

			var updatedPolicy *models.Policy
			var err error
			switch tt.Policy.Scope {
			case retention.ScopeServer:
				updatedPolicy, err = GetServerPolicy()
			case retention.ScopeProject:
				updatedPolicy, err = GetProjectPolicy(1)
			case retention.ScopeRepository:
				updatedPolicy, err = GetRepoPolicy(1, 1)
			}

			require.NoError(t, err)
			require.NotNil(t, updatedPolicy)

			assertPolicyEffectivelyEqual(t, tt.Policy, updatedPolicy)
		})
	}
}

func TestDeletePolicy(t *testing.T) {
	tests := []struct {
		Name   string
		Policy *models.Policy
	}{
		{Name: "Server Policy", Policy: exampleServerPolicy},
		{Name: "Project Policy", Policy: exampleProjectPolicy},
		{Name: "Repo Policy", Policy: exampleRepoPolicy},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			require.NoError(t, DeletePolicy(tt.Policy.ID))
		})
	}

	sp, err := GetServerPolicy()
	require.NoError(t, err)
	require.Nil(t, sp)

	pp, err := GetProjectPolicy(1)
	require.NoError(t, err)
	require.Nil(t, pp)

	rp, err := GetRepoPolicy(1, 1)
	require.NoError(t, err)
	require.Nil(t, rp)
}

// assertPolicyEffectivelyEqual compares most properties of two policies for equality.
//
// This is an easy way to check that policy representations are mostly equal
// while accounting for differences in encoding (like ints being float64 after
// deserialization or dates being in the wrong timezone).
func assertPolicyEffectivelyEqual(t *testing.T, expected, actual *models.Policy) {
	require.NotNil(t, expected)
	require.NotNil(t, actual)

	assert.Equal(t, actual.ID, expected.ID, "ID Not Equal")
	assert.Equal(t, actual.Name, expected.Name, "Name Not Equal")
	assert.Equal(t, actual.Enabled, expected.Enabled, "Enabled flag not equal")

	assert.Equal(t, actual.Scope, expected.Scope, "Scope not equal")
	assert.Equal(t, actual.FallThroughAction, expected.FallThroughAction, "FallThroughAction not equal")

	if expected.Project != nil {
		assert.Equal(t, actual.Project.ProjectID, expected.Project.ProjectID, "Project ID Not Equal")
	} else {
		require.Nil(t, actual.Project)
	}

	if expected.Repository != nil {
		assert.Equal(t, actual.Repository.RepositoryID, expected.Repository.RepositoryID, "Repository ID Not Equal")
	} else {
		require.Nil(t, actual.Repository)
	}

	if len(expected.Filters) == 0 {
		require.Empty(t, actual.Filters, "Filter list not empty")
		return
	}

	require.Equal(t, len(expected.Filters), len(actual.Filters), "Filter list different length")
	for i, ef := range expected.Filters {
		af := actual.Filters[i]

		assert.Equal(t, af.ID, ef.ID, "Filter ID not equal")
		assert.Equal(t, af.Type, ef.Type, "Filter type not equal")

		assert.Equal(t, af.RawOptions, ef.RawOptions, "Filter Options not equal")
	}
}
