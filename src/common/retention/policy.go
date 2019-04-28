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

import "time"

// Scope identifies the scope of a specific retention policy
type Scope int

const (
	// ScopeServer identifies a retention policy defined server-wide
	ScopeServer Scope = iota

	// ScopeProject identifies a retention policy defined for a specific project
	ScopeProject

	// ScopeRepository identifies a retention policy defined for a specific repository
	ScopeRepository
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

// Policy contains an ordered slice of FilterMetadata used to construct filter chains
// during tag retention procession
type Policy struct {
	ID      int64
	Name    string
	Enabled bool

	Scope             Scope
	FallThroughAction FallThroughAction

	ProjectID    int64
	RepositoryID int64

	// When a filter chain is constructed for this policy, these filters will
	// be chained together in the order they appear in the slice
	Filters []*FilterMetadata

	CreatedAt time.Time
	UpdatedAt time.Time
}
