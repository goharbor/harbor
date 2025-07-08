// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	common_http "github.com/goharbor/harbor/src/common/http"
	trans "github.com/goharbor/harbor/src/controller/replication/transfer"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

var (
	blobRetryCnt, chunkRetryCnt int
	replicationChunkSize        int64
	errStopped                  = errors.New("stopped")
	// default chunk size is 10MB
	defaultChunkSize = 10 * 1024 * 1024
)

func init() {
	blobRetryCnt, _ = strconv.Atoi(os.Getenv("COPY_BLOB_RETRY_COUNT"))
	if blobRetryCnt <= 0 {
		blobRetryCnt = 5
	}

	chunkRetryCnt, _ = strconv.Atoi(os.Getenv("COPY_CHUNK_RETRY_COUNT"))
	if chunkRetryCnt <= 0 {
		chunkRetryCnt = 5
	}

	replicationChunkSize, _ = strconv.ParseInt(os.Getenv("REPLICATION_CHUNK_SIZE"), 10, 64)
	if replicationChunkSize <= 0 {
		replicationChunkSize = int64(defaultChunkSize)
	}

	if err := trans.RegisterFactory(model.ResourceTypeImage, factory); err != nil {
		log.Errorf("failed to register transfer factory: %v", err)
	}
	if err := trans.RegisterFactory(model.ResourceTypeArtifact, factory); err != nil {
		log.Errorf("failed to register transfer factory: %v", err)
	}
}

type repository struct {
	repository string
	tags       []string
}

func factory(logger trans.Logger, stopFunc trans.StopFunc) (trans.Transfer, error) {
	return &transfer{
		logger:    logger,
		isStopped: stopFunc,
	}, nil
}

type transfer struct {
	logger    trans.Logger
	isStopped trans.StopFunc
	src       adapter.ArtifactRegistry
	dst       adapter.ArtifactRegistry
}

func (t *transfer) Transfer(src *model.Resource, dst *model.Resource, opts *trans.Options) error {
	// initialize
	if err := t.initialize(src, dst); err != nil {
		return err
	}

	// delete the artifacts/tags on the destination registry
	if dst.Deleted {
		// delete tag
		if dst.IsDeleteTag {
			return t.deleteTag(dst)
		}
		// delete artifact
		return t.delete(t.convert(dst))
	}

	// copy the repository from source registry to the destination
	return t.copy(t.convert(src), t.convert(dst), dst.Override, opts)
}

func (t *transfer) convert(resource *model.Resource) *repository {
	repository := &repository{
		repository: resource.Metadata.Repository.Name,
	}
	for _, artifact := range resource.Metadata.Artifacts {
		if len(artifact.Tags) > 0 {
			repository.tags = append(repository.tags, artifact.Tags...)
			continue
		}
		// no tags
		if len(artifact.Digest) > 0 {
			repository.tags = append(repository.tags, artifact.Digest)
		}
	}
	if len(repository.tags) > 0 {
		return repository
	}
	// fallback to vtags
	repository.tags = resource.Metadata.Vtags
	return repository
}

func (t *transfer) initialize(src *model.Resource, dst *model.Resource) error {
	// create client for source registry
	srcReg, err := createRegistry(src.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for source registry: %v", err)
		return err
	}
	t.src = srcReg
	t.logger.Infof("client for source registry [type: %s, URL: %s, insecure: %v] created",
		src.Registry.Type, src.Registry.URL, src.Registry.Insecure)

	// create client for destination registry
	dstReg, err := createRegistry(dst.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for destination registry: %v", err)
		return err
	}
	t.dst = dstReg
	t.logger.Infof("client for destination registry [type: %s, URL: %s, insecure: %v] created",
		dst.Registry.Type, dst.Registry.URL, dst.Registry.Insecure)

	return nil
}

func createRegistry(reg *model.Registry) (adapter.ArtifactRegistry, error) {
	factory, err := adapter.GetFactory(reg.Type)
	if err != nil {
		return nil, err
	}
	ad, err := factory.Create(reg)
	if err != nil {
		return nil, err
	}
	registry, ok := ad.(adapter.ArtifactRegistry)
	if !ok {
		return nil, errors.New("the adapter doesn't implement the \"ArtifactRegistry\" interface")
	}
	return registry, nil
}

func (t *transfer) shouldStop() bool {
	isStopped := t.isStopped()
	if isStopped {
		t.logger.Info("the job is stopped")
	}
	return isStopped
}

