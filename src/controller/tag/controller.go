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

package tag

import (
	"context"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match"
	"github.com/goharbor/harbor/src/pkg/immutabletag/match/rule"
	"github.com/goharbor/harbor/src/pkg/signature"
	"github.com/goharbor/harbor/src/pkg/tag"
	model_tag "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"time"
)

var (
	// Ctl is a global tag controller instance
	Ctl = NewController()
)

// Controller manages the tags
type Controller interface {
	// Ensure
	Ensure(ctx context.Context, repositoryID, artifactID int64, name string) error
	// Count returns the total count of tags according to the query.
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List tags according to the query
	List(ctx context.Context, query *q.Query, option *Option) (tags []*Tag, err error)
	// Get the tag specified by ID
	Get(ctx context.Context, id int64, option *Option) (tag *Tag, err error)
	// Create the tag and returns the ID
	Create(ctx context.Context, tag *Tag) (id int64, err error)
	// Update the tag. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, tag *Tag, props ...string) (err error)
	// Delete the tag specified by ID with limitation check
	Delete(ctx context.Context, id int64) (err error)
	// DeleteTags deletes all tags
	DeleteTags(ctx context.Context, ids []int64) (err error)
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		tagMgr:       tag.Mgr,
		artMgr:       artifact.Mgr,
		immutableMtr: rule.NewRuleMatcher(),
	}
}

type controller struct {
	tagMgr       tag.Manager
	artMgr       artifact.Manager
	immutableMtr match.ImmutableTagMatcher
}

// Ensure ...
func (c *controller) Ensure(ctx context.Context, repositoryID, artifactID int64, name string) error {
	query := &q.Query{
		Keywords: map[string]interface{}{
			"repository_id": repositoryID,
			"name":          name,
		},
	}
	tags, err := c.List(ctx, query, &Option{
		WithImmutableStatus: true,
	})
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
		// existing tag must check the immutable status and signature
		if tag.Immutable {
			return errors.New(nil).WithCode(errors.PreconditionCode).
				WithMessage("the tag %s configured as immutable, cannot be updated", tag.Name)
		}
		// the tag exists under the repository, but it is attached to other artifact
		// update it to point to the provided artifact
		tag.ArtifactID = artifactID
		tag.PushTime = time.Now()
		return c.Update(ctx, tag, "ArtifactID", "PushTime")
	}

	// the tag doesn't exist under the repository, create it
	// use orm.WithTransaction here to avoid the issue:
	// https://www.postgresql.org/message-id/002e01c04da9%24a8f95c20%2425efe6c1%40lasting.ro
	if err = orm.WithTransaction(func(ctx context.Context) error {
		tag := &Tag{}
		tag.RepositoryID = repositoryID
		tag.ArtifactID = artifactID
		tag.Name = name
		tag.PushTime = time.Now()
		_, err = c.Create(ctx, tag)
		return err
	})(ctx); err != nil && !errors.IsConflictErr(err) {
		return err
	}

	return nil
}

// Count ...
func (c *controller) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return c.tagMgr.Count(ctx, query)
}

// List ...
func (c *controller) List(ctx context.Context, query *q.Query, option *Option) ([]*Tag, error) {
	tgs, err := c.tagMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var tags []*Tag
	for _, tg := range tgs {
		tags = append(tags, c.assembleTag(ctx, tg, option))
	}
	return tags, nil
}

// Get ...
func (c *controller) Get(ctx context.Context, id int64, option *Option) (tag *Tag, err error) {
	tag = &Tag{}
	daoTag, err := c.tagMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	tag.Tag = *daoTag

	if option == nil {
		return tag, nil
	}

	return c.assembleTag(ctx, &tag.Tag, option), nil
}

// Create ...
func (c *controller) Create(ctx context.Context, tag *Tag) (id int64, err error) {
	return c.tagMgr.Create(ctx, &(tag.Tag))
}

// Update ...
func (c *controller) Update(ctx context.Context, tag *Tag, props ...string) (err error) {
	return c.tagMgr.Update(ctx, &tag.Tag, props...)
}

// Delete needs to check the signature and immutable status
func (c *controller) Delete(ctx context.Context, id int64) (err error) {
	option := &Option{
		WithImmutableStatus: true,
		WithSignature:       true,
	}
	tag, err := c.Get(ctx, id, option)
	if err != nil {
		return err
	}
	if tag.Immutable {
		return errors.New(nil).WithCode(errors.PreconditionCode).
			WithMessage("the tag %s configured as immutable, cannot be deleted", tag.Name)
	}
	if tag.Signed {
		return errors.New(nil).WithCode(errors.PreconditionCode).
			WithMessage("the tag %s with signature cannot be deleted", tag.Name)
	}
	return c.tagMgr.Delete(ctx, id)
}

// DeleteTags ...
func (c *controller) DeleteTags(ctx context.Context, ids []int64) (err error) {
	// in order to leverage the signature and immutable status check
	for _, id := range ids {
		if err := c.Delete(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

// assemble several part into a single tag
func (c *controller) assembleTag(ctx context.Context, tag *model_tag.Tag, option *Option) *Tag {
	t := &Tag{
		Tag: *tag,
	}
	if option == nil {
		return t
	}
	if option.WithImmutableStatus {
		c.populateImmutableStatus(ctx, t)
	}
	if option.WithSignature {
		c.populateTagSignature(ctx, t, option)
	}
	return t
}

func (c *controller) populateImmutableStatus(ctx context.Context, tag *Tag) {
	artifact, err := c.artMgr.Get(ctx, tag.ArtifactID)
	if err != nil {
		return
	}
	_, repoName := utils.ParseRepository(artifact.RepositoryName)
	matched, err := c.immutableMtr.Match(artifact.ProjectID, selector.Candidate{
		Repository:  repoName,
		Tags:        []string{tag.Name},
		NamespaceID: artifact.ProjectID,
	})
	if err != nil {
		return
	}
	tag.Immutable = matched
}

func (c *controller) populateTagSignature(ctx context.Context, tag *Tag, option *Option) {
	artifact, err := c.artMgr.Get(ctx, tag.ArtifactID)
	if err != nil {
		return
	}
	if option.SignatureChecker == nil {
		chk, err := signature.GetManager().GetCheckerByRepo(ctx, artifact.RepositoryName)
		if err != nil {
			log.Error(err)
			return
		}
		option.SignatureChecker = chk
	}
	tag.Signed = option.SignatureChecker.IsTagSigned(tag.Name, artifact.Digest)
}
