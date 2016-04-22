/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package token

import (
	"crypto"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/docker/libtrust"
)

const (
	issuer     = "registry-token-issuer"
	privateKey = "/etc/ui/private_key.pem"
	expiration = 5 //minute
)

// GetResourceActions ...
func GetResourceActions(scopes []string) []*token.ResourceActions {
	var res []*token.ResourceActions
	for _, s := range scopes {
		if s == "" {
			continue
		}
		items := strings.Split(s, ":")
		res = append(res, &token.ResourceActions{
			Type:    items[0],
			Name:    items[1],
			Actions: strings.Split(items[2], ","),
		})
	}
	return res
}

// FilterAccess modify the action list in access based on permission
// determine if the request needs to be authenticated.
func FilterAccess(username string, authenticated bool, a *token.ResourceActions) {

	if a.Type == "registry" && a.Name == "catalog" {
		return
	}

	//clear action list to assign to new acess element after perm check.
	a.Actions = []string{}
	if a.Type == "repository" {
		if strings.Contains(a.Name, "/") { //Only check the permission when the requested image has a namespace, i.e. project
			projectName := a.Name[0:strings.LastIndex(a.Name, "/")]
			var permission string
			if authenticated {
				isAdmin, err := dao.IsAdminRole(username)
				if err != nil {
					log.Errorf("Error occurred in IsAdminRole: %v", err)
				}
				if isAdmin {
					exist, err := dao.ProjectExists(projectName)
					if err != nil {
						log.Errorf("Error occurred in CheckExistProject: %v", err)
						return
					}
					if exist {
						permission = "RWM"
					} else {
						permission = ""
						log.Infof("project %s does not exist, set empty permission for admin\n", projectName)
					}
				} else {
					permission, err = dao.GetPermission(username, projectName)
					if err != nil {
						log.Errorf("Error occurred in GetPermission: %v", err)
						return
					}
				}
			}
			if strings.Contains(permission, "W") {
				a.Actions = append(a.Actions, "push")
			}
			if strings.Contains(permission, "M") {
				a.Actions = append(a.Actions, "*")
			}
			if strings.Contains(permission, "R") || dao.IsProjectPublic(projectName) {
				a.Actions = append(a.Actions, "pull")
			}
		}
	}
	log.Infof("current access, type: %s, name:%s, actions:%v \n", a.Type, a.Name, a.Actions)
}

// GenTokenForUI is for the UI process to call, so it won't establish a https connection from UI to proxy.
func GenTokenForUI(username string, service string, scopes []string) (string, error) {
	access := GetResourceActions(scopes)
	for _, a := range access {
		FilterAccess(username, true, a)
	}
	return MakeToken(username, service, access)
}

// MakeToken makes a valid jwt token based on parms.
func MakeToken(username, service string, access []*token.ResourceActions) (string, error) {
	pk, err := libtrust.LoadKeyFile(privateKey)
	if err != nil {
		return "", err
	}
	tk, err := makeTokenCore(issuer, username, service, expiration, access, pk)
	if err != nil {
		return "", err
	}
	rs := fmt.Sprintf("%s.%s", tk.Raw, base64UrlEncode(tk.Signature))
	return rs, nil
}

//make token core
func makeTokenCore(issuer, subject, audience string, expiration int,
	access []*token.ResourceActions, signingKey libtrust.PrivateKey) (*token.Token, error) {

	joseHeader := &token.Header{
		Type:       "JWT",
		SigningAlg: "RS256",
		KeyID:      signingKey.KeyID(),
	}

	jwtID, err := randString(16)
	if err != nil {
		return nil, fmt.Errorf("Error to generate jwt id: %s", err)
	}

	now := time.Now()

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
		return nil, fmt.Errorf("unable to marshal jose header: %s", err)
	}
	if claimSetBytes, err = json.Marshal(claimSet); err != nil {
		return nil, fmt.Errorf("unable to marshal claim set: %s", err)
	}

	encodedJoseHeader := base64UrlEncode(joseHeaderBytes)
	encodedClaimSet := base64UrlEncode(claimSetBytes)
	payload := fmt.Sprintf("%s.%s", encodedJoseHeader, encodedClaimSet)

	var signatureBytes []byte
	if signatureBytes, _, err = signingKey.Sign(strings.NewReader(payload), crypto.SHA256); err != nil {
		return nil, fmt.Errorf("unable to sign jwt payload: %s", err)
	}

	signature := base64UrlEncode(signatureBytes)
	tokenString := fmt.Sprintf("%s.%s", payload, signature)
	return token.NewToken(tokenString)
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
