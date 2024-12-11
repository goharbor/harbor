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
	"container/list"
	"context"
	stderrors "errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/opencontainers/go-digest"

	"github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/chart"
	"github.com/goharbor/harbor/src/controller/artifact/processor/cnab"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/artifact/processor/sbom"
	"github.com/goharbor/harbor/src/controller/artifact/processor/wasm"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/icon"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/accessory"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/artifactrash"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	"github.com/goharbor/harbor/src/pkg/blob"
	"github.com/goharbor/harbor/src/pkg/immutable/match"
	"github.com/goharbor/harbor/src/pkg/immutable/match/rule"
	"github.com/goharbor/harbor/src/pkg/label"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/pkg/repository"
	model_tag "github.com/goharbor/harbor/src/pkg/tag/model/tag"
)

var (
	// Ctl is a global artifact controller instance
	Ctl                 = NewController()
	skippedContentTypes = map[string]struct{}{
		"application/vnd.in-toto+json": {},
	}
)

var (
	// ErrBreak error to break walk
	ErrBreak = stderrors.New("break")

	// ErrSkip error to skip walk the children of the artifact
	ErrSkip = stderrors.New("skip")

	// icon digests for each known type
	defaultIcons = map[string]string{
		image.ArtifactTypeImage: icon.DigestOfIconImage,
		chart.ArtifactTypeChart: icon.DigestOfIconChart,
		cnab.ArtifactTypeCNAB:   icon.DigestOfIconCNAB,
		wasm.ArtifactTypeWASM:   icon.DigestOfIconWASM,
		sbom.ArtifactTypeSBOM:   icon.DigestOfIconAccSBOM,
	}
)

// Controller defines the operations related with artifacts and tags
type Controller interface {
	// Ensure the artifact specified by the digest exists under the repository,
	// creates it if it doesn't exist. If tags are provided, ensure they exist
	// and are attached to the artifact. If the tags don't exist, create them first.
	// The "created" will be set as true when the artifact is created
	Ensure(ctx context.Context, repository, digest string, option *ArtOption) (created bool, id int64, err error)
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
	// UpdatePullTime updates the pull time for the artifact. If the tagID is provides, update the pull
	// time of the tag as well
	UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) (err error)
	// GetAddition returns the addition of the artifact.
	// The addition is different according to the artifact type:
	// build history for image; values.yaml, readme and dependencies for chart, etc
	GetAddition(ctx context.Context, artifactID int64, additionType string) (addition *processor.Addition, err error)
	// AddLabel to the specified artifact
	AddLabel(ctx context.Context, artifactID int64, labelID int64) (err error)
	// RemoveLabel from the specified artifact
	RemoveLabel(ctx context.Context, artifactID int64, labelID int64) (err error)
	// Walk walks the artifact tree rooted at root, calling walkFn for each artifact in the tree, including root.
	Walk(ctx context.Context, root *Artifact, walkFn func(*Artifact) error, option *Option) error
	// HasUnscannableLayer check artifact with digest if has unscannable layer
	HasUnscannableLayer(ctx context.Context, dgst string) (bool, error)
	// ListWithLatest list the artifacts when the latest_in_repository in the query was set
	ListWithLatest(ctx context.Context, query *q.Query, option *Option) (artifacts []*Artifact, err error)
}

// NewController creates an instance of the default artifact controller
func NewController() Controller {
	return &controller{
		tagCtl:       tag.Ctl,
		repoMgr:      pkg.RepositoryMgr,
		artMgr:       pkg.ArtifactMgr,
		artrashMgr:   artifactrash.Mgr,
		blobMgr:      blob.Mgr,
		labelMgr:     label.Mgr,
		immutableMtr: rule.NewRuleMatcher(),
		regCli:       registry.Cli,
		abstractor:   NewAbstractor(),
		accessoryMgr: accessory.Mgr,
	}
}

type controller struct {
	tagCtl       tag.Controller
	repoMgr      repository.Manager
	artMgr       artifact.Manager
	artrashMgr   artifactrash.Manager
	blobMgr      blob.Manager
	labelMgr     label.Manager
	immutableMtr match.ImmutableTagMatcher
	regCli       registry.Client
	abstractor   Abstractor
	accessoryMgr accessory.Manager
}

type ArtOption struct {
	Tags []string
	Accs []*accessorymodel.AccessoryData
}

