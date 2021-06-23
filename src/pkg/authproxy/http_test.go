package authproxy

import (
	"os"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/usergroup"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/authentication/v1beta1"
	"k8s.io/client-go/rest"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	result := m.Run()
	if result != 0 {
		os.Exit(result)
	}
}

func TestUserFromReviewStatus(t *testing.T) {
	type result struct {
		hasErr      bool
		username    string
		groupLen    int
		adminInAuth bool
	}
	cases := []struct {
		input          v1beta1.TokenReviewStatus
		adminGroups    []string
		adminUsernames []string
		expect         result
	}{
		{
			input: v1beta1.TokenReviewStatus{
				Authenticated: false,
				Error:         "connection error",
			},
			adminGroups:    []string{"admin"},
			adminUsernames: []string{},
			expect: result{
				hasErr: true,
			},
		},
		{
			input: v1beta1.TokenReviewStatus{
				Authenticated: true,
				User: v1beta1.UserInfo{
					Username: "jack",
					UID:      "u-1",
				},
			},
			adminGroups:    []string{"admin"},
			adminUsernames: []string{},
			expect: result{
				hasErr:      false,
				username:    "jack",
				groupLen:    0,
				adminInAuth: false,
			},
		},
		{
			input: v1beta1.TokenReviewStatus{
				Authenticated: true,
				User: v1beta1.UserInfo{
					Username: "daniel",
					Groups:   []string{"group1", "group2"},
				},
				Error: "",
			},
			adminGroups:    []string{"group2", "admin"},
			adminUsernames: []string{},
			expect: result{
				hasErr:      false,
				username:    "daniel",
				groupLen:    2,
				adminInAuth: true,
			},
		},
		{
			input: v1beta1.TokenReviewStatus{
				Authenticated: true,
				User: v1beta1.UserInfo{
					Username: "daniel",
					Groups:   []string{"group1", "group2"},
				},
				Error: "",
			},
			adminGroups:    []string{},
			adminUsernames: []string{"daniel", "admin"},
			expect: result{
				hasErr:      false,
				username:    "daniel",
				groupLen:    2,
				adminInAuth: true,
			},
		},
	}
	for _, c := range cases {
		u, err := UserFromReviewStatus(c.input, c.adminGroups, c.adminUsernames)
		if c.expect.hasErr == true {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, c.expect.username, u.Username)
			assert.Equal(t, c.expect.groupLen, len(u.GroupIDs))
			assert.Equal(t, c.expect.adminInAuth, u.AdminRoleInAuth)
		}
		if u != nil {
			for _, gid := range u.GroupIDs {
				t.Logf("Deleting group %d", gid)
				if err := usergroup.Mgr.Delete(orm.Context(), gid); err != nil {
					panic(err)
				}
			}
		}
	}
}

