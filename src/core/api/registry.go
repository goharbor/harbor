package api

import (
	"fmt"
	"net/http"
	"strconv"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/policy"
	"github.com/goharbor/harbor/src/replication/registry"
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

	t.manager = replication.RegistryMgr
	t.policyCtl = replication.PolicyCtl
}

// Ping checks health status of a registry
func (t *RegistryAPI) Ping() {
	req := struct {
		ID             *int64  `json:"id"`
		Type           *string `json:"type"`
		URL            *string `json:"url"`
		CredentialType *string `json:"credential_type"`
		AccessKey      *string `json:"access_key"`
		AccessSecret   *string `json:"access_secret"`
		Insecure       *bool   `json:"insecure"`
	}{}
	t.DecodeJSONReq(&req)

	reg := &model.Registry{}
	var err error
	if req.ID != nil {
		reg, err = t.manager.Get(*req.ID)
		if err != nil {
			t.HandleInternalServerError(fmt.Sprintf("failed to get registry %d: %v", *req.ID, err))
			return
		}

		if reg == nil {
			t.HandleNotFound(fmt.Sprintf("registry %d not found", *req.ID))
			return
		}
	}
	if req.Type != nil {
		reg.Type = model.RegistryType(*req.Type)
	}
	if req.URL != nil {
		url, err := utils.ParseEndpoint(*req.URL)
		if err != nil {
			t.HandleBadRequest(err.Error())
			return
		}

		// Prevent SSRF security issue #3755
		reg.URL = url.Scheme + "://" + url.Host + url.Path
	}
	if req.CredentialType != nil {
		if reg.Credential == nil {
			reg.Credential = &model.Credential{}
		}
		reg.Credential.Type = model.CredentialType(*req.CredentialType)
	}
	if req.AccessKey != nil {
		if reg.Credential == nil {
			reg.Credential = &model.Credential{}
		}
		reg.Credential.AccessKey = *req.AccessKey
	}
	if req.AccessSecret != nil {
		if reg.Credential == nil {
			reg.Credential = &model.Credential{}
		}
		reg.Credential.AccessSecret = *req.AccessSecret
	}
	if req.Insecure != nil {
		reg.Insecure = *req.Insecure
	}
	if len(reg.Type) == 0 || len(reg.URL) == 0 {
		t.HandleBadRequest("type or url cannot be empty")
		return
	}

	status, err := registry.CheckHealthStatus(reg)
	if err != nil {
		e, ok := err.(*common_http.Error)
		if ok && e.Code == http.StatusUnauthorized {
			t.HandleBadRequest("invalid credential")
			return
		}
		t.HandleInternalServerError(fmt.Sprintf("failed to check health of registry %s: %v", reg.URL, err))
		return
	}

	if status != model.Healthy {
		t.HandleBadRequest("")
		return
	}
	return
}

// Get gets a registry by id.
func (t *RegistryAPI) Get() {
	id := t.GetIDFromURL()

	r, err := t.manager.Get(id)
	if err != nil {
		log.Errorf("failed to get registry %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if r == nil {
		t.HandleNotFound(fmt.Sprintf("registry %d not found", id))
		return
	}

	// Hide access secret
	if r.Credential != nil && len(r.Credential.AccessSecret) != 0 {
		r.Credential.AccessSecret = "*****"
	}

	t.Data["json"] = r
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
	for _, r := range registries {
		if r.Credential != nil && len(r.Credential.AccessSecret) != 0 {
			r.Credential.AccessSecret = "*****"
		}
	}

	t.Data["json"] = registries
	t.ServeJSON()
	return
}

// Post creates a registry
func (t *RegistryAPI) Post() {
	r := &model.Registry{}
	t.DecodeJSONReqAndValidate(r)

	reg, err := t.manager.GetByName(r.Name)
	if err != nil {
		log.Errorf("failed to get registry %s: %v", r.Name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}
	if reg != nil {
		t.HandleConflict(fmt.Sprintf("name '%s' is already used", r.Name))
		return
	}

	if r.Type == model.RegistryTypeLocalHarbor {
		n, _, err := t.manager.List(&model.RegistryQuery{
			Type: string(model.RegistryTypeLocalHarbor),
		})
		if err != nil {
			t.HandleInternalServerError(fmt.Sprintf("failed to list registries: %v", err))
			return
		}
		if n > 0 {
			t.HandleBadRequest(fmt.Sprintf("can only add one registry whose type is %s", model.RegistryTypeLocalHarbor))
			return
		}
	}

	status, err := registry.CheckHealthStatus(r)
	if err != nil {
		t.HandleBadRequest(fmt.Sprintf("health check to registry %s failed: %v", r.URL, err))
		return
	}
	if status != model.Healthy {
		t.HandleBadRequest(fmt.Sprintf("registry %s is unhealthy: %s", r.URL, status))
		return
	}

	id, err := t.manager.Add(r)
	if err != nil {
		log.Errorf("Add registry '%s' error: %v", r.URL, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	t.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put updates a registry
func (t *RegistryAPI) Put() {
	id := t.GetIDFromURL()

	r, err := t.manager.Get(id)
	if err != nil {
		log.Errorf("Get registry by id %d error: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		return
	}

	if r == nil {
		t.HandleNotFound(fmt.Sprintf("Registry %d not found", id))
		return
	}

	req := models.RegistryUpdateRequest{}
	t.DecodeJSONReq(&req)

	originalName := r.Name

	if req.Name != nil {
		r.Name = *req.Name
	}
	if req.Description != nil {
		r.Description = *req.Description
	}
	if req.URL != nil {
		r.URL = *req.URL
	}
	if req.CredentialType != nil {
		r.Credential.Type = (model.CredentialType)(*req.CredentialType)
	}
	if req.AccessKey != nil {
		r.Credential.AccessKey = *req.AccessKey
	}
	if req.AccessSecret != nil {
		r.Credential.AccessSecret = *req.AccessSecret
	}
	if req.Insecure != nil {
		r.Insecure = *req.Insecure
	}

	t.Validate(r)

	if r.Name != originalName {
		reg, err := t.manager.GetByName(r.Name)
		if err != nil {
			log.Errorf("Get registry by name '%s' error: %v", r.Name, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			return
		}

		if reg != nil {
			t.HandleConflict("name is already used")
			return
		}
	}

	status, err := registry.CheckHealthStatus(r)
	if err != nil {
		t.HandleBadRequest(fmt.Sprintf("health check to registry %s failed: %v", r.URL, err))
		return
	}
	if status != model.Healthy {
		t.HandleBadRequest(fmt.Sprintf("registry %s is unhealthy: %s", r.URL, status))
		return
	}

	if err := t.manager.Update(r); err != nil {
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
	if err != nil || id <= 0 {
		t.HandleBadRequest(fmt.Sprintf("invalid registry ID %s", t.GetString(":id")))
		return
	}
	registry, err := t.manager.Get(id)
	if err != nil {
		t.HandleInternalServerError(fmt.Sprintf("failed to get registry %d: %v", id, err))
		return
	}
	if registry == nil {
		t.HandleNotFound(fmt.Sprintf("registry %d not found", id))
		return
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
// TODO remove
func (t *RegistryAPI) GetNamespace() {
	/*
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
	*/
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