func (c *controller) Ensure(ctx context.Context, repository, digest string, option *ArtOption) (bool, int64, error) {
	created, artifact, err := c.ensureArtifact(ctx, repository, digest)
	if err != nil {
		return false, 0, err
	}
	if option != nil {
		for _, tag := range option.Tags {
			if _, err = c.tagCtl.Ensure(ctx, artifact.RepositoryID, artifact.ID, tag); err != nil {
				return false, 0, err
			}
		}
		for _, acc := range option.Accs {
			if err = c.accessoryMgr.Ensure(ctx, artifact.Digest, artifact.RepositoryName, artifact.ID, acc.ArtifactID, acc.Size, acc.Digest, acc.Type); err != nil {
				return false, 0, err
			}
		}
	}
	if created {
		// fire event for create
		e := &metadata.PushArtifactEventMetadata{
			Ctx:      ctx,
			Artifact: artifact,
		}

		if option != nil && len(option.Tags) > 0 {
			e.Tag = option.Tags[0]
		}
		notification.AddEvent(ctx, e)
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
	if !errors.IsErr(err, errors.NotFoundCode) {
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
	artifact.Type = processor.Get(artifact.MediaType).GetArtifactType(ctx, artifact)

	// create it
	// use orm.WithTransaction here to avoid the issue:
	// https://www.postgresql.org/message-id/002e01c04da9%24a8f95c20%2425efe6c1%40lasting.ro
	created := false
	if err = orm.WithTransaction(func(ctx context.Context) error {
		id, err := c.artMgr.Create(ctx, artifact)
		if err != nil {
			return err
		}
		created = true
		artifact.ID = id
		return nil
	})(orm.SetTransactionOpNameToContext(ctx, "tx-ensure-artifact")); err != nil {
		// got error that isn't conflict error, return directly
		if !errors.IsConflictErr(err) {
			return false, nil, err
		}
		log.Debugf("failed to create artifact %s@%s: %v", repository, digest, err)
		// if got conflict error, try to get the artifact again
		artifact, err = c.artMgr.GetByDigest(ctx, repository, digest)
		if err != nil {
			return false, nil, err
		}
	}

	return created, artifact, nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.artMgr.Count(ctx, query)
}

func (c *controller) List(ctx context.Context, query *q.Query, option *Option) ([]*Artifact, error) {
	arts, err := c.artMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}

	var res []*Artifact
	for _, art := range arts {
		res = append(res, c.assembleArtifact(ctx, art, option))
	}
	return res, nil
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
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("artifact %s:%s not found", repository, tag)
	}
	return c.Get(ctx, tags[0].ArtifactID, option)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	accs, err := c.accessoryMgr.List(ctx, q.New(q.KeyWords{"ArtifactID": id}))
	if err != nil {
		return err
	}
	return orm.WithTransaction(func(ctx context.Context) error {
		return c.deleteDeeply(ctx, id, true, len(accs) > 0)
	})(orm.SetTransactionOpNameToContext(ctx, "tx-delete-artifact-delete"))
}

