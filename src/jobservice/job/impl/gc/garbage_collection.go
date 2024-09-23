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

package gc

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/errgroup"

	"github.com/goharbor/harbor/src/common/registryctl"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/artifactrash"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	"github.com/goharbor/harbor/src/pkg/blob"
	blobModels "github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/pkg/registry/interceptor/readonly"
	"github.com/goharbor/harbor/src/registryctl/client"
)

var (
	regCtlInit = registryctl.Init
	errGcStop  = errors.New("stopped")
)

const (
	dialConnectionTimeout = 30 * time.Second
	dialReadTimeout       = time.Minute + 10*time.Second
	dialWriteTimeout      = 10 * time.Second
	blobPrefix            = "blobs::*"
	repoPrefix            = "repository::*"
)

// GarbageCollector is the struct to run registry's garbage collection
type GarbageCollector struct {
	artCtl            artifact.Controller
	artrashMgr        artifactrash.Manager
	blobMgr           blob.Manager
	registryCtlClient client.Client
	logger            logger.Interface
	redisURL          string
	deleteUntagged    bool
	dryRun            bool
	// holds all of trashed artifacts' digest and repositories.
	// The source data of trashedArts is the table ArtifactTrash and it's only used as a dictionary by sweep when to delete a manifest.
	// As table blob has no repositories data, and the repositories are required when to delete a manifest, so use the table ArtifactTrash to capture them.
	trashedArts map[string][]model.ArtifactTrash
	// hold all of GC candidates(non-referenced blobs), it's captured by mark and consumed by sweep.
	deleteSet       []*blobModels.Blob
	timeWindowHours int64
	workers         int
}

// MaxFails implements the interface in job/Interface
func (gc *GarbageCollector) MaxFails() uint {
	return 1
}

// MaxCurrency is implementation of same method in Interface.
func (gc *GarbageCollector) MaxCurrency() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (gc *GarbageCollector) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (gc *GarbageCollector) Validate(_ job.Parameters) error {
	return nil
}

func (gc *GarbageCollector) init(ctx job.Context, params job.Parameters) error {
	regCtlInit()
	gc.logger = ctx.GetLogger()
	gc.deleteSet = make([]*blobModels.Blob, 0)
	gc.trashedArts = make(map[string][]model.ArtifactTrash, 0)

	// UT will use the mock client, ctl and mgr
	if os.Getenv("UTTEST") != "true" {
		gc.registryCtlClient = registryctl.RegistryCtlClient
		gc.artCtl = artifact.Ctl
		gc.artrashMgr = artifactrash.NewManager()
		gc.blobMgr = blob.NewManager()
	}
	if err := gc.registryCtlClient.Health(); err != nil {
		gc.logger.Errorf("failed to start gc as registry controller is unreachable: %v", err)
		return err
	}
	gc.parseParams(params)
	return nil
}

// parseParams set the parameters according to the GC API call.
func (gc *GarbageCollector) parseParams(params job.Parameters) {
	// redis url
	gc.redisURL = params["redis_url_reg"].(string)

	// delete untagged: default is to delete the untagged artifact
	gc.deleteUntagged = true
	deleteUntagged, exist := params["delete_untagged"]
	if exist {
		if untagged, ok := deleteUntagged.(bool); ok && !untagged {
			gc.deleteUntagged = untagged
		}
	}

	// time window: default is 2 hours, and for testing/debugging, it can be set to 0.
	gc.timeWindowHours = 2
	timeWindow, exist := params["time_window"]
	if exist {
		if timeWindow, ok := timeWindow.(float64); ok {
			gc.timeWindowHours = int64(timeWindow)
		}
	}

	// dry run: default is false. And for dry run we can have button in the UI.
	gc.dryRun = false
	dryRun, exist := params["dry_run"]
	if exist {
		if dryRun, ok := dryRun.(bool); ok && dryRun {
			gc.dryRun = dryRun
		}
	}

	// gc workers: default is 1. The business unit of removing blobs.
	gc.workers = 1
	ws, exist := params["workers"]
	if exist {
		if workers, ok := ws.(float64); ok {
			if int(workers) > 0 {
				gc.workers = int(workers)
			}
		}
	}

	gc.logger.Infof("Garbage Collection parameters: [delete_untagged: %t, dry_run: %t, time_window: %d, workers: %d]",
		gc.deleteUntagged, gc.dryRun, gc.timeWindowHours, gc.workers)
}

