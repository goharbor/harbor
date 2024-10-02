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

package internal

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/controller/artifact"
	sbomprocessor "github.com/goharbor/harbor/src/controller/artifact/processor/sbom"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	pkgArt "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/scan/report"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
	"github.com/goharbor/harbor/src/pkg/scan/sbom"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	// defaultAsyncFlushDuration is the default flush interval.
	defaultAsyncFlushDuration = 10 * time.Second
)

var (
	asyncFlushDuration time.Duration
)

func init() {
	// get async flush duration from env, if not provide,
	// use default value: 10*time.second
	envDuration := os.Getenv("ARTIFACT_PULL_ASYNC_FLUSH_DURATION")
	if len(envDuration) == 0 {
		// use default value
		asyncFlushDuration = defaultAsyncFlushDuration
	} else {
		duration, err := strconv.ParseInt(envDuration, 10, 64)
		if err != nil {
			log.Warningf("error to parse ARTIFACT_PULL_ASYNC_FLUSH_DURATION: %v, will use default value: %v", err, defaultAsyncFlushDuration)
			asyncFlushDuration = defaultAsyncFlushDuration
		} else {
			asyncFlushDuration = time.Duration(duration) * time.Second
		}
	}
}

// ArtifactEventHandler preprocess artifact event data
type ArtifactEventHandler struct {
	// execMgr for managing executions
	execMgr task.ExecutionManager
	// reportMgr for managing scan reports
	reportMgr report.Manager
	// sbomReportMgr
	sbomReportMgr sbom.Manager
	// artMgr for managing artifacts
	artMgr pkgArt.Manager

	once sync.Once
	// pullCountStore caches the pull count group by repository
	// map[repositoryID]counts
	pullCountStore map[int64]uint64
	// pullCountLock mutex for pullCountStore
	pullCountLock sync.Mutex
	// pullTimeStore caches the latest pull time group by artifact
	// map[artifactID:tagName]time
	pullTimeStore map[string]time.Time
	// pullTimeLock mutex for pullTimeStore
	pullTimeLock sync.Mutex
}

// Name ...
func (a *ArtifactEventHandler) Name() string {
	return "InternalArtifact"
}