// "isRoot" is used to specify whether the artifact is the root parent artifact
// the error handling logic for the root parent artifact and others is different
// "isAccessory" is used to specify whether the artifact is an accessory.
func (c *controller) deleteDeeply(ctx context.Context, id int64, isRoot, isAccessory bool) error {
	art, err := c.Get(ctx, id, &Option{WithTag: true, WithAccessory: true, WithLabel: true})
	if err != nil {
		// return nil if the nonexistent artifact isn't the root parent
		if !isRoot && errors.IsErr(err, errors.NotFoundCode) {
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
			return errors.New(nil).WithCode(errors.ViolateForeignKeyConstraintCode).
				WithMessage("the deleting artifact is referenced by others")
		}
		// the child artifact is referenced by other artifacts, skip
		return nil
	}

	if isAccessory {
		if err := c.accessoryMgr.DeleteAccessories(ctx, q.New(q.KeyWords{"ArtifactID": art.ID, "Digest": art.Digest})); err != nil && !errors.IsErr(err, errors.NotFoundCode) {
			return err
		}
	}

	// delete accessories if contains any
	for _, acc := range art.Accessories {
		// only hard ref accessory should be removed
		if acc.IsHard() {
			// if this acc artifact has parent(is child), set isRoot to false
			parents, err := c.artMgr.ListReferences(ctx, &q.Query{
				Keywords: map[string]interface{}{
					"ChildID": acc.GetData().ArtifactID,
				},
			})
			if err != nil {
				return err
			}
			if err = c.deleteDeeply(ctx, acc.GetData().ArtifactID, len(parents) == 0, true); err != nil {
				return err
			}
		}
	}

	// delete child artifacts if contains any
	for _, reference := range art.References {
		// delete reference
		if err = c.artMgr.DeleteReference(ctx, reference.ID); err != nil &&
			!errors.IsErr(err, errors.NotFoundCode) {
			return err
		}
		// if the child artifact is an accessory, set isAccessory to true
		accs, err := c.accessoryMgr.List(ctx, q.New(q.KeyWords{"ArtifactID": reference.ChildID}))
		if err != nil {
			return err
		}
		if err = c.deleteDeeply(ctx, reference.ChildID, false, len(accs) > 0); err != nil {
			return err
		}
	}

	// delete all tags that attached to the root artifact
	if isRoot {
		var ids []int64
		for _, tag := range art.Tags {
			ids = append(ids, tag.ID)
		}
		if err = c.tagCtl.DeleteTags(ctx, ids); err != nil {
			return err
		}
	}

	// remove labels added to the artifact
	if err := c.labelMgr.RemoveAllFrom(ctx, id); err != nil {
		return err
	}

	// delete the artifact itself
	if err = c.artMgr.Delete(ctx, art.ID); err != nil {
		// the child artifact doesn't exist, skip
		if !isRoot && errors.IsErr(err, errors.NotFoundCode) {
			return nil
		}
		return err
	}

	blobs, err := c.blobMgr.List(ctx, q.New(q.KeyWords{"artifactDigest": art.Digest}))
	if err != nil {
		return err
	}

	// clean associations between blob and project when the blob is not needed by project
	if err := c.blobMgr.CleanupAssociationsForProject(ctx, art.ProjectID, blobs); err != nil {
		return err
	}

	// use orm.WithTransaction here to avoid the issue:
	// https://www.postgresql.org/message-id/002e01c04da9%24a8f95c20%2425efe6c1%40lasting.ro
	if err = orm.WithTransaction(func(ctx context.Context) error {
		_, err = c.artrashMgr.Create(ctx, &model.ArtifactTrash{
			MediaType:         art.MediaType,
			ManifestMediaType: art.ManifestMediaType,
			RepositoryName:    art.RepositoryName,
			Digest:            art.Digest,
		})
		return err
	})(orm.SetTransactionOpNameToContext(ctx, "tx-delete-artifact-deeply")); err != nil && !errors.IsErr(err, errors.ConflictCode) {
		return err
	}

	// only fire event for the root parent artifact
	if isRoot {
		var tags, labels []string
		for _, tag := range art.Tags {
			tags = append(tags, tag.Name)
		}

		for _, label := range art.Labels {
			labels = append(labels, label.Name)
		}

		notification.AddEvent(ctx, &metadata.DeleteArtifactEventMetadata{
			Ctx:      ctx,
			Artifact: &art.Artifact,
			Tags:     tags,
			Labels:   labels,
		})
	}

	return nil
}

func (c *controller) Copy(ctx context.Context, srcRepo, reference, dstRepo string) (int64, error) {
	dstAccs := make([]*accessorymodel.AccessoryData, 0)
	return c.copyDeeply(ctx, srcRepo, reference, dstRepo, true, false, &dstAccs)
}

