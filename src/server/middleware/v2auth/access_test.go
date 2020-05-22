package v2auth

import (
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/lib"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestAccessList(t *testing.T) {
	req1, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/", nil)
	req2, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/_catalog", nil)
	req3, _ := http.NewRequest(http.MethodPost, "https://registry.test/v2/library/ubuntu/blobs/uploads/?mount=sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f&from=base/ubuntu", nil)
	ctx3 := lib.WithArtifactInfo(context.Background(), lib.ArtifactInfo{
		Repository:          "library/ubuntu",
		BlobMountRepository: "base/ubuntu",
		BlobMountDigest:     "sha256:08e4a417ff4e3913d8723a05cc34055db01c2fd165b588e049c5bad16ce6094f",
	})
	req3 = req3.WithContext(ctx3)
	req4, _ := http.NewRequest(http.MethodGet, "https://registry.test/v2/goharbor/registry/manifests/v1.0", nil)
	req4d, _ := http.NewRequest(http.MethodDelete, "https://registry.test/v2/goharbor/registry/manifests/v1.0", nil)
	ctx4 := lib.WithArtifactInfo(context.Background(), lib.ArtifactInfo{
		Repository: "goharbor/registry",
		Tag:        "v1.0",
		Reference:  "v1.0",
	})
	req4 = req4.WithContext(ctx4)
	req4d = req4d.WithContext(ctx4)
	req5, _ := http.NewRequest(http.MethodGet, "https://registry.test/api/v2.0/users", nil)
	cases := []struct {
		input  *http.Request
		expect []access
	}{
		{
			input: req1,
			expect: []access{{
				target: login,
			}},
		},
		{
			input: req2,
			expect: []access{{
				target: catalog,
			}},
		},
		{
			input: req3,
			expect: []access{{
				target: repository,
				name:   "library/ubuntu",
				action: rbac.ActionPush,
			},
				{
					target: repository,
					name:   "base/ubuntu",
					action: rbac.ActionPull,
				}},
		},
		{
			input: req4,
			expect: []access{{
				target: repository,
				name:   "goharbor/registry",
				action: rbac.ActionPull,
			}},
		},
		{
			input: req4d,
			expect: []access{{
				target: repository,
				name:   "goharbor/registry",
				action: rbac.ActionDelete,
			}},
		},
		{
			input:  req5,
			expect: []access{},
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expect, accessList(c.input))
	}
}

func TestScopeStr(t *testing.T) {
	cases := []struct {
		acs   access
		scope string
	}{
		{
			acs: access{
				target: login,
			},
			scope: "",
		},
		{
			acs: access{
				target: catalog,
			},
			scope: "",
		},
		{
			acs: access{
				target: repository,
				name:   "goharbor/registry",
				action: rbac.ActionPull,
			},
			scope: "repository:goharbor/registry:pull",
		},
		{
			acs: access{
				target: repository,
				name:   "library/golang",
				action: rbac.ActionPush,
			},
			scope: "repository:library/golang:pull,push",
		},
		{
			acs: access{
				target: repository,
				name:   "library/golang",
				action: rbac.ActionDelete,
			},
			scope: "repository:library/golang:delete",
		},
		{
			acs: access{
				target: repository,
				name:   "library/golang",
				action: rbac.ActionList,
			},
			scope: "",
		},
		{
			acs:   access{},
			scope: "",
		},
	}

	for _, c := range cases {
		a := c.acs
		assert.Equal(t, c.scope, a.scopeStr(context.Background()))
	}
}
