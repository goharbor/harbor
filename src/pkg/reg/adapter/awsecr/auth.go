// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package awsecr

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecr"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/lib/log"
)

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type awsAuthCredential struct {
	accessKey string
	awssvc    *ecr.Client

	cacheToken   *cacheToken
	cacheExpired *time.Time
}

type cacheToken struct {
	endpoint string
	user     string
	password string
	host     string
}

// DefaultCacheExpiredTime is expired timeout for aws auth token
const DefaultCacheExpiredTime = time.Hour * 1

func (a *awsAuthCredential) Modify(req *http.Request) error {
	// url maybe redirect to s3
	if !strings.Contains(req.URL.Host, ".ecr.") {
		return nil
	}
	if !a.isTokenValid() {
		endpoint, user, pass, expiresAt, err := a.getAuthorization(req.URL.String())

		if err != nil {
			return err
		}
		u, err := url.Parse(endpoint)
		if err != nil {
			return err
		}
		a.cacheToken = &cacheToken{}
		a.cacheToken.host = u.Host
		a.cacheToken.user = user
		a.cacheToken.password = pass
		a.cacheToken.endpoint = endpoint
		t := time.Now().Add(DefaultCacheExpiredTime)
		if t.Before(*expiresAt) {
			a.cacheExpired = &t
		} else {
			a.cacheExpired = expiresAt
		}
	}
	req.Host = a.cacheToken.host
	req.URL.Host = a.cacheToken.host
	req.SetBasicAuth(a.cacheToken.user, a.cacheToken.password)
	return nil
}

func getAwsCfg(region, accessKey, accessSecret string, insecure bool, caCertificate string, forceEndpoint *string) (*ecr.Client, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithHTTPClient(&http.Client{
			Transport: commonhttp.GetHTTPTransport(
				commonhttp.WithInsecure(insecure),
				commonhttp.WithCACert(caCertificate),
			),
		}),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, accessSecret, ""),
		),
	}

	if forceEndpoint != nil {
		opts = append(opts, config.WithBaseEndpoint(*forceEndpoint))
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return nil, err
	}

	client := ecr.NewFromConfig(cfg)

	return client, nil
}

func (a *awsAuthCredential) getAuthorization(url string) (string, string, string, *time.Time, error) {
	id, _, err := parseAccountRegion(url)
	if err != nil {
		return "", "", "", nil, err
	}

	var input *ecr.GetAuthorizationTokenInput
	if id != "" {
		input = &ecr.GetAuthorizationTokenInput{
			RegistryIds: []string{id},
		}
	}
	svc := a.awssvc
	result, err := svc.GetAuthorizationToken(context.TODO(), input)
	if err != nil {
		return "", "", "", nil, fmt.Errorf("ecr get authorization token: %w", err)
	}

	// Double check
	if len(result.AuthorizationData) == 0 {
		return "", "", "", nil, errors.New("no authorization token returned")
	}

	theOne := result.AuthorizationData[0]
	expiresAt := theOne.ExpiresAt
	payload, _ := base64.StdEncoding.DecodeString(*theOne.AuthorizationToken)
	pair := strings.SplitN(string(payload), ":", 2)

	log.Debugf("Aws Ecr getAuthorization %s result: %d %s...", a.accessKey, len(pair[1]), pair[1][:25])

	return *(theOne.ProxyEndpoint), pair[0], pair[1], expiresAt, nil
}

func (a *awsAuthCredential) isTokenValid() bool {
	if a.cacheToken == nil {
		return false
	}
	if a.cacheExpired == nil {
		return false
	}
	if time.Now().After(*a.cacheExpired) {
		a.cacheExpired = nil
		a.cacheToken = nil
		return false
	}
	return true
}

// NewAuth new aws auth
func NewAuth(accessKey string, awssvc *ecr.Client) Credential {
	return &awsAuthCredential{
		accessKey: accessKey,
		awssvc:    awssvc,
	}
}
