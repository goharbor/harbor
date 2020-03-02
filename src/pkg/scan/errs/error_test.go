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

package errs

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ErrorSuite is a test suite for testing Error.
type ErrorSuite struct {
	suite.Suite
}

// TestError is the entry point of ErrorSuite.
func TestError(t *testing.T) {
	suite.Run(t, &ErrorSuite{})
}

// TestErrorNew ...
func (suite *ErrorSuite) TestErrorNew() {
	err := New("error-testing")
	require.Error(suite.T(), err)

	suite.Equal(true, AsError(err, Common))
	suite.Condition(func() (success bool) {
		return -1 != strings.Index(err.Error(), "error-testing")
	})
	suite.Condition(func() (success bool) {
		success = strings.Contains(err.Error(), codeTexts(Common))
		return
	})
}

// TestErrorWrap ...
func (suite *ErrorSuite) TestErrorWrap() {
	err := errors.New("error-stack")
	e := Wrap(err, "wrap-message")
	require.Error(suite.T(), e)

	suite.Equal(true, AsError(e, Common))
	suite.Condition(func() (success bool) {
		success = -1 != strings.Index(e.Error(), "error-stack")
		return
	})
	suite.Condition(func() (success bool) {
		success = -1 != strings.Index(e.Error(), "wrap-message")
		return
	})
	suite.Condition(func() (success bool) {
		success = strings.Contains(e.Error(), codeTexts(Common))
		return
	})
}

// TestErrorErrorf ...
func (suite *ErrorSuite) TestErrorErrorf() {
	err := Errorf("a=%d", 1000)
	require.Error(suite.T(), err)

	suite.Equal(true, AsError(err, Common))
	suite.Condition(func() (success bool) {
		success = strings.Contains(err.Error(), "a=1000")
		return
	})
	suite.Condition(func() (success bool) {
		success = strings.Contains(err.Error(), codeTexts(Common))
		return
	})
}

// TestErrorString ...
func (suite *ErrorSuite) TestErrorString() {
	err := New("well-formatted-error")
	require.Error(suite.T(), err)

	str := err.(*Error).String()
	require.Condition(suite.T(), func() (success bool) {
		success = len(str) > 0
		return
	})

	e := &Error{}
	er := json.Unmarshal([]byte(str), e)
	suite.NoError(er)
	suite.Equal(e.Message, "well-formatted-error")
}

// TestErrorWithCode ...
func (suite *ErrorSuite) TestErrorWithCode() {
	err := New("error-with-code")
	require.Error(suite.T(), err)

	err = WithCode(Conflict, err)
	require.Error(suite.T(), err)
	suite.Equal(true, AsError(err, Conflict))
	suite.Condition(func() (success bool) {
		success = strings.Contains(err.Error(), codeTexts(Conflict))
		return
	})
}