func (t *transfer) copy(src *repository, dst *repository, override bool, opts *trans.Options) error {
	srcRepo := src.repository
	dstRepo := dst.repository
	t.logger.Infof("copying %s:[%s](source registry) to %s:[%s](destination registry)...",
		srcRepo, strings.Join(src.tags, ","), dstRepo, strings.Join(dst.tags, ","))
	if opts.Speed > 0 {
		t.logger.Infof("limit network speed at %d kb/s", opts.Speed)
	}

	var err error
	for i := range src.tags {
		if e := t.copyArtifact(srcRepo, src.tags[i], dstRepo, dst.tags[i], override, opts); e != nil {
			if e == errStopped {
				return nil
			}
			t.logger.Errorf(e.Error())
			err = e
		}
	}
	if err != nil {
		err = errors.New("got error during the whole transfer period, mark the job failure")
		t.logger.Error(err)
		return err
	}

	t.logger.Infof("copy %s:[%s](source registry) to %s:[%s](destination registry) completed",
		srcRepo, strings.Join(src.tags, ","), dstRepo, strings.Join(dst.tags, ","))
	return nil
}

func (t *transfer) copyArtifact(srcRepo, srcRef, dstRepo, dstRef string, override bool, opts *trans.Options) error {
	t.logger.Infof("copying %s:%s(source registry) to %s:%s(destination registry)...",
		srcRepo, srcRef, dstRepo, dstRef)
	// pull the manifest from the source registry
	manifest, digest, err := t.pullManifest(srcRepo, srcRef)
	if err != nil {
		return err
	}

	// check the existence of the artifact on the destination registry
	exist, digest2, err := t.exist(dstRepo, dstRef)
	if err != nil {
		return err
	}
	if exist {
		// the same artifact already exists
		if digest == digest2 {
			t.logger.Infof("the artifact %s:%s already exists on the destination registry, skip",
				dstRepo, dstRef)
			return nil
		}
		// the same name artifact exists, but not allowed to override
		if !override {
			t.logger.Warningf("the same name artifact %s:%s exists on the destination registry, but the \"override\" is set to false, skip",
				dstRepo, dstRef)
			return nil
		}
		// the same name artifact exists, but allowed to override
		t.logger.Warningf("the same name artifact %s:%s exists on the destination registry and the \"override\" is set to true, continue...",
			dstRepo, dstRef)
	}

	// copy contents between the source and destination registries
	for _, content := range manifest.References() {
		if err = t.copyContent(content, srcRepo, dstRepo, opts); err != nil {
			return err
		}
	}

	// push the manifest to the destination registry
	if err := t.pushManifest(manifest, dstRepo, dstRef); err != nil {
		return err
	}

	t.logger.Infof("copy %s:%s(source registry) to %s:%s(destination registry) completed",
		srcRepo, srcRef, dstRepo, dstRef)
	return nil
}

// copy the content from source registry to destination according to its media type
func (t *transfer) copyContent(content distribution.Descriptor, srcRepo, dstRepo string, opts *trans.Options) error {
	digest := content.Digest.String()
	switch content.MediaType {
	// when the media type of pulled manifest is index,
	// the contents it contains are a few manifests/indexes
	case v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList,
		v1.MediaTypeImageManifest, schema2.MediaTypeManifest,
		schema1.MediaTypeSignedManifest, schema1.MediaTypeManifest:
		// as using digest as the reference, so set the override to true directly
		return t.copyArtifact(srcRepo, digest, dstRepo, digest, true, opts)
	// handle foreign layer
	case schema2.MediaTypeForeignLayer:
		t.logger.Infof("the layer %s is a foreign layer, skip", digest)
		return nil
	// copy layer or artifact config
	// the media type of the layer or config can be "application/octet-stream",
	// schema1.MediaTypeManifestLayer, schema2.MediaTypeLayer, schema2.MediaTypeImageConfig
	default:
		if opts.CopyByChunk {
			// copy by chunk
			return t.copyChunkWithRetry(srcRepo, dstRepo, digest, content.Size, opts.Speed)
		}
		// copy by blob
		return t.copyBlobWithRetry(srcRepo, dstRepo, digest, content.Size, opts.Speed)
	}
}

func (t *transfer) copyBlobWithRetry(srcRepo, dstRepo, digest string, sizeFromDescriptor int64, speed int32) error {
	var err error
	for i, backoff := 1, 2*time.Second; i <= blobRetryCnt; i, backoff = i+1, backoff*2 {
		t.logger.Infof("copying the blob %s(the %dth running)...", digest, i)
		if err = t.copyBlob(srcRepo, dstRepo, digest, sizeFromDescriptor, speed); err == nil {
			t.logger.Infof("copy the blob %s completed", digest)
			return nil
		}
		if i == blobRetryCnt || err == errStopped {
			break
		}
		t.logger.Infof("will retry %v later", backoff)
		time.Sleep(backoff)
	}
	return err
}

