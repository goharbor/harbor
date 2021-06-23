package csrf

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	conf := map[string]interface{}{}
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
	srv := Middleware()(&handler{})
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
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, c.req)
		assert.Equal(t, c.statusCode, rec.Result().StatusCode)
		assert.Equal(t, c.returnToken, rec.Result().Header.Get(tokenHeader) != "")
	}
}

func hasCookie(resp *http.Response, name string) bool {
	for _, c := range resp.Cookies() {
		if c != nil && c.Name == name {
			return true
		}
	}
	return false
}

func TestSecureCookie(t *testing.T) {
	assert.True(t, secureCookie())
	conf := map[string]interface{}{
		common.ExtEndpoint: "http://harbor.test",
	}
	config.InitWithSettings(conf)

	assert.False(t, secureCookie())
	conf = map[string]interface{}{}
	config.InitWithSettings(conf)
}
