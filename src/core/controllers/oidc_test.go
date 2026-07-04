package controllers

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/beego/beego/v2/server/web"
	jose "github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	utilstest "github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/core/middlewares"
	cachepkg "github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/memory"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/oidc"
)

var (
	fakeProviderOnce sync.Once
	fakeProvider     *fakeOIDCProvider
)

func init() {
	web.Router(common.OIDCLoginPath, &OIDCController{}, "get:RedirectLogin")
	web.Router("/c/oidc/cli-token", &OIDCController{}, "get:CLIToken")
	web.Router(common.OIDCCallbackPath, &OIDCController{}, "get:Callback")
}

func TestGetSessionType(t *testing.T) {
	tests := []struct {
		name          string
		refreshToken  string
		expectedType  string
		expectedError bool
	}{
		{
			name:          "Valid",
			refreshToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0eXAiOiJvZmZsaW5lIn0.d9fcdba7c10fc1263bf682947afabaecf3496070cd2d5a5e7b3c79dbf1545c1f",
			expectedType:  "offline",
			expectedError: false,
		},
		{
			name:          "Invalid",
			refreshToken:  "invalidToken",
			expectedType:  "",
			expectedError: true,
		},
		{
			name:          "Missing 'typ' claim",
			refreshToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhbGciOiJIUzI1NiJ9.d9fcdba7c10fc1263bf682947afabaecf3496070cd2d5a5e7b3c79dbf1545c1f",
			expectedType:  "",
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ, err := getSessionType(tt.refreshToken)
			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedType, typ)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedType, typ)
			}
		})
	}
}

func TestOIDCCLIKey(t *testing.T) {
	assert.Equal(t, "oidc_cli_state:state-123", oidcCLIStateKey("state-123"))
	assert.Equal(t, "oidc_cli_poll:poll-123", oidcCLIPollKey("poll-123"))
}

func TestUnmarshalOIDCToken(t *testing.T) {
	raw, err := json.Marshal(&oidc.Token{
		RawIDToken: "raw-id-token",
	})
	assert.NoError(t, err)

	token, err := unmarshalOIDCToken(raw)
	assert.NoError(t, err)
	assert.Equal(t, "raw-id-token", token.RawIDToken)
}

func TestOIDCCLILoginPendingAndReadyFlow(t *testing.T) {
	cacheSetup(t)
	provider := sharedFakeOIDCProvider(t)
	configureOIDCTest(t, provider.server.URL, true)
	handler := newOIDCTestHandler()

	loginResp := startCLILogin(t, handler)
	assert.NotEmpty(t, loginResp.RedirectURL)
	assert.NotEmpty(t, loginResp.PollToken)
	assert.Contains(t, loginResp.RedirectURL, provider.server.URL+"/authorize")
	state := stateFromRedirectURL(t, loginResp.RedirectURL)

	pendingResp := cliTokenResponse(t, handler, loginResp.PollToken)
	assert.Equal(t, http.StatusAccepted, pendingResp.Code)
	assert.Equal(t, oidcCLIStatusPending, decodeCLITokenResponse(t, pendingResp).Status)

	callbackReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?code=test-code&state=%s", common.OIDCCallbackPath, url.QueryEscape(state)), nil)
	callbackRec := httptest.NewRecorder()
	handler.ServeHTTP(callbackRec, callbackReq)
	require.Equal(t, http.StatusFound, callbackRec.Code)

	readyResp := cliTokenResponse(t, handler, loginResp.PollToken)
	assert.Equal(t, http.StatusOK, readyResp.Code)
	ready := decodeCLITokenResponse(t, readyResp)
	assert.Equal(t, oidcCLIStatusReady, ready.Status)
	assert.Equal(t, provider.rawIDToken, ready.IDToken)
	assert.Equal(t, provider.refreshToken, ready.RefreshToken)
	assert.Equal(t, provider.username, ready.Username)
	assert.NotZero(t, ready.ExpiresAt)

	afterConsumeResp := cliTokenResponse(t, handler, loginResp.PollToken)
	assert.Equal(t, http.StatusGone, afterConsumeResp.Code)
	assert.Equal(t, oidcCLIStatusExpired, decodeCLITokenResponse(t, afterConsumeResp).Status)
}

func TestOIDCCLILoginFailureFlow(t *testing.T) {
	cacheSetup(t)
	provider := sharedFakeOIDCProvider(t)
	configureOIDCTest(t, provider.server.URL, true)
	handler := newOIDCTestHandler()

	loginResp := startCLILogin(t, handler)
	state := stateFromRedirectURL(t, loginResp.RedirectURL)

	callbackReq := httptest.NewRequest(
		http.MethodGet,
		fmt.Sprintf("%s?state=%s&error=access_denied&error_description=%s", common.OIDCCallbackPath, url.QueryEscape(state), url.QueryEscape("user denied access")),
		nil,
	)
	callbackRec := httptest.NewRecorder()
	handler.ServeHTTP(callbackRec, callbackReq)
	require.Equal(t, http.StatusBadRequest, callbackRec.Code)

	failedResp := cliTokenResponse(t, handler, loginResp.PollToken)
	assert.Equal(t, http.StatusBadRequest, failedResp.Code)
	failed := decodeCLITokenResponse(t, failedResp)
	assert.Equal(t, oidcCLIStatusFailed, failed.Status)
	assert.Contains(t, failed.Error, "access_denied")
}

