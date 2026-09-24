package dockerhub

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

type clientTestTransport func(*http.Request) (*http.Response, error)

func (transport clientTestTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	return transport(request)
}

func clientTestResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func blockClientNetwork(testingContext *testing.T) func() {
	testingContext.Helper()
	transport, ok := commonhttp.GetHTTPTransport().(*http.Transport)
	require.True(testingContext, ok)
	originalProxy := transport.Proxy
	transport.Proxy = func(request *http.Request) (*url.URL, error) {
		return nil, fmt.Errorf("unexpected network request: %s", request.URL)
	}
	return func() { transport.Proxy = originalProxy }
}

func newTestClient(testingContext *testing.T, credential *model.Credential, transport http.RoundTripper) *Client {
	testingContext.Helper()
	restoreNetwork := blockClientNetwork(testingContext)
	defer restoreNetwork()
	client, err := NewClient(&model.Registry{URL: baseURL, Credential: credential})
	require.NoError(testingContext, err)
	require.Equal(testingContext, config.RegistryHTTPClientTimeout(), client.client.Timeout)
	require.Positive(testingContext, client.client.Timeout)
	client.client.Transport = transport
	return client
}

func TestClientLazyAuthenticationAndRefresh(testingContext *testing.T) {
	credential := &model.Credential{AccessKey: "test-user", AccessSecret: "test-secret"}
	logins := 0
	transport := clientTestTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == loginPath {
			require.Equal(testingContext, http.MethodPost, request.Method)
			require.Equal(testingContext, "application/json", request.Header.Get("Content-Type"))
			var supplied LoginCredential
			require.NoError(testingContext, json.NewDecoder(request.Body).Decode(&supplied))
			require.Equal(testingContext, LoginCredential{Identifier: credential.AccessKey, Secret: credential.AccessSecret}, supplied)
			logins++
			return clientTestResponse(http.StatusOK, fmt.Sprintf(`{"access_token":"token-%d"}`, logins)), nil
		}
		return clientTestResponse(http.StatusOK, request.Header.Get("Authorization")), nil
	})
	firstClient := newTestClient(testingContext, credential, transport)
	secondClient := newTestClient(testingContext, credential, transport)
	require.Zero(testingContext, logins)

	checkAuthorization := func(client *Client, expected string) {
		testingContext.Helper()
		response, err := client.Do(http.MethodGet, listNamespacePath, nil)
		require.NoError(testingContext, err)
		defer response.Body.Close()
		body, err := io.ReadAll(response.Body)
		require.NoError(testingContext, err)
		require.Equal(testingContext, expected, string(body))
	}
	checkAuthorization(firstClient, "Bearer token-1")
	checkAuthorization(firstClient, "Bearer token-1")
	require.Equal(testingContext, 1, logins)
	checkAuthorization(secondClient, "Bearer token-2")
	checkAuthorization(firstClient, "Bearer token-1")
	require.Equal(testingContext, 2, logins)
	firstClient.mu.Lock()
	firstClient.tokenExpiry = time.Now().Add(-time.Second)
	firstClient.mu.Unlock()
	checkAuthorization(firstClient, "Bearer token-3")
	require.Equal(testingContext, 3, logins)
}

func TestClientConcurrentRefresh(testingContext *testing.T) {
	var logins atomic.Int32
	client := newTestClient(testingContext, &model.Credential{AccessKey: "test-user", AccessSecret: "test-secret"}, nil)
	for round := 1; round <= 2; round++ {
		loginStarted := make(chan struct{}, 32)
		finishLogin := make(chan struct{})
		client.client.Transport = clientTestTransport(func(request *http.Request) (*http.Response, error) {
			if request.URL.Path == loginPath {
				sequence := logins.Add(1)
				loginStarted <- struct{}{}
				<-finishLogin
				return clientTestResponse(http.StatusOK, fmt.Sprintf(`{"access_token":"token-%d"}`, sequence)), nil
			}
			if request.Header.Get("Authorization") != fmt.Sprintf("Bearer token-%d", round) {
				return nil, fmt.Errorf("request used the wrong token")
			}
			return clientTestResponse(http.StatusOK, "{}"), nil
		})
		client.mu.Lock()
		client.tokenExpiry = time.Now().Add(-time.Second)
		client.mu.Unlock()
		var workers sync.WaitGroup
		start := make(chan struct{})
		results := make(chan error, 32)
		for index := 0; index < cap(results); index++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				<-start
				response, err := client.Do(http.MethodGet, listNamespacePath, nil)
				if response != nil {
					response.Body.Close()
				}
				results <- err
			}()
		}
		close(start)
		select {
		case <-loginStarted:
		case <-time.After(5 * time.Second):
			close(finishLogin)
			testingContext.Fatal("authentication request did not start")
		}
		lockAvailable := client.mu.TryLock()
		if lockAvailable {
			client.mu.Unlock()
		}
		close(finishLogin)
		workers.Wait()
		require.False(testingContext, lockAvailable, "the in-flight refresh must hold the client lock")
		close(results)
		for err := range results {
			require.NoError(testingContext, err)
		}
		require.EqualValues(testingContext, round, logins.Load())
	}
}

