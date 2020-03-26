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

package event

import "github.com/goharbor/harbor/src/replication/model"

// const definitions
const (
	EventTypeArtifactPush   = "artifact_push"
	EventTypeArtifactDelete = "artifact_delete"
	EventTypeTagDelete      = "tag_delete"
	EventTypeChartUpload    = "chart_upload"
	EventTypeChartDelete    = "chart_delete"
)

// Event is the model that defines the image/chart pull/push event
type Event struct {
	Type     string
	Resource *model.Resource
}
