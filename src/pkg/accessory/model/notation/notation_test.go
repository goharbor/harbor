package notation

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/accessory/model"
	htesting "github.com/goharbor/harbor/src/testing"
)

type NotationTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
	subDigest string
}

func (suite *NotationTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.subDigest = suite.DigestString()
	suite.accessory, _ = model.New(model.TypeNotationSignature,
		model.AccessoryData{
			ArtifactID:        1,
			SubArtifactDigest: suite.subDigest,
			Size:              4321,
			Digest:            suite.digest,
		})
}

func (suite *NotationTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetData().ID)
}

func (suite *NotationTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetData().ArtifactID)
}

func (suite *NotationTestSuite) TestSubGetArtID() {
	suite.Equal(suite.subDigest, suite.accessory.GetData().SubArtifactDigest)
}

func (suite *NotationTestSuite) TestSubGetSize() {
	suite.Equal(int64(4321), suite.accessory.GetData().Size)
}

func (suite *NotationTestSuite) TestSubGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetData().Digest)
}

func (suite *NotationTestSuite) TestSubGetType() {
	suite.Equal(model.TypeNotationSignature, suite.accessory.GetData().Type)
}

func (suite *NotationTestSuite) TestSubGetRefType() {
	suite.Equal(model.RefHard, suite.accessory.Kind())
}

func (suite *NotationTestSuite) TestIsSoft() {
	suite.False(suite.accessory.IsSoft())
}

func (suite *NotationTestSuite) TestIsHard() {
	suite.True(suite.accessory.IsHard())
}

func (suite *NotationTestSuite) TestDisplay() {
	suite.False(suite.accessory.Display())
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(NotationTestSuite))
}
