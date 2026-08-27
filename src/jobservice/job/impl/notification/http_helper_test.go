package notification

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib"
)

type mockDialer struct {
	dialFunc func(ctx context.Context, network, address string) (net.Conn, error)
}

func (m *mockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return m.dialFunc(ctx, network, address)
}

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
		// Mock DNS resolution to return a public IP
		cleanupDNS := lib.SetLookupIPAddrForTest(func(ctx context.Context, host string) ([]net.IPAddr, error) {
			if host == "example.com" {
				return []net.IPAddr{{IP: net.ParseIP("1.1.1.1")}}, nil
			}
			return net.DefaultResolver.LookupIPAddr(ctx, host)
		})
		defer cleanupDNS()

		// Start local TLS test server with self-signed certificate
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		}))
		defer server.Close()

		// Mock dialer to redirect connection to the local test server
		originalDialer := dialer
		t.Cleanup(func() {
			dialer = originalDialer
		})
		dialer = &mockDialer{
			dialFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
			},
		}

		rt := &ssrfProxyRoundTripper{
			insecure:   true, // skip verification for self-signed test server cert
			underlying: http.DefaultTransport,
		}

		req, err := http.NewRequest(http.MethodGet, "https://example.com/foo", nil)
		require.NoError(t, err)
		req = req.WithContext(context.WithValue(req.Context(), useProxyKey, true))

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("response body stays readable after RoundTrip returns", func(t *testing.T) {
		// Mock DNS resolution to return a public IP
		cleanupDNS := lib.SetLookupIPAddrForTest(func(ctx context.Context, host string) ([]net.IPAddr, error) {
			if host == "example.com" {
				return []net.IPAddr{{IP: net.ParseIP("1.1.1.1")}}, nil
			}
			return net.DefaultResolver.LookupIPAddr(ctx, host)
		})
		defer cleanupDNS()

		// Stream the body in delayed chunks after the headers are sent, so
		// reading it can only succeed if the per-attempt context set up by
		// RoundTrip is still alive once RoundTrip has already returned.
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			flusher := w.(http.Flusher)
			for _, chunk := range []string{"first-", "second-", "third"} {
				_, _ = w.Write([]byte(chunk))
				flusher.Flush()
				time.Sleep(80 * time.Millisecond)
			}
		}))
		defer server.Close()

		originalDialer := dialer
		t.Cleanup(func() {
			dialer = originalDialer
		})
		dialer = &mockDialer{
			dialFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
				return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
			},
		}

		rt := &ssrfProxyRoundTripper{
			insecure:   true, // skip verification for self-signed test server cert
			underlying: http.DefaultTransport,
		}

		req, err := http.NewRequest(http.MethodGet, "https://example.com/foo", nil)
		require.NoError(t, err)
		req = req.WithContext(context.WithValue(req.Context(), useProxyKey, true))

		resp, err := rt.RoundTrip(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.Equal(t, "first-second-third", string(body))
	})
}
