// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/backend"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/remote"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/utils"

	"github.com/containerd/containerd/images"
	digest "github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// Opt configures Nydus cache
type Opt struct {
	// Maximum records(bootstrap layer + blob layer) in cache image.
	MaxRecords uint
	// Make cache image manifest compatible with the docker v2 media
	// type defined in github.com/containerd/containerd/images.
	DockerV2Format bool
	// The blob layer record will not be written to cache image if
	// the backend be specified, because the blob layer will be uploaded
	// to backend.
	Backend backend.Backend
}

// Cache creates an image to store cache records in its image manifest,
// every record presents the relationship like:
//
// source_layer_chainid -> (nydus_blob_layer_digest, nydus_bootstrap_layer_digest)
// If the converter hit cache record during build source layer, we can
// skip the layer building, see cache image example: examples/manifest/cache_manifest.json.
//
// Here is the build cache workflow:
// 1. Import cache records from registry;
// 2. Check cache record using source layer ChainID before layer build,
//    skip layer build if the cache hit;
// 3. Export new cache records to registry;
type Cache struct {
	opt Opt
	// Remote is responsible for pulling & pushing cache image
	remote *remote.Remote
	// Store the pulled records from registry
	pulledRecords map[digest.Digest]*CacheRecord
	// Store the records prepared to push to registry
	pushedRecords []*CacheRecord
}

// New creates Nydus cache instance,
func New(remote *remote.Remote, opt Opt) (*Cache, error) {
	cache := &Cache{
		opt:    opt,
		remote: remote,
		// source_layer_chainid -> cache_record
		pulledRecords: make(map[digest.Digest]*CacheRecord),
		pushedRecords: []*CacheRecord{},
	}

	return cache, nil
}

func (cache *Cache) recordToLayer(record *CacheRecord) (*ocispec.Descriptor, *ocispec.Descriptor) {
	bootstrapCacheMediaType := ocispec.MediaTypeImageLayerGzip
	if cache.opt.DockerV2Format {
		bootstrapCacheMediaType = images.MediaTypeDockerSchema2LayerGzip
	}

	bootstrapCacheDesc := &ocispec.Descriptor{
		MediaType: bootstrapCacheMediaType,
		Digest:    record.NydusBootstrapDesc.Digest,
		Size:      record.NydusBootstrapDesc.Size,
		Annotations: map[string]string{
			utils.LayerAnnotationNydusBootstrap:     "true",
			utils.LayerAnnotationNydusSourceChainID: record.SourceChainID.String(),
			// Use the annotation to record bootstrap layer DiffID.
			utils.LayerAnnotationUncompressed: record.NydusBootstrapDiffID.String(),
		},
	}

	var blobCacheDesc *ocispec.Descriptor
	if record.NydusBlobDesc != nil {
		// Record blob layer to cache image if the blob be pushed
		// to registry instead of storage backend.
		if cache.opt.Backend == nil {
			blobCacheDesc = &ocispec.Descriptor{
				MediaType: utils.MediaTypeNydusBlob,
				Digest:    record.NydusBlobDesc.Digest,
				Size:      record.NydusBlobDesc.Size,
				Annotations: map[string]string{
					utils.LayerAnnotationNydusBlob:          "true",
					utils.LayerAnnotationNydusSourceChainID: record.SourceChainID.String(),
				},
			}
		} else {
			bootstrapCacheDesc.Annotations[utils.LayerAnnotationNydusBlobDigest] = record.NydusBlobDesc.Digest.String()
			bootstrapCacheDesc.Annotations[utils.LayerAnnotationNydusBlobSize] = strconv.FormatInt(record.NydusBlobDesc.Size, 10)
		}
	}

	return bootstrapCacheDesc, blobCacheDesc
}

func (cache *Cache) exportRecordsToLayers() []ocispec.Descriptor {
	layers := []ocispec.Descriptor{}

	for _, record := range cache.pushedRecords {
		bootstrapCacheDesc, blobCacheDesc := cache.recordToLayer(record)
		layers = append(layers, *bootstrapCacheDesc)
		if blobCacheDesc != nil {
			layers = append(layers, *blobCacheDesc)
		}
	}

	return layers
}

