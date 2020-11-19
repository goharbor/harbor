package robot

import (
	"github.com/goharbor/harbor/src/pkg/robot2/model"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ModelTestSuite struct {
	suite.Suite
}

func (suite *ModelTestSuite) TestSetLevel() {
	r := Robot{
		Robot: model.Robot{
			ProjectID: 0,
		},
	}
	r.setLevel()

	suite.Equal(LEVELSYSTEM, r.Level)

	r = Robot{
		Robot: model.Robot{
			ProjectID: 1,
		},
	}
	r.setLevel()
	suite.Equal(LEVELPROJECT, r.Level)
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, &ModelTestSuite{})
}