func TestGetTLSConfig(t *testing.T) {
	certificate := `-----BEGIN CERTIFICATE-----
MIIFXzCCA0egAwIBAgIUY7f2ECRISPMeb1iVNvV5iQsIErUwDQYJKoZIhvcNAQEL
BQAwUjELMAkGA1UEBhMCQ04xDDAKBgNVBAgMA1BFSzERMA8GA1UEBwwIQmVpIEpp
bmcxDzANBgNVBAoMBlZNd2FyZTERMA8GA1UEAwwISGFyYm9yQ0EwHhcNMTkxMTE2
MjI1NjQ0WhcNMjAxMTE1MjI1NjQ0WjBdMQswCQYDVQQGEwJDTjEMMAoGA1UECAwD
UEVLMREwDwYDVQQHDAhCZWkgSmluZzEPMA0GA1UECgwGVk13YXJlMRwwGgYDVQQD
DBNqdC1kZXYuaGFyYm9yLmxvY2FsMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIIC
CgKCAgEAwQ/TQTHwnHwEC/KyHP8Tyv/v35GwXRGW6s1MYoqVnyQMPud0scLHAA2u
PZv2F5jy7PtnhcR0ZHGf05L/igY1utV7/4F2aFgOq0ExYMKxvzilitdcvsxmfTLI
m2pwS8+kH/1s1xR9/7ZPlPSdHuxcHgjtMqorljJykRyq0RBLvXCG+fmAY91kdLil
XWiuIU73lNpZHuXEDl4m2XUzb9cuhwvaHYs7aT6BhwqAJZUjwURUqMe1PIOo7vkQ
cKUHe3u3Fg/vbxfecEr3AHcKfIqm5fwI9vdzj5BP3lGT9hrxduAwI6SgehxGGWP4
aN/cKGIKt/2kzgFoQi/d5p3RBkLVNP/sEyAt9dLJj12ovkQwJzdKDVOy50t3ws9g
Mf3rUUb/wdZADK26lxolep9EXVe4kuWpOo1RvdI+lJJvWc3QaJIoVbr9LM8QN3e7
Iyk3pYRyaQj9EKZ4k0RgWVbIZfRLy1LkGMqmCcIqunHVdGDDjbO/ri8z0sKocMGl
qrqcBTPYmsau7PEcfzJY8k/HUDYdhZgIv2C1iLBl6eoTVDRbrGFcu8LzleWx2/19
OA1Cx7S8WyzN+9mjygqwEYc6qMtoeutAkOA5J8JkxBp0yqjUEnAB6E7R07xQP8AY
IKq5oVpkbD8WRI3w7l/X0AAkDtnijbgYWTfPVGihXHhRtkr/b9cCAwEAAaMiMCAw
HgYDVR0RBBcwFYITanQtZGV2LmhhcmJvci5sb2NhbDANBgkqhkiG9w0BAQsFAAOC
AgEAqwU10WwhI5W8w62vOpT+PKSXRVjHKhm3ltaIzAN7S772hiGl6L81cP9dXZ5y
FN0tFUtUVK01JRJHJaduXNwx24HlwRPNp7mLa4IPpeeVfG14/QCoOd8vxHtKG+V6
zE7Jx2FBVfUJ7P4SngEv4QfvZPt+lCXGK3V1RRTpkLD2knhBfu85rjPi+VW56Z7b
Jb2IEmVRlfR7Z0oYm8z3Obt2XuLIC1/8NtfxthggKr6DeSwZSJXrdGVbyvdiAmk2
iHQch0+UTkRDinL0je6WWbxBoAPXsWA9Hc69o8kmjcXUa99/i8FrC2QxPDUoxxMn
1zWk0jct2Tsr3VZ5HnaI5e8ifG7RUcE5Vr6w7MI5P44Q88zhboP1ShYQ/s513cu2
heELKvO3+mqv96lERtkUUwe8tm1zoPKzQI6ecGuqaTcMbXAGax+ud5XnUlz4xzTI
cByAsQ9DNhYIcOftnfz349zkHeWmMum4uiQwfp/+OrqX+O8U0eJYhlfu9vqCU05T
3mE8Hw5veNdLaZx+mzUVIDzrOB3fh/O62J9CsaZKtxwgLlGiT2ltuC1xUqn3DL8s
pkgODrJUf0p5dhcnLyA2nZolRV1rtwlgJstnEV4JpG1MwtmAZYZUilLvnfpVxTtA
y1bQusZMygQezfCuEzsewF+OpANFovCTUEs6s5vyoVNP8lk=
-----END CERTIFICATE-----
`
	cases := []struct {
		input  *models.HTTPAuthProxy
		expect rest.TLSClientConfig
	}{
		{
			input: &models.HTTPAuthProxy{
				Endpoint:            "https://127.0.0.1/login",
				TokenReviewEndpoint: "https://127.0.0.1/tokenreview",
				VerifyCert:          false,
				SkipSearch:          false,
				ServerCertificate:   "",
			},
			expect: rest.TLSClientConfig{
				Insecure: true,
			},
		},
		{
			input: &models.HTTPAuthProxy{
				Endpoint:            "https://127.0.0.1/login",
				TokenReviewEndpoint: "https://127.0.0.1/tokenreview",
				VerifyCert:          true,
				SkipSearch:          false,
				ServerCertificate:   certificate,
			},
			expect: rest.TLSClientConfig{
				Insecure: false,
				CAData:   []byte(certificate),
			},
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.expect, getTLSConfig(c.input))
	}
}
