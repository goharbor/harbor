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
	"errors"
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/core/config"
)

// HealthStatus describes whether a target is healthy or not
type HealthStatus string

const (
	// Healthy indicates target is healthy
	Healthy = "healthy"
	// Unhealthy indicates target is unhealthy
	Unhealthy = "unhealthy"
	// Unknown indicates health status of target is unknown
	Unknown = "unknown"
)

// Manager defines the methods that a target manager should implement
type Manager interface {
	GetRegistry(int64) (*models.RepTarget, error)
	AddRegistry(*models.RepTarget) (int64, error)
	UpdateRegistry(*models.RepTarget) error
	DeleteRegistry(int64) error
	HealthCheck() error
}

// DefaultManager implement the Manager interface
type DefaultManager struct{}

// NewDefaultManager returns an instance of DefaultManger
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// GetRegistry gets a registry by id
func (m *DefaultManager) GetRegistry(id int64) (*models.RepTarget, error) {
	target, err := dao.GetRepTarget(id)
	if err != nil {
		return nil, err
	}

	if target == nil {
		return nil, fmt.Errorf("target '%d' does not exist", id)
	}

	// decrypt the password
	if len(target.Password) > 0 {
		key, err := config.SecretKey()
		if err != nil {
			return nil, err
		}
		pwd, err := utils.ReversibleDecrypt(target.Password, key)
		if err != nil {
			return nil, err
		}
		target.Password = pwd
	}
	return target, nil
}

// AddRegistry adds a new registry
func (m *DefaultManager) AddRegistry(registry *models.RepTarget) (int64, error) {
	var err error
	if len(registry.Password) != 0 {
		key, err := config.SecretKey()
		if err != nil {
			return -1, err
		}
		registry.Password, err = utils.ReversibleEncrypt(registry.Password, key)
		if err != nil {
			log.Errorf("failed to encrypt password: %v", err)
			return -1, err
		}
	}

	id, err := dao.AddRepTarget(*registry)
	if err != nil {
		log.Errorf("failed to add registry: %v", err)
	}
	return id, nil
}

// UpdateRegistry updates a registry
func (m *DefaultManager) UpdateRegistry(registry *models.RepTarget) error {
	// Encrypt the password if set
	if len(registry.Password) > 0 {
		key, err := config.SecretKey()
		if err != nil {
			return err
		}
		pwd, err := utils.ReversibleEncrypt(registry.Password, key)
		if err != nil {
			return err
		}
		registry.Password = pwd
	}

	return dao.UpdateRepTarget(*registry)
}

// DeleteRegistry deletes a registry
func (m *DefaultManager) DeleteRegistry(id int64) error {
	policies, err := dao.GetRepPolicyByTarget(id)
	if err != nil {
		log.Errorf("Get policies related to registry %d error: %v", id, err)
		return err
	}

	if len(policies) > 0 {
		msg := fmt.Sprintf("Can't delete registry with replication policies, %d found", len(policies))
		log.Error(msg)
		return errors.New(msg)
	}

	if err = dao.DeleteRepTarget(id); err != nil {
		log.Errorf("Delete registry %d error: %v", id, err)
		return err
	}

	return nil
}

// HealthCheck checks health status of every registries and update their status. It will check whether a registry
// is reachable and the credential is valid
func (m *DefaultManager) HealthCheck() error {
	registries, err := dao.FilterRepTargets("")
	if err != nil {
		return err
	}

	errCount := 0
	for _, r := range registries {
		status, _ := healthStatus(r)
		r.Health = string(status)
		err := m.UpdateRegistry(r)
		if err != nil {
			log.Warningf("Update health status for '%s' error: %v", r.URL, err)
			errCount++
			continue
		}
	}

	if errCount > 0 {
		return fmt.Errorf("%d out of %d registries failed to update health status", errCount, len(registries))
	}
	return nil
}

func healthStatus(r *models.RepTarget) (HealthStatus, error) {
	transport := registry.GetHTTPTransport(r.Insecure)
	credential := auth.NewBasicAuthCredential(r.Username, r.Password)
	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential)
	registry, err := registry.NewRegistry(r.URL, &http.Client{
		Transport: registry.NewTransport(transport, authorizer),
	})
	if err != nil {
		return Unknown, err
	}

	err = registry.Ping()
	if err != nil {
		return Unhealthy, err
	}

	return Healthy, nil
}
