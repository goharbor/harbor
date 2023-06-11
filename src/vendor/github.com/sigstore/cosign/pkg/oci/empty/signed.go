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
	"errors"
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/sigstore/cosign/pkg/oci"
)

type signedImage struct {
	v1.Image
	digest       v1.Hash
	signature    oci.Signatures
	attestations oci.Signatures
}

func (se *signedImage) Signatures() (oci.Signatures, error) {
	return se.signature, nil
}

func (se *signedImage) Attestations() (oci.Signatures, error) {
	return se.attestations, nil
}

func (se *signedImage) Attachment(name string) (oci.File, error) {
	return nil, errors.New("no attachments")
}

func (se *signedImage) Digest() (v1.Hash, error) {
	if se.digest.Hex == "" {
		return v1.Hash{}, fmt.Errorf("digest not available")
	}
	return se.digest, nil
}

func SignedImage(ref name.Reference) (oci.SignedImage, error) {
	var err error
	d := v1.Hash{}
	base := empty.Image
	if digest, ok := ref.(name.Digest); ok {
		d, err = v1.NewHash(digest.DigestStr())
		if err != nil {
			return nil, err
		}
	}
	return &signedImage{
		Image:        base,
		digest:       d,
		signature:    Signatures(),
		attestations: Signatures(),
	}, nil
}
