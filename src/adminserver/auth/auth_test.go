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
)

func TestAuthenticate(t *testing.T) {
	secret := "correct"
	req1, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	req2, err := http.NewRequest("", "", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req2.AddCookie(&http.Cookie{
		Name:  "secret",
		Value: secret,
	})

	cases := []struct {
		secrets map[string]string
		req     *http.Request
		result  bool
	}{
		{nil, req1, true},
		{map[string]string{"secret1": "incorrect"}, req2, false},
		{map[string]string{"secret1": "incorrect", "secret2": secret}, req2, true},
	}

	for _, c := range cases {
		authenticator := NewSecretAuthenticator(c.secrets)
		authenticated, err := authenticator.Authenticate(c.req)
		assert.Nil(t, err, "unexpected error")
		assert.Equal(t, c.result, authenticated, "unexpected result")
	}
}
