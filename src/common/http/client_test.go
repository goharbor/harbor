package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHTTPTransport(t *testing.T) {
	transport := GetHTTPTransport(InsecureTransport)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	transport = GetHTTPTransport(SecureTransport)
	assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
}
