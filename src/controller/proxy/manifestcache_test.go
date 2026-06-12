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
	"testing"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	libCache "github.com/goharbor/harbor/src/lib/cache"
	cacheMemory "github.com/goharbor/harbor/src/lib/cache/memory"
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
}

func buildTestManifestList(descriptors ...manifestlist.ManifestDescriptor) *manifestlist.DeserializedManifestList {
	man, err := manifestlist.FromDescriptors(descriptors)
	if err != nil {
		panic(err)
	}
	return man
}

func (suite *CacheTestSuite) SetupTest() {
	originManifestListWait := maxManifestListWait
	originManifestWait := maxManifestWait
	originSleepInterval := sleepIntervalSec
	maxManifestListWait = 1
	maxManifestWait = 1
	sleepIntervalSec = 0
	suite.T().Cleanup(func() {
		maxManifestListWait = originManifestListWait
		maxManifestWait = originManifestWait
		sleepIntervalSec = originSleepInterval
	})
	suite.local = localInterfaceMock{}
	cache, err := cacheMemory.New(libCache.Options{Prefix: "cache:", Codec: libCache.DefaultCodec()})
	suite.Require().NoError(err)
	suite.mListCache = &ManifestListCache{local: &suite.local, cache: cache}
	suite.mCache = &ManifestCache{local: &suite.local, manifestListCache: suite.mListCache}
}

func (suite *CacheTestSuite) TestUpdateManifestList() {
	ctx := context.Background()
	amdDig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	armDig := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manifestDescriptors := []manifestlist.ManifestDescriptor{
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
	}
	manList := buildTestManifestList(manifestDescriptors...)
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

func (suite *CacheTestSuite) TestPushManifestList() {
	ctx := context.Background()
	amdDig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	armDig := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manifestDescriptors := []manifestlist.ManifestDescriptor{
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
	}
	manList := buildTestManifestList(manifestDescriptors...)
	repo := "library/hello-world"
	artInfo1 := lib.ArtifactInfo{
		Repository: repo,
		Digest:     amdDig,
		Tag:        "",
	}
	ar := &artifact.Artifact{}
	_, payload, err := manList.Payload()
	suite.Nil(err)
	originDigest := digest.FromBytes(payload)
	trimmedMan, err := manifestlist.FromDescriptors([]manifestlist.ManifestDescriptor{manifestDescriptors[0]})
	suite.Require().NoError(err)
	_, trimmedPayload, err := trimmedMan.Payload()
	suite.Require().NoError(err)
	trimmedDigest := string(digest.FromBytes(trimmedPayload))

	suite.local.On("GetManifest", ctx, artInfo1).Return(ar, nil)
	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)

	suite.local.On("PushManifest", repo, trimmedDigest, mock.Anything).Return(nil)
	suite.local.On("PushManifest", repo, "latest", mock.Anything).Return(nil)
	suite.local.On("UpdatePullTime", ctx, mock.Anything).Return(nil)

	err = suite.mListCache.push(ctx, lib.ArtifactInfo{Repository: repo, Digest: string(originDigest), Tag: "latest"}, manList, manifestlist.MediaTypeManifestList)
	suite.Require().Nil(err)
}

