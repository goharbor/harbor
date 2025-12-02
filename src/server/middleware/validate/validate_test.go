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

package validate

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {
	tests := []struct {
		name         string
		inputURL     string
		expectedCode int
		shouldPass   bool
	}{
		{
			name:         "normal query passes",
			inputURL:     "/api/v2.0/projects?name=test",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "null byte rejected",
			inputURL:     "/api/v2.0/projects?name=test%00inject",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "multiple null bytes rejected",
			inputURL:     "/api/v2.0/projects?name=%00test%00%00value%00",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "invalid UTF-8 rejected",
			inputURL:     "/api/v2.0/projects?name=test%FF%FEvalue",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "overlong UTF-8 null rejected",
			inputURL:     "/api/v2.0/projects?name=test%C0%80value",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "unicode preserved",
			inputURL:     "/api/v2.0/projects?name=项目名称",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "emoji preserved",
			inputURL:     "/api/v2.0/projects?name=test🚀project",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
		{
			name:         "multiple params with null byte rejected",
			inputURL:     "/api/v2.0/projects?name=test%00&page=1",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "url in param with null byte rejected",
			inputURL:     "/api/v2.0/projects?name=http://evil.com/path%3F%00.php",
			expectedCode: http.StatusBadRequest,
			shouldPass:   false,
		},
		{
			name:         "special chars preserved",
			inputURL:     "/api/v2.0/projects?q=name%3D~test",
			expectedCode: http.StatusOK,
			shouldPass:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handlerCalled := false

			handler := Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, tt.inputURL, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, tt.shouldPass, handlerCalled, "handler called mismatch")
		})
	}
}

func TestValidateQueryString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expectErr bool
	}{
		{
			name:      "empty string",
			input:     "",
			expectErr: false,
		},
		{
			name:      "normal query",
			input:     "name=test&page=1",
			expectErr: false,
		},
		{
			name:      "null byte in value",
			input:     "name=test%00value",
			expectErr: true,
		},
		{
			name:      "preserves special chars",
			input:     "q=name%3D~test",
			expectErr: false,
		},
		{
			name:      "unicode valid",
			input:     "name=%E4%B8%AD%E6%96%87",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateQueryString(tt.input)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMiddlewareNoQueryString(t *testing.T) {
	called := false
	handler := Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v2.0/projects", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestMiddlewarePostBody(t *testing.T) {
	// POST body is not validated by this middleware (JSON parser handles it)
	handler := Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v2.0/projects", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
