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

package replication

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	commondao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/project"
	replicationctl "github.com/goharbor/harbor/src/controller/replication"
	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	pkgartifact "github.com/goharbor/harbor/src/pkg/artifact"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	regmodel "github.com/goharbor/harbor/src/pkg/reg/model"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	replicationtesting "github.com/goharbor/harbor/src/testing/controller/replication"
)

// TestMain registers (but does not migrate) a "default" ORM alias so that
// orm.Context(), which handlePushArtifact/handleCreateTag call unconditionally,
// doesn't panic. The tests below never issue a real query against it -
// project.Ctl and replication.Ctl are fully mocked - so no schema is needed,
// only a reachable Postgres to satisfy beego orm's registration check.
func TestMain(m *testing.M) {
	port, _ := strconv.Atoi(os.Getenv("POSTGRESQL_PORT"))
	db := commondao.NewPGSQL(
		os.Getenv("POSTGRESQL_HOST"),
		os.Getenv("POSTGRESQL_PORT"),
		os.Getenv("POSTGRESQL_USR"),
		os.Getenv("POSTGRESQL_PWD"),
		os.Getenv("POSTGRESQL_DATABASE"),
		"disable",
		50, 100, 0, 0,
	)
	if port == 0 || db.Register() != nil {
		// no reachable test database configured; skip rather than fail the
		// whole package, since these tests only need the "default" ORM
		// alias to exist, not real data.
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// setupAlwaysMatchPolicy configures the replication controller mock to return a
// single enabled, event-triggered, filter-less policy so every resource passed
// to Handle reaches replication.Ctl.Start, letting the test observe exactly
// what "public" metadata value was propagated for the pushed artifact.
func setupAlwaysMatchPolicy(t *testing.T, wantPublic string) *replicationtesting.Controller {
	repCtl := &replicationtesting.Controller{}
	repCtl.On("ListPolicies", mock.Anything, mock.Anything).
		Return([]*repctlmodel.Policy{{ID: 1, Enabled: true}}, nil).Once()
	repCtl.On("Start", mock.Anything, mock.Anything, mock.MatchedBy(func(resource *regmodel.Resource) bool {
		return resource.Metadata.Repository.Metadata["public"] == wantPublic
	}), mock.Anything).Return(int64(1), nil).Once()
	t.Cleanup(func() { repCtl.AssertExpectations(t) })
	return repCtl
}

func TestHandlePushArtifact_PropagatesAccessLevel(t *testing.T) {
	cases := []struct {
		name       string
		metadata   map[string]string
		wantPublic string
	}{
		{"public project", map[string]string{"public": "true"}, "true"},
		{"private project", map[string]string{"public": "false"}, "false"},
		{"auth_only project", map[string]string{"public": "auth_only"}, "auth_only"},
		{"missing metadata defaults to false", map[string]string{}, "false"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			originalProjectCtl := project.Ctl
			originalReplicationCtl := replicationctl.Ctl
			defer func() {
				project.Ctl = originalProjectCtl
				replicationctl.Ctl = originalReplicationCtl
			}()

			projectCtl := &projecttesting.Controller{}
			projectCtl.On("Get", mock.Anything, mock.Anything, mock.Anything).
				Return(&proModels.Project{ProjectID: 1, Metadata: tc.metadata}, nil).Once()
			project.Ctl = projectCtl

			replicationctl.Ctl = setupAlwaysMatchPolicy(t, tc.wantPublic)

			h := &Handler{}
			err := h.Handle(context.Background(), &event.PushArtifactEvent{
				ArtifactEvent: &event.ArtifactEvent{
					Repository: "library/hello-world",
					Artifact:   &pkgartifact.Artifact{ProjectID: 1, Type: "IMAGE", Digest: "sha256:abc"},
				},
			})
			require.NoError(t, err)

			projectCtl.AssertExpectations(t)
		})
	}
}

func TestHandleCreateTag_PropagatesAccessLevel(t *testing.T) {
	originalProjectCtl := project.Ctl
	originalReplicationCtl := replicationctl.Ctl
	defer func() {
		project.Ctl = originalProjectCtl
		replicationctl.Ctl = originalReplicationCtl
	}()

	projectCtl := &projecttesting.Controller{}
	projectCtl.On("Get", mock.Anything, mock.Anything, mock.Anything).
		Return(&proModels.Project{ProjectID: 1, Metadata: map[string]string{"public": "auth_only"}}, nil).Once()
	project.Ctl = projectCtl

	replicationctl.Ctl = setupAlwaysMatchPolicy(t, "auth_only")

	h := &Handler{}
	err := h.Handle(context.Background(), &event.CreateTagEvent{
		Repository:       "library/hello-world",
		Tag:              "latest",
		AttachedArtifact: &pkgartifact.Artifact{ProjectID: 1, Type: "IMAGE", Digest: "sha256:abc"},
	})
	require.NoError(t, err)

	projectCtl.AssertExpectations(t)
}
