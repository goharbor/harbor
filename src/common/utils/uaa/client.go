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

// DefaultClient leverages oauth2 pacakge for oauth features
type defaultClient struct {
	httpClient *http.Client
	oauth2Cfg  *oauth2.Config
	//TODO: add public key, etc...
}

func (dc *defaultClient) PasswordAuth(username, password string) (*oauth2.Token, error) {
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, dc.httpClient)
	return dc.oauth2Cfg.PasswordCredentialsToken(ctx, username, password)
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
	}, nil
}