func newOIDCTestHandler() http.Handler {
	handler := http.Handler(web.BeeApp.Handlers)
	mws := middlewares.MiddleWares()
	for i := len(mws) - 1; i >= 0; i-- {
		if mws[i] == nil {
			continue
		}
		handler = mws[i](handler)
	}
	return handler
}

func startCLILogin(t *testing.T, handler http.Handler) *oidcCLILoginResponse {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, common.OIDCLoginPath+"?mode=cli", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	resp := &oidcCLILoginResponse{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), resp))
	return resp
}

func cliTokenResponse(t *testing.T, handler http.Handler, pollToken string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/c/oidc/cli-token?poll_token="+url.QueryEscape(pollToken), nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func stateFromRedirectURL(t *testing.T, redirectURL string) string {
	t.Helper()

	u, err := url.Parse(redirectURL)
	require.NoError(t, err)
	state := u.Query().Get("state")
	require.NotEmpty(t, state)
	return state
}

func decodeCLITokenResponse(t *testing.T, rec *httptest.ResponseRecorder) *oidcCLITokenResponse {
	t.Helper()

	resp := &oidcCLITokenResponse{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), resp))
	return resp
}

func cacheSetup(t *testing.T) {
	t.Helper()
	require.NoError(t, cachepkg.Initialize(cachepkg.Memory, ""))
}

func configureOIDCTest(t *testing.T, endpoint string, autoOnboard bool) {
	t.Helper()

	cfg := map[string]any{}
	for key, value := range utilstest.GetDefaultConfigMap() {
		cfg[key] = value
	}
	cfg[common.AUTHMode] = common.OIDCAuth
	cfg[common.ExtEndpoint] = "http://harbor.test"
	cfg[common.OIDCName] = "test-oidc"
	cfg[common.OIDCEndpoint] = endpoint
	cfg[common.OIDCCLientID] = "harbor-cli-test"
	cfg[common.OIDCClientSecret] = "harbor-cli-secret"
	cfg[common.OIDCVerifyCert] = true
	cfg[common.OIDCAutoOnboard] = autoOnboard
	cfg[common.OIDCScope] = "openid,profile,email"
	cfg[common.OIDCGroupsClaim] = "groups"
	cfg[common.OIDCUserClaim] = "preferred_username"
	config.InitWithSettings(cfg)
}

type fakeOIDCProvider struct {
	server       *httptest.Server
	privateKey   *rsa.PrivateKey
	keyID        string
	issuer       string
	rawIDToken   string
	refreshToken string
	username     string
	subject      string
}

func newFakeOIDCProvider(t *testing.T) *fakeOIDCProvider {
	t.Helper()

	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	provider := &fakeOIDCProvider{
		privateKey:   privateKey,
		keyID:        "cli-test-key",
		refreshToken: "refresh-token-cli-test",
		username:     fmt.Sprintf("cli_oidc_%d", time.Now().UnixNano()),
		subject:      fmt.Sprintf("cli-sub-%d", time.Now().UnixNano()),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"issuer":                 provider.issuer,
				"authorization_endpoint": provider.issuer + "/authorize",
				"token_endpoint":         provider.issuer + "/token",
				"jwks_uri":               provider.issuer + "/jwks",
				"userinfo_endpoint":      provider.issuer + "/userinfo",
			})
		case "/jwks":
			_ = json.NewEncoder(w).Encode(jose.JSONWebKeySet{
				Keys: []jose.JSONWebKey{
					{
						Key:       &provider.privateKey.PublicKey,
						KeyID:     provider.keyID,
						Use:       "sig",
						Algorithm: string(jose.RS256),
					},
				},
			})
		case "/token":
			idToken := provider.issueIDToken(t)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"access_token":  "access-token-cli-test",
				"refresh_token": provider.refreshToken,
				"id_token":      idToken,
				"token_type":    "Bearer",
				"expires_in":    3600,
			})
		case "/userinfo":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"sub":                provider.subject,
				"iss":                provider.issuer,
				"name":               provider.username,
				"preferred_username": provider.username,
				"email":              provider.username + "@example.com",
				"groups":             []string{"developers"},
			})
		default:
			http.NotFound(w, r)
		}
	}))

	provider.server = server
	provider.issuer = server.URL
	provider.rawIDToken = provider.issueIDToken(t)

	return provider
}

func sharedFakeOIDCProvider(t *testing.T) *fakeOIDCProvider {
	t.Helper()

	fakeProviderOnce.Do(func() {
		fakeProvider = newFakeOIDCProvider(t)
	})

	require.NotNil(t, fakeProvider)
	return fakeProvider
}

func (p *fakeOIDCProvider) issueIDToken(t *testing.T) string {
	t.Helper()

	if p.rawIDToken != "" {
		return p.rawIDToken
	}

	claims := jwt.MapClaims{
		"iss":                p.issuer,
		"sub":                p.subject,
		"aud":                "harbor-cli-test",
		"exp":                time.Now().Add(30 * time.Minute).Unix(),
		"iat":                time.Now().Add(-1 * time.Minute).Unix(),
		"name":               p.username,
		"preferred_username": p.username,
		"email":              p.username + "@example.com",
		"groups":             []string{"developers"},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = p.keyID

	raw, err := token.SignedString(p.privateKey)
	require.NoError(t, err)

	p.rawIDToken = raw
	return raw
}
