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

package label

import (
	"fmt"
	"github.com/goharbor/harbor/src/pkg/retention/res"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// LabelSelectorTestSuite is a suite for testing the label selector
type LabelSelectorTestSuite struct {
	suite.Suite

	artifacts []*res.Candidate
}

// TestLabelSelector is entrance for LabelSelectorTestSuite
func TestLabelSelector(t *testing.T) {
	suite.Run(t, new(LabelSelectorTestSuite))
}

// SetupSuite to do preparation work
func (suite *LabelSelectorTestSuite) SetupSuite() {
	suite.artifacts = []*res.Candidate{
		{
			NamespaceID:  1,
			Namespace:    "library",
			Repository:   "harbor",
			Tag:          "1.9",
			Kind:         res.Image,
			PushedTime:   time.Now().Unix() - 3600,
			PulledTime:   time.Now().Unix(),
			CreationTime: time.Now().Unix() - 7200,
			Labels:       []string{"label1", "label2", "label3"},
		},
		{
			NamespaceID:  1,
			Namespace:    "library",
			Repository:   "harbor",
			Tag:          "dev",
			Kind:         res.Image,
			PushedTime:   time.Now().Unix() - 3600,
			PulledTime:   time.Now().Unix(),
			CreationTime: time.Now().Unix() - 7200,
			Labels:       []string{"label1", "label4", "label5"},
		},
	}
}

// TestWithLabelsUnMatched tests the selector of `with` labels but nothing matched
func (suite *LabelSelectorTestSuite) TestWithLabelsUnMatched() {
	withNothing := &selector{
		decoration: With,
		labels:     []string{"label6"},
	}

	selected, err := withNothing.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(selected))
}

// TestWithLabelsMatched tests the selector of `with` labels and matched something
func (suite *LabelSelectorTestSuite) TestWithLabelsMatched() {
	with1 := &selector{
		decoration: With,
		labels:     []string{"label2"},
	}

	selected, err := with1.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:1.9"}, selected)
	})

	with2 := &selector{
		decoration: With,
		labels:     []string{"label1"},
	}

	selected2, err := with2.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(selected2))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:1.9", "harbor:dev"}, selected2)
	})
}

// TestWithoutExistingLabels tests the selector of `without` existing labels
func (suite *LabelSelectorTestSuite) TestWithoutExistingLabels() {
	withoutExisting := &selector{
		decoration: Without,
		labels:     []string{"label1"},
	}

	selected, err := withoutExisting.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, len(selected))
}

// TestWithoutNoneExistingLabels tests the selector of `without` non-existing labels
func (suite *LabelSelectorTestSuite) TestWithoutNoneExistingLabels() {
	withoutNonExisting := &selector{
		decoration: Without,
		labels:     []string{"label6"},
	}

	selected, err := withoutNonExisting.Select(suite.artifacts)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, len(selected))
	assert.Condition(suite.T(), func() bool {
		return expect([]string{"harbor:1.9", "harbor:dev"}, selected)
	})
}

// Check whether the returned result matched the expected ones (only check repo:tag)
func expect(expected []string, candidates []*res.Candidate) bool {
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
