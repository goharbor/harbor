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

package models

import (
	"time"

	"github.com/goharbor/harbor/src/core/api"
)

// TagRecord represents all pertinent metadata about a tag
type TagRecord struct {
	Project    *Project
	Repository *RepoRecord
	Tag        *api.TagResp
}

// TagAction records when a filter takes an action upon a tag
type TagAction struct {
	// The tag the action applies to
	Target *TagRecord
	// The filter that took this action
	ActingFilter Filter
}

// Filter is a tag filter in a Retention Policy Filter Chain
type Filter interface {
	// Process takes tags from the input channel and writes them to one of the three output channels.
	// Tags are written to toKeep if the tags should be explicitly kept by the Filter
	// Tags are written to toDelete if the tags should be explicitly deleted by the Filter
	// Tags are written to next if the retention policy does not apply to the provided tag
	//      or if the policy does not care if the tag is kept or deleted
	//
	// Filters do not own any of the provided channels and should **not** close them under any circumstance
	Process(input <-chan *TagRecord, toKeep, toDelete chan<- *TagAction, next chan<- *TagRecord) error

	// InitializeFor re-initializes the filter for tags from the specified project and repository
	//
	// Filters that maintain per-project or per-repository tracking metadata should reset it when
	// this method is called.
	InitializeFor(project *Project, repo *RepoRecord)
}

// FilterMetadata defines the metadata needed to construct various Filter instances
type FilterMetadata struct {
	ID int64
	// The type of the filter to construct
	Type string
	// Parameters used to construct the filter
	Options map[string]interface{}
}

// RetentionScope identifies the scope of a specific retention policy
type RetentionScope int

const (
	// RetentionScopeServer identifies a retention policy defined server-wide
	RetentionScopeServer RetentionScope = iota
	// RetentionScopeProject identifies a retention policy defined for a specific project
	RetentionScopeProject
	// RetentionScopeRepository identifies a retention policy defined for a specific repository
	RetentionScopeRepository
)

// FallThroughAction determines what action the policy should take when a tag has not
// been explicitly kept nor explicitly deleted by all filters in the filter chain
type FallThroughAction int

const (
	// KeepExtraTags indicates that tags which are not explicitly kept or deleted are implicitly kept
	KeepExtraTags FallThroughAction = iota
	// DeleteExtraTags indicates that tags which are not explicitly kept or deleted are implicitly deleted
	DeleteExtraTags
)

// RetentionPolicy contains an ordered slice of FilterMetadata used to construct filter chains
// during tag retention procession
type RetentionPolicy struct {
	ID      int64
	Name    string
	Enabled bool

	Scope             RetentionScope
	FallThroughAction FallThroughAction

	ProjectID    int64
	RepositoryID int64

	// When a filter chain is constructed for this policy, these filters will
	// be chained together in the order they appear in the slice
	Filters []*FilterMetadata

	CreatedAt time.Time
	UpdatedAt time.Time
}
