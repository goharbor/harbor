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
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/goharbor/harbor/src/common/utils/log"
)

const (
	// Internal TLS ENV
	internalTLSEnable   = "INTERNAL_TLS_ENABLED"
	internalTLSKeyPath  = "INTERNAL_TLS_KEY_PATH"
	internalTLSCertPath = "INTERNAL_TLS_CERT_PATH"
	internalTrustCAPath = "INTERNAL_TLS_TRUST_CA_PATH"
)

// InternalTLSEnabled returns if internal TLS enabled
func InternalTLSEnabled() bool {
	iTLSEnabled := os.Getenv(internalTLSEnable)
	if strings.ToLower(iTLSEnabled) == "true" {
		return true
	}
	return false
}

// GetInternalCA used to get internal cert file from Env
func GetInternalCA(caPool *x509.CertPool) *x509.CertPool {
	if caPool == nil {
		caPool = x509.NewCertPool()
	}

	caPath := os.Getenv(internalTrustCAPath)
	if caPath != "" {
		caCert, err := ioutil.ReadFile(caPath)
		if err != nil {
			log.Errorf("read ca file %s failure %w", caPath, err)
		}
		if ok := caPool.AppendCertsFromPEM(caCert); !ok {
			log.Errorf("append ca to ca pool fail")
		} else {
			log.Infof("append trustCA %s success", caPath)
		}
	}

	return caPool
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
