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
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// privateNetworkAccessEnv allows toggling private network access for outbound webhook and Slack notifications.
// Defaults to true (allowed) to avoid regressions for existing on-premise and Kubernetes deployments.
const privateNetworkAccessEnv = "HARBOR_ALLOW_PRIVATE_NETWORK_ACCESS"

// Exceptions and denials follow the IANA IPv4/IPv6 special-purpose address registries.
// A destination is "public" only when it is global-unicast, not private, and not in any
// non-public range below. IsGlobalUnicast() alone is insufficient because it admits
// special-purpose ranges such as 198.18.0.0/15 and fec0::/10.
var (
	publicIPv4Exceptions = []netip.Prefix{
		netip.MustParsePrefix("192.0.0.9/32"),
		netip.MustParsePrefix("192.0.0.10/32"),
	}
	publicIPv6Exceptions = []netip.Prefix{
		netip.MustParsePrefix("2001:1::1/128"),
		netip.MustParsePrefix("2001:1::2/128"),
		netip.MustParsePrefix("2001:1::3/128"),
		netip.MustParsePrefix("2001:3::/32"),
		netip.MustParsePrefix("2001:4:112::/48"),
		netip.MustParsePrefix("2001:20::/28"),
		netip.MustParsePrefix("2001:30::/28"),
	}
	publicIPv6Prefix  = netip.MustParsePrefix("2000::/3")
	nonPublicPrefixes = []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),
		netip.MustParsePrefix("100.64.0.0/10"),
		netip.MustParsePrefix("192.0.0.0/24"),
		netip.MustParsePrefix("192.0.2.0/24"),
		netip.MustParsePrefix("192.88.99.0/24"),
		netip.MustParsePrefix("198.18.0.0/15"),
		netip.MustParsePrefix("198.51.100.0/24"),
		netip.MustParsePrefix("203.0.113.0/24"),
		netip.MustParsePrefix("224.0.0.0/4"),
		netip.MustParsePrefix("240.0.0.0/4"),
		netip.MustParsePrefix("::/96"),
		netip.MustParsePrefix("64:ff9b::/96"),
		netip.MustParsePrefix("64:ff9b:1::/48"),
		netip.MustParsePrefix("100::/64"),
		netip.MustParsePrefix("100:0:0:1::/64"),
		netip.MustParsePrefix("2001::/23"),
		netip.MustParsePrefix("2001:db8::/32"),
		netip.MustParsePrefix("2002::/16"),
		netip.MustParsePrefix("3fff::/20"),
		netip.MustParsePrefix("5f00::/16"),
		netip.MustParsePrefix("fc00::/7"),
		netip.MustParsePrefix("fe80::/10"),
		netip.MustParsePrefix("fec0::/10"),
		netip.MustParsePrefix("ff00::/8"),
	}
	metadataTargetHosts = map[string]struct{}{
		"metadata.azure.com":       {},
		"metadata.google.internal": {},
	}
)

func newGuardedDialer() *net.Dialer {
	return &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
}

// privateNetworkAccessAllowed reports whether private egress is permitted.
// Defaults to true when unset or empty to prevent regression on existing internal webhook targets.
func privateNetworkAccessAllowed() bool {
	val := strings.TrimSpace(os.Getenv(privateNetworkAccessEnv))
	if val == "" {
		return true
	}
	allowed, err := strconv.ParseBool(val)
	return err == nil && allowed
}

// blockPrivateNetwork is a net.Dialer.Control callback. It runs after DNS resolution and on
// every dial (including redirect hops), so it defeats DNS rebinding. It permits private-network
// access by default, and blocks non-public destinations when HARBOR_ALLOW_PRIVATE_NETWORK_ACCESS
// is explicitly set to false (or a malformed value).
func blockPrivateNetwork(_ string, address string, _ syscall.RawConn) error {
	if privateNetworkAccessAllowed() {
		return nil
	}
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("failed to parse dial address %q: %w", address, err)
	}
	addressIP, err := netip.ParseAddr(host)
	if err != nil {
		return fmt.Errorf("failed to parse dial IP %q: %w", host, err)
	}
	addressIP = addressIP.Unmap()
	if !isPublicNetworkAddress(addressIP) {
		return fmt.Errorf("connections to private network address %s are blocked", addressIP)
	}
	return nil
}

func isPublicNetworkAddress(address netip.Addr) bool {
	if !address.IsValid() || address.Zone() != "" {
		return false
	}

	address = address.Unmap()
	if address.Is4() && containsAddress(publicIPv4Exceptions, address) {
		return true
	}
	if address.Is6() {
		if containsAddress(publicIPv6Exceptions, address) {
			return true
		}
		if !publicIPv6Prefix.Contains(address) {
			return false
		}
	}
	if !address.IsGlobalUnicast() || address.IsPrivate() {
		return false
	}

	return !containsAddress(nonPublicPrefixes, address)
}

func containsAddress(prefixes []netip.Prefix, address netip.Addr) bool {
	for _, prefix := range prefixes {
		if prefix.Contains(address) {
			return true
		}
	}
	return false
}

