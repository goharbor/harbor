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

func TestTeamsJobMaxFails(t *testing.T) {
	rep := &TeamsJob{}
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

func TestTeamsJobShouldRetry(t *testing.T) {
	rep := &TeamsJob{}
	assert.True(t, rep.ShouldRetry())
}

func TestTeamsJobValidate(t *testing.T) {
	rep := &TeamsJob{}
	assert.NotNil(t, rep.Validate(nil))

	jp := job.Parameters{
		"address": "https://mydomain.webhook.office.com/webhookb2/akshat123/IncomingWebhook/akshat456/akshat789",
		"payload": "teams payload",
	}
	assert.Nil(t, rep.Validate(jp))
}

func TestTeamsJobRun(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(logger)

	rep := &TeamsJob{}

	// test teams request
	ts := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)

			// test request method
			assert.Equal(t, http.MethodPost, r.Method)
			// test request body
			assert.Equal(t, string(body), `{"key": "value"}`)
		}))
	defer ts.Close()
	params := map[string]interface{}{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          ts.URL,
	}
	// test correct teams response
	assert.Nil(t, rep.Run(ctx, params))

	tsWrong := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
	defer tsWrong.Close()
	paramsWrong := map[string]interface{}{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          tsWrong.URL,
	}
	// test incorrect teams response
	assert.NotNil(t, rep.Run(ctx, paramsWrong))
}
