package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/ng/registry"
)

// RegistryAPI handles requests to /api/registries/{}. It manages registries integrated to Harbor.
type RegistryAPI struct {
	BaseController
	manager registry.Manager
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

	t.manager = registry.NewDefaultManager()
	if t.manager == nil {
		log.Error("failed to create registry manager")
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Get gets a registry by id.
func (t *RegistryAPI) Get() {
	id := t.GetIDFromURL()

	registry, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("failed to get registry %d: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if registry == nil {
		t.HandleNotFound(fmt.Sprintf("registry %d not found", id))
		return
	}

	// Hide password
	registry.Password = ""

	t.Data["json"] = registry
	t.ServeJSON()
}

// List lists all registries that match a given registry name.
func (t *RegistryAPI) List() {
	name := t.GetString("name")
	registries, err := dao.FilterRepTargets(name)
	if err != nil {
		log.Errorf("failed to filter registries %s: %v", name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	// Hide passwords
	for _, registry := range registries {
		registry.Password = ""
	}

	t.Data["json"] = registries
	t.ServeJSON()
	return
}

// Post creates a registry
func (t *RegistryAPI) Post() {
	registry := &models.RepTarget{}
	t.DecodeJSONReqAndValidate(registry)

	reg, err := dao.GetRepTargetByName(registry.Name)
	if err != nil {
		log.Errorf("failed to get registry %s: %v", registry.Name, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if reg != nil {
		t.HandleConflict(fmt.Sprintf("name '%s' is already used"), registry.Name)
		return
	}

	reg, err = dao.GetRepTargetByEndpoint(registry.URL)
	if err != nil {
		log.Errorf("failed to get registry by URL [ %s ]: %v", registry.URL, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if reg != nil {
		t.HandleConflict(fmt.Sprintf("registry with endpoint '%s' already exists", registry.URL))
		return
	}

	id, err := t.manager.AddRegistry(registry)
	if err != nil {
		log.Errorf("Add registry '%s' error: %v", registry.URL, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	t.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// Put updates a registry
func (t *RegistryAPI) Put() {
	id := t.GetIDFromURL()

	registry, err := t.manager.GetRegistry(id)
	if err != nil {
		log.Errorf("Get registry by id %d error: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	req := struct {
		Name     *string `json:"name"`
		Endpoint *string `json:"endpoint"`
		Username *string `json:"username"`
		Password *string `json:"password"`
		Insecure *bool   `json:"insecure"`
	}{}
	t.DecodeJSONReq(&req)

	originalName := registry.Name
	originalURL := registry.URL

	if req.Name != nil {
		registry.Name = *req.Name
	}
	if req.Endpoint != nil {
		registry.URL = *req.Endpoint
	}
	if req.Username != nil {
		registry.Username = *req.Username
	}
	if req.Password != nil {
		registry.Password = *req.Password
	}
	if req.Insecure != nil {
		registry.Insecure = *req.Insecure
	}

	t.Validate(registry)

	if registry.Name != originalName {
		reg, err := dao.GetRepTargetByName(registry.Name)
		if err != nil {
			log.Errorf("Get registry by name '%s' error: %v", registry.Name, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if reg != nil {
			t.HandleConflict("name is already used")
			return
		}
	}

	if registry.URL != originalURL {
		reg, err := dao.GetRepTargetByEndpoint(registry.URL)
		if err != nil {
			log.Errorf("Get registry by URL '%s' error: %v", registry.URL, err)
			t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if reg != nil {
			t.HandleConflict(fmt.Sprintf("registry with endpoint '%s' already exists", registry.URL))
			return
		}
	}

	if err := t.manager.UpdateRegistry(registry); err != nil {
		log.Errorf("Update registry %d error: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Delete deletes a registry
func (t *RegistryAPI) Delete() {
	id := t.GetIDFromURL()

	registry, err := dao.GetRepTarget(id)
	if err != nil {
		log.Errorf("Get registry %d error: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if registry == nil {
		t.HandleNotFound(fmt.Sprintf("target %d not found", id))
		return
	}

	if err := t.manager.DeleteRegistry(id); err != nil {
		log.Errorf("Delete registry %d error: %v", id, err)
		t.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}
