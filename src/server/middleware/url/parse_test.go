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
