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

package manifest

import (
	"context"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// ChildClassifier classifies a child descriptor of an OCI index or Docker manifest list.
type ChildClassifier interface {
	// Classify returns a nil candidate and a nil error when the descriptor is not
	// its concern, leaving it to the next classifier.
	Classify(ctx context.Context, repository string, descriptor v1.Descriptor, siblings []v1.Descriptor) (*artifact.AccessoryCandidate, error)
}

// ChildClassifiers holds the registered classifiers, in registration order.
var ChildClassifiers []ChildClassifier

// RegisterChildClassifier appends a classifier to the chain. A nil would panic on
// every index push, far from the faulty registration, so it is dropped here.
func RegisterChildClassifier(classifier ChildClassifier) {
	if classifier == nil {
		log.Errorf("refusing to register a nil child classifier")
		return
	}
	ChildClassifiers = append(ChildClassifiers, classifier)
}

// ClassifyChild returns the first accessory candidate claimed for the descriptor.
// A nil candidate means it is an ordinary platform child of the index.
func ClassifyChild(ctx context.Context, repository string, descriptor v1.Descriptor, siblings []v1.Descriptor) (*artifact.AccessoryCandidate, error) {
	for _, classifier := range ChildClassifiers {
		candidate, err := classifier.Classify(ctx, repository, descriptor, siblings)
		if err != nil {
			return nil, err
		}
		if candidate != nil {
			return candidate, nil
		}
	}
	//nolint:nilnil // no classifier claimed the descriptor
	return nil, nil
}
