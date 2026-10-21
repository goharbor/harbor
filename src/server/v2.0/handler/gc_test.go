package handler

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/gc"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
)

func TestValidateWorkers(t *testing.T) {
	assert.False(t, validateWorkers(0))
	assert.False(t, validateWorkers(15))
	assert.False(t, validateWorkers(-1))
	assert.True(t, validateWorkers(1))
	assert.True(t, validateWorkers(5))
}

func authorizedGCContext() context.Context {
	sec := &securitytesting.Context{}
	sec.On("IsAuthenticated").Return(true)
	sec.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	return security.NewContext(context.Background(), sec)
}

func TestCreateGCSchedule_NilScheduleObj(t *testing.T) {
	api := &gcAPI{}
	params := operation.CreateGCScheduleParams{
		Schedule: &models.Schedule{Parameters: map[string]any{}},
	}

	resp := api.CreateGCSchedule(authorizedGCContext(), params)

	rec := httptest.NewRecorder()
	resp.WriteResponse(rec, nil)
	assert.Equal(t, 400, rec.Code)
}

func TestUpdateGCSchedule_NilScheduleObj(t *testing.T) {
	api := &gcAPI{}
	params := operation.UpdateGCScheduleParams{
		Schedule: &models.Schedule{Parameters: map[string]any{}},
	}

	resp := api.UpdateGCSchedule(authorizedGCContext(), params)

	rec := httptest.NewRecorder()
	resp.WriteResponse(rec, nil)
	assert.Equal(t, 400, rec.Code)
}
