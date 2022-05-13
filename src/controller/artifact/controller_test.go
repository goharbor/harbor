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
	"time"

	"github.com/goharbor/harbor/src/controller/artifact/processor/chart"
	"github.com/goharbor/harbor/src/controller/artifact/processor/cnab"
	"github.com/goharbor/harbor/src/controller/artifact/processor/image"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/icon"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/label/model"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
	model_tag "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	tagtesting "github.com/goharbor/harbor/src/testing/controller/tag"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/pkg/accessory"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
	arttesting "github.com/goharbor/harbor/src/testing/pkg/artifact"
	artrashtesting "github.com/goharbor/harbor/src/testing/pkg/artifactrash"
	"github.com/goharbor/harbor/src/testing/pkg/blob"
	"github.com/goharbor/harbor/src/testing/pkg/immutable"
	"github.com/goharbor/harbor/src/testing/pkg/label"
	"github.com/goharbor/harbor/src/testing/pkg/registry"
	repotesting "github.com/goharbor/harbor/src/testing/pkg/repository"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// TODO find another way to test artifact controller, it's hard to maintain currently

type fakeAbstractor struct {
	mock.Mock
}

func (f *fakeAbstractor) AbstractMetadata(ctx context.Context, artifact *artifact.Artifact) error {
	args := f.Called()
	return args.Error(0)
}

type controllerTestSuite struct {
	suite.Suite
	ctl          *controller
	repoMgr      *repotesting.Manager
	artMgr       *arttesting.Manager
	artrashMgr   *artrashtesting.FakeManager
	blobMgr      *blob.Manager
	tagCtl       *tagtesting.FakeController
	labelMgr     *label.Manager
	abstractor   *fakeAbstractor
	immutableMtr *immutable.FakeMatcher
	regCli       *registry.FakeClient
	accMgr       *accessory.Manager
}

func (c *controllerTestSuite) SetupTest() {
	c.repoMgr = &repotesting.Manager{}
	c.artMgr = &arttesting.Manager{}
	c.artrashMgr = &artrashtesting.FakeManager{}
	c.blobMgr = &blob.Manager{}
	c.tagCtl = &tagtesting.FakeController{}
	c.labelMgr = &label.Manager{}
	c.abstractor = &fakeAbstractor{}
	c.immutableMtr = &immutable.FakeMatcher{}
	c.accMgr = &accessorytesting.Manager{}
	c.regCli = &registry.FakeClient{}
	c.ctl = &controller{
		repoMgr:      c.repoMgr,
		artMgr:       c.artMgr,
		artrashMgr:   c.artrashMgr,
		blobMgr:      c.blobMgr,
		tagCtl:       c.tagCtl,
		labelMgr:     c.labelMgr,
		abstractor:   c.abstractor,
		immutableMtr: c.immutableMtr,
		regCli:       c.regCli,
		accessoryMgr: c.accMgr,
	}
}

func (c *controllerTestSuite) TestAssembleArtifact() {
	art := &artifact.Artifact{
		ID:             1,
		Digest:         "sha256:123",
		RepositoryName: "library/hello-world",
	}
	option := &Option{
		WithTag: true,
		TagOption: &tag.Option{
			WithImmutableStatus: false,
		},
		WithLabel:     true,
		WithAccessory: true,
	}
	tg := &tag.Tag{
		Tag: model_tag.Tag{
			ID:           1,
			RepositoryID: 1,
			ArtifactID:   1,
			Name:         "latest",
			PushTime:     time.Now(),
			PullTime:     time.Now(),
		},
	}
	c.tagCtl.On("List").Return([]*tag.Tag{tg}, nil)
	ctx := lib.WithAPIVersion(nil, "2.0")
	lb := &model.Label{
		ID:   1,
		Name: "label",
	}
	c.labelMgr.On("ListByArtifact", mock.Anything, mock.Anything).Return([]*model.Label{
		lb,
	}, nil)
	acc := &basemodel.Default{
		Data: accessorymodel.AccessoryData{
			ID:            1,
			ArtifactID:    2,
			SubArtifactID: 1,
			Type:          accessorymodel.TypeCosignSignature,
		},
	}
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{
		acc,
	}, nil)
	artifact := c.ctl.assembleArtifact(ctx, art, option)
	c.Require().NotNil(artifact)
	c.Equal(art.ID, artifact.ID)
	c.Equal(icon.DigestOfIconDefault, artifact.Icon)
	c.Contains(artifact.Tags, tg)
	c.Contains(artifact.Labels, lb)
	c.Contains(artifact.Accessories, acc)
	// TODO check other fields of option
}

