//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxy

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	_ "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/reg"
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/registry"
	testproxy "github.com/goharbor/harbor/src/testing/controller/proxy"
	testreg "github.com/goharbor/harbor/src/testing/pkg/reg"
)

type localInterfaceMock struct {
	mock.Mock
}

func (l *localInterfaceMock) SendPullEvent(ctx context.Context, repo, tag string) {
	panic("implement me")
}

func (l *localInterfaceMock) GetManifest(ctx context.Context, art lib.ArtifactInfo) (*artifact.Artifact, error) {
	args := l.Called(ctx, art)

	var a *artifact.Artifact
	if args.Get(0) != nil {
		a = args.Get(0).(*artifact.Artifact)
	}
	return a, args.Error(1)
}

func (l *localInterfaceMock) SameArtifact(ctx context.Context, repo, tag, dig string) (bool, error) {
	panic("implement me")
}

func (l *localInterfaceMock) BlobExist(ctx context.Context, art lib.ArtifactInfo) (bool, error) {
	args := l.Called(ctx, art)
	return args.Bool(0), args.Error(1)
}

func (l *localInterfaceMock) PushBlob(localRepo string, desc distribution.Descriptor, bReader io.ReadCloser) error {
	panic("implement me")
}

func (l *localInterfaceMock) PushManifest(repo string, tag string, manifest distribution.Manifest) error {
	args := l.Called(repo, tag, manifest)
	return args.Error(0)
}

func (l *localInterfaceMock) PushManifestList(ctx context.Context, repo string, tag string, man distribution.Manifest) error {
	panic("implement me")
}

func (l *localInterfaceMock) CheckDependencies(ctx context.Context, repo string, man distribution.Manifest) []distribution.Descriptor {
	args := l.Called(ctx, repo, man)
	return args.Get(0).([]distribution.Descriptor)
}

func (l *localInterfaceMock) DeleteManifest(repo, ref string) {
}

func (l *localInterfaceMock) UpdatePullTime(ctx context.Context, art lib.ArtifactInfo) error {
	args := l.Called(ctx, art)
	return args.Error(0)
}

type proxyControllerTestSuite struct {
	suite.Suite
	local  *localInterfaceMock
	remote *testproxy.RemoteInterface
	ctr    Controller
	proj   *proModels.Project
}

func (p *proxyControllerTestSuite) SetupTest() {
	p.local = &localInterfaceMock{}
	p.remote = &testproxy.RemoteInterface{}
	p.proj = &proModels.Project{RegistryID: 1}
	p.ctr = &controller{
		blobCtl:     blob.Ctl,
		artifactCtl: artifact.Ctl,
		local:       p.local,
	}
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_True() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)

	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote, p.proj)
	p.Assert().Nil(err)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(true, desc, nil)
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote, p.proj)
	p.Assert().Nil(err)
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_429() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, errors.New("too many requests").WithCode(errors.RateLimitCode))
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	_, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote, p.proj)
	p.Assert().NotNil(err)
	errors.IsRateLimitError(err)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_429ToLocal() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, errors.New("too many requests").WithCode(errors.RateLimitCode))
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote, p.proj)
	p.Assert().Nil(err)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifestWithTag_True() {
	ctx := context.Background()
	art := lib.ArtifactInfo{Repository: "library/hello-world", Tag: "latest"}
	desc := &distribution.Descriptor{}
	// Set up project with ProxyCacheLocalOnNotFound enabled
	proj := &proModels.Project{
		RegistryID: 1,
		Metadata:   map[string]string{proModels.ProMetaProxyCacheLocalOnNotFound: "true"},
	}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote, proj)
	p.Assert().Nil(err)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifestWithTag_NotFoundWhenOptionDisabled() {
	ctx := context.Background()
	art := lib.ArtifactInfo{Repository: "library/hello-world", Tag: "latest"}
	desc := &distribution.Descriptor{}
	// Set up project without ProxyCacheLocalOnNotFound enabled
	proj := &proModels.Project{
		RegistryID: 1,
		Metadata:   map[string]string{},
	}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote, proj)
	p.Assert().NotNil(err)
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalBlob_True() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(true, nil)
	result := p.ctr.UseLocalBlob(ctx, art)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalBlob_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(false, nil)
	result := p.ctr.UseLocalBlob(ctx, art)
	p.Assert().False(result)
}

func TestProxyControllerTestSuite(t *testing.T) {
	suite.Run(t, &proxyControllerTestSuite{})
}

