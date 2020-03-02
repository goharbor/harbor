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

package list

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ListSuite struct {
	suite.Suite

	l *SyncList
}

func TestListSuite(t *testing.T) {
	suite.Run(t, &ListSuite{})
}

func (suite *ListSuite) SetupSuite() {
	suite.l = New()

	suite.l.Push("a0")
	suite.l.Push("a1")
	suite.l.Push("b0")
	suite.l.Push("a2")

	suite.Equal(4, suite.l.l.Len())
}

func (suite *ListSuite) TestIterate() {
	suite.l.Iterate(func(ele interface{}) bool {
		if s, ok := ele.(string); ok {
			if strings.HasPrefix(s, "b") {
				return true
			}
		}

		return false
	})

	suite.Equal(3, suite.l.l.Len())
}
