package security

import (
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	securitysecret "github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/lib/config"
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

func TestFromJobservice(t *testing.T) {
	// no security ctx should return false
	req1, _ := http.NewRequest(http.MethodHead, "/api", nil)
	assert.False(t, FromJobservice(req1))
	// other username should return false
	req2, _ := http.NewRequest(http.MethodHead, "/api", nil)
	secCtx1 := local.NewSecurityContext(&models.User{UserID: 1, Username: "test-user"})
	req2 = req2.WithContext(security.NewContext(req2.Context(), secCtx1))
	assert.False(t, FromJobservice(req2))
	// secret ctx from jobservice should return true
	req3, _ := http.NewRequest(http.MethodHead, "/api", nil)
	config.Init()
	secCtx2 := securitysecret.NewSecurityContext(config.JobserviceSecret(), config.SecretStore)
	req3 = req3.WithContext(security.NewContext(req3.Context(), secCtx2))
	assert.True(t, FromJobservice(req3))
}

func TestFromJobRetention(t *testing.T) {
	// return false if req is nil
	assert.False(t, FromJobRetention(nil))
	// return false if req has no header
	req1, err := http.NewRequest("GET", "http://localhost:8080/api", nil)
	assert.NoError(t, err)
	assert.False(t, FromJobRetention(req1))
	// return true if header has retention vendor type
	req2, err := http.NewRequest("GET", "http://localhost:8080/api", nil)
	assert.NoError(t, err)
	req2.Header.Set("VendorType", "RETENTION")
	assert.True(t, FromJobRetention(req2))
}
