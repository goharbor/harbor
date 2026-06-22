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
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidatePublicHTTPURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		want    string
		wantErr bool
	}{
		{name: "public IPv4", rawURL: "http://1.1.1.1/hook?token=secret#fragment", want: "http://1.1.1.1/hook"},
		{name: "public host", rawURL: "https://example.com/hook", want: "https://example.com/hook"},
		{name: "user info stripped", rawURL: "https://user:pass@example.com/hook", want: "https://example.com/hook"},
		{name: "loopback", rawURL: "http://127.0.0.1/hook", wantErr: true},
		{name: "localhost", rawURL: "http://localhost/hook", wantErr: true},
		{name: "private IPv4", rawURL: "http://10.0.0.1/hook", wantErr: true},
		{name: "metadata IPv4", rawURL: "http://169.254.169.254/latest/meta-data", wantErr: true},
		{name: "IPv6 loopback", rawURL: "http://[::1]/hook", wantErr: true},
		{name: "IPv6 unique local", rawURL: "http://[fc00::1]/hook", wantErr: true},
		{name: "cluster local", rawURL: "http://service.namespace.svc.cluster.local/hook", wantErr: true},
		{name: "CGNAT IPv4", rawURL: "http://100.64.0.1/hook", wantErr: true},
		{name: "CGNAT IPv4 upper bound", rawURL: "http://100.127.255.254/hook", wantErr: true},
		{name: "local system IPv4", rawURL: "http://0.1.2.3/hook", wantErr: true},
		{name: "TEST-NET-1 IPv4", rawURL: "http://192.0.2.1/hook", wantErr: true},
		{name: "TEST-NET-2 IPv4", rawURL: "http://198.51.100.2/hook", wantErr: true},
		{name: "TEST-NET-3 IPv4", rawURL: "http://203.0.113.5/hook", wantErr: true},
		{name: "Class E reserved IPv4", rawURL: "http://240.0.0.1/hook", wantErr: true},
		{name: "IPv6 discard-only", rawURL: "http://[100::1]/hook", wantErr: true},
		{name: "IPv6 documentation", rawURL: "http://[2001:db8::1]/hook", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidatePublicHTTPURL(context.Background(), tt.rawURL, false)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidatePublicHTTPURLResolve(t *testing.T) {
	originalLookup := lookupIPAddr
	t.Cleanup(func() {
		lookupIPAddr = originalLookup
	})

	lookupIPAddr = func(_ context.Context, host string) ([]net.IPAddr, error) {
		require.Equal(t, "example.com", host)
		return []net.IPAddr{{IP: net.ParseIP("10.0.0.1")}}, nil
	}

	_, err := ValidatePublicHTTPURL(context.Background(), "https://example.com/hook", true)
	require.Error(t, err)
}

func TestPublicDialAddressUsesResolvedPublicIP(t *testing.T) {
	originalLookup := lookupIPAddr
	t.Cleanup(func() {
		lookupIPAddr = originalLookup
	})

	lookupIPAddr = func(_ context.Context, host string) ([]net.IPAddr, error) {
		require.Equal(t, "example.com", host)
		return []net.IPAddr{{IP: net.ParseIP("1.1.1.1")}}, nil
	}

	address, err := publicDialAddress(context.Background(), "example.com", "443")
	require.NoError(t, err)
	require.Equal(t, "1.1.1.1:443", address)
}

type mockDialer struct {
	dialFunc func(ctx context.Context, network, address string) (net.Conn, error)
}

func (m *mockDialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	return m.dialFunc(ctx, network, address)
}

func TestPublicDialContextFallback(t *testing.T) {
	originalLookup := lookupIPAddr
	originalDialer := publicDialer
	t.Cleanup(func() {
		lookupIPAddr = originalLookup
		publicDialer = originalDialer
	})

	lookupIPAddr = func(_ context.Context, host string) ([]net.IPAddr, error) {
		require.Equal(t, "example.com", host)
		return []net.IPAddr{
			{IP: net.ParseIP("1.1.1.1")},
			{IP: net.ParseIP("2.2.2.2")},
		}, nil
	}

	dialedAddrs := []string{}
	publicDialer = &mockDialer{
		dialFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
			dialedAddrs = append(dialedAddrs, address)
			if address == "1.1.1.1:443" {
				return nil, net.InvalidAddrError("unreachable")
			}
			return &net.TCPConn{}, nil
		},
	}

	conn, err := PublicDialContext(context.Background(), "tcp", "example.com:443")
	require.NoError(t, err)
	require.NotNil(t, conn)
	require.Equal(t, []string{"1.1.1.1:443", "2.2.2.2:443"}, dialedAddrs)
}

func TestPublicDialContextAllFail(t *testing.T) {
	originalLookup := lookupIPAddr
	originalDialer := publicDialer
	t.Cleanup(func() {
		lookupIPAddr = originalLookup
		publicDialer = originalDialer
	})

	lookupIPAddr = func(_ context.Context, host string) ([]net.IPAddr, error) {
		return []net.IPAddr{
			{IP: net.ParseIP("1.1.1.1")},
			{IP: net.ParseIP("2.2.2.2")},
		}, nil
	}

	publicDialer = &mockDialer{
		dialFunc: func(ctx context.Context, network, address string) (net.Conn, error) {
			return nil, net.InvalidAddrError("unreachable: " + address)
		},
	}

	conn, err := PublicDialContext(context.Background(), "tcp", "example.com:443")
	require.Error(t, err)
	require.Nil(t, conn)
	require.Contains(t, err.Error(), "unreachable: 2.2.2.2:443")
}
