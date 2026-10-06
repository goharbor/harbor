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

package scan

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/pkg/robot/model"
)

// makeBearerAuthorization must verify the TLS certificate of
// the internal token service (it carries the scan robot's Basic credentials) and
// must bound the response body it reads.

// TestMakeBearerAuthorizationVerifiesTLS asserts the token request rejects an
// untrusted certificate. httptest.NewTLSServer serves an ad-hoc self-signed cert
// that is not in any trust store, so a verifying client must fail the handshake.
// On main (transport built with WithInsecure(true)) verification is skipped, the
// request succeeds, and this assertion fails — proving the fix.
func TestMakeBearerAuthorizationVerifiesTLS(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"token": "forged"})
	}))
	defer srv.Close()

	robotAccount := &model.Robot{Name: "robot$scan", Secret: "s3cret"}

	_, err := makeBearerAuthorization(robotAccount, srv.URL, "library/x")
	require.Error(t, err, "token request must reject an untrusted TLS certificate")
	var certErr *tls.CertificateVerificationError
	require.ErrorAs(t, err, &certErr)
}

// TestMakeBearerAuthorizationBoundsBody asserts the response body read is bounded.
// The server returns a body larger than maxBearerTokenResponseSize over plain HTTP
// (TLS is irrelevant to the bound). On the branch the read is capped and returns an
// "exceeds" error; on main the whole body is buffered and json.Unmarshal fails with
// a different ("invalid character") error, so this assertion fails — proving the fix.
func TestMakeBearerAuthorizationBoundsBody(t *testing.T) {
	huge := bytes.Repeat([]byte("A"), maxBearerTokenResponseSize+1024)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(huge)
	}))
	defer srv.Close()

	robotAccount := &model.Robot{Name: "robot$scan", Secret: "s3cret"}

	_, err := makeBearerAuthorization(robotAccount, srv.URL, "library/x")
	require.Error(t, err, "oversized token response must be rejected")
	require.Contains(t, err.Error(), "exceeds")
}
