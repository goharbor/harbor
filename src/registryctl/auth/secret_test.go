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

package auth

import (
	"fmt"
	"net/http"
	"testing"

	commonsecret "github.com/goharbor/harbor/src/common/secret"
	"github.com/stretchr/testify/assert"
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

func TestNilRequest(t *testing.T) {
	secret := "Correct"
	req, err := http.NewRequest("", "", nil)
	req = nil
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	_ = commonsecret.AddToRequest(req, secret)

	authenticator := NewSecretHandler(map[string]string{"secret1": "correct"})
	err = authenticator.AuthorizeRequest(req)
	assert.Equal(t, err, ErrNoSecret)
}

func TestNoSecret(t *testing.T) {
	secret := ""
	req, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	_ = commonsecret.AddToRequest(req, secret)

	authenticator := NewSecretHandler(map[string]string{})
	err = authenticator.AuthorizeRequest(req)
	assert.Equal(t, err, ErrNoSecret)
}

func TestIncorrectHarborSecret(t *testing.T) {
	secret := "correct"
	req, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	_ = commonsecret.AddToRequest(req, secret)

	// Set req header to an incorrect value to trigger error return
	req.Header.Set("Authorization", fmt.Sprintf("%s%s", "WrongPrefix", secret))
	authenticator := NewSecretHandler(map[string]string{"secret1": "correct"})
	err = authenticator.AuthorizeRequest(req)
	assert.Equal(t, err, ErrInvalidCredential)
}
