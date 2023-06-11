//
// Copyright 2021 The Sigstore Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package empty

import (
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/types"

	"github.com/sigstore/cosign/pkg/oci"
)

// Signatures constructs an empty oci.Signatures.
func Signatures() oci.Signatures {
	base := empty.Image
	if !oci.DockerMediaTypes() {
		base = mutate.MediaType(base, types.OCIManifestSchema1)
		base = mutate.ConfigMediaType(base, types.OCIConfigJSON)
	}
	return &emptyImage{
		Image: base,
	}
}

type emptyImage struct {
	v1.Image
}

var _ oci.Signatures = (*emptyImage)(nil)

// Get implements oci.Signatures
func (*emptyImage) Get() ([]oci.Signature, error) {
	return nil, nil
}
