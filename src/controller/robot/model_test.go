package robot

import (
	"github.com/goharbor/harbor/src/pkg/permission/types"
	"github.com/goharbor/harbor/src/pkg/robot/model"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ModelTestSuite struct {
	htesting.Suite
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

func (suite *ModelTestSuite) TestSetEditable() {
	r := Robot{
		Robot: model.Robot{
			ProjectID: 0,
		},
	}
	r.setEditable()
	suite.Equal(false, r.Editable)

	r = Robot{
		Robot: model.Robot{
			Name:        "testcreate",
			Description: "testcreate",
			Duration:    0,
		},
		ProjectName: "library",
		Level:       LEVELPROJECT,
		Permissions: []*Permission{
			{
				Kind:      "project",
				Namespace: "library",
				Access: []*types.Policy{
					{
						Resource: "repository",
						Action:   "push",
					},
					{
						Resource: "repository",
						Action:   "pull",
					},
				},
			},
		},
	}
	r.setEditable()
	suite.Equal(true, r.Editable)
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, &ModelTestSuite{})
}
