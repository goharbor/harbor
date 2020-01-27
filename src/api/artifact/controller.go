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
	"github.com/goharbor/harbor/src/api/artifact/abstractor/resolver/image"
	"github.com/goharbor/harbor/src/api/artifact/descriptor"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/internal"
	"github.com/goharbor/harbor/src/pkg/art"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
	"github.com/goharbor/harbor/src/pkg/label"
	"github.com/opencontainers/go-digest"
	"strings"

	// registry image resolvers
	_ "github.com/goharbor/harbor/src/api/artifact/abstractor/resolver/image"
	// register chart resolver
	_ "github.com/goharbor/harbor/src/api/artifact/abstractor/resolver/chart"
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
	// Get the artifact specified by repository name and reference, the reference can be tag or digest,
	// specify the properties returned with option
	GetByReference(ctx context.Context, repository, reference string, option *Option) (artifact *Artifact, err error)
	// Delete the artifact specified by ID. All tags attached to the artifact are deleted as well
	Delete(ctx context.Context, id int64) (err error)
	// ListTags lists the tags according to the query, specify the properties returned with option
	ListTags(ctx context.Context, query *q.Query, option *TagOption) (total int64, tags []*Tag, err error)
	// CreateTag creates a tag
	CreateTag(ctx context.Context, tag *Tag) (id int64, err error)
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
	// TODO move this to GC controller?
	// Prune removes the useless artifact records. The underlying registry data will
	// be removed during garbage collection
	// Prune(ctx context.Context, option *Option) error
}

// NewController creates an instance of the default artifact controller
func NewController() Controller {
	return &controller{
		repoMgr:      repository.Mgr,
		artMgr:       artifact.Mgr,
		tagMgr:       tag.Mgr,
		labelMgr:     label.Mgr,
		abstractor:   abstractor.NewAbstractor(),
		immutableMtr: rule.NewRuleMatcher(),
	}
}

// TODO concurrency summary

type controller struct {
	repoMgr      repository.Manager
	artMgr       artifact.Manager
	tagMgr       tag.Manager
	labelMgr     label.Manager
	abstractor   abstractor.Abstractor
	immutableMtr match.ImmutableTagMatcher
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
	// abstract the metadata for the artifact
	if err = c.abstractor.AbstractMetadata(ctx, artifact); err != nil {
		return false, 0, err
	}

	// populate the artifact type
	typee, err := descriptor.GetArtifactType(artifact.MediaType)
	if err != nil {
		return false, 0, err
	}
	artifact.Type = typee

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

func (c *controller) GetByReference(ctx context.Context, repository, reference string, option *Option) (*Artifact, error) {
	// the reference is tag
	if _, err := digest.Parse(reference); err != nil {
		return c.getByTag(ctx, repository, reference, option)
	}
	// the reference is digest
	return c.getByDigest(ctx, repository, reference, option)
}

func (c *controller) getByDigest(ctx context.Context, repository, digest string, option *Option) (*Artifact, error) {
	repo, err := c.repoMgr.GetByName(ctx, repository)
	if err != nil {
		return nil, err
	}
	_, artifacts, err := c.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": repo.RepositoryID,
			"Digest":       digest,
		},
	}, option)
	if err != nil {
		return nil, err
	}
	if len(artifacts) == 0 {
		return nil, ierror.New(nil).WithCode(ierror.NotFoundCode).
			WithMessage("artifact %s@%s not found", repository, digest)
	}
	return artifacts[0], nil
}

func (c *controller) getByTag(ctx context.Context, repository, tag string, option *Option) (*Artifact, error) {
	repo, err := c.repoMgr.GetByName(ctx, repository)
	if err != nil {
		return nil, err
	}
	_, tags, err := c.tagMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": repo.RepositoryID,
			"Name":         tag,
		},
	})
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
	// remove labels added to the artifact
	if err := c.labelMgr.RemoveAllFrom(ctx, id); err != nil {
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

	if err := c.artMgr.Delete(ctx, id); err != nil {
		return err
	}

	// TODO fire delete artifact event
	return nil
}

