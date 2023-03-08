package cosign

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/accessory/model"
	htesting "github.com/goharbor/harbor/src/testing"
)

type CosignTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
	subDigest string
}

func (suite *CosignTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.subDigest = suite.DigestString()
	suite.accessory, _ = model.New(model.TypeCosignSignature,
		model.AccessoryData{
			ArtifactID:        1,
			SubArtifactDigest: suite.subDigest,
			Size:              4321,
			Digest:            suite.digest,
		})
}

func (suite *CosignTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetData().ID)
}

func (suite *CosignTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetData().ArtifactID)
}

func (suite *CosignTestSuite) TestSubGetArtID() {
	suite.Equal(suite.subDigest, suite.accessory.GetData().SubArtifactDigest)
}

func (suite *CosignTestSuite) TestSubGetSize() {
	suite.Equal(int64(4321), suite.accessory.GetData().Size)
}

func (suite *CosignTestSuite) TestSubGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetData().Digest)
}

func (suite *CosignTestSuite) TestSubGetType() {
	suite.Equal(model.TypeCosignSignature, suite.accessory.GetData().Type)
}

func (suite *CosignTestSuite) TestSubGetRefType() {
	suite.Equal(model.RefHard, suite.accessory.Kind())
}

func (suite *CosignTestSuite) TestIsSoft() {
	suite.False(suite.accessory.IsSoft())
}

func (suite *CosignTestSuite) TestIsHard() {
	suite.True(suite.accessory.IsHard())
}

func (suite *CosignTestSuite) TestDisplay() {
	suite.False(suite.accessory.Display())
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(CosignTestSuite))
}
