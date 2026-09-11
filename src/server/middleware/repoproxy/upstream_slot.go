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

package repoproxy

import (
	"context"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/redis"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/proxy/connection"
)

// upstreamAccess is how a request should proceed once acquireUpstreamSlot
// returns.
type upstreamAccess int

const (
	// fetchUpstream: the request holds an upstream connection slot and must
	// release it when done.
	fetchUpstream upstreamAccess = iota
	// serveLocal: another request brought the content into the local registry
	// while this one was waiting; serve it from there.
	serveLocal
)

const upstreamSlotPollInterval = time.Second

// acquireUpstreamSlot enforces the project's max_upstream_conn. With every
// slot taken it waits for one to free up or for existsLocally to report the
// content, for as long as the client stays connected. The returned release
// gives the slot back and must be called once the upstream work is done; it
// is a no-op when the project sets no limit or the content is served locally.
func acquireUpstreamSlot(ctx context.Context, p *proModels.Project, art lib.ArtifactInfo, existsLocally func() bool) (upstreamAccess, func(), error) {
	release := func() {}
	if p.MaxUpstreamConnection() <= 0 {
		return fetchUpstream, release, nil
	}
	client, err := redis.GetHarborClient()
	if err != nil {
		return fetchUpstream, release, errors.NewErrs(err)
	}
	key := upstreamRegistryConnectionKey(art)
	log.Debugf("upstream registry connection limit key: %s", key)
	acquire := func() bool {
		return connection.Limiter.Acquire(ctx, client, key, p.MaxUpstreamConnection())
	}
	access, err := waitForUpstreamSlot(ctx, acquire, existsLocally, upstreamSlotPollInterval)
	if err != nil || access == serveLocal {
		return access, release, err
	}
	// Background context: the request's context may already be canceled by
	// the time the slot is refreshed or released.
	stop := make(chan struct{})
	go keepSlotAlive(stop, slotRefreshInterval, func() {
		connection.Limiter.Refresh(context.Background(), client, key)
	})
	var once sync.Once
	release = func() {
		once.Do(func() {
			close(stop)
			connection.Limiter.Release(context.Background(), client, key)
		})
	}
	return fetchUpstream, release, nil
}

// waitForUpstreamSlot blocks until acquire succeeds, existsLocally reports
// the content or the client goes away, sleeping interval between attempts.
// There is no timeout of its own: a slot whose holder died expires on its own
// (see connection.SlotTTL), and a live holder is worth waiting for however
// long its pull takes.
func waitForUpstreamSlot(ctx context.Context, acquire func() bool, existsLocally func() bool, interval time.Duration) (upstreamAccess, error) {
	for {
		if acquire() {
			return fetchUpstream, nil
		}
		select {
		case <-ctx.Done():
			return fetchUpstream, ctx.Err()
		case <-time.After(interval):
		}
		if existsLocally() {
			return serveLocal, nil
		}
	}
}

// slotRefreshInterval leaves two missed refreshes before a live holder's slot
// expires under it.
const slotRefreshInterval = connection.SlotTTL / 3

// keepSlotAlive calls refresh every interval until stop is closed.
func keepSlotAlive(stop <-chan struct{}, interval time.Duration, refresh func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			refresh()
		}
	}
}
