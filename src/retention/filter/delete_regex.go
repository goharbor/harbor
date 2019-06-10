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
	"regexp"

	"github.com/goharbor/harbor/src/common/models"

	"github.com/goharbor/harbor/src/common/retention"
)

const (
	// TypeDeleteRegex tells the filter builder to construct a DeleteRegex filter for the associated metadata
	TypeDeleteRegex = "retention:filter:delete_regex"
)

type deleteRegex struct {
	match *regexp.Regexp
}

// NewDeleteRegex constructs a filter implementing retention.Filter. It accepts a single argument, "match",
// which must be a string consisting of a valid regular expression.
func NewDeleteRegex(metadata map[string]interface{}) (retention.Filter, error) {
	if raw, ok := metadata[MetaDataKeyMatch]; ok {
		if rawString, ok := raw.(string); ok {
			regex, err := regexp.Compile(rawString)
			if err == nil {
				return &deleteRegex{match: regex}, nil
			}

			return nil, ErrInvalidMetadata(MetaDataKeyMatch, err.Error())
		}

		return nil, ErrWrongMetadataType(MetaDataKeyMatch, "string")
	}

	return nil, ErrMissingMetadata(MetaDataKeyMatch)
}

// InitializeFor for a DeleteRegex filter does nothing
func (f *deleteRegex) InitializeFor(project *models.Project, repo *models.RepoRecord) {}

// Process for a DeleteRegex filter returns retention.FilterActionDelete if the tag name matches the
// regular expression in "match". Otherwise, it returns FilterActionNoDecision
func (f *deleteRegex) Process(tag *retention.TagRecord) (retention.FilterAction, error) {
	if f.match.MatchString(tag.Name) {
		return retention.FilterActionDelete, nil
	}

	return retention.FilterActionNoDecision, nil
}
