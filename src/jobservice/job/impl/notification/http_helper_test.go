package notification

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHttpHelper(t *testing.T) {
	c1 := httpHelper.clients[insecure]
	assert.NotNil(t, c1)
	assert.Equal(t, 3*time.Second, c1.Timeout)

	c2 := httpHelper.clients[secure]
	assert.NotNil(t, c2)
	assert.Equal(t, 3*time.Second, c1.Timeout)

	_, ok := httpHelper.clients["notExists"]
	assert.False(t, ok)
}

func TestSsrfProxyRoundTripper(t *testing.T) {
	t.Run("no proxy key", func(t *testing.T) {
		dummyRT := roundTripFunc(func(req *http.Request) (*http.Response, error) {
			// Unmodified request should have original URL host
			assert.Equal(t, "example.com", req.URL.Host)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("ok")),
			}, nil
		})

		rt := &ssrfProxyRoundTripper{
			insecure:   false,
			underlying: dummyRT,
		}

		req, err := http.NewRequest(http.MethodGet, "https://example.com/foo", nil)
		require.NoError(t, err)

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("with proxy key and private IP", func(t *testing.T) {
		rt := &ssrfProxyRoundTripper{
			insecure:   false,
			underlying: http.DefaultTransport,
		}

		req, err := http.NewRequest(http.MethodGet, "https://127.0.0.1/foo", nil)
		require.NoError(t, err)
		req = req.WithContext(context.WithValue(req.Context(), useProxyKey, true))

		_, err = rt.RoundTrip(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "blocked non-public host")
	})

	t.Run("with proxy key and public host", func(t *testing.T) {
		rt := &ssrfProxyRoundTripper{
			insecure:   false,
			underlying: http.DefaultTransport,
		}

		req, err := http.NewRequest(http.MethodGet, "https://example.com/foo", nil)
		require.NoError(t, err)
		req = req.WithContext(context.WithValue(req.Context(), useProxyKey, true))

		resp, err := rt.RoundTrip(req)
		if err == nil {
			defer resp.Body.Close()
			// Status can be 200 or 404 or any other HTTP status since it reached example.com
			assert.True(t, resp.StatusCode > 0)
		} else {
			// If network is not reachable (e.g. offline builder), we allow dial/connect errors,
			// but we shouldn't get validation errors.
			assert.True(t, strings.Contains(err.Error(), "dial") || strings.Contains(err.Error(), "connect") || strings.Contains(err.Error(), "lookup") || strings.Contains(err.Error(), "no such host") || strings.Contains(err.Error(), "timeout"))
		}
	})
}
