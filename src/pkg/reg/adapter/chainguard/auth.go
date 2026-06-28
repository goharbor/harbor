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

package chainguard

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

const (
	defaultIssuer   = "https://issuer.enforce.dev"
	defaultAudience = "cgr.dev"
	basicUser       = "_token"
)

// dockerPasswordAuthorizer exchanges an OIDC token for a Chainguard registry token (STS), then
// sets Docker-style basic auth expected by cgr.dev (username "_token", password = exchanged token).
type dockerPasswordAuthorizer struct {
	mu sync.Mutex

	reg *model.Registry

	cachedPassword string
	cacheExpiry    time.Time
}

func newDockerPasswordAuthorizer(reg *model.Registry) *dockerPasswordAuthorizer {
	return &dockerPasswordAuthorizer{reg: reg}
}

func (a *dockerPasswordAuthorizer) Modify(req *http.Request) error {
	pass, err := a.registryPassword()
	if err != nil {
		return err
	}
	req.SetBasicAuth(basicUser, pass)
	return nil
}

func (a *dockerPasswordAuthorizer) registryPassword() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cachedPassword != "" && time.Now().Before(a.cacheExpiry.Add(-30*time.Second)) {
		return a.cachedPassword, nil
	}

	identity, idToken, err := loadCredentials(a.reg)
	if err != nil {
		return "", err
	}

	tok, exp, err := exchange(a.reg, defaultIssuer, defaultAudience, identity, idToken)
	if err != nil {
		return "", err
	}
	a.cachedPassword = tok
	a.cacheExpiry = exp
	return tok, nil
}

type stsExchangeRequest struct {
	Aud      []string `json:"aud"`
	Identity string   `json:"identity"`
}

type stsRawToken struct {
	Token        string          `json:"token"`
	RefreshToken string          `json:"refresh_token"`
	Expiry       json.RawMessage `json:"expiry"`
}

func exchange(reg *model.Registry, issuer, audience, identity, idToken string) (password string, exp time.Time, err error) {
	if identity == "" {
		return "", time.Time{}, errors.New("chainguard identity ID is required (access key)")
	}

	body, err := json.Marshal(stsExchangeRequest{
		Aud:      []string{audience},
		Identity: identity,
	})
	if err != nil {
		return "", time.Time{}, err
	}

	u := strings.TrimRight(issuer, "/") + "/sts/exchange"
	req, err := http.NewRequest(http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(idToken))

	client := &http.Client{
		Transport: commonhttp.GetHTTPTransport(
			commonhttp.WithInsecure(reg.Insecure),
			commonhttp.WithCACert(reg.CACertificate),
		),
		Timeout: config.RegistryHTTPClientTimeout(),
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("sts exchange %s: %s: %s", u, resp.Status, truncateForErr(respBody))
	}

	var raw stsRawToken
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return "", time.Time{}, err
	}
	if raw.Token == "" {
		return "", time.Time{}, errors.New("sts exchange returned empty token")
	}

	exp = tokenExpiry(raw.Token)
	if ts := parseProtoTimestampJSON(raw.Expiry); !ts.IsZero() {
		exp = ts
	}
	if exp.IsZero() {
		exp = time.Now().Add(50 * time.Minute)
	}
	return raw.Token, exp, nil
}

func truncateForErr(b []byte) string {
	s := string(b)
	if len(s) > 512 {
		return s[:512] + "..."
	}
	return s
}

func parseProtoTimestampJSON(raw json.RawMessage) time.Time {
	if len(raw) == 0 || string(raw) == "null" {
		return time.Time{}
	}
	var s string
	if json.Unmarshal(raw, &s) == nil && s != "" {
		if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
			return t
		}
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			return t
		}
	}
	var obj struct {
		Seconds int64 `json:"seconds"`
		Nanos   int32 `json:"nanos"`
	}
	if json.Unmarshal(raw, &obj) == nil && obj.Seconds != 0 {
		return time.Unix(obj.Seconds, int64(obj.Nanos))
	}
	return time.Time{}
}

func tokenExpiry(jwt string) time.Time {
	parts := strings.Split(jwt, ".")
	if len(parts) < 2 {
		return time.Time{}
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}
	}
	var claims struct {
		Exp int64 `json:"exp"`
	}
	if json.Unmarshal(payload, &claims) != nil || claims.Exp == 0 {
		return time.Time{}
	}
	return time.Unix(claims.Exp, 0)
}

func resolveSecretToToken(secret string) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errors.New("empty OIDC token or path")
	}
	if fi, err := os.Stat(secret); err == nil && !fi.IsDir() {
		b, err := os.ReadFile(secret)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	return secret, nil
}

// loadCredentials reads the Chainguard identity ID from the registry access key (UI) and the OIDC
// JWT from the access secret (inline string or path to a file containing the token).
func loadCredentials(reg *model.Registry) (identity, idToken string, err error) {
	if reg.Credential == nil {
		return "", "", errors.New("credential is required")
	}
	identity = strings.TrimSpace(reg.Credential.AccessKey)
	if identity == "" {
		return "", "", errors.New("chainguard identity ID is required (access key)")
	}
	raw, err := resolveSecretToToken(reg.Credential.AccessSecret)
	if err != nil {
		return "", "", err
	}
	idToken = strings.TrimSpace(raw)
	if idToken == "" {
		return "", "", errors.New("OIDC token is required (access secret: string or file path)")
	}
	return identity, idToken, nil
}
