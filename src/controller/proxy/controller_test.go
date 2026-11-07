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
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	testproxy "github.com/goharbor/harbor/src/testing/controller/proxy"
	cachetesting "github.com/goharbor/harbor/src/testing/lib/cache"
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

func (p *proxyControllerTestSuite) SetupSuite() {
	copyContext = func(ctx context.Context) context.Context { return ctx }
}

func (p *proxyControllerTestSuite) TearDownSuite() {
	copyContext = orm.Copy
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

func (p *proxyControllerTestSuite) cachedManifestList(tagged *artifact.Artifact) (lib.ArtifactInfo, *cachetesting.Cache) {
	repo := "library/hello-world"
	manList, _ := testManifestList(p.T())
	dig := payloadDigest(p.T(), manList)
	ct, payload, err := manList.Payload()
	p.Require().NoError(err)
	c := &cachetesting.Cache{}
	p.ctr = &controller{
		local:           p.local,
		cache:           c,
		handlerRegistry: map[string]ManifestCacheHandler{ct: &ManifestListCache{local: p.local, cache: c}},
	}
	art := lib.ArtifactInfo{Repository: repo, Tag: "latest"}
	withDigest := lib.ArtifactInfo{Repository: repo, Tag: "latest", Digest: dig}
	p.local.On("GetManifest", mock.Anything, art).Return(tagged, nil)
	p.remote.On("ManifestExist", repo, "latest").Return(true, &distribution.Descriptor{Digest: digest.Digest(dig)}, nil)
	c.On("Fetch", mock.Anything, manifestListKey(repo, withDigest), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]byte) = payload
	}).Return(nil)
	c.On("Fetch", mock.Anything, manifestListContentTypeKey(repo, withDigest), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*string) = ct
	}).Return(nil)
	return withDigest, c
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_CachedManifestListPushedWhenTagMissing() {
	art, c := p.cachedManifestList(nil)
	p.local.On("GetManifest", mock.Anything, lib.ArtifactInfo{Repository: art.Repository, Digest: testAmdDigest}).Return(&artifact.Artifact{}, nil)
	p.local.On("GetManifest", mock.Anything, lib.ArtifactInfo{Repository: art.Repository, Digest: testArmDigest}).Return(&artifact.Artifact{}, nil)
	c.On("Save", mock.Anything, TrimmedManifestlist+art.Digest, art.Digest).Return(nil).Once()
	c.On("Save", mock.Anything, trimmedDigestKey(art.Repository, art.Digest), art.Digest, manifestListCacheInterval).Return(nil).Once()
	p.local.On("UpdatePullTime", mock.Anything, mock.Anything).Return(nil)
	pushed := make(chan string, 1)
	p.local.On("PushManifest", art.Repository, "latest", mock.Anything).Run(func(args mock.Arguments) {
		pushed <- payloadDigest(p.T(), args.Get(2).(distribution.Manifest))
	}).Return(nil).Once()

	useLocal, ml, err := p.ctr.UseLocalManifest(context.Background(), lib.ArtifactInfo{Repository: art.Repository, Tag: art.Tag}, p.remote, p.proj)
	p.Require().NoError(err)
	p.True(useLocal)
	p.NotNil(ml)
	select {
	case dig := <-pushed:
		p.Equal(art.Digest, dig)
	case <-time.After(5 * time.Second):
		p.Fail("the cached manifest list is not pushed by tag")
	}
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_CachedManifestListAlreadyTagged() {
	tagged := &artifact.Artifact{}
	art, _ := p.cachedManifestList(tagged)
	tagged.Digest = art.Digest

	useLocal, ml, err := p.ctr.UseLocalManifest(context.Background(), lib.ArtifactInfo{Repository: art.Repository, Tag: art.Tag}, p.remote, p.proj)
	p.Require().NoError(err)
	p.True(useLocal)
	p.NotNil(ml)
	p.local.AssertNotCalled(p.T(), "PushManifest", mock.Anything, mock.Anything, mock.Anything)
}

func (p *proxyControllerTestSuite) TestManifestListTagged() {
	art := lib.ArtifactInfo{Repository: "library/hello-world", Tag: "latest", Digest: testAmdDigest}
	withDigest := func(dig string) *artifact.Artifact {
		a := &artifact.Artifact{}
		a.Digest = dig
		return a
	}
	cases := []struct {
		name    string
		a       *artifact.Artifact
		trimmed string
		want    bool
	}{
		{name: "tag missing", a: nil, want: false},
		{name: "same digest", a: withDigest(testAmdDigest), want: true},
		{name: "trimmed digest", a: withDigest(testArmDigest), trimmed: testArmDigest, want: true},
		{name: "stale tag", a: withDigest(testArmDigest), trimmed: "sha256:" + strings.Repeat("0", 64), want: false},
		{name: "trimmed digest not cached", a: withDigest(testArmDigest), want: false},
	}
	for _, tc := range cases {
		p.Run(tc.name, func() {
			c := &cachetesting.Cache{}
			call := c.On("Fetch", mock.Anything, trimmedDigestKey(art.Repository, art.Digest), mock.Anything)
			if tc.trimmed != "" {
				call.Run(func(args mock.Arguments) { *args.Get(2).(*string) = tc.trimmed }).Return(nil)
			} else {
				call.Return(cache.ErrNotFound)
			}
			ctl := &controller{cache: c}
			p.Equal(tc.want, ctl.manifestListTagged(context.Background(), tc.a, art))
		})
	}
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
