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

func setEgressEnv(t *testing.T, policy, legacy, allowlist string) {
	t.Helper()
	t.Setenv(webhookEgressPolicyEnv, policy)
	t.Setenv(privateNetworkAccessEnv, legacy)
	t.Setenv(webhookEgressAllowlistEnv, allowlist)
}

func TestValidateNetworkTarget(t *testing.T) {
	setEgressEnv(t, "public_only", "", "")
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
			require.NoError(t, ValidateNetworkTarget(context.Background(), resolver, host))
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
		"168.63.129.16",
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
			require.Error(t, ValidateNetworkTarget(context.Background(), resolver, host))
		})
	}
}

// TestValidateNetworkTargetDefault covers the default policy: internal webhook receivers keep
// working, while targets that reach the host itself or cloud metadata stay blocked.
func TestValidateNetworkTargetDefault(t *testing.T) {
	setEgressEnv(t, "", "", "")
	resolver := &stubNetworkResolver{answers: map[string][]netip.Addr{
		"receiver.ns.svc.cluster.local": {netip.MustParseAddr("10.96.12.7")},
		"rebind.example.com":            {netip.MustParseAddr("169.254.169.254")},
	}}

	for _, host := range []string{
		"receiver.ns.svc.cluster.local",
		"8.8.8.8",
		"10.0.0.1",
		"100.64.0.1",
		"172.16.0.1",
		"192.168.0.1",
		"198.18.0.1",
		"fc00::1",
		"fd12:3456::1",
	} {
		t.Run("allows "+host, func(t *testing.T) {
			require.NoError(t, ValidateNetworkTarget(context.Background(), resolver, host))
		})
	}

	for _, host := range []string{
		"localhost",
		"metadata.google.internal",
		"rebind.example.com",
		"0.0.0.0",
		"127.0.0.1",
		"169.254.169.254",
		"100.100.100.200",
		"168.63.129.16",
		"224.0.0.1",
		"255.255.255.255",
		"::",
		"::1",
		"::ffff:169.254.169.254",
		"::169.254.169.254",
		"64:ff9b::a9fe:a9fe",
		"2002:a9fe:a9fe::1",
		"2001:0:a9fe:a9fe::1",
		"fd00:ec2::254",
		"fd20:ce::254",
		"fe80::1",
		"fe80::1%eth0",
		"ff02::1",
	} {
		t.Run("blocks "+host, func(t *testing.T) {
			require.ErrorContains(t, ValidateNetworkTarget(context.Background(), resolver, host), "restricted")
		})
	}
}

// TestEgressLevelPrecedence pins the compatibility contract: the policy wins, the deprecated boolean
// keeps its 2.15 meaning when the policy is unset, and invalid values fail closed.
func TestEgressLevelPrecedence(t *testing.T) {
	for _, tc := range []struct {
		policy, legacy string
		want           egressLevel
	}{
		{"", "", egressBlockRestricted},
		{" ", " ", egressBlockRestricted},
		{"", "true", egressAllowAll},
		{"", "1", egressAllowAll},
		{"", "false", egressPublicOnly},
		{"", "flase", egressPublicOnly},
		{"allow_all", "", egressAllowAll},
		{"Block_Restricted", "", egressBlockRestricted},
		{"public_only", "", egressPublicOnly},
		{"public_only", "true", egressPublicOnly},
		{"allow_all", "false", egressAllowAll},
		{"allow-all", "true", egressPublicOnly},
	} {
		t.Run(tc.policy+"/"+tc.legacy, func(t *testing.T) {
			assert.Equal(t, tc.want, parseEgressLevel(tc.policy, tc.legacy))
		})
	}
}

// TestValidateNetworkTargetPerLevel covers what each level lets through.
func TestValidateNetworkTargetPerLevel(t *testing.T) {
	for _, tc := range []struct {
		policy, legacy              string
		public, private, restricted bool
	}{
		{"allow_all", "", true, true, true},
		{"", "true", true, true, true},
		{"block_restricted", "", true, true, false},
		{"", "", true, true, false},
		{"public_only", "", true, false, false},
		{"", "false", true, false, false},
	} {
		t.Run(tc.policy+"/"+tc.legacy, func(t *testing.T) {
			setEgressEnv(t, tc.policy, tc.legacy, "")
			for target, allowed := range map[string]bool{
				"8.8.8.8":                  tc.public,
				"10.0.0.1":                 tc.private,
				"fd12::1":                  tc.private,
				"169.254.169.254":          tc.restricted,
				"127.0.0.1":                tc.restricted,
				"metadata.google.internal": tc.restricted,
			} {
				err := ValidateNetworkTarget(context.Background(), nil, target)
				if allowed {
					assert.NoError(t, err, target)
				} else {
					assert.Error(t, err, target)
				}
				if address, parseErr := netip.ParseAddr(target); parseErr == nil {
					err := checkEgress("tcp", net.JoinHostPort(address.String(), "443"), nil)
					assert.Equal(t, allowed, err == nil, "dial %s: %v", target, err)
				}
			}
		})
	}
}

