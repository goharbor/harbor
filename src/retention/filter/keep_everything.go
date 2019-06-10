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

// TypeKeepEverything tells the filter builder to construct a KeepEverything filter for the associated metadata
const TypeKeepEverything = "retention:filter:keep_everything"

// KeepEverything is an empty struct that implements retention.Filter. It takes no arguments.
type KeepEverything struct{}

// InitializeFor on a KeepEverything Filter does nothing
func (*KeepEverything) InitializeFor(project *models.Project, repo *models.RepoRecord) {}

// Process for the DeleteEverything Filter simply returns retention.FilterActionKeep
func (*KeepEverything) Process(tag *retention.TagRecord) (retention.FilterAction, error) {
	return retention.FilterActionKeep, nil
}
