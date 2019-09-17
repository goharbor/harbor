// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package doublestar

import (
	"fmt"
	"github.com/goharbor/harbor/src/pkg/reselector"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// RegExpSelectorTestSuite is a suite for testing the label selector
type RegExpSelectorTestSuite struct {
	suite.Suite

	artifacts []*reselector.Candidate
}

// TestRegExpSelector is entrance for RegExpSelectorTestSuite
func TestRegExpSelector(t *testing.T) {
	suite.Run(t, new(RegExpSelectorTestSuite))
}

// SetupSuite to do preparation work
func (suite *RegExpSelectorTestSuite) SetupSuite() {
	suite.artifacts = []*reselector.Candidate{
		{
			NamespaceID:  1,
			Namespace:    "library",
			Repository:   "harbor",
			Tag:          "latest",
			Kind:         reselector.Image,
			PushedTime:   time.Now().Unix() - 3600,
			PulledTime:   time.Now().Unix(),
			CreationTime: time.Now().Unix() - 7200,
			Labels:       []string{"label1", "label2", "label3"},
		},
		{
			NamespaceID:  2,
			Namespace:    "retention",
			Repository:   "redis",
			Tag:          "4.0",
			Kind:         reselector.Image,
			PushedTime:   time.Now().Unix() - 3600,
			PulledTime:   time.Now().Unix(),
			CreationTime: time.Now().Unix() - 7200,
			Labels:       []string{"label1", "label4", "label5"},
		},
		{
			NamespaceID:  2,
			Namespace:    "retention",
			Repository:   "redis",
			Tag:          "4.1",
			Kind:         reselector.Image,
			PushedTime:   time.Now().Unix() - 3600,
			PulledTime:   time.Now().Unix(),
			CreationTime: time.Now().Unix() - 7200,
			Labels:       []string{"label1", "label4", "label5"},
		},
	}
}

// TestTagMatches tests the tag `matches` case
func (suite *RegExpSelectorTestSuite) TestTagMatches() {
	tagMatches := &selector{
		decoration: Matches,
		pattern:    "{latest,4.*}",
	}

	selected, err := tagMatches.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 3, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:latest", "redis:4.0", "redis:4.1"}, selected)
	})

	tagMatches2 := &selector{
		decoration: Matches,
		pattern:    "4.*",
	}

	selected, err = tagMatches2.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"redis:4.0", "redis:4.1"}, selected)
	})
}

// TestTagExcludes tests the tag `excludes` case
func (suite *RegExpSelectorTestSuite) TestTagExcludes() {
	tagExcludes := &selector{
		decoration: Excludes,
		pattern:    "{latest,4.*}",
	}

	selected, err := tagExcludes.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(selected))

	tagExcludes2 := &selector{
		decoration: Excludes,
		pattern:    "4.*",
	}

	selected, err = tagExcludes2.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:latest"}, selected)
	})
}

// TestRepoMatches tests the repository `matches` case
func (suite *RegExpSelectorTestSuite) TestRepoMatches() {
	repoMatches := &selector{
		decoration: RepoMatches,
		pattern:    "{redis}",
	}

	selected, err := repoMatches.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"redis:4.0", "redis:4.1"}, selected)
	})

	repoMatches2 := &selector{
		decoration: RepoMatches,
		pattern:    "har*",
	}

	selected, err = repoMatches2.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:latest"}, selected)
	})
}

// TestRepoExcludes tests the repository `excludes` case
func (suite *RegExpSelectorTestSuite) TestRepoExcludes() {
	repoExcludes := &selector{
		decoration: RepoExcludes,
		pattern:    "{redis}",
	}

	selected, err := repoExcludes.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:latest"}, selected)
	})

	repoExcludes2 := &selector{
		decoration: RepoExcludes,
		pattern:    "har*",
	}

	selected, err = repoExcludes2.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"redis:4.0", "redis:4.1"}, selected)
	})
}

// TestNSMatches tests the namespace `matches` case
func (suite *RegExpSelectorTestSuite) TestNSMatches() {
	repoMatches := &selector{
		decoration: NSMatches,
		pattern:    "{library}",
	}

	selected, err := repoMatches.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:latest"}, selected)
	})

	repoMatches2 := &selector{
		decoration: RepoMatches,
		pattern:    "re*",
	}

	selected, err = repoMatches2.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"redis:4.0", "redis:4.1"}, selected)
	})
}

// TestNSExcludes tests the namespace `excludes` case
func (suite *RegExpSelectorTestSuite) TestNSExcludes() {
	repoExcludes := &selector{
		decoration: NSExcludes,
		pattern:    "{library}",
	}

	selected, err := repoExcludes.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"redis:4.0", "redis:4.1"}, selected)
	})

	repoExcludes2 := &selector{
		decoration: NSExcludes,
		pattern:    "re*",
	}

	selected, err = repoExcludes2.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:latest"}, selected)
	})
}

// Check whether the returned result matched the expected ones (only check repo:tag)
func expect(expected []string, candidates []*reselector.Candidate) bool {
	hash := make(map[string]bool)

	for _, art := range candidates {
		hash[fmt.Sprintf("%s:%s", art.Repository, art.Tag)] = true
	}

	for _, exp := range expected {
		if _, ok := hash[exp]; !ok {
			return ok
		}
	}

	return true
}
