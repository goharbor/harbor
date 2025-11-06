package notification

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
)

func TestMatrixJobMaxFails(t *testing.T) {
	rep := &MatrixJob{}
	t.Run("default max fails", func(t *testing.T) {
		assert.Equal(t, uint(3), rep.MaxFails())
	})

	t.Run("user defined max fails", func(t *testing.T) {
		t.Setenv(maxFails, "15")
		assert.Equal(t, uint(15), rep.MaxFails())
	})

	t.Run("user defined wrong max fails", func(t *testing.T) {
		t.Setenv(maxFails, "abc")
		assert.Equal(t, uint(3), rep.MaxFails())
	})
}

func TestMatrixJobShouldRetry(t *testing.T) {
	rep := &MatrixJob{}
	assert.True(t, rep.ShouldRetry())
}

func TestMatrixJobValidate(t *testing.T) {
	rep := &MatrixJob{}
	assert.NotNil(t, rep.Validate(nil))

	jp := job.Parameters{
		"address": "https://matrix.org/_matrix/client/r0/rooms/!room:matrix.org/send/m.room.message",
		"payload": "matrix payload",
	}
	assert.Nil(t, rep.Validate(jp))
}

func TestMatrixJobRun(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(logger)

	rep := &MatrixJob{}

	// test matrix request
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)

			// test request method
			assert.Equal(t, http.MethodPost, r.Method)
			// test request body
			assert.Equal(t, string(body), `{"key": "value"}`)
		}))
	defer ts.Close()
	params := map[string]any{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          ts.URL,
	}
	// test correct matrix response
	assert.Nil(t, rep.Run(ctx, params))

	tsWrong := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
	defer tsWrong.Close()
	paramsWrong := map[string]any{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          tsWrong.URL,
	}
	// test incorrect matrix response
	assert.NotNil(t, rep.Run(ctx, paramsWrong))
}