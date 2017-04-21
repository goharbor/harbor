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

package token

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
)

const (
	issuer = "harbor-token-issuer"
)

var privateKey string

func init() {
	privateKey = "/etc/ui/private_key.pem"
}

// GetResourceActions ...
func GetResourceActions(scopes []string) []*token.ResourceActions {
	log.Debugf("scopes: %+v", scopes)
	var res []*token.ResourceActions
	for _, s := range scopes {
		if s == "" {
			continue
		}
		items := strings.Split(s, ":")
		length := len(items)

		typee := items[0]

		name := ""
		if length > 1 {
			name = items[1]
		}

		actions := []string{}
		if length > 2 {
			actions = strings.Split(items[2], ",")
		}

		res = append(res, &token.ResourceActions{
			Type:    typee,
			Name:    name,
			Actions: actions,
		})
	}
	return res
}

//filterAccess iterate a list of resource actions and try to use the filter that matches the resource type to filter the actions.
func filterAccess(access []*token.ResourceActions, u userInfo, filters map[string]accessFilter) error {
	var err error
	for _, a := range access {
		f, ok := filters[a.Type]
		if !ok {
			a.Actions = []string{}
			log.Warningf("No filter found for access type: %s, skip filter, the access of resource '%s' will be set empty.", a.Type, a.Name)
			continue
		}
		err = f.filter(u, a)
		log.Debugf("user: %s, access: %v", u.name, a)
		if err != nil {
			return err
		}
	}
	return nil
}

//RegistryTokenForUI calls genTokenForUI to get raw token for registry
func RegistryTokenForUI(username string, service string, scopes []string) (string, int, *time.Time, error) {
	return genTokenForUI(username, service, scopes, registryFilterMap)
}

//NotaryTokenForUI calls genTokenForUI to get raw token for notary
func NotaryTokenForUI(username string, service string, scopes []string) (string, int, *time.Time, error) {
	return genTokenForUI(username, service, scopes, notaryFilterMap)
}

// genTokenForUI is for the UI process to call, so it won't establish a https connection from UI to proxy.
func genTokenForUI(username string, service string, scopes []string, filters map[string]accessFilter) (string, int, *time.Time, error) {
	isAdmin, err := dao.IsAdminRole(username)
	if err != nil {
		return "", 0, nil, err
	}
	u := userInfo{
		name:    username,
		allPerm: isAdmin,
	}
	access := GetResourceActions(scopes)
	err = filterAccess(access, u, filters)
	if err != nil {
		return "", 0, nil, err
	}
	return MakeRawToken(username, service, access)
}

// MakeRawToken makes a valid jwt token based on parms.
func MakeRawToken(username, service string, access []*token.ResourceActions) (token string, expiresIn int, issuedAt *time.Time, err error) {
	pk, err := libtrust.LoadKeyFile(privateKey)
	if err != nil {
		return "", 0, nil, err
	}
	expiration, err := config.TokenExpiration()
	if err != nil {
		return "", 0, nil, err
	}

	tk, expiresIn, issuedAt, err := makeTokenCore(issuer, username, service, expiration, access, pk)
	if err != nil {
		return "", 0, nil, err
	}
	rs := fmt.Sprintf("%s.%s", tk.Raw, base64UrlEncode(tk.Signature))
	return rs, expiresIn, issuedAt, nil
}

type tokenJSON struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	IssuedAt  string `json:"issued_at"`
}

func makeToken(username, service string, access []*token.ResourceActions) (*tokenJSON, error) {
	raw, expires, issued, err := MakeRawToken(username, service, access)
	if err != nil {
		return nil, err
	}
	return &tokenJSON{raw, expires, issued.Format(time.RFC3339)}, nil
}

func permToActions(p string) []string {
	res := []string{}
	if strings.Contains(p, "W") {
		res = append(res, "push")
	}
	if strings.Contains(p, "M") {
		res = append(res, "*")
	}
	if strings.Contains(p, "R") {
		res = append(res, "pull")
	}
	return res
}

//make token core
func makeTokenCore(issuer, subject, audience string, expiration int,
	access []*token.ResourceActions, signingKey libtrust.PrivateKey) (t *token.Token, expiresIn int, issuedAt *time.Time, err error) {

	joseHeader := &token.Header{
		Type:       "JWT",
		SigningAlg: "RS256",
		KeyID:      signingKey.KeyID(),
	}

	jwtID, err := randString(16)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("Error to generate jwt id: %s", err)
	}

	now := time.Now().UTC()
	issuedAt = &now
	expiresIn = expiration * 60

	claimSet := &token.ClaimSet{
		Issuer:     issuer,
		Subject:    subject,
		Audience:   audience,
		Expiration: now.Add(time.Duration(expiration) * time.Minute).Unix(),
		NotBefore:  now.Unix(),
		IssuedAt:   now.Unix(),
		JWTID:      jwtID,
		Access:     access,
	}

	var joseHeaderBytes, claimSetBytes []byte

	if joseHeaderBytes, err = json.Marshal(joseHeader); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to marshal jose header: %s", err)
	}
	if claimSetBytes, err = json.Marshal(claimSet); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to marshal claim set: %s", err)
	}

	encodedJoseHeader := base64UrlEncode(joseHeaderBytes)
	encodedClaimSet := base64UrlEncode(claimSetBytes)
	payload := fmt.Sprintf("%s.%s", encodedJoseHeader, encodedClaimSet)

	var signatureBytes []byte
	if signatureBytes, _, err = signingKey.Sign(strings.NewReader(payload), crypto.SHA256); err != nil {
		return nil, 0, nil, fmt.Errorf("unable to sign jwt payload: %s", err)
	}

	signature := base64UrlEncode(signatureBytes)
	tokenString := fmt.Sprintf("%s.%s", payload, signature)
	t, err = token.NewToken(tokenString)
	return
}

func randString(length int) (string, error) {
	const alphanum = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rb := make([]byte, length)
	_, err := rand.Read(rb)
	if err != nil {
		return "", err
	}
	for i, b := range rb {
		rb[i] = alphanum[int(b)%len(alphanum)]
	}
	return string(rb), nil
}

func base64UrlEncode(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}