// as we call the docker registry APIs in the registry client directly,
// this bypass our own logic(ensure, fire event, etc.) inside the registry handlers,
// these logic must be covered explicitly here.
// "copyDeeply" iterates the child artifacts and copy them first
func (c *controller) copyDeeply(ctx context.Context, srcRepo, reference, dstRepo string, isRoot, isAcc bool, dstAccs *[]*accessorymodel.AccessoryData) (int64, error) {
	var option *Option
	option = &Option{WithTag: true, WithAccessory: true}
	if isAcc {
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
		// the child artifact already exists under the destination repository, skip
		if !isRoot {
			return dstArt.ID, nil
		}
		// the root parent already exists, goto next step to ensure artifact: create artifact & references, copy tags & accessories.
		goto ensureArt
	}
	if !errors.IsErr(err, errors.NotFoundCode) {
		return 0, err
	}

	// the artifact doesn't exist under the destination repository, continue to copy
	// copy child artifacts if contains any
	for _, reference := range srcArt.References {
		if _, err = c.copyDeeply(ctx, srcRepo, reference.ChildDigest, dstRepo, false, false, dstAccs); err != nil {
			return 0, err
		}
	}

	// copy accessory if contains any
	for _, acc := range srcArt.Accessories {
		accs, err := c.accessoryMgr.List(ctx, q.New(q.KeyWords{"SubjectArtifactRepo": srcRepo, "SubjectArtifactDigest": acc.GetData().Digest}))
		if err != nil {
			return 0, err
		}
		// copy the fork which root is the accessory self with a temp array
		// to avoid infinite recursion, disable this part in UT.
		if os.Getenv("UTTEST") != "true" {
			if len(accs) > 0 {
				tmpDstAccs := make([]*accessorymodel.AccessoryData, 0)
				_, err = c.copyDeeply(ctx, srcRepo, acc.GetData().Digest, dstRepo, true, false, &tmpDstAccs)
				if err != nil {
					return 0, err
				}
			}
		}
		dstAcc := &accessorymodel.AccessoryData{
			Digest: acc.GetData().Digest,
			Type:   acc.GetData().Type,
			Size:   acc.GetData().Size,
		}
		*dstAccs = append(*dstAccs, dstAcc)
		id, err := c.copyDeeply(ctx, srcRepo, acc.GetData().Digest, dstRepo, false, true, dstAccs)
		if err != nil {
			return 0, err
		}
		dstAcc.ArtifactID = id
	}

	// copy the parent artifact into the backend docker registry
	if err := c.regCli.Copy(srcRepo, digest, dstRepo, digest, false); err != nil {
		return 0, err
	}

ensureArt:
	// only copy the tags of outermost artifact
	var tags []string
	for _, tag := range srcArt.Tags {
		tags = append(tags, tag.Name)
	}
	// ensure the parent artifact exist in the database
	artopt := &ArtOption{
		Tags: tags,
	}
	if !isAcc {
		artopt.Accs = *dstAccs
	}
	_, id, err := c.Ensure(ctx, dstRepo, digest, artopt)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (c *controller) UpdatePullTime(ctx context.Context, artifactID int64, tagID int64, time time.Time) error {
	if err := c.artMgr.UpdatePullTime(ctx, artifactID, time); err != nil {
		return err
	}
	// update tag pull time if artifact has tag
	if tagID != 0 {
		tg, err := c.tagCtl.Get(ctx, tagID, nil)
		if err != nil {
			return err
		}
		if tg.ArtifactID != artifactID {
			return fmt.Errorf("tag %d isn't attached to artifact %d", tagID, artifactID)
		}
		return c.tagCtl.Update(ctx, &tag.Tag{
			Tag: model_tag.Tag{
				ID:       tg.ID,
				PullTime: time,
			},
		}, "PullTime")
	}

	return nil
}

func (c *controller) GetAddition(ctx context.Context, artifactID int64, addition string) (*processor.Addition, error) {
	artifact, err := c.artMgr.Get(ctx, artifactID)
	if err != nil {
		return nil, err
	}
	return processor.Get(artifact.MediaType).AbstractAddition(ctx, artifact, addition)
}

func (c *controller) AddLabel(ctx context.Context, artifactID int64, labelID int64) (err error) {
	defer func() {
		if err == nil {
			// trigger label artifact event
			e := &event.Event{}
			metaData := &metadata.ArtifactLabeledMetadata{
				ArtifactID: artifactID,
				LabelID:    labelID,
				Ctx:        ctx,
			}
			if err := e.Build(ctx, metaData); err == nil {
				if err := e.Publish(ctx); err != nil {
					log.Error(errors.Wrap(err, "mark label to resource handler: event publish"))
				}
			} else {
				log.Error(errors.Wrap(err, "mark label to resource handler: event build"))
			}
		}
	}()
	err = c.labelMgr.AddTo(ctx, labelID, artifactID)
	return
}

func (c *controller) RemoveLabel(ctx context.Context, artifactID int64, labelID int64) error {
	return c.labelMgr.RemoveFrom(ctx, labelID, artifactID)
}

func (c *controller) Walk(ctx context.Context, root *Artifact, walkFn func(*Artifact) error, option *Option) error {
	queue := list.New()
	queue.PushBack(root)

	walked := map[string]bool{}

	for queue.Len() > 0 {
		elem := queue.Front()
		queue.Remove(elem)

		artifact := elem.Value.(*Artifact)
		walked[artifact.Digest] = true

		if err := walkFn(artifact); err != nil {
			if err == ErrBreak {
				return nil
			} else if err == ErrSkip {
				continue
			}

			return err
		}

		if len(artifact.References) > 0 {
			var ids []int64
			for _, ref := range artifact.References {
				ids = append(ids, ref.ChildID)
			}

			// HACK: base=* in KeyWords to filter all artifacts
			children, err := c.List(ctx, q.New(q.KeyWords{"id__in": ids, "base": "*"}), option)
			if err != nil {
				return err
			}

			for _, child := range children {
				if !walked[child.Digest] {
					queue.PushBack(child)
				}
				if len(child.Accessories) != 0 {
					for _, acc := range child.Accessories {
						accArt, err := c.Get(ctx, acc.GetData().ArtifactID, option)
						if err != nil {
							return err
						}
						if !walked[accArt.Digest] {
							queue.PushBack(accArt)
						}
					}
				}
			}
		}

		if len(artifact.Accessories) > 0 {
			for _, acc := range artifact.Accessories {
				accArt, err := c.Get(ctx, acc.GetData().ArtifactID, option)
				if err != nil {
					return err
				}
				if !walked[accArt.Digest] {
					queue.PushBack(accArt)
				}
			}
		}
	}

	return nil
}

// assemble several part into a single artifact
func (c *controller) assembleArtifact(ctx context.Context, art *artifact.Artifact, option *Option) *Artifact {
	artifact := &Artifact{
		Artifact: *art,
	}
	// populate addition links
	c.populateAdditionLinks(ctx, artifact)

	// populate icon for the known artifact types
	c.populateIcon(artifact)

	if option == nil {
		return artifact
	}
	if option.WithTag {
		c.populateTags(ctx, artifact, option.TagOption)
	}
	if option.WithLabel {
		c.populateLabels(ctx, artifact)
	}
	if option.WithAccessory {
		c.populateAccessories(ctx, artifact)
	}
	return artifact
}

func (c *controller) populateIcon(art *Artifact) {
	if len(art.Icon) == 0 {
		if i, ok := defaultIcons[art.Type]; ok {
			art.Icon = i
		} else {
			art.Icon = icon.DigestOfIconDefault
		}
	}
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
	types := processor.Get(artifact.MediaType).ListAdditionTypes(ctx, &artifact.Artifact)
	if len(types) > 0 {
		version := lib.GetAPIVersion(ctx)
		for _, t := range types {
			artifact.SetAdditionLink(strings.ToLower(t), version)
		}
	}
}

func (c *controller) populateAccessories(ctx context.Context, art *Artifact) {
	accs, err := c.accessoryMgr.List(ctx, q.New(q.KeyWords{"SubjectArtifactID": art.ID}))
	if err != nil {
		log.Errorf("failed to list accessories of artifact %d: %v", art.ID, err)
		return
	}
	art.Accessories = accs
}

// HasUnscannableLayer check if it is a in-toto sbom, if it contains any blob with a content_type is application/vnd.in-toto+json, then consider as in-toto sbom
func (c *controller) HasUnscannableLayer(ctx context.Context, dgst string) (bool, error) {
	if len(dgst) == 0 {
		return false, nil
	}
	blobs, err := c.blobMgr.GetByArt(ctx, dgst)
	if err != nil {
		return false, err
	}
	for _, b := range blobs {
		if _, exist := skippedContentTypes[b.ContentType]; exist {
			log.Debugf("the artifact with digest %v is unscannable, because it contains content type: %v", dgst, b.ContentType)
			return true, nil
		}
	}
	return false, nil
}

// ListWithLatest ...
func (c *controller) ListWithLatest(ctx context.Context, query *q.Query, option *Option) (artifacts []*Artifact, err error) {
	arts, err := c.artMgr.ListWithLatest(ctx, query)
	if err != nil {
		return nil, err
	}
	var res []*Artifact
	for _, art := range arts {
		res = append(res, c.assembleArtifact(ctx, art, option))
	}
	return res, nil
}
