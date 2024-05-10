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

package sbom

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/accessory/model"
	htesting "github.com/goharbor/harbor/src/testing"
)

type SBOMTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
	subDigest string
}

func (suite *SBOMTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.subDigest = suite.DigestString()
	suite.accessory, _ = model.New(model.TypeHarborSBOM,
		model.AccessoryData{
			ArtifactID:        1,
			SubArtifactDigest: suite.subDigest,
			Size:              4321,
			Digest:            suite.digest,
		})
}

func (suite *SBOMTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetData().ID)
}

func (suite *SBOMTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetData().ArtifactID)
}

func (suite *SBOMTestSuite) TestSubGetArtID() {
	suite.Equal(suite.subDigest, suite.accessory.GetData().SubArtifactDigest)
}

func (suite *SBOMTestSuite) TestSubGetSize() {
	suite.Equal(int64(4321), suite.accessory.GetData().Size)
}

func (suite *SBOMTestSuite) TestSubGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetData().Digest)
}

func (suite *SBOMTestSuite) TestSubGetType() {
	suite.Equal(model.TypeHarborSBOM, suite.accessory.GetData().Type)
}

func (suite *SBOMTestSuite) TestSubGetRefType() {
	suite.Equal(model.RefHard, suite.accessory.Kind())
}

func (suite *SBOMTestSuite) TestIsSoft() {
	suite.False(suite.accessory.IsSoft())
}

func (suite *SBOMTestSuite) TestIsHard() {
	suite.True(suite.accessory.IsHard())
}

func (suite *SBOMTestSuite) TestDisplay() {
	suite.False(suite.accessory.Display())
}

func TestSBOMTestSuite(t *testing.T) {
	suite.Run(t, new(SBOMTestSuite))
}
