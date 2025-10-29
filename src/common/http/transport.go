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
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/trace"
)

const (
	// InsecureTransport used to get the insecure http Transport
	InsecureTransport = iota
	// SecureTransport used to get the external secure http Transport
	SecureTransport
)

var (
	secureHTTPTransport   http.RoundTripper
	insecureHTTPTransport http.RoundTripper
)

func init() {
	insecureHTTPTransport = NewTransport(WithInsecureSkipVerify(true))
	if InternalTLSEnabled() {
		secureHTTPTransport = NewTransport(WithInternalTLSConfig())
	} else {
		secureHTTPTransport = NewTransport()
	}
}

func AddTracingWithGlobalTransport() {
	insecureHTTPTransport = otelhttp.NewTransport(insecureHTTPTransport, trace.HarborHTTPTraceOptions...)
	secureHTTPTransport = otelhttp.NewTransport(secureHTTPTransport, trace.HarborHTTPTraceOptions...)
}

// Use this instead of Default Transport in library because it sets ForceAttemptHTTP2 to true
// And that options introduced in go 1.13 will cause the https requests hang forever in replication environment
func newDefaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		TLSClientConfig:       &tls.Config{},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

// WithInternalTLSConfig returns a TransportOption that configures the transport to use the internal TLS configuration
func WithInternalTLSConfig() func(*http.Transport) {
	return func(tr *http.Transport) {
		tlsConfig, err := GetInternalTLSConfig()
		if err != nil {
			panic(err)
		}
		tr.TLSClientConfig = tlsConfig
	}
}

// WithInsecureSkipVerify returns a TransportOption that configures the transport to skip verification of the server's certificate
func WithInsecureSkipVerify(skipVerify bool) func(*http.Transport) {
	return func(tr *http.Transport) {
		tr.TLSClientConfig.InsecureSkipVerify = skipVerify
	}
}

// ValidateCACertificate validates if the provided certificate is in valid PEM format
func ValidateCACertificate(caCert string) error {
	if caCert == "" {
		return nil
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM([]byte(caCert)) {
		return errors.New("invalid CA certificate: not in valid PEM format")
	}
	return nil
}

// WithCustomCACert returns a TransportOption that configures custom CA certificate
func WithCustomCACert(caCert string) func(*http.Transport) {
	return func(tr *http.Transport) {
		if caCert == "" {
			return
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM([]byte(caCert)) {
			log.Errorf("Failed to append CA certificate to pool - invalid PEM format")
			return
		}

		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{}
		}
		tr.TLSClientConfig.RootCAs = caCertPool
	}
}

// NewTransport returns a new http.Transport with the specified options
func NewTransport(opts ...func(*http.Transport)) http.RoundTripper {
	tr := newDefaultTransport()
	for _, opt := range opts {
		opt(tr)
	}
	return tr
}

// TransportConfig is the configuration for http transport
type TransportConfig struct {
	Insecure      bool
	CACertificate string
}

// TransportOption is the option for http transport
type TransportOption func(*TransportConfig)

// WithInsecure returns a TransportOption that configures the transport to skip verification of the server's certificate
func WithInsecure(skipVerify bool) TransportOption {
	return func(cfg *TransportConfig) {
		cfg.Insecure = skipVerify
	}
}

// WithCACert returns a TransportOption that configures custom CA certificate
func WithCACert(caCert string) TransportOption {
	return func(cfg *TransportConfig) {
		cfg.CACertificate = caCert
	}
}

// GetHTTPTransport returns HttpTransport based on insecure configuration and CA certificate.
//
// Priority:
//  1. Custom CA certificate (if provided) - creates a new transport with custom CA
//  2. Insecure mode (if enabled) - returns shared transport that skips TLS verification
//  3. Default - returns shared transport that uses system CA pool
//
// Backward Compatibility:
// Existing Harbor installations that rely on system-level CA trust stores will continue
// to work after upgrade. The custom CA certificate feature is optional, and when no
// custom CA is provided, the system CA pool is used (option 3 above).
func GetHTTPTransport(opts ...TransportOption) http.RoundTripper {
	cfg := &TransportConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	if cfg.CACertificate != "" {
		return NewTransport(
			WithCustomCACert(cfg.CACertificate),
		)
	}

	if cfg.Insecure {
		return insecureHTTPTransport
	}
	return secureHTTPTransport
}
