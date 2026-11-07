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
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
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
}

func (suite *CacheTestSuite) SetupTest() {
	suite.local = localInterfaceMock{}
	suite.cache = &cachetesting.Cache{}
	suite.mListCache = &ManifestListCache{local: &suite.local, cache: suite.cache}
	suite.mCache = &ManifestCache{local: &suite.local}
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

	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest"}).Return(nil, nil)
	suite.cache.On("Save", ctx, TrimmedManifestlist+dig, trimmedDig).Return(nil).Once()
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), trimmedDig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("PushManifest", repo, "latest", trimmed).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest", Digest: trimmedDig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.push(ctx, art, trimmed))
	suite.local.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushManifestListByDigest() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, trimmed := testManifestList(suite.T())
	dig, trimmedDig := payloadDigest(suite.T(), manList), payloadDigest(suite.T(), trimmed)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig}

	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Tag: trimmedDig}).Return(nil, nil)
	suite.cache.On("Save", ctx, TrimmedManifestlist+dig, trimmedDig).Return(nil).Once()
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), trimmedDig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("PushManifest", repo, trimmedDig, trimmed).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: trimmedDig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.push(ctx, art, trimmed))
	suite.local.AssertExpectations(suite.T())
	suite.cache.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushManifestListAlreadyTagged() {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, trimmed := testManifestList(suite.T())
	dig, trimmedDig := payloadDigest(suite.T(), manList), payloadDigest(suite.T(), trimmed)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}

	tagged := &artifact.Artifact{}
	tagged.Digest = trimmedDig
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest"}).Return(tagged, nil)
	suite.cache.On("Save", ctx, TrimmedManifestlist+dig, trimmedDig).Return(nil).Once()
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), trimmedDig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Tag: "latest", Digest: trimmedDig}).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.push(ctx, art, trimmed))
	suite.local.AssertNotCalled(suite.T(), "PushManifest", mock.Anything, mock.Anything, mock.Anything)
	suite.local.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushManifestListNoManifestCached() {
	manList, _ := testManifestList(suite.T())
	empty, err := manifestlist.FromDescriptors(nil)
	suite.Require().NoError(err)
	art := lib.ArtifactInfo{Repository: "library/hello-world", Digest: payloadDigest(suite.T(), manList), Tag: "latest"}

	suite.Error(suite.mListCache.push(context.Background(), art, empty))
	suite.local.AssertNotCalled(suite.T(), "PushManifest", mock.Anything, mock.Anything, mock.Anything)
	suite.cache.AssertNotCalled(suite.T(), "Save", mock.Anything, mock.Anything, mock.Anything)
}

func (suite *CacheTestSuite) waitAndPushAllCached(remoteDigest string) (lib.ArtifactInfo, string) {
	ctx := context.Background()
	repo := "library/hello-world"
	manList, _ := testManifestList(suite.T())
	dig := payloadDigest(suite.T(), manList)
	art := lib.ArtifactInfo{Repository: repo, Digest: dig, Tag: "latest"}
	remote := &testproxy.RemoteInterface{}
	remote.On("ManifestExist", "hello-world", "latest").Return(true, &distribution.Descriptor{Digest: digest.Digest(remoteDigest)}, nil).Once()
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: testAmdDigest}).Return(&artifact.Artifact{}, nil)
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: testArmDigest}).Return(&artifact.Artifact{}, nil)
	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)
	suite.cache.On("Save", ctx, TrimmedManifestlist+dig, dig).Return(nil).Once()
	suite.cache.On("Save", ctx, trimmedDigestKey(repo, dig), dig, manifestListCacheInterval).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, mock.Anything).Return(nil).Once()

	suite.Require().NoError(suite.mListCache.waitAndPush(ctx, "hello-world", art, manList, remote))
	remote.AssertExpectations(suite.T())
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

	suite.mCache.CacheContent(ctx, repo, man, artInfo, nil, "")
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
