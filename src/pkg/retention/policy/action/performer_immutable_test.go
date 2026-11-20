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

package action

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/lib/q"
	isel "github.com/goharbor/harbor/src/lib/selector"
	immumodel "github.com/goharbor/harbor/src/pkg/immutable/model"
)

// fakeImmutableCtl implements immutable.Controller; ListImmutableRules fails to
// simulate a transient database error during a retention run.
type fakeImmutableCtl struct {
	listErr error
}

func (f *fakeImmutableCtl) GetImmutableRule(ctx context.Context, id int64) (*immumodel.Metadata, error) {
	return nil, errors.New("not implemented in fake")
}
func (f *fakeImmutableCtl) CreateImmutableRule(ctx context.Context, m *immumodel.Metadata) (int64, error) {
	return 0, errors.New("not implemented in fake")
}
func (f *fakeImmutableCtl) DeleteImmutableRule(ctx context.Context, id int64) error {
	return errors.New("not implemented in fake")
}
func (f *fakeImmutableCtl) UpdateImmutableRule(ctx context.Context, projectID int64, m *immumodel.Metadata) error {
	return errors.New("not implemented in fake")
}
func (f *fakeImmutableCtl) ListImmutableRules(ctx context.Context, query *q.Query) ([]*immumodel.Metadata, error) {
	return nil, f.listErr
}
func (f *fakeImmutableCtl) Count(ctx context.Context, query *q.Query) (int64, error) {
	return 0, errors.New("not implemented in fake")
}
func (f *fakeImmutableCtl) DeleteImmutableRuleByProject(ctx context.Context, projectID int64) error {
	return errors.New("not implemented in fake")
}

// TestRetentionDeletesImmutableArtifactOnMatcherError is the regression for the
// fail-open defect in isImmutable: when the immutable-rule listing fails (for
// example a transient database failure during a retention run), the deletion
// decision must fail closed. Before the fix, the candidate was treated as
// mutable and cleared for deletion.
func TestRetentionDeletesImmutableArtifactOnMatcherError(t *testing.T) {
	old := immutable.Ctr
	defer func() { immutable.Ctr = old }()
	immutable.Ctr = &fakeImmutableCtl{listErr: errors.New("simulated transient DB failure")}

	artifact := &isel.Candidate{
		NamespaceID: 1,
		Namespace:   "library",
		Repository:  "harbor",
		Kind:        "image",
		Tags:        []string{"release-1.0"},
		Digest:      "sha256:ea2b4bad746c80a91b3512b4d9e3e2f0e8f9a4a7e1c0b6d5a3f2e1d0c9b8a7965",
	}

	performer := NewRetainAction([]*isel.Candidate{artifact}, true /* isDryRun */)
	results, err := performer.Perform(context.Background(), nil /* nothing retained */)
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Error(t, results[0].Error,
		"artifact covered by an immutable rule must not be cleared for deletion when the immutable-rule listing fails")
}
