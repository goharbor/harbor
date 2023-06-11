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

// ImageIndex returns an oci.SignedImageIndex form of the v1.ImageIndex with
// no signatures.
func ImageIndex(i v1.ImageIndex) oci.SignedImageIndex {
	return &index{
		v1Index: i,
	}
}

type v1Index v1.ImageIndex

type index struct {
	v1Index
}

var _ oci.SignedImageIndex = (*index)(nil)

// SignedImage implements oci.SignedImageIndex
func (ii *index) SignedImage(h v1.Hash) (oci.SignedImage, error) {
	i, err := ii.Image(h)
	if err != nil {
		return nil, err
	}
	return Image(i), nil
}

// SignedImageIndex implements oci.SignedImageIndex
func (ii *index) SignedImageIndex(h v1.Hash) (oci.SignedImageIndex, error) {
	i, err := ii.ImageIndex(h)
	if err != nil {
		return nil, err
	}
	return ImageIndex(i), nil
}

// Signatures implements oci.SignedImageIndex
func (*index) Signatures() (oci.Signatures, error) {
	return empty.Signatures(), nil
}

// Attestations implements oci.SignedImageIndex
func (*index) Attestations() (oci.Signatures, error) {
	return empty.Signatures(), nil
}

// Attestations implements oci.SignedImage
func (*index) Attachment(name string) (oci.File, error) {
	return nil, errors.New("unimplemented")
}
