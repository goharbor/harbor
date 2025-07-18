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
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/beego/beego/v2/client/orm"
	"github.com/docker/distribution"
	schema1 "github.com/docker/distribution/manifest/schema1"
	_ "github.com/jackc/pgx/v4/stdlib" // registry pgx driver
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/lib"
	_ "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/config/inmemory"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	testproxy "github.com/goharbor/harbor/src/testing/controller/proxy"
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
	args := l.Called(localRepo, desc, bReader)
	return args.Error(0)
}

func (l *localInterfaceMock) PullBlob(localRepo string, digest digest.Digest) (int64, io.ReadCloser, error) {
	args := l.Called(localRepo, digest)
	var size int64
	var bReader io.ReadCloser
	if args.Get(0) != nil {
		size = args.Get(0).(int64)
	}
	if args.Get(1) != nil {
		bReader = io.NopCloser(bytes.NewReader(args.Get(1).([]byte)))
	}
	return size, bReader, args.Error(2)
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

type proxyControllerTestSuite struct {
	suite.Suite
	cfg    config.Manager
	local  *localInterfaceMock
	remote *testproxy.RemoteInterface
	ctr    Controller
	proj   *proModels.Project
}

func (p *proxyControllerTestSuite) SetupSuite() {
	// Register a global in memory config manager for testing
	config.Register(common.InMemoryCfgManager, inmemory.NewInMemoryManager())
	config.DefaultCfgManager = common.InMemoryCfgManager

	// This doesn't really do anything other than allow us to run `orm.Copy` in the background
	// prior to submitting events that depend on a database connection. The database connection
	// is not actually used in these proxy controller tests.
	if err := orm.RegisterDriver("pgx", orm.DRPostgres); err != nil {
		p.T().Fatalf("Failed to register test database driver: %v", err)
	}
	db, _, err := sqlmock.New()
	if err != nil {
		p.T().Fatalf("Failed to create sqlmock: %v", err)
	}
	if err := orm.AddAliasWthDB("default", "pgx", db); err != nil {
		p.T().Fatalf("Failed to add alias with db: %v", err)
	}
}

func (p *proxyControllerTestSuite) SetupTest() {
	cfg, err := config.GetManager(config.DefaultCfgManager)
	if err != nil {
		p.T().Fatalf("Failed to get config manager: %v", err)
	}

	p.cfg = cfg
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
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Return(true, desc, nil)
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	result, _, err := p.ctr.UseLocalManifest(ctx, art, p.remote)
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

func (p *proxyControllerTestSuite) TestUseLocalBlob_True() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(true, nil)
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	result := p.ctr.UseLocalBlob(ctx, art)
	p.Assert().True(result)
}

func (p *proxyControllerTestSuite) TestUseLocalBlob_False() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: dig}
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(false, nil)
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil).Maybe()
	result := p.ctr.UseLocalBlob(ctx, art)
	p.Assert().False(result)
}

func (p *proxyControllerTestSuite) TestUseLocalManifest_EnsureSingleRemoteRequest() {
	ctx := context.Background()
	artInfo := lib.ArtifactInfo{
		Repository:  "library/hello-world",
		ProjectName: "library",
		Digest:      "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b",
	}
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)

	// force a not_found error
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// wait for requests to queue up
		time.Sleep(10 * time.Millisecond)
	}).Return(false, nil, nil).Once()

	// run 10 concurrent requests
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := p.ctr.UseLocalManifest(ctx, artInfo, p.remote)
			p.Assert().True(errors.IsNotFoundErr(err))
		}()
	}

	wg.Wait()

	// expect only one remote request
	p.remote.AssertExpectations(p.T())
}

