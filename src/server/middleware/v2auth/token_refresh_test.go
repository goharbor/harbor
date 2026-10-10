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

package v2auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/beego/beego/v2/server/web"
	beegosession "github.com/beego/beego/v2/server/web/session"
	registrytoken "github.com/docker/distribution/registry/auth/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/project"
	svctoken "github.com/goharbor/harbor/src/core/service/token"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	tokenpkg "github.com/goharbor/harbor/src/pkg/token"
	v2 "github.com/goharbor/harbor/src/pkg/token/claims/v2"
	"github.com/goharbor/harbor/src/server/middleware/artifactinfo"
	securitymiddleware "github.com/goharbor/harbor/src/server/middleware/security"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
)

type refreshHTTPEvent struct {
	Method     string   `json:"method"`
	Path       string   `json:"path"`
	AuthScheme string   `json:"auth_scheme,omitempty"`
	Status     int      `json:"status"`
	Challenge  string   `json:"challenge,omitempty"`
	Scopes     []string `json:"requested_scopes,omitempty"`
}

type refreshMeasurement struct {
	Scenario      string             `json:"scenario"`
	Iteration     int                `json:"iteration"`
	FinalStatus   int                `json:"final_status"`
	Expected      int                `json:"expected_status"`
	Authorized    int                `json:"authorized_handler_calls"`
	TokenRequests int                `json:"token_requests"`
	DurationMS    float64            `json:"duration_ms"`
	Events        []refreshHTTPEvent `json:"events"`
}

type refreshRecorder struct {
	http.ResponseWriter
	status int
}