func (suite *CacheTestSuite) TestPushManifestList_RegistersPendingWhenChildrenMissing() {
	ctx := context.Background()
	repo := "library/hello-world"
	amdDig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	armDig := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manifestDescriptors := []manifestlist.ManifestDescriptor{{
		Descriptor: distribution.Descriptor{Digest: digest.Digest(amdDig), Size: 3253, MediaType: schema2.MediaTypeManifest},
		Platform:   manifestlist.PlatformSpec{Architecture: "amd64", OS: "linux"},
	}, {
		Descriptor: distribution.Descriptor{Digest: digest.Digest(armDig), Size: 3253, MediaType: schema2.MediaTypeManifest},
		Platform:   manifestlist.PlatformSpec{Architecture: "arm64", OS: "linux"},
	}}
	manList := buildTestManifestList(manifestDescriptors...)
	_, payload, err := manList.Payload()
	suite.Require().NoError(err)
	parentDigest := string(digest.FromBytes(payload))

	suite.local.On("GetManifest", ctx, mock.Anything).Return(nil, nil)

	err = suite.mListCache.push(ctx, lib.ArtifactInfo{Repository: repo, Digest: parentDigest, Tag: "latest"}, manList, manifestlist.MediaTypeManifestList)
	suite.Require().NoError(err)

	for _, childDigest := range []string{amdDig, armDig} {
		var state pendingManifestListState
		err = suite.mListCache.cache.Fetch(ctx, pendingManifestListKey(repo, childDigest, parentDigest), &state)
		suite.Require().NoError(err)
		suite.Equal(repo, state.Repository)
		suite.Equal(parentDigest, state.Digest)
		suite.Equal("latest", state.Tag)
	}
	trimmedKey := TrimmedManifestlist + parentDigest
	err = suite.mListCache.cache.Fetch(ctx, trimmedKey, new(string))
	suite.ErrorIs(err, libCache.ErrNotFound)
}

func (suite *CacheTestSuite) TestManifestCache_ReconcilesPendingManifestList() {
	ctx := context.Background()
	repo := "library/hello-world"
	cache, err := cacheMemory.New(libCache.Options{Prefix: "cache:", Codec: libCache.DefaultCodec()})
	suite.Require().NoError(err)
	local := &localInterfaceMock{}
	mListCache := &ManifestListCache{local: local, cache: cache}
	mCache := &ManifestCache{local: local, manifestListCache: mListCache}
	childManifest, childDesc, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(ociManifest))
	suite.Require().NoError(err)
	amdDig := string(childDesc.Digest)
	armDig := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manifestDescriptors := []manifestlist.ManifestDescriptor{{
		Descriptor: distribution.Descriptor{Digest: digest.Digest(amdDig), Size: 3253, MediaType: schema2.MediaTypeManifest},
		Platform:   manifestlist.PlatformSpec{Architecture: "amd64", OS: "linux"},
	}, {
		Descriptor: distribution.Descriptor{Digest: digest.Digest(armDig), Size: 3253, MediaType: schema2.MediaTypeManifest},
		Platform:   manifestlist.PlatformSpec{Architecture: "arm64", OS: "linux"},
	}}
	manList := buildTestManifestList(manifestDescriptors...)
	_, listPayload, err := manList.Payload()
	suite.Require().NoError(err)
	parentDigest := string(digest.FromBytes(listPayload))

	childArt := lib.ArtifactInfo{Repository: repo, Digest: string(childDesc.Digest), Tag: "linux-amd64"}

	local.On("GetManifest", ctx, mock.Anything).Return(nil, nil).Times(len(manifestDescriptors) * maxManifestListWait)
	err = mListCache.push(ctx, lib.ArtifactInfo{Repository: repo, Digest: parentDigest, Tag: "latest"}, manList, manifestlist.MediaTypeManifestList)
	suite.Require().NoError(err)

	fullArt := &artifact.Artifact{}
	local.ExpectedCalls = nil
	local.Calls = nil
	local.On("CheckDependencies", ctx, repo, childManifest).Return([]distribution.Descriptor{}).Once()
	local.On("PushManifest", repo, childArt.Digest, childManifest).Return(nil).Once()
	local.On("PushManifest", repo, childArt.Tag, childManifest).Return(nil).Once()
	local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: amdDig}).Return(fullArt, nil).Maybe()
	local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: armDig}).Return(nil, nil).Maybe()
	trimmedMan, err := manifestlist.FromDescriptors([]manifestlist.ManifestDescriptor{manifestDescriptors[0]})
	suite.Require().NoError(err)
	_, trimmedPayload, err := trimmedMan.Payload()
	suite.Require().NoError(err)
	trimmedDigest := string(digest.FromBytes(trimmedPayload))
	local.On("PushManifest", repo, trimmedDigest, mock.Anything).Return(nil).Once()
	local.On("PushManifest", repo, "latest", mock.Anything).Return(nil).Once()
	local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: trimmedDigest, Tag: "latest"}).Return(nil).Once()

	mCache.CacheContent(ctx, repo, childManifest, childArt, nil, "")

	var cachedTrimmed string
	err = cache.Fetch(ctx, TrimmedManifestlist+parentDigest, &cachedTrimmed)
	suite.Require().NoError(err)
	suite.Equal(trimmedDigest, cachedTrimmed)

	for _, childDigest := range []string{amdDig, armDig} {
		var state pendingManifestListState
		err = cache.Fetch(ctx, pendingManifestListKey(repo, childDigest, parentDigest), &state)
		suite.Require().NoError(err)
		suite.Equal(parentDigest, state.Digest)
		suite.Equal("latest", state.Tag)
	}

	local.AssertExpectations(suite.T())
}