func (c *controllerTestSuite) TestPopulateIcon() {
	cases := []struct {
		art *artifact.Artifact
		ico string
	}{
		{
			art: &artifact.Artifact{
				ID:     1,
				Digest: "sha256:123",
				Type:   image.ArtifactTypeImage,
			},
			ico: icon.DigestOfIconImage,
		},
		{
			art: &artifact.Artifact{
				ID:     2,
				Digest: "sha256:456",
				Type:   cnab.ArtifactTypeCNAB,
			},
			ico: icon.DigestOfIconCNAB,
		},
		{
			art: &artifact.Artifact{
				ID:     3,
				Digest: "sha256:1234",
				Type:   chart.ArtifactTypeChart,
			},
			ico: icon.DigestOfIconChart,
		},
		{
			art: &artifact.Artifact{
				ID:     4,
				Digest: "sha256:1234",
				Type:   "other",
			},
			ico: icon.DigestOfIconDefault,
		},
		{
			art: &artifact.Artifact{
				ID:     5,
				Digest: "sha256:2345",
				Type:   image.ArtifactTypeImage,
				Icon:   "sha256:abcd",
			},
			ico: "sha256:abcd",
		},
	}
	for _, cs := range cases {
		a := &Artifact{
			Artifact: *cs.art,
		}
		c.ctl.populateIcon(a)
		c.Equal(cs.ico, a.Icon)
	}
}

func (c *controllerTestSuite) TestEnsureArtifact() {
	digest := "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180"

	// the artifact already exists
	c.artMgr.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).Return(&artifact.Artifact{
		ID: 1,
	}, nil)
	created, art, err := c.ctl.ensureArtifact(orm.NewContext(nil, &ormtesting.FakeOrmer{}), "library/hello-world", digest)
	c.Require().Nil(err)
	c.False(created)
	c.Equal(int64(1), art.ID)

	// reset the mock
	c.SetupTest()

	// the artifact doesn't exist
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		ProjectID: 1,
	}, nil)
	c.artMgr.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.NotFoundError(nil))
	c.artMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	c.abstractor.On("AbstractMetadata").Return(nil)
	created, art, err = c.ctl.ensureArtifact(orm.NewContext(nil, &ormtesting.FakeOrmer{}), "library/hello-world", digest)
	c.Require().Nil(err)
	c.True(created)
	c.Equal(int64(1), art.ID)
}

func (c *controllerTestSuite) TestEnsure() {
	digest := "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180"

	// both the artifact and the tag don't exist
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		ProjectID: 1,
	}, nil)
	c.artMgr.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.NotFoundError(nil))
	c.artMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	c.abstractor.On("AbstractMetadata").Return(nil)
	c.tagCtl.On("Ensure").Return(nil)
	c.accMgr.On("Ensure").Return(nil)
	_, id, err := c.ctl.Ensure(orm.NewContext(nil, &ormtesting.FakeOrmer{}), "library/hello-world", digest, &ArtOption{
		Tags: []string{"latest"},
	})
	c.Require().Nil(err)
	c.repoMgr.AssertExpectations(c.T())
	c.artMgr.AssertExpectations(c.T())
	c.tagCtl.AssertExpectations(c.T())
	c.abstractor.AssertExpectations(c.T())
	c.Equal(int64(1), id)
}

func (c *controllerTestSuite) TestCount() {
	c.artMgr.On("List", mock.Anything, mock.Anything).Return([]*artifact.Artifact{
		{
			ID:           1,
			RepositoryID: 1,
		},
	}, nil)
	acc := &basemodel.Default{
		Data: accessorymodel.AccessoryData{
			ID:            1,
			ArtifactID:    2,
			SubArtifactID: 1,
			Type:          accessorymodel.TypeCosignSignature,
		},
	}
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{
		acc,
	}, nil)
	total, err := c.ctl.Count(nil, nil)
	c.Require().Nil(err)
	c.Equal(int64(0), total)
}

