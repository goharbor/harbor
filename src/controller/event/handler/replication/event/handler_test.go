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

package event

import (
	"context"
	"github.com/goharbor/harbor/src/controller/replication"
	repmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/accessory"
	accmodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/tag"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	"github.com/goharbor/harbor/src/pkg/task"
	mockrep "github.com/goharbor/harbor/src/testing/controller/replication"
	mockacc "github.com/goharbor/harbor/src/testing/pkg/accessory"
	mockaccmodel "github.com/goharbor/harbor/src/testing/pkg/accessory/model"
	mockart "github.com/goharbor/harbor/src/testing/pkg/artifact"
	mocktag "github.com/goharbor/harbor/src/testing/pkg/tag"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHandleAppendsAccessoriesPreservingEventTags(t *testing.T) {
	ctl := &mockrep.Controller{}
	old := replication.Ctl
	replication.Ctl = ctl
	t.Cleanup(func() { replication.Ctl = old })
	p := &repmodel.Policy{Filters: tagFilter("release")}
	ctl.On("ListPolicies", mock.Anything, mock.Anything).Return([]*repmodel.Policy{p}, nil).Once()
	ctl.On("Start", mock.Anything, p, mock.MatchedBy(func(r *model.Resource) bool {
		return len(r.Metadata.Artifacts) == 2 && r.Metadata.Artifacts[0].Digest == "subject" && len(r.Metadata.Artifacts[0].Tags) == 2 && r.Metadata.Artifacts[0].Tags[0] == "release" && r.Metadata.Artifacts[1].Digest == "signature"
	}), task.ExecutionTriggerEvent).Return(int64(1), nil).Once()
	mockDiscovery(t, "subject", "signature")
	require.NoError(t, Handle(context.Background(), &Event{Type: EventTypeArtifactPush, Resource: seed("subject", "release", "other")}))
	ctl.AssertExpectations(t)
}
func TestDeletionPreservesOriginalResource(t *testing.T) {
	for _, eventType := range []string{EventTypeTagDelete, EventTypeArtifactDelete} {
		t.Run(eventType, func(t *testing.T) {
			ctl := &mockrep.Controller{}
			old := replication.Ctl
			replication.Ctl = ctl
			t.Cleanup(func() { replication.Ctl = old })
			p := &repmodel.Policy{ReplicateDeletion: true, Filters: tagFilter("release")}
			r := seed("deleted", "release")
			r.Deleted = true
			r.IsDeleteTag = eventType == EventTypeTagDelete
			if !r.IsDeleteTag {
				r.Metadata.Artifacts[0].Tags = append(r.Metadata.Artifacts[0].Tags, "other")
			}
			ctl.On("ListPolicies", mock.Anything, mock.Anything).Return([]*repmodel.Policy{p}, nil).Once()
			ctl.On("Start", mock.Anything, p, mock.MatchedBy(func(actual *model.Resource) bool {
				return actual == r && (r.IsDeleteTag || len(actual.Metadata.Artifacts[0].Tags) == 2)
			}), task.ExecutionTriggerEvent).Return(int64(1), nil).Once()
			require.NoError(t, Handle(context.Background(), &Event{Type: eventType, Resource: r}))
			ctl.AssertExpectations(t)
		})
	}
}

func TestNonmatchingEventDoesNotReadGraph(t *testing.T) {
	for _, filters := range [][]*model.Filter{
		tagFilter("release"),
		{{Type: model.FilterTypeLabel, Value: []string{"approved"}}},
	} {
		ctl := &mockrep.Controller{}
		old := replication.Ctl
		replication.Ctl = ctl
		p := &repmodel.Policy{Filters: filters}
		ctl.On("ListPolicies", mock.Anything, mock.Anything).Return([]*repmodel.Policy{p}, nil).Once()
		err := Handle(context.Background(), &Event{Type: EventTypeArtifactPush, Resource: seed("root", "other")})
		replication.Ctl = old
		require.NoError(t, err)
		ctl.AssertExpectations(t)
		ctl.AssertNotCalled(t, "Start", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	}
}
func TestPoliciesUseEventLabelsAndPreserveOriginalTags(t *testing.T) {
	ctl := &mockrep.Controller{}
	old := replication.Ctl
	replication.Ctl = ctl
	t.Cleanup(func() { replication.Ctl = old })
	policies := []*repmodel.Policy{}
	for _, tag := range []string{"one", "two"} {
		p := &repmodel.Policy{Filters: append(tagFilter(tag), &model.Filter{Type: model.FilterTypeLabel, Value: []string{"approved"}})}
		policies = append(policies, p)
		ctl.On("Start", mock.Anything, p, mock.MatchedBy(func(r *model.Resource) bool {
			return len(r.Metadata.Artifacts) == 2 && len(r.Metadata.Artifacts[0].Tags) == 2 && r.Metadata.Artifacts[0].Tags[0] == "one" && r.Metadata.Artifacts[0].Tags[1] == "two" && r.Metadata.Artifacts[1].Digest == "sig"
		}), task.ExecutionTriggerEvent).Return(int64(1), nil).Once()
	}
	ctl.On("ListPolicies", mock.Anything, mock.Anything).Return(policies, nil).Once()
	r := seed("root", "one", "two")
	r.Metadata.Artifacts[0].Labels = []string{"approved"}
	mockDiscovery(t, "root", "sig")
	require.NoError(t, Handle(context.Background(), &Event{Type: EventTypeArtifactPush, Resource: r}))
	require.Equal(t, []string{"one", "two"}, r.Metadata.Artifacts[0].Tags)
	ctl.AssertExpectations(t)
}

// Exercise Handle through the real reader while keeping database access mocked.
func mockDiscovery(t *testing.T, root, signature string) {
	mockArtifactGraph(t, graphReader{
		root:      {artifact: model.Artifact{Digest: root}, accessories: []string{signature}},
		signature: {artifact: model.Artifact{Digest: signature}},
	})
}

func mockArtifactGraph(t *testing.T, graph graphReader) {
	t.Helper()
	am, tm, acm := &mockart.Manager{}, &mocktag.Manager{}, &mockacc.Manager{}
	oldArt, oldTag, oldAcc := pkg.ArtifactMgr, tag.Mgr, accessory.Mgr
	pkg.ArtifactMgr, tag.Mgr, accessory.Mgr = am, tm, acm
	t.Cleanup(func() {
		pkg.ArtifactMgr, tag.Mgr, accessory.Mgr = oldArt, oldTag, oldAcc
		am.AssertExpectations(t)
		tm.AssertExpectations(t)
		acm.AssertExpectations(t)
	})
	var id int64
	for digest, node := range graph {
		id++
		artifactID := id
		art := &artifact.Artifact{ID: artifactID, Digest: digest}
		for _, child := range node.children {
			art.References = append(art.References, &artifact.Reference{ChildDigest: child})
		}
		am.On("GetByDigest", mock.Anything, "test/harbor-core", digest).Return(art, nil).Once()
		tm.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool { return query.Keywords["ArtifactID"] == artifactID })).Return([]*tagmodel.Tag{}, nil).Once()
		var accs []accmodel.Accessory
		for _, accessoryDigest := range node.accessories {
			a := &mockaccmodel.Accessory{}
			a.On("GetData").Return(accmodel.AccessoryData{Digest: accessoryDigest}).Once()
			t.Cleanup(func() { a.AssertExpectations(t) })
			accs = append(accs, a)
		}
		acm.On("List", mock.Anything, mock.MatchedBy(func(query *q.Query) bool {
			return query.Keywords["SubjectArtifactRepo"] == "test/harbor-core" && query.Keywords["SubjectArtifactDigest"] == digest
		})).Return(accs, nil).Once()
	}
}

