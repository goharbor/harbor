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
	"math/rand"
	"time"

	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/reg"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/replication"
)

// Ctl is a global registry controller instance
var Ctl = NewController()
var regularHealthCheckInterval = 5 * time.Minute

// Controller defines the registry related operations
type Controller interface {
	// Create the registry
	Create(ctx context.Context, registry *model.Registry) (id int64, err error)
	// Count returns the count of registries according to the query
	Count(ctx context.Context, query *q.Query) (count int64, err error)
	// List registries according to the query
	List(ctx context.Context, query *q.Query) (registries []*model.Registry, err error)
	// Get the registry specified by ID
	Get(ctx context.Context, id int64) (registry *model.Registry, err error)
	// Update the specified registry
	Update(ctx context.Context, registry *model.Registry, props ...string) (err error)
	// Delete the registry specified by ID
	Delete(ctx context.Context, id int64) (err error)
	// GetInfo returns the basic information and capabilities of the registry
	GetInfo(ctx context.Context, id int64) (info *model.RegistryInfo, err error)
	// IsHealthy checks whether the provided registry is healthy or not
	IsHealthy(ctx context.Context, registry *model.Registry) (healthy bool, err error)
	// ListRegistryProviderTypes returns all the registered registry provider type
	ListRegistryProviderTypes(ctx context.Context) (types []string, err error)
	// ListRegistryProviderInfos returns all the registered registry provider information
	ListRegistryProviderInfos(ctx context.Context) (infos map[string]*model.AdapterPattern, err error)
	// StartRegularHealthCheck for all registries
	StartRegularHealthCheck(ctx context.Context, closing, done chan struct{})
}

// NewController creates an instance of the registry controller
func NewController() Controller {
	return &controller{
		regMgr: reg.Mgr,
		repMgr: replication.Mgr,
		proMgr: pkg.ProjectMgr,
	}
}

type controller struct {
	regMgr reg.Manager
	repMgr replication.Manager
	proMgr project.Manager
}

func (c *controller) Create(ctx context.Context, registry *model.Registry) (int64, error) {
	if err := c.validate(ctx, registry); err != nil {
		return 0, err
	}
	return c.regMgr.Create(ctx, registry)
}

func (c *controller) validate(ctx context.Context, registry *model.Registry) error {
	if len(registry.Name) == 0 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("name cannot be empty")
	}
	if len(registry.Name) > 64 {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("the max length of name is 64")
	}
	url, err := lib.ValidateHTTPURL(registry.URL)
	if err != nil {
		return err
	}
	registry.URL = url

	healthy, err := c.IsHealthy(ctx, registry)
	if err != nil {
		return err
	}
	if !healthy {
		return errors.New(nil).WithCode(errors.BadRequestCode).WithMessage("the registry is unhealthy")
	}
	registry.Status = model.Healthy
	return nil
}

func (c *controller) Count(ctx context.Context, query *q.Query) (int64, error) {
	return c.regMgr.Count(ctx, query)
}

func (c *controller) List(ctx context.Context, query *q.Query) ([]*model.Registry, error) {
	return c.regMgr.List(ctx, query)
}

func (c *controller) Get(ctx context.Context, id int64) (*model.Registry, error) {
	return c.regMgr.Get(ctx, id)
}

func (c *controller) Update(ctx context.Context, registry *model.Registry, props ...string) error {
	if err := c.validate(ctx, registry); err != nil {
		return err
	}
	return c.regMgr.Update(ctx, registry, props...)
}

func (c *controller) Delete(ctx context.Context, id int64) error {
	// referenced by replication policy as source registry
	count, err := c.repMgr.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"src_registry_id": id,
		},
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(nil).WithCode(errors.PreconditionCode).WithMessage("the registry %d is referenced by replication policies, cannot delete it", id)
	}
	// referenced by replication policy as destination registry
	count, err = c.repMgr.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"dest_registry_id": id,
		},
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(nil).WithCode(errors.PreconditionCode).WithMessage("the registry %d is referenced by replication policies, cannot delete it", id)
	}
	// referenced by proxy cache project
	count, err = c.proMgr.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"registry_id": id,
		},
	})
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(nil).WithCode(errors.PreconditionCode).WithMessage("the registry %d is referenced by proxy cache project, cannot delete it", id)
	}

	return c.regMgr.Delete(ctx, id)
}

func (c *controller) IsHealthy(ctx context.Context, registry *model.Registry) (bool, error) {
	adapter, err := c.regMgr.CreateAdapter(ctx, registry)
	if err != nil {
		return false, err
	}
	status, err := adapter.HealthCheck()
	if err != nil {
		return false, err
	}
	return status == model.Healthy, nil
}

func (c *controller) GetInfo(ctx context.Context, id int64) (*model.RegistryInfo, error) {
	var (
		registry *model.Registry
		err      error
	)
	registry, err = c.regMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	adapter, err := c.regMgr.CreateAdapter(ctx, registry)
	if err != nil {
		return nil, err
	}
	info, err := adapter.Info()
	if err != nil {
		return nil, err
	}

	// currently, only the local Harbor registry supports the event based trigger, append it here
	if id == 0 {
		info.SupportedTriggers = append(info.SupportedTriggers, model.TriggerTypeEventBased)
	}
	info = process(info)
	return info, nil
}

func (c *controller) ListRegistryProviderTypes(ctx context.Context) ([]string, error) {
	return c.regMgr.ListRegistryProviderTypes(ctx)
}

func (c *controller) ListRegistryProviderInfos(ctx context.Context) (map[string]*model.AdapterPattern, error) {
	return c.regMgr.ListRegistryProviderInfos(ctx)
}

func (c *controller) StartRegularHealthCheck(ctx context.Context, closing, done chan struct{}) {
	// Wait some random time before starting health checking. If Harbor is deployed in HA mode
	// with multiple instances, this will avoid instances check health in the same time.
	<-time.After(time.Duration(rand.Int63n(int64(regularHealthCheckInterval))))

	ticker := time.NewTicker(regularHealthCheckInterval)
	log.Infof("Start regular health check for registries with interval %v", regularHealthCheckInterval)
	for {
		select {
		case <-ticker.C:
			registries, err := c.regMgr.List(ctx, nil)
			if err != nil {
				log.Errorf("failed to list registries: %v", err)
				continue
			}
			for _, registry := range registries {
				isHealthy, err := c.IsHealthy(ctx, registry)
				if err != nil {
					log.Errorf("failed to check health of registry %d: %v", registry.ID, err)
					continue
				}
				status := model.Healthy
				if !isHealthy {
					status = model.Unhealthy
				}
				if registry.Status == status {
					continue
				}
				registry.Status = status
				if err = c.regMgr.Update(ctx, registry, "Status"); err != nil {
					log.Errorf("failed to update the status of registry %d: %v", registry.ID, err)
					continue
				}
				log.Debugf("update the status of registry %d to %s", registry.ID, status)
			}
		case <-closing:
			log.Info("Stop registry health checker")
			// No cleanup works to do, signal done directly
			close(done)
			return
		}
	}
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
