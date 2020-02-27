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
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver"
	"github.com/goharbor/harbor/src/api/artifact/descriptor"
	"github.com/goharbor/harbor/src/api/tag"
	"github.com/goharbor/harbor/src/internal"
	"github.com/goharbor/harbor/src/pkg/artifactrash"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
	"github.com/goharbor/harbor/src/pkg/label"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/pkg/signature"
	"github.com/opencontainers/go-digest"
	"strings"
	"time"
	// registry image resolvers
	_ "github.com/goharbor/harbor/src/api/artifact/abstractor/resolver/image"
	// register chart resolver
	_ "github.com/goharbor/harbor/src/api/artifact/abstractor/resolver/chart"
	// register CNAB resolver
	_ "github.com/goharbor/harbor/src/api/artifact/abstractor/resolver/cnab"
	"github.com/goharbor/harbor/src/common/utils/log"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
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
	Ensure(ctx context.Context, repository, digest string, tags ...string) (created bool, id int64, err error)
	// Count returns the total count of artifacts according to the query.
	// The artifacts that referenced by others and without tags are not counted
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List artifacts according to the query, specify the properties returned with option
	// The artifacts that referenced by others and without tags are not returned
	List(ctx context.Context, query *q.Query, option *Option) (artifacts []*Artifact, err error)
	// Get the artifact specified by ID, specify the properties returned with option
	Get(ctx context.Context, id int64, option *Option) (artifact *Artifact, err error)
	// Get the artifact specified by repository name and reference, the reference can be tag or digest,
	// specify the properties returned with option
	GetByReference(ctx context.Context, repository, reference string, option *Option) (artifact *Artifact, err error)
	// Delete the artifact specified by artifact ID
	Delete(ctx context.Context, id int64) (err error)
	// Copy the artifact specified by "srcRepo" and "reference" into the repository specified by "dstRepo"
	Copy(ctx context.Context, srcRepo, reference, dstRepo string) (id int64, err error)
	// ListTags lists the tags according to the query, specify the properties returned with option
	ListTags(ctx context.Context, query *q.Query, option *tag.Option) (tags []*tag.Tag, err error)
	// CreateTag creates a tag
	CreateTag(ctx context.Context, tag *tag.Tag) (id int64, err error)
	// DeleteTag deletes the tag specified by tagID
	DeleteTag(ctx context.Context, tagID int64) (err error)
	// UpdatePullTime updates the pull time for the artifact. If the tagID is provides, update the pull
	// time of the tag as well
	UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) (err error)
	// GetAddition returns the addition of the artifact.
	// The addition is different according to the artifact type:
	// build history for image; values.yaml, readme and dependencies for chart, etc
	GetAddition(ctx context.Context, artifactID int64, additionType string) (addition *resolver.Addition, err error)
	// AddLabel to the specified artifact
	AddLabel(ctx context.Context, artifactID int64, labelID int64) (err error)
	// RemoveLabel from the specified artifact
	RemoveLabel(ctx context.Context, artifactID int64, labelID int64) (err error)
}

// NewController creates an instance of the default artifact controller
func NewController() Controller {
	return &controller{
		tagCtl:       tag.Ctl,
		repoMgr:      repository.Mgr,
		artMgr:       artifact.Mgr,
		artrashMgr:   artifactrash.Mgr,
		blobMgr:      blob.Mgr,
		sigMgr:       signature.GetManager(),
		labelMgr:     label.Mgr,
		abstractor:   abstractor.NewAbstractor(),
		immutableMtr: rule.NewRuleMatcher(),
		regCli:       registry.Cli,
	}
}

// TODO concurrency summary

type controller struct {
	tagCtl       tag.Controller
	repoMgr      repository.Manager
	artMgr       artifact.Manager
	artrashMgr   artifactrash.Manager
	blobMgr      blob.Manager
	sigMgr       signature.Manager
	labelMgr     label.Manager
	abstractor   abstractor.Abstractor
	immutableMtr match.ImmutableTagMatcher
	regCli       registry.Client
}