func (c *controllerTestSuite) TestList() {
	query := &q.Query{}
	option := &Option{
		WithTag:       true,
		WithAccessory: true,
	}
	c.artMgr.On("List", mock.Anything, mock.Anything).Return([]*artifact.Artifact{
		{
			ID:           1,
			RepositoryID: 1,
		},
	}, nil)
	c.tagCtl.On("List").Return([]*tag.Tag{
		{
			Tag: model_tag.Tag{
				ID:           1,
				RepositoryID: 1,
				ArtifactID:   1,
				Name:         "latest",
			},
		},
	}, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		Name: "library/hello-world",
	}, nil)
	c.repoMgr.On("List", mock.Anything, mock.Anything).Return([]*repomodel.RepoRecord{
		{RepositoryID: 1, Name: "library/hello-world"},
	}, nil)
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)
	artifacts, err := c.ctl.List(nil, query, option)
	c.Require().Nil(err)
	c.Require().Len(artifacts, 1)
	c.Equal(int64(1), artifacts[0].ID)
	c.Require().Len(artifacts[0].Tags, 1)
	c.Equal(int64(1), artifacts[0].Tags[0].ID)
	c.Equal(0, len(artifacts[0].Accessories))
}

func (c *controllerTestSuite) TestGet() {
	c.artMgr.On("Get", mock.Anything, mock.Anything, mock.Anything).Return(&artifact.Artifact{
		ID:           1,
		RepositoryID: 1,
	}, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	art, err := c.ctl.Get(nil, 1, nil)
	c.Require().Nil(err)
	c.Require().NotNil(art)
	c.Equal(int64(1), art.ID)
}

func (c *controllerTestSuite) TestGetByDigest() {
	// not found
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		RepositoryID: 1,
	}, nil)
	c.artMgr.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.NotFoundError(nil))
	art, err := c.ctl.getByDigest(nil, "library/hello-world",
		"sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180", nil)
	c.Require().NotNil(err)
	c.True(errors.IsErr(err, errors.NotFoundCode))

	// reset the mock
	c.SetupTest()

	// success
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		RepositoryID: 1,
	}, nil)
	c.artMgr.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).Return(&artifact.Artifact{
		ID:           1,
		RepositoryID: 1,
	}, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	art, err = c.ctl.getByDigest(nil, "library/hello-world",
		"sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180", nil)
	c.Require().Nil(err)
	c.Require().NotNil(art)
	c.Equal(int64(1), art.ID)
}

func (c *controllerTestSuite) TestGetByTag() {
	// not found
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		RepositoryID: 1,
	}, nil)
	c.tagCtl.On("List").Return(nil, nil)
	art, err := c.ctl.getByTag(nil, "library/hello-world", "latest", nil)
	c.Require().NotNil(err)
	c.True(errors.IsErr(err, errors.NotFoundCode))

	// reset the mock
	c.SetupTest()

	// success
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		RepositoryID: 1,
	}, nil)
	c.tagCtl.On("List").Return([]*tag.Tag{
		{
			Tag: model_tag.Tag{
				ID:           1,
				RepositoryID: 1,
				Name:         "latest",
				ArtifactID:   1,
			},
		},
	}, nil)
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{
		ID: 1,
	}, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	art, err = c.ctl.getByTag(nil, "library/hello-world", "latest", nil)
	c.Require().Nil(err)
	c.Require().NotNil(art)
	c.Equal(int64(1), art.ID)
}

func (c *controllerTestSuite) TestGetByReference() {
	// reference is digest
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		RepositoryID: 1,
	}, nil)
	c.artMgr.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).Return(&artifact.Artifact{
		ID:           1,
		RepositoryID: 1,
	}, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	art, err := c.ctl.GetByReference(nil, "library/hello-world",
		"sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180", nil)
	c.Require().Nil(err)
	c.Require().NotNil(art)
	c.Equal(int64(1), art.ID)

	// reset the mock
	c.SetupTest()

	// reference is tag
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		RepositoryID: 1,
	}, nil)
	c.tagCtl.On("List").Return([]*tag.Tag{
		{
			Tag: model_tag.Tag{
				ID:           1,
				RepositoryID: 1,
				Name:         "latest",
				ArtifactID:   1,
			},
		},
	}, nil)
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{
		ID: 1,
	}, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	art, err = c.ctl.GetByReference(nil, "library/hello-world", "latest", nil)
	c.Require().Nil(err)
	c.Require().NotNil(art)
	c.Equal(int64(1), art.ID)
}

