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

package auth

import (
	"github.com/goharbor/harbor/src/common/http/modifier"
	"github.com/goharbor/harbor/src/core/config"
	"sync"
)

// NewBasicAuthorizer create an authorizer to add basic auth header as is set in the parameter
func NewBasicAuthorizer(u, p string) modifier.Modifier {
	return NewBasicAuthCredential(u, p)
}

var (
	defaultAuthorizer modifier.Modifier
	once              sync.Once
)

// DefaultBasicAuthorizer returns the basic authorizer that sets the basic auth as configured in env variables
func DefaultBasicAuthorizer() modifier.Modifier {
	once.Do(func() {
		u, p := config.RegistryCredential()
		defaultAuthorizer = NewBasicAuthCredential(u, p)
	})
	return defaultAuthorizer
}
