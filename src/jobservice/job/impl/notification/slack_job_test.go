package notification

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
)

func TestSlackJobMaxFails(t *testing.T) {
	rep := &SlackJob{}
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

func TestSlackJobShouldRetry(t *testing.T) {
	rep := &SlackJob{}
	assert.True(t, rep.ShouldRetry())
}

func TestSlackJobValidate(t *testing.T) {
	rep := &SlackJob{}
	assert.NotNil(t, rep.Validate(nil))

	jp := job.Parameters{
		"address": "https://webhook.slack.com/hsdouihhsd988",
		"payload": "slack payload",
	}
	assert.Nil(t, rep.Validate(jp))
}

func TestSlackJobRun(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(logger)

	rep := &SlackJob{}

	originalClients := httpHelper.clients
	t.Cleanup(func() {
		httpHelper.clients = originalClients
	})
	httpHelper.clients = map[string]*http.Client{}
	httpHelper.clients[secure] = &http.Client{
		CheckRedirect: noRedirect,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			body, _ := io.ReadAll(req.Body)

			// test request method
			assert.Equal(t, http.MethodPost, req.Method)
			// test request body
			assert.Equal(t, string(body), `{"key": "value"}`)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}
	httpHelper.clients[insecure] = httpHelper.clients[secure]

	params := map[string]any{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          "http://1.1.1.1",
	}
	// test correct slack response
	assert.Nil(t, rep.Run(ctx, params))

	httpHelper.clients[insecure] = &http.Client{
		CheckRedirect: noRedirect,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
				Request:    req,
			}, nil
		}),
	}
	paramsWrong := map[string]any{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          "http://1.1.1.1",
	}
	// test incorrect slack response
	assert.NotNil(t, rep.Run(ctx, paramsWrong))
}

func TestSlackJobRunRejectsPrivateTarget(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	rep := &SlackJob{}
	params := map[string]any{
		"payload": `{"key": "value"}`,
		"address": "http://169.254.169.254/latest/meta-data",
	}

	assert.NotNil(t, rep.Run(ctx, params))
}
