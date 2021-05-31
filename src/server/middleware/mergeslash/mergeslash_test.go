package mergeslash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type handler struct {
	path string
}

func (h *handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.path = req.URL.Path
	w.WriteHeader(200)
}

func TestMergeSlash(t *testing.T) {
	next := &handler{}
	rec := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "https://test.local/api/v2.0/systeminfo/", nil)
	req2, _ := http.NewRequest(http.MethodGet, "https://test.local/v2//////_catalog", nil)
	req3, _ := http.NewRequest(http.MethodPost, "https://test.local/v2/library///////ubuntu//blobs/uploads///////", nil)
	req4, _ := http.NewRequest(http.MethodGet, "https://test.local//api/v2.0///////artifacts?scan_overview=false", nil)
	cases := []struct {
		req          *http.Request
		expectedPath string
	}{
		{
			req:          req1,
			expectedPath: "/api/v2.0/systeminfo/",
		},
		{
			req:          req2,
			expectedPath: "/v2/_catalog",
		},
		{
			req:          req3,
			expectedPath: "/v2/library/ubuntu/blobs/uploads/",
		},
		{
			req:          req4,
			expectedPath: "/api/v2.0/artifacts",
		},
	}
	for _, tt := range cases {
		Middleware()(next).ServeHTTP(rec, tt.req)
		assert.Equal(t, tt.expectedPath, next.path)
	}
}
