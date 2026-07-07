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
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/config"
	pkgartifact "github.com/goharbor/harbor/src/pkg/artifact"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
	repotesting "github.com/goharbor/harbor/src/testing/pkg/repository"

	"github.com/goharbor/harbor/src/pkg"
)

func TestConstructArtifactPayload_RepoType(t *testing.T) {
	config.DefaultCfgManager = common.InMemoryCfgManager

	originalRepositoryMgr := pkg.RepositoryMgr
	repoMgr := &repotesting.Manager{}
	repoMgr.On("GetByName", mock.Anything, mock.Anything).
		Return(&repomodel.RepoRecord{}, nil)
	pkg.RepositoryMgr = repoMgr
	defer func() { pkg.RepositoryMgr = originalRepositoryMgr }()

	cases := []struct {
		name     string
		metadata map[string]string
		wantType string
	}{
		{"public", map[string]string{"public": "true"}, proModels.ProjectPublic},
		{"private", map[string]string{"public": "false"}, proModels.ProjectPrivate},
		{"auth_only", map[string]string{"public": "auth_only"}, proModels.ProjectAuthOnly},
	}

	a := &Handler{}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			evt := &event.ArtifactEvent{
				EventType:  "PUSH_ARTIFACT",
				Repository: "library/hello-world",
				Artifact:   &pkgartifact.Artifact{ProjectID: 1, Digest: "sha256:abc"},
			}
			prj := &proModels.Project{
				ProjectID: 1,
				Name:      "library",
				Metadata:  tc.metadata,
			}

			payload, err := a.constructArtifactPayload(context.Background(), evt, prj)
			require.NoError(t, err)
			require.Equal(t, tc.wantType, payload.EventData.Repository.RepoType)
		})
	}
}
