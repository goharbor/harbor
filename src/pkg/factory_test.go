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

package pkg

import (
	"testing"

	"github.com/goharbor/harbor/src/pkg/artifact"
	cachedArtifact "github.com/goharbor/harbor/src/pkg/cached/artifact/redis"
	cachedProject "github.com/goharbor/harbor/src/pkg/cached/project/redis"
	cachedProjectMeta "github.com/goharbor/harbor/src/pkg/cached/project_metadata/redis"
	cachedRepo "github.com/goharbor/harbor/src/pkg/cached/repository/redis"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/project/metadata"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/stretchr/testify/assert"
)

func TestInitArtifactMgr(t *testing.T) {
	// cache not enable
	assert.NotNil(t, ArtifactMgr)
	assert.IsType(t, artifact.NewManager(), ArtifactMgr)

	// cache enable
	initArtifactMgr(true)
	assert.NotNil(t, ArtifactMgr)
	assert.IsType(t, cachedArtifact.NewManager(artifact.NewManager()), ArtifactMgr)
}

func TestInitProjectMgr(t *testing.T) {
	// cache not enable
	assert.NotNil(t, ProjectMgr)
	assert.IsType(t, project.New(), ProjectMgr)

	// cache enable
	initProjectMgr(true)
	assert.NotNil(t, ProjectMgr)
	assert.IsType(t, cachedProject.NewManager(project.New()), ProjectMgr)
}

func TestInitProjectMetaMgr(t *testing.T) {
	// cache not enable
	assert.NotNil(t, ProjectMetaMgr)
	assert.IsType(t, metadata.New(), ProjectMetaMgr)

	// cache enable
	initProjectMetaMgr(true)
	assert.NotNil(t, ProjectMetaMgr)
	assert.IsType(t, cachedProjectMeta.NewManager(metadata.New()), ProjectMetaMgr)
}

func TestInitRepositoryMgr(t *testing.T) {
	// cache not enable
	assert.NotNil(t, RepositoryMgr)
	assert.IsType(t, repository.New(), RepositoryMgr)

	// cache enable
	initRepositoryMgr(true)
	assert.NotNil(t, RepositoryMgr)
	assert.IsType(t, cachedRepo.NewManager(repository.New()), RepositoryMgr)
}
