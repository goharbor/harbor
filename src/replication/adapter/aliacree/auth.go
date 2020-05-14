package aliacree

import (
	"fmt"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"

	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr_ee"
)

var _ Credential = &acrEEAuthCredential{}

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type acrEEAuthCredential struct {
	region              string
	instanceID          string
	accessKey           string
	secretKey           string
	cacheToken          *registryTemporaryToken
	cacheTokenExpiredAt time.Time
}

type registryTemporaryToken struct {
	user     string
	password string
}

// NewAuth will get a temporary docker registry username and password via acr-ee service API.
func NewAuth(region, instanceID, accessKey, secretKey string) Credential {
	return &acrEEAuthCredential{
		region:     region,
		instanceID: instanceID,
		accessKey:  accessKey,
		secretKey:  secretKey,
		cacheToken: &registryTemporaryToken{},
	}
}

func (a *acrEEAuthCredential) Modify(r *http.Request) (err error) {
	if !a.isCacheTokenValid() {
		log.Debugf("[acrEEAuthCredential.Modify.updateToken]Host: %s\n", r.Host)
		var client *cr_ee.Client
		client, err = cr_ee.NewClientWithAccessKey(a.region, a.accessKey, a.secretKey)
		if err != nil {
			return
		}

		req := cr_ee.CreateGetAuthorizationTokenRequest()
		req.Domain = fmt.Sprintf(endpointTpl, a.region)
		req.RegionId = a.region
		req.InstanceId = a.instanceID
		resp, err := client.GetAuthorizationToken(req)
		if err != nil {
			return err
		}
		if !resp.IsSuccess() {
			return fmt.Errorf("aliyun response unsuccessful")
		}
		a.cacheTokenExpiredAt = timeUnix(resp.ExpireTime).ToTime()
		a.cacheToken.user = resp.TempUsername
		a.cacheToken.password = resp.AuthorizationToken
	} else {
		log.Debug("[acrEEAuthCredential] USE CACHE TOKEN!!!")
	}

	r.SetBasicAuth(a.cacheToken.user, a.cacheToken.password)
	return
}

func (a *acrEEAuthCredential) isCacheTokenValid() bool {
	if &a.cacheTokenExpiredAt == nil {
		return false
	}
	if a.cacheToken == nil {
		return false
	}
	if time.Now().After(a.cacheTokenExpiredAt) {
		return false
	}

	return true
}
