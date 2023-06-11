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

package remote

import (
	"errors"
	"net/http"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"github.com/sigstore/cosign/pkg/oci"
	"github.com/sigstore/cosign/pkg/oci/empty"
	"github.com/sigstore/cosign/pkg/oci/internal/signature"
)

// Signatures fetches the signatures image represented by the named reference.
// If the tag is not found, this returns an empty oci.Signatures.
func Signatures(ref name.Reference, opts ...Option) (oci.Signatures, error) {
	o := makeOptions(ref.Context(), opts...)
	img, err := remoteImage(ref, o.ROpt...)
	var te *transport.Error
	if errors.As(err, &te) {
		if te.StatusCode != http.StatusNotFound {
			return nil, te
		}
		return empty.Signatures(), nil
	} else if err != nil {
		return nil, err
	}
	return &sigs{
		Image: img,
	}, nil
}

type sigs struct {
	v1.Image
}

var _ oci.Signatures = (*sigs)(nil)

// Get implements oci.Signatures
func (s *sigs) Get() ([]oci.Signature, error) {
	m, err := s.Manifest()
	if err != nil {
		return nil, err
	}
	signatures := make([]oci.Signature, 0, len(m.Layers))
	for _, desc := range m.Layers {
		layer, err := s.Image.LayerByDigest(desc.Digest)
		if err != nil {
			return nil, err
		}
		signatures = append(signatures, signature.New(layer, desc))
	}
	return signatures, nil
}
