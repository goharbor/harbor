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

package notification

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib"
)

const (
	secure   = "secure"
	insecure = "insecure"

	// Max retry has the same meaning as max fails.
	maxFails = "JOBSERVICE_WEBHOOK_JOB_MAX_RETRY"
	// http client timeout for webhook job(seconds).
	httpClientTimeout = "JOBSERVICE_WEBHOOK_JOB_HTTP_CLIENT_TIMEOUT"
)

var (
	// timeout records the timeout for http client
	timeout    time.Duration
	httpHelper *HTTPHelper
)

func init() {
	// default timeout is 3 seconds
	timeout = 3 * time.Second
	if envTimeout, exist := os.LookupEnv(httpClientTimeout); exist {
		t, err := strconv.ParseInt(envTimeout, 10, 64)
		if err != nil {
			logger.Warningf("Failed to parse timeout from environment, error: %v", err)
			return
		}

		timeout = time.Duration(t) * time.Second
		logger.Debugf("Set the http client timeout to %v for webhook job", timeout)
	}
}

// HTTPHelper in charge of sending notification messages to remote endpoint
type HTTPHelper struct {
	clients map[string]*http.Client
}

func init() {
	httpHelper = &HTTPHelper{
		clients: map[string]*http.Client{},
	}
	httpHelper.clients[secure] = &http.Client{
		Transport:     webhookTransport(false),
		Timeout:       timeout,
		CheckRedirect: noRedirect,
	}
	httpHelper.clients[insecure] = &http.Client{
		Transport:     webhookTransport(true),
		Timeout:       timeout,
		CheckRedirect: noRedirect,
	}
}

type dialerInterface interface {
	DialContext(ctx context.Context, network, addr string) (net.Conn, error)
}

type contextKey string

const (
	useProxyKey   contextKey = "useProxy"
	targetHostKey contextKey = "targetHost"
)

var (
	dialer dialerInterface = &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
)

func webhookTransport(insecure bool) http.RoundTripper {
	opts := []func(*http.Transport){
		func(tr *http.Transport) {
			tr.Proxy = http.ProxyFromEnvironment
			tr.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
				if useProxy, ok := ctx.Value(useProxyKey).(bool); ok && useProxy {
					return dialer.DialContext(ctx, network, addr)
				}
				return lib.PublicDialContext(ctx, network, addr)
			}
		},
	}
	if insecure {
		opts = append(opts, commonhttp.WithInsecureSkipVerify(true))
	}
	underlying := commonhttp.NewTransport(opts...)
	return &ssrfProxyRoundTripper{
		insecure:   insecure,
		underlying: underlying,
	}
}

type ssrfProxyRoundTripper struct {
	insecure     bool
	underlying   http.RoundTripper
	transportsMu sync.Mutex
	transports   map[string]*http.Transport
	hosts        []string // for LRU cache eviction
}

func (s *ssrfProxyRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	useProxy, ok := req.Context().Value(useProxyKey).(bool)
	if !ok || !useProxy {
		return s.underlying.RoundTrip(req)
	}

	host := req.URL.Hostname()
	port := req.URL.Port()
	if port == "" {
		if req.URL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	// Resolve and validate target host, pinning the IP address to prevent DNS-rebinding SSRF
	dialAddrs, err := lib.PublicDialAddresses(req.Context(), host, port)
	if err != nil {
		return nil, err
	}

	s.transportsMu.Lock()
	if s.transports == nil {
		s.transports = make(map[string]*http.Transport)
	}
	tr, exists := s.transports[host]
	if !exists {
		// LRU Eviction: If map has 100 or more entries, evict the oldest
		if len(s.hosts) >= 100 {
			oldest := s.hosts[0]
			s.hosts = s.hosts[1:]
			if oldTr, ok := s.transports[oldest]; ok {
				oldTr.CloseIdleConnections()
				delete(s.transports, oldest)
			}
		}

		tr = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, addr)
			},
			DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				dialHost, _, err := net.SplitHostPort(addr)
				if err != nil {
					dialHost = addr
				}
				serverName := dialHost
				if targetHost, ok := ctx.Value(targetHostKey).(string); ok {
					if dialHost == targetHost || net.ParseIP(dialHost) != nil {
						serverName = targetHost
					}
				}
				plainConn, err := dialer.DialContext(ctx, network, addr)
				if err != nil {
					return nil, err
				}
				tlsConfig := &tls.Config{
					ServerName:         serverName,
					InsecureSkipVerify: s.insecure,
				}
				tlsConn := tls.Client(plainConn, tlsConfig)
				if err := tlsConn.HandshakeContext(ctx); err != nil {
					plainConn.Close()
					return nil, err
				}
				return tlsConn, nil
			},
			TLSClientConfig: &tls.Config{
				ServerName:         host,
				InsecureSkipVerify: s.insecure,
			},
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		}
		s.transports[host] = tr
		s.hosts = append(s.hosts, host)
	} else {
		// Update LRU position: move host to the end
		for i, h := range s.hosts {
			if h == host {
				s.hosts = append(s.hosts[:i], s.hosts[i+1:]...)
				break
			}
		}
		s.hosts = append(s.hosts, host)
	}
	s.transportsMu.Unlock()

	var lastErr error
	for _, dialAddr := range dialAddrs {
		if err := req.Context().Err(); err != nil {
			return nil, err
		}

		ip, _, err := net.SplitHostPort(dialAddr)
		if err != nil {
			lastErr = err
			continue
		}

		ctx := context.WithValue(req.Context(), targetHostKey, host)
		clonedReq := req.Clone(ctx)
		if clonedReq.Host == "" {
			clonedReq.Host = req.URL.Host
		}
		clonedReq.URL.Host = net.JoinHostPort(ip, port)

		if req.GetBody != nil {
			body, err := req.GetBody()
			if err != nil {
				lastErr = err
				continue
			}
			clonedReq.Body = body
		}

		resp, err := tr.RoundTrip(clonedReq)
		if err == nil {
			return resp, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func noRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}
