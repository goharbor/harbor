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

package api

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/pkg/q"
	s "github.com/goharbor/harbor/src/pkg/scan/api/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/pkg/errors"
)

// ScannerAPI provides the API for managing the plugin scanners
type ScannerAPI struct {
	// The base controller to provide common utilities
	BaseController

	// Controller for the plug scanners
	c s.Controller
}

// Prepare sth. for the subsequent actions
func (sa *ScannerAPI) Prepare() {
	// Call super prepare method
	sa.BaseController.Prepare()

	// Check access permissions
	if !sa.SecurityCtx.IsAuthenticated() {
		sa.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	if !sa.SecurityCtx.IsSysAdmin() {
		sa.SendForbiddenError(errors.New(sa.SecurityCtx.GetUsername()))
		return
	}

	// Use the default controller
	sa.c = s.DefaultController
}

// Get the specified scanner
func (sa *ScannerAPI) Get() {
	if r := sa.get(); r != nil {
		// Response to the client
		sa.Data["json"] = r
		sa.ServeJSON()
	}
}

// Metadata returns the metadata of the given scanner.
func (sa *ScannerAPI) Metadata() {
	uuid := sa.GetStringFromPath(":uuid")

	meta, err := sa.c.GetMetadata(uuid)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: get metadata"))
		return
	}

	// Response to the client
	sa.Data["json"] = meta
	sa.ServeJSON()
}

// List all the scanners
func (sa *ScannerAPI) List() {
	p, pz, err := sa.GetPaginationParams()
	if err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: list all"))
		return
	}

	query := &q.Query{
		PageSize:   pz,
		PageNumber: p,
	}

	// Get query key words
	kws := make(map[string]interface{})
	properties := []string{"name", "description", "url", "ex_name", "ex_url"}
	for _, k := range properties {
		kw := sa.GetString(k)
		if len(kw) > 0 {
			kws[k] = kw
		}
	}

	if len(kws) > 0 {
		query.Keywords = kws
	}

	all, err := sa.c.ListRegistrations(query)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: list all"))
		return
	}

	// Response to the client
	sa.Data["json"] = all
	sa.ServeJSON()
}

// Create a new scanner
func (sa *ScannerAPI) Create() {
	r := &scanner.Registration{}

	if err := sa.DecodeJSONReq(r); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: create"))
		return
	}

	if err := r.Validate(false); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: create"))
		return
	}

	// Explicitly check if conflict
	if !sa.checkDuplicated("name", r.Name) ||
		!sa.checkDuplicated("url", r.URL) {
		return
	}

	// All newly created should be non default one except the 1st one
	r.IsDefault = false

	uuid, err := sa.c.CreateRegistration(r)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: create"))
		return
	}

	location := fmt.Sprintf("%s/%s", sa.Ctx.Request.RequestURI, uuid)
	sa.Ctx.ResponseWriter.Header().Add("Location", location)

	resp := make(map[string]string, 1)
	resp["uuid"] = uuid

	// Response to the client
	sa.Ctx.ResponseWriter.WriteHeader(http.StatusCreated)
	sa.Data["json"] = resp
	sa.ServeJSON()
}

// Update a scanner
func (sa *ScannerAPI) Update() {
	r := sa.get()
	if r == nil {
		// meet error
		return
	}

	// Immutable registration is not allowed
	if r.Immutable {
		sa.SendForbiddenError(errors.Errorf("registration %s is not allowed to update as it is immutable: scanner API: update", r.Name))
		return
	}

	// full dose updated
	rr := &scanner.Registration{}
	if err := sa.DecodeJSONReq(rr); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: update"))
		return
	}

	if err := r.Validate(true); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: update"))
		return
	}

	// Name changed?
	if r.Name != rr.Name {
		if !sa.checkDuplicated("name", rr.Name) {
			return
		}
	}

	// URL changed?
	if r.URL != rr.URL {
		if !sa.checkDuplicated("url", rr.URL) {
			return
		}
	}

	getChanges(r, rr)

	if err := sa.c.UpdateRegistration(r); err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: update"))
		return
	}

	location := fmt.Sprintf("%s/%s", sa.Ctx.Request.RequestURI, r.UUID)
	sa.Ctx.ResponseWriter.Header().Add("Location", location)

	// Response to the client
	sa.Data["json"] = r
	sa.ServeJSON()
}

// Delete the scanner
func (sa *ScannerAPI) Delete() {
	r := sa.get()
	if r == nil {
		// meet error
		return
	}

	// Immutable registration is not allowed
	if r.Immutable {
		sa.SendForbiddenError(errors.Errorf("registration %s is not allowed to delete as it is immutable: scanner API: update", r.Name))
		return
	}

	deleted, err := sa.c.DeleteRegistration(r.UUID)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: delete"))
		return
	}

	sa.Data["json"] = deleted
	sa.ServeJSON()
}

// SetAsDefault sets the given registration as default one
func (sa *ScannerAPI) SetAsDefault() {
	uid := sa.GetStringFromPath(":uuid")

	m := make(map[string]interface{})
	if err := sa.DecodeJSONReq(&m); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: set as default"))
		return
	}

	if v, ok := m["is_default"]; ok {
		if isDefault, y := v.(bool); y && isDefault {
			if err := sa.c.SetDefaultRegistration(uid); err != nil {
				sa.SendInternalServerError(errors.Wrap(err, "scanner API: set as default"))
			}

			return
		}
	}

	// Not supported
	sa.SendForbiddenError(errors.Errorf("not supported: %#v", m))
}

// Ping the registration.
func (sa *ScannerAPI) Ping() {
	r := &scanner.Registration{}

	if err := sa.DecodeJSONReq(r); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: ping"))
		return
	}

	if err := r.Validate(false); err != nil {
		sa.SendBadRequestError(errors.Wrap(err, "scanner API: ping"))
		return
	}

	if _, err := sa.c.Ping(r); err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: ping"))
		return
	}
}

// get the specified scanner
func (sa *ScannerAPI) get() *scanner.Registration {
	uid := sa.GetStringFromPath(":uuid")

	r, err := sa.c.GetRegistration(uid)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: get"))
		return nil
	}

	if r == nil {
		// NOT found
		sa.SendNotFoundError(errors.Errorf("scanner: %s", uid))
		return nil
	}

	return r
}

func (sa *ScannerAPI) checkDuplicated(property, value string) bool {
	// Explicitly check if conflict
	kw := make(map[string]interface{})
	kw[property] = value

	query := &q.Query{
		Keywords: kw,
	}

	l, err := sa.c.ListRegistrations(query)
	if err != nil {
		sa.SendInternalServerError(errors.Wrap(err, "scanner API: check existence"))
		return false
	}

	if len(l) > 0 {
		sa.SendConflictError(errors.Errorf("duplicated entries: %s:%s", property, value))
		return false
	}

	return true
}

func getChanges(e *scanner.Registration, eChange *scanner.Registration) {
	e.Name = eChange.Name
	e.Description = eChange.Description
	e.URL = eChange.URL
	e.Auth = eChange.Auth
	e.AccessCredential = eChange.AccessCredential
	e.Disabled = eChange.Disabled
	e.SkipCertVerify = eChange.SkipCertVerify
}
