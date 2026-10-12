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
	"unicode"

	"github.com/goharbor/harbor/src/lib/log"
)

// webhookEgressPolicyEnv selects which destinations the webhook and Slack notification clients
// may reach: allow_all, block_restricted (the default) or public_only. See egressLevel.
const webhookEgressPolicyEnv = "HARBOR_WEBHOOK_EGRESS_POLICY"

// webhookEgressAllowlistEnv lists destinations the notification clients may always reach, whatever
// the policy: IPs, CIDRs, hostnames, and "*.domain" suffixes, separated by commas or whitespace. An
// allowlisted hostname is not resolved by Harbor for the check, so it is trusted wherever DNS
// points it.
const webhookEgressAllowlistEnv = "HARBOR_WEBHOOK_EGRESS_ALLOWLIST"

// privateNetworkAccessEnv is the deprecated boolean from 2.15. When the policy is not set it keeps
// its meaning: true allows everything, false allows public destinations only.
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
	// restrictedPrefixes are blocked unless the policy is allow_all or they are allowlisted: they reach
	// the Harbor host itself, cloud instance metadata and control endpoints, or translate into an
	// IPv4 address that could. Loopback, link-local and multicast are matched by netip helpers.
	restrictedPrefixes = []netip.Prefix{
		netip.MustParsePrefix("0.0.0.0/8"),
		netip.MustParsePrefix("100.100.100.200/32"), // Alibaba Cloud metadata, inside CGNAT space
		netip.MustParsePrefix("168.63.129.16/32"),   // Azure WireServer, inside public space
		netip.MustParsePrefix("240.0.0.0/4"),
		netip.MustParsePrefix("::/96"),
		netip.MustParsePrefix("64:ff9b::/96"),
		netip.MustParsePrefix("64:ff9b:1::/48"),
		netip.MustParsePrefix("2001::/32"),
		netip.MustParsePrefix("2002::/16"),
		netip.MustParsePrefix("fd00:ec2::254/128"), // AWS IMDS over IPv6, inside ULA space
		netip.MustParsePrefix("fd20:ce::254/128"),  // GCP metadata over IPv6, inside ULA space
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

// egressLevel is how far the notification clients may reach beyond public addresses.
type egressLevel int

const (
	// egressAllowAll performs no checks, as before 2.15.3.
	egressAllowAll egressLevel = iota
	// egressBlockRestricted allows private networks but not loopback, link-local, cloud metadata
	// and the other restricted destinations, which no legitimate webhook receiver needs.
	egressBlockRestricted
	// egressPublicOnly allows public addresses only.
	egressPublicOnly
)

var egressLevelNames = map[egressLevel]string{
	egressAllowAll:        "allow_all",
	egressBlockRestricted: "block_restricted",
	egressPublicOnly:      "public_only",
}

func (level egressLevel) String() string {
	return egressLevelNames[level]
}

// egressPolicy decides which destinations the notification clients may reach. The allowlist
// overrides the level.
type egressPolicy struct {
	level           egressLevel
	allowedPrefixes []netip.Prefix
	// allowedHosts holds exact hostnames; entries starting with "." match any subdomain.
	allowedHosts []string
}

var egressPolicyCache struct {
	sync.Mutex
	loaded                    bool
	policyEnv, legacyEnv, raw string
	policy                    egressPolicy
}

// currentEgressPolicy reads the policy from the environment, re-parsing only when it changed so
// that invalid and deprecated values are logged once rather than on every dial.
func currentEgressPolicy() egressPolicy {
	policyEnv, legacyEnv, allowlist := os.Getenv(webhookEgressPolicyEnv), os.Getenv(privateNetworkAccessEnv), os.Getenv(webhookEgressAllowlistEnv)
	egressPolicyCache.Lock()
	defer egressPolicyCache.Unlock()
	cache := &egressPolicyCache
	if !cache.loaded || cache.policyEnv != policyEnv || cache.legacyEnv != legacyEnv || cache.raw != allowlist {
		cache.policy = parseEgressPolicy(policyEnv, legacyEnv, allowlist)
		cache.policyEnv, cache.legacyEnv, cache.raw = policyEnv, legacyEnv, allowlist
		cache.loaded = true
	}
	return cache.policy
}

func parseEgressPolicy(policyEnv, legacyEnv, allowlist string) egressPolicy {
	policy := egressPolicy{level: parseEgressLevel(policyEnv, legacyEnv)}
	for _, entry := range strings.FieldsFunc(allowlist, func(r rune) bool { return r == ',' || unicode.IsSpace(r) }) {
		entry = strings.TrimSuffix(strings.ToLower(entry), ".")
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			policy.allowedPrefixes = append(policy.allowedPrefixes, netip.PrefixFrom(prefix.Addr().Unmap(), unmappedBits(prefix)).Masked())
			continue
		}
		if address, err := netip.ParseAddr(entry); err == nil && address.Zone() == "" {
			address = address.Unmap()
			policy.allowedPrefixes = append(policy.allowedPrefixes, netip.PrefixFrom(address, address.BitLen()))
			continue
		}
		// only "*.domain" is a wildcard; it is stored as ".domain"
		host, isWildcard := strings.CutPrefix(entry, "*.")
		if !isValidAllowlistHost(host) {
			log.Warningf("ignoring invalid %s entry %q", webhookEgressAllowlistEnv, entry)
			continue
		}
		if isWildcard {
			host = "." + host
		}
		policy.allowedHosts = append(policy.allowedHosts, host)
	}
	return policy
}

