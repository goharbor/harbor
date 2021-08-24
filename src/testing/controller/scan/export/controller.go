package export

import (
	"context"
	"github.com/goharbor/harbor/src/pkg/scan/export"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/mock"
	"time"
)

// Controller exposes a mock implementation of the scan data export controller
type Controller struct {
	mock.Mock
}

func (_m *Controller) Start(ctx context.Context, criteria export.Criteria) (executionId int64, err error) {
	ret := _m.Called(ctx, criteria)
	return ret.Get(0).(int64), nil
}

func (_m *Controller) GetExecution(ctx context.Context, executionId int64) (*export.Execution, error) {
	ret := _m.Called(ctx, executionId)
	return ret.Get(0).(*export.Execution), nil

}

func (_m *Controller) StartCleanup(ctx context.Context) {
	_m.Called(ctx)
}

func (_m *Controller) GetTask(ctx context.Context, executionId int64) (*task.Task, error) {
	return &task.Task{
		ID:             2,
		VendorType:     "SCAN_DATA_EXPORT",
		ExecutionID:    1,
		Status:         "Success",
		StatusMessage:  "",
		RunCount:       1,
		JobID:          "JobId",
		ExtraAttrs:     nil,
		CreationTime:   time.Time{},
		StartTime:      time.Time{},
		UpdateTime:     time.Time{},
		EndTime:        time.Time{},
		StatusRevision: 0,
	}, nil
}

// Manager is a mock implementation of scan data export manager

type Manager struct {
	mock.Mock
}

func (_m *Manager) Fetch(ctx context.Context, params export.Params) ([]export.Data, error) {
	ret := _m.Called(ctx, params)
	return ret.Get(0).([]export.Data), ret.Error(1)
}

type DigestCalculator struct {
	mock.Mock
}

func (_m *DigestCalculator) Calculate(fileName string) (digest.Digest, error) {
	ret := _m.Called(fileName)
	return ret.Get(0).(digest.Digest), ret.Error(1)
}

type CleanupManager struct {
	mock.Mock
}

func (_m *CleanupManager) Configure(settings *export.CleanupSettings) {
	_m.Called(settings)
}

func (_m *CleanupManager) Execute(ctx context.Context) error {
	ret := _m.Called(ctx)
	return ret.Error(0)
}
