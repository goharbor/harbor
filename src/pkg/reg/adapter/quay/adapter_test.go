package quay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common/utils/test"
	adp "github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/stretchr/testify/assert"
)

func getMockAdapter(t *testing.T) (*adapter, *httptest.Server) {
	server := test.NewServer(
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/quay/busybox/manifests/latest",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				// sample test data
				data := `{"name":"quay/busybox","tag":"latest","architecture":"amd64","fsLayers":[{"blobSum":"sha256:ee780d08a5b4de5192a526d422987f451d9a065e6da42aefe8c3b20023a250c7"},{"blobSum":"sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"},{"blobSum":"sha256:a3ed95caeb02ffe68cdd9fd84406680ae93d633cb16422d00e8a7c22955b46d4"},{"blobSum":"sha256:9c075fe2c773108d2fe2c18ea170548b0ee30ef4e5e072d746e3f934e788b734"}],"history":[{"v1Compatibility":"{\"architecture\":\"amd64\",\"config\":{\"Env\":[\"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin\",\"TERM=xterm\",\"container=podman\",\"HOSTNAME=\"],\"Cmd\":[\"/bin/sh\"],\"WorkingDir\":\"/\"},\"created\":\"2020-09-04T16:08:39.002329276Z\",\"id\":\"94cc2cc706952747940d82296ae52449f4bb31a5ee40bb5a36145929775d80a5\",\"os\":\"linux\",\"parent\":\"17d40c067945c6ba5bd16e84fd2454b645bc85471e9222f64cc4af9dfbead85d\"}"},{"v1Compatibility":"{\"id\":\"17d40c067945c6ba5bd16e84fd2454b645bc85471e9222f64cc4af9dfbead85d\",\"parent\":\"6e85bfb29b9b917e6b97a3aa6339f150072a7953210013390c7cbf405eaea31c\",\"comment\":\"Updated at 2020-09-04 15:06:14 +0000\",\"created\":\"2020-09-04T15:06:15.059006882Z\",\"container_config\":{\"Cmd\":[\"sh echo \\\"2020-09-04 15:06:14 +0000\\\" \\u003e foo\"]},\"throwaway\":true}"},{"v1Compatibility":"{\"id\":\"6e85bfb29b9b917e6b97a3aa6339f150072a7953210013390c7cbf405eaea31c\",\"parent\":\"dacca7329c6e094ed91cc673249ff788cb7959ef6f53517f363660145bbc7464\",\"created\":\"2020-09-01T00:36:18.335938403Z\",\"container_config\":{\"Cmd\":[\"/bin/sh -c #(nop)  CMD [\\\"sh\\\"]\"]},\"throwaway\":true}"},{"v1Compatibility":"{\"id\":\"dacca7329c6e094ed91cc673249ff788cb7959ef6f53517f363660145bbc7464\",\"created\":\"2020-09-01T00:36:18.002487153Z\",\"container_config\":{\"Cmd\":[\"/bin/sh -c #(nop) ADD file:4e5169fa630e0afede3b7db9a6d0ca063df3fe69cc2873a6c50e9801d61f563f in / \"]}}"}],"schemaVersion":1,"signatures":[{"header":{"jwk":{"crv":"P-256","kid":"BU64:Q7B6:FC7P:O23C:H46C:5J5O:DRH6:F7XL:MB2U:YD4Z:4IT2:4XAT","kty":"EC","x":"xOLxWrHtPC8j0WFZd4l5TErbzO_p5FWiYYamVEOtEjI","y":"Ze-LCvjtbpEJRSKbRlQXQYlBcAjoTNNTj_GvaDieXy0"},"alg":"ES256"},"signature":"8XK-3ScG84TOXGxX-Cwi0j1nEjqOIaR0U0n6_N4BpK0X9hCvf3aMEbeln_bO2mrKTfuJWP5KAbN-SgekcJm92Q","protected":"eyJmb3JtYXRMZW5ndGgiOjE4OTUsImZvcm1hdFRhaWwiOiJmUSIsInRpbWUiOiIyMDIwLTA5LTA0VDE2OjA5OjA5WiJ9"}]}`
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(data))
			},
		},
		&test.RequestHandlerMapping{
			Method:  http.MethodGet,
			Pattern: "/v2/",
			Handler: func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Method, r.URL)
				fmt.Println(555)
				w.WriteHeader(http.StatusOK)
			},
		},
	)

	factory, _ := adp.GetFactory(model.RegistryTypeQuay)
	ad, err := factory.Create(&model.Registry{
		Type: model.RegistryTypeQuay,
		URL:  server.URL,
	})
	assert.Nil(t, err)
	a := ad.(*adapter)
	return a, server
}

func TestAdapter_NewAdapter(t *testing.T) {
	factory, err := adp.GetFactory("BadName")
	assert.Nil(t, factory)
	assert.NotNil(t, err)

	factory, err = adp.GetFactory(model.RegistryTypeQuay)
	assert.Nil(t, err)
	assert.NotNil(t, factory)
}

func TestAdapter_HealthCheck(t *testing.T) {
	a, s := getMockAdapter(t)
	defer s.Close()

	health, err := a.HealthCheck()
	assert.Nil(t, err)
	assert.Equal(t, string(health), model.Healthy)
}

func TestAdapter_Info(t *testing.T) {
	a, s := getMockAdapter(t)
	defer s.Close()

	info, err := a.Info()
	assert.Nil(t, err)
	t.Log(info)
}

func TestAdapter_PullManifests(t *testing.T) {
	a, s := getMockAdapter(t)
	defer s.Close()

	registry, _, err := a.PullManifest("quay/busybox", "latest")

	assert.Nil(t, err)
	assert.NotNil(t, registry)
}
