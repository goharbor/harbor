package csrf

import (
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

func resetMiddleware() {
	once = sync.Once{}
}

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	conf := map[string]any{}
	config.InitWithSettings(conf)
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

type handler struct {
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestMiddleware(t *testing.T) {
	cases := []struct {
		req         *http.Request
		statusCode  int
		returnToken bool
	}{
		{
			req:         httptest.NewRequest(http.MethodGet, "/", nil),
			statusCode:  http.StatusOK,
			returnToken: true,
		},
		{
			req:         httptest.NewRequest(http.MethodDelete, "/", nil),
			statusCode:  http.StatusForbidden,
			returnToken: true,
		},
		{
			req:         httptest.NewRequest(http.MethodGet, "/api/2.0/projects", nil), // should be skipped
			statusCode:  http.StatusOK,
			returnToken: false,
		},
		{
			req:         httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/latest", nil), // should be skipped
			statusCode:  http.StatusOK,
			returnToken: false,
		},
	}
	for _, c := range cases {
		srv := Middleware()(&handler{})
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, c.req)
		assert.Equal(t, c.statusCode, rec.Result().StatusCode)
		assert.Equal(t, c.returnToken, rec.Result().Header.Get(tokenHeader) != "")
	}
}

func TestMiddlewareInvalidKey(t *testing.T) {
	originalEnv := os.Getenv(csrfKeyEnv)
	defer os.Setenv(csrfKeyEnv, originalEnv)

	t.Run("invalid CSRF key", func(t *testing.T) {
		os.Setenv(csrfKeyEnv, "invalidkey")
		resetMiddleware()
		middleware := Middleware()
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be reached when CSRF key is invalid")
		})

		handler := middleware(testHandler)
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestSecureCookie(t *testing.T) {
	assert.True(t, secureCookie())
	conf := map[string]any{
		common.ExtEndpoint: "http://harbor.test",
	}
	config.InitWithSettings(conf)

	assert.False(t, secureCookie())
	conf = map[string]any{}
	config.InitWithSettings(conf)
}