func TestClientAnonymousAccess(testingContext *testing.T) {
	for _, credential := range []*model.Credential{nil, {}} {
		requests := 0
		client := newTestClient(testingContext, credential, clientTestTransport(func(request *http.Request) (*http.Response, error) {
			requests++
			require.Equal(testingContext, listNamespacePath, request.URL.Path)
			require.Empty(testingContext, request.Header.Get("Authorization"))
			return clientTestResponse(http.StatusOK, "{}"), nil
		}))
		response, err := client.Do(http.MethodGet, listNamespacePath, nil)
		require.NoError(testingContext, err)
		response.Body.Close()
		require.Equal(testingContext, 1, requests)
	}
}

func TestClientAuthenticationFailures(testingContext *testing.T) {
	for _, testCase := range []struct {
		name   string
		status int
		body   string
	}{
		{name: "rate limited", status: http.StatusTooManyRequests, body: `{"detail":"Rate limit exceeded"}`},
		{name: "unauthorized", status: http.StatusUnauthorized, body: `{"detail":"Invalid credentials"}`},
		{name: "malformed JSON", status: http.StatusOK, body: "{"},
		{name: "missing token", status: http.StatusOK, body: "{}"},
		{name: "empty token", status: http.StatusOK, body: `{"access_token":""}`},
	} {
		testingContext.Run(testCase.name, func(testingContext *testing.T) {
			logins, apiRequests := 0, 0
			client := newTestClient(testingContext, &model.Credential{AccessKey: "test-user", AccessSecret: "test-secret"}, clientTestTransport(func(request *http.Request) (*http.Response, error) {
				if request.URL.Path == loginPath {
					logins++
					if logins == 1 {
						return clientTestResponse(testCase.status, testCase.body), nil
					}
					return clientTestResponse(http.StatusOK, `{"access_token":"recovered"}`), nil
				}
				apiRequests++
				require.Equal(testingContext, "Bearer recovered", request.Header.Get("Authorization"))
				return clientTestResponse(http.StatusOK, "{}"), nil
			}))
			response, err := client.Do(http.MethodGet, listNamespacePath, nil)
			require.Error(testingContext, err)
			require.Nil(testingContext, response)
			require.Equal(testingContext, 1, logins)
			require.Zero(testingContext, apiRequests)
			require.Empty(testingContext, client.token)
			require.True(testingContext, client.tokenExpiry.IsZero())
			response, err = client.Do(http.MethodGet, listNamespacePath, nil)
			require.NoError(testingContext, err)
			response.Body.Close()
			require.Equal(testingContext, 2, logins)
			require.Equal(testingContext, 1, apiRequests)
		})
	}
}

func TestClientIncompleteCredentialsAreNotAnonymous(testingContext *testing.T) {
	for _, credential := range []*model.Credential{{AccessKey: "test-user"}, {AccessSecret: "test-secret"}} {
		logins := 0
		client := newTestClient(testingContext, credential, clientTestTransport(func(request *http.Request) (*http.Response, error) {
			require.Equal(testingContext, loginPath, request.URL.Path)
			logins++
			return clientTestResponse(http.StatusUnauthorized, "{}"), nil
		}))
		response, err := client.Do(http.MethodGet, listNamespacePath, nil)
		require.Error(testingContext, err)
		require.Nil(testingContext, response)
		require.Equal(testingContext, 1, logins)
	}
}