// NetworkResolver resolves target hostnames for public-network validation.
type NetworkResolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

// ValidatePublicNetworkTarget rejects a target host that does not resolve entirely to public
// addresses. Literal IPs are classified directly; hostnames are resolved and every returned
// address must be public. It short-circuits when private access is allowed so the check matches
// the dial-time guard's escape hatch.
func ValidatePublicNetworkTarget(ctx context.Context, resolver NetworkResolver, host string) error {
	if privateNetworkAccessAllowed() {
		return nil
	}
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "" {
		return fmt.Errorf("target must include a hostname")
	}
	if isBlockedTargetHostname(host) {
		return fmt.Errorf("target hostname %q is not public", host)
	}

	address, err := netip.ParseAddr(host)
	if err == nil {
		if !isPublicNetworkAddress(address) {
			return fmt.Errorf("target address %q is not public", host)
		}
		return nil
	}

	if resolver == nil {
		return fmt.Errorf("target hostname %q cannot be resolved", host)
	}
	addresses, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil || len(addresses) == 0 {
		return fmt.Errorf("target hostname %q cannot be resolved", host)
	}
	for _, address := range addresses {
		if !isPublicNetworkAddress(address) {
			return fmt.Errorf("target hostname %q resolves to a non-public address", host)
		}
	}
	return nil
}

func isBlockedTargetHostname(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	for metadataHost := range metadataTargetHosts {
		if host == metadataHost || strings.HasSuffix(host, "."+metadataHost) {
			return true
		}
	}
	return false
}

// WithPublicNetworkOnly blocks connections to non-public destinations. Intended for the
// notification (webhook/slack) HTTP clients, whose targets are attacker-controlled. Direct
// connections are checked at dial time on the resolved address. When the transport's proxy
// (HTTP_PROXY/HTTPS_PROXY/NO_PROXY) applies, the dial goes to the operator's proxy instead, so
// the target host is validated before the request is handed to it. Honors the
// HARBOR_ALLOW_PRIVATE_NETWORK_ACCESS escape hatch.
func WithPublicNetworkOnly() func(*http.Transport) {
	return withPublicNetworkOnly(net.DefaultResolver)
}

func withPublicNetworkOnly(resolver NetworkResolver) func(*http.Transport) {
	return func(transport *http.Transport) {
		guard := &publicNetworkGuard{resolver: resolver, upstreamProxy: transport.Proxy}
		dialer := newGuardedDialer()
		dialer.Control = blockPrivateNetwork
		guard.guardedDial = dialer.DialContext
		guard.proxyDial = newGuardedDialer().DialContext
		if guard.upstreamProxy != nil {
			transport.Proxy = guard.proxy
		}
		transport.DialContext = guard.dialContext
	}
}

type publicNetworkGuard struct {
	resolver      NetworkResolver
	upstreamProxy func(*http.Request) (*url.URL, error)
	guardedDial   func(context.Context, string, string) (net.Conn, error)
	proxyDial     func(context.Context, string, string) (net.Conn, error)
	// proxyAddrs holds the host:port of every proxy the upstream function returned. Only dials to
	// exactly these addresses skip the private-address check: the proxy is operator configuration
	// and is commonly on a private network.
	proxyAddrs sync.Map
}

func (guard *publicNetworkGuard) proxy(req *http.Request) (*url.URL, error) {
	proxyURL, err := guard.upstreamProxy(req)
	if err != nil || privateNetworkAccessAllowed() {
		return proxyURL, err
	}
	if proxyURL == nil {
		// A direct request whose target is the proxy's own address would otherwise inherit the
		// proxy's dial exemption.
		if _, isProxy := guard.proxyAddrs.Load(canonicalAddr(req.URL)); isProxy {
			return nil, fmt.Errorf("connections to private network address %s are blocked", req.URL.Host)
		}
		return nil, nil
	}
	// The proxy, not this process, resolves and connects to the target, so the dial-time check
	// cannot see it. Validate the target before handing it over.
	if err := ValidatePublicNetworkTarget(req.Context(), guard.resolver, req.URL.Hostname()); err != nil {
		return nil, err
	}
	guard.proxyAddrs.Store(canonicalAddr(proxyURL), struct{}{})
	return proxyURL, nil
}

func (guard *publicNetworkGuard) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if _, isProxy := guard.proxyAddrs.Load(address); isProxy {
		return guard.proxyDial(ctx, network, address)
	}
	return guard.guardedDial(ctx, network, address)
}

var defaultPorts = map[string]string{"http": "80", "https": "443", "socks5": "1080", "socks5h": "1080"}

// canonicalAddr mirrors net/http's host:port form for a URL, which is the address the transport
// passes to DialContext.
func canonicalAddr(target *url.URL) string {
	port := target.Port()
	if port == "" {
		port = defaultPorts[strings.ToLower(target.Scheme)]
	}
	return net.JoinHostPort(target.Hostname(), port)
}
