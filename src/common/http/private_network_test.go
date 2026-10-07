// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package http

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubNetworkResolver struct {
	answers map[string][]netip.Addr
	errors  map[string]error
	calls   []string
}

func (resolver *stubNetworkResolver) LookupNetIP(_ context.Context, _ string, host string) ([]netip.Addr, error) {
	resolver.calls = append(resolver.calls, host)
	if err := resolver.errors[host]; err != nil {
		return nil, err
	}
	return resolver.answers[host], nil
}

func TestValidatePublicNetworkTarget(t *testing.T) {
	t.Setenv(privateNetworkAccessEnv, "false")
	resolver := &stubNetworkResolver{
		answers: map[string][]netip.Addr{
			"example.com": {netip.MustParseAddr("8.8.8.8")},
			"mixed.example.com": {
				netip.MustParseAddr("8.8.8.8"),
				netip.MustParseAddr("127.0.0.1"),
			},
		},
		errors: map[string]error{
			"failure.example.com": errors.New("resolver failed"),
		},
	}

	for _, host := range []string{
		"example.com",
		"example.com.",
		"8.8.8.8",
		"192.0.0.9",
		"192.0.0.10",
		"2001:1::1",
		"2606:4700:4700::1111",
	} {
		t.Run("allows "+host, func(t *testing.T) {
			require.NoError(t, ValidatePublicNetworkTarget(context.Background(), resolver, host))
		})
	}

	for _, host := range []string{
		"",
		"localhost",
		"api.localhost.",
		"metadata.google.internal",
		"sub.metadata.azure.com.",
		"mixed.example.com",
		"missing.example.com",
		"failure.example.com",
		"0.0.0.0",
		"10.0.0.1",
		"100.64.0.1",
		"127.0.0.1",
		"169.254.169.254",
		"172.16.0.1",
		"192.168.0.1",
		"198.18.0.1",
		"224.0.0.1",
		"240.0.0.1",
		"::1",
		"::ffff:127.0.0.1",
		"fc00::1",
		"fe80::1",
		"fe80::1%eth0",
	} {
		t.Run("blocks "+host, func(t *testing.T) {
			require.Error(t, ValidatePublicNetworkTarget(context.Background(), resolver, host))
		})
	}
}

// TestValidatePublicNetworkTargetEscapeHatch proves the HARBOR_ALLOW_PRIVATE_NETWORK_ACCESS
// env allows private targets by default when unset/empty, blocks when false or malformed,
// and permits when explicitly true.
func TestValidatePublicNetworkTargetEscapeHatch(t *testing.T) {
	// Unset / empty: allowed by default without regression
	t.Setenv(privateNetworkAccessEnv, "")
	require.NoError(t, ValidatePublicNetworkTarget(context.Background(), nil, "127.0.0.1"))

	// Explicit false: blocked
	t.Setenv(privateNetworkAccessEnv, "false")
	require.Error(t, ValidatePublicNetworkTarget(context.Background(), nil, "127.0.0.1"))

	// Malformed / invalid: fails closed (blocked) to prevent accidental SSRF bypass
	t.Setenv(privateNetworkAccessEnv, "flase")
	require.Error(t, ValidatePublicNetworkTarget(context.Background(), nil, "127.0.0.1"))
	t.Setenv(privateNetworkAccessEnv, "invalid")
	require.Error(t, ValidatePublicNetworkTarget(context.Background(), nil, "127.0.0.1"))

	// Explicit true: allowed
	t.Setenv(privateNetworkAccessEnv, "true")
	require.NoError(t, ValidatePublicNetworkTarget(context.Background(), nil, "127.0.0.1"))
}

func TestWithPublicNetworkOnly(t *testing.T) {
	t.Setenv(privateNetworkAccessEnv, "false")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	dial := func() error {
		transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
		WithPublicNetworkOnly()(transport)
		conn, err := transport.DialContext(context.Background(), "tcp", listener.Addr().String())
		if conn != nil {
			_ = conn.Close()
		}
		return err
	}

	// Default: the loopback dial is refused before the connection is made.
	require.ErrorContains(t, dial(), "private network address")

	// Escape hatch: the same dial succeeds when private access is allowed.
	t.Setenv(privateNetworkAccessEnv, "true")
	require.NoError(t, dial())
}

type recordingProxy struct {
	mu       sync.Mutex
	requests []string
	url      *url.URL
}