func (r *refreshRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// TestTokenScopeRefresh exercises a real go-containerregistry client over TLS
// against Harbor's artifact parsing, JWT verification and v2 authorization.
// Project lookup, token-service permissions and the downstream registry handler
// are fixtures; this does not require a deployed Harbor or registry storage.
// Set HARBOR_18856_SAMPLES and HARBOR_18856_RESULTS to collect repeatable traces.
func TestTokenScopeRefresh(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	keyPath := filepath.Join(t.TempDir(), "private_key.pem")
	require.NoError(t, os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key),
	}), 0600))
	t.Setenv("TOKEN_PRIVATE_KEY_PATH", keyPath)
	originalSessions := web.GlobalSessions
	web.GlobalSessions, err = beegosession.NewManager("memory", &beegosession.ManagerConfig{
		CookieName: "issue18856", Gclifetime: 3600, Maxlifetime: 3600,
	})
	require.NoError(t, err)
	defer func() { web.GlobalSessions = originalSessions }()

	ctl := &projecttesting.Controller{}
	lookup := func(_ context.Context, idOrName any, _ ...project.Option) *proModels.Project {
		id := int64(1)
		switch v := idOrName.(type) {
		case int64:
			id = v
		case string:
			if v == "project_2" {
				id = 2
			}
		}
		return &proModels.Project{ProjectID: id, Name: fmt.Sprintf("project_%d", id), CreationTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)}
	}
	mock.OnAnything(ctl, "Get").Return(lookup, nil)
	mock.OnAnything(ctl, "GetByName").Return(func(ctx context.Context, n string, _ ...project.Option) *proModels.Project {
		return lookup(ctx, n)
	}, nil)
	originalCtl, originalChecker := project.Ctl, checker
	project.Ctl, checker = ctl, reqChecker{ctl: ctl}
	defer func() { project.Ctl, checker = originalCtl, originalChecker }()

	var mu sync.Mutex
	var events []refreshHTTPEvent
	var authorized int
	downstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		mu.Lock()
		authorized++
		mu.Unlock()
		switch r.Method {
		case http.MethodPut:
			w.WriteHeader(http.StatusCreated)
		case http.MethodPost:
			if r.URL.Query().Get("mount") != "" {
				w.WriteHeader(http.StatusCreated)
			} else {
				w.WriteHeader(http.StatusAccepted)
			}
		case http.MethodDelete:
			w.WriteHeader(http.StatusAccepted)
		default:
			w.WriteHeader(http.StatusOK)
		}
	})
	registryHandler := artifactinfo.Middleware()(securitymiddleware.Middleware()(
		securitymiddleware.UnauthorizedMiddleware()(Middleware()(downstream))))
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &refreshRecorder{ResponseWriter: w, status: http.StatusOK}
		defer func() {
			scheme, _, _ := strings.Cut(r.Header.Get(authHeader), " ")
			e := refreshHTTPEvent{Method: r.Method, Path: r.URL.Path, AuthScheme: scheme,
				Status: rec.status, Challenge: rec.Header().Get("WWW-Authenticate")}
			if r.URL.Path == "/service/token" {
				e.Scopes = r.URL.Query()["scope"]
			}
			mu.Lock()
			events = append(events, e)
			mu.Unlock()
		}()
		if r.URL.Path != "/service/token" {
			registryHandler.ServeHTTP(rec, r)
			return
		}
		username, password, ok := r.BasicAuth()
		if !ok || password != "fixture-password" || r.URL.Query().Get("service") != svctoken.Registry {
			rec.WriteHeader(http.StatusUnauthorized)
			return
		}
		var scopes []string
		for _, value := range r.URL.Query()["scope"] {
			scopes = append(scopes, strings.Fields(value)...)
		}
		// The client retains old scopes when requesting additional actions.
		// Issue one access entry per repository, with the union of requested
		// actions, so a cached pull scope cannot overwrite the new delete scope.
		var access []*registrytoken.ResourceActions
		resources := make(map[string]*registrytoken.ResourceActions)
		for _, a := range svctoken.GetResourceActions(scopes) {
			resource := a.Type + ":" + a.Name
			if existing, ok := resources[resource]; ok {
				existing.Actions = append(existing.Actions, a.Actions...)
			} else {
				resources[resource] = a
				access = append(access, a)
			}
		}
		if username == "denied" {
			for _, a := range access {
				a.Actions = nil
			}
		}
		now := time.Now()
		claims := &v2.Claims{RegisteredClaims: jwt.RegisteredClaims{
			Issuer: v2.Issuer, Subject: username, Audience: jwt.ClaimStrings{svctoken.Registry},
			IssuedAt: jwt.NewNumericDate(now), NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute)),
		}, Access: access}
		tok, tokenErr := tokenpkg.New(&tokenpkg.Options{SignMethod: jwt.SigningMethodRS256,
			PrivateKey: pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})}, claims)
		if tokenErr != nil {
			rec.WriteHeader(http.StatusInternalServerError)
			return
		}
		raw, tokenErr := tok.Raw()
		if tokenErr != nil {
			rec.WriteHeader(http.StatusInternalServerError)
			return
		}
		rec.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(rec).Encode(map[string]any{"token": raw, "expires_in": 60}); err != nil {
			t.Errorf("encode token: %v", err)
		}
	}))
	defer server.Close()
	reg, err := name.NewRegistry(strings.TrimPrefix(server.URL, "https://"))
	require.NoError(t, err)

	samples := 1
	if value := os.Getenv("HARBOR_18856_SAMPLES"); value != "" {
		samples, err = strconv.Atoi(value)
		require.NoError(t, err)
		require.Positive(t, samples)
	}
	var measurements []refreshMeasurement
	defer func() {
		if resultPath := os.Getenv("HARBOR_18856_RESULTS"); resultPath != "" {
			data, err := json.MarshalIndent(measurements, "", "  ")
			require.NoError(t, err)
			require.NoError(t, os.WriteFile(resultPath, append(data, '\n'), 0600))
		}
	}()
	for _, tc := range []struct {
		name          string
		method        string
		path          string
		initialScopes []string
		wantScope     string
		status        int
		denied        bool
	}{
		{"pull", http.MethodGet, "/v2/project_1/image/manifests/latest", nil, "repository:project_1/image:pull", http.StatusOK, false},
		{"head", http.MethodHead, "/v2/project_1/image/manifests/latest", nil, "repository:project_1/image:pull", http.StatusOK, false},
		{"push", http.MethodPut, "/v2/project_1/image/manifests/latest", nil, "repository:project_1/image:pull,push", http.StatusCreated, false},
		{"upload", http.MethodPost, "/v2/project_1/image/blobs/uploads/", nil, "repository:project_1/image:pull,push", http.StatusAccepted, false},
		{"delete_from_pull_token", http.MethodDelete, "/v2/project_1/image/manifests/latest", []string{"repository:project_1/image:pull"}, "repository:project_1/image:delete", http.StatusAccepted, false},
		{"cross_project_mount", http.MethodPost, "/v2/project_1/image/blobs/uploads/?mount=sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa&from=project_2/image", []string{"repository:project_1/image:pull,push"}, "repository:project_1/image:pull,push repository:project_2/image:pull", http.StatusCreated, false},
		{"sufficient_scope", http.MethodGet, "/v2/project_1/image/manifests/latest", []string{"repository:project_1/image:pull"}, "", http.StatusOK, false},
		{"denied_scope", http.MethodGet, "/v2/project_1/image/manifests/latest", nil, "repository:project_1/image:pull", http.StatusUnauthorized, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			successes := 0
			scopedChallenges := 0
			for i := 0; i < samples; i++ {
				mu.Lock()
				events, authorized = nil, 0
				mu.Unlock()
				username := "allowed"
				if tc.denied {
					username = "denied"
				}
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				started := time.Now()
				rt, err := transport.NewWithContext(ctx, reg,
					authn.FromConfig(authn.AuthConfig{Username: username, Password: "fixture-password"}),
					server.Client().Transport, tc.initialScopes)
				require.NoError(t, err)
				req, err := http.NewRequestWithContext(ctx, tc.method, server.URL+tc.path, nil)
				require.NoError(t, err)
				resp, err := (&http.Client{Transport: rt}).Do(req)
				require.NoError(t, err)
				_, err = io.Copy(io.Discard, resp.Body)
				require.NoError(t, err)
				require.NoError(t, resp.Body.Close())
				cancel()
				mu.Lock()
				m := refreshMeasurement{Scenario: tc.name, Iteration: i + 1, FinalStatus: resp.StatusCode,
					Expected: tc.status, Authorized: authorized, DurationMS: float64(time.Since(started).Microseconds()) / 1000,
					Events: append([]refreshHTTPEvent(nil), events...)}
				mu.Unlock()
				var challengedScope string
				for _, event := range m.Events {
					if event.Path == "/service/token" {
						m.TokenRequests++
					}
					if event.Path != "/v2/" && event.Status == http.StatusUnauthorized {
						challengedScope = event.Challenge
					}
				}
				measurements = append(measurements, m)
				if strings.Contains(challengedScope, `scope="`+tc.wantScope+`"`) {
					scopedChallenges++
				}
				if m.FinalStatus == tc.status {
					successes++
				}
				if tc.denied {
					require.Zero(t, m.Authorized, "denied token must never reach the registry handler")
				} else if m.FinalStatus == tc.status {
					require.Equal(t, 1, m.Authorized)
				}
				if m.FinalStatus == tc.status && tc.wantScope != "" {
					require.Equal(t, 2, m.TokenRequests, "one initial token and one refresh")
				} else if tc.wantScope == "" {
					require.Equal(t, 1, m.TokenRequests, "sufficient token needs no refresh")
				}
			}
			t.Logf("%s: %d/%d expected HTTP %d", tc.name, successes, samples, tc.status)
			require.Equal(t, samples, successes)
			if tc.wantScope != "" {
				require.Equal(t, samples, scopedChallenges, "each insufficient token must receive the required scopes")
			}
		})
	}
}
