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

package signed

import (
	"errors"

	v1 "github.com/google/go-containerregistry/pkg/v1"

	"github.com/sigstore/cosign/pkg/oci"
	"github.com/sigstore/cosign/pkg/oci/empty"
)

// Image returns an oci.SignedImage form of the v1.Image with no signatures.
func Image(i v1.Image) oci.SignedImage {
	return &image{
		Image: i,
	}
}

type image struct {
	v1.Image
}

var _ oci.SignedImage = (*image)(nil)

// Signatures implements oci.SignedImage
func (*image) Signatures() (oci.Signatures, error) {
	return empty.Signatures(), nil
}

// Attestations implements oci.SignedImage
func (*image) Attestations() (oci.Signatures, error) {
	return empty.Signatures(), nil
}

// Attestations implements oci.SignedImage
func (*image) Attachment(name string) (oci.File, error) {
	return nil, errors.New("unimplemented")
}
