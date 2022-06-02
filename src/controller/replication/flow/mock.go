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

package flow

import "github.com/goharbor/harbor/src/pkg/reg/adapter"

// define a new interface to combine the two interfaces of adapter for mockery to generate the mocks

// nolint:deadcode
// for make gen_mocks use
type registryAdapter interface {
	adapter.Adapter
	adapter.ArtifactRegistry
}

//go:generate mockery --dir . --name registryAdapter --output . --outpkg flow --filename mock_adapter_test.go --structname mockAdapter
//go:generate mockery --dir ../../../pkg/reg/adapter --name Factory --output . --outpkg flow --filename mock_adapter_factory_test.go --structname mockFactory
