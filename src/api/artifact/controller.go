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

package artifact

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/api/artifact/abstractor"
	// registry image resolvers
	_ "github.com/goharbor/harbor/src/api/artifact/abstractor/resolver/image"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/tag"
	tm "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"time"
)

var (
	// Ctl is a global artifact controller instance
	Ctl = NewController()
)

// Controller defines the operations related with artifacts and tags
type Controller interface {
	// Ensure the artifact specified by the digest exists under the repository,
	// creates it if it doesn't exist. If tags are provided, ensure they exist
	// and are attached to the artifact. If the tags don't exist, create them first.
	// The "created" will be set as true when the artifact is created
	Ensure(ctx context.Context, repositoryID int64, digest string, tags ...string) (created bool, id int64, err error)
	// List artifacts according to the query, specify the properties returned with option
	List(ctx context.Context, query *q.Query, option *Option) (total int64, artifacts []*Artifact, err error)
	// Get the artifact specified by ID, specify the properties returned with option
	Get(ctx context.Context, id int64, option *Option) (artifact *Artifact, err error)
	// Delete the artifact specified by ID. All tags attached to the artifact are deleted as well
	Delete(ctx context.Context, id int64) (err error)
	// Tags returns the tags according to the query, specify the properties returned with option
	Tags(ctx context.Context, query *q.Query, option *TagOption) (total int64, tags []*Tag, err error)
	// DeleteTag deletes the tag specified by tagID
	DeleteTag(ctx context.Context, tagID int64) (err error)
	// UpdatePullTime updates the pull time for the artifact. If the tagID is provides, update the pull
	// time of the tag as well
	UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) (err error)
	// GetSubResource returns the sub resource of the artifact
	// The sub resource is different according to the artifact type:
	// build history for image; values.yaml, readme and dependencies for chart, etc
	GetSubResource(ctx context.Context, artifactID int64, resource string) (*Resource, error)
	// TODO move this to GC controller?
	// Prune removes the useless artifact records. The underlying registry data will
	// be removed during garbage collection
	// Prune(ctx context.Context, option *Option) error
}

// NewController creates an instance of the default artifact controller
func NewController() Controller {
	return &controller{
		repoMgr:    repository.Mgr,
		artMgr:     artifact.Mgr,
		tagMgr:     tag.Mgr,
		abstractor: abstractor.NewAbstractor(),
	}
}

// TODO concurrency summary

type controller struct {
	repoMgr    repository.Manager
	artMgr     artifact.Manager
	tagMgr     tag.Manager
	abstractor abstractor.Abstractor
}

func (c *controller) Ensure(ctx context.Context, repositoryID int64, digest string, tags ...string) (bool, int64, error) {
	created, id, err := c.ensureArtifact(ctx, repositoryID, digest)
	if err != nil {
		return false, 0, err
	}
	for _, tag := range tags {
		if err = c.ensureTag(ctx, repositoryID, id, tag); err != nil {
			return false, 0, err
		}
	}
	return created, id, nil
}

// ensure the artifact exists under the repository, create it if doesn't exist.
func (c *controller) ensureArtifact(ctx context.Context, repositoryID int64, digest string) (bool, int64, error) {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"digest":        digest,
		},
	}
	_, artifacts, err := c.artMgr.List(ctx, query)
	if err != nil {
		return false, 0, err
	}
	// the artifact already exists under the repository, return directly
	if len(artifacts) > 0 {
		return false, artifacts[0].ID, nil
	}

	// the artifact doesn't exist under the repository, create it first
	repository, err := c.repoMgr.Get(ctx, repositoryID)
	if err != nil {
		return false, 0, err
	}
	artifact := &artifact.Artifact{
		ProjectID:    repository.ProjectID,
		RepositoryID: repositoryID,
		Digest:       digest,
		PushTime:     time.Now(),
	}
	// abstract the specific information for the artifact
	if err = c.abstractor.Abstract(ctx, artifact); err != nil {
		return false, 0, err
	}

	// create it
	id, err := c.artMgr.Create(ctx, artifact)
	if err != nil {
		// if got conflict error, try to get the artifact again
		if ierror.IsConflictErr(err) {
			_, artifacts, err = c.artMgr.List(ctx, query)
			if err != nil {
				return false, 0, err
			}
			if len(artifacts) > 0 {
				return false, artifacts[0].ID, nil
			}
		}
		return false, 0, err
	}
	return true, id, nil
}