func TestHandleStartsReplicationForIndexSubjectBackEdge(t *testing.T) {
	ctl := &mockrep.Controller{}
	old := replication.Ctl
	replication.Ctl = ctl
	t.Cleanup(func() { replication.Ctl = old })
	policy := &repmodel.Policy{Enabled: true, Trigger: &model.Trigger{Type: model.TriggerTypeEventBased}, Filters: tagFilter("release")}
	ctl.On("ListPolicies", mock.Anything, mock.Anything).Return([]*repmodel.Policy{policy}, nil).Once()
	ctl.On("Start", mock.Anything, policy, mock.MatchedBy(func(r *model.Resource) bool {
		return len(r.Metadata.Artifacts) == 2 && r.Metadata.Artifacts[0].Digest == "index" && r.Metadata.Artifacts[1].Digest == "sig"
	}), task.ExecutionTriggerEvent).Return(int64(1), nil).Once()
	mockArtifactGraph(t, graphReader{
		"index":   {children: []string{"subject"}, accessories: []string{"sig"}},
		"subject": {accessories: []string{"index"}},
		"sig":     {artifact: model.Artifact{Digest: "sig"}},
	})
	require.NoError(t, Handle(context.Background(), &Event{Type: EventTypeArtifactPush, Resource: seed("index", "release")}))
	ctl.AssertExpectations(t)
}
