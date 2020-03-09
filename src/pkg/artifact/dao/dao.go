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

package dao

import (
	"context"
	"fmt"
	"strings"

	beegoorm "github.com/astaxie/beego/orm"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/q"
)

// DAO is the data access object interface for artifact
type DAO interface {
	// Count returns the total count of artifacts according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List artifacts according to the query. The artifacts that referenced by others and
	// without tags are not returned
	List(ctx context.Context, query *q.Query) (artifacts []*Artifact, err error)
	// Get the artifact specified by ID
	Get(ctx context.Context, id int64) (*Artifact, error)
	// GetByDigest returns the artifact specified by repository and digest
	GetByDigest(ctx context.Context, repository, digest string) (artifact *Artifact, err error)
	// Create the artifact
	Create(ctx context.Context, artifact *Artifact) (id int64, err error)
	// Delete the artifact specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// Update updates the artifact. Only the properties specified by "props" will be updated if it is set
	Update(ctx context.Context, artifact *Artifact, props ...string) (err error)
	// CreateReference creates the artifact reference
	CreateReference(ctx context.Context, reference *ArtifactReference) (id int64, err error)
	// ListReferences lists the artifact references according to the query
	ListReferences(ctx context.Context, query *q.Query) (references []*ArtifactReference, err error)
	// DeleteReference specified by ID
	DeleteReference(ctx context.Context, id int64) (err error)
	// DeleteReferences deletes the references referenced by the artifact specified by parent ID
	DeleteReferences(ctx context.Context, parentID int64) (err error)
}

const (
	// both tagged and untagged artifacts
	both = `IN (
		SELECT DISTINCT art.id FROM artifact art
		LEFT JOIN tag ON art.id=tag.artifact_id 
		LEFT JOIN artifact_reference ref ON art.id=ref.child_id
		WHERE tag.id IS NOT NULL OR ref.id IS NULL)`
	// only untagged artifacts
	untagged = `IN (
		SELECT DISTINCT art.id FROM artifact art
		LEFT JOIN tag ON art.id=tag.artifact_id
		WHERE tag.id IS NULL)`
	// only tagged artifacts
	tagged = `IN (
		SELECT DISTINCT art.id FROM artifact art
		JOIN tag ON art.id=tag.artifact_id
		WHERE tag.id IS NOT NULL)`
)

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Count(ctx context.Context, query *q.Query) (int64, error) {
	if query != nil {
		// ignore the page number and size
		query = &q.Query{
			Keywords: query.Keywords,
		}
	}
	qs, err := querySetter(ctx, query)
	if err != nil {
		return 0, err
	}
	return qs.Count()
}
func (d *dao) List(ctx context.Context, query *q.Query) ([]*Artifact, error) {
	artifacts := []*Artifact{}
	qs, err := querySetter(ctx, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&artifacts); err != nil {
		return nil, err
	}
	return artifacts, nil
}
func (d *dao) Get(ctx context.Context, id int64) (*Artifact, error) {
	artifact := &Artifact{
		ID: id,
	}
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err = ormer.Read(artifact); err != nil {
		if e := orm.AsNotFoundError(err, "artifact %d not found", id); e != nil {
			err = e
		}
		return nil, err
	}
	return artifact, nil
}

func (d *dao) GetByDigest(ctx context.Context, repository, digest string) (*Artifact, error) {
	qs, err := orm.QuerySetter(ctx, &Artifact{}, &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryName": repository,
			"Digest":         digest,
		},
	})
	if err != nil {
		return nil, err
	}
	artifacts := []*Artifact{}
	if _, err = qs.All(&artifacts); err != nil {
		return nil, err
	}
	if len(artifacts) == 0 {
		return nil, ierror.New(nil).WithCode(ierror.NotFoundCode).
			WithMessage("artifact %s@%s not found", repository, digest)
	}
	return artifacts[0], nil
}

func (d *dao) Create(ctx context.Context, artifact *Artifact) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(artifact)
	if err != nil {
		if e := orm.AsConflictError(err, "artifact %s already exists under the repository %d",
			artifact.Digest, artifact.RepositoryID); e != nil {
			err = e
		}
	}
	return id, err
}
func (d *dao) Delete(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&Artifact{
		ID: id,
	})
	if err != nil {
		if e := orm.AsForeignKeyError(err,
			"the artifact %d is referenced by other resources", id); e != nil {
			err = e
		}
		return err
	}
	if n == 0 {
		return ierror.NotFoundError(nil).WithMessage("artifact %d not found", id)
	}

	return nil
}
func (d *dao) Update(ctx context.Context, artifact *Artifact, props ...string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Update(artifact, props...)
	if err != nil {
		return err
	}
	if n == 0 {
		return ierror.NotFoundError(nil).WithMessage("artifact %d not found", artifact.ID)
	}
	return nil
}
func (d *dao) CreateReference(ctx context.Context, reference *ArtifactReference) (int64, error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	id, err := ormer.Insert(reference)
	if err != nil {
		if e := orm.AsConflictError(err, "reference already exists, parent artifact ID: %d, child artifact ID: %d",
			reference.ParentID, reference.ChildID); e != nil {
			err = e
		} else if e := orm.AsForeignKeyError(err, "the reference tries to reference a non existing artifact, parent artifact ID: %d, child artifact ID: %d",
			reference.ParentID, reference.ChildID); e != nil {
			err = e
		}
	}
	return id, err
}
func (d *dao) ListReferences(ctx context.Context, query *q.Query) ([]*ArtifactReference, error) {
	references := []*ArtifactReference{}
	qs, err := orm.QuerySetter(ctx, &ArtifactReference{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&references); err != nil {
		return nil, err
	}
	return references, nil
}

func (d *dao) DeleteReference(ctx context.Context, id int64) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&ArtifactReference{ID: id})
	if err != nil {
		return err
	}
	if n == 0 {
		return ierror.NotFoundError(nil).WithMessage("artifact reference %d not found", id)
	}
	return nil
}

