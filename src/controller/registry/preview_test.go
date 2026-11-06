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

package registry

import (
	"context"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/testing/mock"
	testingadapter "github.com/goharbor/harbor/src/testing/pkg/reg/adapter"
)

type resolvingAdapter struct {
	*testingadapter.Adapter
	repository, reference string
	filter                *model.FileFilter
}

func (a *resolvingAdapter) Resolve(repository, reference string, filter *model.FileFilter) (*model.ResolveResult, error) {
	a.repository, a.reference, a.filter = repository, reference, filter
	return &model.ResolveResult{Repository: repository, Reference: reference, Revision: "abc"}, nil
}

func (r *registryTestSuite) TestGetInfoSupportedPreview() {
	info := &model.RegistryInfo{Type: "x"}
	mock.OnAnything(r.regMgr, "Get").Return(&model.Registry{ID: 1, Name: "plain"}, nil).Once()
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(r.adapter, nil).Once()
	mock.OnAnything(r.adapter, "Info").Return(info, nil).Once()
	got, err := r.ctl.GetInfo(context.TODO(), 1)
	r.Require().NoError(err)
	r.False(got.SupportedPreview)

	resolver := &resolvingAdapter{Adapter: &testingadapter.Adapter{}}
	mock.OnAnything(r.regMgr, "Get").Return(&model.Registry{ID: 2, Name: "resolver"}, nil).Once()
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(resolver, nil).Once()
	mock.OnAnything(resolver.Adapter, "Info").Return(&model.RegistryInfo{Type: "y"}, nil).Once()
	got, err = r.ctl.GetInfo(context.TODO(), 2)
	r.Require().NoError(err)
	r.True(got.SupportedPreview)
}

func (r *registryTestSuite) TestPreview() {
	resolver := &resolvingAdapter{Adapter: &testingadapter.Adapter{}}
	mock.OnAnything(r.regMgr, "Get").Return(&model.Registry{ID: 2, Name: "resolver"}, nil).Once()
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(resolver, nil).Once()
	filter := &model.FileFilter{Include: []string{"*.json"}}
	res, err := r.ctl.Preview(context.TODO(), 2, "org/model", "main", filter)
	r.Require().NoError(err)
	r.Equal("abc", res.Revision)
	r.Equal("org/model", resolver.repository)
	r.Equal("main", resolver.reference)
	r.Equal(filter, resolver.filter)
}

func (r *registryTestSuite) TestPreviewUnsupported() {
	mock.OnAnything(r.regMgr, "Get").Return(&model.Registry{ID: 1, Name: "plain"}, nil).Once()
	mock.OnAnything(r.regMgr, "CreateAdapter").Return(r.adapter, nil).Once()
	_, err := r.ctl.Preview(context.TODO(), 1, "library/nginx", "latest", nil)
	r.Require().Error(err)
	r.True(errors.IsErr(err, errors.BadRequestCode))
	r.Contains(err.Error(), "does not support previews")
}

func (r *registryTestSuite) TestPreviewRejectsBadInput() {
	_, err := r.ctl.Preview(context.TODO(), 1, " ", "main", nil)
	r.True(errors.IsErr(err, errors.BadRequestCode))

	_, err = r.ctl.Preview(context.TODO(), 1, "org/model", "main", &model.FileFilter{Include: []string{"[a"}})
	r.True(errors.IsErr(err, errors.BadRequestCode))
	r.regMgr.AssertNotCalled(r.T(), "Get", mock.Anything, mock.Anything)
}
