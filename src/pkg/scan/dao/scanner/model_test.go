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

package scanner

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ModelTestSuite tests the utility functions of the model
type ModelTestSuite struct {
	suite.Suite
}

// TestModel is the entry of the model test suite
func TestModel(t *testing.T) {
	suite.Run(t, new(ModelTestSuite))
}

// TestJSON tests the marshal and unmarshal functions
func (suite *ModelTestSuite) TestJSON() {
	r := &Registration{
		Name:        "forUT",
		Description: "sample registration",
		URL:         "https://sample.scanner.com",
	}

	json, err := r.ToJSON()
	require.NoError(suite.T(), err)
	assert.Condition(suite.T(), func() (success bool) {
		success = len(json) > 0
		return
	})

	r2 := &Registration{}
	err = r2.FromJSON(json)
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), "forUT", r2.Name)
}

// TestValidate tests the validate function
func (suite *ModelTestSuite) TestValidate() {
	r := &Registration{}

	err := r.Validate(true)
	require.Error(suite.T(), err)

	r.UUID = "uuid"
	err = r.Validate(true)
	require.Error(suite.T(), err)

	r.Name = "forUT"
	err = r.Validate(true)
	require.Error(suite.T(), err)

	r.URL = "a.b.c"
	err = r.Validate(true)
	require.NoError(suite.T(), err)

	r.URL = "http://a.b.c"
	err = r.Validate(true)
	require.NoError(suite.T(), err)

	err = r.Validate(true)
	require.NoError(suite.T(), err)
}
