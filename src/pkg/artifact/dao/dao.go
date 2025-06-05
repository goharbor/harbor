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
	"time"

	beegoorm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
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
	// UpdatePullTime updates artifact pull time by ID.
	UpdatePullTime(ctx context.Context, id int64, pullTime time.Time) (err error)
	// CreateReference creates the artifact reference
	CreateReference(ctx context.Context, reference *ArtifactReference) (id int64, err error)
	// ListReferences lists the artifact references according to the query
	ListReferences(ctx context.Context, query *q.Query) (references []*ArtifactReference, err error)
	// DeleteReference specified by ID
	DeleteReference(ctx context.Context, id int64) (err error)
	// DeleteReferences deletes the references referenced by the artifact specified by parent ID
	DeleteReferences(ctx context.Context, parentID int64) (err error)
	// ListWithLatest ...
	ListWithLatest(ctx context.Context, query *q.Query) (artifacts []*Artifact, err error)
}

const (
	// the QuerySetter of beego doesn't support "EXISTS" directly, use qs.FilterRaw("id", "=id AND xxx") to workaround the limitation
	// base filter: both tagged and untagged artifacts
	both = `=id AND (
		EXISTS (SELECT 1 FROM tag WHERE tag.artifact_id = T0.id)
		OR 
		NOT EXISTS (SELECT 1 FROM artifact_reference ref WHERE ref.child_id = T0.id)
	)`
	// tag filter: only untagged artifacts
	// the "untagged" filter is based on "base" filter, so we consider the tag only
	untagged = `=id AND NOT EXISTS(
		SELECT 1 FROM tag WHERE tag.artifact_id = T0.id
	)`
	// tag filter: only tagged artifacts
	tagged = `=id AND EXISTS (
		SELECT 1 FROM tag WHERE tag.artifact_id = T0.id
	)`
	// accessory filter: filter out the accessory
	notacc = `=id AND NOT EXISTS (
		SELECT 1 FROM artifact_accessory aa WHERE aa.artifact_id = T0.id
	)`
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
	qs, err := querySetter(ctx, query, orm.WithSortDisabled(true))
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
		Keywords: map[string]any{
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
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("artifact %s@%s not found", repository, digest)
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
		return errors.NotFoundError(nil).WithMessagef("artifact %d not found", id)
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
		return errors.NotFoundError(nil).WithMessagef("artifact %d not found", artifact.ID)
	}
	return nil
}

func (d *dao) UpdatePullTime(ctx context.Context, id int64, pullTime time.Time) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	// can only be retained to the second if not format
	formatPullTime := pullTime.Format("2006-01-02 15:04:05.999999")
	// update db only if pull_time is null or pull_time < (in-coming)pullTime
	sql := "UPDATE artifact SET pull_time = ? WHERE id = ? AND (pull_time IS NULL OR pull_time < ?)"
	args := []any{formatPullTime, id, formatPullTime}

	_, err = ormer.Raw(sql, args...).Exec()
	if err != nil {
		return err
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
		return errors.NotFoundError(nil).WithMessagef("artifact reference %d not found", id)
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
		Keywords: map[string]any{
			"parent_id": parentID,
		},
	})
	if err != nil {
		return err
	}
	_, err = qs.Delete()
	return err
}

