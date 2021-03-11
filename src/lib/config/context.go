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

package config

import (
	"context"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/config/metadata"
)

type cfgMgrKey struct{}

// Manager defines the operation for config
type Manager interface {
	Load(ctx context.Context) error
	Set(ctx context.Context, key string, value interface{})
	Save(ctx context.Context) error
	Get(ctx context.Context, key string) *metadata.ConfigureValue
	UpdateConfig(ctx context.Context, cfgs map[string]interface{}) error
	GetUserCfgs(ctx context.Context) map[string]interface{}
	ValidateCfg(ctx context.Context, cfgs map[string]interface{}) error
	GetAll(ctx context.Context) map[string]interface{}
	GetDatabaseCfg() *models.Database
}

// FromContext returns CfgManager from context
func FromContext(ctx context.Context) (Manager, bool) {
	m, ok := ctx.Value(cfgMgrKey{}).(Manager)
	return m, ok
}

// NewContext returns context with CfgManager
func NewContext(ctx context.Context, m Manager) context.Context {
	return context.WithValue(ctx, cfgMgrKey{}, m)
}
