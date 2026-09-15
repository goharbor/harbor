package url

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestURL(t *testing.T) {
	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/req1?mount=sha256&from=test", nil)
	rec1 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec1, req1)
	assert.Equal(http.StatusOK, rec1.Code)

	req2 := httptest.NewRequest(http.MethodPost, "/req2?mount=sha256&from=test;", nil)
	rec2 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec2, req2)
	assert.Equal(http.StatusBadRequest, rec2.Code)

	req3 := httptest.NewRequest(http.MethodGet, "/req3?foo=bar?", nil)
	rec3 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec3, req3)
	assert.Equal(http.StatusOK, rec3.Code)

	req4 := httptest.NewRequest(http.MethodGet, "/req4", nil)
	rec4 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec4, req4)
	assert.Equal(http.StatusOK, rec4.Code)
}

func TestURLWithNullBytes(t *testing.T) {
	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Null byte in value should be rejected
	req1 := httptest.NewRequest(http.MethodGet, "/req1?page=1%00", nil)
	rec1 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec1, req1)
	assert.Equal(http.StatusBadRequest, rec1.Code)

	// Null byte in key should be rejected
	req2 := httptest.NewRequest(http.MethodGet, "/req2?test%00key=value", nil)
	rec2 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec2, req2)
	assert.Equal(http.StatusBadRequest, rec2.Code)
}

func TestURLWithInvalidUTF8(t *testing.T) {
	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Invalid UTF-8 in value should be rejected
	req1 := httptest.NewRequest(http.MethodGet, "/req1?page=%FF%FE", nil)
	rec1 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec1, req1)
	assert.Equal(http.StatusBadRequest, rec1.Code)

	// Invalid UTF-8 in key should be rejected
	req2 := httptest.NewRequest(http.MethodGet, "/req2?%FF%FEkey=value", nil)
	rec2 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec2, req2)
	assert.Equal(http.StatusBadRequest, rec2.Code)

	// Overlong UTF-8 encoding (C0 80 = null byte) should be rejected
	req3 := httptest.NewRequest(http.MethodGet, "/req3?page=%C0%80", nil)
	rec3 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec3, req3)
	assert.Equal(http.StatusBadRequest, rec3.Code)
}

func TestContainsInvalidChars(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid string", "hello", false},
		{"valid utf8", "héllo", false},
		{"valid unicode", "日本語", false},
		{"null byte", "hel\x00lo", true},
		{"invalid utf8", "hel\xfflo", true},
		{"overlong null", string([]byte{0xC0, 0x80}), true},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsInvalidChars(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestQueryParameterKeyValidation(t *testing.T) {
	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		url          string
		expectedCode int
	}{
		// Valid keys
		{"valid key", "/api?page=1", http.StatusOK},
		{"valid key with underscore", "/api?page_size=10", http.StatusOK},
		{"valid key with unicode", "/api?日本語=value", http.StatusOK},
		// Invalid keys - null bytes
		{"null byte at start of key", "/api?%00key=value", http.StatusBadRequest},
		{"null byte in middle of key", "/api?ke%00y=value", http.StatusBadRequest},
		{"null byte at end of key", "/api?key%00=value", http.StatusBadRequest},
		// Invalid keys - invalid UTF-8
		{"invalid utf8 in key", "/api?%FFkey=value", http.StatusBadRequest},
		{"overlong encoding in key", "/api?%C0%80key=value", http.StatusBadRequest},
		// Mixed valid/invalid - key invalid but value valid
		{"invalid key valid value", "/api?test%00=validvalue", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			rec := httptest.NewRecorder()
			Middleware()(next).ServeHTTP(rec, req)
			assert.Equal(tt.expectedCode, rec.Code, "URL: %s", tt.url)
		})
	}
}
