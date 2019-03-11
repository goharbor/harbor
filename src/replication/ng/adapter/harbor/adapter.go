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

package harbor

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	adp "github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

const (
	harbor model.RegistryType = "Harbor"
)

func init() {
	// TODO add more information to the info
	info := &adp.Info{
		Type:                   harbor,
		SupportedResourceTypes: []model.ResourceType{model.ResourceTypeRepository},
	}
	if err := adp.RegisterFactory(info, func(registry *model.Registry) (adp.Adapter, error) {
		return newAdapter(registry), nil
	}); err != nil {
		log.Errorf("failed to register factory for %s: %v", harbor, err)
		return
	}
	log.Infof("the factory for adapter %s registered", harbor)
}

// TODO implement the functions
type adapter struct {
	*adp.DefaultImageRegistry
}

func newAdapter(registry *model.Registry) *adapter {
	return &adapter{}
}

func (a *adapter) ListNamespaces(*model.NamespaceQuery) ([]*model.Namespace, error) {
	return nil, nil
}
func (a *adapter) CreateNamespace(*model.Namespace) error {
	return nil
}
func (a *adapter) GetNamespace(string) (*model.Namespace, error) {
	return nil, nil
}
