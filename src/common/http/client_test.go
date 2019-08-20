package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHTTPTransport(t *testing.T) {
	transport := GetHTTPTransport(true)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify)
	transport = GetHTTPTransport(false)
	assert.False(t, transport.TLSClientConfig.InsecureSkipVerify)
}
