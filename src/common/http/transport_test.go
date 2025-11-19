package http

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateSelfSignedCert(t *testing.T) string {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "Test CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return string(certPEM)
}

func TestGetHTTPTransport(t *testing.T) {
	transport := GetHTTPTransport()
	assert.Equal(t, secureHTTPTransport, transport, "Transport should be secure")
	transport = GetHTTPTransport(WithInsecure(true))
	assert.Equal(t, insecureHTTPTransport, transport, "Transport should be insecure")
}

func TestValidateCACertificate(t *testing.T) {
	validCert := generateSelfSignedCert(t)

	tests := []struct {
		name    string
		cert    string
		wantErr bool
	}{
		{
			name:    "empty certificate",
			cert:    "",
			wantErr: false,
		},
		{
			name:    "valid self-signed certificate",
			cert:    validCert,
			wantErr: false,
		},
		{
			name:    "invalid certificate - not PEM format",
			cert:    "this is not a certificate",
			wantErr: true,
		},
		{
			name:    "invalid certificate - missing header",
			cert:    "MIIDXTCCAkWgAwIBAgIJAKZ7cGiVgJqRMA0GCSqGSIb3DQEBCwUAMEU=",
			wantErr: true,
		},
		{
			name: "invalid certificate - wrong PEM type",
			cert: `-----BEGIN PRIVATE KEY-----
MIIDXTCCAkWgAwIBAgIJAKZ7cGiVgJqRMA0GCSqGSIb3DQEBCwUAMEU=
-----END PRIVATE KEY-----`,
			wantErr: true,
		},
		{
			name: "invalid certificate - corrupted base64",
			cert: `-----BEGIN CERTIFICATE-----
MIIDXTCCAkWgAwIBAgIJAKZ7cGiVgJqRMA0GCSqGSIb3DQEBCwUAMEU!!!INVALID!!!
-----END CERTIFICATE-----`,
			wantErr: true,
		},
		{
			name: "invalid certificate - malformed PEM",
			cert: `-----BEGIN CERTIFICATE-----
INVALID_DATA
-----END CERTIFICATE-----`,
			wantErr: true,
		},
		{
			name:    "whitespace only",
			cert:    "   \n\t  ",
			wantErr: false, // Trimmed to empty string, which is valid (no error)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCACertificate(tt.cert)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid CA certificate")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestWithCustomCACert(t *testing.T) {
	validCert := generateSelfSignedCert(t)

	tests := []struct {
		name          string
		cert          string
		expectRootCAs bool
		expectCerts   int
	}{
		{
			name:          "empty certificate",
			cert:          "",
			expectRootCAs: false,
			expectCerts:   0,
		},
		{
			name:          "valid single certificate",
			cert:          validCert,
			expectRootCAs: true,
			expectCerts:   1,
		},
		{
			name:          "invalid PEM format - no warning, early return",
			cert:          "not a valid certificate",
			expectRootCAs: false,
			expectCerts:   0,
		},
		{
			name: "corrupted certificate data - parse error",
			cert: `-----BEGIN CERTIFICATE-----
INVALID_BASE64_DATA!!!
-----END CERTIFICATE-----`,
			expectRootCAs: false,
			expectCerts:   0,
		},
		{
			name: "wrong PEM type - skipped",
			cert: `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC=
-----END PRIVATE KEY-----`,
			expectRootCAs: false,
			expectCerts:   0,
		},
		{
			name:          "whitespace only",
			cert:          "   \n\t  ",
			expectRootCAs: false,
			expectCerts:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &http.Transport{}
			opt := WithCustomCACert(tt.cert)
			opt(tr)

			if tt.expectRootCAs {
				require.NotNil(t, tr.TLSClientConfig, "TLSClientConfig should be set")
				require.NotNil(t, tr.TLSClientConfig.RootCAs, "RootCAs should be set")
				subjects := tr.TLSClientConfig.RootCAs.Subjects()
				assert.Len(t, subjects, tt.expectCerts, "Should have expected number of certificates")
			} else {
				// For invalid/empty certs, the function returns early without setting RootCAs
				if tr.TLSClientConfig != nil {
					assert.Nil(t, tr.TLSClientConfig.RootCAs, "RootCAs should not be set for invalid/empty cert")
				}
			}
		})
	}
}

func TestGetHTTPTransportWithCACert(t *testing.T) {
	validCert := generateSelfSignedCert(t)

	t.Run("valid certificate", func(t *testing.T) {
		transport := GetHTTPTransport(WithCACert(validCert))
		require.NotNil(t, transport, "transport should not be nil")

		httpTransport, ok := transport.(*http.Transport)
		require.True(t, ok, "should be *http.Transport")
		require.NotNil(t, httpTransport.TLSClientConfig, "TLSClientConfig should be set")
		require.NotNil(t, httpTransport.TLSClientConfig.RootCAs, "RootCAs should be set")

		subjects := httpTransport.TLSClientConfig.RootCAs.Subjects()
		assert.Len(t, subjects, 1, "Should have one certificate")
	})

	t.Run("invalid certificate - returns default secure transport", func(t *testing.T) {
		invalidCert := `-----BEGIN CERTIFICATE-----
INVALID!!!
-----END CERTIFICATE-----`

		transport := GetHTTPTransport(WithCACert(invalidCert))
		require.NotNil(t, transport, "transport should not be nil")

		httpTransport, ok := transport.(*http.Transport)
		require.True(t, ok, "should be *http.Transport")

		if httpTransport.TLSClientConfig != nil {
			assert.False(t, httpTransport.TLSClientConfig.InsecureSkipVerify, "Should not skip TLS verification")
		}
	})

	t.Run("empty certificate", func(t *testing.T) {
		transport := GetHTTPTransport(WithCACert(""))
		assert.Equal(t, secureHTTPTransport, transport, "Should return secure transport for empty cert")
	})
}

func TestGetHTTPTransportPriority(t *testing.T) {
	validCert := generateSelfSignedCert(t)

	t.Run("CA cert takes priority over insecure", func(t *testing.T) {
		transport := GetHTTPTransport(WithInsecure(true), WithCACert(validCert))
		httpTransport, ok := transport.(*http.Transport)
		require.True(t, ok, "should be *http.Transport")
		require.NotNil(t, httpTransport.TLSClientConfig, "TLSClientConfig should be set")
		require.NotNil(t, httpTransport.TLSClientConfig.RootCAs, "CA cert should take priority over insecure")

		subjects := httpTransport.TLSClientConfig.RootCAs.Subjects()
		assert.Len(t, subjects, 1, "Should have one certificate")
	})

	t.Run("insecure when no CA cert", func(t *testing.T) {
		transport := GetHTTPTransport(WithInsecure(true))
		assert.Equal(t, insecureHTTPTransport, transport, "Should return insecure transport")
	})
}

func TestTransportInsecureSkipVerify(t *testing.T) {
	transport := GetHTTPTransport(WithInsecure(true))
	httpTransport, ok := transport.(*http.Transport)
	require.True(t, ok)
	require.NotNil(t, httpTransport.TLSClientConfig)
	assert.True(t, httpTransport.TLSClientConfig.InsecureSkipVerify)
}

func TestTransportSecureDefault(t *testing.T) {
	transport := GetHTTPTransport()
	httpTransport, ok := transport.(*http.Transport)
	require.True(t, ok)
	if httpTransport.TLSClientConfig != nil {
		assert.False(t, httpTransport.TLSClientConfig.InsecureSkipVerify)
	}
}

func TestMultipleCertificates(t *testing.T) {
	cert1 := generateSelfSignedCert(t)
	cert2 := generateSelfSignedCert(t)

	multipleCerts := cert1 + "\n" + cert2

	t.Run("multiple valid certificates", func(t *testing.T) {
		tr := &http.Transport{}
		opt := WithCustomCACert(multipleCerts)
		opt(tr)

		require.NotNil(t, tr.TLSClientConfig, "TLSClientConfig should be set")
		require.NotNil(t, tr.TLSClientConfig.RootCAs, "RootCAs should be set")

		subjects := tr.TLSClientConfig.RootCAs.Subjects()
		assert.Len(t, subjects, 2, "Should have two certificates")
	})

	t.Run("validate multiple certificates", func(t *testing.T) {
		err := ValidateCACertificate(multipleCerts)
		require.NoError(t, err, "Should validate multiple certificates")
	})
}

func TestMixedPEMBlocks(t *testing.T) {
	validCert := generateSelfSignedCert(t)

	mixedPEM := validCert + `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC=
-----END PRIVATE KEY-----
`

	t.Run("mixed PEM blocks - only certificates extracted", func(t *testing.T) {
		tr := &http.Transport{}
		opt := WithCustomCACert(mixedPEM)
		opt(tr)

		require.NotNil(t, tr.TLSClientConfig, "TLSClientConfig should be set")
		require.NotNil(t, tr.TLSClientConfig.RootCAs, "RootCAs should be set")

		subjects := tr.TLSClientConfig.RootCAs.Subjects()
		assert.Len(t, subjects, 1, "Should have one certificate")
	})
}

func TestGetHTTPTransportWithMultipleCerts(t *testing.T) {
	cert1 := generateSelfSignedCert(t)
	cert2 := generateSelfSignedCert(t)
	multipleCerts := cert1 + "\n" + cert2

	transport := GetHTTPTransport(WithCACert(multipleCerts))
	require.NotNil(t, transport)

	httpTransport, ok := transport.(*http.Transport)
	require.True(t, ok, "should be *http.Transport")
	require.NotNil(t, httpTransport.TLSClientConfig)
	require.NotNil(t, httpTransport.TLSClientConfig.RootCAs)

	subjects := httpTransport.TLSClientConfig.RootCAs.Subjects()
	assert.Len(t, subjects, 2, "Should have two certificates")
}
