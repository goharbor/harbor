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
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type NopCloseRequestTestSuite struct {
	suite.Suite
}

func (suite *NopCloseRequestTestSuite) TestReusableBody() {
	r, _ := http.NewRequest(http.MethodPost, "/", strings.NewReader("body"))

	body, err := ioutil.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte("body"), body)

	body, err = ioutil.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte(""), body)

	r, _ = http.NewRequest(http.MethodPost, "/", strings.NewReader("body"))
	r = NopCloseRequest(r)

	body, err = ioutil.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte("body"), body)

	body, err = ioutil.ReadAll(r.Body)
	suite.Nil(err)
	suite.Equal([]byte("body"), body)
}

func TestNopCloseRequestTestSuite(t *testing.T) {
	suite.Run(t, &NopCloseRequestTestSuite{})
}
