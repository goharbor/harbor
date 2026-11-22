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

package proxy

import (
	"context"
	"time"

	"github.com/docker/distribution"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/retry"
)

const (
	// localPushAttempts is the total number of attempts, the first try plus the
	// retries, made when pushing content into the local registry.
	localPushAttempts = 3
	// localPushInitialInterval and localPushMaxInterval bound the backoff
	// between those attempts.
	localPushInitialInterval = 200 * time.Millisecond
	localPushMaxInterval     = 400 * time.Millisecond
)

// retryLocalPush runs push until it succeeds or localPushAttempts attempts have been made.
//
// Caching content pulled through a proxy-cache project means pushing it into the
// local registry, and neither push is a single request: PushBlob opens an upload
// with a POST and then PUTs to the location that returned, and PushManifest is a
// PUT that core follows with a read-back GET to build the artifact record. When the
// registry runs as several replicas behind a non-sticky service, those requests can
// land on different replicas, and the later one fails with BLOB_UPLOAD_UNKNOWN or
// MANIFEST_UNKNOWN although nothing is actually wrong. A single attempt turns that,
// and any other transient local error, into content that stays out of the cache
// until some client happens to pull it again. Both operations are keyed by digest
// and so are safe to repeat.
func retryLocalPush(ctx context.Context, what string, push func() error) error {
	attempt := 0
	return retry.Retry(func() error {
		// Callers currently pass a context detached from the pulling client's
		// request, since caching outlives that request, but stop retrying if a
		// cancellable one is ever passed and it is done.
		if err := ctx.Err(); err != nil {
			return retry.Abort(err)
		}

		attempt++
		err := push()
		if err == nil {
			if attempt > 1 {
				log.Infof("pushed %s to local registry on attempt %d/%d", what, attempt, localPushAttempts)
			}
			return nil
		}
		if attempt >= localPushAttempts {
			log.Errorf("giving up pushing %s to local registry after %d attempts, error: %v", what, attempt, err)
			return retry.Abort(err)
		}
		return err
	},
		retry.InitialInterval(localPushInitialInterval),
		retry.MaxInterval(localPushMaxInterval),
		// retry.Retry only checks its timeout between attempts and never interrupts
		// a push already in flight, so this is a backstop rather than a deadline:
		// the attempt count is what bounds the loop. Leave room for every attempt to
		// take the full registry client timeout that already bounds each one.
		retry.Timeout(localPushAttempts*config.RegistryHTTPClientTimeout()),
		retry.Callback(func(err error, sleep time.Duration) {
			log.Warningf("failed to push %s to local registry (attempt %d/%d), retrying in %v, error: %v",
				what, attempt, localPushAttempts, sleep, err)
		}),
	)
}

// putBlobToLocal streams a blob from the remote registry into the local one.
//
// The reader is re-opened for every attempt: a failed push may already have
// consumed part of it, and replaying a drained reader would upload a truncated
// blob instead of failing cleanly.
func putBlobToLocal(ctx context.Context, local localInterface, remoteRepo, localRepo string, desc distribution.Descriptor, r RemoteInterface) error {
	log.Debugf("Put blob to local registry!, sourceRepo:%v, localRepo:%v, digest: %v", remoteRepo, localRepo, desc.Digest)
	return retryLocalPush(ctx, "blob "+localRepo+":"+string(desc.Digest), func() error {
		_, bReader, err := r.BlobReader(remoteRepo, string(desc.Digest))
		if err != nil {
			log.Errorf("failed to create blob reader, error %v", err)
			return err
		}
		defer bReader.Close()
		return local.PushBlob(localRepo, desc, bReader)
	})
}

// pushManifestToLocal pushes a manifest into the local registry, retrying on failure.
func pushManifestToLocal(ctx context.Context, local localInterface, repo, ref string, man distribution.Manifest) error {
	return retryLocalPush(ctx, "manifest "+repo+":"+ref, func() error {
		return local.PushManifest(repo, ref, man)
	})
}