func (c *controller) ensureTag(ctx context.Context, repositoryID, artifactID int64, name string) error {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"name":          name,
		},
	}
	_, tags, err := c.tagMgr.List(ctx, query)
	if err != nil {
		return err
	}
	// the tag already exists under the repository
	if len(tags) > 0 {
		tag := tags[0]
		// the tag already exists under the repository and is attached to the artifact, return directly
		if tag.ArtifactID == artifactID {
			return nil
		}
		// the tag exists under the repository, but it is attached to other artifact
		// update it to point to the provided artifact
		tag.ArtifactID = artifactID
		tag.PushTime = time.Now()
		return c.tagMgr.Update(ctx, tag, "ArtifactID", "PushTime")
	}
	// the tag doesn't exist under the repository, create it
	_, err = c.tagMgr.Create(ctx, &tm.Tag{
		RepositoryID: repositoryID,
		ArtifactID:   artifactID,
		Name:         name,
		PushTime:     time.Now(),
	})
	// ignore the conflict error
	if err != nil && ierror.IsConflictErr(err) {
		return nil
	}
	return err
}

func (c *controller) List(ctx context.Context, query *q.Query, option *Option) (int64, []*Artifact, error) {
	total, arts, err := c.artMgr.List(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	var artifacts []*Artifact
	for _, art := range arts {
		artifacts = append(artifacts, c.assembleArtifact(ctx, art, option))
	}
	return total, artifacts, nil
}
func (c *controller) Get(ctx context.Context, id int64, option *Option) (*Artifact, error) {
	art, err := c.artMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.assembleArtifact(ctx, art, option), nil
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	// delete artifact first in case the artifact is referenced by other artifact
	if err := c.artMgr.Delete(ctx, id); err != nil {
		return err
	}

	// delete all tags that attached to the artifact
	_, tags, err := c.tagMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"artifact_id": id,
		},
	})
	if err != nil {
		return err
	}
	for _, tag := range tags {
		if err = c.DeleteTag(ctx, tag.ID); err != nil {
			return err
		}
	}
	// TODO fire delete artifact event
	return nil
}
func (c *controller) Tags(ctx context.Context, query *q.Query, option *TagOption) (int64, []*Tag, error) {
	total, tgs, err := c.tagMgr.List(ctx, query)
	if err != nil {
		return 0, nil, err
	}
	var tags []*Tag
	for _, tg := range tgs {
		tags = append(tags, c.assembleTag(ctx, tg, option))
	}
	return total, tags, nil
}

func (c *controller) DeleteTag(ctx context.Context, tagID int64) error {
	// immutable checking is covered in middleware
	// TODO check signature
	// TODO delete label
	// TODO fire delete tag event
	return c.tagMgr.Delete(ctx, tagID)
}

func (c *controller) UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) error {
	tag, err := c.tagMgr.Get(ctx, tagID)
	if err != nil {
		return err
	}
	if tag.ArtifactID != artifactID {
		return fmt.Errorf("tag %d isn't attached to artifact %d", tagID, artifactID)
	}
	if err := c.artMgr.UpdatePullTime(ctx, artifactID, time); err != nil {
		return err
	}
	return c.tagMgr.Update(ctx, &tm.Tag{
		ID: tagID,
	}, "PullTime")
}
func (c *controller) GetSubResource(ctx context.Context, artifactID int64, resource string) (*Resource, error) {
	// TODO implement
	return nil, nil
}

// assemble several part into a single artifact
func (c *controller) assembleArtifact(ctx context.Context, art *artifact.Artifact, option *Option) *Artifact {
	artifact := &Artifact{
		Artifact: *art,
	}
	if option == nil {
		return artifact
	}
	// populate tags
	if option.WithTag {
		_, tgs, err := c.tagMgr.List(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"artifact_id": artifact.ID,
			},
		})
		if err == nil {
			// assemble tags
			for _, tg := range tgs {
				artifact.Tags = append(artifact.Tags, c.assembleTag(ctx, tg, option.TagOption))
			}
		} else {
			log.Errorf("failed to list tag of artifact %d: %v", artifact.ID, err)
		}
	}
	// TODO populate other properties: scan, signature etc.
	return artifact
}

// assemble several part into a single tag
func (c *controller) assembleTag(ctx context.Context, tag *tm.Tag, option *TagOption) *Tag {
	t := &Tag{
		Tag: *tag,
	}
	if option == nil {
		return t
	}
	// TODO populate label, signature, immutable status for tag
	return t
}