func (d *dao) DeleteReferences(ctx context.Context, parentID int64) error {
	// make sure the parent artifact exist
	_, err := d.Get(ctx, parentID)
	if err != nil {
		return err
	}
	qs, err := orm.QuerySetter(ctx, &ArtifactReference{}, &q.Query{
		Keywords: map[string]interface{}{
			"parent_id": parentID,
		},
	})
	if err != nil {
		return err
	}
	_, err = qs.Delete()
	return err
}

func querySetter(ctx context.Context, query *q.Query) (beegoorm.QuerySeter, error) {
	qs, err := orm.QuerySetter(ctx, &Artifact{}, query)
	if err != nil {
		return nil, err
	}
	qs, err = setBaseQuery(qs, query)
	if err != nil {
		return nil, err
	}
	qs, err = setTagQuery(qs, query)
	if err != nil {
		return nil, err
	}
	qs, err = setLabelQuery(qs, query)
	if err != nil {
		return nil, err
	}
	return qs, nil
}

// handle q=base=*
// when "q=base=*" is specified in the query, the base collection is the all artifacts of database,
// otherwise the base connection is only the tagged artifacts and untagged artifacts that aren't
// referenced by others
func setBaseQuery(qs beegoorm.QuerySeter, query *q.Query) (beegoorm.QuerySeter, error) {
	if query == nil || len(query.Keywords) == 0 {
		qs = qs.FilterRaw("id", both)
		return qs, nil
	}
	base, exist := query.Keywords["base"]
	if !exist {
		qs = qs.FilterRaw("id", both)
		return qs, nil
	}
	b, ok := base.(string)
	if !ok || b != "*" {
		return qs, ierror.New(nil).WithCode(ierror.BadRequestCode).
			WithMessage(`the value of "base" query can only be exact match value with "*"`)
	}
	// the base is specified as "*"
	return qs, nil
}

// handle query string: q=tags=value q=tags=~value
func setTagQuery(qs beegoorm.QuerySeter, query *q.Query) (beegoorm.QuerySeter, error) {
	if query == nil || len(query.Keywords) == 0 {
		return qs, nil
	}
	tags, exist := query.Keywords["tags"]
	if !exist {
		tags, exist = query.Keywords["Tags"]
		if !exist {
			return qs, nil
		}
	}

	// fuzzy match
	f, ok := tags.(*q.FuzzyMatchValue)
	if ok {
		sql := fmt.Sprintf(`IN (
		SELECT DISTINCT art.id FROM artifact art
		JOIN tag ON art.id=tag.artifact_id
		WHERE tag.name LIKE '%%%s%%')`, orm.Escape(f.Value))
		qs = qs.FilterRaw("id", sql)
		return qs, nil
	}
	// exact match, only handle "*" for listing tagged artifacts and "nil" for listing untagged artifacts
	s, ok := tags.(string)
	if ok {
		if s == "*" {
			qs = qs.FilterRaw("id", tagged)
			return qs, nil
		}
		if s == "nil" {
			qs = qs.FilterRaw("id", untagged)
			return qs, nil
		}
	}
	return qs, ierror.New(nil).WithCode(ierror.BadRequestCode).
		WithMessage(`the value of "tags" query can only be fuzzy match value or exact match value with "*" and "nil"`)
}

// handle query string: q=labels=(1 2 3)
func setLabelQuery(qs beegoorm.QuerySeter, query *q.Query) (beegoorm.QuerySeter, error) {
	if query == nil || len(query.Keywords) == 0 {
		return qs, nil
	}
	labels, exist := query.Keywords["labels"]
	if !exist {
		labels, exist = query.Keywords["Labels"]
		if !exist {
			return qs, nil
		}
	}
	al, ok := labels.(*q.AndList)
	if !ok {
		return qs, ierror.New(nil).WithCode(ierror.BadRequestCode).
			WithMessage(`the value of "labels" query can only be integer list with intersetion relationship`)
	}
	var collections []string
	for _, value := range al.Values {
		labelID, ok := value.(int64)
		if !ok {
			return qs, ierror.New(nil).WithCode(ierror.BadRequestCode).
				WithMessage(`the value of "labels" query can only be integer list with intersetion relationship`)
		}
		collections = append(collections, fmt.Sprintf(`SELECT artifact_id FROM label_reference WHERE label_id=%d`, labelID))
	}
	qs = qs.FilterRaw("id", fmt.Sprintf(`IN (%s)`, strings.Join(collections, " INTERSECT ")))
	return qs, nil
}
