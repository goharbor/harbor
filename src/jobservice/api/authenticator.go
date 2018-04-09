// Copyright 2018 The Harbor Authors. All rights reserved.

package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/vmware/harbor/src/jobservice/config"
	"github.com/vmware/harbor/src/jobservice/utils"
)

const (
	secretPrefix = "Harbor-Secret"
	authHeader   = "Authorization"
)

//Authenticator defined behaviors of doing auth checking.
type Authenticator interface {
	//Auth incoming request
	//
	//req *http.Request: the incoming request
	//
	//Returns:
	// nil returned if successfully done
	// otherwise an error returned
	DoAuth(req *http.Request) error
}

//SecretAuthenticator implements interface 'Authenticator' based on simple secret.
type SecretAuthenticator struct{}

//DoAuth implements same method in interface 'Authenticator'.
func (sa *SecretAuthenticator) DoAuth(req *http.Request) error {
	if req == nil {
		return errors.New("nil request")
	}

	h := strings.TrimSpace(req.Header.Get(authHeader))
	if utils.IsEmptyStr(h) {
		return fmt.Errorf("header '%s' missing", authHeader)
	}

	if !strings.HasPrefix(h, secretPrefix) {
		return fmt.Errorf("'%s' should start with '%s' but got '%s' now", authHeader, secretPrefix, h)
	}

	secret := strings.TrimSpace(strings.TrimPrefix(h, secretPrefix))
	//incase both two are empty
	if utils.IsEmptyStr(secret) {
		return errors.New("empty secret is not allowed")
	}

	expectedSecret := config.GetUIAuthSecret()
	if expectedSecret != secret {
		return errors.New("unauthorized")
	}

	return nil
}
