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

package artifact

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/q"
	"time"
)

var (
	// Ctl is a global artifact controller instance
	Ctl = NewController()
)

// Controller defines the operations related with artifacts and tags
type Controller interface {
	// Ensure the artifact specified by the digest exists under the repository,
	// creates it if it doesn't exist. If tags are provided, ensure they exist
	// and are attached to the artifact. If the tags don't exist, create them first
	Ensure(ctx context.Context, repository *models.RepoRecord, digest string, tags ...string) (err error)
	// List artifacts according to the query and option
	List(ctx context.Context, query *q.Query, option *Option) (total int64, artifacts []*Artifact, err error)
	// Get the artifact specified by ID
	Get(ctx context.Context, id int64, option *Option) (artifact *Artifact, err error)
	// Delete the artifact specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// DeleteTag deletes the tag specified by ID
	DeleteTag(ctx context.Context, id int64) (err error)
	// UpdatePullTime updates the pull time for the artifact. If the tag is provides, update the pull
	// time of the tag as well
	UpdatePullTime(ctx context.Context, artifactID int64, tag string, time time.Time) (err error)
	// GetSubResource returns the sub resource of the artifact
	// The sub resource is different according to the artifact type:
	// build history for image; values.yaml, readme and dependencies for chart, etc
	GetSubResource(ctx context.Context, artifactID int64, resource string) (*Resource, error)
	// TODO move this to GC controller?
	// Prune removes the useless artifact records. The underlying registry data will
	// be removed during garbage collection
	// Prune(ctx context.Context, option *Option) error
}

// NewController creates an instance of the default artifact controller
func NewController() Controller {
	// TODO implement
	return nil
}

// As a redis lock is applied during the artifact pushing, we do not to handle the concurrent issues
// for artifacts and tags
