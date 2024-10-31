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
	"fmt"
	"io"
	"testing"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	_ "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	testproxy "github.com/goharbor/harbor/src/testing/controller/proxy"
	"github.com/goharbor/harbor/src/testing/lib/cache"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
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

func (l *localInterfaceMock) PullManifest(repo string, ref string) (distribution.Manifest, string, error) {
	args := l.Called(repo, ref)

	var d distribution.Manifest
	if arg := args.Get(0); arg != nil {
		d = arg.(distribution.Manifest)
	}

	return d, args.String(1), args.Error(2)
}

func (l *localInterfaceMock) PushManifest(repo string, tag string, manifest distribution.Manifest) error {
	args := l.Called(repo, tag, manifest)
	return args.Error(0)
}

func (l *localInterfaceMock) PushManifestList(ctx context.Context, repo string, tag string, man distribution.Manifest) error {
	panic("implement me")
}

func (l *localInterfaceMock) CheckDependencies(ctx context.Context, repo string, man distribution.Manifest) []distribution.Descriptor {
	panic("implement me")
}

func (l *localInterfaceMock) DeleteManifest(repo, ref string) {
}

type proxyControllerTestSuite struct {
	suite.Suite
	local  *localInterfaceMock
	remote *testproxy.RemoteInterface
	ctr    *controller
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

	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().Nil(err)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	repo := "library/hello-world"
	art := lib.ArtifactInfo{Repository: repo, Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(true, desc, nil)
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	p.local.On("PullManifest", repo, string(desc.Digest)).Times(1).Return(nil, "", fmt.Errorf("could not pull manifest"))
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().Nil(err)
	p.Assert().False(result)
	p.local.AssertExpectations(p.T())
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_429() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, errors.New("too many requests").WithCode(errors.RateLimitCode))
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	_, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
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
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().Nil(err)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifestWithTag_False() {
	ctx := context.Background()
	art := lib.ArtifactInfo{Repository: "library/hello-world", Tag: "latest"}
	desc := &distribution.Descriptor{}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(false, desc, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
	p.Assert().True(errors.IsNotFoundErr(err))
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifestWithTag_LocalRepoTrueManifest() {
	manifest := `{ 
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.manifest.v1+json",
		"config": {
			 "mediaType": "application/vnd.example.config.v1+json",
			 "digest": "sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03",
			 "size": 123
		},
		"layers": [
			 {
					"mediaType": "application/vnd.example.data.v1.tar+gzip",
					"digest": "sha256:e258d248fda94c63753607f7c4494ee0fcbe92f1a76bfdac795c9d84101eb317",
					"size": 1234
			 }
		],
		"annotations": {
			 "com.example.key1": "value1"
		}
	}`
	man, desc, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifest))
	p.Require().NoError(err)
	mediaType, payload, err := man.Payload()
	p.Require().NoError(err)

	ctx := context.Background()
	repo := "library/hello-world"
	art := lib.ArtifactInfo{Repository: repo, Tag: "latest"}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(true, &desc, nil)
	p.local.On("PullManifest", repo, string(desc.Digest)).Times(1).Return(man, string(desc.Digest), nil)

	result, manifests, err := p.ctr.UseLocalManifest(ctx, art, p.remote)

	p.Assert().NoError(err)
	p.Assert().True(result)
	p.Assert().NotNil(manifests)
	p.Assert().Equal(mediaType, manifests.ContentType())
	p.Assert().Equal(string(desc.Digest), manifests.Digest())
	p.Assert().Equal(payload, manifests.Content())

	p.local.AssertExpectations(p.T())
}

func (p *proxyControllerTestSuite) TestUseLocalManifestWithTag_CacheTrueManifestList() {
	c := cache.NewCache(p.T())
	p.ctr.cache = c

	ctx := context.Background()
	repo := "library/hello-world"
	art := lib.ArtifactInfo{Repository: repo, Tag: "latest"}
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	desc := &distribution.Descriptor{Digest: digest.Digest(dig)}
	content := "some content"
	contentType := "some content type"
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(true, desc, nil)
	p.local.On("PullManifest", repo, string(desc.Digest)).Times(1).Return(nil, "", fmt.Errorf("could not pull manifest"))
	artInfoWithDigest := art
	artInfoWithDigest.Digest = dig
	c.On("Fetch", mock.Anything, manifestListKey(art.Repository, artInfoWithDigest), mock.Anything).
		Times(1).
		Run(func(args mock.Arguments) {
			ct := args.Get(2).(*[]byte)
			*ct = []byte(content)
		}).
		Return(nil)
	c.On("Fetch", mock.Anything, manifestListContentTypeKey(art.Repository, artInfoWithDigest), mock.Anything).
		Times(1).
		Run(func(args mock.Arguments) {
			ct := args.Get(2).(*string)
			*ct = contentType
		}).
		Return(nil)

	result, manifests, err := p.ctr.UseLocalManifest(ctx, art, p.remote)

	p.Assert().NoError(err)
	p.Assert().True(result)
	p.Assert().NotNil(manifests)
	p.Assert().Equal(contentType, manifests.ContentType())
	p.Assert().Equal(string(desc.Digest), manifests.Digest())
	p.Assert().Equal([]byte(content), manifests.Content())

	p.local.AssertExpectations(p.T())
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
			got := getRemoteRepo(tt.in)
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
