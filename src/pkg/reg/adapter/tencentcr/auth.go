package tencentcr

import (
	"net/http"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
	tcr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tcr/v20190924"
)

// Credential ...
type Credential modifier.Modifier

var _ Credential = &qcloudAuthCredential{}

func (q *qcloudAuthCredential) Modify(r *http.Request) (err error) {
	if !q.isCacheTokenValid() {
		err = q.getTempInstanceToken()
		log.Debugf("qcloudAuthCredential.Modify.isCacheTokenValid.updateToken=%s, err=%v", q.cacheTokenExpiredAt, err)
		if err != nil {
			return
		}
	}
	r.SetBasicAuth(q.cacheTokener.username, q.cacheTokener.token)
	log.Debugf("[qcloudAuthCredential.Modify]Host: %v, header: %#v", r.Host, r.Header)
	return
}

func (q *qcloudAuthCredential) isCacheTokenValid() (ok bool) {
	if q.cacheTokenExpiredAt.IsZero() {
		return
	}
	if q.cacheTokener == nil {
		return
	}
	if time.Now().After(q.cacheTokenExpiredAt) {
		return
	}
	return true
}

// Implements interface Credential
type qcloudAuthCredential struct {
	registryID          *string
	client              *tcr.Client
	cacheTokener        *temporaryTokener
	cacheTokenExpiredAt time.Time
}

type temporaryTokener struct {
	username string
	token    string
}

// NewAuth ...
func NewAuth(registryID *string, client *tcr.Client) Credential {
	return &qcloudAuthCredential{
		registryID:   registryID,
		client:       client,
		cacheTokener: &temporaryTokener{},
	}
}

func (q *qcloudAuthCredential) getTempInstanceToken() (err error) {
	var req = tcr.NewCreateInstanceTokenRequest()
	req.RegistryId = q.registryID
	var resp *tcr.CreateInstanceTokenResponse
	resp, err = q.client.CreateInstanceToken(req)
	if err != nil {
		return
	}

	q.cacheTokener = &temporaryTokener{*resp.Response.Username, *resp.Response.Token}
	q.cacheTokenExpiredAt = time.Unix(*resp.Response.ExpTime/1e3, *resp.Response.ExpTime%1e3)
	log.Debugf("[qcloudAuthCredential.getTempInstanceToken]Update temp token=%#v, cacheTokenExpiredAt=%s, unix=%v", q.cacheTokener,
		q.cacheTokenExpiredAt.UTC().String(), *resp.Response.ExpTime)

	return
}

func isSecretID(key string) (ok bool) {
	return strings.Index(key, "AKID") == 0
}