// Run implements the interface in job/Interface
func (gc *GarbageCollector) Run(ctx job.Context, params job.Parameters) error {
	if err := gc.init(ctx, params); err != nil {
		return err
	}

	gc.logger.Infof("start to run gc in job.")

	// mark
	if err := gc.mark(ctx); err != nil {
		if err == errGcStop {
			gc.logger.Info("received the stop signal, quit GC job.")
			return nil
		}
		gc.logger.Errorf("failed to execute GC job at mark phase, error: %v", err)
		return err
	}

	// sweep
	if !gc.dryRun {
		if err := gc.sweep(ctx); err != nil {
			if err == errGcStop {
				// we may already delete several artifacts before receiving the stop signal, so try to clean up the cache
				gc.logger.Info("received the stop signal, quit GC job after cleaning up the cache.")
				return gc.cleanCache(ctx.SystemContext())
			}
			gc.logger.Errorf("failed to execute GC job at sweep phase, error: %v", err)
			return err
		}

		if err := gc.cleanCache(ctx.SystemContext()); err != nil {
			return err
		}
	}
	gc.logger.Infof("success to run gc in job.")
	return nil
}

// mark
func (gc *GarbageCollector) mark(ctx job.Context) error {
	arts, err := gc.deletedArt(ctx)
	if err != nil {
		gc.logger.Errorf("failed to get deleted Artifacts in gc job, with error: %v", err)
		return err
	}
	// just log it, the job will continue to execute, the orphan blobs that created in the quota exceeding case can be removed.
	if len(arts) == 0 {
		gc.logger.Warning("no removed artifacts.")
	}
	gc.trashedArts = arts

	// get gc candidates, and set the repositories.
	// AS the reference count is calculated by joining table project_blob and blob, here needs to call removeUntaggedBlobs to remove these non-used blobs from table project_blob firstly.
	orphanBlobs, err := gc.markOrSweepUntaggedBlobs(ctx)
	if err != nil {
		return err
	}

	blobs, err := gc.uselessBlobs(ctx)
	if err != nil {
		gc.logger.Errorf("failed to get gc candidate: %v", err)
		return err
	}
	if len(orphanBlobs) != 0 {
		blobs = append(blobs, orphanBlobs...)
	}
	if len(blobs) == 0 {
		if err := saveGCRes(ctx, int64(0), int64(0), int64(0)); err != nil {
			gc.logger.Errorf("failed to save the garbage collection results, errMsg=%v", err)
		}
		gc.logger.Info("no need to execute GC as there is no non referenced artifacts.")
		return nil
	}

	// update delete status for the candidates.
	blobCt := 0
	mfCt := 0
	makeSize := int64(0)

	for _, blob := range blobs {
		if !gc.dryRun {
			if gc.shouldStop(ctx) {
				return errGcStop
			}
			blob.Status = blobModels.StatusDelete
			count, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
			if err != nil {
				gc.logger.Warningf("failed to mark gc candidate, skip it.: %s, error: %v", blob.Digest, err)
				continue
			}
			if count == 0 {
				gc.logger.Warningf("no blob found to mark gc candidate, skip it. ID:%d, digest:%s", blob.ID, blob.Digest)
				continue
			}
		}
		gc.logger.Infof("blob eligible for deletion: %s", blob.Digest)
		gc.deleteSet = append(gc.deleteSet, blob)
		if blob.IsManifest() {
			mfCt++
		} else {
			blobCt++
		}
		// do not count the foreign layer size as it's actually not in the storage.
		if !blob.IsForeignLayer() {
			makeSize = makeSize + blob.Size
		}
	}
	gc.logger.Infof("%d blobs and %d manifests eligible for deletion", blobCt, mfCt)
	gc.logger.Infof("The GC could free up %d MB space, the size is a rough estimation.", makeSize/1024/1024)

	if gc.dryRun {
		if err := saveGCRes(ctx, makeSize, int64(blobCt), int64(mfCt)); err != nil {
			gc.logger.Errorf("failed to save the garbage collection results, errMsg=%v", err)
		}
	}
	return nil
}

