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

package scan

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/testing/mock"
	mocks "github.com/goharbor/harbor/src/testing/pkg/scan/scanner"
	"github.com/stretchr/testify/assert"
)

func TestEnsureScanners(t *testing.T) {

	t.Run("Should do nothing when list of wanted scanners is empty", func(t *testing.T) {
		err := EnsureScanners(context.TODO(), []scanner.Registration{})
		assert.NoError(t, err)
	})

	t.Run("Should return error when listing scanners fails", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{
				"name__in": []string{"scanner"},
			},
		}).Return(nil, errors.New("DB error"))

		err := EnsureScanners(context.TODO(), []scanner.Registration{
			{Name: "scanner", URL: "http://scanner:8080"},
		})

		assert.EqualError(t, err, "listing scanners: DB error")
		mgr.AssertExpectations(t)
	})

	t.Run("Should create only non-existing scanners", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{
				"name__in": []string{
					"trivy",
				},
			},
		}).Return([]*scanner.Registration{}, nil)
		mgr.On("Create", mock.Anything, &scanner.Registration{
			Name: "trivy",
			URL:  "http://trivy:8080",
		}).Return("uuid-trivy", nil)

		err := EnsureScanners(context.TODO(), []scanner.Registration{
			{Name: "trivy", URL: "http://trivy:8080"},
		})

		assert.NoError(t, err)
		mgr.AssertExpectations(t)
	})

	t.Run("Should update scanners", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{
				"name__in": []string{
					"trivy",
				},
			},
		}).Return([]*scanner.Registration{
			{Name: "trivy", URL: "http://trivy:8080"},
		}, nil)
		mgr.On("Update", mock.Anything, &scanner.Registration{
			Name: "trivy",
			URL:  "http://trivy:8443",
		}).Return(nil)

		err := EnsureScanners(context.TODO(), []scanner.Registration{
			{Name: "trivy", URL: "http://trivy:8443"},
		})

		assert.NoError(t, err)
		mgr.AssertExpectations(t)
	})

}

func TestEnsureDefaultScanner(t *testing.T) {

	t.Run("Should return error when getting default scanner fails", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("GetDefault", mock.Anything).Return(nil, errors.New("DB error"))

		err := EnsureDefaultScanner(context.TODO(), "trivy")
		assert.EqualError(t, err, "getting default scanner: DB error")
		mgr.AssertExpectations(t)
	})

	t.Run("Should do nothing when the default scanner is already set", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("GetDefault", mock.Anything).Return(&scanner.Registration{
			Name: "trivy",
		}, nil)

		err := EnsureDefaultScanner(context.TODO(), "trivy")
		assert.NoError(t, err)
		mgr.AssertExpectations(t)
	})

	t.Run("Should return error when listing scanners fails", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("GetDefault", mock.Anything).Return(nil, nil)
		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{"name": "trivy"},
		}).Return(nil, errors.New("DB error"))

		err := EnsureDefaultScanner(context.TODO(), "trivy")
		assert.EqualError(t, err, "listing scanners: DB error")
		mgr.AssertExpectations(t)
	})

	t.Run("Should return error when listing scanners returns unexpected scanners count", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("GetDefault", mock.Anything).Return(nil, nil)
		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{"name": "trivy"},
		}).Return([]*scanner.Registration{
			{Name: "trivy"},
			{Name: "trivy"},
		}, nil)

		err := EnsureDefaultScanner(context.TODO(), "trivy")
		assert.EqualError(t, err, "expected only one scanner with name trivy but got 2")
		mgr.AssertExpectations(t)
	})

	t.Run("Should set the default scanner when it is not set", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("GetDefault", mock.Anything).Return(nil, nil)
		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{"name": "trivy"},
		}).Return([]*scanner.Registration{
			{
				Name: "trivy",
				UUID: "trivy-uuid",
				URL:  "http://trivy:8080",
			},
		}, nil)
		mgr.On("SetAsDefault", mock.Anything, "trivy-uuid").Return(nil)

		err := EnsureDefaultScanner(context.TODO(), "trivy")
		assert.NoError(t, err)
		mgr.AssertExpectations(t)
	})

	t.Run("Should return error when setting the default scanner fails", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("GetDefault", mock.Anything).Return(nil, nil)
		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{"name": "trivy"},
		}).Return([]*scanner.Registration{
			{
				Name: "trivy",
				UUID: "trivy-uuid",
				URL:  "http://trivy:8080",
			},
		}, nil)
		mgr.On("SetAsDefault", mock.Anything, "trivy-uuid").Return(errors.New("DB error"))

		err := EnsureDefaultScanner(context.TODO(), "trivy")
		assert.EqualError(t, err, "setting trivy as default scanner: DB error")
		mgr.AssertExpectations(t)
	})

}

func TestRemoveImmutableScanners(t *testing.T) {

	t.Run("Should do nothing when list of names is empty", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		err := RemoveImmutableScanners(context.TODO(), []string{})
		assert.NoError(t, err)
		mgr.AssertExpectations(t)
	})

	t.Run("Should return error when listing scanners fails", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{
				"immutable": true,
				"name__in":  []string{"scanner"},
			},
		}).Return(nil, errors.New("DB error"))

		err := RemoveImmutableScanners(context.TODO(), []string{"scanner"})
		assert.EqualError(t, err, "listing scanners: DB error")
		mgr.AssertExpectations(t)
	})

	t.Run("Should delete multiple scanners", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		registrations := []*scanner.Registration{
			{
				Name: "scanner-1",
				UUID: "uuid-1",
				URL:  "http://scanner-1",
			},
			{
				Name: "scanner-2",
				UUID: "uuid-2",
				URL:  "http://scanner-2",
			}}

		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{
				"immutable": true,
				"name__in": []string{
					"scanner-1",
					"scanner-2",
				},
			},
		}).Return(registrations, nil)
		mgr.On("Delete", mock.Anything, "uuid-1").Return(nil)
		mgr.On("Delete", mock.Anything, "uuid-2").Return(nil)

		err := RemoveImmutableScanners(context.TODO(), []string{
			"scanner-1",
			"scanner-2",
		})
		assert.NoError(t, err)
		mgr.AssertExpectations(t)
	})

	t.Run("Should return error when deleting any scanner fails", func(t *testing.T) {
		mgr := &mocks.Manager{}
		scannerManager = mgr

		registrations := []*scanner.Registration{
			{
				Name: "scanner-1",
				UUID: "uuid-1",
				URL:  "http://scanner-1",
			},
			{
				Name: "scanner-2",
				UUID: "uuid-2",
				URL:  "http://scanner-2",
			}}

		mgr.On("List", mock.Anything, &q.Query{
			Keywords: map[string]interface{}{
				"immutable": true,
				"name__in": []string{
					"scanner-1",
					"scanner-2",
				},
			},
		}).Return(registrations, nil)
		mgr.On("Delete", mock.Anything, "uuid-1").Return(nil)
		mgr.On("Delete", mock.Anything, "uuid-2").Return(errors.New("DB error"))

		err := RemoveImmutableScanners(context.TODO(), []string{
			"scanner-1",
			"scanner-2",
		})
		assert.EqualError(t, err, "deleting scanner: uuid-2: DB error")
		mgr.AssertExpectations(t)
	})

}
