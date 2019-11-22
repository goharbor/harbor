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
	"testing"

	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	sc "github.com/goharbor/harbor/src/pkg/scan/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/scanner/mocks"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type managerOptions struct {
	registrations   []*scanner.Registration
	listError       error
	getError        error
	getDefaultError error
	createError     error
	createErrorFn   func(*scanner.Registration) error
}

func newManager(opts *managerOptions) sc.Manager {
	if opts == nil {
		opts = &managerOptions{}
	}

	data := map[string]*scanner.Registration{}
	for _, reg := range opts.registrations {
		data[reg.URL] = reg
	}

	mgr := &mocks.Manager{}

	listFn := func(query *q.Query) []*scanner.Registration {
		if opts.listError != nil {
			return nil
		}

		url := query.Keywords["url"]

		var results []*scanner.Registration
		for key, reg := range data {
			if url == key {
				results = append(results, reg)
			}
		}

		return results
	}

	getFn := func(url string) *scanner.Registration {
		if opts.getError != nil {
			return nil
		}

		return data[url]
	}

	getDefaultFn := func() *scanner.Registration {
		if opts.getDefaultError != nil {
			return nil
		}

		for _, reg := range data {
			if reg.IsDefault {
				return reg
			}
		}

		return nil
	}

	createFn := func(reg *scanner.Registration) string {
		if opts.createError != nil {
			return ""
		}

		data[reg.URL] = reg

		return reg.URL
	}

	createError := func(reg *scanner.Registration) error {
		if opts.createErrorFn != nil {
			return opts.createErrorFn(reg)
		}

		return opts.createError
	}

	mgr.On("List", mock.AnythingOfType("*q.Query")).Return(listFn, opts.listError)
	mgr.On("Get", mock.AnythingOfType("string")).Return(getFn, opts.getError)
	mgr.On("GetDefault").Return(getDefaultFn, opts.getDefaultError)
	mgr.On("Create", mock.AnythingOfType("*scanner.Registration")).Return(createFn, createError)

	return mgr
}

func TestEnsureScanner(t *testing.T) {
	assert := assert.New(t)

	registrations := []*scanner.Registration{
		{URL: "reg1"},
	}

	// registration with the url exist in the system
	scannerManager = newManager(
		&managerOptions{
			registrations: registrations,
		},
	)
	assert.Nil(EnsureScanner(&scanner.Registration{URL: "reg1"}))

	// list registrations got error
	scannerManager = newManager(
		&managerOptions{
			listError: errors.New("list registrations internal error"),
		},
	)
	assert.Error(EnsureScanner(&scanner.Registration{URL: "reg1"}))

	// create registration got error
	scannerManager = newManager(
		&managerOptions{
			createError: errors.New("create registration internal error"),
		},
	)
	assert.Error(EnsureScanner(&scanner.Registration{URL: "reg1"}))

	// get default registration got error
	scannerManager = newManager(
		&managerOptions{
			getDefaultError: errors.New("get default registration internal error"),
		},
	)
	assert.Error(EnsureScanner(&scanner.Registration{URL: "reg1"}))

	// create registration when no registrations in the system
	scannerManager = newManager(nil)
	assert.Nil(EnsureScanner(&scanner.Registration{URL: "reg1"}))
	reg1, err := scannerManager.Get("reg1")
	assert.Nil(err)
	assert.NotNil(reg1)
	assert.True(reg1.IsDefault)

	// create registration when there are registrations in the system
	scannerManager = newManager(
		&managerOptions{
			registrations: registrations,
		},
	)
	assert.Nil(EnsureScanner(&scanner.Registration{URL: "reg2"}))
	reg2, err := scannerManager.Get("reg2")
	assert.Nil(err)
	assert.NotNil(reg2)
	assert.True(reg2.IsDefault)

	// create registration when there are registrations in the system and the default registration exist
	scannerManager = newManager(
		&managerOptions{
			registrations: []*scanner.Registration{
				{URL: "reg1", IsDefault: true},
			},
		},
	)
	assert.Nil(EnsureScanner(&scanner.Registration{URL: "reg3"}))
	reg3, err := scannerManager.Get("reg3")
	assert.Nil(err)
	assert.NotNil(reg3)
	assert.False(reg3.IsDefault)
}

func TestEnsureScannerWithResolveConflict(t *testing.T) {
	assert := assert.New(t)

	registrations := []*scanner.Registration{
		{URL: "reg1"},
	}

	// create registration got ErrDupRows when its name is Clair
	scannerManager = newManager(
		&managerOptions{
			registrations: registrations,

			createErrorFn: func(reg *scanner.Registration) error {
				if reg.Name == "Clair" {
					return errors.Wrap(types.ErrDupRows, "failed to create reg")
				}

				return nil
			},
		},
	)

	assert.Nil(EnsureScanner(&scanner.Registration{Name: "Clair", URL: "reg1"}))
	assert.Error(EnsureScanner(&scanner.Registration{Name: "Clair", URL: "reg2"}))
	assert.Nil(EnsureScanner(&scanner.Registration{Name: "Clair", URL: "reg2"}, true))
}
