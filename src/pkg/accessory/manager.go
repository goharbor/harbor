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

package accessory

import (
	"context"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory/dao"
	"github.com/goharbor/harbor/src/pkg/accessory/model"

	_ "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
)

var (
	// Mgr is a global artifact manager instance
	Mgr = NewManager()
)

// Manager is the only interface of artifact module to provide the management functions for artifacts
type Manager interface {
	// Get the artifact specified by the ID
	Get(ctx context.Context, id int64) (accessory model.Accessory, err error)
	// Count returns the total count of tags according to the query.
	Count(ctx context.Context, query *q.Query) (total int64, err error)
	// List tags according to the query
	List(ctx context.Context, query *q.Query) (accs []model.Accessory, err error)
	// Create the tag and returns the ID
	Create(ctx context.Context, accessory model.AccessoryData) (id int64, err error)
	// Delete the tag specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// DeleteAccessories deletes accessories according to the query
	DeleteAccessories(ctx context.Context, q *q.Query) (err error)
}

// NewManager returns an instance of the default manager
func NewManager() Manager {
	return &manager{
		dao.New(),
	}
}

var _ Manager = &manager{}

type manager struct {
	dao dao.DAO
}

func (m *manager) Get(ctx context.Context, id int64) (model.Accessory, error) {
	acc, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return model.New(acc.Type, model.AccessoryData{
		ID:            acc.ID,
		ArtifactID:    acc.ArtifactID,
		SubArtifactID: acc.SubjectArtifactID,
		Size:          acc.Size,
		Digest:        acc.Digest,
		CreatTime:     acc.CreationTime,
	})
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]model.Accessory, error) {
	accsDao, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var accs []model.Accessory
	for _, accD := range accsDao {
		acc, err := model.New(accD.Type, model.AccessoryData{
			ID:            accD.ID,
			ArtifactID:    accD.ArtifactID,
			SubArtifactID: accD.SubjectArtifactID,
			Size:          accD.Size,
			Digest:        accD.Digest,
			CreatTime:     accD.CreationTime,
		})
		if err != nil {
			return nil, errors.New(err).WithCode(errors.BadRequestCode)
		}
		accs = append(accs, acc)
	}
	return accs, nil
}

func (m *manager) Create(ctx context.Context, accessory model.AccessoryData) (int64, error) {
	acc := &dao.Accessory{
		ArtifactID:        accessory.ArtifactID,
		SubjectArtifactID: accessory.SubArtifactID,
		Size:              accessory.Size,
		Digest:            accessory.Digest,
		Type:              accessory.Type,
	}
	return m.dao.Create(ctx, acc)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

func (m *manager) DeleteAccessories(ctx context.Context, q *q.Query) error {
	_, err := m.dao.DeleteAccessories(ctx, q)
	return err
}
