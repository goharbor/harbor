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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	awsecrapi "github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	libconfig "github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Credential ...
type Credential modifier.Modifier

// Implements interface Credential
type awsAuthCredential struct {
	awssvc *awsecrapi.Client

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

var sourceIdentityPattern = regexp.MustCompile(`^[\w+=,.@-]{2,64}$`)

type roleCredential struct {
	RoleARN                  string `json:"role_arn"`
	SourceIdentity           string `json:"source_identity,omitempty"`
	SourceAccessKey          string `json:"source_access_key,omitempty"`
	WebIdentityTokenFilePath string `json:"web_identity_token_file,omitempty"`
}

func (a *awsAuthCredential) Modify(req *http.Request) error {
	// url maybe redirect to s3
	if !strings.Contains(strings.ToLower(req.URL.Host), ".ecr.") {
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

func getAwsSvc(region string, credential *model.Credential, insecure bool, caCertificate string, forceEndpoint *string) (*awsecrapi.Client, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
		config.WithHTTPClient(&http.Client{
			Transport: commonhttp.GetHTTPTransport(
				commonhttp.WithInsecure(insecure),
				commonhttp.WithCACert(caCertificate),
			),
			Timeout: libconfig.RegistryHTTPClientTimeout(),
		}),
	}

	log.Debug("Aws Ecr getAuthorization")
	credentialType := model.CredentialTypeBasic
	if credential != nil && credential.Type != "" {
		credentialType = credential.Type
	}

	var role roleCredential
	switch credentialType {
	case model.CredentialTypeBasic:
		if credential != nil {
			if credential.AccessKey == "" && credential.AccessSecret != "" {
				return nil, errors.New("AWS access key is required when an access secret is configured")
			}
			if credential.AccessKey != "" && credential.AccessSecret == "" {
				return nil, errors.New("AWS access secret is required when an access key is configured")
			}
			if credential.AccessKey != "" {
				opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
					credential.AccessKey, credential.AccessSecret, "")))
			}
		}
	case model.CredentialTypeAWSWebIdentity, model.CredentialTypeAWSAssumeRole:
		var err error
		role, err = parseRoleCredential(credentialType, credential)
		if err != nil {
			return nil, err
		}
		if role.SourceAccessKey != "" {
			opts = append(opts, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				role.SourceAccessKey, credential.AccessSecret, "")))
		}
	default:
		return nil, fmt.Errorf("unsupported AWS credential type %q", credentialType)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), opts...)
	if err != nil {
		return nil, err
	}

	if forceEndpoint != nil {
		cfg.BaseEndpoint = forceEndpoint
	}

	switch credentialType {
	case model.CredentialTypeAWSWebIdentity:
		provider := stscreds.NewWebIdentityRoleProvider(
			sts.NewFromConfig(cfg), role.RoleARN, stscreds.IdentityTokenFile(role.WebIdentityTokenFilePath))
		cfg.Credentials = aws.NewCredentialsCache(provider)
	case model.CredentialTypeAWSAssumeRole:
		provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), role.RoleARN, func(options *stscreds.AssumeRoleOptions) {
			if role.SourceIdentity != "" {
				options.SourceIdentity = aws.String(role.SourceIdentity)
			}
		})
		cfg.Credentials = aws.NewCredentialsCache(provider)
	}

	svc := awsecrapi.NewFromConfig(cfg)
	return svc, nil
}

func parseRoleCredential(credentialType string, credential *model.Credential) (roleCredential, error) {
	var role roleCredential
	if credential == nil {
		return role, errors.New("AWS role credential is required")
	}

	decoder := json.NewDecoder(strings.NewReader(credential.AccessKey))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&role); err != nil {
		return role, fmt.Errorf("invalid AWS role credential: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return role, errors.New("invalid AWS role credential: multiple JSON values")
	}
	if role.RoleARN == "" {
		return role, errors.New("AWS role ARN is required")
	}

	switch credentialType {
	case model.CredentialTypeAWSWebIdentity:
		if role.WebIdentityTokenFilePath == "" {
			return role, errors.New("AWS web identity token file is required")
		}
		if !filepath.IsAbs(role.WebIdentityTokenFilePath) {
			return role, errors.New("AWS web identity token file must be an absolute local path")
		}
		if role.SourceIdentity != "" {
			return role, errors.New("AWS web identity source identity must be supplied by the token claim")
		}
		if role.SourceAccessKey != "" || credential.AccessSecret != "" {
			return role, errors.New("AWS web identity does not accept source access keys")
		}
	case model.CredentialTypeAWSAssumeRole:
		if role.WebIdentityTokenFilePath != "" {
			return role, errors.New("AWS assume role does not accept a web identity token file")
		}
		if role.SourceAccessKey == "" && credential.AccessSecret != "" {
			return role, errors.New("AWS source access key is required when a source access secret is configured")
		}
		if role.SourceAccessKey != "" && credential.AccessSecret == "" {
			return role, errors.New("AWS source access secret is required when a source access key is configured")
		}
		if role.SourceIdentity != "" && (!sourceIdentityPattern.MatchString(role.SourceIdentity) ||
			strings.HasPrefix(strings.ToLower(role.SourceIdentity), "aws:")) {
			return role, errors.New("invalid AWS source identity")
		}
	}

	return role, nil
}

func (a *awsAuthCredential) getAuthorization(url string) (string, string, string, *time.Time, error) {
	id, _, err := parseAccountRegion(url)
	if err != nil {
		return "", "", "", nil, err
	}

	var input *awsecrapi.GetAuthorizationTokenInput
	if id != "" {
		input = &awsecrapi.GetAuthorizationTokenInput{RegistryIds: []string{id}}
	} else {
		input = &awsecrapi.GetAuthorizationTokenInput{}
	}
	svc := a.awssvc
	result, err := svc.GetAuthorizationToken(context.TODO(), input)
	if err != nil {
		var aerr smithy.APIError
		if errors.As(err, &aerr) {
			return "", "", "", nil, fmt.Errorf("%s: %s", aerr.ErrorCode(), aerr.ErrorMessage())
		}

		return "", "", "", nil, err
	}

	// Double check
	if len(result.AuthorizationData) == 0 {
		return "", "", "", nil, errors.New("no authorization token returned")
	}

	theOne := result.AuthorizationData[0]
	expiresAt := theOne.ExpiresAt
	payload, _ := base64.StdEncoding.DecodeString(*theOne.AuthorizationToken)
	pair := strings.SplitN(string(payload), ":", 2)

	log.Debugf("Aws Ecr getAuthorization succeeded, token length: %d", len(pair[1]))

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
func NewAuth(awssvc *awsecrapi.Client) Credential {
	return &awsAuthCredential{
		awssvc: awssvc,
	}
}
