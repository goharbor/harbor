package googlegar

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type registryAccessToken struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
	IssuedAt  string `json:"issued_at"`
}

var (
	username string = "_json_key"
	password string = "ppp"
	insecure bool   = false

	token *registryAccessToken = &registryAccessToken{
		Token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
	}
	expectedAuthHeader string = fmt.Sprintf("Bearer %s", token.Token)

	server *httptest.Server
)

func setup() {
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.Header().Add("WWW-Authenticate", "Bearer realm=\"http://"+r.Host+"/v2/token\"")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// return token
		if strings.HasPrefix(r.URL.Path, "/v2/token") {
			b, _ := json.Marshal(token)
			w.Write(b)
			return
		}

		fmt.Printf("not matched: %s\n", r.URL.Path)
		w.WriteHeader(http.StatusOK)
	}))
}

func teardown() {
	server.Close()
}

func TestAuthorizer_AddAuthorization(t *testing.T) {
	setup()
	defer teardown()

	// we are expecting a Authorization header to be added
	auth := NewAuthorizer(username, password, insecure)
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/v2/", nil)
	auth.Modify(req)
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		t.Errorf("expected authorization header")
	}
	if authHeader != expectedAuthHeader {
		t.Errorf("unexpected token - expected: %s, got: %s", expectedAuthHeader, authHeader)
	}

	// we are expecting a Authorization header to be added
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/artifacts-uploads", nil)
	auth.Modify(req)
	authHeader = req.Header.Get("Authorization")
	if authHeader == "" {
		t.Errorf("expected authorization header")
	}
	if authHeader != "Bearer "+token.Token {
		t.Errorf("unexpected token - expected: %s, got: %s", token.Token, authHeader)
	}
}

func TestAuthorizer_NoAuthorization(t *testing.T) {
	setup()
	defer teardown()

	// we are expecting a Authorization header to be added
	auth := NewAuthorizer(username, password, insecure)
	// authorizer should be initilized with an authorization header
	req, _ := http.NewRequest(http.MethodGet, server.URL+"/v2/", nil)
	auth.Modify(req)
	authHeader := req.Header.Get("Authorization")
	if authHeader == "" {
		t.Errorf("expected authorization header")
	}
	if authHeader != expectedAuthHeader {
		t.Errorf("unexpected token - expected: %s, got: %s", expectedAuthHeader, authHeader)
	}

	// making a non /v2 or /artifacts-uploads call should not add the 'Authorization' Header
	req, _ = http.NewRequest(http.MethodGet, server.URL+"/some-other-path", nil)
	auth.Modify(req)
	if req.Header.Get("Authorization") != "" {
		t.Errorf("not expected authorization header")
	}
}
