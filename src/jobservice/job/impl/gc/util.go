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

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// delKeys ...
func delKeys(ctx context.Context, c cache.Cache, pattern string) error {
	iter, err := c.Scan(ctx, pattern)
	if err != nil {
		return errors.Wrap(err, "failed to scan keys")
	}

	for iter.Next(ctx) {
		if err := c.Delete(ctx, iter.Val()); err != nil {
			return errors.Wrap(err, "failed to clean registry cache")
		}
	}

	return nil
}

// v2DeleteManifest calls the registry API to remove manifest
func v2DeleteManifest(repository, digest string) error {
	exist, _, err := registry.Cli.ManifestExist(repository, digest)
	if err != nil {
		return err
	}
	// it could be happened at remove manifest success but fail to delete harbor DB.
	// when the GC job executes again, the manifest should not exist.
	if !exist {
		return nil
	}
	return registry.Cli.DeleteManifest(repository, digest)
}

// ignoreNotFound ignores the NotFoundErr error
func ignoreNotFound(f func() error) error {
	if err := f(); err != nil && !errors.IsNotFoundErr(err) {
		return err
	}
	return nil
}

// divide if it is divisible, it gives the quotient. if it's not, it gives the remainder.
func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("the divided cannot be zero")
	}

	quotient := a / b
	remainder := a % b

	if quotient == 0 {
		return remainder, nil
	}
	return quotient, nil
}