func (suite *CacheTestSuite) TestPushManifestList_DeletesPreviousTrimmedDigest() {
	ctx := context.Background()
	repo := "library/hello-world"
	firstDig := "sha256:1a9ec845ee94c202b2d5da74a24f0ed2058318bfa9879fa541efaecba272e86b"
	secondDig := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manList := buildTestManifestList(
		manifestlist.ManifestDescriptor{Descriptor: distribution.Descriptor{Digest: digest.Digest(firstDig), Size: 1, MediaType: schema2.MediaTypeManifest}, Platform: manifestlist.PlatformSpec{Architecture: "amd64", OS: "linux"}},
		manifestlist.ManifestDescriptor{Descriptor: distribution.Descriptor{Digest: digest.Digest(secondDig), Size: 1, MediaType: schema2.MediaTypeManifest}, Platform: manifestlist.PlatformSpec{Architecture: "arm64", OS: "linux"}},
	)
	_, payload, err := manList.Payload()
	suite.Require().NoError(err)
	parentDigest := string(digest.FromBytes(payload))
	firstTrimmed, err := manifestlist.FromDescriptors([]manifestlist.ManifestDescriptor{manList.Manifests[0]})
	suite.Require().NoError(err)
	_, firstPayload, err := firstTrimmed.Payload()
	suite.Require().NoError(err)
	firstTrimmedDigest := string(digest.FromBytes(firstPayload))

	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: firstDig}).Return(&artifact.Artifact{}, nil).Once()
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: secondDig}).Return(nil, nil).Once()
	suite.local.On("PushManifest", repo, firstTrimmedDigest, mock.Anything).Return(nil).Once()
	suite.local.On("PushManifest", repo, "latest", mock.Anything).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: firstTrimmedDigest, Tag: "latest"}).Return(nil).Once()
	err = suite.mListCache.push(ctx, lib.ArtifactInfo{Repository: repo, Digest: parentDigest, Tag: "latest"}, manList, manifestlist.MediaTypeManifestList)
	suite.Require().NoError(err)

	trimmedDigest, err := suite.mListCache.getTrimmedDigest(ctx, parentDigest)
	suite.Require().NoError(err)
	suite.Equal(firstTrimmedDigest, trimmedDigest)

	fullArt := &artifact.Artifact{}
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: firstDig}).Return(fullArt, nil).Once()
	suite.local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: secondDig}).Return(fullArt, nil).Once()
	suite.local.On("PushManifest", repo, parentDigest, mock.Anything).Return(nil).Once()
	suite.local.On("PushManifest", repo, "latest", mock.Anything).Return(nil).Once()
	suite.local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: parentDigest, Tag: "latest"}).Return(nil).Once()
	suite.local.On("DeleteManifest", repo, firstTrimmedDigest).Once()
	err = suite.mListCache.push(ctx, lib.ArtifactInfo{Repository: repo, Digest: parentDigest, Tag: "latest"}, manList, manifestlist.MediaTypeManifestList)
	suite.Require().NoError(err)

	trimmedDigest, err = suite.mListCache.getTrimmedDigest(ctx, parentDigest)
	suite.Require().NoError(err)
	suite.Equal(parentDigest, trimmedDigest)
}

