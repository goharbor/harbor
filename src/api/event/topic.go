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

import (
	"github.com/goharbor/harbor/src/pkg/artifact"
	"time"
)

// the event consumers can refer to this file to find all topics and the corresponding event structures

// const definition
const (
	TopicCreateProject    = "CREATE_PROJECT"
	TopicDeleteProject    = "DELETE_PROJECT"
	TopicDeleteRepository = "DELETE_REPOSITORY"
	TopicPushArtifact     = "PUSH_ARTIFACT"
	TopicPullArtifact     = "PULL_ARTIFACT"
	TopicDeleteArtifact   = "DELETE_ARTIFACT"
	TopicCreateTag        = "CREATE_TAG"
	TopicDeleteTag        = "DELETE_TAG"
)

// CreateProjectEvent is the creating project event
type CreateProjectEvent struct {
	Project  string
	Operator string
	OccurAt  time.Time
}

// DeleteProjectEvent is the deleting project event
type DeleteProjectEvent struct {
	Project  string
	Operator string
	OccurAt  time.Time
}

// DeleteRepositoryEvent is the deleting repository event
type DeleteRepositoryEvent struct {
	Repository string
	Operator   string
	OccurAt    time.Time
}

// PushArtifactEvent is the pushing artifact event
type PushArtifactEvent struct {
	Repository string
	Artifact   *artifact.Artifact
	Tag        string // when the artifact is pushed by digest, the tag here will be null
	Operator   string
	OccurAt    time.Time
}

// PullArtifactEvent is the pulling artifact event
type PullArtifactEvent struct {
	Repository string
	Artifact   *artifact.Artifact
	Tag        string // when the artifact is pulled by digest, the tag here will be null
	Operator   string
	OccurAt    time.Time
}

// DeleteArtifactEvent is the deleting artifact event
type DeleteArtifactEvent struct {
	Repository string
	Artifact   *artifact.Artifact
	Tags       []string // all the tags that attached to the deleted artifact
	Operator   string
	OccurAt    time.Time
}

// CreateTagEvent is the creating tag event
type CreateTagEvent struct {
	Repository       string
	Tag              string
	AttachedArtifact *artifact.Artifact
	Operator         string
	OccurAt          time.Time
}

// DeleteTagEvent is the deleting tag event
type DeleteTagEvent struct {
	Repository       string
	Tag              string
	AttachedArtifact *artifact.Artifact
	Operator         string
	OccurAt          time.Time
}
