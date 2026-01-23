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
	"encoding/pem"
	"fmt"
	"net"
	"net/http"
	"strings"
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

// normalizePEM normalizes PEM data by trimming whitespace and converting
func normalizePEM(cert string) string {
	cert = strings.TrimSpace(cert)
	cert = strings.ReplaceAll(cert, "\r\n", "\n")
	cert = strings.ReplaceAll(cert, "\r", "\n")
	return cert
}

// ValidateCACertificate validates whether the provided CA certificate string
// contains at least one valid PEM-encoded x509 certificate.
func ValidateCACertificate(caCert string) error {
	caCert = normalizePEM(caCert)
	if caCert == "" {
		return nil
	}

	// Attempt to parse one or more certificates from the provided PEM
	certs, err := parseCertificatesFromPEM(caCert)
	if err != nil {
		return fmt.Errorf("invalid CA certificate: %w", err)
	}

	if len(certs) == 0 {
		return errors.New("invalid CA certificate: no valid certificates found in PEM data")
	}

	return nil
}

// parseCertificatesFromPEM decodes all PEM blocks and parses certificates.
func parseCertificatesFromPEM(pemData string) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	rest := []byte(pemData)

	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %v", err)
		}

		certs = append(certs, cert)
	}

	return certs, nil
}

// WithCustomCACert returns a TransportOption that configures custom CA certificates (supports chains)
func WithCustomCACert(caCert string) func(*http.Transport) {
	return func(tr *http.Transport) {
		caCert = normalizePEM(caCert)
		if caCert == "" {
			log.Debugf("No custom CA certificate provided; skipping configuration")
			return
		}

		certs, err := parseCertificatesFromPEM(caCert)
		if err != nil {
			log.Warningf("Failed to parse CA certificate: %v", err)
			return
		}

		if len(certs) == 0 {
			log.Warningf("No valid certificates found in provided CA PEM")
			return
		}

		caCertPool := x509.NewCertPool()
		for _, cert := range certs {
			caCertPool.AddCert(cert)
		}

		if tr.TLSClientConfig == nil {
			tr.TLSClientConfig = &tls.Config{}
		}

		tr.TLSClientConfig.RootCAs = caCertPool
		log.Debugf("Configured HTTP transport with custom CA certificate.")
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
