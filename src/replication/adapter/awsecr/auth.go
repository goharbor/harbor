package awsecr

import (
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsecrapi "github.com/aws/aws-sdk-go/service/ecr"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/common/utils/log"
	"net/http"
	"net/url"
	"strings"
)

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type awsAuthCredential struct {
	region        string
	access_key    string
	access_secret string
}

func (a *awsAuthCredential) Modify(req *http.Request) error {
	endpoint, user, pass, err := a.getAuthorization()

	if err != nil {
		return err
	}

	if u, err := url.Parse(endpoint); err != nil {
		return err
	} else {
		req.Host = u.Host
		req.URL.Host = u.Host
		req.SetBasicAuth(user, pass)
		return nil
	}
}

func (a *awsAuthCredential) getAuthorization() (string, string, string, error) {
	log.Infof("Aws Ecr getAuthorization %s", a.access_key)
	cred := credentials.NewStaticCredentials(
		a.access_key,
		a.access_secret,
		"")
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: cred,
		Region:      &a.region,
	}))

	svc := awsecrapi.New(sess)

	result, err := svc.GetAuthorizationToken(nil)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			return "", "", "", fmt.Errorf("%s: %s", aerr.Code(), aerr.Error())
		}

		return "", "", "", err
	}

	// Double check
	if len(result.AuthorizationData) == 0 {
		return "", "", "", errors.New("no authorization token returned")
	}

	theOne := result.AuthorizationData[0]
	payload, _ := base64.StdEncoding.DecodeString(*theOne.AuthorizationToken)
	pair := strings.SplitN(string(payload), ":", 2)

	log.Debugf("Aws Ecr getAuthorization %s result: %d %s...", a.access_key, len(pair[1]), pair[1][:25])

	return *(theOne.ProxyEndpoint), pair[0], pair[1], nil
}

func NewAuth(region, access_key, access_secret string) Credential {
	return &awsAuthCredential{
		region,
		access_key,
		access_secret,
	}
}
