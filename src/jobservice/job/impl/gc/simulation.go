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

package gc

import (
	"context"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	pkgart "github.com/goharbor/harbor/src/pkg/artifact"
)

// deletionSimulation decides which artifacts artifact.Controller.Delete would remove, without removing anything.
// It follows the rules of deleteDeeply in src/controller/artifact/controller.go and has to be kept in sync with it.
type deletionSimulation struct {
	artCtl artifact.Controller
	artMgr pkgart.Manager
	// the artifacts and references deleted so far, in the order of deletion
	deleted     []*artifact.Artifact
	deletedRefs []int64
	artDeleted  map[int64]bool
	refDeleted  map[int64]bool
}

func newDeletionSimulation(artCtl artifact.Controller, artMgr pkgart.Manager) *deletionSimulation {
	return &deletionSimulation{
		artCtl:     artCtl,
		artMgr:     artMgr,
		artDeleted: map[int64]bool{},
		refDeleted: map[int64]bool{},
	}
}

// delete simulates the deletion of the artifact and everything a real deletion takes with it.
func (s *deletionSimulation) delete(ctx context.Context, id int64) error {
	deleted, deletedRefs := len(s.deleted), len(s.deletedRefs)
	if err := s.deleteDeeply(ctx, id, true); err != nil {
		// a real deletion runs in one transaction, so a failed one deletes nothing
		for _, art := range s.deleted[deleted:] {
			delete(s.artDeleted, art.ID)
		}
		for _, ref := range s.deletedRefs[deletedRefs:] {
			delete(s.refDeleted, ref)
		}
		s.deleted, s.deletedRefs = s.deleted[:deleted], s.deletedRefs[:deletedRefs]
		return err
	}
	return nil
}

func (s *deletionSimulation) deleteDeeply(ctx context.Context, id int64, isRoot bool) error {
	if s.artDeleted[id] {
		if isRoot {
			return errors.NotFoundError(nil).WithMessagef("artifact %d is already deleted", id)
		}
		return nil
	}
	art, err := s.artCtl.Get(ctx, id, &artifact.Option{WithTag: true, WithAccessory: true})
	if err != nil {
		if !isRoot && errors.IsErr(err, errors.NotFoundCode) {
			return nil
		}
		return err
	}

	// the child artifact is referenced by some tags, skip
	if !isRoot && len(art.Tags) > 0 {
		return nil
	}
	parents, err := s.parents(ctx, id)
	if err != nil {
		return err
	}
	if parents > 0 {
		if isRoot {
			return errors.New(nil).WithCode(errors.ViolateForeignKeyConstraintCode).
				WithMessage("the deleting artifact is referenced by others")
		}
		// the child artifact is referenced by other artifacts, skip
		return nil
	}

	for _, acc := range art.Accessories {
		// only hard ref accessory should be removed
		if !acc.IsHard() {
			continue
		}
		accParents, err := s.parents(ctx, acc.GetData().ArtifactID)
		if err != nil {
			return err
		}
		if err := s.deleteDeeply(ctx, acc.GetData().ArtifactID, accParents == 0); err != nil {
			return err
		}
	}

	for _, reference := range art.References {
		s.deletedRefs = append(s.deletedRefs, reference.ID)
		s.refDeleted[reference.ID] = true
		if err := s.deleteDeeply(ctx, reference.ChildID, false); err != nil {
			return err
		}
	}

	s.deleted = append(s.deleted, art)
	s.artDeleted[art.ID] = true
	return nil
}

// parents counts the references to the artifact that the simulation has not deleted yet
func (s *deletionSimulation) parents(ctx context.Context, id int64) (int, error) {
	references, err := s.artMgr.ListReferences(ctx, q.New(q.KeyWords{"ChildID": id}))
	if err != nil {
		return 0, err
	}
	count := 0
	for _, reference := range references {
		if !s.refDeleted[reference.ID] {
			count++
		}
	}
	return count, nil
}
