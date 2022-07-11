package export

import (
	"crypto/sha256"
	"fmt"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
)

type DigestCalculatorTestSuite struct {
	suite.Suite
}

func (suite *DigestCalculatorTestSuite) TestDigestCalculation() {
	fileName := "/tmp/testfile.txt"
	data := []byte("test")
	ioutil.WriteFile(fileName, data, os.ModePerm)
	digestCalc := SHA256ArtifactDigestCalculator{}
	digest, err := digestCalc.Calculate(fileName)
	suite.NoError(err)

	hasher := sha256.New()
	hasher.Write(data)
	expectedDigest := fmt.Sprintf("sha256:%x", hasher.Sum(nil))
	suite.Equal(expectedDigest, digest.String())
}

func TestDigestCalculatorTestSuite(t *testing.T) {
	suite.Run(t, &DigestCalculatorTestSuite{})
}
