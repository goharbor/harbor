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

package image

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type indexProcessTestSuite struct {
	suite.Suite
	processor *indexProcessor
}

func (i *indexProcessTestSuite) SetupTest() {
	i.processor = &indexProcessor{}
}

func (i *indexProcessTestSuite) TestGetArtifactType() {
	i.Assert().Equal(ArtifactTypeImage, i.processor.GetArtifactType(nil, nil))
}

func TestIndexProcessTestSuite(t *testing.T) {
	suite.Run(t, &indexProcessTestSuite{})
}
