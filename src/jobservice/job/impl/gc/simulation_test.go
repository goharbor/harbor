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
	"testing"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	pkgart "github.com/goharbor/harbor/src/pkg/artifact"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	"github.com/goharbor/harbor/src/testing/mock"
	arttesting "github.com/goharbor/harbor/src/testing/pkg/artifact"
)

func TestDeletionSimulationDiscardsFailedDeletion(t *testing.T) {
	artCtl := &artifacttesting.Controller{}
	artMgr := &arttesting.Manager{}
	index := func(id int64, refs ...*pkgart.Reference) *artifact.Artifact {
		return &artifact.Artifact{Artifact: pkgart.Artifact{ID: id, ManifestMediaType: v1.MediaTypeImageIndex, References: refs}}
	}
	parents := func(childID int64, refs ...*pkgart.Reference) {
		artMgr.On("ListReferences", mock.Anything, testifymock.MatchedBy(func(query *q.Query) bool {
			return query.Keywords["ChildID"] == childID
		})).Return(refs, nil)
	}

	// index 1 fails on its second child, index 2 shares the first child with it
	ref1to10 := &pkgart.Reference{ID: 100, ParentID: 1, ChildID: 10}
	ref1to11 := &pkgart.Reference{ID: 101, ParentID: 1, ChildID: 11}
	ref2to10 := &pkgart.Reference{ID: 102, ParentID: 2, ChildID: 10}
	artCtl.On("Get", mock.Anything, int64(1), mock.Anything).Return(index(1, ref1to10, ref1to11), nil)
	artCtl.On("Get", mock.Anything, int64(2), mock.Anything).Return(index(2, ref2to10), nil)
	artCtl.On("Get", mock.Anything, int64(10), mock.Anything).Return(&artifact.Artifact{Artifact: pkgart.Artifact{ID: 10}}, nil)
	artCtl.On("Get", mock.Anything, int64(11), mock.Anything).Return(nil, errors.New("database unavailable"))
	parents(1)
	parents(2)
	parents(10, ref1to10, ref2to10)
	parents(11, ref1to11)

	s := newDeletionSimulation(artCtl, artMgr)
	require.Error(t, s.delete(context.TODO(), 1))
	require.NoError(t, s.delete(context.TODO(), 2))

	// the failed deletion of index 1 keeps its reference to artifact 10, so artifact 10 survives index 2 as well
	var deleted []int64
	for _, art := range s.deleted {
		deleted = append(deleted, art.ID)
	}
	assert.Equal(t, []int64{2}, deleted)
}
