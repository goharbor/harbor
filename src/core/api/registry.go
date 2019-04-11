package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/replication/ng/event"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/replication/ng"
	"github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/goharbor/harbor/src/replication/ng/policy"
	"github.com/goharbor/harbor/src/replication/ng/registry"
)

// RegistryAPI handles requests to /api/registries/{}. It manages registries integrated to Harbor.
type RegistryAPI struct {
	BaseController
	manager   registry.Manager
	policyCtl policy.Controller
}

// Prepare validates the user
func (t *RegistryAPI) Prepare() {
	t.BaseController.Prepare()
	if !t.SecurityCtx.IsAuthenticated() {
		t.HandleUnauthorized()
		return
	}

	if !t.SecurityCtx.IsSysAdmin() {
		t.HandleForbidden(t.SecurityCtx.GetUsername())
		return
	}

	t.manager = ng.RegistryMgr
	t.policyCtl = ng.PolicyCtl
}

// Ping checks health status of a registry
func (t *RegistryAPI) Ping() {
	r := &model.Registry{}
	t.DecodeJSONReqAndValidate(r)

	var err error
	id := r.ID
	if id != 0 {
		r, err = t.manager.Get(id)
		if err != nil {
			log.Errorf("failed to get registry %s: %v", r.Name, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		if r == nil {
			t.CustomAbort(http.StatusNotFound, fmt.Sprintf("Registry %d not found", id))
			return
		}
	}

	if len(r.URL) == 0 {
		t.CustomAbort(http.StatusBadRequest, "URL can't be emptry")
		return
	}

	status, err := registry.CheckHealthStatus(r)
	if err != nil {
		t.CustomAbort(http.StatusInternalServerError, fmt.Sprintf("Ping registry %s error: %v", r.URL, err))
		return
	}

	if status != model.Healthy {
		t.CustomAbort(http.StatusBadRequest, fmt.Sprintf("Ping registry %d failed", r.ID))
	}
	return
}

// Get gets a registry by id.
func (t *RegistryAPI) Get() {
	id := t.GetIDFromURL()

	registry, err := t.manager.Get(id)
	if err != nil {
		log.Errorf("failed to get registry %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if registry == nil {
		t.HandleNotFound(fmt.Sprintf("registry %d not found", id))
		return
	}

	// Hide access secret
	if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
		registry.Credential.AccessSecret = "*****"
	}

	t.Data["json"] = registry
	t.ServeJSON()
}

// List lists all registries that match a given registry name.
func (t *RegistryAPI) List() {
	name := t.GetString("name")

	_, registries, err := t.manager.List(&model.RegistryQuery{
		Name: name,
	})
	if err != nil {
		log.Errorf("failed to list registries %s: %v", name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	// Hide passwords
	for _, registry := range registries {
		if registry.Credential != nil && len(registry.Credential.AccessSecret) != 0 {
			registry.Credential.AccessSecret = "*****"
		}
	}

	t.Data["json"] = registries
	t.ServeJSON()
	return
}

// Post creates a registry
func (t *RegistryAPI) Post() {
	registry := &model.Registry{}
	t.DecodeJSONReqAndValidate(registry)

	reg, err := t.manager.GetByName(registry.Name)
	if err != nil {
		log.Errorf("failed to get registry %s: %v", registry.Name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if reg != nil {
		t.HandleConflict(fmt.Sprintf("name '%s' is already used", registry.Name))
		return
	}

	id, err := t.manager.Add(registry)
	if err != nil {
		log.Errorf("Add registry '%s' error: %v", registry.URL, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	t.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put updates a registry
func (t *RegistryAPI) Put() {
	id := t.GetIDFromURL()

	registry, err := t.manager.Get(id)
	if err != nil {
		log.Errorf("Get registry by id %d error: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if registry == nil {
		t.HandleNotFound(fmt.Sprintf("Registry %d not found", id))
		return
	}

	req := models.RegistryUpdateRequest{}
	t.DecodeJSONReq(&req)

	originalName := registry.Name

	if req.Name != nil {
		registry.Name = *req.Name
	}
	if req.Description != nil {
		registry.Description = *req.Description
	}
	if req.URL != nil {
		registry.URL = *req.URL
	}
	if req.CredentialType != nil {
		registry.Credential.Type = (model.CredentialType)(*req.CredentialType)
	}
	if req.AccessKey != nil {
		registry.Credential.AccessKey = *req.AccessKey
	}
	if req.AccessSecret != nil {
		registry.Credential.AccessSecret = *req.AccessSecret
	}
	if req.Insecure != nil {
		registry.Insecure = *req.Insecure
	}

	t.Validate(registry)

	if registry.Name != originalName {
		reg, err := t.manager.GetByName(registry.Name)
		if err != nil {
			log.Errorf("Get registry by name '%s' error: %v", registry.Name, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		if reg != nil {
			t.HandleConflict("name is already used")
			return
		}
	}

	if err := t.manager.Update(registry); err != nil {
		log.Errorf("Update registry %d error: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
}

// Delete deletes a registry
func (t *RegistryAPI) Delete() {
	id := t.GetIDFromURL()

	registry, err := t.manager.Get(id)
	if err != nil {
		msg := fmt.Sprintf("Get registry %d error: %v", id, err)
		log.Error(msg)
		t.HandleInternalServerError(msg)
		return
	}

	if registry == nil {
		t.HandleNotFound(fmt.Sprintf("Registry %d not found", id))
		return
	}

	// Check whether there are replication policies that use this registry as source registry.
	total, _, err := t.policyCtl.List([]*model.PolicyQuery{
		{
			SrcRegistry: id,
		},
	}...)
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("List replication policies with source registry %d error: %v", id, err))
		return
	}
	if total > 0 {
		msg := fmt.Sprintf("Can't delete registry %d,  %d replication policies use it as source registry", id, total)
		log.Error(msg)
		t.HandleStatusPreconditionFailed(msg)
		return
	}

	// Check whether there are replication policies that use this registry as destination registry.
	total, _, err = t.policyCtl.List([]*model.PolicyQuery{
		{
			DestRegistry: id,
		},
	}...)
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("List replication policies with destination registry %d error: %v", id, err))
		return
	}
	if total > 0 {
		msg := fmt.Sprintf("Can't delete registry %d,  %d replication policies use it as destination registry", id, total)
		log.Error(msg)
		t.HandleStatusPreconditionFailed(msg)
		return
	}

	if err := t.manager.Remove(id); err != nil {
		msg := fmt.Sprintf("Delete registry %d error: %v", id, err)
		log.Error(msg)
		t.HandleInternalServerError(msg)
		return
	}
}

// GetInfo returns the base info and capability declarations of the registry
func (t *RegistryAPI) GetInfo() {
	id, err := t.GetInt64FromPath(":id")
	// "0" is used for the ID of the local Harbor registry
	if err != nil || id < 0 {
		t.HandleBadRequest(fmt.Sprintf("invalid registry ID %s", t.GetString(":id")))
		return
	}
	var registry *model.Registry
	if id == 0 {
		registry = event.GetLocalRegistry()
	} else {
		registry, err = t.manager.Get(id)
		if err != nil {
			t.HandleInternalServerError(fmt.Sprintf("failed to get registry %d: %v", id, err))
			return
		}
		if registry == nil {
			t.HandleNotFound(fmt.Sprintf("registry %d not found", id))
			return
		}
	}

	factory, err := adapter.GetFactory(registry.Type)
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("failed to get the adapter factory for registry type %s: %v", registry.Type, err))
		return
	}
	adp, err := factory(registry)
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("failed to create the adapter for registry %d: %v", registry.ID, err))
		return
	}
	info, err := adp.Info()
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("failed to get registry info %d: %v", id, err))
		return
	}
	t.WriteJSONData(process(info))
}

// GetNamespace get the namespace of a registry
func (t *RegistryAPI) GetNamespace() {
	var registry *model.Registry
	var err error

	id, err := t.GetInt64FromPath(":id")
	if err != nil || id < 0 {
		t.HandleBadRequest(fmt.Sprintf("invalid registry ID %s", t.GetString(":id")))
		return
	}
	if id > 0 {
		registry, err = t.manager.Get(id)
		if err != nil {
			t.HandleInternalServerError(fmt.Sprintf("failed to get registry %d: %v", id, err))
			return
		}
	} else if id == 0 {
		registry = event.GetLocalRegistry()
	}

	if registry == nil {
		t.HandleNotFound(fmt.Sprintf("registry %d not found", id))
		return
	}

	if !adapter.HasFactory(registry.Type) {
		t.HandleInternalServerError(fmt.Sprintf("no adapter factory found for %s", registry.Type))
		return
	}

	regFactory, err := adapter.GetFactory(registry.Type)
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("fail to get adapter factory %s", registry.Type))
		return
	}
	regAdapter, err := regFactory(registry)
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("fail to get adapter %s", registry.Type))
		return
	}

	query := &model.NamespaceQuery{
		Name: t.GetString("name"),
	}
	npResults, err := regAdapter.ListNamespaces(query)
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("fail to list namespaces %s %v", registry.Type, err))
		return
	}

	t.Data["json"] = npResults
	t.ServeJSON()
}

// merge "SupportedResourceTypes" into "SupportedResourceFilters" for UI to render easier
func process(info *model.RegistryInfo) *model.RegistryInfo {
	if info == nil {
		return nil
	}
	in := &model.RegistryInfo{
		Type:              info.Type,
		Description:       info.Description,
		SupportedTriggers: info.SupportedTriggers,
	}
	filters := []*model.FilterStyle{}
	for _, filter := range info.SupportedResourceFilters {
		if filter.Type != model.FilterTypeResource {
			filters = append(filters, filter)
		}
	}
	values := []string{}
	for _, resourceType := range info.SupportedResourceTypes {
		values = append(values, string(resourceType))
	}
	filters = append(filters, &model.FilterStyle{
		Type:   model.FilterTypeResource,
		Style:  model.FilterStyleTypeRadio,
		Values: values,
	})
	in.SupportedResourceFilters = filters

	return in
}
