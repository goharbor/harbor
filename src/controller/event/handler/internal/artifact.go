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
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
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

// Handler preprocess artifact event data
type Handler struct {
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
func (a *Handler) Name() string {
	return "InternalArtifact"
}

// Handle ...
func (a *Handler) Handle(ctx context.Context, value interface{}) error {
	switch v := value.(type) {
	case *event.PullArtifactEvent:
		return a.onPull(ctx, v.ArtifactEvent)
	case *event.PushArtifactEvent:
		return a.onPush(ctx, v.ArtifactEvent)
	default:
		log.Errorf("Can not handler this event type! %#v", v)
	}
	return nil
}

// IsStateful ...
func (a *Handler) IsStateful() bool {
	return false
}

func (a *Handler) onPull(ctx context.Context, event *event.ArtifactEvent) error {
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

func (a *Handler) updatePullTimeInCache(ctx context.Context, event *event.ArtifactEvent) {
	var tagName string
	if len(event.Tags) != 0 {
		tagName = event.Tags[0]
	}

	key := fmt.Sprintf("%d:%s", event.Artifact.ID, tagName)

	a.pullTimeLock.Lock()
	defer a.pullTimeLock.Unlock()

	a.pullTimeStore[key] = time.Now()
}

func (a *Handler) addPullCountInCache(ctx context.Context, event *event.ArtifactEvent) {
	a.pullCountLock.Lock()
	defer a.pullCountLock.Unlock()

	a.pullCountStore[event.Artifact.RepositoryID] = a.pullCountStore[event.Artifact.RepositoryID] + 1
}

func (a *Handler) syncFlushPullTime(ctx context.Context, artifactID int64, tagName string, time time.Time) {
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

func (a *Handler) syncFlushPullCount(ctx context.Context, repositoryID int64, count uint64) {
	if err := repository.Ctl.AddPullCount(ctx, repositoryID, count); err != nil {
		log.Warningf("failed to add pull count repository %d, %v", repositoryID, err)
	}
}

func (a *Handler) asyncFlushPullTime(ctx context.Context) {
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

func (a *Handler) asyncFlushPullCount(ctx context.Context) {
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

func (a *Handler) onPush(ctx context.Context, event *event.ArtifactEvent) error {
	go func() {
		if err := autoScan(ctx, &artifact.Artifact{Artifact: *event.Artifact}, event.Tags...); err != nil {
			log.Errorf("scan artifact %s@%s failed, error: %v", event.Artifact.RepositoryName, event.Artifact.Digest, err)
		}
	}()

	return nil
}
