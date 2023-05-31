package gc

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
)

type callbackTestSuite struct {
	suite.Suite
	execMgr *tasktesting.ExecutionManager
	taskMgr *tasktesting.Manager
}

func (c *callbackTestSuite) SetupTest() {
	c.execMgr = &tasktesting.ExecutionManager{}
	c.taskMgr = &tasktesting.Manager{}
}

func (c *callbackTestSuite) TestCheckIn() {
	t := &task.Task{
		ID:     1,
		Status: "Success",
	}

	sc := &job.StatusChange{
		CheckIn: "",
	}

	c.taskMgr.On("Get", mock.Anything, int64(1)).Return(&task.Task{ID: 1, ExecutionID: 1}, nil)
	c.execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{ID: 1}, nil)
	c.execMgr.On("UpdateExtraAttrs", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	gcCheckIn(context.Background(), t, sc)
}

func TestCallBackTestSuite(t *testing.T) {
	suite.Run(t, &callbackTestSuite{})
}