// TestValidateNetworkTargetAllowlist covers HARBOR_PRIVATE_NETWORK_ALLOWLIST: listed destinations
// pass even with private access disabled, including restricted ones the operator names
// explicitly, and allowlisted hostnames are trusted without resolution.
func TestValidateNetworkTargetAllowlist(t *testing.T) {
	setEgressEnv(t, "public_only", "", "10.1.0.0/16, hooks.internal *.svc.cluster.local,127.0.0.1,::ffff:192.168.0.0/112,not_a_host!,*.")
	resolver := &stubNetworkResolver{answers: map[string][]netip.Addr{
		"listed-range.example.com": {netip.MustParseAddr("10.1.4.4")},
		"other-range.example.com":  {netip.MustParseAddr("10.2.4.4")},
	}}

	for _, host := range []string{
		"10.1.2.3",
		"listed-range.example.com",
		"hooks.internal",
		"HOOKS.internal.",
		"receiver.ns.svc.cluster.local",
		"127.0.0.1",
		"192.168.7.7",
		"8.8.8.8",
	} {
		t.Run("allows "+host, func(t *testing.T) {
			require.NoError(t, ValidateNetworkTarget(context.Background(), resolver, host))
		})
	}
	for _, host := range []string{
		"10.2.0.1",
		"other-range.example.com",
		"svc.cluster.local",
		"evilhooks.internal",
		"127.0.0.2",
		"169.254.169.254",
	} {
		t.Run("blocks "+host, func(t *testing.T) {
			require.Error(t, ValidateNetworkTarget(context.Background(), resolver, host))
		})
	}
	assert.NotContains(t, resolver.calls, "hooks.internal", "allowlisted hostnames are not resolved")
}

func TestCheckEgress(t *testing.T) {
	setEgressEnv(t, "", "", "")
	require.NoError(t, checkEgress("tcp4", "10.0.0.1:443", nil))
	require.NoError(t, checkEgress("tcp6", "[fd12::1]:443", nil))
	require.ErrorContains(t, checkEgress("tcp4", "169.254.169.254:80", nil), "restricted network address")
	require.ErrorContains(t, checkEgress("tcp6", "[::ffff:127.0.0.1]:80", nil), "restricted network address")

	setEgressEnv(t, "public_only", "", "")
	require.ErrorContains(t, checkEgress("tcp4", "10.0.0.1:443", nil), "private network address")
	require.NoError(t, checkEgress("tcp4", "8.8.8.8:443", nil))
}

func TestWithEgressGuard(t *testing.T) {
	setEgressEnv(t, "", "", "")
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = listener.Close() })

	dial := func(address string) error {
		transport := &http.Transport{Proxy: http.ProxyFromEnvironment}
		WithEgressGuard()(transport)
		conn, err := transport.DialContext(context.Background(), "tcp", address)
		if conn != nil {
			_ = conn.Close()
		}
		return err
	}
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)

	// Loopback is refused before the connection is made, even with private access allowed.
	require.ErrorContains(t, dial(listener.Addr().String()), "restricted network address")

	// The allowlist reopens it, by address or by the hostname the transport dials.
	t.Setenv(webhookEgressAllowlistEnv, "127.0.0.1")
	require.NoError(t, dial(listener.Addr().String()))
	t.Setenv(webhookEgressAllowlistEnv, "localhost")
	require.NoError(t, dial(net.JoinHostPort("localhost", port)))
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

// TestWithEgressGuardHonoursProxy covers the jobservice HTTP(S)_PROXY path: a public target
// must still be sent through the configured proxy, even though the proxy itself sits on a
// private address that the dial-time guard would otherwise refuse.
func TestWithEgressGuardHonoursProxy(t *testing.T) {
	setEgressEnv(t, "public_only", "", "")
	proxy := newRecordingProxy(t)

	transport := &http.Transport{Proxy: http.ProxyURL(proxy.url)}
	WithEgressGuard()(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	resp, err := client.Get("http://8.8.8.8/hook")
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, []string{"http://8.8.8.8/hook"}, proxy.seen())
}

