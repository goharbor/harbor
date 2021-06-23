//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package validate

import (
	"context"
	"github.com/goharbor/harbor/src/lib/config"
)

// Rule rules to validate the configure input
// Steps to add validation rule for configure parameter
// 1. Create/find the model need to validate, add validate tag in the field.
// 2. Implement the Rule interface like LdapGroupValidateRule
// 3. Add the rule to validateRules in pkg/config/manager.go
type Rule interface {
	// Validate validates a specific group of configuration items, cfgs contains the config need to be updated, the final config should merged with the cfgMgr
	Validate(ctx context.Context, cfgMgr config.Manager, cfgs map[string]interface{}) error
}
