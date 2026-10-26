package gc

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRegistryRedisURLPassedToJob(t *testing.T) {
	const (
		oldURL = "redis://old-redis:6379/1"
		newURL = "redis://new-redis:6379/2"
	)
	schedule := func(extraAttrs string) string {
		return `{"trigger":null,"deleteuntagged":true,"deletetag":false,"dryrun":false,"workers":1,"extra_attrs":` + extraAttrs + `}`
	}

	cases := []struct {
		name    string
		env     string
		trigger string
		param   string
		policy  Policy
		want    any
	}{
		{
			name:    "scheduled run uses the current env over a stale stored url",
			env:     newURL,
			trigger: task.ExecutionTriggerSchedule,
			param:   schedule(`{"redis_url_reg":"` + oldURL + `","time_window":2}`),
			want:    newURL,
		},
		{
			name:    "scheduled run falls back to the stored url when env is empty",
			env:     "",
			trigger: task.ExecutionTriggerSchedule,
			param:   schedule(`{"redis_url_reg":"` + oldURL + `","time_window":2}`),
			want:    oldURL,
		},
		{
			name:    "scheduled run without stored extra attributes uses the current env",
			env:     newURL,
			trigger: task.ExecutionTriggerSchedule,
			param:   schedule(`null`),
			want:    newURL,
		},
		{
			name:    "manual run passes the url from the policy unchanged",
			env:     newURL,
			trigger: task.ExecutionTriggerManual,
			policy:  Policy{ExtraAttrs: map[string]any{"redis_url_reg": oldURL}},
			want:    oldURL,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("_REDIS_URL_REG", tc.env)

			execMgr := &tasktesting.ExecutionManager{}
			taskMgr := &tasktesting.Manager{}
			var gotTrigger string
			var gotParams map[string]any
			execMgr.On("Create", mock.Anything, job.GarbageCollectionVendorType, int64(-1), mock.Anything, mock.Anything).
				Run(func(args mock.Arguments) { gotTrigger = args.String(3) }).
				Return(int64(1), nil)
			taskMgr.On("Create", mock.Anything, int64(1), mock.Anything).
				Run(func(args mock.Arguments) { gotParams = args.Get(2).(*task.Job).Parameters }).
				Return(int64(1), nil)

			origCtl := Ctl
			Ctl = &controller{exeMgr: execMgr, taskMgr: taskMgr}
			t.Cleanup(func() { Ctl = origCtl })

			var err error
			if tc.trigger == task.ExecutionTriggerSchedule {
				err = gcCallback(context.Background(), tc.param)
			} else {
				_, err = Ctl.Start(context.Background(), tc.policy, tc.trigger)
			}
			require.NoError(t, err)
			assert.Equal(t, tc.trigger, gotTrigger)
			assert.Equal(t, tc.want, gotParams["redis_url_reg"])
		})
	}
}
