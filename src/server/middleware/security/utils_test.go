package security

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBearerToken(t *testing.T) {
	req1, _ := http.NewRequest(http.MethodHead, "/api", nil)
	req1.Header.Set("Authorization", "Bearer token")
	req2, _ := http.NewRequest(http.MethodPut, "/api", nil)
	req2.SetBasicAuth("", "")
	req3, _ := http.NewRequest(http.MethodPut, "/api", nil)
	cases := []struct {
		request *http.Request
		token   string
	}{
		{
			request: req1,
			token:   "token",
		},
		{
			request: req2,
			token:   "",
		},
		{
			request: req3,
			token:   "",
		},
		{
			request: nil,
			token:   "",
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.token, bearerToken(c.request))
	}
}