func (t *transfer) copyChunkWithRetry(srcRepo, dstRepo, digest string, sizeFromDescriptor int64, speed int32) error {
	var (
		err      error
		location string

		start int64 = -1
		end   int64 = -1
	)

	for i, backoff := 1, 2*time.Second; i <= chunkRetryCnt; i, backoff = i+1, backoff*2 {
		t.logger.Infof("copying the blob %s by chunk(chunkSize: %d)(the %dth running)...", digest, replicationChunkSize, i)
		if err = t.copyBlobByChunk(srcRepo, dstRepo, digest, sizeFromDescriptor, &start, &end, &location, speed); err == nil {
			t.logger.Infof("copy the blob %s by chunk completed", digest)
			return nil
		}
		if i == chunkRetryCnt || err == errStopped {
			break
		}
		t.logger.Infof("will retry %v later", backoff)
		time.Sleep(backoff)
	}

	return err
}

// tryMountBlob try to check existence and mount, return true if mounted.
func (t *transfer) tryMountBlob(_, dstRepo, digest string) (bool, error) {
	if t.shouldStop() {
		return false, errStopped
	}
	exist, err := t.dst.BlobExist(dstRepo, digest)
	if err != nil {
		t.logger.Errorf("failed to check the existence of blob %s on the destination registry: %v", digest, err)
		return false, err
	}
	if exist {
		t.logger.Infof("the blob %s already exists on the destination registry, skip", digest)
		// we think the blob is mounted if it is existed.
		return true, nil
	}

	mount, repository, err := t.dst.CanBeMount(digest)
	if err != nil {
		t.logger.Errorf("failed to check whether the blob %s can be mounted on the destination registry: %v", digest, err)
		return false, err
	}
	if mount {
		if err = t.dst.MountBlob(repository, digest, dstRepo); err != nil {
			t.logger.Errorf("failed to mount the blob %s on the destination registry: %v", digest, err)
			return false, err
		}
		t.logger.Infof("the blob %s mounted from the repository %s on the destination registry directly", digest, repository)
		return true, nil
	}

	return false, nil
}

// copy the layer or artifact config from the source registry to destination
// the size parameter is taken from manifests.
func (t *transfer) copyBlob(srcRepo, dstRepo, digest string, sizeFromDescriptor int64, speed int32) error {
	mounted, err := t.tryMountBlob(srcRepo, dstRepo, digest)
	if err != nil {
		return err
	}
	// return earlier if it is mounted
	if mounted {
		return nil
	}

	return t.copyBlobByMonolithic(srcRepo, dstRepo, digest, sizeFromDescriptor, speed)
}

func (t *transfer) copyBlobByMonolithic(srcRepo, dstRepo, digest string, sizeFromDescriptor int64, speed int32) error {
	size, data, err := t.src.PullBlob(srcRepo, digest)
	if err != nil {
		t.logger.Errorf("failed to pulling the blob %s: %v", digest, err)
		return err
	}
	if speed > 0 {
		data = lib.NewReader(data, speed)
	}
	defer data.Close()
	// get size 0 from PullBlob, use size from distribution.Descriptor instead.
	if size == 0 {
		size = sizeFromDescriptor
		t.logger.Debugf("the blob size from remote registry is 0, use size %d from manifests instead", size)
	}

	t.logger.Debugf("the blob size is %d bytes", size)

	if err = t.dst.PushBlob(dstRepo, digest, size, data); err != nil {
		t.logger.Errorf("failed to pushing the blob %s, size %d: %v", digest, size, err)
		return err
	}

	return nil
}

