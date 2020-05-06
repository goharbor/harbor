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
	"github.com/goharbor/harbor/src/lib/log"
	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/adapter/harbor/base"
	v1 "github.com/goharbor/harbor/src/replication/adapter/harbor/v1"
	v2 "github.com/goharbor/harbor/src/replication/adapter/harbor/v2"
	"github.com/goharbor/harbor/src/replication/model"
)

func init() {
	if err := adp.RegisterFactory(model.RegistryTypeHarbor, new(factory)); err != nil {
		log.Errorf("failed to register factory for %s: %v", model.RegistryTypeHarbor, err)
		return
	}
	log.Infof("the factory for adapter %s registered", model.RegistryTypeHarbor)
}

type factory struct {
}

func (f *factory) Create(r *model.Registry) (adp.Adapter, error) {
	base, err := base.New(r)
	if err != nil {
		return nil, err
	}
	version := base.GetAPIVersion()
	// no API version, it's instance of Harbor 1.x
	if len(version) == 0 {
		log.Debug("no API version, create the v1 adapter")
		return v1.New(base), nil
	}
	log.Debugf("API version is %s, create the v2 adapter", version)
	return v2.New(base), nil
}

func (f *factory) AdapterPattern() *model.AdapterPattern {
	return nil
}