func (suite *CacheTestSuite) TestManifestCache_ReconcilesPendingWhenDigestPushSucceedsButTagPushFails() {
	ctx := context.Background()
	repo := "library/hello-world"
	cache, err := cacheMemory.New(libCache.Options{Prefix: "cache:", Codec: libCache.DefaultCodec()})
	suite.Require().NoError(err)
	local := &localInterfaceMock{}
	mListCache := &ManifestListCache{local: local, cache: cache}
	mCache := &ManifestCache{local: local, manifestListCache: mListCache}
	childManifest, childDesc, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(ociManifest))
	suite.Require().NoError(err)
	childDigest := string(childDesc.Digest)
	otherDigest := "sha256:92c7f9c92844bbbb5d0a101b22f7c2a7949e40f8ea90c8b3bc396879d95e899a"
	manList := buildTestManifestList(
		manifestlist.ManifestDescriptor{Descriptor: distribution.Descriptor{Digest: digest.Digest(childDigest), Size: 1, MediaType: schema2.MediaTypeManifest}, Platform: manifestlist.PlatformSpec{Architecture: "amd64", OS: "linux"}},
		manifestlist.ManifestDescriptor{Descriptor: distribution.Descriptor{Digest: digest.Digest(otherDigest), Size: 1, MediaType: schema2.MediaTypeManifest}, Platform: manifestlist.PlatformSpec{Architecture: "arm64", OS: "linux"}},
	)
	_, listPayload, err := manList.Payload()
	suite.Require().NoError(err)
	parentDigest := string(digest.FromBytes(listPayload))

	local.On("GetManifest", ctx, mock.Anything).Return(nil, nil).Times(2)
	err = mListCache.push(ctx, lib.ArtifactInfo{Repository: repo, Digest: parentDigest, Tag: "latest"}, manList, manifestlist.MediaTypeManifestList)
	suite.Require().NoError(err)

	fullArt := &artifact.Artifact{}
	local.ExpectedCalls = nil
	local.Calls = nil
	local.On("CheckDependencies", ctx, repo, childManifest).Return([]distribution.Descriptor{}).Once()
	local.On("PushManifest", repo, childDigest, childManifest).Return(nil).Once()
	local.On("PushManifest", repo, "linux-amd64", childManifest).Return(fmt.Errorf("tag push failed")).Once()
	local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: childDigest}).Return(fullArt, nil).Maybe()
	local.On("GetManifest", ctx, lib.ArtifactInfo{Repository: repo, Digest: otherDigest}).Return(nil, nil).Maybe()
	trimmedMan, err := manifestlist.FromDescriptors([]manifestlist.ManifestDescriptor{manList.Manifests[0]})
	suite.Require().NoError(err)
	_, trimmedPayload, err := trimmedMan.Payload()
	suite.Require().NoError(err)
	trimmedDigest := string(digest.FromBytes(trimmedPayload))
	local.On("PushManifest", repo, trimmedDigest, mock.Anything).Return(nil).Once()
	local.On("PushManifest", repo, "latest", mock.Anything).Return(nil).Once()
	local.On("UpdatePullTime", ctx, lib.ArtifactInfo{Repository: repo, Digest: trimmedDigest, Tag: "latest"}).Return(nil).Once()

	mCache.CacheContent(ctx, repo, childManifest, lib.ArtifactInfo{Repository: repo, Digest: childDigest, Tag: "linux-amd64"}, nil, "")

	var cachedTrimmed string
	err = cache.Fetch(ctx, TrimmedManifestlist+parentDigest, &cachedTrimmed)
	suite.Require().NoError(err)
	suite.Equal(trimmedDigest, cachedTrimmed)
	local.AssertExpectations(suite.T())
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

	_, err = suite.mCache.push(artInfo, man)
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

	_, err = suite.mCache.push(artInfo, man)
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
