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

package validate

import (
	"context"
	"regexp"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
)

// OIDCGroupFilterValidateRule validates the OIDC group filter, which is a regular expression.
// A filter that does not compile is ignored when a user logs in and every group of the
// token is onboarded, so it must be rejected when it is saved.
type OIDCGroupFilterValidateRule struct {
}

// Validate checks the OIDC group filter in cfgs, an empty filter means no filtering.
// A filter that is present but is not a string is rejected too, as it would be stored
// as text that is not a valid regular expression.
func (o OIDCGroupFilterValidateRule) Validate(_ context.Context, _ config.Manager, cfgs map[string]any) error {
	val, exist := cfgs[common.OIDCGroupFilter]
	if !exist {
		return nil
	}
	filter, ok := val.(string)
	if !ok {
		return errors.Errorf("invalid %s: expected a string but got %T", common.OIDCGroupFilter, val)
	}
	if len(filter) == 0 {
		return nil
	}
	if _, err := regexp.Compile(filter); err != nil {
		return errors.Errorf("invalid %s %q: %v", common.OIDCGroupFilter, filter, err)
	}
	return nil
}
