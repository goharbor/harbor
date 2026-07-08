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

package lib

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
)

type NopCloseRequestTestSuite struct {
	suite.Suite
}

func (suite *NopCloseRequestTestSuite) TestReusableBody() {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("body"))

	body, err := io.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte("body"), body)

	body, err = io.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte(""), body)

	r, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader("body"))
	r = NopCloseRequest(r)

	body, err = io.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte("body"), body)

	body, err = io.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte("body"), body)
}

func TestNopCloseRequestTestSuite(t *testing.T) {
	suite.Run(t, &NopCloseRequestTestSuite{})
}

type ReadRequestBodyTestSuite struct {
	suite.Suite
}

func (suite *ReadRequestBodyTestSuite) TestWithinLimit() {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("body"))

	data, err := ReadRequestBody(r, 8)
	suite.Nil(err)
	suite.Equal([]byte("body"), data)

	// body is restored and re-readable for downstream consumers
	rest, err := io.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte("body"), rest)
}

func (suite *ReadRequestBodyTestSuite) TestAtLimit() {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("body"))

	data, err := ReadRequestBody(r, 4)
	suite.Nil(err)
	suite.Equal([]byte("body"), data)
}

func (suite *ReadRequestBodyTestSuite) TestOverLimit() {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("body"))

	data, err := ReadRequestBody(r, 3)
	suite.Nil(data)
	suite.True(errors.IsErr(err, errors.RequestEntityTooLargeCode))
}

func (suite *ReadRequestBodyTestSuite) TestUnbounded() {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("body"))

	data, err := ReadRequestBody(r, 0)
	suite.Nil(err)
	suite.Equal([]byte("body"), data)
}

func TestReadRequestBodyTestSuite(t *testing.T) {
	suite.Run(t, &ReadRequestBodyTestSuite{})
}
