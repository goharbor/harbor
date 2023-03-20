package subject

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/accessory/model"
	htesting "github.com/goharbor/harbor/src/testing"
)

type SubjectTestSuite struct {
	htesting.Suite
	accessory model.Accessory
	digest    string
	subDigest string
}

func (suite *SubjectTestSuite) SetupSuite() {
	suite.digest = suite.DigestString()
	suite.subDigest = suite.DigestString()
	suite.accessory, _ = model.New(model.TypeSubject,
		model.AccessoryData{
			ArtifactID:        1,
			SubArtifactDigest: suite.subDigest,
			Size:              4321,
			Digest:            suite.digest,
		})
}

func (suite *SubjectTestSuite) TestGetID() {
	suite.Equal(int64(0), suite.accessory.GetData().ID)
}

func (suite *SubjectTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetData().ArtifactID)
}

func (suite *SubjectTestSuite) TestSubGetArtID() {
	suite.Equal(suite.subDigest, suite.accessory.GetData().SubArtifactDigest)
}

func (suite *SubjectTestSuite) TestSubGetSize() {
	suite.Equal(int64(4321), suite.accessory.GetData().Size)
}

func (suite *SubjectTestSuite) TestSubGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetData().Digest)
}

func (suite *SubjectTestSuite) TestSubGetType() {
	suite.Equal(model.TypeSubject, suite.accessory.GetData().Type)
}

func (suite *SubjectTestSuite) TestSubGetRefType() {
	suite.Equal(model.RefHard, suite.accessory.Kind())
}

func (suite *SubjectTestSuite) TestIsSoft() {
	suite.False(suite.accessory.IsSoft())
}

func (suite *SubjectTestSuite) TestIsHard() {
	suite.True(suite.accessory.IsHard())
}

func (suite *SubjectTestSuite) TestDisplay() {
	suite.False(suite.accessory.Display())
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, new(SubjectTestSuite))
}