// Handle ...
func (a *ArtifactEventHandler) Handle(ctx context.Context, value interface{}) error {
	switch v := value.(type) {
	case *event.PullArtifactEvent:
		return a.onPull(ctx, v.ArtifactEvent)
	case *event.PushArtifactEvent:
		return a.onPush(ctx, v.ArtifactEvent)
	case *event.DeleteArtifactEvent:
		return a.onDelete(ctx, v.ArtifactEvent)
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	return nil
}

// IsStateful ...
func (a *ArtifactEventHandler) IsStateful() bool {
	return false
}

func (a *ArtifactEventHandler) onPull(ctx context.Context, event *event.ArtifactEvent) error {
	if config.ScannerSkipUpdatePullTime(ctx) && isScannerUser(ctx, event) {
		return nil
	}
	// if duration is equal to 0 or negative, keep original sync mode.
	if asyncFlushDuration <= 0 {
		var tagName string
		if len(event.Tags) > 0 {
			tagName = event.Tags[0]
		}

		if !config.PullTimeUpdateDisable(ctx) {
			a.syncFlushPullTime(ctx, event.Artifact.ID, tagName, time.Now())
		}

		if !config.PullCountUpdateDisable(ctx) {
			a.syncFlushPullCount(ctx, event.Artifact.RepositoryID, 1)
		}

		return nil
	}

	// async mode, update in cache firstly and flush to db by workers periodically.
	a.once.Do(func() {
		if !config.PullTimeUpdateDisable(ctx) {
			a.pullTimeStore = make(map[string]time.Time)
			go a.asyncFlushPullTime(orm.Context())
		}

		if !config.PullCountUpdateDisable(ctx) {
			a.pullCountStore = make(map[int64]uint64)
			go a.asyncFlushPullCount(orm.Context())
		}
	})

	if !config.PullTimeUpdateDisable(ctx) {
		a.updatePullTimeInCache(ctx, event)
	}

	if !config.PullCountUpdateDisable(ctx) {
		a.addPullCountInCache(ctx, event)
	}

	return nil
}

func (a *ArtifactEventHandler) updatePullTimeInCache(_ context.Context, event *event.ArtifactEvent) {
	var tagName string
	if len(event.Tags) != 0 {
		tagName = event.Tags[0]
	}

	key := fmt.Sprintf("%d:%s", event.Artifact.ID, tagName)

	a.pullTimeLock.Lock()
	defer a.pullTimeLock.Unlock()

	a.pullTimeStore[key] = time.Now()
}

func (a *ArtifactEventHandler) addPullCountInCache(_ context.Context, event *event.ArtifactEvent) {
	a.pullCountLock.Lock()
	defer a.pullCountLock.Unlock()

	a.pullCountStore[event.Artifact.RepositoryID] = a.pullCountStore[event.Artifact.RepositoryID] + 1
}

func (a *ArtifactEventHandler) syncFlushPullTime(ctx context.Context, artifactID int64, tagName string, time time.Time) {
	var tagID int64

	if tagName != "" {
		tags, err := tag.Ctl.List(ctx, q.New(
			map[string]interface{}{
				"ArtifactID": artifactID,
				"Name":       tagName,
			}), nil)
		if err != nil {
			log.Warningf("failed to list tags when to update pull time, %v", err)
		} else {
			if len(tags) != 0 {
				tagID = tags[0].ID
			}
		}
	}

	if err := artifact.Ctl.UpdatePullTime(ctx, artifactID, tagID, time); err != nil {
		log.Warningf("failed to update pull time for artifact %d, %v", artifactID, err)
	}
}

func (a *ArtifactEventHandler) syncFlushPullCount(ctx context.Context, repositoryID int64, count uint64) {
	if err := repository.Ctl.AddPullCount(ctx, repositoryID, count); err != nil {
		log.Warningf("failed to add pull count repository %d, %v", repositoryID, err)
	}
}

func (a *ArtifactEventHandler) asyncFlushPullTime(ctx context.Context) {
	for {
		<-time.After(asyncFlushDuration)
		a.pullTimeLock.Lock()

		for key, time := range a.pullTimeStore {
			keys := strings.Split(key, ":")
			artifactID, err := strconv.ParseInt(keys[0], 10, 64)
			if err != nil {
				log.Warningf("failed to parse artifact id %s, %v", key, err)
				continue
			}

			var tagName string
			if len(keys) > 1 && keys[1] != "" {
				tagName = keys[1]
			}

			a.syncFlushPullTime(ctx, artifactID, tagName, time)
		}

		a.pullTimeStore = make(map[string]time.Time)
		a.pullTimeLock.Unlock()
	}
}

func (a *ArtifactEventHandler) asyncFlushPullCount(ctx context.Context) {
	for {
		<-time.After(asyncFlushDuration)
		a.pullCountLock.Lock()

		for repositoryID, count := range a.pullCountStore {
			a.syncFlushPullCount(ctx, repositoryID, count)
		}

		a.pullCountStore = make(map[int64]uint64)
		a.pullCountLock.Unlock()
	}
}

func (a *ArtifactEventHandler) onPush(ctx context.Context, event *event.ArtifactEvent) error {
	go func() {
		if event.Operator != "" {
			ctx = context.WithValue(ctx, operator.ContextKey{}, event.Operator)
		}

		if err := autoScan(ctx, &artifact.Artifact{Artifact: *event.Artifact}, event.Tags...); err != nil {
			log.Errorf("scan artifact %s@%s failed, error: %v", event.Artifact.RepositoryName, event.Artifact.Digest, err)
		}

		log.Debugf("auto generate sbom is triggered for artifact event %+v", event)
		if err := autoGenSBOM(ctx, &artifact.Artifact{Artifact: *event.Artifact}); err != nil {
			log.Errorf("generate sbom for artifact %s@%s failed, error: %v", event.Artifact.RepositoryName, event.Artifact.Digest, err)
		}
	}()

	return nil
}

func (a *ArtifactEventHandler) onDelete(ctx context.Context, event *event.ArtifactEvent) error {
	execMgr := task.ExecMgr
	reportMgr := report.Mgr
	artMgr := pkg.ArtifactMgr
	// for UT mock
	if a.execMgr != nil {
		execMgr = a.execMgr
	}
	if a.reportMgr != nil {
		reportMgr = a.reportMgr
	}
	if a.artMgr != nil {
		artMgr = a.artMgr
	}

	ids := []int64{event.Artifact.ID}
	digests := []string{event.Artifact.Digest}
	if len(event.Artifact.References) > 0 {
		for _, ref := range event.Artifact.References {
			ids = append(ids, ref.ChildID)
			digests = append(digests, ref.ChildDigest)
		}
	}
	// check if the digest also referenced by other artifacts, should exclude it to delete the scan report if still referenced by others.
	unrefDigests := []string{}
	for _, digest := range digests {
		// with the base=* to query all artifacts includes untagged and references
		count, err := artMgr.Count(ctx, q.New(q.KeyWords{"digest": digest, "base": "*"}))
		if err != nil {
			log.Errorf("failed to count the artifact with the digest %s, error: %v", digest, err)
			continue
		}

		if count == 0 {
			unrefDigests = append(unrefDigests, digest)
		}
	}
	// clean up the scan executions of this artifact and it's references by id
	log.Debugf("delete the associated scan executions of artifacts %v as the artifacts have been deleted", ids)
	for _, id := range ids {
		if err := execMgr.DeleteByVendor(ctx, job.ImageScanJobVendorType, id); err != nil {
			log.Errorf("failed to delete scan executions of artifact %d, error: %v", id, err)
		}
	}

	// clean up the scan reports of this artifact and it's references by digest
	log.Debugf("delete the associated scan reports of artifacts %v as the artifacts have been deleted", unrefDigests)
	if err := reportMgr.DeleteByDigests(ctx, unrefDigests...); err != nil {
		log.Errorf("failed to delete scan reports of artifact %v, error: %v", unrefDigests, err)
	}

	// delete sbom_report when the subject artifact is deleted
	if err := sbom.Mgr.DeleteByArtifactID(ctx, event.Artifact.ID); err != nil {
		log.Errorf("failed to delete sbom reports of artifact ID %v, error: %v", event.Artifact.ID, err)
	}

	// delete sbom_report when the accessory artifact is deleted
	if event.Artifact.Type == sbomprocessor.ArtifactTypeSBOM && len(event.Artifact.Digest) > 0 {
		if err := sbom.Mgr.DeleteByExtraAttr(ctx, v1.MimeTypeSBOMReport, "sbom_digest", event.Artifact.Digest); err != nil {
			log.Errorf("failed to delete sbom reports of with sbom digest %v, error: %v", event.Artifact.Digest, err)
		}
	}
	return nil
}

// isScannerUser check if the current user is a scanner user by its prefix
// usually a scanner user should be named like `robot$<projectName>+<Scanner UUID (8byte)>-<Scanner Name>-<UUID>`
// verify it by the prefix `robot$<projectName>+<Scanner UUID (8byte)>`
func isScannerUser(ctx context.Context, event *event.ArtifactEvent) bool {
	if len(event.Operator) == 0 {
		return false
	}
	robotPrefix := config.RobotPrefix(ctx)
	scannerPrefix := config.ScannerRobotPrefix(ctx)
	prefix := fmt.Sprintf("%s%s+%s", robotPrefix, parseProjectName(event.Repository), scannerPrefix)
	return strings.HasPrefix(event.Operator, prefix)
}

func parseProjectName(repoName string) string {
	if strings.Contains(repoName, "/") {
		return strings.Split(repoName, "/")[0]
	}
	return ""
}
