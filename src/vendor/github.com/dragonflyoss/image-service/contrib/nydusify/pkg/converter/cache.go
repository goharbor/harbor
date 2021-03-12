// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"context"
	"fmt"
	"os"

	"github.com/opencontainers/go-digest"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/backend"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/cache"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter/provider"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/remote"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/utils"
)

type cacheGlue struct {
	cache *cache.Cache
	// Remote object for cache image
	cacheRemote *remote.Remote
	// Remote object for target image
	remote *remote.Remote
}

func newCacheGlue(
	ctx context.Context, maxRecords uint, dockerV2Format bool, remote *remote.Remote, cacheRemote *remote.Remote, backend backend.Backend,
) (*cacheGlue, error) {
	if cacheRemote == nil {
		return &cacheGlue{}, nil
	}

	pullDone := logger.Log(ctx, fmt.Sprintf("[CACH] Import from %s", cacheRemote.Ref), nil)

	// Pull Nydus cache image from remote registry
	cache, err := cache.New(cacheRemote, cache.Opt{
		MaxRecords:     maxRecords,
		DockerV2Format: dockerV2Format,
		Backend:        backend,
	})
	if err != nil {
		return nil, pullDone(errors.Wrap(err, "Init nydus cache"))
	}

	// Ingore the error of importing cache image, it doesn't affect
	// the build workflow.
	if err := cache.Import(ctx); err != nil {
		logrus.Warnf("Failed to import cache: %s", err)
	}

	return &cacheGlue{
		cache:       cache,
		cacheRemote: cacheRemote,
		remote:      remote,
	}, pullDone(nil)
}

func (cg *cacheGlue) Pull(
	ctx context.Context, sourceLayerChainID digest.Digest,
) (*cache.CacheRecord, error) {
	if cg.cache == nil {
		return nil, nil
	}

	var cacheRecord *cache.CacheRecord

	// Using ChainID to ensure we can find corresponding overlayed
	// Nydus blob/bootstrap layer in cache records.
	_cacheRecord, bootstrapReader, blobReader, err := cg.cache.Check(ctx, sourceLayerChainID)
	if err == nil && _cacheRecord != nil {
		pullDone := logger.Log(ctx, "[CACH] Check layer", provider.LoggerFields{
			"ChainID": sourceLayerChainID,
		})
		// Pull the cached layer from cache image, then push to target namespace/repo,
		// because the blob data is not shared between diffrent namespace in registry,
		// this operation ensure that Nydus image own these layers.
		cacheRecord = _cacheRecord
		defer bootstrapReader.Close()
		if err := cg.remote.Push(ctx, *cacheRecord.NydusBootstrapDesc, true, bootstrapReader); err != nil {
			return nil, pullDone(errors.Wrapf(err, "Push cached bootstrap layer"))
		}
		if blobReader != nil && cacheRecord.NydusBlobDesc != nil {
			defer blobReader.Close()
			if err := cg.remote.Push(ctx, *cacheRecord.NydusBlobDesc, true, blobReader); err != nil {
				return nil, pullDone(errors.Wrapf(err, "Push cached blob layer"))
			}
		}
		pullDone(nil)
	}

	return cacheRecord, nil
}

func (cg *cacheGlue) Push(ctx context.Context, layer *buildLayer) error {
	if cg.cache == nil {
		return nil
	}

	pushDone := logger.Log(ctx, "[CACH] Push layer", provider.LoggerFields{
		"ChainID": layer.source.ChainID(),
	})

	// Push bootstrap layer to cache image
	bootstrapReader, err := utils.PackTargz(
		layer.bootstrapPath, utils.BootstrapFileNameInLayer, true,
	)
	if err != nil {
		return pushDone(errors.Wrapf(err, "Compress bootstrap layer"))
	}
	defer bootstrapReader.Close()
	if err := cg.cache.Push(ctx, *layer.bootstrapDesc, bootstrapReader); err != nil {
		return pushDone(errors.Wrapf(err, "Push target bootstrap layer to nydus cache"))
	}

	// Push blob layer to cache image
	if layer.backend == nil && layer.blobPath != "" {
		blobFile, err := os.Open(layer.blobPath)
		if err != nil {
			return pushDone(errors.Wrapf(err, "Open blob file"))
		}
		defer blobFile.Close()
		if err := cg.cache.Push(ctx, *layer.blobDesc, blobFile); err != nil {
			return pushDone(errors.Wrapf(err, "Push target blob layer to nydus cache"))
		}
	}

	return pushDone(nil)
}

func (cg *cacheGlue) PullBootstrap(
	ctx context.Context, chainID digest.Digest, pulledBootstrapPath string,
) error {
	if cg.cache == nil {
		return nil
	}

	cacheRecord, bootstrapReader, blobReader, _ := cg.cache.Check(ctx, chainID)
	if cacheRecord != nil {
		defer bootstrapReader.Close()
		if blobReader != nil {
			defer blobReader.Close()
		}
		bootstrapDesc := cacheRecord.NydusBootstrapDesc
		pullDone := logger.Log(ctx, fmt.Sprintf("[CACH] Pull bootstrap"), provider.LoggerFields{
			"ChainID": chainID,
		})
		// Pull the bootstrap layer recorded in cache image for build workflow
		if err := cg.cache.PullBootstrap(ctx, bootstrapDesc, pulledBootstrapPath); err != nil {
			return pullDone(errors.Wrapf(err, "Pull bootstrap from cache image"))
		}
		return pullDone(nil)
	}

	return fmt.Errorf("Not found bootstrap in cache")
}

func (cg *cacheGlue) Export(
	ctx context.Context, buildLayers []*buildLayer,
) error {
	if cg.cache == nil {
		return nil
	}

	pushDone := logger.Log(ctx, fmt.Sprintf("[CACH] Export to %s", cg.cacheRemote.Ref), nil)

	// Re-import cache from remote registry to avoid conflicts with another
	// conversion progress as much as possible
	cg.cache.Import(ctx)

	cacheRecords := []*cache.CacheRecord{}
	for _, layer := range buildLayers {
		record := layer.GetCacheRecord()
		cacheRecords = append(cacheRecords, &record)
	}
	cg.cache.Record(cacheRecords)

	// Push cache image to remote registry
	if err := cg.cache.Export(ctx); err != nil {
		logrus.Warnf("Failed to export cache: %s", err)
	}

	return pushDone(nil)
}