func TestClientInvalidatesUnauthorizedWithoutReplay(testingContext *testing.T) {
	for _, status := range []int{http.StatusUnauthorized, http.StatusForbidden} {
		testingContext.Run(http.StatusText(status), func(testingContext *testing.T) {
			logins, apiRequests := 0, 0
			client := newTestClient(testingContext, &model.Credential{AccessKey: "test-user", AccessSecret: "test-secret"}, clientTestTransport(func(request *http.Request) (*http.Response, error) {
				if request.URL.Path == loginPath {
					logins++
					return clientTestResponse(http.StatusOK, fmt.Sprintf(`{"access_token":"token-%d"}`, logins)), nil
				}
				apiRequests++
				require.Equal(testingContext, fmt.Sprintf("Bearer token-%d", logins), request.Header.Get("Authorization"))
				if apiRequests == 1 {
					require.Equal(testingContext, http.MethodPost, request.Method)
					return clientTestResponse(status, "{}"), nil
				}
				return clientTestResponse(http.StatusOK, "{}"), nil
			}))
			response, err := client.Do(http.MethodPost, createNamespacePath, strings.NewReader(`{"orgname":"test"}`))
			require.NoError(testingContext, err)
			require.Equal(testingContext, status, response.StatusCode)
			response.Body.Close()
			require.Equal(testingContext, 1, apiRequests)
			require.Equal(testingContext, 1, logins)
			response, err = client.Do(http.MethodGet, listNamespacePath, nil)
			require.NoError(testingContext, err)
			response.Body.Close()
			require.Equal(testingContext, 2, apiRequests)
			if status == http.StatusUnauthorized {
				require.Equal(testingContext, 2, logins)
			} else {
				require.Equal(testingContext, 1, logins)
			}
		})
	}
}

func TestClientLateUnauthorizedDoesNotInvalidateFreshToken(testingContext *testing.T) {
	var logins atomic.Int32
	started, release := make(chan struct{}), make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })
	client := newTestClient(testingContext, &model.Credential{AccessKey: "test-user", AccessSecret: "test-secret"}, clientTestTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == loginPath {
			return clientTestResponse(http.StatusOK, fmt.Sprintf(`{"access_token":"token-%d"}`, logins.Add(1))), nil
		}
		if request.URL.Path == "/slow" {
			close(started)
			<-release
			return clientTestResponse(http.StatusUnauthorized, "{}"), nil
		}
		return clientTestResponse(http.StatusOK, "{}"), nil
	}))
	finished := make(chan error, 1)
	go func() {
		response, err := client.Do(http.MethodGet, "/slow", nil)
		if response != nil {
			response.Body.Close()
		}
		finished <- err
	}()
	<-started
	client.mu.Lock()
	client.tokenExpiry = time.Now().Add(-time.Second)
	client.mu.Unlock()
	response, err := client.Do(http.MethodGet, listNamespacePath, nil)
	require.NoError(testingContext, err)
	response.Body.Close()
	releaseOnce.Do(func() { close(release) })
	require.NoError(testingContext, <-finished)
	response, err = client.Do(http.MethodGet, listNamespacePath, nil)
	require.NoError(testingContext, err)
	response.Body.Close()
	require.EqualValues(testingContext, 2, logins.Load())
}

func TestClientAuthenticationTimeout(testingContext *testing.T) {
	client := newTestClient(testingContext, &model.Credential{AccessKey: "test-user", AccessSecret: "test-secret"}, clientTestTransport(func(request *http.Request) (*http.Response, error) {
		<-request.Context().Done()
		return nil, request.Context().Err()
	}))
	client.client.Timeout = 20 * time.Millisecond
	response, err := client.Do(http.MethodGet, listNamespacePath, nil)
	require.ErrorIs(testingContext, err, context.DeadlineExceeded)
	require.Nil(testingContext, response)
	client.client.Timeout = config.RegistryHTTPClientTimeout()
	client.client.Transport = clientTestTransport(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == loginPath {
			return clientTestResponse(http.StatusOK, `{"access_token":"recovered"}`), nil
		}
		return clientTestResponse(http.StatusOK, "{}"), nil
	})
	response, err = client.Do(http.MethodGet, listNamespacePath, nil)
	require.NoError(testingContext, err)
	response.Body.Close()
}