func newRecordingProxy(t *testing.T) *recordingProxy {
	t.Helper()
	proxy := &recordingProxy{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy.mu.Lock()
		proxy.requests = append(proxy.requests, r.URL.String())
		proxy.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	proxyURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	proxy.url = proxyURL
	return proxy
}

func (proxy *recordingProxy) seen() []string {
	proxy.mu.Lock()
	defer proxy.mu.Unlock()
	return append([]string(nil), proxy.requests...)
}

// TestWithPublicNetworkOnlyHonoursProxy covers the jobservice HTTP(S)_PROXY path: a public target
// must still be sent through the configured proxy, even though the proxy itself sits on a
// private address that the dial-time guard would otherwise refuse.
func TestWithPublicNetworkOnlyHonoursProxy(t *testing.T) {
	t.Setenv(privateNetworkAccessEnv, "false")
	proxy := newRecordingProxy(t)

	transport := &http.Transport{Proxy: http.ProxyURL(proxy.url)}
	WithPublicNetworkOnly()(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	resp, err := client.Get("http://8.8.8.8/hook")
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{"http://8.8.8.8/hook"}, proxy.seen())
}

// TestWithPublicNetworkOnlyProxyBlocksPrivateTargets proves that routing through a proxy does not
// reopen the SSRF: non-public targets are refused before the proxy is ever contacted.
func TestWithPublicNetworkOnlyProxyBlocksPrivateTargets(t *testing.T) {
	t.Setenv(privateNetworkAccessEnv, "false")
	proxy := newRecordingProxy(t)
	resolver := &stubNetworkResolver{answers: map[string][]netip.Addr{
		"public.example.com":  {netip.MustParseAddr("8.8.8.8")},
		"rebound.example.com": {netip.MustParseAddr("10.0.0.1")},
	}}

	transport := &http.Transport{Proxy: http.ProxyURL(proxy.url)}
	withPublicNetworkOnly(resolver)(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	for _, target := range []string{
		"http://169.254.169.254/latest/meta-data/",
		"http://10.0.0.1/",
		"https://[fd00::1]/",
		"http://metadata.google.internal/",
		"http://rebound.example.com/",
		"http://unresolvable.example.com/",
	} {
		t.Run(target, func(t *testing.T) {
			_, err := client.Get(target)
			require.Error(t, err)
		})
	}
	assert.Empty(t, proxy.seen(), "no private target may reach the proxy")

	resp, err := client.Get("http://public.example.com/hook")
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, []string{"http://public.example.com/hook"}, proxy.seen())

}

// TestWithPublicNetworkOnlyProxyAddressNotReachableDirectly covers a NO_PROXY entry that matches
// the proxy itself: a direct request to the proxy's address must not inherit the dial exemption
// the proxy gets once it has been used.
func TestWithPublicNetworkOnlyProxyAddressNotReachableDirectly(t *testing.T) {
	t.Setenv(privateNetworkAccessEnv, "false")
	proxy := newRecordingProxy(t)
	transport := &http.Transport{Proxy: func(req *http.Request) (*url.URL, error) {
		if req.URL.Host == proxy.url.Host {
			return nil, nil
		}
		return proxy.url, nil
	}}
	withPublicNetworkOnly(&stubNetworkResolver{})(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	resp, err := client.Get("http://8.8.8.8/hook")
	require.NoError(t, err)
	_ = resp.Body.Close()

	_, err = client.Get(proxy.url.String() + "/internal")
	require.ErrorContains(t, err, "private network address")
	assert.Equal(t, []string{"http://8.8.8.8/hook"}, proxy.seen())
}

// TestWithPublicNetworkOnlyNoProxyStillGuardsDial covers NO_PROXY: when the proxy function
// declines a target, the direct dial keeps the resolved-address check.
func TestWithPublicNetworkOnlyNoProxyStillGuardsDial(t *testing.T) {
	t.Setenv(privateNetworkAccessEnv, "false")
	target := newRecordingProxy(t)
	transport := &http.Transport{Proxy: func(*http.Request) (*url.URL, error) { return nil, nil }}
	withPublicNetworkOnly(&stubNetworkResolver{})(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	_, err := client.Get(target.url.String())
	require.ErrorContains(t, err, "private network address")
	assert.Empty(t, target.seen())
}

// TestWithPublicNetworkOnlyProxyEscapeHatch keeps the opt-out working on the proxied path.
func TestWithPublicNetworkOnlyProxyEscapeHatch(t *testing.T) {
	t.Setenv(privateNetworkAccessEnv, "true")
	proxy := newRecordingProxy(t)
	transport := &http.Transport{Proxy: http.ProxyURL(proxy.url)}
	withPublicNetworkOnly(&stubNetworkResolver{})(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	resp, err := client.Get("http://10.0.0.1/hook")
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, []string{"http://10.0.0.1/hook"}, proxy.seen())
}