func (gc *GarbageCollector) sweep(ctx job.Context) error {
	gc.logger = ctx.GetLogger()
	sweepSize := int64(0)
	blobCnt := int64(0)
	mfCnt := int64(0)
	total := len(gc.deleteSet)

	// split the full set into pieces (count workers)
	if total <= 0 || gc.workers <= 0 {
		return nil
	}
	blobChunkSize, err := divide(total, gc.workers)
	if err != nil {
		return err
	}
	blobChunkCount := (total + blobChunkSize - 1) / blobChunkSize
	blobChunks := make([][]*blobModels.Blob, blobChunkCount)
	for i, start := 0, 0; i < blobChunkCount; i, start = i+1, start+blobChunkSize {
		end := start + blobChunkSize
		if end > total {
			end = total
		}
		blobChunks[i] = gc.deleteSet[start:end]
	}

	g := new(errgroup.Group)
	g.SetLimit(gc.workers)
	index := int64(0)
	for _, blobChunk := range blobChunks {
		blobChunk := blobChunk
		g.Go(func() error {
			uid := uuid.New().String()
			for _, blob := range blobChunk {
				if gc.shouldStop(ctx) {
					return errGcStop
				}

				localIndex := atomic.AddInt64(&index, 1)
				// set the status firstly, if the blob is updated by any HEAD/PUT request, it should be fail and skip.
				blob.Status = blobModels.StatusDeleting
				count, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
				if err != nil {
					gc.logger.Errorf("[%s][%d/%d] failed to mark gc candidate deleting, skip: %s, %s", uid, localIndex, total, blob.Digest, blob.Status)
					continue
				}
				if count == 0 {
					gc.logger.Warningf("[%s][%d/%d] no blob found to mark gc candidate deleting, ID:%d, digest:%s", uid, localIndex, total, blob.ID, blob.Digest)
					continue
				}

				// remove tags and revisions of a manifest
				skippedBlob := false
				if _, exist := gc.trashedArts[blob.Digest]; exist && blob.IsManifest() {
					for _, art := range gc.trashedArts[blob.Digest] {
						// Harbor cannot know the existing tags in the backend from its database, so let the v2 DELETE manifest to remove all of them.
						gc.logger.Infof("[%s][%d/%d] delete the manifest with registry v2 API: %s, %s, %s",
							uid, localIndex, total, art.RepositoryName, blob.ContentType, blob.Digest)
						if err := retry.Retry(func() error {
							return ignoreNotFound(func() error {
								err := v2DeleteManifest(art.RepositoryName, blob.Digest)
								// if the system is in read-only mode, return an Abort error to skip retrying
								if err == readonly.Err {
									return retry.Abort(err)
								}
								return err
							})
						}, retry.Callback(func(err error, sleep time.Duration) {
							gc.logger.Infof("[%s][%d/%d] failed to exec v2DeleteManifest, error: %v, will retry again after: %s", uid, localIndex, total, err, sleep)
						})); err != nil {
							gc.logger.Errorf("[%s][%d/%d] failed to delete manifest with v2 API, %s, %s, %v", uid, localIndex, total, art.RepositoryName, blob.Digest, err)
							if err := ignoreNotFound(func() error {
								return gc.markDeleteFailed(ctx, blob)
							}); err != nil {
								gc.logger.Errorf("[%s][%d/%d] failed to call gc.markDeleteFailed() after v2DeleteManifest() error out: %s, %v", uid, localIndex, total, blob.Digest, err)
								return err
							}
							// if the system is set to read-only mode, return directly
							if err == readonly.Err {
								return err
							}
							skippedBlob = true
							continue
						}
						// for manifest, it has to delete the revisions folder of each repository
						gc.logger.Infof("[%s][%d/%d] delete manifest from storage: %s", uid, localIndex, total, blob.Digest)
						if err := retry.Retry(func() error {
							return ignoreNotFound(func() error {
								err := gc.registryCtlClient.DeleteManifest(art.RepositoryName, blob.Digest)
								// if the system is in read-only mode, return an Abort error to skip retrying
								if err == readonly.Err {
									return retry.Abort(err)
								}
								return err
							})
						}, retry.Callback(func(err error, sleep time.Duration) {
							gc.logger.Infof("[%s][%d/%d] failed to exec DeleteManifest, error: %v, will retry again after: %s", uid, localIndex, total, err, sleep)
						})); err != nil {
							gc.logger.Errorf("[%s][%d/%d] failed to remove manifest from storage: %s, %s, errMsg=%v", uid, localIndex, total, art.RepositoryName, blob.Digest, err)
							if err := ignoreNotFound(func() error {
								return gc.markDeleteFailed(ctx, blob)
							}); err != nil {
								gc.logger.Errorf("[%s][%d/%d] failed to call gc.markDeleteFailed() after gc.registryCtlClient.DeleteManifest() error out: %s, %s, %v", uid, localIndex, total, art.RepositoryName, blob.Digest, err)
								return err
							}
							// if the system is set to read-only mode, return directly
							if err == readonly.Err {
								return err
							}
							skippedBlob = true
							continue
						}

						gc.logger.Infof("[%s][%d/%d] delete artifact blob record from database: %d, %s, %s", uid, localIndex, total, art.ID, art.RepositoryName, art.Digest)
						if err := ignoreNotFound(func() error {
							return gc.blobMgr.CleanupAssociationsForArtifact(ctx.SystemContext(), art.Digest)
						}); err != nil {
							gc.logger.Errorf("[%s][%d/%d] failed to call gc.blobMgr.CleanupAssociationsForArtifact(): %v, errMsg=%v", uid, localIndex, total, art.Digest, err)
							return err
						}

						gc.logger.Infof("[%s][%d/%d] delete artifact trash record from database: %d, %s, %s", uid, localIndex, total, art.ID, art.RepositoryName, art.Digest)
						if err := ignoreNotFound(func() error {
							return gc.artrashMgr.Delete(ctx.SystemContext(), art.ID)
						}); err != nil {
							gc.logger.Errorf("[%s][%d/%d] failed to call gc.artrashMgr.Delete(): %v, errMsg=%v", uid, localIndex, total, art.ID, err)
							return err
						}
					}
				}

				// skip deleting the blob if the manifest's tag/revision is not deleted
				if skippedBlob {
					continue
				}

				// delete all the blobs, which include config, layer and manifest
				// for the foreign layer, as it's not stored in the storage, no need to call the delete api and count size, but still have to delete the DB record.
				if !blob.IsForeignLayer() {
					gc.logger.Infof("[%s][%d/%d] delete blob from storage: %s", uid, localIndex, total, blob.Digest)
					if err := retry.Retry(func() error {
						return ignoreNotFound(func() error {
							err := gc.registryCtlClient.DeleteBlob(blob.Digest)
							// if the system is in read-only mode, return an Abort error to skip retrying
							if err == readonly.Err {
								return retry.Abort(err)
							}
							return err
						})
					}, retry.Callback(func(err error, sleep time.Duration) {
						gc.logger.Infof("[%s][%d/%d] failed to exec DeleteBlob, error: %v, will retry again after: %s", uid, localIndex, total, err, sleep)
					})); err != nil {
						gc.logger.Errorf("[%s][%d/%d] failed to delete blob from storage: %s, %s, errMsg=%v", uid, localIndex, total, blob.Digest, blob.Status, err)
						if err := ignoreNotFound(func() error {
							return gc.markDeleteFailed(ctx, blob)
						}); err != nil {
							gc.logger.Errorf("[%s][%d/%d] failed to call gc.markDeleteFailed() after gc.registryCtlClient.DeleteBlob() error out: %s, %v", uid, localIndex, total, blob.Digest, err)
							return err
						}
						// if the system is set to read-only mode, return directly
						if err == readonly.Err {
							return err
						}
						continue
					}
					atomic.AddInt64(&sweepSize, blob.Size)
				}

				gc.logger.Infof("[%s][%d/%d] delete blob record from database: %d, %s", uid, localIndex, total, blob.ID, blob.Digest)
				if err := ignoreNotFound(func() error {
					return gc.blobMgr.Delete(ctx.SystemContext(), blob.ID)
				}); err != nil {
					gc.logger.Errorf("[%s][%d/%d] failed to delete blob from database: %s, %s, errMsg=%v", uid, localIndex, total, blob.Digest, blob.Status, err)
					if err := ignoreNotFound(func() error {
						return gc.markDeleteFailed(ctx, blob)
					}); err != nil {
						gc.logger.Errorf("[%s][%d/%d] failed to call gc.markDeleteFailed() after gc.blobMgr.Delete() error out, %d, %s %v", uid, localIndex, total, blob.ID, blob.Digest, err)
						return err
					}
					return err
				}

				if blob.IsManifest() {
					atomic.AddInt64(&mfCnt, 1)
				} else {
					atomic.AddInt64(&blobCnt, 1)
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		gc.logger.Errorf("failed to execute mark(), error out, %v", err)
		return err
	}

	gc.logger.Infof("%d blobs and %d manifests are actually deleted", blobCnt, mfCnt)
	gc.logger.Infof("The GC job actual frees up %d MB space.", sweepSize/1024/1024)

	if err := saveGCRes(ctx, sweepSize, blobCnt, mfCnt); err != nil {
		gc.logger.Errorf("failed to save the garbage collection results, errMsg=%v", err)
	}

	return nil
}

// cleanCache is to clean the registry cache for GC.
// To do this is because the issue https://github.com/docker/distribution/issues/2094
func (gc *GarbageCollector) cleanCache(ctx context.Context) error {
	u, err := url.Parse(gc.redisURL)
	if err != nil {
		gc.logger.Errorf("failed to parse redis url %s, error: %v", gc.redisURL, err)
		return err
	}

	c, err := cache.New(u.Scheme, cache.Address(gc.redisURL))
	if err != nil {
		gc.logger.Errorf("failed to get redis client: %v", err)
		return err
	}

	// clean all keys in registry redis DB.

	// sample of keys in registry redis:
	// 1) "blobs::sha256:1a6fd470b9ce10849be79e99529a88371dff60c60aab424c077007f6979b4812"
	// 2) "repository::library/hello-world::blobs::sha256:4ab4c602aa5eed5528a6620ff18a1dc4faef0e1ab3a5eddeddb410714478c67f"
	patterns := []string{blobPrefix, repoPrefix}
	for _, pattern := range patterns {
		if err := delKeys(ctx, c, pattern); err != nil {
			gc.logger.Errorf("failed to clean registry cache %v, pattern %s", err, pattern)
			return err
		}
	}

	gc.logger.Info("cache clean up completed")

	return nil
}

// deletedArt contains the two parts of artifact, no actually deletion for dry run mode.
// 1, required part, the artifacts were removed from Harbor.
// 2, optional part, the untagged artifacts.
func (gc *GarbageCollector) deletedArt(ctx job.Context) (map[string][]model.ArtifactTrash, error) {
	if os.Getenv("UTTEST") == "true" {
		gc.logger = ctx.GetLogger()
	}

	// allTrashedArts contains the artifacts that actual removed and simulate removed(for dry run).
	allTrashedArts := make([]model.ArtifactTrash, 0)

	// artMap : map[digest : []ArtifactTrash list]
	artMap := make(map[string][]model.ArtifactTrash)
	// handle the optional ones, and the artifact controller will move them into trash.
	if gc.deleteUntagged {
		untaggedArts, err := gc.artCtl.List(ctx.SystemContext(), &q.Query{
			Keywords: map[string]interface{}{
				"Tags": "nil",
			},
		}, &artifact.Option{WithAccessory: true})
		if err != nil {
			return artMap, err
		}
		gc.logger.Info("start to delete untagged artifact (no actually deletion for dry-run mode)")
		for _, untagged := range untaggedArts {
			// for dryRun, just simulate the artifact deletion, move the artifact to artifact trash
			if gc.dryRun {
				var simulateDeletions []model.ArtifactTrash
				err = gc.artCtl.Walk(ctx.SystemContext(), untagged, func(a *artifact.Artifact) error {
					simulateDeletion := model.ArtifactTrash{
						MediaType:         a.MediaType,
						ManifestMediaType: a.ManifestMediaType,
						RepositoryName:    a.RepositoryName,
						Digest:            a.Digest,
						CreationTime:      time.Now(),
					}
					simulateDeletions = append(simulateDeletions, simulateDeletion)
					return nil
				}, &artifact.Option{WithAccessory: true})
				if err != nil {
					gc.logger.Errorf("walk the artifact %s failed, error: %v", untagged.Digest, err)
					continue
				}
				allTrashedArts = append(allTrashedArts, simulateDeletions...)
			} else {
				if gc.shouldStop(ctx) {
					return nil, errGcStop
				}
				if err := gc.artCtl.Delete(ctx.SystemContext(), untagged.ID); err != nil {
					// the failure ones can be GCed by the next execution
					gc.logger.Errorf("failed to delete untagged:%d artifact in DB, error, %v", untagged.ID, err)
					continue
				}
			}
			gc.logger.Infof("delete the untagged artifact: ProjectID:(%d)-RepositoryName(%s)-MediaType:(%s)-Digest:(%s)",
				untagged.ProjectID, untagged.RepositoryName, untagged.ManifestMediaType, untagged.Digest)
		}
		gc.logger.Info("end to delete untagged artifact (no actually deletion for dry-run mode)")
	}

	// filter gets all of actually deleted artifact, here do not need time window as the manifest candidate has to remove all of its reference.
	// For dryRun, no need to get the actual deletion artifacts since the return map is for the mark phase to call v2 remove manifest.
	if !gc.dryRun {
		actualDeletions, err := gc.artrashMgr.Filter(ctx.SystemContext(), 0)
		if err != nil {
			return artMap, err
		}
		allTrashedArts = append(allTrashedArts, actualDeletions...)
	}

	// group the deleted artifact by digest. The repositories of blob is needed when to delete as a manifest.
	if len(allTrashedArts) > 0 {
		gc.logger.Info("artifact trash candidates.")
		for _, art := range allTrashedArts {
			gc.logger.Info(art.String())
			_, exist := artMap[art.Digest]
			if !exist {
				artMap[art.Digest] = []model.ArtifactTrash{art}
			} else {
				repos := artMap[art.Digest]
				repos = append(repos, art)
				artMap[art.Digest] = repos
			}
		}
	}

	return artMap, nil
}

// mark or sweep the untagged blobs in each project, these blobs are not referenced by any manifest and will be cleaned by GC
// * dry-run, find and return the untagged blobs
// * non dry-run, remove the reference of the untagged blobs
func (gc *GarbageCollector) markOrSweepUntaggedBlobs(ctx job.Context) ([]*blobModels.Blob, error) {
	var orphanBlobs []*blobModels.Blob
	for result := range project.ListAll(ctx.SystemContext(), 50, nil, project.Metadata(false)) {
		if gc.shouldStop(ctx) {
			return nil, errGcStop
		}
		if result.Error != nil {
			gc.logger.Errorf("remove untagged blobs for all projects got error: %v", result.Error)
			continue
		}
		p := result.Data

		ps := 1000
		lastBlobID := int64(0)
		timeRG := q.Range{
			Max: time.Now().Add(-time.Duration(gc.timeWindowHours) * time.Hour).Format(time.RFC3339),
		}

		for {
			if gc.shouldStop(ctx) {
				gc.logger.Info("received the stop signal, quit GC job.")
				return nil, errGcStop
			}
			blobRG := q.Range{
				Min: lastBlobID,
			}
			query := &q.Query{
				Keywords: map[string]interface{}{
					"update_time": &timeRG,
					"projectID":   p.ProjectID,
					"id":          &blobRG,
				},
				PageNumber: 1,
				PageSize:   int64(ps),
				Sorts: []*q.Sort{
					q.NewSort("id", false),
				},
			}
			blobs, err := gc.blobMgr.List(ctx.SystemContext(), query)
			if err != nil {
				gc.logger.Errorf("failed to get blobs of project: %d, %v", p.ProjectID, err)
				break
			}
			if gc.dryRun {
				unassociated, err := gc.blobMgr.FindBlobsShouldUnassociatedWithProject(ctx.SystemContext(), p.ProjectID, blobs)
				if err != nil {
					gc.logger.Errorf("failed to find untagged blobs of project: %d, %v", p.ProjectID, err)
					break
				}
				orphanBlobs = append(orphanBlobs, unassociated...)
			} else {
				if err := gc.blobMgr.CleanupAssociationsForProject(ctx.SystemContext(), p.ProjectID, blobs); err != nil {
					gc.logger.Errorf("failed to clean untagged blobs of project: %d, %v", p.ProjectID, err)
					break
				}
			}
			if len(blobs) < ps {
				break
			}
			lastBlobID = blobs[len(blobs)-1].ID
		}
	}
	return orphanBlobs, nil
}

func (gc *GarbageCollector) uselessBlobs(ctx job.Context) ([]*blobModels.Blob, error) {
	var blobs []*blobModels.Blob
	var err error

	blobs, err = gc.blobMgr.UselessBlobs(ctx.SystemContext(), gc.timeWindowHours)
	if err != nil {
		gc.logger.Errorf("failed to get gc useless blobs: %v", err)
		return blobs, err
	}

	// For dryRun, it needs to append the blobs that are associated with untagged artifact.
	// Do it since the it doesn't remove the untagged artifact in dry run mode. All the blobs of untagged artifact are referenced by project,
	// so they cannot get by the above UselessBlobs method.
	// In dryRun mode, trashedArts only contains the mock deletion artifact.
	if gc.dryRun {
		for artDigest := range gc.trashedArts {
			artBlobs, err := gc.blobMgr.GetByArt(ctx.SystemContext(), artDigest)
			if err != nil {
				return blobs, err
			}
			blobs = append(blobs, artBlobs...)
		}
	}

	return blobs, err
}

// markDeleteFailed set the blob status to StatusDeleteFailed
func (gc *GarbageCollector) markDeleteFailed(ctx job.Context, blob *blobModels.Blob) error {
	blob.Status = blobModels.StatusDeleteFailed
	count, err := gc.blobMgr.UpdateBlobStatus(ctx.SystemContext(), blob)
	if err != nil {
		gc.logger.Errorf("failed to mark gc candidate delete failed: %s, %s", blob.Digest, blob.Status)
		return errors.Wrapf(err, "failed to mark gc candidate delete failed: %s, %s", blob.Digest, blob.Status)
	}
	if count == 0 {
		return errors.New(nil).WithMessage("no blob found to mark delete failed, ID:%d, digest:%s", blob.ID, blob.Digest).WithCode(errors.NotFoundCode)
	}
	return nil
}

func (gc *GarbageCollector) shouldStop(ctx job.Context) bool {
	opCmd, exit := ctx.OPCommand()
	if exit && opCmd.IsStop() {
		return true
	}
	return false
}

func saveGCRes(ctx job.Context, sweepSize, blobs, manifests int64) error {
	gcObj := struct {
		SweepSize int64 `json:"freed_space"`
		Blobs     int64 `json:"purged_blobs"`
		Manifests int64 `json:"purged_manifests"`
	}{
		SweepSize: sweepSize,
		Blobs:     blobs,
		Manifests: manifests,
	}
	c, err := json.Marshal(gcObj)
	if err != nil {
		return err
	}
	_ = ctx.Checkin(string(c))
	return nil
}