func (c *controller) Ensure(ctx context.Context, repository, digest string, tags ...string) (bool, int64, error) {
	created, artifact, err := c.ensureArtifact(ctx, repository, digest)
	if err != nil {
		return false, 0, err
	}
	for _, tag := range tags {
		if err = c.tagCtl.Ensure(ctx, artifact.RepositoryID, artifact.ID, tag); err != nil {
			return false, 0, err
		}
	}
	return created, artifact.ID, nil
}

// ensure the artifact exists under the repository, create it if doesn't exist.
func (c *controller) ensureArtifact(ctx context.Context, repository, digest string) (bool, *artifact.Artifact, error) {
	art, err := c.artMgr.GetByDigest(ctx, repository, digest)
	// the artifact already exists under the repository, return directly
	if err == nil {
		return false, art, nil
	}

	// got other error when get the artifact, return the error
	if !ierror.IsErr(err, ierror.NotFoundCode) {
		return false, nil, err
	}

	// the artifact doesn't exist under the repository, create it first
	repo, err := c.repoMgr.GetByName(ctx, repository)
	if err != nil {
		return false, nil, err
	}
	artifact := &artifact.Artifact{
		ProjectID:      repo.ProjectID,
		RepositoryID:   repo.RepositoryID,
		RepositoryName: repository,
		Digest:         digest,
		PushTime:       time.Now(),
	}
	// abstract the metadata for the artifact
	if err = c.abstractor.AbstractMetadata(ctx, artifact); err != nil {
		return false, nil, err
	}

	// populate the artifact type
	artifact.Type = descriptor.GetArtifactType(artifact.MediaType)

	// create it
	id, err := c.artMgr.Create(ctx, artifact)
	if err != nil {
		// if got conflict error, try to get the artifact again
		if ierror.IsConflictErr(err) {
			art, err = c.artMgr.GetByDigest(ctx, repository, digest)
			if err == nil {
				return false, art, nil
			}
			return false, nil, err
		}
		return false, nil, err
	}
	artifact.ID = id
	return true, artifact, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.artMgr.Count(ctx, query)
}

func (c *controller) List(ctx context.Context, query *q.Query, option *Option) ([]*Artifact, error) {
	arts, err := c.artMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}

	var artifacts []*Artifact
	for _, art := range arts {
		artifacts = append(artifacts, c.assembleArtifact(ctx, art, option))
	}
	return artifacts, nil
}

func (c *controller) Get(ctx context.Context, id int64, option *Option) (*Artifact, error) {
	art, err := c.artMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.assembleArtifact(ctx, art, option), nil
}

func (c *controller) GetByReference(ctx context.Context, repository, reference string, option *Option) (*Artifact, error) {
	// the reference is tag
	if _, err := digest.Parse(reference); err != nil {
		return c.getByTag(ctx, repository, reference, option)
	}
	// the reference is digest
	return c.getByDigest(ctx, repository, reference, option)
}

func (c *controller) getByDigest(ctx context.Context, repository, digest string, option *Option) (*Artifact, error) {
	art, err := c.artMgr.GetByDigest(ctx, repository, digest)
	if err != nil {
		return nil, err
	}
	return c.assembleArtifact(ctx, art, option), nil
}

func (c *controller) getByTag(ctx context.Context, repository, tag string, option *Option) (*Artifact, error) {
	repo, err := c.repoMgr.GetByName(ctx, repository)
	if err != nil {
		return nil, err
	}
	tags, err := c.tagCtl.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": repo.RepositoryID,
			"Name":         tag,
		},
	}, nil)
	if err != nil {
		return nil, err
	}
	if len(tags) == 0 {
		return nil, ierror.New(nil).WithCode(ierror.NotFoundCode).
			WithMessage("artifact %s:%s not found", repository, tag)
	}
	return c.Get(ctx, tags[0].ArtifactID, option)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	return c.deleteDeeply(ctx, id, true)
}

