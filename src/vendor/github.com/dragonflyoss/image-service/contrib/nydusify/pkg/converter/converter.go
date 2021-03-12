// Copyright 2020 Ant Group. All rights reserved.
//
// SPDX-License-Identifier: Apache-2.0

package converter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/backend"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/build"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter/provider"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/remote"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/utils"
)

// PullWorkerCount specifies source layer pull concurrency
var PullWorkerCount uint = 5

// PushWorkerCount specifies Nydus layer push concurrency
var PushWorkerCount uint = 5

var logger provider.ProgressLogger

var (
	errInvalidCache = errors.New("Invalid cache")
)

type mountJob struct {
	err    error
	ctx    context.Context
	layer  *buildLayer
	umount func() error
}

func (job *mountJob) Do() error {
	var umount func() error
	umount, job.err = job.layer.Mount(job.ctx)
	job.umount = umount
	return job.err
}

func (job *mountJob) Err() error {
	return job.err
}

type Opt struct {
	Logger         provider.ProgressLogger
	SourceProvider provider.SourceProvider

	TargetRemote *remote.Remote

	CacheRemote     *remote.Remote
	CacheMaxRecords uint

	NydusImagePath string
	WorkDir        string
	PrefetchDir    string
	WhiteoutSpec   string

	MultiPlatform  bool
	DockerV2Format bool

	BackendType   string
	BackendConfig string
}

type Converter struct {
	Opt
}

func New(opt Opt) (*Converter, error) {
	return &Converter{
		Opt: opt,
	}, nil
}

func (cvt *Converter) convert(ctx context.Context) error {
	logger = cvt.Logger

	logrus.Infoln(fmt.Sprintf("Converting to %s", cvt.TargetRemote.Ref))

	// Init backend to upload Nydus blob if the backend config
	// option be specified
	var _backend backend.Backend
	var err error
	if cvt.BackendConfig != "" {
		_backend, err = backend.NewBackend(cvt.BackendType, cvt.BackendConfig)
		if err != nil {
			return errors.Wrap(err, "Init backend")
		}
	}

	// Try to pull Nydus cache image from remote registry
	cg, err := newCacheGlue(
		ctx, cvt.CacheMaxRecords, cvt.DockerV2Format, cvt.TargetRemote, cvt.CacheRemote, _backend,
	)
	if err != nil {
		return errors.Wrap(err, "Pull cache image")
	}

	// BuildWorkflow builds nydus blob/bootstrap layer by layer
	bootstrapsDir := filepath.Join(cvt.WorkDir, "bootstraps")
	if err := os.RemoveAll(bootstrapsDir); err != nil {
		return errors.Wrap(err, "Remove bootstrap directory")
	}
	if err := os.MkdirAll(bootstrapsDir, 0755); err != nil {
		return errors.Wrap(err, "Create bootstrap directory")
	}
	buildWorkflow, err := build.NewWorkflow(build.WorkflowOption{
		NydusImagePath: cvt.NydusImagePath,
		PrefetchDir:    cvt.PrefetchDir,
		TargetDir:      cvt.WorkDir,
		WhiteoutSpec:   cvt.WhiteoutSpec,
	})
	if err != nil {
		return errors.Wrap(err, "Create build flow")
	}

	sourceLayers, err := cvt.SourceProvider.Layers(ctx)
	if err != nil {
		return errors.Wrap(err, "Get source layers")
	}
	pullWorker := utils.NewQueueWorkerPool(PullWorkerCount, uint(len(sourceLayers)))
	pushWorker := utils.NewWorkerPool(PushWorkerCount, uint(len(sourceLayers)))
	buildLayers := []*buildLayer{}

	// Pull and mount source layer in pull worker
	var parentBuildLayer *buildLayer
	for idx, sourceLayer := range sourceLayers {
		buildLayer := &buildLayer{
			index:          idx,
			buildWorkflow:  buildWorkflow,
			bootstrapsDir:  bootstrapsDir,
			cacheGlue:      cg,
			backend:        _backend,
			remote:         cvt.TargetRemote,
			source:         sourceLayer,
			parent:         parentBuildLayer,
			dockerV2Format: cvt.DockerV2Format,
		}
		parentBuildLayer = buildLayer
		buildLayers = append(buildLayers, buildLayer)
		job := mountJob{
			ctx:   ctx,
			layer: buildLayer,
		}

		if err := pullWorker.Put(&job); err != nil {
			return errors.Wrap(err, "Put layer pull job to worker")
		}
	}

	// Build source layer to Nydus layer (bootstrap & blob) once the first source
	// layer be mounted in pull worker, and then put Nydus layer to the push worker,
	// it can be uploaded to remote registry
	for _, jobChan := range pullWorker.Waiter() {
		_job := <-jobChan
		if _job.Err() != nil {
			return errors.Wrap(_job.Err(), "Pull source layer")
		}
		job := _job.(*mountJob)

		// Skip building if we found the cache record in cache image
		if job.layer.Cached() {
			continue
		}

		// Build source layer to Nydus layer by invoking Nydus image builder
		if err := job.layer.Build(ctx); err != nil {
			return errors.Wrap(err, "Build source layer")
		}

		// Push Nydus layer (bootstrap & blob) to target registry
		pushWorker.Put(func() error {
			return job.layer.Push(ctx)
		})
	}

	// Wait all layer push job finish, then we can push image manifest on next
	if err := pushWorker.Wait(); err != nil {
		return errors.Wrap(err, "Push Nydus layer")
	}

	// Push OCI manifest, Nydus manifest and manifest index
	mm := &manifestManager{
		sourceProvider: cvt.SourceProvider,
		remote:         cvt.TargetRemote,
		backend:        _backend,
		multiPlatform:  cvt.MultiPlatform,
		dockerV2Format: cvt.DockerV2Format,
	}
	pushDone := logger.Log(ctx, "[MANI] Push manifest", nil)
	if err := mm.Push(ctx, buildLayers); err != nil {
		// When encounter http 400 error during pushing manifest to remote registry, means the
		// manifest is invalid, maybe the cache layer is not available in registry with a high
		// probability caused by registry GC, for example the cache image be overwritten by another
		// conversion progress, and the registry GC be triggered in the same time
		if cvt.CacheRemote != nil && strings.Contains(err.Error(), "unexpected status: 400") {
			logrus.Warn(errors.Wrap(err, "Push manifest"))
			return pushDone(errInvalidCache)
		}
		return pushDone(errors.Wrap(err, "Push target manifest"))
	}
	pushDone(nil)

	// Push Nydus cache image to remote registry
	if err := cg.Export(ctx, buildLayers); err != nil {
		return errors.Wrap(err, "Get cache record")
	}

	logrus.Infoln(fmt.Sprintf("Converted to %s", cvt.TargetRemote.Ref))

	return nil
}

// Convert converts source image to target (Nydus) image
func (cvt *Converter) Convert(ctx context.Context) error {
	if err := cvt.convert(ctx); err != nil {
		if errors.Is(err, errInvalidCache) {
			// Retry to convert without cache if the cache is invalid. we can't ensure the
			// cache is always valid during conversion progress, the registry will refuse
			// the Nydus manifest included invalid layer (purged by registry GC) pulled from
			// cache record, so retry without cache is a middle ground at this point
			cvt.CacheRemote = nil
			retryDone := logger.Log(ctx, "Retrying to convert without cache", nil)
			return retryDone(cvt.convert(ctx))
		}
		return errors.Wrap(err, "Failed to convert")
	}
	return nil
}