// parseEgressLevel applies the policy, then the deprecated boolean, then the default. An invalid
// value falls back to public_only so that a typo can never widen egress.
func parseEgressLevel(policyEnv, legacyEnv string) egressLevel {
	if value := strings.ToLower(strings.TrimSpace(policyEnv)); value != "" {
		for level, name := range egressLevelNames {
			if value == name {
				return level
			}
		}
		log.Warningf("invalid %s=%q, allowing public webhook targets only", webhookEgressPolicyEnv, policyEnv)
		return egressPublicOnly
	}
	value := strings.TrimSpace(legacyEnv)
	if value == "" {
		return egressBlockRestricted
	}
	allowed, err := strconv.ParseBool(value)
	switch {
	case err != nil:
		log.Warningf("invalid %s=%q, allowing public webhook targets only", privateNetworkAccessEnv, legacyEnv)
		return egressPublicOnly
	case allowed:
		log.Warningf("%s is deprecated, use %s=%s", privateNetworkAccessEnv, webhookEgressPolicyEnv, egressAllowAll)
		return egressAllowAll
	default:
		log.Warningf("%s is deprecated, use %s=%s", privateNetworkAccessEnv, webhookEgressPolicyEnv, egressPublicOnly)
		return egressPublicOnly
	}
}

// unmappedBits keeps an IPv4-mapped prefix such as ::ffff:10.0.0.0/104 equivalent to 10.0.0.0/8.
func unmappedBits(prefix netip.Prefix) int {
	if prefix.Addr().Is4In6() {
		return max(prefix.Bits()-96, 0)
	}
	return prefix.Bits()
}

// isValidAllowlistHost accepts DNS names only. The last label must not be numeric, so an IP
// address can never enter the list as a hostname and skip address classification.
func isValidAllowlistHost(host string) bool {
	labels := strings.Split(host, ".")
	if host == "" || strings.Trim(labels[len(labels)-1], "0123456789") == "" {
		return false
	}
	for _, label := range labels {
		if label == "" || len(label) > 63 || strings.HasPrefix(label, "-") || strings.HasSuffix(label, "-") {
			return false
		}
		for _, r := range label {
			if (r < 'a' || r > 'z') && (r < '0' || r > '9') && r != '-' && r != '_' {
				return false
			}
		}
	}
	return true
}

// allowsHost reports whether host is on the allowlist. host must be lower-case without a
// trailing dot.
func (policy egressPolicy) allowsHost(host string) bool {
	for _, allowed := range policy.allowedHosts {
		if host == allowed || strings.HasPrefix(allowed, ".") && strings.HasSuffix(host, allowed) {
			return true
		}
	}
	return false
}

// checkAddress returns an error when the policy blocks address.
func (policy egressPolicy) checkAddress(address netip.Addr) error {
	address = address.Unmap()
	switch {
	case policy.level == egressAllowAll, containsAddress(policy.allowedPrefixes, address.WithZone("")):
		return nil
	case isRestrictedNetworkAddress(address):
		return fmt.Errorf("connections to restricted network address %s are blocked", address)
	case isPublicNetworkAddress(address), policy.level == egressBlockRestricted:
		return nil
	default:
		return fmt.Errorf("connections to private network address %s are blocked", address)
	}
}

// logBlocked tells the operator which setting reopens a blocked target. The error returned to the
// caller stays free of configuration names because project admins see it.
func (policy egressPolicy) logBlocked(target string, err error) {
	log.Warningf("webhook egress policy %s blocked %s: %v; allow it with %s or %s",
		policy.level, target, err, webhookEgressAllowlistEnv, webhookEgressPolicyEnv)
}

func isRestrictedNetworkAddress(address netip.Addr) bool {
	return !address.IsValid() || address.Zone() != "" || address.IsLoopback() || address.IsUnspecified() ||
		address.IsLinkLocalUnicast() || address.IsMulticast() || containsAddress(restrictedPrefixes, address)
}

