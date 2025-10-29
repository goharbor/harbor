package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHTTPTransport(t *testing.T) {
	transport := GetHTTPTransport()
	assert.Equal(t, secureHTTPTransport, transport, "Transport should be secure")
	transport = GetHTTPTransport(WithInsecure(true))
	assert.Equal(t, insecureHTTPTransport, transport, "Transport should be insecure")
}

func TestValidateCACertificate(t *testing.T) {
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
			name: "invalid certificate - wrong header",
			cert: `-----BEGIN PRIVATE KEY-----
MIIDXTCCAkWgAwIBAgIJAKZ7cGiVgJqRMA0GCSqGSIb3DQEBCwUAMEU=
-----END PRIVATE KEY-----`,
			wantErr: true,
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
