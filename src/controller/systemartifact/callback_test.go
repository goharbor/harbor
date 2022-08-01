package systemartifact

import (
	"context"
	"errors"
	"testing"

	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/controller/systemartifact"
	"github.com/goharbor/harbor/src/testing/mock"
)

type CallbackTestSuite struct {
	suite.Suite
	cleanupController *systemartifact.Controller
}

func (suite *CallbackTestSuite) SetupSuite() {
	suite.cleanupController = &systemartifact.Controller{}
	cleanupController = suite.cleanupController
}

func (suite *CallbackTestSuite) TestCleanupCallbackSuccess() {
	{
		ctx := context.TODO()
		suite.cleanupController.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(nil)
		triggerScheduleMatcher := testifymock.MatchedBy(func(trigger string) bool {
			return trigger == task.ExecutionTriggerSchedule
		})
		err := cleanupCallBack(ctx, "")
		suite.NoErrorf(err, "Unexpected error : %v", err)
		suite.cleanupController.AssertCalled(suite.T(), "Start", mock.Anything, true, triggerScheduleMatcher)
	}
	{
		suite.cleanupController = nil
		suite.cleanupController = &systemartifact.Controller{}
		cleanupController = suite.cleanupController
	}

	{
		ctx := context.TODO()
		suite.cleanupController.On("Start", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("test error"))
		triggerScheduleMatcher := testifymock.MatchedBy(func(trigger string) bool {
			return trigger == task.ExecutionTriggerSchedule
		})
		err := cleanupCallBack(ctx, "")
		suite.Error(err)
		suite.cleanupController.AssertCalled(suite.T(), "Start", mock.Anything, true, triggerScheduleMatcher)
	}

}

func TestCallbackTestSuite(t *testing.T) {
	suite.Run(t, &CallbackTestSuite{})
}
