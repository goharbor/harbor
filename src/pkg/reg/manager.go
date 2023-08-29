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

package reg

import (
	"context"

	commonthttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/dao"
	"github.com/goharbor/harbor/src/pkg/reg/model"

	// register the AliACR adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/aliacr"
	// register the AwsEcr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/awsecr"
	// register the AzureAcr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/azurecr"
	// register the DockerHub adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/dockerhub"
	// register the DTR adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/dtr"
	// register the Github Container Registry adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/githubcr"
	// register the GitLab adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/gitlab"
	// register the Google Gcr adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/googlegcr"
	// register the Harbor adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/harbor"
	// register the huawei adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/huawei"
	// register the Jfrog Artifactory adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/jfrog"
	// register the Native adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/native"
	// register the Quay.io adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/quay"
	// register the TencentCloud TCR adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/tencentcr"
	// register the VolcEngine CR Registry adapter
	_ "github.com/goharbor/harbor/src/pkg/reg/adapter/volcenginecr"
)

var (
	// Mgr is the global registry manager instance
	Mgr = NewManager()
)

// Manager defines the registry related operations
type Manager interface {
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
	// CreateAdapter for the provided registry
	CreateAdapter(ctx context.Context, registry *model.Registry) (adapter adapter.Adapter, err error)
	// ListRegistryProviderTypes returns all the registered registry provider type
	ListRegistryProviderTypes(ctx context.Context) (types []string, err error)
	// ListRegistryProviderInfos returns all the registered registry provider information
	ListRegistryProviderInfos(ctx context.Context) (infos map[string]*model.AdapterPattern, err error)
}

// NewManager creates an instance of registry manager
func NewManager() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context, registry *model.Registry) (int64, error) {
	reg, err := toDaoModel(registry)
	if err != nil {
		return 0, err
	}
	return m.dao.Create(ctx, reg)
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.dao.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.Registry, error) {
	registries, err := m.dao.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var regs []*model.Registry
	for _, registry := range registries {
		r, err := fromDaoModel(registry)
		if err != nil {
			return nil, err
		}
		regs = append(regs, r)
	}
	return regs, nil
}

func (m *manager) Get(ctx context.Context, id int64) (*model.Registry, error) {
	if id == 0 {
		return getLocalRegistry(), nil
	}
	registry, err := m.dao.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return fromDaoModel(registry)
}

func (m *manager) Update(ctx context.Context, registry *model.Registry, props ...string) error {
	reg, err := toDaoModel(registry)
	if err != nil {
		return err
	}
	return m.dao.Update(ctx, reg, props...)
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

func (m *manager) CreateAdapter(_ context.Context, registry *model.Registry) (adapter.Adapter, error) {
	factory, err := adapter.GetFactory(registry.Type)
	if err != nil {
		return nil, err
	}
	return factory.Create(registry)
}

func (m *manager) ListRegistryProviderTypes(_ context.Context) ([]string, error) {
	return adapter.ListRegisteredAdapterTypes(), nil
}

func (m *manager) ListRegistryProviderInfos(_ context.Context) (infos map[string]*model.AdapterPattern, err error) {
	return adapter.ListRegisteredAdapterInfos(), nil
}

// getLocalRegistry returns the info of the local Harbor registry
func getLocalRegistry() *model.Registry {
	return &model.Registry{
		Type:            model.RegistryTypeHarbor,
		Name:            "Local",
		URL:             config.InternalCoreURL(),
		TokenServiceURL: config.InternalTokenServiceEndpoint(),
		Status:          "healthy",
		Credential: &model.Credential{
			Type: model.CredentialTypeSecret,
			// use secret to do the auth for the local Harbor
			AccessSecret: config.JobserviceSecret(),
		},
		Insecure: !commonthttp.InternalTLSEnabled(),
	}
}

// decrypt checks whether access secret is set in the registry, if so, decrypt it.
func decrypt(secret string) (string, error) {
	if len(secret) == 0 {
		return "", nil
	}
	secretKey, err := config.SecretKey()
	if err != nil {
		return "", nil
	}
	decrypted, err := utils.ReversibleDecrypt(secret, secretKey)
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
	secretKey, err := config.SecretKey()
	if err != nil {
		return "", nil
	}
	encrypted, err := utils.ReversibleEncrypt(secret, secretKey)
	if err != nil {
		return "", err
	}

	return encrypted, nil
}

// FromDaoModel converts DAO layer registry model to replication model.
// Also, if access secret is provided, decrypt it.
func fromDaoModel(registry *dao.Registry) (*model.Registry, error) {
	r := &model.Registry{
		ID:           registry.ID,
		Name:         registry.Name,
		Description:  registry.Description,
		Type:         registry.Type,
		Credential:   &model.Credential{},
		URL:          registry.URL,
		Insecure:     registry.Insecure,
		Status:       registry.Status,
		CreationTime: registry.CreationTime,
		UpdateTime:   registry.UpdateTime,
	}

	if len(registry.AccessKey) != 0 {
		credentialType := registry.CredentialType
		if len(credentialType) == 0 {
			credentialType = model.CredentialTypeBasic
		}
		decrypted, err := decrypt(registry.AccessSecret)
		if err != nil {
			return nil, err
		}
		r.Credential = &model.Credential{
			Type:         credentialType,
			AccessKey:    registry.AccessKey,
			AccessSecret: decrypted,
		}
	}

	return r, nil
}

// ToDaoModel converts registry model from replication to DAO layer model.
// Also, if access secret is provided, encrypt it.
func toDaoModel(registry *model.Registry) (*dao.Registry, error) {
	m := &dao.Registry{
		ID:           registry.ID,
		URL:          registry.URL,
		Name:         registry.Name,
		Type:         string(registry.Type),
		Insecure:     registry.Insecure,
		Description:  registry.Description,
		Status:       registry.Status,
		CreationTime: registry.CreationTime,
		UpdateTime:   registry.UpdateTime,
	}

	if registry.Credential != nil && len(registry.Credential.AccessKey) != 0 {
		credentialType := registry.Credential.Type
		if len(credentialType) == 0 {
			credentialType = model.CredentialTypeBasic
		}
		encrypted, err := encrypt(registry.Credential.AccessSecret)
		if err != nil {
			return nil, err
		}

		m.CredentialType = string(credentialType)
		m.AccessKey = registry.Credential.AccessKey
		m.AccessSecret = encrypted
	}

	return m, nil
}
