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
	"context"
	"strconv"

	"github.com/goharbor/harbor/src/controller/event"
	repevent "github.com/goharbor/harbor/src/controller/event/handler/replication/event"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Handler ...
type Handler struct {
}

// Name ...
func (r *Handler) Name() string {
	return "Replication"
}

// Handle ...
func (r *Handler) Handle(ctx context.Context, value interface{}) error {
	pushArtEvent, ok := value.(*event.PushArtifactEvent)
	if ok {
		return r.handlePushArtifact(ctx, pushArtEvent)
	}
	deleteArtEvent, ok := value.(*event.DeleteArtifactEvent)
	if ok {
		return r.handleDeleteArtifact(ctx, deleteArtEvent)
	}
	createTagEvent, ok := value.(*event.CreateTagEvent)
	if ok {
		return r.handleCreateTag(ctx, createTagEvent)
	}
	deleteTagEvent, ok := value.(*event.DeleteTagEvent)
	if ok {
		return r.handleDeleteTag(ctx, deleteTagEvent)
	}
	return nil
}

// IsStateful ...
func (r *Handler) IsStateful() bool {
	return false
}

func (r *Handler) handlePushArtifact(ctx context.Context, event *event.PushArtifactEvent) error {
	art := event.Artifact
	public := false
	prj, err := project.Ctl.Get(orm.Context(), art.ProjectID, project.Metadata(true))
	if err != nil {
		log.Errorf("failed to get project: %d, error: %v", art.ProjectID, err)
		return err
	}
	public = prj.IsPublic()

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
	return repevent.Handle(ctx, e)
}

func (r *Handler) handleDeleteArtifact(ctx context.Context, event *event.DeleteArtifactEvent) error {
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
	return repevent.Handle(ctx, e)
}

func (r *Handler) handleCreateTag(ctx context.Context, event *event.CreateTagEvent) error {
	art := event.AttachedArtifact
	public := false
	prj, err := project.Ctl.Get(orm.Context(), art.ProjectID, project.Metadata(true))
	if err != nil {
		log.Errorf("failed to get project: %d, error: %v", art.ProjectID, err)
		return err
	}
	public = prj.IsPublic()

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
	return repevent.Handle(ctx, e)
}

func (r *Handler) handleDeleteTag(ctx context.Context, event *event.DeleteTagEvent) error {
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
	return repevent.Handle(ctx, e)
}
