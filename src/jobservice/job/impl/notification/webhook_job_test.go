package notification

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func TestMaxFails(t *testing.T) {
	rep := &WebhookJob{}
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

func TestShouldRetry(t *testing.T) {
	rep := &WebhookJob{}
	assert.True(t, rep.ShouldRetry())
}

func TestValidate(t *testing.T) {
	rep := &WebhookJob{}
	assert.Nil(t, rep.Validate(nil))
}

func TestRun(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}

	ctx.On("GetLogger").Return(logger)

	rep := &WebhookJob{}

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
			// test request header
			assert.Equal(t, "auth_test", req.Header.Get("Authorization"))
			assert.Empty(t, req.Header.Get("Host"))
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
		"header":           `{"Authorization": ["auth_test"], "Host": ["metadata.google.internal"]}`,
	}
	// test correct webhook response
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
		"header":           `{"Authorization": ["auth_test"]}`,
	}
	// test incorrect webhook response
	assert.NotNil(t, rep.Run(ctx, paramsWrong))
}

func TestRunRejectsPrivateTarget(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	rep := &WebhookJob{}
	params := map[string]any{
		"payload": `{"key": "value"}`,
		"address": "http://169.254.169.254/latest/meta-data",
	}

	require.Error(t, rep.Run(ctx, params))
}

func TestNoRedirect(t *testing.T) {
	client := &http.Client{
		CheckRedirect: noRedirect,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: http.StatusFound,
				Body:       io.NopCloser(strings.NewReader("")),
				Header:     make(http.Header),
				Request:    req,
			}
			resp.Header.Set("Location", "http://1.1.1.2")
			return resp, nil
		}),
	}

	resp, err := client.Get("http://1.1.1.1")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusFound, resp.StatusCode)
}
