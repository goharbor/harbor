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

func (suite *ModelTestSuite) TestIsSysLevel() {
	r := Robot{
		Robot: model.Robot{
			ProjectID: 0,
		},
	}
	r.setLevel()
	suite.True(r.IsSysLevel())

	r = Robot{
		Robot: model.Robot{
			ProjectID: 1,
		},
	}
	r.setLevel()
	suite.False(r.IsSysLevel())
}

func (suite *ModelTestSuite) TestSetEditable() {
	r := Robot{
		Robot: model.Robot{
			ProjectID: 0,
		},
	}
	r.setEditable()
	suite.False(r.Editable)

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
	suite.True(r.Editable)
}

func (suite *ModelTestSuite) TestIsCoverAll() {
	p := &Permission{
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
		Scope: "/project/*",
	}
	suite.True(p.IsCoverAll())

	p.Scope = "/system"
	suite.False(p.IsCoverAll())
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, &ModelTestSuite{})
}
