package query

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/suite"
)

// QueryTestSuite tests q
type QueryTestSuite struct {
	suite.Suite
}

// TestQueryTestSuite is entry of go test
func TestQueryTestSuite(t *testing.T) {
	suite.Run(t, new(QueryTestSuite))
}

// TestExtraParams tests extra parameters
func (suite *QueryTestSuite) TestExtraParams() {
	extras := make(ExtraParameters)
	extras.Set("a", 100)
	v, ok := extras.Get("a")

	assert.Equal(suite.T(), true, ok)
	assert.Equal(suite.T(), 100, v.(int))
}
