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

package quota

import (
	"context"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/gtask"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
)

// Deferred, coalesced quota usage refresh.
//
// On the push path the usage figure is maintained by the synchronous
// reservation itself (RefreshForProjectMiddleware only runs on deletes).
// When the reservation is skipped - see Request()'s unlimited fast path -
// something else must keep the usage figure current: the request path
// marks the project dirty in memory (no database access), and a background
// task recomputes and stores the usage once per dirty project per
// interval, regardless of how many requests marked it.
//
// The flush recomputes usage from the database (CalculateUsage) instead of
// accumulating in-memory deltas, so it is idempotent: a mark lost to a
// process restart, or a concurrent flush from another core replica, can
// only delay convergence - never corrupt the stored value. A failed flush
// re-marks the project and is retried on the next interval.
//
// The whole mechanism - and the unlimited fast path in Request() that
// relies on it - is opt-in: it activates only when the
// QUOTA_ASYNC_REFRESH_DURATION env var is set to a positive number of
// seconds (the flush interval). Without it, quota behavior is unchanged.
// Setting the env var also switches RefreshMiddleware to the coalesced
// path (see AsyncRefreshEnabled).
const defaultDeferredRefreshInterval = 10 * time.Second

const asyncRefreshDurationEnv = "QUOTA_ASYNC_REFRESH_DURATION"

var asyncRefreshConfigured bool

// perRefreshTimeout bounds a single project's recompute inside the flush
// loop; see refreshDirty.
const perRefreshTimeout = 30 * time.Second

var dirty = struct {
	sync.Mutex
	keys map[refKey]struct{}
}{keys: map[refKey]struct{}{}}

type refKey struct {
	reference   string
	referenceID string
}

func init() {
	interval, configured := parseRefreshInterval(os.Getenv(asyncRefreshDurationEnv))
	asyncRefreshConfigured = configured
	gtask.DefaultPool().AddTask(flushDirtyQuota, interval)
}

// parseRefreshInterval turns the QUOTA_ASYNC_REFRESH_DURATION value into the
// flush interval. It returns the default interval and configured=false for
// an empty or invalid value; an invalid value is additionally logged.
func parseRefreshInterval(env string) (time.Duration, bool) {
	if len(env) == 0 {
		return defaultDeferredRefreshInterval, false
	}
	seconds, err := strconv.ParseInt(env, 10, 64)
	// the upper bound keeps seconds*time.Second representable in a
	// time.Duration; beyond it the multiplication overflows negative
	// and the flush loop would spin without sleeping
	if err != nil || seconds <= 0 || seconds > int64(math.MaxInt64/int64(time.Second)) {
		log.Warningf("invalid %s %q, using default flush interval %s", asyncRefreshDurationEnv, env, defaultDeferredRefreshInterval)
		return defaultDeferredRefreshInterval, false
	}
	return time.Duration(seconds) * time.Second, true
}

// AsyncRefreshEnabled returns true when QUOTA_ASYNC_REFRESH_DURATION is set
// to a positive number of seconds; only then does Request() skip the
// reservation for unlimited quotas, and RefreshMiddleware coalesces its
// per-request full recomputes through the deferred flush as well.
func AsyncRefreshEnabled() bool {
	return asyncRefreshConfigured
}

// MarkRefresh records that the usage of the reference object changed and
// must be recomputed by the next flush. It performs no database access.
func MarkRefresh(reference, referenceID string) {
	dirty.Lock()
	dirty.keys[refKey{reference: reference, referenceID: referenceID}] = struct{}{}
	dirty.Unlock()
}

// flushDirtyQuota recomputes and stores the usage of every project marked
// dirty since the previous flush.
func flushDirtyQuota(_ context.Context) {
	refreshDirty(orm.Context())
}

func refreshDirty(ctx context.Context) {
	dirty.Lock()
	keys := dirty.keys
	dirty.keys = make(map[refKey]struct{}, len(keys))
	dirty.Unlock()

	for key := range keys {
		// Bound each recompute: Refresh retries CAS conflicts for up to
		// defaultRetryTimeout, and one project stuck in conflicts must not
		// hold the single flush goroutine and starve the other dirty
		// projects. A timed-out refresh is re-marked and tried next flush.
		refreshCtx, cancel := context.WithTimeout(ctx, perRefreshTimeout)
		// IgnoreLimitation: the recompute records the truth about used
		// storage; enforcement happens at reserve time.
		err := Ctl.Refresh(refreshCtx, key.reference, key.referenceID, IgnoreLimitation(true))
		cancel()
		switch {
		case err == nil:
		case errors.IsNotFoundErr(err):
			// the reference was deleted after being marked - nothing left
			// to refresh, do not re-mark or the key retries forever
		default:
			log.Errorf("deferred quota refresh for %s %s failed, will retry next flush, error: %v", key.reference, key.referenceID, err)
			MarkRefresh(key.reference, key.referenceID)
		}
	}
}