func (c *controllerTestSuite) TestDeleteDeeply() {
	// root artifact and doesn't exist
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(nil, errors.NotFoundError(nil))
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)
	err := c.ctl.deleteDeeply(orm.NewContext(nil, &ormtesting.FakeOrmer{}), 1, true, false)
	c.Require().NotNil(err)
	c.Assert().True(errors.IsErr(err, errors.NotFoundCode))

	// reset the mock
	c.SetupTest()

	// child artifact and doesn't exist
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(nil, errors.NotFoundError(nil))
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)
	err = c.ctl.deleteDeeply(orm.NewContext(nil, &ormtesting.FakeOrmer{}), 1, false, false)
	c.Require().Nil(err)

	// reset the mock
	c.SetupTest()

	// child artifact and contains tags
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{ID: 1}, nil)
	c.artMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	c.tagCtl.On("List").Return([]*tag.Tag{
		{
			Tag: model_tag.Tag{
				ID: 1,
			},
		},
	}, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	c.artrashMgr.On("Create").Return(0, nil)
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)
	err = c.ctl.deleteDeeply(orm.NewContext(nil, &ormtesting.FakeOrmer{}), 1, false, false)
	c.Require().Nil(err)

	// reset the mock
	c.SetupTest()

	// root artifact is referenced by other artifacts
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{ID: 1}, nil)
	c.tagCtl.On("List").Return(nil, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	c.artMgr.On("ListReferences", mock.Anything, mock.Anything).Return([]*artifact.Reference{
		{
			ID: 1,
		},
	}, nil)
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)
	err = c.ctl.deleteDeeply(orm.NewContext(nil, &ormtesting.FakeOrmer{}), 1, true, false)
	c.Require().NotNil(err)

	// reset the mock
	c.SetupTest()

	// child artifact contains no tag but referenced by other artifacts
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{ID: 1}, nil)
	c.tagCtl.On("List").Return(nil, nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	c.artMgr.On("ListReferences", mock.Anything, mock.Anything).Return([]*artifact.Reference{
		{
			ID: 1,
		},
	}, nil)
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)
	err = c.ctl.deleteDeeply(nil, 1, false, false)
	c.Require().Nil(err)

	// reset the mock
	c.SetupTest()

	// accessory contains tag
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{ID: 1}, nil)
	c.artMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	c.tagCtl.On("List").Return([]*tag.Tag{
		{
			Tag: model_tag.Tag{
				ID: 1,
			},
		},
	}, nil)
	c.tagCtl.On("DeleteTags", mock.Anything, mock.Anything).Return(nil)
	c.labelMgr.On("RemoveAllFrom", mock.Anything, mock.Anything).Return(nil)
	c.artMgr.On("ListReferences", mock.Anything, mock.Anything).Return([]*artifact.Reference{}, nil)
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)
	c.accMgr.On("DeleteAccessories", mock.Anything, mock.Anything).Return(nil)
	c.blobMgr.On("List", mock.Anything, mock.Anything).Return(nil, nil)
	c.blobMgr.On("CleanupAssociationsForProject", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{}, nil)
	c.artrashMgr.On("Create").Return(0, nil)
	err = c.ctl.deleteDeeply(orm.NewContext(nil, &ormtesting.FakeOrmer{}), 1, true, true)
	c.Require().Nil(err)

}

