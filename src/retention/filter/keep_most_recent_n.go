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

package filter

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/retention"
)

const (
	// FilterTypeKeepMostRecentN tells the filter builder to construct a keepMostRecentN filter for associated metadata
	FilterTypeKeepMostRecentN = "retention:filter:keep_most_recent_n"

	// MetaDataKeyN is the ke in the metadata map for the `n` value
	MetaDataKeyN = "n"
)

type keepMostRecentN struct {
	N         int
	keptSoFar int
}

// NewKeepMostRecentN constructs a new filter for the provided metadata
func NewKeepMostRecentN(metadata map[string]interface{}) (*keepMostRecentN, error) {
	if n, ok := metadata[MetaDataKeyN]; ok {
		if intN, ok := n.(int); ok && intN > 0 {
			return &keepMostRecentN{N: intN}, nil
		} else if ok {
			return nil, ErrInvalidMetadata(MetaDataKeyN, "cannot be negative")
		}

		return nil, ErrWrongMetadataType(MetaDataKeyN, "int")
	}

	return nil, ErrMissingMetadata(MetaDataKeyN)
}

func (f *keepMostRecentN) InitializeFor(project *models.Project, repo *models.RepoRecord) {
	f.keptSoFar = 0
}

func (f *keepMostRecentN) Process(tag *retention.TagRecord) (retention.FilterAction, error) {
	f.keptSoFar++
	if f.keptSoFar > f.N {
		return retention.FilterActionDelete, nil
	}

	return retention.FilterActionKeep, nil
}
