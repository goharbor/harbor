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

	repmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/testing/mock"
)

const (
	typeModelSource = "test-model-source"
	typePlainSource = "test-plain-source"
)

type infoAdapter struct {
	adapter.Adapter
	info *model.RegistryInfo
}

func (i *infoAdapter) Info() (*model.RegistryInfo, error) { return i.info, nil }

type infoFactory struct{ info *model.RegistryInfo }

func (f infoFactory) Create(*model.Registry) (adapter.Adapter, error) {
	return &infoAdapter{info: f.info}, nil
}
func (f infoFactory) AdapterPattern() *model.AdapterPattern { return nil }

func init() {
	_ = adapter.RegisterFactory(typeModelSource, infoFactory{info: &model.RegistryInfo{
		SupportedResourceFilters: []*model.FilterStyle{{Type: model.FilterTypeName}, {Type: model.FilterTypeFile}},
	}})
	_ = adapter.RegisterFactory(typePlainSource, infoFactory{info: &model.RegistryInfo{
		SupportedResourceFilters: []*model.FilterStyle{{Type: model.FilterTypeName}},
	}})
}

func (r *replicationTestSuite) createPullPolicy(srcType string, filters ...*model.Filter) error {
	r.regMgr.On("Get", mock.Anything, int64(0)).Return(&model.Registry{ID: 0, Type: model.RegistryTypeHarbor}, nil)
	r.regMgr.On("Get", mock.Anything, int64(5)).Return(&model.Registry{ID: 5, Name: "src", Type: srcType}, nil)
	mock.OnAnything(r.repMgr, "Create").Return(int64(1), nil)
	_, err := r.ctl.CreatePolicy(context.TODO(), &repmodel.Policy{
		Name:         "pull",
		SrcRegistry:  &model.Registry{ID: 5},
		DestRegistry: &model.Registry{ID: 0},
		Filters:      filters,
		Enabled:      true,
	})
	return err
}

func (r *replicationTestSuite) TestCreatePolicyFileFilterSupported() {
	r.Require().NoError(r.createPullPolicy(typeModelSource, &model.Filter{Type: model.FilterTypeFile, Value: "*.json"}))
}

func (r *replicationTestSuite) TestCreatePolicyFileFilterUnsupported() {
	err := r.createPullPolicy(typePlainSource, &model.Filter{Type: model.FilterTypeFile, Value: "*.json"})
	r.Require().Error(err)
	r.True(errors.IsErr(err, errors.BadRequestCode))
	r.Contains(err.Error(), "does not support the file filter")
	r.repMgr.AssertNotCalled(r.T(), "Create", mock.Anything, mock.Anything)
}

func (r *replicationTestSuite) TestCreatePolicyEmptyFileFilterNeedsNoSupport() {
	r.Require().NoError(r.createPullPolicy(typePlainSource, &model.Filter{Type: model.FilterTypeFile, Value: ""}))
}

func (r *replicationTestSuite) TestCreatePushPolicyRejectsSourceOptions() {
	r.regMgr.On("Get", mock.Anything, mock.Anything).Return(&model.Registry{ID: 3, Type: typeModelSource}, nil)
	_, err := r.ctl.CreatePolicy(context.TODO(), &repmodel.Policy{
		Name:         "push",
		DestRegistry: &model.Registry{ID: 3},
		Filters:      []*model.Filter{{Type: model.FilterTypeFile, Value: "*.json"}},
	})
	r.Require().Error(err)
	r.True(errors.IsErr(err, errors.BadRequestCode))
}
