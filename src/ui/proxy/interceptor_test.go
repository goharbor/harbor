package proxy

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestMatchPullManifest(t *testing.T) {
	assert := assert.New(t)
	req1, _ := http.NewRequest("POST", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	res1, _, _ := MatchPullManifest(req1)
	assert.False(res1, "%s %v is not a request to pull manifest", req1.Method, req1.URL)

	req2, _ := http.NewRequest("GET", "http://192.168.0.3:80/v2/library/ubuntu/manifests/14.04", nil)
	res2, repo2, tag2 := MatchPullManifest(req2)
	assert.True(res2, "%s %v is a request to pull manifest", req2.Method, req2.URL)
	assert.Equal("library/ubuntu", repo2)
	assert.Equal("14.04", tag2)

	req3, _ := http.NewRequest("GET", "https://192.168.0.5:443/v1/library/ubuntu/manifests/14.04", nil)
	res3, _, _ := MatchPullManifest(req3)
	assert.False(res3, "%s %v is not a request to pull manifest", req3.Method, req3.URL)

	req4, _ := http.NewRequest("GET", "https://192.168.0.5/v2/library/ubuntu/manifests/14.04", nil)
	res4, repo4, tag4 := MatchPullManifest(req4)
	assert.True(res4, "%s %v is a request to pull manifest", req4.Method, req4.URL)
	assert.Equal("library/ubuntu", repo4)
	assert.Equal("14.04", tag4)

	req5, _ := http.NewRequest("GET", "https://myregistry.com/v2/path1/path2/golang/manifests/1.6.2", nil)
	res5, repo5, tag5 := MatchPullManifest(req5)
	assert.True(res5, "%s %v is a request to pull manifest", req5.Method, req5.URL)
	assert.Equal("path1/path2/golang", repo5)
	assert.Equal("1.6.2", tag5)

	req6, _ := http.NewRequest("GET", "https://myregistry.com/v2/myproject/registry/manifests/sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", nil)
	res6, repo6, tag6 := MatchPullManifest(req6)
	assert.True(res6, "%s %v is a request to pull manifest", req6.Method, req6.URL)
	assert.Equal("myproject/registry", repo6)
	assert.Equal("sha256:ca4626b691f57d16ce1576231e4a2e2135554d32e13a85dcff380d51fdd13f6a", tag6)
}
