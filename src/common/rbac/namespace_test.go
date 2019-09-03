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

package rbac

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProjectNamespaceTestSuite struct {
	suite.Suite
}

func (suite *ProjectNamespaceTestSuite) TestResource() {
	var namespace Namespace

	namespace = &projectNamespace{projectID: int64(1)}

	suite.Equal(namespace.Resource(Resource("image")), Resource("/project/1/image"))
}

func (suite *ProjectNamespaceTestSuite) TestIdentity() {
	namespace, _ := Resource("/project/1/image").GetNamespace()
	suite.Equal(namespace.Identity(), int64(1))
}

func TestProjectNamespaceTestSuite(t *testing.T) {
	suite.Run(t, new(ProjectNamespaceTestSuite))
}
