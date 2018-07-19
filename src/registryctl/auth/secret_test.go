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

package auth

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	commonsecret "github.com/vmware/harbor/src/common/secret"
)

func TestAuthorizeRequestInvalid(t *testing.T) {
	secret := "correct"
	req, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	_ = commonsecret.AddToRequest(req, secret)

	authenticator := NewSecretHandler(map[string]string{"secret1": "incorrect"})
	err = authenticator.AuthorizeRequest(req)
	assert.Equal(t, err, ErrInvalidCredential)

}

func TestAuthorizeRequestValid(t *testing.T) {
	secret := "correct"
	req, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	_ = commonsecret.AddToRequest(req, secret)

	authenticator := NewSecretHandler(map[string]string{"secret1": "correct"})
	err = authenticator.AuthorizeRequest(req)
	assert.Nil(t, err)

}