func (c *controllerTestSuite) TestCopy() {
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{
		ID:     1,
		Digest: "sha256:418fb88ec412e340cdbef913b8ca1bbe8f9e8dc705f9617414c1f2c8db980180",
	}, nil)
	c.repoMgr.On("GetByName", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		RepositoryID: 1,
		Name:         "library/hello-world",
	}, nil)
	c.artMgr.On("Count", mock.Anything, mock.Anything).Return(int64(0), nil)
	c.artMgr.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.NotFoundError(nil))
	c.tagCtl.On("List").Return([]*tag.Tag{
		{
			Tag: model_tag.Tag{
				ID:   1,
				Name: "latest",
			},
		},
	}, nil)
	acc := &basemodel.Default{
		Data: accessorymodel.AccessoryData{
			ID:            1,
			ArtifactID:    2,
			SubArtifactID: 1,
			Type:          accessorymodel.TypeCosignSignature,
		},
	}
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{
		acc,
	}, nil)
	c.tagCtl.On("Update").Return(nil)
	c.repoMgr.On("Get", mock.Anything, mock.Anything).Return(&repomodel.RepoRecord{
		RepositoryID: 1,
		Name:         "library/hello-world",
	}, nil)
	c.abstractor.On("AbstractMetadata").Return(nil)
	c.artMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	c.regCli.On("Copy").Return(nil)
	c.tagCtl.On("Ensure").Return(nil)
	c.accMgr.On("Ensure", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	_, err := c.ctl.Copy(orm.NewContext(nil, &ormtesting.FakeOrmer{}), "library/hello-world", "latest", "library/hello-world2")
	c.Require().Nil(err)
}

func (c *controllerTestSuite) TestUpdatePullTime() {
	// artifact ID and tag ID matches
	c.tagCtl.On("Get").Return(&tag.Tag{
		Tag: model_tag.Tag{
			ID:         1,
			ArtifactID: 1,
		},
	}, nil)
	c.tagCtl.On("Update").Return(nil)
	c.artMgr.On("UpdatePullTime", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := c.ctl.UpdatePullTime(nil, 1, 1, time.Now())
	c.Require().Nil(err)
	c.artMgr.AssertExpectations(c.T())
	c.tagCtl.AssertExpectations(c.T())

	// reset the mock
	c.SetupTest()

	// artifact ID and tag ID doesn't match
	c.tagCtl.On("Get").Return(&tag.Tag{
		Tag: model_tag.Tag{
			ID:         1,
			ArtifactID: 2,
		},
	}, nil)
	c.artMgr.On("UpdatePullTime", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = c.ctl.UpdatePullTime(nil, 1, 1, time.Now())
	c.Require().NotNil(err)
	c.tagCtl.AssertExpectations(c.T())

	// if no tag, should not update tag
	c.SetupTest()
	c.tagCtl.On("Update").Return(nil)
	c.artMgr.On("UpdatePullTime", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err = c.ctl.UpdatePullTime(nil, 1, 0, time.Now())
	c.Require().Nil(err)
	c.artMgr.AssertExpectations(c.T())
	// should not call tag Update
	c.tagCtl.AssertNotCalled(c.T(), "Update")
}

func (c *controllerTestSuite) TestGetAddition() {
	c.artMgr.On("Get", mock.Anything, mock.Anything).Return(&artifact.Artifact{}, nil)
	_, err := c.ctl.GetAddition(nil, 1, "addition")
	c.Require().NotNil(err)
}

func (c *controllerTestSuite) TestAddTo() {
	c.labelMgr.On("AddTo", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := c.ctl.AddLabel(context.Background(), 1, 1)
	c.Require().Nil(err)
}

func (c *controllerTestSuite) TestRemoveFrom() {
	c.labelMgr.On("RemoveFrom", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := c.ctl.RemoveLabel(nil, 1, 1)
	c.Require().Nil(err)
}

func (c *controllerTestSuite) TestWalk() {
	c.artMgr.On("List", mock.Anything, mock.Anything).Return([]*artifact.Artifact{
		{Digest: "d1", ManifestMediaType: v1.MediaTypeImageManifest},
		{Digest: "d2", ManifestMediaType: v1.MediaTypeImageManifest},
	}, nil)
	c.accMgr.On("List", mock.Anything, mock.Anything).Return([]accessorymodel.Accessory{}, nil)

	{
		root := &Artifact{}

		var n int
		c.ctl.Walk(context.TODO(), root, func(a *Artifact) error {
			n++
			return nil
		}, nil)

		c.Equal(1, n)
	}

	{
		root := &Artifact{}
		root.References = []*artifact.Reference{
			{ParentID: 1, ChildID: 2},
			{ParentID: 1, ChildID: 3},
		}

		var n int
		c.ctl.Walk(context.TODO(), root, func(a *Artifact) error {
			n++
			return nil
		}, nil)

		c.Equal(3, n)
	}
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}