// copyBlobByChunk copy blob by chunk with specified start and end range.
// The <range> refers to the byte range of the chunk, and MUST be inclusive on both ends. The first chunk's range MUST begin with 0.
func (t *transfer) copyBlobByChunk(srcRepo, dstRepo, digest string, sizeFromDescriptor int64, start, end *int64, location *string, speed int32) error {
	mounted, err := t.tryMountBlob(srcRepo, dstRepo, digest)
	if err != nil {
		return err
	}
	// return earlier if it is mounted.
	if mounted {
		return nil
	}

	// fallback to copy by monolithic if the blob size is equal or less than chunk size.
	if sizeFromDescriptor <= replicationChunkSize {
		return t.copyBlobByMonolithic(srcRepo, dstRepo, digest, sizeFromDescriptor, speed)
	}

	// end range should equal (blobSize - 1)
	endRange := sizeFromDescriptor - 1
	for {
		// update the start and end for upload
		*start = *end + 1
		// since both ends are closed intervals, it is necessary to subtract one byte
		*end = min(*start+replicationChunkSize-1, endRange)

		t.logger.Infof("copying the blob chunk: %d-%d/%d", *start, *end, sizeFromDescriptor)
		_, data, err := t.src.PullBlobChunk(srcRepo, digest, sizeFromDescriptor, *start, *end)
		if err != nil {
			t.logger.Errorf("failed to pulling the blob chunk: %d-%d/%d, error: %v", *start, *end, sizeFromDescriptor, err)
			return err
		}

		if speed > 0 {
			data = lib.NewReader(data, speed)
		}
		// failureEnd will only be used for adjusting content range when issue happened during push the chunk.
		var failureEnd int64
		*location, failureEnd, err = t.dst.PushBlobChunk(dstRepo, digest, sizeFromDescriptor, data, *start, *end, *location)
		if err != nil {
			t.logger.Errorf("failed to pushing the blob chunk: %d-%d/%d, error: %v", *start, *end, sizeFromDescriptor, err)
			data.Close()
			*end = failureEnd
			return err
		}

		data.Close()

		t.logger.Infof("copy the blob chunk: %d-%d/%d completed", *start, *end, sizeFromDescriptor)
		// if the end equals (blobSize-1), that means it is last chunk, return if this is the last chunk
		if *end == endRange {
			break
		}
	}

	return nil
}

func (t *transfer) pullManifest(repository, reference string) (
	distribution.Manifest, string, error) {
	if t.shouldStop() {
		return nil, "", errStopped
	}
	t.logger.Infof("pulling the manifest of artifact %s:%s ...", repository, reference)
	manifest, digest, err := t.src.PullManifest(repository, reference)
	if err != nil {
		t.logger.Errorf("failed to pull the manifest of artifact %s:%s: %v", repository, reference, err)
		return nil, "", err
	}
	t.logger.Infof("the manifest of artifact %s:%s pulled", repository, reference)

	return manifest, digest, nil
}

func (t *transfer) exist(repository, tag string) (bool, string, error) {
	exist, desc, err := t.dst.ManifestExist(repository, tag)
	if err != nil {
		t.logger.Errorf("failed to check the existence of the manifest of artifact %s:%s on the destination registry: %v",
			repository, tag, err)
		return false, "", err
	}
	var dig string
	if desc != nil {
		dig = string(desc.Digest)
	}
	return exist, dig, nil
}

func (t *transfer) pushManifest(manifest distribution.Manifest, repository, tag string) error {
	if t.shouldStop() {
		return errStopped
	}
	t.logger.Infof("pushing the manifest of artifact %s:%s ...", repository, tag)
	mediaType, payload, err := manifest.Payload()
	if err != nil {
		t.logger.Errorf("failed to push manifest of artifact %s:%s: %v",
			repository, tag, err)
		return err
	}
	if _, err := t.dst.PushManifest(repository, tag, mediaType, payload); err != nil {
		t.logger.Errorf("failed to push manifest of artifact %s:%s: %v",
			repository, tag, err)
		return err
	}
	t.logger.Infof("the manifest of artifact %s:%s pushed",
		repository, tag)
	return nil
}

func (t *transfer) delete(repo *repository) error {
	if t.shouldStop() {
		return nil
	}

	repository := repo.repository
	for _, tag := range repo.tags {
		exist, _, err := t.dst.ManifestExist(repository, tag)
		if err != nil {
			t.logger.Errorf("failed to check the existence of the manifest of artifact %s:%s on the destination registry: %v",
				repository, tag, err)
			return err
		}
		if !exist {
			t.logger.Infof("the artifact %s:%s doesn't exist on the destination registry, skip",
				repository, tag)
			continue
		}
		if err := t.dst.DeleteManifest(repository, tag); err != nil {
			t.logger.Errorf("failed to delete the manifest of artifact %s:%s on the destination registry: %v",
				repository, tag, err)
			return err
		}
		t.logger.Infof("the manifest of artifact %s:%s is deleted", repository, tag)
	}
	return nil
}

func (t *transfer) deleteTag(res *model.Resource) error {
	if t.shouldStop() {
		return nil
	}
	repository := res.Metadata.Repository.Name
	for _, art := range res.Metadata.Artifacts {
		for _, tag := range art.Tags {
			if err := t.dst.DeleteTag(repository, tag); err != nil {
				if e, ok := err.(*common_http.Error); ok && e.Code == http.StatusNotFound {
					t.logger.Infof("the tag %s:%s doesn't exist on the destination registry, skip", repository, tag)
					continue
				}
				t.logger.Errorf("failed to delete the tag %s:%s on the destination registry: %v",
					repository, tag, err)
				return err
			}
			t.logger.Infof("the tag %s:%s is deleted", repository, tag)
		}
	}
	return nil
}
