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

package signature

import (
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/sigstore/cosign/pkg/cosign/bundle"
	"github.com/sigstore/cosign/pkg/oci"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
)

const (
	sigkey    = "dev.cosignproject.cosign/signature"
	certkey   = "dev.sigstore.cosign/certificate"
	chainkey  = "dev.sigstore.cosign/chain"
	BundleKey = "dev.sigstore.cosign/bundle"
)

type sigLayer struct {
	v1.Layer
	desc v1.Descriptor
}

func New(l v1.Layer, desc v1.Descriptor) oci.Signature {
	return &sigLayer{
		Layer: l,
		desc:  desc,
	}
}

var _ oci.Signature = (*sigLayer)(nil)

// Annotations implements oci.Signature
func (s *sigLayer) Annotations() (map[string]string, error) {
	return s.desc.Annotations, nil
}

// Payload implements oci.Signature
func (s *sigLayer) Payload() ([]byte, error) {
	// Compressed is a misnomer here, we just want the raw bytes from the registry.
	r, err := s.Layer.Compressed()
	if err != nil {
		return nil, err
	}
	payload, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return payload, nil
}

// Base64Signature implements oci.Signature
func (s *sigLayer) Base64Signature() (string, error) {
	b64sig, ok := s.desc.Annotations[sigkey]
	if !ok {
		return "", fmt.Errorf("signature layer %s is missing %q annotation", s.desc.Digest, sigkey)
	}
	return b64sig, nil
}

// Cert implements oci.Signature
func (s *sigLayer) Cert() (*x509.Certificate, error) {
	certPEM := s.desc.Annotations[certkey]
	if certPEM == "" {
		return nil, nil
	}
	certs, err := cryptoutils.LoadCertificatesFromPEM(strings.NewReader(certPEM))
	if err != nil {
		return nil, err
	}
	return certs[0], nil
}

// Chain implements oci.Signature
func (s *sigLayer) Chain() ([]*x509.Certificate, error) {
	chainPEM := s.desc.Annotations[chainkey]
	if chainPEM == "" {
		return nil, nil
	}
	certs, err := cryptoutils.LoadCertificatesFromPEM(strings.NewReader(chainPEM))
	if err != nil {
		return nil, err
	}
	return certs, nil
}

// Bundle implements oci.Signature
func (s *sigLayer) Bundle() (*bundle.RekorBundle, error) {
	val := s.desc.Annotations[BundleKey]
	if val == "" {
		return nil, nil
	}
	var b bundle.RekorBundle
	if err := json.Unmarshal([]byte(val), &b); err != nil {
		return nil, fmt.Errorf("unmarshaling bundle: %w", err)
	}
	return &b, nil
}
