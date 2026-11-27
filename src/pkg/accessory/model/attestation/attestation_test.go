package attestation

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/accessory/model"
)

type AttestationTestSuite struct {
	suite.Suite
	accessory model.Accessory
	digest    string
	subDigest string
}

func (suite *AttestationTestSuite) SetupSuite() {
	suite.digest = "sha256:1111111111111111111111111111111111111111111111111111111111111111"
	suite.subDigest = "sha256:2222222222222222222222222222222222222222222222222222222222222222"
	suite.accessory, _ = model.New(model.TypeInTotoAttestation, model.AccessoryData{
		ArtifactID:        1,
		SubArtifactDigest: suite.subDigest,
		Size:              1234,
		Digest:            suite.digest,
	})
}

func (suite *AttestationTestSuite) TestGetArtID() {
	suite.Equal(int64(1), suite.accessory.GetData().ArtifactID)
}

func (suite *AttestationTestSuite) TestGetDigest() {
	suite.Equal(suite.digest, suite.accessory.GetData().Digest)
}

func (suite *AttestationTestSuite) TestGetType() {
	suite.Equal(model.TypeInTotoAttestation, suite.accessory.GetData().Type)
}

func (suite *AttestationTestSuite) TestKind() {
	suite.Equal(model.RefHard, suite.accessory.Kind())
}

func (suite *AttestationTestSuite) TestIsHard() {
	suite.True(suite.accessory.IsHard())
	suite.False(suite.accessory.IsSoft())
}

func (suite *AttestationTestSuite) TestDisplay() {
	suite.False(suite.accessory.Display())
}

func TestAttestationTestSuite(t *testing.T) {
	suite.Run(t, new(AttestationTestSuite))
}

func TestInTotoTypeRegistered(t *testing.T) {
	acc, err := model.New(model.TypeInTotoAttestation, model.AccessoryData{})
	require.NoError(t, err)
	require.Equal(t, model.TypeInTotoAttestation, acc.GetData().Type)
}