func (p *proxyControllerTestSuite) TestProxyManifest_EnsureSingleRemoteRequest() {
	ctx := context.Background()
	artInfo := lib.ArtifactInfo{
		Repository:  "library/hello-world",
		ProjectName: "library",
		Digest:      "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b",
	}
	man := &schema1.SignedManifest{
		Manifest: schema1.Manifest{
			Name: "library/hello-world",
			Tag:  "latest",
		},
	}

	// Set up mock expectations for methods called by background goroutines
	p.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)

	// return a valid manifest
	p.remote.On("Manifest", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// wait for requests to queue up
		time.Sleep(10 * time.Millisecond)
	}).Return(man, "", nil).Once()

	// run 10 concurrent requests
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			actualMan, err := p.ctr.ProxyManifest(ctx, artInfo, p.remote)
			p.Assert().Nil(err)
			p.Assert().Equal(man, actualMan)
		}()
	}

	wg.Wait()

	// expect only one remote request
	p.remote.AssertExpectations(p.T())
}

func (p *proxyControllerTestSuite) TestHeadManifest_EnsureSingleRemoteRequest() {
	ctx := context.Background()
	dig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	artInfo := lib.ArtifactInfo{
		Repository:  "library/hello-world",
		ProjectName: "library",
		Digest:      dig,
	}

	// return a valid manifest
	p.remote.On("ManifestExist", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		time.Sleep(10 * time.Millisecond)
	}).Return(true, &distribution.Descriptor{Digest: digest.Digest(dig)}, nil).Once()

	// run 10 concurrent requests
	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			exists, desc, err := p.ctr.HeadManifest(ctx, artInfo, p.remote)
			p.Assert().Nil(err)
			p.Assert().True(exists)
			p.Assert().Equal(digest.Digest(dig), desc.Digest)
		}()
	}

	wg.Wait()

	// expect only one remote request
	p.remote.AssertExpectations(p.T())
}

func (p *proxyControllerTestSuite) TestProxyBlob_EnableAsyncLocalCaching_True() {
	p.cfg.Set(context.Background(), common.EnableAsyncLocalCaching, true)

	enabled := config.EnableAsyncLocalCaching()
	p.Assert().True(enabled)

	// expect blob to be pulled from remote
	p.local.On("PushBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	p.remote.On("BlobReader", mock.Anything, mock.Anything).Return(int64(100), io.NopCloser(bytes.NewReader([]byte("test"))), nil)

	size, bReader, err := p.ctr.ProxyBlob(
		context.Background(),
		&proModels.Project{},
		lib.ArtifactInfo{
			Repository: "library/hello-world",
			Digest:     "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b",
		},
		p.remote,
	)
	p.Assert().Nil(err)
	p.Assert().Equal(int64(100), size)
	blob, _ := io.ReadAll(bReader)
	p.Assert().Equal([]byte("test"), blob)

	// expect blob to be pushed to local in background eventually
	p.Eventually(func() bool {
		return p.local.AssertExpectations(p.T())
	}, 1*time.Second, 100*time.Millisecond, "expecting PushBlob to be called once")
}

func (p *proxyControllerTestSuite) TestProxyBlob_EnableAsyncLocalCaching_False() {
	p.cfg.Set(context.Background(), common.EnableAsyncLocalCaching, false)

	enabled := config.EnableAsyncLocalCaching()
	p.Assert().False(enabled)

	// only expect remote requests to be made once
	p.local.On("BlobExist", mock.Anything, mock.Anything).Return(false, nil)
	p.local.On("PullBlob", mock.Anything, mock.Anything).Return(int64(100), []byte("test"), nil)
	p.local.On("PushBlob", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	p.remote.On("BlobReader", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		// wait for requests to queue up
		time.Sleep(10 * time.Millisecond)
	}).Return(int64(100), io.NopCloser(bytes.NewReader([]byte("test"))), nil).Once()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			size, bReader, err := p.ctr.ProxyBlob(
				context.Background(),
				&proModels.Project{},
				lib.ArtifactInfo{
					Repository: "library/hello-world",
					Digest:     "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b",
				},
				p.remote,
			)
			p.Assert().Nil(err)
			p.Assert().Equal(int64(100), size)

			blob, _ := io.ReadAll(bReader)
			p.Assert().Equal([]byte("test"), blob)
		}()
	}

	wg.Wait()

	// expect only one remote request
	p.remote.AssertExpectations(p.T())
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