func (cache *Cache) layerToRecord(layer *ocispec.Descriptor) *CacheRecord {
	sourceChainIDStr, ok := layer.Annotations[utils.LayerAnnotationNydusSourceChainID]
	if !ok {
		return nil
	}
	sourceChainID := digest.Digest(sourceChainIDStr)
	if sourceChainID.Validate() != nil {
		return nil
	}
	if layer.Annotations == nil {
		return nil
	}

	// Handle bootstrap cache layer
	if layer.Annotations[utils.LayerAnnotationNydusBootstrap] == "true" {
		uncompressedDigestStr := layer.Annotations[utils.LayerAnnotationUncompressed]
		if uncompressedDigestStr == "" {
			return nil
		}
		bootstrapDiffID := digest.Digest(uncompressedDigestStr)
		if bootstrapDiffID.Validate() != nil {
			return nil
		}
		bootstrapDesc := ocispec.Descriptor{
			MediaType: layer.MediaType,
			Digest:    layer.Digest,
			Size:      layer.Size,
			Annotations: map[string]string{
				utils.LayerAnnotationNydusBootstrap: "true",
				utils.LayerAnnotationUncompressed:   uncompressedDigestStr,
			},
		}
		var nydusBlobDesc *ocispec.Descriptor
		if layer.Annotations[utils.LayerAnnotationNydusBlobDigest] != "" &&
			layer.Annotations[utils.LayerAnnotationNydusBlobSize] != "" {
			blobDigest := digest.Digest(layer.Annotations[utils.LayerAnnotationNydusBlobDigest])
			if blobDigest.Validate() != nil {
				return nil
			}
			blobSize, err := strconv.ParseInt(layer.Annotations[utils.LayerAnnotationNydusBlobSize], 10, 64)
			if err != nil {
				return nil
			}
			nydusBlobDesc = &ocispec.Descriptor{
				MediaType: utils.MediaTypeNydusBlob,
				Digest:    blobDigest,
				Size:      blobSize,
				Annotations: map[string]string{
					utils.LayerAnnotationNydusBlob: "true",
				},
			}
		}
		return &CacheRecord{
			SourceChainID:        sourceChainID,
			NydusBootstrapDesc:   &bootstrapDesc,
			NydusBlobDesc:        nydusBlobDesc,
			NydusBootstrapDiffID: bootstrapDiffID,
		}
	}

	// Handle blob cache layer
	if layer.Annotations[utils.LayerAnnotationNydusBlob] == "true" {
		nydusBlobDesc := &ocispec.Descriptor{
			MediaType: layer.MediaType,
			Digest:    layer.Digest,
			Size:      layer.Size,
			Annotations: map[string]string{
				utils.LayerAnnotationNydusBlob: "true",
			},
		}
		return &CacheRecord{
			SourceChainID: sourceChainID,
			NydusBlobDesc: nydusBlobDesc,
		}
	}

	return nil
}

func mergeRecord(old, new *CacheRecord) *CacheRecord {
	if old == nil {
		old = &CacheRecord{
			SourceChainID: new.SourceChainID,
		}
	}

	if new.NydusBootstrapDesc != nil {
		old.NydusBootstrapDesc = new.NydusBootstrapDesc
		old.NydusBootstrapDiffID = new.NydusBootstrapDiffID
	}

	if new.NydusBlobDesc != nil {
		old.NydusBlobDesc = new.NydusBlobDesc
	}

	return old
}

func (cache *Cache) importLayersToRecords(layers []ocispec.Descriptor) {
	pulledRecords := make(map[digest.Digest]*CacheRecord)
	pushedRecords := []*CacheRecord{}

	for idx := range layers {
		record := cache.layerToRecord(&layers[idx])
		if record != nil {
			// Merge bootstrap and related blob layer to record
			newRecord := mergeRecord(
				pulledRecords[record.SourceChainID],
				record,
			)
			pulledRecords[record.SourceChainID] = newRecord
			pushedRecords = append(pushedRecords, newRecord)
		}
	}

	cache.pulledRecords = pulledRecords
	cache.pushedRecords = pushedRecords
}

// Export pushes cache manifest index to remote registry
func (cache *Cache) Export(ctx context.Context) error {
	if len(cache.pushedRecords) == 0 {
		return nil
	}

	layers := cache.exportRecordsToLayers()

	// Prepare empty image config, just for registry API compatibility,
	// manifest requires a valid config field.
	configMediaType := ocispec.MediaTypeImageConfig
	if cache.opt.DockerV2Format {
		configMediaType = images.MediaTypeDockerSchema2Config
	}
	config := ocispec.Image{
		Config: ocispec.ImageConfig{},
		RootFS: ocispec.RootFS{},
	}
	configDesc, configBytes, err := utils.MarshalToDesc(config, configMediaType)
	if err != nil {
		return errors.Wrap(err, "Marshal cache config")
	}
	if err := cache.remote.Push(ctx, *configDesc, false, bytes.NewReader(configBytes)); err != nil {
		return errors.Wrap(err, "Push cache config")
	}

	// Push cache manifest to remote registry
	mediaType := ocispec.MediaTypeImageManifest
	if cache.opt.DockerV2Format {
		mediaType = images.MediaTypeDockerSchema2Manifest
	}

	manifest := CacheManifest{
		MediaType: mediaType,
		Manifest: ocispec.Manifest{
			Versioned: specs.Versioned{
				SchemaVersion: 2,
			},
			// Just for registry API compatibility, registry required a
			// valid config field.
			Config: *configDesc,
			Layers: layers,
			Annotations: map[string]string{
				utils.ManifestNydusCache: utils.ManifestNydusCacheVersion,
			},
		},
	}

	manifestDesc, manifestBytes, err := utils.MarshalToDesc(manifest, manifest.MediaType)
	if err != nil {
		return errors.Wrap(err, "Push cache manifest")
	}

	if err := cache.remote.Push(ctx, *manifestDesc, false, bytes.NewReader(manifestBytes)); err != nil {
		return errors.Wrap(err, "Push cache manifest")
	}

	return nil
}