func (d *dao) ListWithLatest(ctx context.Context, query *q.Query) (artifacts []*Artifact, err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	sql := `SELECT a.*
	FROM artifact a
	JOIN (
		SELECT repository_name, MAX(push_time) AS latest_push_time
	FROM artifact
	WHERE project_id = ? and %s = ?
	GROUP BY repository_name
	) latest ON a.repository_name = latest.repository_name AND a.push_time = latest.latest_push_time`

	queryParam := make([]any, 0)
	var ok bool
	var pid any
	if pid, ok = query.Keywords["ProjectID"]; !ok {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage(`the value of "ProjectID" must be set`)
	}
	queryParam = append(queryParam, pid)

	var attributionValue any
	if attributionValue, ok = query.Keywords["media_type"]; ok {
		sql = fmt.Sprintf(sql, "media_type")
	} else if attributionValue, ok = query.Keywords["artifact_type"]; ok {
		sql = fmt.Sprintf(sql, "artifact_type")
	}

	if attributionValue == "" {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage(`the value of "media_type" or "artifact_type" must be set`)
	}
	queryParam = append(queryParam, attributionValue)

	sql, queryParam = orm.PaginationOnRawSQL(query, sql, queryParam)
	arts := []*Artifact{}
	_, err = ormer.Raw(sql, queryParam...).QueryRows(&arts)
	if err != nil {
		return nil, err
	}

	return arts, nil
}

func querySetter(ctx context.Context, query *q.Query, options ...orm.Option) (beegoorm.QuerySeter, error) {
	qs, err := orm.QuerySetter(ctx, &Artifact{}, query, options...)
	if err != nil {
		return nil, err
	}
	qs, err = setBaseQuery(qs, query)
	if err != nil {
		return nil, err
	}
	qs, err = setTagQuery(ctx, qs, query)
	if err != nil {
		return nil, err
	}
	qs, err = setLabelQuery(qs, query)
	if err != nil {
		return nil, err
	}
	qs, err = setAccessoryQuery(qs, query)
	if err != nil {
		return nil, err
	}
	return qs, nil
}

// handle q=base=*
// when "q=base=*" is specified in the query, the base collection is the all artifacts of database,
// otherwise the base collection is only the tagged artifacts and untagged artifacts that aren't
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
		return qs, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage(`the value of "base" query can only be exact match value with "*"`)
	}
	// the base is specified as "*"
	return qs, nil
}

// handle query string: q=tags=value q=tags=~value
func setTagQuery(ctx context.Context, qs beegoorm.QuerySeter, query *q.Query) (beegoorm.QuerySeter, error) {
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
		// get the id list first to avoid the sql injection
		inClause, err := orm.CreateInClause(ctx, `SELECT DISTINCT art.id FROM artifact art
			JOIN tag ON art.id=tag.artifact_id
			WHERE tag.name LIKE ?`, "%"+orm.Escape(f.Value)+"%")
		if err != nil {
			return nil, err
		}
		qs = qs.FilterRaw("id", inClause)
		return qs, nil
	}
	// exact match:
	// "*" for listing tagged artifacts
	// "nil" for listing untagged artifacts
	// others for get the artifact with the specified tag
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

		// get the id list first to avoid the sql injection
		inClause, err := orm.CreateInClause(ctx, `SELECT DISTINCT art.id FROM artifact art
			JOIN tag ON art.id=tag.artifact_id
			WHERE tag.name = ?`, s)
		if err != nil {
			return nil, err
		}
		qs = qs.FilterRaw("id", inClause)
		return qs, nil
	}
	return qs, errors.New(nil).WithCode(errors.BadRequestCode).
		WithMessage(`the value of "tags" query can only be fuzzy match value or exact match value`)
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
		return qs, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage(`the value of "labels" query can only be integer list with intersetion relationship`)
	}
	var collections []string
	for _, value := range al.Values {
		labelID, ok := value.(int64)
		if !ok {
			return qs, errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessage(`the value of "labels" query can only be integer list with intersetion relationship`)
		}
		// param "labelID" is integer, no need to sanitize
		collections = append(collections, fmt.Sprintf(`SELECT artifact_id FROM label_reference WHERE label_id=%d`, labelID))
	}
	qs = qs.FilterRaw("id", fmt.Sprintf(`IN (%s)`, strings.Join(collections, " INTERSECT ")))
	return qs, nil
}

// filter out the accessory for results
func setAccessoryQuery(qs beegoorm.QuerySeter, query *q.Query) (beegoorm.QuerySeter, error) {
	if query == nil {
		return qs, nil
	}

	qs = qs.FilterRaw("id", notacc)
	return qs, nil
}
