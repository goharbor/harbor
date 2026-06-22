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

package lib

import (
	"context"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
)

type dialer interface {
	DialContext(ctx context.Context, network, address string) (net.Conn, error)
}

var (
	lookupIPAddr        = net.DefaultResolver.LookupIPAddr
	publicDialer dialer = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}
)

// ValidatePublicHTTPURL checks whether rawURL is an HTTP(S) URL whose target is
// suitable for server-side outbound requests. It removes user info, query, and
// fragment components and rejects loopback, private, link-local, and other
// non-public destinations.
func ValidatePublicHTTPURL(ctx context.Context, rawURL string, resolve bool) (string, error) {
	u, err := parsePublicHTTPURL(rawURL)
	if err != nil {
		return "", err
	}

	if err := validatePublicHost(ctx, u.Hostname(), resolve); err != nil {
		return "", err
	}

	return u.String(), nil
}

// PublicDialContext resolves address, rejects non-public resolved IPs, and dials
// the selected IP directly so the transport does not perform an unchecked DNS
// lookup after validation.
func PublicDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	dialAddresses, err := publicDialAddresses(ctx, host, port)
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, dialAddress := range dialAddresses {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		conn, err := publicDialer.DialContext(ctx, network, dialAddress)
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func parsePublicHTTPURL(rawURL string) (*url.URL, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("empty URL")
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, errors.New(err).WithCode(errors.BadRequestCode)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("invalid HTTP scheme: %s", u.Scheme)
	}
	if u.Hostname() == "" {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("empty host")
	}

	u.User = nil
	u.RawQuery = ""
	u.Fragment = ""
	return u, nil
}

func validatePublicHost(ctx context.Context, host string, resolve bool) error {
	host = strings.Trim(strings.TrimSuffix(strings.ToLower(host), "."), "[]")
	if host == "" {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("empty host")
	}
	if isLocalHostname(host) {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("blocked non-public host: %s", host)
	}

	if ip := parseIP(host); ip != nil {
		return validatePublicIP(ip, host)
	}
	if !resolve {
		return nil
	}

	addrs, err := lookupIPAddr(ctx, host)
	if err != nil {
		return errors.New(err).WithCode(errors.BadRequestCode).WithMessagef("failed to resolve host: %s", host)
	}
	if len(addrs) == 0 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("failed to resolve host: %s", host)
	}
	for _, addr := range addrs {
		if err := validatePublicIP(addr.IP, host); err != nil {
			return err
		}
	}
	return nil
}

func publicDialAddress(ctx context.Context, host, port string) (string, error) {
	dialAddresses, err := publicDialAddresses(ctx, host, port)
	if err != nil {
		return "", err
	}
	return dialAddresses[0], nil
}

func publicDialAddresses(ctx context.Context, host, port string) ([]string, error) {
	host = strings.Trim(strings.TrimSuffix(strings.ToLower(host), "."), "[]")
	if err := validatePublicHost(ctx, host, false); err != nil {
		return nil, err
	}
	if ip := parseIP(host); ip != nil {
		return []string{net.JoinHostPort(ip.String(), port)}, nil
	}

	addrs, err := lookupIPAddr(ctx, host)
	if err != nil {
		return nil, errors.New(err).WithCode(errors.BadRequestCode).WithMessagef("failed to resolve host: %s", host)
	}
	if len(addrs) == 0 {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("failed to resolve host: %s", host)
	}
	var dialAddrs []string
	for _, addr := range addrs {
		if err := validatePublicIP(addr.IP, host); err != nil {
			return nil, err
		}
		dialAddrs = append(dialAddrs, net.JoinHostPort(addr.IP.String(), port))
	}
	return dialAddrs, nil
}

func isLocalHostname(host string) bool {
	return host == "localhost" ||
		strings.HasSuffix(host, ".localhost") ||
		strings.HasSuffix(host, ".local") ||
		strings.HasSuffix(host, ".internal") ||
		strings.HasSuffix(host, ".svc") ||
		strings.HasSuffix(host, ".svc.cluster.local") ||
		strings.HasSuffix(host, ".cluster.local")
}

func parseIP(host string) net.IP {
	if i := strings.LastIndex(host, "%"); i >= 0 {
		host = host[:i]
	}
	return net.ParseIP(host)
}

var specialUseIPNets []*net.IPNet

func init() {
	cidrs := []string{
		// IPv4 special-use / non-public ranges
		"0.0.0.0/8",          // Local system
		"10.0.0.0/8",         // Private-Use (RFC 1918)
		"100.64.0.0/10",      // Shared Address Space / CGNAT (RFC 6598)
		"127.0.0.0/8",        // Loopback
		"169.254.0.0/16",     // Link-Local
		"172.16.0.0/12",      // Private-Use (RFC 1918)
		"192.0.0.0/24",       // IETF Protocol Assignments (RFC 6890)
		"192.0.2.0/24",       // Documentation (TEST-NET-1, RFC 5737)
		"192.88.99.0/24",     // 6to4 Relay Anycast (RFC 7526)
		"192.168.0.0/16",     // Private-Use (RFC 1918)
		"198.18.0.0/15",      // Benchmarking (RFC 2544)
		"198.51.100.0/24",    // Documentation (TEST-NET-2, RFC 5737)
		"203.0.113.0/24",     // Documentation (TEST-NET-3, RFC 5737)
		"240.0.0.0/4",        // Reserved / Class E (RFC 1112)
		"255.255.255.255/32", // Limited Broadcast

		// IPv6 special-use / non-public ranges
		"::/128",        // Unspecified
		"::1/128",       // Loopback
		"100::/64",      // Discard-only address block (RFC 6666)
		"2001:2::/48",   // Benchmarking (RFC 5180)
		"2001:db8::/32", // Documentation (RFC 3849)
		"fc00::/7",      // Unique-Local / ULA (RFC 4193)
		"fe80::/10",     // Link-Local Unicast
		"ff00::/8",      // Multicast
	}

	for _, cidr := range cidrs {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			specialUseIPNets = append(specialUseIPNets, ipNet)
		}
	}
}

func validatePublicIP(ip net.IP, host string) error {
	if ip == nil {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("blocked non-public host: %s", host)
	}

	// Reject non-global-unicast addresses
	if !ip.IsGlobalUnicast() {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("blocked non-public host: %s", host)
	}

	// Reject private-use addresses (RFC 1918 and RFC 4193 ULA)
	if ip.IsPrivate() {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("blocked non-public host: %s", host)
	}

	// Reject other special-use IP ranges
	for _, ipNet := range specialUseIPNets {
		if ipNet.Contains(ip) {
			return errors.New(nil).WithCode(errors.BadRequestCode).WithMessagef("blocked non-public host: %s", host)
		}
	}

	return nil
}
