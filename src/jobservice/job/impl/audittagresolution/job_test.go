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

package audittagresolution

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	artpkg "github.com/goharbor/harbor/src/pkg/artifact"
	am "github.com/goharbor/harbor/src/pkg/auditext/model"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	mockAuditExt "github.com/goharbor/harbor/src/testing/pkg/auditext"
	mockArt "github.com/goharbor/harbor/src/testing/pkg/artifact"
	mockTag "github.com/goharbor/harbor/src/testing/pkg/tag"
)

func TestValidateParams(t *testing.T) {
	j := &Job{}
	assert.Error(t, j.Validate(nil))
	assert.Error(t, j.Validate(job.Parameters{}))
	assert.Error(t, j.Validate(job.Parameters{ParamAuditLogID: float64(1)}))
	assert.NoError(t, j.Validate(job.Parameters{
		ParamAuditLogID: float64(1),
		ParamRepository: "library/nginx",
		ParamDigest:     "sha256:abc",
	}))
}

func TestShouldRetry(t *testing.T) {
	j := &Job{}
	assert.True(t, j.ShouldRetry())
}

func TestRunResolvesTag(t *testing.T) {
	auditMgr := mockAuditExt.NewManager(t)
	artifactMgr := mockArt.NewManager(t)
	tagMgr := mockTag.NewManager(t)

	j := &Job{
		auditMgr:    auditMgr,
		artifactMgr: artifactMgr,
		tagMgr:      tagMgr,
	}

	ctx := &mockjobservice.MockJobContext{}
	ctx.On("GetLogger").Return(&mockjobservice.MockJobLogger{})

	params := job.Parameters{
		ParamAuditLogID: float64(42),
		ParamRepository: "library/nginx",
		ParamDigest:     "sha256:abc",
	}

	art := &artpkg.Artifact{}
	art.ID = 10
	artifactMgr.On("GetByDigest", context.TODO(), "library/nginx", "sha256:abc").Return(art, nil)

	tagMgr.On("List", context.TODO(), mock.MatchedBy(func(query *q.Query) bool {
		return query.Keywords["ArtifactID"] == int64(10)
	})).Return([]*tagmodel.Tag{
		{Name: "8.1.0"},
		{Name: "latest"},
	}, nil)

	entry := &am.AuditLogExt{
		ID:                   42,
		OperationDescription: "pull artifact: library/nginx@sha256:abc",
	}
	auditMgr.On("Get", context.TODO(), int64(42)).Return(entry, nil)
	auditMgr.On("Update", context.TODO(), mock.MatchedBy(func(a *am.AuditLogExt) bool {
		return a.OperationDescription == "pull artifact: library/nginx@sha256:abc; resolved tags: 8.1.0, latest"
	}), "OperationDescription").Return(nil)

	err := j.Run(ctx, params)
	assert.NoError(t, err)
}

func TestRunNoTags(t *testing.T) {
	auditMgr := mockAuditExt.NewManager(t)
	artifactMgr := mockArt.NewManager(t)
	tagMgr := mockTag.NewManager(t)

	j := &Job{
		auditMgr:    auditMgr,
		artifactMgr: artifactMgr,
		tagMgr:      tagMgr,
	}

	ctx := &mockjobservice.MockJobContext{}
	ctx.On("GetLogger").Return(&mockjobservice.MockJobLogger{})

	params := job.Parameters{
		ParamAuditLogID: float64(42),
		ParamRepository: "library/nginx",
		ParamDigest:     "sha256:abc",
	}

	art := &artpkg.Artifact{}
	art.ID = 10
	artifactMgr.On("GetByDigest", context.TODO(), "library/nginx", "sha256:abc").Return(art, nil)
	tagMgr.On("List", context.TODO(), mock.Anything).Return([]*tagmodel.Tag{}, nil)

	err := j.Run(ctx, params)
	assert.NoError(t, err)
}

func TestRunArtifactNotFound(t *testing.T) {
	auditMgr := mockAuditExt.NewManager(t)
	artifactMgr := mockArt.NewManager(t)
	tagMgr := mockTag.NewManager(t)

	j := &Job{
		auditMgr:    auditMgr,
		artifactMgr: artifactMgr,
		tagMgr:      tagMgr,
	}

	ctx := &mockjobservice.MockJobContext{}
	ctx.On("GetLogger").Return(&mockjobservice.MockJobLogger{})

	params := job.Parameters{
		ParamAuditLogID: float64(42),
		ParamRepository: "library/nginx",
		ParamDigest:     "sha256:abc",
	}

	artifactMgr.On("GetByDigest", context.TODO(), "library/nginx", "sha256:abc").
		Return(nil, errors.NotFoundError(nil).WithMessage("not found"))

	err := j.Run(ctx, params)
	assert.NoError(t, err)
}
