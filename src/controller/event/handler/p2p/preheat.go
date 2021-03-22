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

package p2p

import (
	"context"

	"github.com/goharbor/harbor/src/controller/tag"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/p2p/preheat"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

// Handler ...
type Handler struct {
}

// Name ...
func (p *Handler) Name() string {
	return "P2PPreheat"
}

// Handle ...
func (p *Handler) Handle(ctx context.Context, value interface{}) error {
	switch value.(type) {
	case *event.PushArtifactEvent:
		pushArtEvent, _ := value.(*event.PushArtifactEvent)
		return p.handlePushArtifact(ctx, pushArtEvent)
	case *event.ScanImageEvent:
		scanImageEvent, _ := value.(*event.ScanImageEvent)
		return p.handleImageScanned(ctx, scanImageEvent)
	case *event.ArtifactLabeledEvent:
		artifactLabeledEvent, _ := value.(*event.ArtifactLabeledEvent)
		return p.handleArtifactLabeled(ctx, artifactLabeledEvent)
	default:
		return errors.New("unsupported type")
	}
}

// IsStateful ...
func (p *Handler) IsStateful() bool {
	return false
}

func (p *Handler) handlePushArtifact(ctx context.Context, event *event.PushArtifactEvent) error {
	if event.Artifact.Type != image.ArtifactTypeImage {
		return nil
	}

	// NOTES: So far, we only support artifact with tags
	if len(event.Tags) == 0 {
		return nil
	}

	log.Debugf("preheat: artifact pushed %s:%s@%s", event.Artifact.RepositoryName, event.Tags, event.Artifact.Digest)

	art, err := artifact.Ctl.Get(ctx, event.Artifact.ID, &artifact.Option{
		WithTag:   true,
		WithLabel: true,
	})
	if err != nil {
		return err
	}

	// Only with the pushed tags, ignore other tags
	pt := make([]*tag.Tag, 0)
	for _, tg := range art.Tags {
		if tg.Name == event.Tags[0] {
			pt = append(pt, tg)
			break
		}
	}
	art.Tags = pt

	_, err = preheat.Enf.PreheatArtifact(ctx, art)
	return err
}

func (p *Handler) handleImageScanned(ctx context.Context, event *event.ScanImageEvent) error {
	// TODO: If the scan is targeting an manifest list, here the artifacts we get are all the children
	//  artifacts of the manifest list. The children artifacts are high probably untagged ones that
	//  will be definitely ignored by the tag filter. We need to find a way to resolve this issue.
	log.Debugf("preheat: image scanned %s:%s", event.Artifact.Repository, event.Artifact.Tag)
	art, err := artifact.Ctl.GetByReference(ctx, event.Artifact.Repository, event.Artifact.Digest,
		&artifact.Option{
			WithTag:   true,
			WithLabel: true,
		})
	if err != nil {
		return err
	}
	_, err = preheat.Enf.PreheatArtifact(ctx, art)
	return err
}

func (p *Handler) handleArtifactLabeled(ctx context.Context, event *event.ArtifactLabeledEvent) error {
	art, err := artifact.Ctl.Get(ctx, event.ArtifactID, &artifact.Option{
		WithTag:   true,
		WithLabel: true,
	})

	if err != nil {
		return err
	}

	// Only care image at this moment
	if art.Type != image.ArtifactTypeImage {
		return nil
	}
	log.Debugf("preheat: artifact labeled %s:%s", art.Artifact.RepositoryName, art.Artifact.Digest)

	_, err = preheat.Enf.PreheatArtifact(ctx, art)
	return err
}
