// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	resourceLabelAPIBasePath = "/api/repositories"
	repo                     = "library/hello-world"
	tag                      = "latest"
	proLibraryLabelID        int64
)

func TestAddToImage(t *testing.T) {
	sysLevelLabelID, err := dao.AddLabel(&models.Label{
		Name:  "sys_level_label",
		Level: common.LabelLevelSystem,
	})
	require.Nil(t, err)
	defer dao.DeleteLabel(sysLevelLabelID)

	proTestLabelID, err := dao.AddLabel(&models.Label{
		Name:      "pro_test_label",
		Level:     common.LabelLevelUser,
		Scope:     common.LabelScopeProject,
		ProjectID: 100,
	})
	require.Nil(t, err)
	defer dao.DeleteLabel(proTestLabelID)

	proLibraryLabelID, err = dao.AddLabel(&models.Label{
		Name:      "pro_library_label",
		Level:     common.LabelLevelUser,
		Scope:     common.LabelScopeProject,
		ProjectID: 1,
	})
	require.Nil(t, err)

	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				url: fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath,
					repo, tag),
				method: http.MethodPost,
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				url: fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath,
					repo, tag),
				method:     http.MethodPost,
				credential: projGuest,
			},
			code: http.StatusForbidden,
		},
		// 404 repo doesn't exist
		{
			request: &testingRequest{
				url:        fmt.Sprintf("%s/library/non-exist-repo/tags/%s/labels", resourceLabelAPIBasePath, tag),
				method:     http.MethodPost,
				credential: projDeveloper,
			},
			code: http.StatusNotFound,
		},
		// 404 image doesn't exist
		{
			request: &testingRequest{
				url:        fmt.Sprintf("%s/%s/tags/non-exist-tag/labels", resourceLabelAPIBasePath, repo),
				method:     http.MethodPost,
				credential: projDeveloper,
			},
			code: http.StatusNotFound,
		},
		// 400
		{
			request: &testingRequest{
				url:        fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath, repo, tag),
				method:     http.MethodPost,
				credential: projDeveloper,
			},
			code: http.StatusBadRequest,
		},
		// 404 label doesn't exist
		{
			request: &testingRequest{
				url: fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath,
					repo, tag),
				method:     http.MethodPost,
				credential: projDeveloper,
				bodyJSON: struct {
					ID int64
				}{
					ID: 1000,
				},
			},
			code: http.StatusNotFound,
		},
		// 400 system level label
		{
			request: &testingRequest{
				url: fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath,
					repo, tag),
				method:     http.MethodPost,
				credential: projDeveloper,
				bodyJSON: struct {
					ID int64
				}{
					ID: sysLevelLabelID,
				},
			},
			code: http.StatusBadRequest,
		},
		// 400 try to add the label of project1 to the image under project2
		{
			request: &testingRequest{
				url: fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath,
					repo, tag),
				method:     http.MethodPost,
				credential: projDeveloper,
				bodyJSON: struct {
					ID int64
				}{
					ID: proTestLabelID,
				},
			},
			code: http.StatusBadRequest,
		},
		// 200
		{
			request: &testingRequest{
				url: fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath,
					repo, tag),
				method:     http.MethodPost,
				credential: projDeveloper,
				bodyJSON: struct {
					ID int64
				}{
					ID: proLibraryLabelID,
				},
			},
			code: http.StatusOK,
		},
	}
	runCodeCheckingCases(t, cases...)
}

func TestGetOfImage(t *testing.T) {
	labels := []*models.Label{}
	err := handleAndParse(&testingRequest{
		url:        fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath, repo, tag),
		method:     http.MethodGet,
		credential: projDeveloper,
	}, &labels)
	require.Nil(t, err)
	require.Equal(t, 1, len(labels))
	assert.Equal(t, proLibraryLabelID, labels[0].ID)
}

func TestRemoveFromImage(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url: fmt.Sprintf("%s/%s/tags/%s/labels/%d", resourceLabelAPIBasePath,
				repo, tag, proLibraryLabelID),
			method:     http.MethodDelete,
			credential: projDeveloper,
		},
		code: http.StatusOK,
	})

	labels := []*models.Label{}
	err := handleAndParse(&testingRequest{
		url: fmt.Sprintf("%s/%s/tags/%s/labels", resourceLabelAPIBasePath,
			repo, tag),
		method:     http.MethodGet,
		credential: projDeveloper,
	}, &labels)
	require.Nil(t, err)
	require.Equal(t, 0, len(labels))
}

func TestAddToRepository(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url:    fmt.Sprintf("%s/%s/labels", resourceLabelAPIBasePath, repo),
			method: http.MethodPost,
			bodyJSON: struct {
				ID int64
			}{
				ID: proLibraryLabelID,
			},
			credential: projDeveloper,
		},
		code: http.StatusOK,
	})
}

func TestGetOfRepository(t *testing.T) {
	labels := []*models.Label{}
	err := handleAndParse(&testingRequest{
		url:        fmt.Sprintf("%s/%s/labels", resourceLabelAPIBasePath, repo),
		method:     http.MethodGet,
		credential: projDeveloper,
	}, &labels)
	require.Nil(t, err)
	require.Equal(t, 1, len(labels))
	assert.Equal(t, proLibraryLabelID, labels[0].ID)
}

func TestRemoveFromRepository(t *testing.T) {
	runCodeCheckingCases(t, &codeCheckingCase{
		request: &testingRequest{
			url: fmt.Sprintf("%s/%s/labels/%d", resourceLabelAPIBasePath,
				repo, proLibraryLabelID),
			method:     http.MethodDelete,
			credential: projDeveloper,
		},
		code: http.StatusOK,
	})

	labels := []*models.Label{}
	err := handleAndParse(&testingRequest{
		url:        fmt.Sprintf("%s/%s/labels", resourceLabelAPIBasePath, repo),
		method:     http.MethodGet,
		credential: projDeveloper,
	}, &labels)
	require.Nil(t, err)
	require.Equal(t, 0, len(labels))

	dao.DeleteLabel(proLibraryLabelID)
}
