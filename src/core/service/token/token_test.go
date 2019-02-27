// Copyright 2018 Project Harbor Authors
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
	"github.com/dgrijalva/jwt-go"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/stretchr/testify/assert"

	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"runtime"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/config"
)

func TestMain(m *testing.M) {
	if err := config.Init(); err != nil {
		panic(err)
	}
	InitCreators()
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestGetResourceActions(t *testing.T) {
	cases := map[string]*token.ResourceActions{
		"::": {
			Type:    "",
			Name:    "",
			Actions: []string{},
		},
		"repository": {
			Type:    "repository",
			Name:    "",
			Actions: []string{},
		},
		"repository:": {
			Type:    "repository",
			Name:    "",
			Actions: []string{},
		},
		"repository:library/hello-world": {
			Type:    "repository",
			Name:    "library/hello-world",
			Actions: []string{},
		},
		"repository:library/hello-world:": {
			Type:    "repository",
			Name:    "library/hello-world",
			Actions: []string{},
		},
		"repository:library/hello-world:pull,push": {
			Type:    "repository",
			Name:    "library/hello-world",
			Actions: []string{"pull", "push"},
		},
		"registry:catalog:*": {
			Type:    "registry",
			Name:    "catalog",
			Actions: []string{"*"},
		},
		"repository:192.168.0.1:443/library/hello-world:pull,push": {
			Type:    "repository",
			Name:    "192.168.0.1:443/library/hello-world",
			Actions: []string{"pull", "push"},
		},
	}

	for k, v := range cases {
		r := GetResourceActions([]string{k})[0]
		assert.EqualValues(t, v, r)
	}
}

func getKeyAndCertPath() (string, string) {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current directory")
	}
	return path.Join(path.Dir(f), "test/private_key.pem"), path.Join(path.Dir(f), "test/root.crt")
}

func getPublicKey(crtPath string) (*rsa.PublicKey, error) {
	crt, err := ioutil.ReadFile(crtPath)
	if err != nil {
		return nil, err
	}
	block, _ := pem.Decode(crt)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	return cert.PublicKey.(*rsa.PublicKey), nil
}

type harborClaims struct {
	jwt.StandardClaims
	// Private claims
	Access []*token.ResourceActions `json:"access"`
}

func TestMakeToken(t *testing.T) {
	pk, crt := getKeyAndCertPath()
	// overwrite the config values for testing.
	privateKey = pk
	ra := []*token.ResourceActions{{
		Type:    "repository",
		Name:    "10.117.4.142/notary-test/hello-world-2",
		Actions: []string{"pull", "push"},
	}}
	svc := "harbor-registry"
	u := "tester"
	tokenJSON, err := MakeToken(u, svc, ra)
	if err != nil {
		t.Errorf("Error while making token: %v", err)
	}
	tokenString := tokenJSON.Token
	// t.Logf("privatekey: %s, crt: %s", tokenString, crt)
	pubKey, err := getPublicKey(crt)
	if err != nil {
		t.Errorf("Error while getting public key from cert: %s", crt)
	}
	tok, err := jwt.ParseWithClaims(tokenString, &harborClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return pubKey, nil
	})
	t.Logf("Token validity: %v", tok.Valid)
	if err != nil {
		t.Errorf("Error while parsing the token: %v", err)
	}
	claims := tok.Claims.(*harborClaims)
	assert.Equal(t, *(claims.Access[0]), *(ra[0]), "Access mismatch")
	assert.Equal(t, claims.Audience, svc, "Audience mismatch")
}

func TestPermToActions(t *testing.T) {
	perm1 := "RWM"
	perm2 := "MRR"
	perm3 := ""
	expect1 := []string{"push", "*", "pull"}
	expect2 := []string{"*", "pull"}
	expect3 := []string{}
	res1 := permToActions(perm1)
	res2 := permToActions(perm2)
	res3 := permToActions(perm3)
	assert.Equal(t, res1, expect1, fmt.Sprintf("actions mismatch for permission: %s", perm1))
	assert.Equal(t, res2, expect2, fmt.Sprintf("actions mismatch for permission: %s", perm2))
	assert.Equal(t, res3, expect3, fmt.Sprintf("actions mismatch for permission: %s", perm3))
}

