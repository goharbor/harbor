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

package static

import (
	"bytes"
	"crypto/x509"
	"io"

	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/sigstore/cosign/pkg/cosign/bundle"
	"github.com/sigstore/cosign/pkg/oci"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
)

const (
	SignatureAnnotationKey   = "dev.cosignproject.cosign/signature"
	CertificateAnnotationKey = "dev.sigstore.cosign/certificate"
	ChainAnnotationKey       = "dev.sigstore.cosign/chain"
	BundleAnnotationKey      = "dev.sigstore.cosign/bundle"
)

// NewSignature constructs a new oci.Signature from the provided options.
func NewSignature(payload []byte, b64sig string, opts ...Option) (oci.Signature, error) {
	o, err := makeOptions(opts...)
	if err != nil {
		return nil, err
	}
	return &staticLayer{
		b:      payload,
		b64sig: b64sig,
		opts:   o,
	}, nil
}

// NewAttestation constructs a new oci.Signature from the provided options.
// Since Attestation is treated just like a Signature but the actual signature
// is baked into the payload, the Signature does not actually have
// the Base64Signature.
func NewAttestation(payload []byte, opts ...Option) (oci.Signature, error) {
	return NewSignature(payload, "", opts...)
}

// Copy constructs a new oci.Signature from the provided one.
func Copy(sig oci.Signature) (oci.Signature, error) {
	payload, err := sig.Payload()
	if err != nil {
		return nil, err
	}
	b64sig, err := sig.Base64Signature()
	if err != nil {
		return nil, err
	}
	var opts []Option

	mt, err := sig.MediaType()
	if err != nil {
		return nil, err
	}
	opts = append(opts, WithLayerMediaType(mt))

	ann, err := sig.Annotations()
	if err != nil {
		return nil, err
	}
	opts = append(opts, WithAnnotations(ann))

	bundle, err := sig.Bundle()
	if err != nil {
		return nil, err
	}
	opts = append(opts, WithBundle(bundle))

	cert, err := sig.Cert()
	if err != nil {
		return nil, err
	}
	if cert != nil {
		rawCert, err := cryptoutils.MarshalCertificateToPEM(cert)
		if err != nil {
			return nil, err
		}
		chain, err := sig.Chain()
		if err != nil {
			return nil, err
		}
		rawChain, err := cryptoutils.MarshalCertificatesToPEM(chain)
		if err != nil {
			return nil, err
		}
		opts = append(opts, WithCertChain(rawCert, rawChain))
	}
	return NewSignature(payload, b64sig, opts...)
}

type staticLayer struct {
	b      []byte
	b64sig string
	opts   *options
}

var _ v1.Layer = (*staticLayer)(nil)
var _ oci.Signature = (*staticLayer)(nil)

// Annotations implements oci.Signature
func (l *staticLayer) Annotations() (map[string]string, error) {
	m := make(map[string]string, len(l.opts.Annotations)+1)
	for k, v := range l.opts.Annotations {
		m[k] = v
	}
	m[SignatureAnnotationKey] = l.b64sig
	return m, nil
}

// Payload implements oci.Signature
func (l *staticLayer) Payload() ([]byte, error) {
	return l.b, nil
}

// Base64Signature implements oci.Signature
func (l *staticLayer) Base64Signature() (string, error) {
	return l.b64sig, nil
}

// Cert implements oci.Signature
func (l *staticLayer) Cert() (*x509.Certificate, error) {
	certs, err := cryptoutils.LoadCertificatesFromPEM(bytes.NewReader(l.opts.Cert))
	if err != nil {
		return nil, err
	}
	if len(certs) == 0 {
		return nil, nil
	}
	return certs[0], nil
}

// Chain implements oci.Signature
func (l *staticLayer) Chain() ([]*x509.Certificate, error) {
	certs, err := cryptoutils.LoadCertificatesFromPEM(bytes.NewReader(l.opts.Chain))
	if err != nil {
		return nil, err
	}
	return certs, nil
}

// Bundle implements oci.Signature
func (l *staticLayer) Bundle() (*bundle.RekorBundle, error) {
	return l.opts.Bundle, nil
}

// Digest implements v1.Layer
func (l *staticLayer) Digest() (v1.Hash, error) {
	h, _, err := v1.SHA256(bytes.NewReader(l.b))
	return h, err
}

// DiffID implements v1.Layer
func (l *staticLayer) DiffID() (v1.Hash, error) {
	h, _, err := v1.SHA256(bytes.NewReader(l.b))
	return h, err
}

// Compressed implements v1.Layer
func (l *staticLayer) Compressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(l.b)), nil
}

// Uncompressed implements v1.Layer
func (l *staticLayer) Uncompressed() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(l.b)), nil
}

// Size implements v1.Layer
func (l *staticLayer) Size() (int64, error) {
	return int64(len(l.b)), nil
}

// MediaType implements v1.Layer
func (l *staticLayer) MediaType() (types.MediaType, error) {
	return l.opts.LayerMediaType, nil
}