func TestProxyCacheRemoteRepo(t *testing.T) {
	cases := []struct {
		name string
		in   lib.ArtifactInfo
		want string
	}{
		{
			name: `normal test`,
			in:   lib.ArtifactInfo{ProjectName: "dockerhub_proxy", Repository: "dockerhub_proxy/firstfloor/hello-world"},
			want: "firstfloor/hello-world",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := GetRemoteRepo(tt.in)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}
func TestGetRef(t *testing.T) {
	cases := []struct {
		name string
		in   lib.ArtifactInfo
		want string
	}{
		{
			name: `normal`,
			in:   lib.ArtifactInfo{Repository: "hello-world", Tag: "latest", Digest: "sha256:aabbcc"},
			want: "sha256:aabbcc",
		},
		{
			name: `digest_only`,
			in:   lib.ArtifactInfo{Repository: "hello-world", Tag: "", Digest: "sha256:aabbcc"},
			want: "sha256:aabbcc",
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			got := getReference(tt.in)
			if got != tt.want {
				t.Errorf(`(%v) = %v; want "%v"`, tt.in, got, tt.want)
			}
		})
	}
}

func TestProxyBlob(t *testing.T) {
	cases := []struct {
		name       string
		disconnect bool
	}{
		{name: "full read"},
		{name: "client disconnect", disconnect: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			data := bytes.Repeat([]byte("layer"), 1<<20)
			dig := digest.FromBytes(data).String()
			var upstreamN atomic.Int32
			stored := make(chan []byte, 1)

			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v2/library/app/blobs/"+dig {
					http.NotFound(w, r)
					return
				}
				upstreamN.Add(1)
				w.Header().Set("Content-Length", strconv.Itoa(len(data)))
				_, _ = w.Write(data)
			}))
			defer upstream.Close()

			local := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodPost:
					w.Header().Set("Location", r.URL.Path+"1")
					w.WriteHeader(http.StatusAccepted)
				case http.MethodPut:
					b, _ := io.ReadAll(r.Body)
					stored <- b
					w.WriteHeader(http.StatusCreated)
				}
			}))
			defer local.Close()

			regMgr := &testreg.Manager{}
			regMgr.On("Get", mock.Anything, int64(1)).Return(&model.Registry{Type: model.RegistryTypeDockerRegistry, URL: upstream.URL, Status: model.Healthy}, nil)
			orig := reg.Mgr
			reg.Mgr = regMgr
			defer func() { reg.Mgr = orig }()

			ctl := &controller{local: &localHelper{registry: registry.NewClient(local.URL, "", "", true)}}
			art := lib.ArtifactInfo{ProjectName: "proxy", Repository: "proxy/library/app", Digest: dig}
			size, r, err := ctl.ProxyBlob(context.Background(), &proModels.Project{RegistryID: 1}, art)
			require.NoError(t, err)
			assert.Equal(t, int64(len(data)), size)
			if tc.disconnect {
				_, err = io.ReadFull(r, make([]byte, 1024))
				require.NoError(t, err)
			} else {
				got, err := io.ReadAll(r)
				require.NoError(t, err)
				assert.Equal(t, data, got)
			}
			require.NoError(t, r.Close())

			select {
			case b := <-stored:
				assert.Equal(t, data, b)
			case <-time.After(10 * time.Second):
				t.Fatal("blob was not stored locally")
			}
			assert.Equal(t, int32(1), upstreamN.Load())
		})
	}
}

func TestTeeBlobUpstreamError(t *testing.T) {
	pr, pw := io.Pipe()
	upstreamErr := errors.New("boom")
	stored := make(chan error, 1)
	go func() {
		_, err := io.ReadAll(pr)
		stored <- err
	}()
	tb := &teeBlob{upstream: io.NopCloser(io.MultiReader(strings.NewReader("partial"), iotest.ErrReader(upstreamErr))), store: pw}

	_, err := io.ReadAll(tb)
	assert.ErrorIs(t, err, upstreamErr)
	require.NoError(t, tb.Close())
	assert.ErrorIs(t, <-stored, upstreamErr)
}

func TestTeeBlobStoreFails(t *testing.T) {
	data := bytes.Repeat([]byte("x"), 100*1024)
	pr, pw := io.Pipe()
	require.NoError(t, pr.Close())
	tb := &teeBlob{upstream: io.NopCloser(bytes.NewReader(data)), store: pw}

	got, err := io.ReadAll(tb)
	require.NoError(t, err)
	require.NoError(t, tb.Close())
	assert.Equal(t, data, got)
}
