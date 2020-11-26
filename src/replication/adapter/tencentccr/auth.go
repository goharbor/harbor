package tencentccr

import (
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/http/modifier"
)

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type tencentAuthCredential struct {
	region    string
	accessKey string
	secretKey string
	insecure  bool

	cacheToken        *cacheToken
	cacheTokenExpired *time.Time
}
// cacheToken ...
type cacheToken struct {
	user     string
	password string
}

// NewAuth new tencent auth
func NewAuth(region, accessKey, secretKey string, insecure bool) Credential {
	return &tencentAuthCredential{
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
		insecure:  insecure,
	}
}

// isTokenValid todo support Tencent TCR
func (t *tencentAuthCredential) isTokenValid() bool {
	if t.cacheToken == nil {
		return false
	}
	if t.cacheTokenExpired == nil {
		return false
	}
	if time.Now().After(*t.cacheTokenExpired) {
		t.cacheTokenExpired = nil
		t.cacheToken = nil
		return false
	}
	return true
}

// Modfy todo support tencent tcr, current use baseauth
func (t *tencentAuthCredential) Modify(req *http.Request) error {
	req.SetBasicAuth(t.accessKey, t.secretKey)
	return nil
}
