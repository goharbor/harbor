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

package cached

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	// Resource type definitions
	// ResourceTypeArtifact defines artifact type.
	ResourceTypeArtifact = "artifact"
	// ResourceTypeProject defines project type.
	ResourceTypeProject = "project"
	// ResourceTypeProject defines project metadata type.
	ResourceTypeProjectMeta = "project_metadata"
	// ResourceTypeRepository defines repository type.
	ResourceTypeRepository = "repository"
	// ResourceTypeManifest defines manifest type.
	ResourceTypeManifest = "manifest"
)

// Manager is the interface for resource cache manager.
// Provides common interfaces for admin to view and manage resource cache.
type Manager interface {
	//  ResourceType returns the resource type.
	//  eg. artifact、project、tag、repository
	ResourceType(ctx context.Context) string
	// CountCache returns current this resource occupied cache count.
	CountCache(ctx context.Context) (int64, error)
	// DeleteCache deletes specific cache by key.
	DeleteCache(ctx context.Context, key string) error
	// FlushAll flush this resource's all cache.
	FlushAll(ctx context.Context) error

	// TODO for more extensions like metrics.
}

// ObjectKey normalizes cache object key.
type ObjectKey struct {
	// namespace as group or prefix, eg. artifact:id
	namespace string
}

// NewObjectKey returns object key with namespace.
func NewObjectKey(ns string) *ObjectKey {
	return &ObjectKey{namespace: ns}
}

// Format formats fields to string.
// eg. namespace: 'artifact'
// Format("id", 100, "digest", "aaa"): "artifact:id:100:digest:aaa"
func (ok *ObjectKey) Format(keysAndValues ...interface{}) (string, error) {
	// keysAndValues must be paired.
	if len(keysAndValues)%2 != 0 {
		return "", errors.Errorf("invalid keysAndValues: %v", keysAndValues...)
	}

	s := ok.namespace
	for i := 0; i < len(keysAndValues); i++ {
		// even is key
		if i%2 == 0 {
			key, match := keysAndValues[i].(string)
			if !match {
				return "", errors.Errorf("key must be string, invalid key type: %#v", keysAndValues[i])
			}

			s += fmt.Sprintf(":%s", key)
		} else {
			switch keysAndValues[i].(type) {
			case int, int16, int32, int64:
				s += fmt.Sprintf(":%d", keysAndValues[i])
			case string:
				s += fmt.Sprintf(":%s", keysAndValues[i])
			default:
				return "", errors.Errorf("unsupported value type: %#v", keysAndValues[i])
			}
		}
	}

	return s, nil
}
