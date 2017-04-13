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
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/docker/distribution/registry/auth/token"
	"github.com/stretchr/testify/assert"

	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/vmware/harbor/src/common/utils/test"
	"github.com/vmware/harbor/src/ui/config"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"testing"
)

func TestMain(m *testing.M) {
	server, err := test.NewAdminserver(nil)
	if err != nil {
		panic(err)
	}
	defer server.Close()

	if err := os.Setenv("ADMIN_SERVER_URL", server.URL); err != nil {
		panic(err)
	}
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
	s := []string{"registry:catalog:*", "repository:10.117.4.142/notary-test/hello-world-2:pull,push"}
	expectedRA := [2]token.ResourceActions{
		token.ResourceActions{
			Type:    "registry",
			Name:    "catalog",
			Actions: []string{"*"},
		},
		token.ResourceActions{
			Type:    "repository",
			Name:    "10.117.4.142/notary-test/hello-world-2",
			Actions: []string{"pull", "push"},
		},
	}
	ra := GetResourceActions(s)
	assert.Equal(t, *ra[0], expectedRA[0], "The Resource Action mismatch")
	assert.Equal(t, *ra[1], expectedRA[1], "The Resource Action mismatch")
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
	//overwrite the config values for testing.
	privateKey = pk
	ra := []*token.ResourceActions{&token.ResourceActions{
		Type:    "repository",
		Name:    "10.117.4.142/notary-test/hello-world-2",
		Actions: []string{"pull", "push"},
	}}
	svc := "harbor-registry"
	u := "tester"
	tokenJSON, err := makeToken(u, svc, ra)
	if err != nil {
		t.Errorf("Error while making token: %v", err)
	}
	tokenString := tokenJSON.Token
	//t.Logf("privatekey: %s, crt: %s", tokenString, crt)
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
	testList := []parserTestRec{parserTestRec{"library/ubuntu:14.04", image{"library", "ubuntu", "14.04"}, false},
		parserTestRec{"test/hello", image{"test", "hello", ""}, false},
		parserTestRec{"myimage:14.04", image{}, true},
		parserTestRec{"org/team/img", image{"org", "team/img", ""}, false},
	}

	p := &basicParser{}
	for _, rec := range testList {
		r, err := p.parse(rec.input)
		if rec.expectError {
			assert.Error(t, err, "Expected error for input: %s", rec.input)
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
	testList := []parserTestRec{parserTestRec{"10.117.4.142:5000/library/ubuntu:14.04", image{"library", "ubuntu", "14.04"}, false},
		parserTestRec{"myimage:14.04", image{}, true},
		parserTestRec{"10.117.4.142:80/library/myimage:14.04", image{}, true},
		parserTestRec{"library/myimage:14.04", image{}, true},
		parserTestRec{"10.117.4.142:5000/myimage:14.04", image{}, true},
		parserTestRec{"10.117.4.142:5000/org/team/img", image{"org", "team/img", ""}, false},
	}
	for _, rec := range testList {
		r, err := p.parse(rec.input)
		if rec.expectError {
			assert.Error(t, err, "Expected error for input: %s", rec.input)
		} else {
			assert.Nil(t, err, "Expected no error for input: %s", rec.input)
			assert.Equal(t, rec.expect, *r, "result mismatch for input: %s", rec.input)
		}
	}
}

func TestFilterAccess(t *testing.T) {
	//TODO put initial data in DB to verify repository filter.
	var err error
	s := []string{"registry:catalog:*"}
	a1 := GetResourceActions(s)
	a2 := GetResourceActions(s)
	a3 := GetResourceActions(s)
	u1 := userInfo{"jack", true}
	u2 := userInfo{"jack", false}
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
	err = filterAccess(a1, u1, registryFilterMap)
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, ra1, *a1[0], "Mismatch after registry filter Map")
	err = filterAccess(a2, u1, notaryFilterMap)
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, ra2, *a2[0], "Mismatch after notary filter Map")
	err = filterAccess(a3, u2, registryFilterMap)
	assert.Nil(t, err, "Unexpected error: %v", err)
	assert.Equal(t, ra2, *a3[0], "Mismatch after registry filter Map")
}