// "isRoot" is used to specify whether the artifact is the root parent artifact
// the error handling logic for the root parent artifact and others is different
func (c *controller) deleteDeeply(ctx context.Context, id int64, isRoot bool) error {
	art, err := c.Get(ctx, id, &Option{WithTag: true})
	if err != nil {
		// return nil if the nonexistent artifact isn't the root parent
		if !isRoot && ierror.IsErr(err, ierror.NotFoundCode) {
			return nil
		}
		return err
	}

	// the child artifact is referenced by some tags, skip
	if !isRoot && len(art.Tags) > 0 {
		return nil
	}
	parents, err := c.artMgr.ListReferences(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ChildID": id,
		},
	})
	if err != nil {
		return err
	}
	if len(parents) > 0 {
		// the root artifact is referenced by other artifacts
		if isRoot {
			return ierror.New(nil).WithCode(ierror.ViolateForeignKeyConstraintCode).
				WithMessage("the deleting artifact is referenced by others")
		}
		// the child artifact is referenced by other artifacts, skip
		return nil
	}
	// delete child artifacts if contains any
	for _, reference := range art.References {
		// delete reference
		if err = c.artMgr.DeleteReference(ctx, reference.ID); err != nil &&
			!ierror.IsErr(err, ierror.NotFoundCode) {
			return err
		}
		if err = c.deleteDeeply(ctx, reference.ChildID, false); err != nil {
			return err
		}
	}

	// delete all tags that attached to the root artifact
	if isRoot {
		if err = c.tagCtl.DeleteTags(ctx, art.Tags); err != nil {
			return err
		}
	}

	// remove labels added to the artifact
	if err := c.labelMgr.RemoveAllFrom(ctx, id); err != nil {
		return err
	}

	blobs, err := c.blobMgr.List(ctx, blob.ListParams{ArtifactDigest: art.Digest})
	if err != nil {
		return err
	}

	// clean associations between blob and project when when the blob is not needed by project
	if err := c.blobMgr.CleanupAssociationsForProject(ctx, art.ProjectID, blobs); err != nil {
		return err
	}

	// delete the artifact itself
	if err = c.artMgr.Delete(ctx, art.ID); err != nil {
		// the child artifact doesn't exist, skip
		if !isRoot && ierror.IsErr(err, ierror.NotFoundCode) {
			return nil
		}
		return err
	}

	_, err = c.artrashMgr.Create(ctx, &model.ArtifactTrash{
		MediaType:         art.MediaType,
		ManifestMediaType: art.ManifestMediaType,
		RepositoryName:    art.RepositoryName,
		Digest:            art.Digest,
	})
	if err != nil && !ierror.IsErr(err, ierror.ConflictCode) {
		return err
	}
	// TODO fire delete artifact event

	return nil
}

func (c *controller) Copy(ctx context.Context, srcRepo, reference, dstRepo string) (int64, error) {
	return c.copyDeeply(ctx, srcRepo, reference, dstRepo, true)
}

// as we call the docker registry APIs in the registry client directly,
// this bypass our own logic(ensure, fire event, etc.) inside the registry handlers,
// these logic must be covered explicitly here.
// "copyDeeply" iterates the child artifacts and copy them first
func (c *controller) copyDeeply(ctx context.Context, srcRepo, reference, dstRepo string, isRoot bool) (int64, error) {
	var option *Option
	// only get the tags of the root parent
	if isRoot {
		option = &Option{WithTag: true}
	}

	srcArt, err := c.GetByReference(ctx, srcRepo, reference, option)
	if err != nil {
		return 0, err
	}

	digest := srcArt.Digest

	// check the existence of artifact in the destination repository
	dstArt, err := c.GetByReference(ctx, dstRepo, digest, option)
	if err == nil {
		// return conflict error if the root parent artifact already exists under the destination repository
		if isRoot {
			return 0, ierror.New(nil).WithCode(ierror.ConflictCode).
				WithMessage("the artifact %s@%s already exists", dstRepo, digest)
		}
		// the child artifact already under the destination repository, skip
		return dstArt.ID, nil
	}
	if !ierror.IsErr(err, ierror.NotFoundCode) {
		return 0, err
	}

	// the artifact doesn't exist under the destination repository, continue to copy
	// copy child artifacts if contains any
	for _, reference := range srcArt.References {
		if _, err = c.copyDeeply(ctx, srcRepo, reference.ChildDigest, dstRepo, false); err != nil {
			return 0, err
		}
	}

	// copy the parent artifact into the backend docker registry
	if err := c.regCli.Copy(srcRepo, digest, dstRepo, digest, false); err != nil {
		return 0, err
	}

	// only copy the tags of outermost artifact
	var tags []string
	for _, tag := range srcArt.Tags {
		tags = append(tags, tag.Name)
	}
	// ensure the parent artifact exist in the database
	_, id, err := c.Ensure(ctx, dstRepo, digest, tags...)
	if err != nil {
		return 0, err
	}
	// TODO fire event
	return id, nil
}

