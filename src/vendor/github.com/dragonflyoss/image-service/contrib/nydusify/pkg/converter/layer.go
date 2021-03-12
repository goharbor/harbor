// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/containerd/containerd/images"
	"github.com/dustin/go-humanize"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/backend"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/build"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/cache"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter/provider"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/remote"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/utils"
)

type buildLayer struct {
	index  int
	source provider.SourceLayer

	backend        backend.Backend
	remote         *remote.Remote
	buildWorkflow  *build.Workflow
	cacheGlue      *cacheGlue
	bootstrapsDir  string
	dockerV2Format bool

	cacheRecord     *cache.CacheRecord
	blobDesc        *ocispec.Descriptor
	bootstrapDesc   *ocispec.Descriptor
	bootstrapDiffID *digest.Digest
	parent          *buildLayer
	sourceDir       string
	blobPath        string
	bootstrapPath   string
}

func (layer *buildLayer) pushBlob(ctx context.Context) (*ocispec.Descriptor, error) {
	// Note: filepath.Base(blobPath) is a sha256 hex string
	blobID := filepath.Base(layer.blobPath)
	blobPath := layer.blobPath

	blobDigest := digest.NewDigestFromEncoded(digest.SHA256, blobID)
	info, err := os.Stat(blobPath)
	if err != nil {
		return nil, errors.Wrap(err, "Stat blob file")
	}

	desc := ocispec.Descriptor{
		Digest:    blobDigest,
		Size:      info.Size(),
		MediaType: utils.MediaTypeNydusBlob,
		Annotations: map[string]string{
			// Use `utils.LayerAnnotationUncompressed` to generate
			// DiffID of layer defined in OCI spec
			utils.LayerAnnotationUncompressed: blobDigest.String(),
			utils.LayerAnnotationNydusBlob:    "true",
		},
	}

	// Upload Nydus blob to backend if backend config be specified
	if layer.backend != nil {
		if err := layer.backend.Upload(blobID, blobPath); err != nil {
			return nil, errors.Wrap(err, "Upload blob to backend")
		}
		return &desc, nil
	}

	blobFile, err := os.Open(blobPath)
	if err != nil {
		return nil, errors.Wrap(err, "Open blob file")
	}
	defer blobFile.Close()

	if err := layer.remote.Push(ctx, desc, true, blobFile); err != nil {
		return nil, errors.Wrap(err, "Push blob layer")
	}

	return &desc, nil
}

func (layer *buildLayer) pushBootstrap(ctx context.Context) (*ocispec.Descriptor, *digest.Digest, error) {
	// TODO: make these PackTargzInfo calls concurrently
	compressedDigest, compressedSize, err := utils.PackTargzInfo(
		layer.bootstrapPath, utils.BootstrapFileNameInLayer, true,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Calculate compressed boostrap digest")
	}

	uncompressedDigest, _, err := utils.PackTargzInfo(
		layer.bootstrapPath, utils.BootstrapFileNameInLayer, false,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Calculate uncompressed boostrap digest")
	}

	compressedReader, err := utils.PackTargz(
		layer.bootstrapPath, utils.BootstrapFileNameInLayer, true,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Compress boostrap layer")
	}
	defer compressedReader.Close()

	bootstrapMediaType := ocispec.MediaTypeImageLayerGzip
	if layer.dockerV2Format {
		bootstrapMediaType = images.MediaTypeDockerSchema2LayerGzip
	}

	desc := ocispec.Descriptor{
		Digest:    compressedDigest,
		Size:      compressedSize,
		MediaType: bootstrapMediaType,
		Annotations: map[string]string{
			// Use `utils.LayerAnnotationUncompressed` to generate
			// DiffID of layer defined in OCI spec
			utils.LayerAnnotationUncompressed:   uncompressedDigest.String(),
			utils.LayerAnnotationNydusBootstrap: "true",
		},
	}

	if err := layer.remote.Push(ctx, desc, true, compressedReader); err != nil {
		return nil, nil, errors.Wrap(err, "Push bootstrap layer")
	}

	return &desc, &uncompressedDigest, nil
}

