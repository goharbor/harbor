package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
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

func TestGCScheduleRequired(t *testing.T) {
	secCtx := &securitytesting.Context{}
	secCtx.On("IsAuthenticated").Return(true)
	secCtx.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	ctx := security.NewContext(context.Background(), secCtx)
	api := &gcAPI{}

	for _, body := range []*models.Schedule{nil, {Parameters: map[string]any{"delete_untagged": true}}} {
		rr := httptest.NewRecorder()
		api.CreateGCSchedule(ctx, operation.CreateGCScheduleParams{
			HTTPRequest: httptest.NewRequest(http.MethodPost, "/api/v2.0/system/gc/schedule", nil),
			Schedule:    body,
		}).WriteResponse(rr, runtime.JSONProducer())
		assert.Equal(t, http.StatusBadRequest, rr.Code)

		rr = httptest.NewRecorder()
		api.UpdateGCSchedule(ctx, operation.UpdateGCScheduleParams{
			HTTPRequest: httptest.NewRequest(http.MethodPut, "/api/v2.0/system/gc/schedule", nil),
			Schedule:    body,
		}).WriteResponse(rr, runtime.JSONProducer())
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	}
}
