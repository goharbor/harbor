package aliacr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cr"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
)

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type aliyunAuthCredential struct {
	region              string
	accessKey           string
	secretKey           string
	instanceId          string
	cacheToken          *registryTemporaryToken
	cacheTokenExpiredAt time.Time
}

type registryTemporaryToken struct {
	user     string
	password string
}

var _ Credential = &aliyunAuthCredential{}

// NewAuth will get a temporary docker registry username and password via aliyun cr service API.
func NewAuth(region, accessKey, secretKey, instanceId string) Credential {
	return &aliyunAuthCredential{
		region:     region,
		accessKey:  accessKey,
		secretKey:  secretKey,
		cacheToken: &registryTemporaryToken{},
		instanceId: instanceId,
	}
}

func (a *aliyunAuthCredential) Modify(r *http.Request) (err error) {
	if !a.isCacheTokenValid() {
		log.Debugf("[aliyunAuthCredential.Modify.updateToken]Host: %s\n", r.Host)
		var client *cr.Client
		client, err = cr.NewClientWithAccessKey(a.region, a.accessKey, a.secretKey)
		if err != nil {
			return err
		}

		request := requests.NewCommonRequest()
		request.Method = "GET"
		request.Scheme = "https" // https | http
		request.Domain = fmt.Sprintf(endpointTpl, a.region)
		request.Version = "2018-12-01"
		request.ApiName = "GetAuthorizationToken"
		request.QueryParams["RegionId"] = a.region
		request.QueryParams["InstanceId"] = a.instanceId
		tokenResponse, err := client.ProcessCommonRequest(request)
		if err != nil {
			return err
		}
		var v authorizationToken
		err = json.Unmarshal(tokenResponse.GetHttpContentBytes(), &v)
		if err != nil {
			return err
		}
		a.cacheTokenExpiredAt = v.ExpireTime.ToTime()
		a.cacheToken.user = v.TempUserName
		a.cacheToken.password = v.AuthorizationToken
	} else {
		log.Debug("[aliyunAuthCredential] USE CACHE TOKEN!!!")
	}

	r.SetBasicAuth(a.cacheToken.user, a.cacheToken.password)
	return
}

func (a *aliyunAuthCredential) isCacheTokenValid() bool {
	if a.cacheTokenExpiredAt.IsZero() {
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
