package project

import (
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/stretchr/testify/assert"
	"testing"
)

type fakeRobotContext struct {
	username   string
	isSysAdmin bool
}

var (
	robotCtx = &fakeRobotContext{username: "robot$tester", isSysAdmin: true}
)

func (ctx *fakeRobotContext) IsAuthenticated() bool {
	return ctx.username != ""
}

func (ctx *fakeRobotContext) GetUsername() string {
	return ctx.username
}

func (ctx *fakeRobotContext) IsSysAdmin() bool {
	return ctx.IsAuthenticated() && ctx.isSysAdmin
}

func (ctx *fakeRobotContext) GetPolicies() []*rbac.Policy {
	return nil
}

func TestGetPolicies(t *testing.T) {
	namespace := rbac.NewProjectNamespace("library", false)
	robot := NewRobot(robotCtx, namespace)
	assert.NotNil(t, robot.GetPolicies())
}
