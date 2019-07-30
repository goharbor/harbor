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

package index

import (
	"sync"

	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg/or"
	"github.com/pkg/errors"
)

const (
	// AlgorithmOR for || algorithm
	AlgorithmOR = "or"
)

// index for keeping the mapping between algorithm and its processor
var index sync.Map

func init() {
	// Register or
	Register(AlgorithmOR, or.New)
}

// Register processor with the algorithm
func Register(algorithm string, processor alg.Factory) {
	if len(algorithm) > 0 && processor != nil {
		index.Store(algorithm, processor)
	}
}

// Get Processor
func Get(algorithm string, params []*alg.Parameter) (alg.Processor, error) {
	v, ok := index.Load(algorithm)
	if !ok {
		return nil, errors.Errorf("no processor registered with algorithm: %s", algorithm)
	}

	if fac, ok := v.(alg.Factory); ok {
		return fac(params), nil
	}

	return nil, errors.Errorf("no valid processor registered for algorithm: %s", algorithm)
}
