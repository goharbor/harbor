// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package uaa

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vmware/harbor/src/common/utils/log"
	"golang.org/x/oauth2"
)

// Client provides funcs to interact with UAA.
type Client interface {
	//PasswordAuth accepts username and password, return a token if it's valid.
	PasswordAuth(username, password string) (*oauth2.Token, error)
	//GetUserInfoByToken send the token to OIDC endpoint to get user info, currently it's also used to validate the token.
	GetUserInfo(token string) (*UserInfo, error)
}

// ClientConfig values to initialize UAA Client
type ClientConfig struct {
	ClientID      string
	ClientSecret  string
	Endpoint      string
	SkipTLSVerify bool
	//Absolut path for CA root used to communicate with UAA, only effective when skipTLSVerify set to false.
	CARootPath string
}

// UserInfo represent the JSON object of a userinfo response from UAA.
// As the response varies, this struct will contain only a subset of attributes
// that may be used in Harbor
type UserInfo struct {
	UserID   string `json:"user_id"`
	Sub      string `json:"sub"`
	UserName string `json:"user_name"`
	Name     string `json:"name"`
	Email    string `json:"email"`
}

// DefaultClient leverages oauth2 pacakge for oauth features
type defaultClient struct {
	httpClient *http.Client
	oauth2Cfg  *oauth2.Config
	endpoint   string
	//TODO: add public key, etc...
}

func (dc *defaultClient) PasswordAuth(username, password string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, dc.httpClient)
	return dc.oauth2Cfg.PasswordCredentialsToken(ctx, username, password)
}

func (dc *defaultClient) GetUserInfo(token string) (*UserInfo, error) {
	userInfoURL := dc.endpoint + "/uaa/userinfo"
	req, err := http.NewRequest(http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "bearer "+token)
	resp, err := dc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	info := &UserInfo{}
	if err := json.Unmarshal(data, info); err != nil {
		return nil, err
	}
	return info, nil
}

// NewDefaultClient creates an instance of defaultClient.
func NewDefaultClient(cfg *ClientConfig) (Client, error) {
	url := cfg.Endpoint
	if !strings.Contains(url, "://") {
		url = "https://" + url
	}
	url = strings.TrimSuffix(url, "/")
	tc := &tls.Config{
		InsecureSkipVerify: cfg.SkipTLSVerify,
	}
	if !cfg.SkipTLSVerify && len(cfg.CARootPath) > 0 {
		content, err := ioutil.ReadFile(cfg.CARootPath)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		//Do not throw error if the certificate is malformed, so we can put a place holder.
		if ok := pool.AppendCertsFromPEM(content); !ok {
			log.Warningf("Failed to append certificate to cert pool, cert path: %s", cfg.CARootPath)
		} else {
			tc.RootCAs = pool
		}
	}
	hc := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tc,
		},
	}

	oc := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: url + "/uaa/oauth/token",
			AuthURL:  url + "/uaa/oauth/authorize",
		},
	}

	return &defaultClient{
		httpClient: hc,
		oauth2Cfg:  oc,
		endpoint:   url,
	}, nil
}
