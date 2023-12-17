package clientinfo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	"github.com/stretchr/testify/assert"
)

func TestMiddleware(t *testing.T) {

	cases := []struct {
		name     string
		init     func()
		teardown func()
		buildReq func() *http.Request
		checkReq func(t *testing.T, r *http.Request)
	}{
		{
			name: "test ip track on using X-Real-IP",
			buildReq: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("X-Real-IP", "192.168.0.1")
				return r
			},
			checkReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "192.168.0.1", lib.GetClientIPAddress(r.Context()))
			},
			init: func() {
				config.InitWithSettings(map[string]interface{}{
					common.AuditLogTrackIPAddress: "true",
				})
			},
			teardown: func() {
				config.InitWithSettings(test.GetUnitTestConfig())
			},
		},
		{
			name: "test ip track on using X-Forwarded-For",
			buildReq: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("X-Forwarded-For", "192.168.0.1")
				return r
			},
			checkReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "192.168.0.1", lib.GetClientIPAddress(r.Context()))
			},
			init: func() {
				config.InitWithSettings(map[string]interface{}{
					common.AuditLogTrackIPAddress: "true",
				})
			},
			teardown: func() {
				config.InitWithSettings(test.GetUnitTestConfig())
			},
		},
		{
			name: "test ip track on, X-Forwarded-For superiors X-Real-IP",
			buildReq: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("X-Real-IP", "192.168.0.10")
				r.Header.Set("X-Forwarded-For", "192.168.0.1")
				return r
			},
			checkReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "192.168.0.1", lib.GetClientIPAddress(r.Context()))
			},
			init: func() {
				config.InitWithSettings(map[string]interface{}{
					common.AuditLogTrackIPAddress: "true",
				})
			},
			teardown: func() {
				config.InitWithSettings(test.GetUnitTestConfig())
			},
		},
		{
			name: "test ip track off using X-Real-IP",
			buildReq: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("X-Real-IP", "192.168.0.1")
				return r
			},
			checkReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "", lib.GetClientIPAddress(r.Context()))
			},
			init: func() {
				config.InitWithSettings(map[string]interface{}{
					common.AuditLogTrackIPAddress: "false",
				})
			},
			teardown: func() {
				config.InitWithSettings(test.GetUnitTestConfig())
			},
		},
		{
			name: "test user agent track on",
			buildReq: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("User-Agent", "harbor-test")
				return r
			},
			checkReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "harbor-test", lib.GetUserAgent(r.Context()))
			},
			init: func() {
				config.InitWithSettings(map[string]interface{}{
					common.AuditLogTrackUserAgent: "true",
				})
			},
			teardown: func() {
				config.InitWithSettings(test.GetUnitTestConfig())
			},
		},
		{
			name: "test user agent track off",
			buildReq: func() *http.Request {
				r := httptest.NewRequest(http.MethodGet, "/", nil)
				r.Header.Set("User-Agent", "harbor-test")
				return r
			},
			checkReq: func(t *testing.T, r *http.Request) {
				assert.Equal(t, "", lib.GetUserAgent(r.Context()))
			},
			init: func() {
				config.InitWithSettings(map[string]interface{}{
					common.AuditLogTrackUserAgent: "false",
				})
			},
			teardown: func() {
				config.InitWithSettings(test.GetUnitTestConfig())
			},
		},
	}
	for _, c := range cases {
		var r *http.Request
		c.init()
		Middleware()(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			r = req
		})).ServeHTTP(httptest.NewRecorder(), c.buildReq())
		c.teardown()
		t.Run(c.name, func(t *testing.T) {
			c.checkReq(t, r)
		})
	}
}
