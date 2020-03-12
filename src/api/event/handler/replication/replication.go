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

package replication

import (
	"github.com/goharbor/harbor/src/api/event"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/replication"
	repevent "github.com/goharbor/harbor/src/replication/event"
	"github.com/goharbor/harbor/src/replication/model"
	"strconv"
)

// Handler ...
type Handler struct {
}

// Handle ...
func (r *Handler) Handle(value interface{}) error {
	pushArtEvent, ok := value.(*event.PushArtifactEvent)
	if ok {
		return r.handlePushArtifact(pushArtEvent)
	}
	deleteArtEvent, ok := value.(*event.DeleteArtifactEvent)
	if ok {
		return r.handleDeleteArtifact(deleteArtEvent)
	}
	createTagEvent, ok := value.(*event.CreateTagEvent)
	if ok {
		return r.handleCreateTag(createTagEvent)
	}
	deleteTagEvent, ok := value.(*event.DeleteTagEvent)
	if ok {
		return r.handleDeleteTag(deleteTagEvent)
	}
	return nil
}

// IsStateful ...
func (r *Handler) IsStateful() bool {
	return false
}

func (r *Handler) handlePushArtifact(event *event.PushArtifactEvent) error {
	art := event.Artifact
	public := false
	project, err := project.Mgr.Get(art.ProjectID)
	if err == nil && project != nil {
		public = project.IsPublic()
	} else {
		log.Error(err)
	}
	project.IsPublic()
	e := &repevent.Event{
		Type: repevent.EventTypeArtifactPush,
		Resource: &model.Resource{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: event.Repository,
					Metadata: map[string]interface{}{
						"public": strconv.FormatBool(public),
					},
				},
				Artifacts: []*model.Artifact{
					{
						Type:   art.Type,
						Digest: art.Digest,
						Tags:   event.Tags,
					}},
			},
		},
	}
	return replication.EventHandler.Handle(e)
}

func (r *Handler) handleDeleteArtifact(event *event.DeleteArtifactEvent) error {
	art := event.Artifact
	e := &repevent.Event{
		Type: repevent.EventTypeArtifactDelete,
		Resource: &model.Resource{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: event.Repository,
				},
				Artifacts: []*model.Artifact{
					{
						Type:   art.Type,
						Digest: art.Digest,
						Tags:   event.Tags,
					}},
			},
			Deleted: true,
		},
	}
	return replication.EventHandler.Handle(e)
}

func (r *Handler) handleCreateTag(event *event.CreateTagEvent) error {
	art := event.AttachedArtifact
	public := false
	project, err := project.Mgr.Get(art.ProjectID)
	if err == nil && project != nil {
		public = project.IsPublic()
	} else {
		log.Error(err)
	}
	project.IsPublic()
	e := &repevent.Event{
		Type: repevent.EventTypeArtifactPush,
		Resource: &model.Resource{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: event.Repository,
					Metadata: map[string]interface{}{
						"public": strconv.FormatBool(public),
					},
				},
				Artifacts: []*model.Artifact{
					{
						Type:   art.Type,
						Digest: art.Digest,
						Tags:   []string{event.Tag},
					}},
			},
		},
	}
	return replication.EventHandler.Handle(e)
}

func (r *Handler) handleDeleteTag(event *event.DeleteTagEvent) error {
	art := event.AttachedArtifact
	e := &repevent.Event{
		Type: repevent.EventTypeTagDelete,
		Resource: &model.Resource{
			Type: model.ResourceTypeArtifact,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: event.Repository,
				},
				Artifacts: []*model.Artifact{
					{
						Type:   art.Type,
						Digest: art.Digest,
						Tags:   []string{event.Tag},
					}},
			},
			Deleted:     true,
			IsDeleteTag: true,
		},
	}
	return replication.EventHandler.Handle(e)
}
