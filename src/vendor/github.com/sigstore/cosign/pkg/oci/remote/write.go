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
	"fmt"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/sigstore/cosign/pkg/oci"
)

// WriteSignedImageIndexImages writes the images within the image index
// This includes the signed image and associated signatures in the image index
// TODO (priyawadhwa@): write the `index.json` itself to the repo as well
// TODO (priyawadhwa@): write the attestations
func WriteSignedImageIndexImages(ref name.Reference, sii oci.SignedImageIndex, opts ...Option) error {
	repo := ref.Context()
	o := makeOptions(repo, opts...)

	// write the image index if there is one
	ii, err := sii.SignedImageIndex(v1.Hash{})
	if err != nil {
		return fmt.Errorf("signed image index: %w", err)
	}
	if ii != nil {
		if err := remote.WriteIndex(ref, ii, o.ROpt...); err != nil {
			return fmt.Errorf("writing index: %w", err)
		}
	}

	// write the image if there is one
	si, err := sii.SignedImage(v1.Hash{})
	if err != nil {
		return fmt.Errorf("signed image: %w", err)
	}
	if si != nil {
		if err := remoteWrite(ref, si, o.ROpt...); err != nil {
			return fmt.Errorf("remote write: %w", err)
		}
	}

	// write the signatures
	sigs, err := sii.Signatures()
	if err != nil {
		return err
	}
	if sigs != nil { // will be nil if there are no associated signatures
		sigsTag, err := SignatureTag(ref, opts...)
		if err != nil {
			return fmt.Errorf("sigs tag: %w", err)
		}
		if err := remoteWrite(sigsTag, sigs, o.ROpt...); err != nil {
			return err
		}
	}

	// write the attestations
	atts, err := sii.Attestations()
	if err != nil {
		return err
	}
	if atts != nil { // will be nil if there are no associated attestations
		attsTag, err := AttestationTag(ref, opts...)
		if err != nil {
			return fmt.Errorf("sigs tag: %w", err)
		}
		return remoteWrite(attsTag, atts, o.ROpt...)
	}
	return nil
}

// WriteSignature publishes the signatures attached to the given entity
// into the provided repository.
func WriteSignatures(repo name.Repository, se oci.SignedEntity, opts ...Option) error {
	o := makeOptions(repo, opts...)

	// Access the signature list to publish
	sigs, err := se.Signatures()
	if err != nil {
		return err
	}

	// Determine the tag to which these signatures should be published.
	h, err := se.(digestable).Digest()
	if err != nil {
		return err
	}
	tag := o.TargetRepository.Tag(normalize(h, o.TagPrefix, o.SignatureSuffix))

	// Write the Signatures image to the tag, with the provided remote.Options
	return remoteWrite(tag, sigs, o.ROpt...)
}

// WriteAttestations publishes the attestations attached to the given entity
// into the provided repository.
func WriteAttestations(repo name.Repository, se oci.SignedEntity, opts ...Option) error {
	o := makeOptions(repo, opts...)

	// Access the signature list to publish
	atts, err := se.Attestations()
	if err != nil {
		return err
	}

	// Determine the tag to which these signatures should be published.
	h, err := se.(digestable).Digest()
	if err != nil {
		return err
	}
	tag := o.TargetRepository.Tag(normalize(h, o.TagPrefix, o.AttestationSuffix))

	// Write the Signatures image to the tag, with the provided remote.Options
	return remoteWrite(tag, atts, o.ROpt...)
}
