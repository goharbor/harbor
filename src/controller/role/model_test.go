package role

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/role/model"
)

type ModelTestSuite struct {
	suite.Suite
}

func (suite *ModelTestSuite) TestSetLevel() {
	r := &Role{Role: model.Role{ID: 1, Name: "myCustomRole"}}
	r.setLevel()
	suite.Equal(LEVELROLE, r.Level)
}

func (suite *ModelTestSuite) TestSetEditable() {
	r := &Role{Role: model.Role{ID: 1, Name: "myCustomRole"}}
	suite.False(r.Editable)
	r.setEditable()
	suite.True(r.Editable)
}

func TestModelTestSuite(t *testing.T) {
	suite.Run(t, &ModelTestSuite{})
}
