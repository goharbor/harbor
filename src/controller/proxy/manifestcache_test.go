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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/cache"
	testproxy "github.com/goharbor/harbor/src/testing/controller/proxy"
	cachetesting "github.com/goharbor/harbor/src/testing/lib/cache"
	"github.com/goharbor/harbor/src/testing/mock"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const ociManifest = `{ 
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

type CacheTestSuite struct {
	suite.Suite
	mCache     *ManifestCache
	mListCache *ManifestListCache
	local      localInterfaceMock
	cache      *cachetesting.Cache
	remote     *testproxy.RemoteInterface
	locker     *testLocker
}

func (suite *CacheTestSuite) SetupTest() {
	suite.local = localInterfaceMock{}
	suite.cache = &cachetesting.Cache{}
	suite.remote = &testproxy.RemoteInterface{}
	suite.locker = &testLocker{}
	suite.mListCache = &ManifestListCache{local: &suite.local, cache: suite.cache, locker: suite.locker}
	suite.mCache = &ManifestCache{local: &suite.local}
}

// testLocker records the locked keys
type testLocker struct {
	mu   sync.Mutex
	keys []string
}

func (l *testLocker) lock(ctx context.Context, key string) (context.Context, func(), error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.keys = append(l.keys, key)
	return ctx, func() {}, nil
}

// lostLocker locks keys, then loses the locks at once
type lostLocker struct{}

func (lostLocker) lock(ctx context.Context, _ string) (context.Context, func(), error) {
	lockCtx, cancel := context.WithCancel(ctx)
	cancel()
	return lockCtx, func() {}, nil
}

func artifactWithDigest(dig string) *artifact.Artifact {
	a := &artifact.Artifact{}
	a.Digest = dig
	return a
}

func (suite *CacheTestSuite) TearDownSuite() {
}
func (suite *CacheTestSuite) TestUpdateManifestList() {
	ctx := context.Background()
	amdDig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	armDig := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manifestList := manifestlist.ManifestList{
		Versioned: manifest.Versioned{
			SchemaVersion: 2,
			MediaType:     manifestlist.MediaTypeManifestList,
		},
		Manifests: []manifestlist.ManifestDescriptor{
			{
				Descriptor: distribution.Descriptor{
					Digest:    digest.Digest(amdDig),
					Size:      3253,
					MediaType: schema2.MediaTypeManifest,
				},
				Platform: manifestlist.PlatformSpec{
					Architecture: "amd64",
					OS:           "linux",
				},
			}, {
				Descriptor: distribution.Descriptor{
					Digest:    digest.Digest(armDig),
					Size:      3253,
					MediaType: schema2.MediaTypeManifest,
				},
				Platform: manifestlist.PlatformSpec{
					Architecture: "arm",
					OS:           "linux",
				},
			},
		},
	}
	manList := &manifestlist.DeserializedManifestList{
		ManifestList: manifestList,
	}
	artInfo1 := lib.ArtifactInfo{
		Repository: "library/hello-world",
		Digest:     amdDig,
		Tag:        "",
	}
	ar := &artifact.Artifact{}
	suite.local.On("GetManifest", ctx, artInfo1).Return(ar, nil)
	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)

	newMan, err := suite.mListCache.updateManifestList(ctx, "library/hello-world", manList)
	suite.Require().Nil(err)
	suite.Assert().Equal(len(newMan.References()), 1)
}

const (
	testAmdDigest = "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	testArmDigest = "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
)

func testManifestList(t *testing.T) (*manifestlist.DeserializedManifestList, *manifestlist.DeserializedManifestList) {
	manList, err := manifestlist.FromDescriptors([]manifestlist.ManifestDescriptor{
		{
			Descriptor: distribution.Descriptor{Digest: digest.Digest(testAmdDigest), Size: 3253, MediaType: schema2.MediaTypeManifest},
			Platform:   manifestlist.PlatformSpec{Architecture: "amd64", OS: "linux"},
		},
		{
			Descriptor: distribution.Descriptor{Digest: digest.Digest(testArmDigest), Size: 3253, MediaType: schema2.MediaTypeManifest},
			Platform:   manifestlist.PlatformSpec{Architecture: "arm", OS: "linux"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	trimmed, err := manifestlist.FromDescriptors(manList.Manifests[:1])
	if err != nil {
		t.Fatal(err)
	}
	return manList, trimmed
}

func payloadDigest(t *testing.T, man distribution.Manifest) string {
	_, payload, err := man.Payload()
	if err != nil {
		t.Fatal(err)
	}
	return string(digest.FromBytes(payload))
}

func (suite *CacheTestSuite) TestPushManifestListByTag() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, trimmed := testManifestList(suite.T())
	dig, trimmedDig := payloadDigest(suite.T(), manList), payloadDigest(suite.T(), trimmed)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}

	suite.noPulledTags(repo, dig)
	suite.cache.On("Fetch", ctx, trimmedDigestKey(repo, dig), mock.Anything).Return(cache.ErrNotFound).Once()
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest"}).Return(nil, nil)
	suite.remote.On("ManifestExist", "hello-world", "latest").Return(true, &distribution.Descriptor{Digest: digest.Digest(dig)}, nil).Once()
	suite.local.On("PushManifest", repo, "latest", trimmed).Return(nil).Once()
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), trimmedDig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest", Digest: trimmedDig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.push(ctx, "hello-world", art, trimmed, suite.remote))
	suite.local.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	suite.remote.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushManifestListByDigest() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, trimmed := testManifestList(suite.T())
	dig, trimmedDig := payloadDigest(suite.T(), manList), payloadDigest(suite.T(), trimmed)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig}

	suite.noPulledTags(repo, dig)
	suite.cache.On("Fetch", ctx, trimmedDigestKey(repo, dig), mock.Anything).Return(cache.ErrNotFound).Once()
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: trimmedDig}).Return(nil, nil)
	suite.local.On("PushManifest", repo, trimmedDig, trimmed).Return(nil).Once()
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), trimmedDig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: trimmedDig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.push(ctx, "hello-world", art, trimmed, suite.remote))
	suite.local.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	suite.remote.AssertNotCalled(suite.T(), "ManifestExist", mock.Anything, mock.Anything)
}

func (suite *CacheTestSuite) TestPushManifestListAlreadyTagged() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, trimmed := testManifestList(suite.T())
	dig, trimmedDig := payloadDigest(suite.T(), manList), payloadDigest(suite.T(), trimmed)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}

	suite.noPulledTags(repo, dig)
	suite.cache.On("Fetch", ctx, trimmedDigestKey(repo, dig), mock.Anything).Return(cache.ErrNotFound).Once()
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest"}).Return(artifactWithDigest(trimmedDig), nil)
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), trimmedDig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest", Digest: trimmedDig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.push(ctx, "hello-world", art, trimmed, suite.remote))
	suite.local.AssertNotCalled(suite.T(), "PushManifest", mock.Anything, mock.Anything, mock.Anything)
	suite.remote.AssertNotCalled(suite.T(), "ManifestExist", mock.Anything, mock.Anything)
	suite.local.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushManifestListTagNotConfirmed() {
	other := digest.Digest("sha256:" + strings.Repeat("0", 64))
	cases := []struct {
		name  string
		exist bool
		desc  *distribution.Descriptor
		err   error
		nilR  bool
	}{
		{name: "tag moved", exist: true, desc: &distribution.Descriptor{Digest: other}},
		{name: "tag deleted"},
		{name: "check failed", err: fmt.Errorf("too many requests")},
		{name: "no digest", exist: true, desc: &distribution.Descriptor{}},
		{name: "no remote", nilR: true},
	}
	for _, tc := range cases {
		suite.Run(tc.name, func() {
			suite.SetupTest()
			ctx := context.Background()
			repo := "library/hello-world"
			manList, trimmed := testManifestList(suite.T())
			dig, trimmedDig := payloadDigest(suite.T(), manList), payloadDigest(suite.T(), trimmed)
			art := lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}
			var r RemoteInterface = suite.remote
			if tc.nilR {
				r = nil
			}

			suite.noPulledTags(repo, dig)
			suite.cache.On("Fetch", ctx, trimmedDigestKey(repo, dig), mock.Anything).Return(cache.ErrNotFound).Once()
			suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest"}).Return(artifactWithDigest(string(other)), nil)
			suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: trimmedDig}).Return(nil, nil)
			suite.remote.On("ManifestExist", "hello-world", "latest").Return(tc.exist, tc.desc, tc.err)
			suite.local.On("PushManifest", repo, trimmedDig, trimmed).Return(nil).Once()
			suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), trimmedDig, manifestListCacheInterval).Return(nil).Once()
			suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: trimmedDig}).Return(nil).Once()

			suite.Require().NoError(suite.mListCache.push(ctx, "hello-world", art, trimmed, r))
			suite.local.AssertNotCalled(suite.T(), "PushManifest", repo, "latest", mock.Anything)
			suite.local.AssertExpectations(suite.T())
			suite.cache.AssertExpectations(suite.T())
		})
	}
}

func (suite *CacheTestSuite) TestPushManifestListTagFailed() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, trimmed := testManifestList(suite.T())
	dig, trimmedDig := payloadDigest(suite.T(), manList), payloadDigest(suite.T(), trimmed)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}
	pushErr := fmt.Errorf("failed to push")

	suite.cache.On("Fetch", ctx, manifestListTagsKey(repo, dig), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]string) = []string{"stable"}
	}).Return(nil).Once()
	suite.cache.On("Fetch", ctx, trimmedDigestKey(repo, dig), mock.Anything).Return(cache.ErrNotFound).Once()
	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)
	suite.remote.On("ManifestExist", "hello-world", mock.Anything).Return(true, &distribution.Descriptor{Digest: digest.Digest(dig)}, nil)
	suite.local.On("PushManifest", repo, "latest", trimmed).Return(pushErr).Once()
	suite.local.On("PushManifest", repo, "stable", trimmed).Return(nil).Once()
	// the trimmed copy exists by the tag stable, so it is recorded
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), trimmedDig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: trimmedDig}).Return(nil).Once()

	suite.ErrorIs(suite.mListCache.push(ctx, "hello-world", art, trimmed, suite.remote), pushErr)
	suite.local.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestTagPointsToGivesUp() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	dig := digest.Digest(testAmdDigest)
	suite.remote.On("ManifestExist", "hello-world", "latest").WaitUntil(time.After(5*time.Second)).Return(true, &distribution.Descriptor{Digest: dig}, nil)

	start := time.Now()
	suite.False(tagPointsTo(ctx, suite.remote, "hello-world", "latest", string(dig)))
	suite.Less(time.Since(start), time.Second)
}

func (suite *CacheTestSuite) TestAddToListReadFailed() {
	ctx := context.Background()
	key := manifestListParentsKey("library/hello-world", testAmdDigest)
	suite.cache.On("Fetch", ctx, key, mock.Anything).Return(fmt.Errorf("%w: i/o timeout", cache.ErrNotFound)).Once()

	suite.Error(suite.mListCache.addToList(ctx, key, testArmDigest))
	suite.cache.AssertNotCalled(suite.T(), "Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func (suite *CacheTestSuite) TestPushManifestListMovesPreviousTags() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	prevDig := "sha256:" + strings.Repeat("1", 64)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig}

	suite.noPulledTags(repo, dig)
	suite.cache.On("Fetch", ctx, trimmedDigestKey(repo, dig), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*string) = prevDig
	}).Return(nil).Once()
	suite.local.On("GetTags", ctx, repo, prevDig).Return([]string{"latest", "stable"}, nil).Once()
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest"}).Return(artifactWithDigest(prevDig), nil)
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Tag: "stable"}).Return(artifactWithDigest(prevDig), nil)
	suite.remote.On("ManifestExist", "hello-world", "latest").Return(true, &distribution.Descriptor{Digest: digest.Digest(dig)}, nil).Once()
	suite.remote.On("ManifestExist", "hello-world", "stable").Return(true, &distribution.Descriptor{Digest: digest.Digest("sha256:" + strings.Repeat("0", 64))}, nil).Once()
	suite.local.On("PushManifest", repo, "latest", manList).Return(nil).Once()
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), dig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: dig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.push(ctx, "hello-world", art, manList, suite.remote))
	suite.local.AssertNotCalled(suite.T(), "PushManifest", repo, "stable", mock.Anything)
	suite.local.AssertNotCalled(suite.T(), "PushManifest", repo, dig, mock.Anything)
	suite.local.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	suite.remote.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPublishManifestListNoManifestCached() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}

	suite.waitingFor(repo, dig, "latest")
	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)

	suite.Require().NoError(suite.mListCache.publish(ctx, "hello-world", art, manList, suite.remote))
	suite.Equal([]string{manifestListLockKey(repo)}, suite.locker.keys)
	suite.local.AssertNotCalled(suite.T(), "PushManifest", mock.Anything, mock.Anything, mock.Anything)
	suite.cache.AssertExpectations(suite.T())
}

// waitingFor mocks the records of the test manifest list dig as waiting for its manifests in repo,
// and as pulled by tag unless it is empty
func (suite *CacheTestSuite) waitingFor(repo, dig, tag string) {
	for _, child := range []string{testAmdDigest, testArmDigest} {
		suite.cache.On("Fetch", mock.Anything, manifestListParentsKey(repo, child), mock.Anything).Return(cache.ErrNotFound).Once()
		suite.cache.On("Save", mock.Anything, manifestListParentsKey(repo, child), []string{dig}, manifestListCacheInterval).Return(nil).Once()
	}
	if len(tag) > 0 {
		suite.cache.On("Save", mock.Anything, manifestListTagsKey(repo, dig), []string{tag}, manifestListCacheInterval).Return(nil).Once()
	}
	suite.noPulledTags(repo, dig)
}

// noPulledTags mocks the test manifest list dig as pulled by no tag before in repo
func (suite *CacheTestSuite) noPulledTags(repo, dig string) {
	suite.cache.On("Fetch", mock.Anything, manifestListTagsKey(repo, dig), mock.Anything).Return(cache.ErrNotFound).Maybe()
}

// cachedManifests mocks the manifests of the test manifest list as cached in repo
func (suite *CacheTestSuite) cachedManifests(repo, dig, tag string) {
	suite.waitingFor(repo, dig, tag)
	suite.local.On("GetManifest", mock.Anything, lib.ArtifactInfo{Repository: repo, Digest: testAmdDigest}).Return(&artifact.Artifact{}, nil)
	suite.local.On("GetManifest", mock.Anything, lib.ArtifactInfo{Repository: repo, Digest: testArmDigest}).Return(&artifact.Artifact{}, nil)
}

func (suite *CacheTestSuite) waitAndPushAllCached(remoteDigest string) (lib.ArtifactInfo, string) {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}
	suite.remote.On("ManifestExist", "hello-world", "latest").Return(true, &distribution.Descriptor{Digest: digest.Digest(remoteDigest)}, nil).Once()
	suite.cachedManifests(repo, dig, "latest")
	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)
	suite.cache.On("Fetch", ctx, trimmedDigestKey(repo, dig), mock.Anything).Return(cache.ErrNotFound).Once()
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), dig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, mock.Anything).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.waitAndPush(ctx, "hello-world", art, manList, suite.remote, maxManifestListWait))
	suite.Equal([]string{manifestListLockKey(repo)}, suite.locker.keys)
	suite.remote.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	return art, dig
}

func (suite *CacheTestSuite) TestWaitAndPushManifestList() {
	manList, _ := testManifestList(suite.T())
	suite.local.On("PushManifest", "library/hello-world", "latest", mock.Anything).Return(nil).Once()
	art, _ := suite.waitAndPushAllCached(payloadDigest(suite.T(), manList))
	suite.local.AssertCalled(suite.T(), "UpdatePullTime", mock.Anything, lib.ArtifactInfo{Repository: art.Repository, Tag: "latest", Digest: art.Digest})
	suite.local.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestWaitAndPushManifestListTagMoved() {
	dig := "sha256:" + strings.Repeat("0", 64)
	suite.local.On("PushManifest", "library/hello-world", mock.Anything, mock.Anything).Return(nil).Once()
	art, _ := suite.waitAndPushAllCached(dig)
	suite.local.AssertCalled(suite.T(), "PushManifest", art.Repository, art.Digest, mock.Anything)
	suite.local.AssertNotCalled(suite.T(), "PushManifest", art.Repository, "latest", mock.Anything)
	suite.local.AssertCalled(suite.T(), "UpdatePullTime", mock.Anything, lib.ArtifactInfo{Repository: art.Repository, Digest: art.Digest})
}

func (suite *CacheTestSuite) TestPushParents() {
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	ct, payload, err := manList.Payload()
	suite.Require().NoError(err)
	parent := lib.ArtifactInfo{Repository: repo, Digest: dig}
	prevDig := "sha256:" + strings.Repeat("1", 64)

	suite.cache.On("Fetch", mock.Anything, manifestListParentsKey(repo, testArmDigest), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]string) = []string{dig}
	}).Return(nil).Once()
	suite.cache.On("Fetch", mock.Anything, manifestListKey(repo, parent), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]byte) = payload
	}).Return(nil).Once()
	suite.cache.On("Fetch", mock.Anything, manifestListContentTypeKey(repo, parent), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*string) = ct
	}).Return(nil).Once()
	suite.cachedManifests(repo, dig, "")
	suite.cache.On("Fetch", mock.Anything, trimmedDigestKey(repo, dig), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*string) = prevDig
	}).Return(nil).Once()
	suite.local.On("GetTags", mock.Anything, repo, prevDig).Return([]string{"latest"}, nil).Once()
	suite.local.On("GetManifest", mock.Anything, lib.ArtifactInfo{Repository: repo, Tag: "latest"}).Return(artifactWithDigest(prevDig), nil)
	suite.remote.On("ManifestExist", "hello-world", "latest").Return(true, &distribution.Descriptor{Digest: digest.Digest(dig)}, nil).Once()
	suite.local.On("PushManifest", repo, "latest", mock.Anything).Return(nil).Once()
	suite.cache.On("Save", mock.Anything, trimmedDigestKey(repo, dig), dig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", mock.Anything, lib.ArtifactInfo{Repository: repo, Digest: dig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.pushParents(context.Background(), "hello-world", repo, testArmDigest, suite.remote))
	suite.local.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	suite.remote.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushParentsNotWaiting() {
	repo := "library/hello-world"
	suite.cache.On("Fetch", mock.Anything, manifestListParentsKey(repo, testArmDigest), mock.Anything).Return(cache.ErrNotFound).Once()

	suite.Require().NoError(suite.mListCache.pushParents(context.Background(), "hello-world", repo, testArmDigest, suite.remote))
	suite.Empty(suite.locker.keys)
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushParentsWhilePublishing() {
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	ct, payload, err := manList.Payload()
	suite.Require().NoError(err)
	parent := lib.ArtifactInfo{Repository: repo, Digest: dig}
	// this core instance publishes the manifest list already
	key := manifestListKey(repo, parent) + ":"
	suite.Require().True(inflightChecker.addRequest(key))
	defer inflightChecker.removeRequest(key)

	suite.cache.On("Fetch", mock.Anything, manifestListParentsKey(repo, testArmDigest), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]string) = []string{dig}
	}).Return(nil).Once()
	suite.cache.On("Fetch", mock.Anything, manifestListKey(repo, parent), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]byte) = payload
	}).Return(nil).Once()
	suite.cache.On("Fetch", mock.Anything, manifestListContentTypeKey(repo, parent), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*string) = ct
	}).Return(nil).Once()
	suite.waitingFor(repo, dig, "")
	suite.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)

	suite.Require().NoError(suite.mListCache.pushParents(context.Background(), "hello-world", repo, testArmDigest, suite.remote))
	suite.Equal([]string{manifestListLockKey(repo)}, suite.locker.keys)
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushParentsByPulledTag() {
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	ct, payload, err := manList.Payload()
	suite.Require().NoError(err)
	parent := lib.ArtifactInfo{Repository: repo, Digest: dig}

	suite.cache.On("Fetch", mock.Anything, manifestListParentsKey(repo, testArmDigest), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]string) = []string{dig}
	}).Return(nil).Once()
	suite.cache.On("Fetch", mock.Anything, manifestListKey(repo, parent), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]byte) = payload
	}).Return(nil).Once()
	suite.cache.On("Fetch", mock.Anything, manifestListContentTypeKey(repo, parent), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*string) = ct
	}).Return(nil).Once()
	for _, child := range []string{testAmdDigest, testArmDigest} {
		suite.cache.On("Fetch", mock.Anything, manifestListParentsKey(repo, child), mock.Anything).Return(cache.ErrNotFound).Once()
		suite.cache.On("Save", mock.Anything, manifestListParentsKey(repo, child), []string{dig}, manifestListCacheInterval).Return(nil).Once()
	}
	suite.local.On("GetManifest", mock.Anything, lib.ArtifactInfo{Repository: repo, Digest: testAmdDigest}).Return(&artifact.Artifact{}, nil)
	suite.local.On("GetManifest", mock.Anything, lib.ArtifactInfo{Repository: repo, Digest: testArmDigest}).Return(&artifact.Artifact{}, nil)
	// the manifest list was pulled by tag while none of its manifests was cached
	suite.cache.On("Fetch", mock.Anything, manifestListTagsKey(repo, dig), mock.Anything).Run(func(args mock.Arguments) {
		*args.Get(2).(*[]string) = []string{"latest"}
	}).Return(nil).Once()
	suite.cache.On("Fetch", mock.Anything, trimmedDigestKey(repo, dig), mock.Anything).Return(cache.ErrNotFound).Once()
	suite.local.On("GetManifest", mock.Anything, lib.ArtifactInfo{Repository: repo, Tag: "latest"}).Return(nil, nil)
	suite.remote.On("ManifestExist", "hello-world", "latest").Return(true, &distribution.Descriptor{Digest: digest.Digest(dig)}, nil).Once()
	suite.local.On("PushManifest", repo, "latest", mock.Anything).Return(nil).Once()
	suite.cache.On("Save", mock.Anything, trimmedDigestKey(repo, dig), dig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", mock.Anything, lib.ArtifactInfo{Repository: repo, Digest: dig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.pushParents(context.Background(), "hello-world", repo, testArmDigest, suite.remote))
	suite.local.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
	suite.remote.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestWaitForManifestsKeepsOtherManifestLists() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	others := make([]string, maxListLength)
	for i := range others {
		others[i] = fmt.Sprintf("sha256:%064d", i)
	}
	for _, child := range []string{testAmdDigest, testArmDigest} {
		suite.cache.On("Fetch", ctx, manifestListParentsKey(repo, child), mock.Anything).Run(func(args mock.Arguments) {
			*args.Get(2).(*[]string) = append([]string{dig}, others...)
		}).Return(nil).Once()
		suite.cache.On("Save", ctx, manifestListParentsKey(repo, child), append(others[1:], dig), manifestListCacheInterval).Return(nil).Once()
	}

	suite.mListCache.waitForManifests(ctx, lib.ArtifactInfo{Repository: repo, Digest: dig}, manList)
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPublishManifestListLockLost() {
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	suite.mListCache.locker = lostLocker{}
	suite.cachedManifests(repo, dig, "latest")
	suite.cache.On("Fetch", mock.Anything, trimmedDigestKey(repo, dig), mock.Anything).Return(cache.ErrNotFound)
	suite.local.On("GetManifest", mock.Anything, mock.Anything).Return(nil, nil)
	suite.remote.On("ManifestExist", "hello-world", "latest").Return(true, &distribution.Descriptor{Digest: digest.Digest(dig)}, nil)

	err := suite.mListCache.publish(context.Background(), "hello-world", lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}, manList, suite.remote)
	suite.ErrorIs(err, context.Canceled)
	suite.local.AssertNotCalled(suite.T(), "PushManifest", mock.Anything, mock.Anything, mock.Anything)
	suite.cache.AssertNotCalled(suite.T(), "Save", mock.Anything, trimmedDigestKey(repo, dig), mock.Anything, mock.Anything)
}

func (suite *CacheTestSuite) TestManifestCache_CacheContent() {
	defer suite.local.AssertExpectations(suite.T())

	manifest := ociManifest
	man, desc, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifest))
	suite.Require().NoError(err)

	ctx := context.Background()
	repo := "library/hello-world"

	artInfo := lib.ArtifactInfo{
		Repository: repo,
		Digest:     string(desc.Digest),
		Tag:        "latest",
	}

	suite.local.On("CheckDependencies", ctx, artInfo.Repository, man).Once().Return([]distribution.Descriptor{})
	suite.local.On("PushManifest", artInfo.Repository, artInfo.Digest, man).Once().Return(nil)
	suite.local.On("PushManifest", artInfo.Repository, artInfo.Tag, man).Once().Return(nil)
	suite.mCache.lists = suite.mListCache
	suite.cache.On("Fetch", ctx, manifestListParentsKey(repo, artInfo.Digest), mock.Anything).Once().Return(cache.ErrNotFound)

	suite.mCache.CacheContent(ctx, repo, man, artInfo, nil, "")
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestManifestCache_push_succeeds() {
	defer suite.local.AssertExpectations(suite.T())

	manifest := ociManifest
	man, desc, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifest))
	suite.Require().NoError(err)

	repo := "library/hello-world"

	artInfo := lib.ArtifactInfo{
		Repository: repo,
		Digest:     string(desc.Digest),
		Tag:        "latest",
	}

	suite.local.On("PushManifest", artInfo.Repository, artInfo.Digest, man).Once().Return(nil)
	suite.local.On("PushManifest", artInfo.Repository, artInfo.Tag, man).Once().Return(nil)

	err = suite.mCache.push(artInfo, man)
	suite.Assert().NoError(err)
}

func (suite *CacheTestSuite) TestManifestCache_push_fails() {
	defer suite.local.AssertExpectations(suite.T())

	manifest := ociManifest
	man, desc, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(manifest))
	suite.Require().NoError(err)

	repo := "library/hello-world"

	artInfo := lib.ArtifactInfo{
		Repository: repo,
		Digest:     string(desc.Digest),
		Tag:        "latest",
	}

	digestErr := fmt.Errorf("error during manifest push referencing digest")
	tagErr := fmt.Errorf("error during manifest push referencing tag")
	suite.local.On("PushManifest", artInfo.Repository, artInfo.Digest, man).Once().Return(digestErr)
	suite.local.On("PushManifest", artInfo.Repository, artInfo.Tag, man).Once().Return(tagErr)

	err = suite.mCache.push(artInfo, man)
	suite.Assert().Error(err)
	wrappedErr, isWrappedErr := err.(interface{ Unwrap() []error })
	suite.Assert().True(isWrappedErr)
	errs := wrappedErr.Unwrap()
	suite.Assert().Len(errs, 2)
	suite.Assert().Contains(errs, digestErr)
	suite.Assert().Contains(errs, tagErr)
}

func TestCacheTestSuite(t *testing.T) {
	suite.Run(t, &CacheTestSuite{})
}
