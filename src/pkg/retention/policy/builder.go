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

package policy

import (
	"github.com/goharbor/harbor/src/pkg/retention/policy/alg"
	"github.com/pkg/errors"
)

// Builder builds the runnable processor from the raw policy
type Builder interface {
	// Builds runnable processor
	//
	//  Arguments:
	//    rawPolicy string : the simple retention policy with JSON format
	//
	//  Returns:
	//    Processor : a processor implementation to process the candidates
	//    error     : common error object if any errors occurred
	Build(rawPolicy string) (alg.Processor, error)
}

// basicBuilder is default implementation of Builder interface
type basicBuilder struct{}

// Build policy processor from the raw policy
func (bb *basicBuilder) Build(rawPolicy string) (alg.Processor, error) {
	if len(rawPolicy) == 0 {
		return nil, errors.New("empty raw policy to build processor")
	}

	// Decode metadata
	liteMeta := &LiteMeta{}
	if err := liteMeta.Decode(rawPolicy); err != nil {
		return nil, errors.Wrap(err, "build policy processor")
	}

	switch liteMeta.Algorithm {
	case AlgorithmOR:
	default:
		return nil, errors.Errorf("algorithm %s is not supported", liteMeta.Algorithm)
	}

	return nil, errors.New("not implemented")
}