// Import pulls cache manifest index from remote registry
func (cache *Cache) Import(ctx context.Context) error {
	manifestDesc, err := cache.remote.Resolve(ctx)
	if err != nil {
		return errors.Wrap(err, "Resolve cache image")
	}

	// Fetch cache manifest from remote registry
	manifestReader, err := cache.remote.Pull(ctx, *manifestDesc, true)
	if err != nil {
		return errors.Wrap(err, "Pull cache image")
	}
	defer manifestReader.Close()

	manifestBytes, err := ioutil.ReadAll(manifestReader)
	if err != nil {
		return errors.Wrap(err, "Read cache manifest")
	}

	var manifest CacheManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return errors.Wrap(err, "Unmarshal cache manifest")
	}

	// Discard the cache mismatched version
	if manifest.Annotations[utils.ManifestNydusCache] != utils.ManifestNydusCacheVersion {
		return fmt.Errorf("Unmatched cache version %s", manifest.Annotations[utils.ManifestNydusCache])
	}

	cache.importLayersToRecords(manifest.Layers)

	return nil
}

// Check checks bootstrap & blob layer exists in registry or storage backend
func (cache *Cache) Check(ctx context.Context, layerChainID digest.Digest) (*CacheRecord, io.ReadCloser, io.ReadCloser, error) {
	record, ok := cache.pulledRecords[layerChainID]
	if !ok {
		return nil, nil, nil, nil
	}

	// Check bootstrap layer on cache
	bootstrapReader, err := cache.remote.Pull(ctx, *record.NydusBootstrapDesc, true)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Check bootstrap layer")
	}

	// Check blob layer on cache
	if record.NydusBlobDesc != nil {
		if cache.opt.Backend == nil {
			blobReader, err := cache.remote.Pull(ctx, *record.NydusBlobDesc, true)
			if err != nil {
				return nil, nil, nil, errors.Wrap(err, "Check blob layer")
			}
			return record, bootstrapReader, blobReader, nil
		}
		exist, err := cache.opt.Backend.Check(record.NydusBlobDesc.Digest.Hex())
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "Check blob on backend")
		}
		if !exist {
			return nil, nil, nil, errors.New("Not found blob on backend")
		}
		return record, bootstrapReader, nil, nil
	}

	return record, bootstrapReader, nil, nil
}

// Record puts new bootstrap & blob layer to cache record, it's a limited queue.
func (cache *Cache) Record(records []*CacheRecord) {
	moveFront := map[digest.Digest]bool{}
	for _, record := range records {
		moveFront[record.SourceChainID] = true
	}

	pushedRecords := records
	for _, record := range cache.pushedRecords {
		if !moveFront[record.SourceChainID] {
			pushedRecords = append(pushedRecords, record)
			if len(pushedRecords) >= int(cache.opt.MaxRecords) {
				break
			}
		}
	}

	if len(pushedRecords) > int(cache.opt.MaxRecords) {
		cache.pushedRecords = pushedRecords[:int(cache.opt.MaxRecords)]
	} else {
		cache.pushedRecords = pushedRecords
	}
}

// PullBootstrap pulls bootstrap layer from registry, and unpack to a specified path,
// we can use it to prepare parent bootstrap for building.
func (cache *Cache) PullBootstrap(ctx context.Context, bootstrapDesc *ocispec.Descriptor, target string) error {
	reader, err := cache.remote.Pull(ctx, *bootstrapDesc, true)
	if err != nil {
		return errors.Wrap(err, "Pull cached bootstrap layer")
	}
	defer reader.Close()

	if err := utils.UnpackFile(reader, utils.BootstrapFileNameInLayer, target); err != nil {
		return errors.Wrap(err, "Unpack cached bootstrap layer")
	}

	return nil
}

// Push pushes cache image to registry
func (cache *Cache) Push(ctx context.Context, desc ocispec.Descriptor, reader io.Reader) error {
	return cache.remote.Push(ctx, desc, true, reader)
}