func (c *controller) CreateTag(ctx context.Context, tag *tag.Tag) (int64, error) {
	// TODO fire event
	return c.tagCtl.Create(ctx, tag)
}
func (c *controller) ListTags(ctx context.Context, query *q.Query, option *tag.Option) ([]*tag.Tag, error) {
	tags, err := c.tagCtl.List(ctx, query, option)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func (c *controller) DeleteTag(ctx context.Context, tagID int64) error {
	// TODO fire delete tag event
	return c.tagCtl.Delete(ctx, tagID)
}

func (c *controller) UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) error {
	tag, err := c.tagCtl.Get(ctx, tagID, nil)
	if err != nil {
		return err
	}
	if tag.ArtifactID != artifactID {
		return fmt.Errorf("tag %d isn't attached to artifact %d", tagID, artifactID)
	}
	if err := c.artMgr.UpdatePullTime(ctx, artifactID, time); err != nil {
		return err
	}
	return c.tagCtl.Update(ctx, tag, "PullTime")
}

func (c *controller) GetAddition(ctx context.Context, artifactID int64, addition string) (*resolver.Addition, error) {
	artifact, err := c.artMgr.Get(ctx, artifactID)
	if err != nil {
		return nil, err
	}
	return c.abstractor.AbstractAddition(ctx, artifact, addition)
}

func (c *controller) AddLabel(ctx context.Context, artifactID int64, labelID int64) error {
	return c.labelMgr.AddTo(ctx, labelID, artifactID)
}

func (c *controller) RemoveLabel(ctx context.Context, artifactID int64, labelID int64) error {
	return c.labelMgr.RemoveFrom(ctx, labelID, artifactID)
}

// assemble several part into a single artifact
func (c *controller) assembleArtifact(ctx context.Context, art *artifact.Artifact, option *Option) *Artifact {
	artifact := &Artifact{
		Artifact: *art,
	}
	// populate addition links
	c.populateAdditionLinks(ctx, artifact)

	if option == nil {
		return artifact
	}
	if option.WithTag {
		c.populateTags(ctx, artifact, option.TagOption)
	}
	if option.WithLabel {
		c.populateLabels(ctx, artifact)
	}
	return artifact
}

func (c *controller) populateTags(ctx context.Context, art *Artifact, option *tag.Option) {
	tags, err := c.tagCtl.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"artifact_id": art.ID,
		},
	}, option)
	if err != nil {
		log.Errorf("failed to list tag of artifact %d: %v", art.ID, err)
		return
	}
	art.Tags = tags
}

func (c *controller) populateLabels(ctx context.Context, art *Artifact) {
	labels, err := c.labelMgr.ListByArtifact(ctx, art.ID)
	if err != nil {
		log.Errorf("failed to list labels of artifact %d: %v", art.ID, err)
		return
	}
	art.Labels = labels
}

func (c *controller) populateAdditionLinks(ctx context.Context, artifact *Artifact) {
	types := descriptor.ListAdditionTypes(artifact.MediaType)
	if len(types) > 0 {
		version := internal.GetAPIVersion(ctx)
		for _, t := range types {
			artifact.SetAdditionLink(strings.ToLower(t), version)
		}
	}
}
