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
	"fmt"
	"os"
	"strings"
)

const (
	// Internal TLS ENV
	internalTLSEnable        = "INTERNAL_TLS_ENABLED"
	internalVerifyClientCert = "INTERNAL_VERIFY_CLIENT_CERT"
	internalTLSKeyPath       = "INTERNAL_TLS_KEY_PATH"
	internalTLSCertPath      = "INTERNAL_TLS_CERT_PATH"
	internalTrustCAPath      = "INTERNAL_TLS_TRUST_CA_PATH"
)

// InternalTLSEnabled returns if internal TLS enabled
func InternalTLSEnabled() bool {
	return strings.ToLower(os.Getenv(internalTLSEnable)) == "true"
}

// InternalEnableVerifyClientCert returns if mTLS enabled
func InternalEnableVerifyClientCert() bool {
	return strings.ToLower(os.Getenv(internalVerifyClientCert)) == "true"
}

// GetInternalCertPair used to get internal cert and key pair from environment
func GetInternalCertPair() (tls.Certificate, error) {
	crtPath := os.Getenv(internalTLSCertPath)
	keyPath := os.Getenv(internalTLSKeyPath)
	cert, err := tls.LoadX509KeyPair(crtPath, keyPath)
	return cert, err
}

// GetInternalTLSConfig return a tls.Config for internal https communicate
func GetInternalTLSConfig() (*tls.Config, error) {
	// genrate key pair
	cert, err := GetInternalCertPair()
	if err != nil {
		return nil, fmt.Errorf("internal TLS enabled but can't get cert file %w", err)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, nil
}
