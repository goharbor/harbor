package manifestinfo

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/utils/test"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"reflect"
)

type mfinfoTestSuite struct {
	suite.Suite
	require *require.Assertions
	assert  *assert.Assertions
}

func (t *mfinfoTestSuite) SetupSuite() {
	t.require = require.New(t.T())
	t.assert = assert.New(t.T())
	test.InitDatabaseFromEnv()
}

func (t *mfinfoTestSuite) TestParseManifestInfoFromPath() {
	mustRequest := func(method, url string) *http.Request {
		req, _ := http.NewRequest(method, url, nil)
		return req
	}

	type args struct {
		req *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    *middleware.ManifestInfo
		wantErr bool
	}{
		{
			"ok for digest",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifests/sha256:3e17b60ab9d92d953fb8ebefa25624c0d23fb95f78dde5572285d10158044059")},
			&middleware.ManifestInfo{
				ProjectID:  1,
				Repository: "library/photon",
				Digest:     "sha256:3e17b60ab9d92d953fb8ebefa25624c0d23fb95f78dde5572285d10158044059",
			},
			false,
		},
		{
			"ok for tag",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifests/latest")},
			&middleware.ManifestInfo{
				ProjectID:  1,
				Repository: "library/photon",
				Tag:        "latest",
			},
			false,
		},
		{
			"project not found",
			args{mustRequest(http.MethodDelete, "/v2/notfound/photon/manifests/latest")},
			nil,
			true,
		},
		{
			"url not match",
			args{mustRequest(http.MethodDelete, "/v2/library/photon/manifest/latest")},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func() {
			got, err := parseManifestInfoFromPath(tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf(err, fmt.Sprintf("ParseManifestInfoFromPath() error = %v, wantErr %v", err, tt.wantErr))
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf(err, fmt.Sprintf("ParseManifestInfoFromPath() = %v, want %v", got, tt.want))
			}
		})
	}
}

func (t *mfinfoTestSuite) TestResolveManifest() {

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	req := httptest.NewRequest(http.MethodDelete, "/v2/library/hello-world/manifests/latest", nil)
	rec := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec, req)
	t.assert.Equal(rec.Code, http.StatusOK)

	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	t.assert.True(ok)
	t.assert.Equal(mf.Tag, "latest")
	t.assert.Equal(mf.ProjectID, int64(1))
}
