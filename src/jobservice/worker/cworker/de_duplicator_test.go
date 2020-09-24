package cworker

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/tests"
)

// DeDuplicatorTestSuite tests functions of DeDuplicator
type DeDuplicatorTestSuite struct {
	suite.Suite
}

// TestDeDuplicatorTestSuite is entry of go test
func TestDeDuplicatorTestSuite(t *testing.T) {
	suite.Run(t, new(DeDuplicatorTestSuite))
}

// TestDeDuplicator ...
func (suite *DeDuplicatorTestSuite) TestDeDuplicator() {
	jobName := "fake_job"
	jobParams := map[string]interface{}{
		"image": "ubuntu:latest",
	}

	rdd := NewDeDuplicator(tests.GiveMeTestNamespace(), tests.GiveMeRedisPool())

	err := rdd.MustUnique(jobName, jobParams)
	require.NoError(suite.T(), err, "must unique 1st time: nil error expected but got %s", err)

	err = rdd.MustUnique(jobName, jobParams)
	assert.Error(suite.T(), err, "must unique 2nd time: non nil error expected but got nil")

	err = rdd.DelUniqueSign(jobName, jobParams)
	assert.NoError(suite.T(), err, "del unique: nil error expected but got %s", err)
}
