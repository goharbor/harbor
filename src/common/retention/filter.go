// Copyright 2019 Project Harbor Authors
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

package retention

import (
	"time"

	"github.com/goharbor/harbor/src/common/models"
)

// FilterAction denotes the action a filter has taken for a given tag record
type FilterAction uint

const (
	// FilterActionKeep explicitly marks the tag as kept
	FilterActionKeep FilterAction = iota

	// FilterActionDelete explicitly marks the tag as deleted
	FilterActionDelete

	// FilterActionNoDecision passes the tag onto the next filter in the chain
	FilterActionNoDecision
)

// FilterMetadata defines the metadata needed to construct various Filter instances
type FilterMetadata struct {
	ID int64
	// The type of the filter to construct
	Type string
	// Parameters used to construct the filter
	Options map[string]interface{}
}

// TagRecord represents all pertinent metadata about a tag
type TagRecord struct {
	Project    *models.Project
	Repository *models.RepoRecord

	Name       string
	CreatedAt  time.Time
	LastPullAt time.Time
	Labels     []*models.Label
}

// Filter is a tag filter in a Retention Policy Filter Chain
type Filter interface {
	// Process determines what to do for a given tag record
	Process(tag *TagRecord) (FilterAction, error)

	// InitializeFor re-initializes the filter for tags from the specified project and repository
	//
	// Filters that maintain per-project or per-repository tracking metadata should reset it when
	// this method is called. Every call to `Process` will be for the same project and repo until
	// `InitializeFor` is called again.
	InitializeFor(project *models.Project, repo *models.RepoRecord)
}