func (c *controller) CreateTag(ctx context.Context, tag *Tag) (int64, error) {
	// TODO fire event
	return c.tagMgr.Create(ctx, &(tag.Tag))
}
func (c *controller) ListTags(ctx context.Context, query *q.Query, option *TagOption) (int64, []*Tag, error) {
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
	// Immutable checking is covered in middleware
	// TODO check signature
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

func (c *controller) GetAddition(ctx context.Context, artifactID int64, addition string) (*resolver.Addition, error) {
	artifact, err := c.artMgr.Get(ctx, artifactID)
	if err != nil {
		return nil, err
	}
	switch addition {
	case image.AdditionTypeVulnerabilities:
		// get the vulnerabilities from scan service
		// TODO implement
		return &resolver.Addition{}, nil
	default:
		return c.abstractor.AbstractAddition(ctx, artifact, addition)
	}
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
	if option == nil {
		return artifact
	}
	if option.WithTag {
		c.populateTags(ctx, artifact, option.TagOption)
	}
	if option.WithLabel {
		c.populateLabels(ctx, artifact)
	}
	if option.WithScanOverview {
		c.populateScanOverview(ctx, artifact)
	}
	if option.WithSignature {
		c.populateSignature(ctx, artifact)
	}
	// populate addition links
	c.populateAdditionLinks(ctx, artifact)
	return artifact
}

func (c *controller) populateTags(ctx context.Context, art *Artifact, option *TagOption) {
	_, tags, err := c.tagMgr.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"artifact_id": art.ID,
		},
	})
	if err != nil {
		log.Errorf("failed to list tag of artifact %d: %v", art.ID, err)
		return
	}
	for _, tag := range tags {
		art.Tags = append(art.Tags, c.assembleTag(ctx, tag, option))
	}
}

// assemble several part into a single tag
func (c *controller) assembleTag(ctx context.Context, tag *tm.Tag, option *TagOption) *Tag {
	t := &Tag{
		Tag: *tag,
	}
	if option == nil {
		return t
	}
	if option.WithImmutableStatus {
		c.populateImmutableStatus(ctx, t)
	}
	return t
}

func (c *controller) populateLabels(ctx context.Context, art *Artifact) {
	labels, err := c.labelMgr.ListByArtifact(ctx, art.ID)
	if err != nil {
		log.Errorf("failed to list labels of artifact %d: %v", art.ID, err)
		return
	}
	art.Labels = labels
}

func (c *controller) populateImmutableStatus(ctx context.Context, tag *Tag) {
	repo, err := c.repoMgr.Get(ctx, tag.RepositoryID)
	if err != nil {
		log.Error(err)
		return
	}
	_, repoName := utils.ParseRepository(repo.Name)
	matched, err := c.immutableMtr.Match(repo.ProjectID, art.Candidate{
		Repository:  repoName,
		Tag:         tag.Name,
		NamespaceID: repo.ProjectID,
	})
	if err != nil {
		log.Error(err)
		return
	}
	tag.Immutable = matched
}

func (c *controller) populateScanOverview(ctx context.Context, art *Artifact) {
	// TODO implement
}

func (c *controller) populateSignature(ctx context.Context, art *Artifact) {
	// TODO implement
	// TODO populate signature on artifact or tag level?
}

func (c *controller) populateAdditionLinks(ctx context.Context, artifact *Artifact) {
	types, err := descriptor.ListAdditionTypes(artifact.MediaType)
	if err != nil {
		log.Error(err.Error())
		return
	}
	if len(types) == 0 {
		return
	}
	repository, err := c.repoMgr.Get(ctx, artifact.RepositoryID)
	if err != nil {
		log.Error(err.Error())
		return
	}
	pro, repo := utils.ParseRepository(repository.Name)
	version := internal.GetAPIVersion(ctx)
	if artifact.AdditionLinks == nil {
		artifact.AdditionLinks = make(map[string]*AdditionLink)
	}
	href := ""
	for _, t := range types {
		t = strings.ToLower(t)
		switch t {
		case image.AdditionTypeVulnerabilities:
			// check whether the scan service is enabled and set the addition link
			// TODO implement
			href = fmt.Sprintf("/api/%s/projects/%s/repositories/%s/artifacts/%s/vulnerabilities",
				version, pro, repo, artifact.Digest)
		default:
			href = fmt.Sprintf("/api/%s/projects/%s/repositories/%s/artifacts/%s/additions/%s",
				version, pro, repo, artifact.Digest, t)
		}
		artifact.AdditionLinks[t] = &AdditionLink{
			HREF:     href,
			Absolute: false,
		}
	}
}
