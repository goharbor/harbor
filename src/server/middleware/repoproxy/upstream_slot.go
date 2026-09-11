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

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/redis"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/proxy/connection"
)

// acquireUpstreamSlot enforces the project's max_upstream_conn. The returned
// release gives the slot back and must be called once the upstream work is
// done; it is a no-op when the project sets no limit.
func acquireUpstreamSlot(ctx context.Context, p *proModels.Project, art lib.ArtifactInfo) (func(), error) {
	release := func() {}
	if p.MaxUpstreamConnection() <= 0 {
		return release, nil
	}
	client, err := redis.GetHarborClient()
	if err != nil {
		return release, errors.NewErrs(err)
	}
	key := upstreamRegistryConnectionKey(art)
	log.Debugf("upstream registry connection limit key: %s", key)
	if !connection.Limiter.Acquire(ctx, client, key, p.MaxUpstreamConnection()) {
		log.Infof("current connection exceed max connections to upstream registry, key: %s", key)
		return release, tooManyRequestsError
	}
	// Background context: the request's context may already be canceled by
	// the time the slot is released.
	release = func() { connection.Limiter.Release(context.Background(), client, key) }
	return release, nil
}