// checkEgress is a net.Dialer.Control callback. It runs after DNS resolution and on every dial,
// so it defeats DNS rebinding.
func checkEgress(_ string, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("failed to parse dial address %q: %w", address, err)
	}
	addressIP, err := netip.ParseAddr(host)
	if err != nil {
		return fmt.Errorf("failed to parse dial IP %q: %w", host, err)
	}
	policy := currentEgressPolicy()
	if err := policy.checkAddress(addressIP); err != nil {
		policy.logBlocked(address, err)
		return err
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

// NetworkResolver resolves target hostnames for egress validation.
type NetworkResolver interface {
	LookupNetIP(context.Context, string, string) ([]netip.Addr, error)
}

// ValidateNetworkTarget rejects a target host the notification clients would refuse to reach.
// Literal IPs are classified directly; hostnames are resolved and every returned address must
// be permitted. Allowlisted hostnames pass without resolution.
func ValidateNetworkTarget(ctx context.Context, resolver NetworkResolver, host string) error {
	policy := currentEgressPolicy()
	if policy.level == egressAllowAll {
		return nil
	}
	if err := validateNetworkTarget(ctx, resolver, policy, host); err != nil {
		policy.logBlocked(host, err)
		return err
	}
	return nil
}

func validateNetworkTarget(ctx context.Context, resolver NetworkResolver, policy egressPolicy, host string) error {
	host = strings.TrimSuffix(strings.ToLower(host), ".")
	if host == "" {
		return fmt.Errorf("target must include a hostname")
	}
	if policy.allowsHost(host) {
		return nil
	}
	if isBlockedTargetHostname(host) {
		return fmt.Errorf("target hostname %q is restricted", host)
	}

	address, err := netip.ParseAddr(host)
	if err == nil {
		if err := policy.checkAddress(address); err != nil {
			return fmt.Errorf("target address %q is blocked: %w", host, err)
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
		if err := policy.checkAddress(address); err != nil {
			return fmt.Errorf("target hostname %q resolves to a blocked address: %w", host, err)
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

// WithEgressGuard enforces the egress policy on a transport. Intended for the notification
// (webhook/slack) HTTP clients, whose targets are user-controlled. Direct connections are checked
// at dial time on the resolved address. When the transport's proxy (HTTP_PROXY/HTTPS_PROXY/NO_PROXY)
// applies, the dial goes to the operator's proxy instead, so the target host is validated before
// the request is handed to it.
func WithEgressGuard() func(*http.Transport) {
	return withEgressGuard(net.DefaultResolver)
}

func withEgressGuard(resolver NetworkResolver) func(*http.Transport) {
	return func(transport *http.Transport) {
		guard := &egressGuard{resolver: resolver, upstreamProxy: transport.Proxy}
		dialer := newGuardedDialer()
		dialer.Control = checkEgress
		guard.guardedDial = dialer.DialContext
		guard.unguardedDial = newGuardedDialer().DialContext
		if guard.upstreamProxy != nil {
			transport.Proxy = guard.proxy
		}
		transport.DialContext = guard.dialContext
	}
}

type egressGuard struct {
	resolver      NetworkResolver
	upstreamProxy func(*http.Request) (*url.URL, error)
	guardedDial   func(context.Context, string, string) (net.Conn, error)
	unguardedDial func(context.Context, string, string) (net.Conn, error)
	// proxyAddrs holds the host:port of every proxy the upstream function returned. Dials to
	// exactly these addresses skip the address check: the proxy is operator configuration and
	// may sit on a restricted address such as a loopback sidecar.
	proxyAddrs sync.Map
}

func (guard *egressGuard) proxy(req *http.Request) (*url.URL, error) {
	proxyURL, err := guard.upstreamProxy(req)
	if err != nil || currentEgressPolicy().level == egressAllowAll {
		return proxyURL, err
	}
	if proxyURL == nil {
		// A direct request whose target is the proxy's own address would otherwise inherit the
		// proxy's dial exemption.
		if _, isProxy := guard.proxyAddrs.Load(canonicalAddr(req.URL)); isProxy {
			if err := ValidateNetworkTarget(req.Context(), guard.resolver, req.URL.Hostname()); err != nil {
				return nil, err
			}
		}
		return nil, nil
	}
	// The proxy, not this process, resolves and connects to the target, so the dial-time check
	// cannot see it. Validate the target before handing it over.
	if err := ValidateNetworkTarget(req.Context(), guard.resolver, req.URL.Hostname()); err != nil {
		return nil, err
	}
	guard.proxyAddrs.Store(canonicalAddr(proxyURL), struct{}{})
	return proxyURL, nil
}

func (guard *egressGuard) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if _, isProxy := guard.proxyAddrs.Load(address); isProxy {
		return guard.unguardedDial(ctx, network, address)
	}
	if host, _, err := net.SplitHostPort(address); err == nil &&
		currentEgressPolicy().allowsHost(strings.TrimSuffix(strings.ToLower(host), ".")) {
		return guard.unguardedDial(ctx, network, address)
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
