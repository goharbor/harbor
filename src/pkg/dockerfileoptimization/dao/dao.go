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
	"time"

	beegoorm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

// DAO is the data access object for dockerfile_optimization.
type DAO interface {
	// Upsert creates or replaces the optimization record for a given artifact.
	Upsert(ctx context.Context, rec *DockerfileOptimization) error
	// GetByArtifact returns the record for (repositoryName, artifactDigest),
	// or a NotFoundError if none exists.
	GetByArtifact(ctx context.Context, repositoryName, artifactDigest string) (*DockerfileOptimization, error)
	// UpdateStatus sets the status and error message of the record for
	// (repositoryName, artifactDigest) without touching its content columns.
	UpdateStatus(ctx context.Context, repositoryName, artifactDigest, status, errMsg string) error
}

// New returns an instance of the default DAO.
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) Upsert(ctx context.Context, rec *DockerfileOptimization) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	_, err = ormer.InsertOrUpdate(rec, "repository_name, artifact_digest")
	return err
}

func (d *dao) UpdateStatus(ctx context.Context, repositoryName, artifactDigest, status, errMsg string) error {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	count, err := ormer.QueryTable(&DockerfileOptimization{}).
		Filter("repository_name", repositoryName).
		Filter("artifact_digest", artifactDigest).
		Update(beegoorm.Params{
			"status":      status,
			"error":       errMsg,
			"update_time": time.Now(),
		})
	if err != nil {
		return err
	}
	if count == 0 {
		return errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("no dockerfile optimization found for artifact %s in %s", artifactDigest, repositoryName)
	}
	return nil
}

func (d *dao) GetByArtifact(ctx context.Context, repositoryName, artifactDigest string) (*DockerfileOptimization, error) {
	results := []*DockerfileOptimization{}
	qs, err := orm.QuerySetter(ctx, &DockerfileOptimization{}, &q.Query{
		Keywords: map[string]any{
			"RepositoryName": repositoryName,
			"ArtifactDigest": artifactDigest,
		},
	})
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&results); err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessagef("no dockerfile optimization found for artifact %s in %s", artifactDigest, repositoryName)
	}
	return results[0], nil
}
