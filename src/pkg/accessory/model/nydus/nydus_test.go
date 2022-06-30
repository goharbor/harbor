package nydus

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/accessory/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
)

type NydusTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
}

func (suite *NydusTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.accessory, _ = model.New(model.TypeNydusAccelerator,
		model.AccessoryData{
			ArtifactID:    1,
			SubArtifactID: 2,
			Size:          4321,
			Digest:        suite.digest,
		})
}

func (suite *NydusTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetData().ID)
}

func (suite *NydusTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetData().ArtifactID)
}

func (suite *NydusTestSuite) TestSubGetArtID() {
	suite.Equal(int64(2), suite.accessory.GetData().SubArtifactID)
}

func (suite *NydusTestSuite) TestSubGetSize() {
	suite.Equal(int64(4321), suite.accessory.GetData().Size)
}

func (suite *NydusTestSuite) TestSubGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetData().Digest)
}

func (suite *NydusTestSuite) TestSubGetType() {
	suite.Equal(model.TypeNydusAccelerator, suite.accessory.GetData().Type)
}

func (suite *NydusTestSuite) TestSubGetRefType() {
	suite.Equal(model.RefHard, suite.accessory.Kind())
}

func (suite *NydusTestSuite) TestIsSoft() {
	suite.False(suite.accessory.IsSoft())
}

func (suite *NydusTestSuite) TestIsHard() {
	suite.True(suite.accessory.IsHard())
}

func (suite *NydusTestSuite) TestDisplay() {
	suite.False(suite.accessory.Display())
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(NydusTestSuite))
}
