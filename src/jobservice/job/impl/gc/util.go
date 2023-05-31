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
	"fmt"

	"github.com/gomodule/redigo/redis"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/registry"
)

// delKeys ...
func delKeys(con redis.Conn, pattern string) error {
	iter := 0
	keys := make([]string, 0)
	for {
		arr, err := redis.Values(con.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return fmt.Errorf("error retrieving '%s' keys: %s", pattern, err)
		}
		iter, err = redis.Int(arr[0], nil)
		if err != nil {
			return fmt.Errorf("unexpected type for Int, got type %T", err)
		}
		k, err := redis.Strings(arr[1], nil)
		if err != nil {
			return fmt.Errorf("converts an array command reply to a []string %v", err)
		}
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}
	for _, key := range keys {
		_, err := con.Do("DEL", key)
		if err != nil {
			return fmt.Errorf("failed to clean registry cache %v", err)
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
