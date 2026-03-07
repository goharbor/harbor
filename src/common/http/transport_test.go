package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetHTTPTransport(t *testing.T) {
	transport := GetHTTPTransport()
	assert.Equal(t, secureHTTPTransport, transport, "Transport should be secure")
	transport = GetHTTPTransport(WithInsecure(true))
	assert.Equal(t, insecureHTTPTransport, transport, "Transport should be insecure")
}