func (layer *buildLayer) Push(ctx context.Context) error {
	// Push Nydus bootstrap layer to remote registry
	bootstrapInfo, err := os.Stat(layer.bootstrapPath)
	if err != nil {
		return errors.Wrap(err, "Get bootstrap layer size")
	}
	bootstrapSize := humanize.Bytes(uint64(bootstrapInfo.Size()))
	pushDone := logger.Log(ctx, "[BOOT] Push bootstrap", provider.LoggerFields{
		"ChainID": layer.source.ChainID(),
		"Size":    bootstrapSize,
	})
	layer.bootstrapDesc, layer.bootstrapDiffID, err = layer.pushBootstrap(ctx)
	if err != nil {
		return pushDone(errors.Wrapf(err, "Push Nydus bootstrap layer"))
	}
	pushDone(nil)

	// Push Nydus blob layer to remote registry
	if layer.blobPath != "" {
		blobDigest := digest.NewDigestFromEncoded(digest.SHA256, filepath.Base(layer.blobPath))
		info, err := os.Stat(layer.blobPath)
		if err != nil {
			return errors.Wrap(err, "Get blob layer size")
		}
		blobSize := humanize.Bytes(uint64(info.Size()))
		var op string
		if layer.backend != nil {
			op = "Upload"
		} else {
			op = "Push"
		}
		pushDone := logger.Log(ctx, fmt.Sprintf("[BLOB] %s blob", op), provider.LoggerFields{
			"Digest": blobDigest,
			"Size":   blobSize,
		})
		layer.blobDesc, err = layer.pushBlob(ctx)
		if err != nil {
			return pushDone(errors.Wrapf(err, "Push Nydus blob layer"))
		}
		pushDone(nil)
	}

	// Also push Nydus bootstrap and blob layer to cache image, because maybe
	// the cache image is located in different namespace/repo
	if err := layer.cacheGlue.Push(ctx, layer); err != nil {
		return errors.Wrapf(err, "Push layer to cache image")
	}

	return nil
}

func (layer *buildLayer) Mount(ctx context.Context) (func() error, error) {
	sourceLayerSize := humanize.Bytes(uint64(layer.source.Size()))

	// Give priority to checking & pulling Nydus layer from cache image
	cacheRecord, err := layer.cacheGlue.Pull(ctx, layer.source.ChainID())
	if err != nil {
		return nil, errors.Wrap(err, "Get cache record")
	}
	if cacheRecord != nil {
		layer.cacheRecord = cacheRecord
		return nil, nil
	}

	bootstrapName := strconv.Itoa(layer.index+1) + "-" + layer.source.ChainID().String()
	layer.bootstrapPath = filepath.Join(layer.bootstrapsDir, bootstrapName)

	// Pull source layer for building on next if no cache hit
	mountDone := logger.Log(ctx, fmt.Sprintf("[SOUR] Pull layer"), provider.LoggerFields{
		"ChainID": layer.source.ChainID(),
		"Size":    sourceLayerSize,
	})
	var umount func() error
	layer.sourceDir, umount, err = layer.source.Mount(ctx)
	if err != nil {
		return nil, mountDone(errors.Wrapf(err, "Mount source layer %s", layer.source.Digest()))
	}

	return umount, mountDone(nil)
}

func (layer *buildLayer) Build(ctx context.Context) error {
	sourceSize := humanize.Bytes(uint64(layer.source.Size()))

	// Build Nydus blob and bootstrap file to temp directory
	buildDone := logger.Log(ctx, fmt.Sprintf("[DUMP] Build layer"), provider.LoggerFields{
		"Digest": layer.source.Digest(),
		"Size":   sourceSize,
	})
	parentBootstrapPath := ""
	parentLayer := layer.parent
	if parentLayer != nil {
		// Try to reuse the bootstrap of parent layer in cache record
		if parentLayer.Cached() {
			bootstrapName := strconv.Itoa(parentLayer.index+1) + "-" + parentLayer.source.ChainID().String()
			parentLayer.bootstrapPath = filepath.Join(parentLayer.bootstrapsDir, bootstrapName+"-cached")
			if err := parentLayer.cacheGlue.PullBootstrap(ctx, parentLayer.source.ChainID(), parentLayer.bootstrapPath); err != nil {
				logrus.Warn(errors.Wrap(err, "Pull bootstrap from cache"))
				// Error occurs, the cache is invalid
				return buildDone(errInvalidCache)
			}
		}
		parentBootstrapPath = parentLayer.bootstrapPath
	}
	blobPath, err := layer.buildWorkflow.Build(layer.sourceDir, parentBootstrapPath, layer.bootstrapPath)
	if err != nil {
		return buildDone(errors.Wrapf(err, "Build source layer %s", layer.source.Digest()))
	}
	layer.blobPath = blobPath

	return buildDone(nil)
}

func (layer *buildLayer) GetCacheRecord() cache.CacheRecord {
	if layer.cacheRecord != nil {
		return *layer.cacheRecord
	}
	return cache.CacheRecord{
		SourceChainID:        layer.source.ChainID(),
		NydusBlobDesc:        layer.blobDesc,
		NydusBootstrapDesc:   layer.bootstrapDesc,
		NydusBootstrapDiffID: *layer.bootstrapDiffID,
	}
}

func (layer *buildLayer) Cached() bool {
	return layer.cacheRecord != nil
}
