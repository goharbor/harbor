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

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/dao"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
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
		status, err := CheckHealthStatus(r)
		if err != nil {
			log.Warningf("Check health status for %s error: %v", r.URL, err)
		}
		r.Status = string(status)
		err = m.Update(r, "status")
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

// CheckHealthStatus checks status of a given registry
func CheckHealthStatus(r *model.Registry) (model.HealthStatus, error) {
	if !adapter.HasFactory(r.Type) {
		return model.Unknown, fmt.Errorf("no adapter factory for type '%s' registered", r.Type)
	}

	factory, err := adapter.GetFactory(r.Type)
	if err != nil {
		return model.Unknown, fmt.Errorf("get adaper for type '%s' error: %v", r.Type, err)
	}

	rAdapter, err := factory(r)
	if err != nil {
		return model.Unknown, fmt.Errorf("generate '%s' type adapter form factory error: %v", r.Type, err)
	}

	return rAdapter.HealthCheck()
}

// decrypt checks whether access secret is set in the registry, if so, decrypt it.
func decrypt(secret string) (string, error) {
	if len(secret) == 0 {
		return "", nil
	}

	decrypted, err := utils.ReversibleDecrypt(secret, config.Config.SecretKey)
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

	encrypted, err := utils.ReversibleEncrypt(secret, config.Config.SecretKey)
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

// fromDaoModel converts DAO layer registry model to replication model.
// Also, if access secret is provided, decrypt it.
func fromDaoModel(registry *models.Registry) (*model.Registry, error) {
	var decrypted string
	var err error
	if len(registry.AccessSecret) != 0 {
		decrypted, err = decrypt(registry.AccessSecret)
		if err != nil {
			return nil, err
		}
	}

	r := &model.Registry{
		ID:           registry.ID,
		Name:         registry.Name,
		Description:  registry.Description,
		Type:         model.RegistryType(registry.Type),
		Credential:   &model.Credential{},
		URL:          registry.URL,
		Insecure:     registry.Insecure,
		Status:       registry.Health,
		CreationTime: registry.CreationTime,
		UpdateTime:   registry.UpdateTime,
	}

	if len(registry.CredentialType) != 0 && len(registry.AccessKey) != 0 {
		r.Credential = &model.Credential{
			Type:         model.CredentialType(registry.CredentialType),
			AccessKey:    registry.AccessKey,
			AccessSecret: decrypted,
		}
	}

	return r, nil
}

// toDaoModel converts registry model from replication to DAO layer model.
// Also, if access secret is provided, encrypt it.
func toDaoModel(registry *model.Registry) (*models.Registry, error) {
	m := &models.Registry{
		ID:           registry.ID,
		URL:          registry.URL,
		Name:         registry.Name,
		Type:         string(registry.Type),
		Insecure:     registry.Insecure,
		Description:  registry.Description,
		Health:       registry.Status,
		CreationTime: registry.CreationTime,
		UpdateTime:   registry.UpdateTime,
	}

	if registry.Credential != nil && len(registry.Credential.AccessKey) != 0 {
		var encrypted string
		var err error
		if len(registry.Credential.AccessSecret) != 0 {
			encrypted, err = encrypt(registry.Credential.AccessSecret)
			if err != nil {
				return nil, err
			}
		}

		m.CredentialType = string(registry.Credential.Type)
		m.AccessKey = registry.Credential.AccessKey
		m.AccessSecret = encrypted
	}

	return m, nil
}
