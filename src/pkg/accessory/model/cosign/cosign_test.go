package cosign

import (
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type CosignTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
}

func (suite *CosignTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.accessory, _ = model.New(model.TypeCosignSignature,
		model.AccessoryData{
			ArtifactID:    1,
			SubArtifactID: 2,
			Size:          4321,
			Digest:        suite.digest,
		})
}

func (suite *CosignTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetData().ID)
}

func (suite *CosignTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetData().ArtifactID)
}

func (suite *CosignTestSuite) TestSubGetArtID() {
	suite.Equal(int64(2), suite.accessory.GetData().SubArtifactID)
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
