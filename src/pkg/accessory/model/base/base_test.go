package base

import (
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type BaseTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
	subDigest string
}

func (suite *BaseTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.subDigest = suite.DigestString()
	suite.accessory, _ = model.New(model.TypeNone,
		model.AccessoryData{
			ArtifactID:    1,
			SubArtifactID: 2,
			Size:          1234,
			Digest:        suite.digest,
		})
}

func (suite *BaseTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetData().ID)
}

func (suite *BaseTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetData().ArtifactID)
}

func (suite *BaseTestSuite) TestSubGetArtID() {
	suite.Equal(int64(2), suite.accessory.GetData().SubArtifactID)
}

func (suite *BaseTestSuite) TestSubGetSize() {
	suite.Equal(int64(1234), suite.accessory.GetData().Size)
}

func (suite *BaseTestSuite) TestSubGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetData().Digest)
}

func (suite *BaseTestSuite) TestSubGetType() {
	suite.Equal(model.TypeNone, suite.accessory.GetData().Type)
}

func (suite *BaseTestSuite) TestSubGetRefType() {
	suite.Equal(model.RefNone, suite.accessory.Kind())
}

func (suite *BaseTestSuite) TestIsSoft() {
	suite.False(suite.accessory.IsSoft())
}

func (suite *BaseTestSuite) TestIsHard() {
	suite.False(suite.accessory.IsHard())
}

func (suite *BaseTestSuite) TestDisplay() {
	suite.False(suite.accessory.Display())
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(BaseTestSuite))
}
