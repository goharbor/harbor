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
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	// TODO use the replication config rather than the core
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/replication/ng/dao"
	"github.com/goharbor/harbor/src/replication/ng/dao/models"
	"github.com/goharbor/harbor/src/replication/ng/model"
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
	// Add new registry
	Add(*model.Registry) (int64, error)
	// List registries, returns total count, registry list and error
	List(...*model.RegistryQuery) (int64, []*model.Registry, error)
	// Get the specified registry
	Get(int64) (*model.Registry, error)
	// GetByName gets registry by name
	GetByName(name string) (*model.Registry, error)
	// Update the registry, the "props" are the properties of registry
	// that need to be updated
	Update(registry *model.Registry, props ...string) error
	// Remove the registry with the specified ID
	Remove(int64) error
	// HealthCheck checks health status of all registries and update result in database
	HealthCheck() error
}

// DefaultManager implement the Manager interface
type DefaultManager struct{}

// NewDefaultManager returns an instance of DefaultManger
func NewDefaultManager() *DefaultManager {
	return &DefaultManager{}
}

// Ensure *DefaultManager has implemented Manager interface.
var _ Manager = (*DefaultManager)(nil)

// Get gets a registry by id
func (m *DefaultManager) Get(id int64) (*model.Registry, error) {
	registry, err := dao.GetRegistry(id)
	if err != nil {
		return nil, err
	}

	if registry == nil {
		return nil, nil
	}

	return fromDaoModel(registry)
}

// GetByName gets a registry by its name
func (m *DefaultManager) GetByName(name string) (*model.Registry, error) {
	registry, err := dao.GetRegistryByName(name)
	if err != nil {
		return nil, err
	}

	if registry == nil {
		return nil, nil
	}

	return fromDaoModel(registry)
}

// List lists registries according to query provided.
func (m *DefaultManager) List(query ...*model.RegistryQuery) (int64, []*model.Registry, error) {
	var registryQueries []*dao.ListRegistryQuery
	if len(query) > 0 {
		// limit being -1 indicates no pagination specified, result in all registries matching name returned.
		listQuery := &dao.ListRegistryQuery{
			Query: query[0].Name,
			Limit: -1,
		}
		if query[0].Pagination != nil {
			listQuery.Offset = query[0].Pagination.Page * query[0].Pagination.Size
			listQuery.Limit = query[0].Pagination.Size
		}

		registryQueries = append(registryQueries, listQuery)
	}
	total, registries, err := dao.ListRegistries(registryQueries...)
	if err != nil {
		return -1, nil, err
	}

	var results []*model.Registry
	for _, r := range registries {
		registry, err := fromDaoModel(r)
		if err != nil {
			return -1, nil, err
		}
		results = append(results, registry)
	}

	return total, results, nil
}

// Add adds a new registry
func (m *DefaultManager) Add(registry *model.Registry) (int64, error) {
	r, err := toDaoModel(registry)
	if err != nil {
		log.Errorf("Convert registry model to dao layer model error: %v", err)
		return -1, err
	}

	id, err := dao.AddRegistry(r)
	if err != nil {
		log.Errorf("Add registry error: %v", err)
		return -1, err
	}

	return id, nil
}

// Update updates a registry
func (m *DefaultManager) Update(registry *model.Registry, props ...string) error {
	// TODO(ChenDe): Only update the given props

	r, err := toDaoModel(registry)
	if err != nil {
		log.Errorf("Convert registry model to dao layer model error: %v", err)
		return err
	}

	return dao.UpdateRegistry(r)
}

// Remove deletes a registry
func (m *DefaultManager) Remove(id int64) error {
	if err := dao.DeleteRegistry(id); err != nil {
		log.Errorf("Delete registry %d error: %v", id, err)
		return err
	}

	return nil
}

// HealthCheck checks health status of every registries and update their status. It will check whether a registry
// is reachable and the credential is valid
func (m *DefaultManager) HealthCheck() error {
	_, registries, err := m.List()
	if err != nil {
		return err
	}

	errCount := 0
	for _, r := range registries {
		status, err := healthStatus(r)
		if err != nil {
			log.Warningf("Check health status for %s error: %v", r.URL, err)
		}
		r.Status = string(status)
		err = m.Update(r)
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

func healthStatus(r *model.Registry) (HealthStatus, error) {
	// TODO(ChenDe): Support other credential type like OAuth, for the moment, only basic auth is supported.
	if r.Credential.Type != model.CredentialTypeBasic {
		return Unknown, fmt.Errorf("unknown credential type '%s', only '%s' supported yet", r.Credential.Type, model.CredentialTypeBasic)
	}

	// TODO(ChenDe): Support health check for other kinds of registry
	if r.Type != model.RegistryTypeHarbor {
		return Unknown, fmt.Errorf("unknown registry type '%s'", model.RegistryTypeHarbor)
	}

	transport := registry.GetHTTPTransport(r.Insecure)
	credential := auth.NewBasicAuthCredential(r.Credential.AccessKey, r.Credential.AccessSecret)
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

// decrypt checks whether access secret is set in the registry, if so, decrypt it.
func decrypt(secret string) (string, error) {
	if len(secret) == 0 {
		return "", nil
	}

	key, err := config.SecretKey()
	if err != nil {
		return "", err
	}
	decrypted, err := utils.ReversibleDecrypt(secret, key)
	if err != nil {
		return "", err
	}

	return decrypted, nil
}

// encrypt checks whether access secret is set in the registry, if so, encrypt it.
func encrypt(secret string) (string, error) {
	if len(secret) == 0 {
		return secret, nil
	}

	key, err := config.SecretKey()
	if err != nil {
		return "", err
	}
	encrypted, err := utils.ReversibleEncrypt(secret, key)
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

// fromDaoModel converts DAO layer registry model to replication model.
// Also, if access secret is provided, decrypt it.
func fromDaoModel(registry *models.Registry) (*model.Registry, error) {
	decrypted, err := decrypt(registry.AccessSecret)
	if err != nil {
		return nil, err
	}

	r := &model.Registry{
		ID:          registry.ID,
		Name:        registry.Name,
		Description: registry.Description,
		Type:        model.RegistryType(registry.Type),
		URL:         registry.URL,
		Credential: &model.Credential{
			Type:         model.CredentialType(registry.CredentialType),
			AccessKey:    registry.AccessKey,
			AccessSecret: decrypted,
		},
		Insecure:     registry.Insecure,
		Status:       registry.Health,
		CreationTime: registry.CreationTime,
		UpdateTime:   registry.UpdateTime,
	}

	return r, nil
}

// toDaoModel converts registry model from replication to DAO layer model.
// Also, if access secret is provided, encrypt it.
func toDaoModel(registry *model.Registry) (*models.Registry, error) {
	var encrypted string
	var err error
	if registry.Credential != nil {
		encrypted, err = encrypt(registry.Credential.AccessSecret)
		if err != nil {
			return nil, err
		}
	}

	return &models.Registry{
		ID:             registry.ID,
		URL:            registry.URL,
		Name:           registry.Name,
		CredentialType: string(registry.Credential.Type),
		AccessKey:      registry.Credential.AccessKey,
		AccessSecret:   encrypted,
		Type:           string(registry.Type),
		Insecure:       registry.Insecure,
		Description:    registry.Description,
		Health:         registry.Status,
		CreationTime:   registry.CreationTime,
		UpdateTime:     registry.UpdateTime,
	}, nil
}