type parserTestRec struct {
	input       string
	expect      image
	expectError bool
}

func TestInit(t *testing.T) {
	InitCreators()
}

func TestBasicParser(t *testing.T) {
	testList := []parserTestRec{{"library/ubuntu:14.04", image{"library", "ubuntu", "14.04"}, false},
		{"test/hello", image{"test", "hello", ""}, false},
		{"myimage:14.04", image{}, true},
		{"org/team/img", image{"org", "team/img", ""}, false},
	}

	p := &basicParser{}
	for _, rec := range testList {
		r, err := p.parse(rec.input)
		if rec.expectError {
			assert.Error(t, err, fmt.Sprintf("Expected error for input: %s", rec.input))
		} else {
			assert.Nil(t, err, "Expected no error for input: %s", rec.input)
			assert.Equal(t, rec.expect, *r, "result mismatch for input: %s", rec.input)
		}
	}
}

func TestEndpointParser(t *testing.T) {
	p := &endpointParser{
		"10.117.4.142:5000",
	}
	testList := []parserTestRec{{"10.117.4.142:5000/library/ubuntu:14.04", image{"library", "ubuntu", "14.04"}, false},
		{"myimage:14.04", image{}, true},
		{"10.117.4.142:80/library/myimage:14.04", image{}, true},
		{"library/myimage:14.04", image{}, true},
		{"10.117.4.142:5000/myimage:14.04", image{}, true},
		{"10.117.4.142:5000/org/team/img", image{"org", "team/img", ""}, false},
	}
	for _, rec := range testList {
		r, err := p.parse(rec.input)
		if rec.expectError {
			assert.Error(t, err, fmt.Sprintf("Expected error for input: %s", rec.input))
		} else {
			assert.Nil(t, err, "Expected no error for input: %s", rec.input)
			assert.Equal(t, rec.expect, *r, "result mismatch for input: %s", rec.input)
		}
	}
}

type fakeSecurityContext struct {
	isAdmin bool
}

func (f *fakeSecurityContext) IsAuthenticated() bool {
	return true
}

func (f *fakeSecurityContext) GetUsername() string {
	return "jack"
}

func (f *fakeSecurityContext) IsSysAdmin() bool {
	return f.isAdmin
}
func (f *fakeSecurityContext) IsSolutionUser() bool {
	return false
}
func (f *fakeSecurityContext) Can(action rbac.Action, resource rbac.Resource) bool {
	return false
}
func (f *fakeSecurityContext) GetMyProjects() ([]*models.Project, error) {
	return nil, nil
}
func (f *fakeSecurityContext) GetProjectRoles(interface{}) []int {
	return nil
}

func TestFilterAccess(t *testing.T) {
	// TODO put initial data in DB to verify repository filter.
	var err error
	s := []string{"registry:catalog:*"}
	a1 := GetResourceActions(s)
	a2 := GetResourceActions(s)
	a3 := GetResourceActions(s)

	ra1 := token.ResourceActions{
		Type:    "registry",
		Name:    "catalog",
		Actions: []string{"*"},
	}
	ra2 := token.ResourceActions{
		Type:    "registry",
		Name:    "catalog",
		Actions: []string{},
	}
	err = filterAccess(a1, &fakeSecurityContext{
		isAdmin: true,
	}, nil, registryFilterMap)
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, ra1, *a1[0], "Mismatch after registry filter Map")

	err = filterAccess(a2, &fakeSecurityContext{
		isAdmin: true,
	}, nil, notaryFilterMap)
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, ra2, *a2[0], "Mismatch after notary filter Map")

	err = filterAccess(a3, &fakeSecurityContext{
		isAdmin: false,
	}, nil, registryFilterMap)
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, ra2, *a3[0], "Mismatch after registry filter Map")
}

func TestParseScopes(t *testing.T) {
	assert := assert.New(t)
	u1 := "/service/token?account=admin&scope=repository%3Alibrary%2Fregistry%3Apush%2Cpull&scope=repository%3Ahello-world%2Fregistry%3Apull&service=harbor-registry"
	r1, _ := url.Parse(u1)
	l1 := parseScopes(r1)
	assert.Equal([]string{"repository:library/registry:push,pull", "repository:hello-world/registry:pull"}, l1)
}