// TestWithEgressGuardProxyBlocksPrivateTargets proves that routing through a proxy does not
// reopen the SSRF: blocked targets are refused before the proxy is ever contacted.
func TestWithEgressGuardProxyBlocksPrivateTargets(t *testing.T) {
	setEgressEnv(t, "public_only", "", "")
	proxy := newRecordingProxy(t)
	resolver := &stubNetworkResolver{answers: map[string][]netip.Addr{
		"public.example.com":  {netip.MustParseAddr("8.8.8.8")},
		"rebound.example.com": {netip.MustParseAddr("10.0.0.1")},
	}}

	transport := &http.Transport{Proxy: http.ProxyURL(proxy.url)}
	withEgressGuard(resolver)(transport)
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

// TestWithEgressGuardProxyAddressNotReachableDirectly covers a NO_PROXY entry that matches the
// proxy itself: a direct request to the proxy's address must not inherit the dial exemption the
// proxy gets once it has been used.
func TestWithEgressGuardProxyAddressNotReachableDirectly(t *testing.T) {
	setEgressEnv(t, "", "", "")
	proxy := newRecordingProxy(t)
	transport := &http.Transport{Proxy: func(req *http.Request) (*url.URL, error) {
		if req.URL.Host == proxy.url.Host {
			return nil, nil
		}
		return proxy.url, nil
	}}
	withEgressGuard(&stubNetworkResolver{})(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	resp, err := client.Get("http://8.8.8.8/hook")
	require.NoError(t, err)
	_ = resp.Body.Close()

	_, err = client.Get(proxy.url.String() + "/internal")
	require.ErrorContains(t, err, "restricted network address")
	assert.Equal(t, []string{"http://8.8.8.8/hook"}, proxy.seen())
}

// TestWithEgressGuardNoProxyStillGuardsDial covers NO_PROXY: when the proxy function declines a
// target, the direct dial keeps the resolved-address check.
func TestWithEgressGuardNoProxyStillGuardsDial(t *testing.T) {
	setEgressEnv(t, "", "", "")
	target := newRecordingProxy(t)
	transport := &http.Transport{Proxy: func(*http.Request) (*url.URL, error) { return nil, nil }}
	withEgressGuard(&stubNetworkResolver{})(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	_, err := client.Get(target.url.String())
	require.ErrorContains(t, err, "restricted network address")
	assert.Empty(t, target.seen())
}

// TestWithEgressGuardProxyPrivateAllowed keeps private targets reachable on the proxied path under
// block_restricted, while restricted ones are still refused before reaching the proxy.
func TestWithEgressGuardProxyPrivateAllowed(t *testing.T) {
	setEgressEnv(t, "block_restricted", "", "")
	proxy := newRecordingProxy(t)
	transport := &http.Transport{Proxy: http.ProxyURL(proxy.url)}
	withEgressGuard(&stubNetworkResolver{})(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	resp, err := client.Get("http://10.0.0.1/hook")
	require.NoError(t, err)
	_ = resp.Body.Close()
	_, err = client.Get("http://169.254.169.254/latest/meta-data/")
	require.ErrorContains(t, err, "restricted")
	assert.Equal(t, []string{"http://10.0.0.1/hook"}, proxy.seen())
}

func TestParseEgressPolicy(t *testing.T) {
	policy := parseEgressPolicy("", "", "10.0.0.0/8 ::ffff:172.16.0.0/108, 192.168.1.10, fe80::1%eth0, *.svc.cluster.local, Hooks.Example.COM., bad host, -, *.")
	assert.Equal(t, egressBlockRestricted, policy.level)
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("172.16.0.0/12"),
		netip.MustParsePrefix("192.168.1.10/32"),
	}, policy.allowedPrefixes)
	assert.Equal(t, []string{".svc.cluster.local", "hooks.example.com", "bad", "host"}, policy.allowedHosts)
}

// TestParseEgressPolicyRejectsMalformedWildcards covers entries that would otherwise land in the
// hostname list as an IP address and skip classification.
func TestParseEgressPolicyRejectsMalformedWildcards(t *testing.T) {
	policy := parseEgressPolicy("public_only", "", "*169.254.169.254, *.169.254.169.254, .example.com, **.example.com, 169.254.169.254.5, *.Example.com")
	assert.Empty(t, policy.allowedPrefixes)
	assert.Equal(t, []string{".example.com"}, policy.allowedHosts)

	setEgressEnv(t, "block_restricted", "", "*169.254.169.254")
	require.Error(t, ValidateNetworkTarget(context.Background(), nil, "169.254.169.254"))
	require.Error(t, checkEgress("tcp4", "169.254.169.254:80", nil))
}

// TestWithEgressGuardAllowlistedHostThroughProxy covers a hostname the operator allowlists: it is
// handed to the proxy without Harbor resolving it, since Harbor may not be able to.
func TestWithEgressGuardAllowlistedHostThroughProxy(t *testing.T) {
	setEgressEnv(t, "public_only", "", "*.corp.internal")
	proxy := newRecordingProxy(t)
	transport := &http.Transport{Proxy: http.ProxyURL(proxy.url)}
	withEgressGuard(&stubNetworkResolver{})(transport)
	client := &http.Client{Transport: transport, Timeout: 2 * time.Second}

	resp, err := client.Get("http://hooks.corp.internal/hook")
	require.NoError(t, err)
	_ = resp.Body.Close()
	_, err = client.Get("http://hooks.elsewhere.internal/hook")
	require.ErrorContains(t, err, "cannot be resolved")
	assert.Equal(t, []string{"http://hooks.corp.internal/hook"}, proxy.seen())
}
