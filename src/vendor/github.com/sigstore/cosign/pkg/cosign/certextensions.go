//
// Copyright 2022 The Sigstore Authors.
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

package cosign

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
)

type CertExtensions struct {
	Cert *x509.Certificate
}

var (
	// Fulcio cert-extensions, documented here: https://github.com/sigstore/fulcio/blob/main/docs/oid-info.md
	CertExtensionOIDCIssuer               = "1.3.6.1.4.1.57264.1.1"
	CertExtensionGithubWorkflowTrigger    = "1.3.6.1.4.1.57264.1.2"
	CertExtensionGithubWorkflowSha        = "1.3.6.1.4.1.57264.1.3"
	CertExtensionGithubWorkflowName       = "1.3.6.1.4.1.57264.1.4"
	CertExtensionGithubWorkflowRepository = "1.3.6.1.4.1.57264.1.5"
	CertExtensionGithubWorkflowRef        = "1.3.6.1.4.1.57264.1.6"

	OIDOtherName = asn1.ObjectIdentifier{1, 3, 6, 1, 4, 1, 57264, 1, 7}

	CertExtensionMap = map[string]string{
		CertExtensionOIDCIssuer:               "oidcIssuer",
		CertExtensionGithubWorkflowTrigger:    "githubWorkflowTrigger",
		CertExtensionGithubWorkflowSha:        "githubWorkflowSha",
		CertExtensionGithubWorkflowName:       "githubWorkflowName",
		CertExtensionGithubWorkflowRepository: "githubWorkflowRepository",
		CertExtensionGithubWorkflowRef:        "githubWorkflowRef",
	}

	// OID for Subject Alternative Name
	SANOID = asn1.ObjectIdentifier{2, 5, 29, 17}
)

func (ce *CertExtensions) certExtensions() map[string]string {
	extensions := map[string]string{}
	for _, ext := range ce.Cert.Extensions {
		readableName, ok := CertExtensionMap[ext.Id.String()]
		if ok {
			extensions[readableName] = string(ext.Value)
		} else {
			extensions[ext.Id.String()] = string(ext.Value)
		}
	}
	return extensions
}

// GetIssuer returns the issuer for a Certificate
func (ce *CertExtensions) GetIssuer() string {
	return ce.certExtensions()["oidcIssuer"]
}

// GetCertExtensionGithubWorkflowTrigger returns the GitHub Workflow Trigger for a Certificate
func (ce *CertExtensions) GetCertExtensionGithubWorkflowTrigger() string {
	return ce.certExtensions()["githubWorkflowTrigger"]
}

// GetExtensionGithubWorkflowSha returns the GitHub Workflow SHA for a Certificate
func (ce *CertExtensions) GetExtensionGithubWorkflowSha() string {
	return ce.certExtensions()["githubWorkflowSha"]
}

// GetCertExtensionGithubWorkflowName returns the GitHub Workflow Name for a Certificate
func (ce *CertExtensions) GetCertExtensionGithubWorkflowName() string {
	return ce.certExtensions()["githubWorkflowName"]
}

// GetCertExtensionGithubWorkflowRepository returns the GitHub Workflow Repository for a Certificate
func (ce *CertExtensions) GetCertExtensionGithubWorkflowRepository() string {
	return ce.certExtensions()["githubWorkflowRepository"]
}

// GetCertExtensionGithubWorkflowRef returns the GitHub Workflow Ref for a Certificate
func (ce *CertExtensions) GetCertExtensionGithubWorkflowRef() string {
	return ce.certExtensions()["githubWorkflowRef"]
}

// TODO: Move (un)marshalling to sigstore/sigstore
// OtherName describes a name related to a certificate which is not in one
// of the standard name formats. RFC 5280, 4.2.1.6:
//
//	OtherName ::= SEQUENCE {
//	     type-id    OBJECT IDENTIFIER,
//	     value      [0] EXPLICIT ANY DEFINED BY type-id }
//
// OtherName for Fulcio-issued certificates only supports UTF-8 strings as values.
type OtherName struct {
	ID    asn1.ObjectIdentifier
	Value string `asn1:"utf8,explicit,tag:0"`
}

// MarshalOtherNameSAN creates a Subject Alternative Name extension
// with an OtherName sequence. RFC 5280, 4.2.1.6:
//
// SubjectAltName ::= GeneralNames
// GeneralNames ::= SEQUENCE SIZE (1..MAX) OF GeneralName
// GeneralName ::= CHOICE {
//
//	otherName                       [0]     OtherName,
//	... }
func MarshalOtherNameSAN(name string, critical bool) (*pkix.Extension, error) {
	o := OtherName{
		ID:    OIDOtherName,
		Value: name,
	}
	bytes, err := asn1.MarshalWithParams(o, "tag:0")
	if err != nil {
		return nil, err
	}

	sans, err := asn1.Marshal([]asn1.RawValue{{FullBytes: bytes}})
	if err != nil {
		return nil, err
	}
	return &pkix.Extension{
		Id:       SANOID,
		Critical: critical,
		Value:    sans,
	}, nil
}

// UnmarshalOtherNameSAN extracts a UTF-8 string from the OtherName
// field in the Subject Alternative Name extension.
func UnmarshalOtherNameSAN(exts []pkix.Extension) (string, error) {
	var otherNames []string

	for _, e := range exts {
		if !e.Id.Equal(SANOID) {
			continue
		}

		var seq asn1.RawValue
		rest, err := asn1.Unmarshal(e.Value, &seq)
		if err != nil {
			return "", err
		} else if len(rest) != 0 {
			return "", fmt.Errorf("trailing data after X.509 extension")
		}
		if !seq.IsCompound || seq.Tag != asn1.TagSequence || seq.Class != asn1.ClassUniversal {
			return "", asn1.StructuralError{Msg: "bad SAN sequence"}
		}

		rest = seq.Bytes
		for len(rest) > 0 {
			var v asn1.RawValue
			rest, err = asn1.Unmarshal(rest, &v)
			if err != nil {
				return "", err
			}

			// skip all GeneralName fields except OtherName
			if v.Tag != 0 {
				continue
			}

			var other OtherName
			_, err := asn1.UnmarshalWithParams(v.FullBytes, &other, "tag:0")
			if err != nil {
				return "", fmt.Errorf("could not parse requested OtherName SAN: %w", err)
			}
			if !other.ID.Equal(OIDOtherName) {
				return "", fmt.Errorf("unexpected OID for OtherName, expected %v, got %v", OIDOtherName, other.ID)
			}
			otherNames = append(otherNames, other.Value)
		}
	}

	if len(otherNames) == 0 {
		return "", errors.New("no OtherName found")
	}
	if len(otherNames) != 1 {
		return "", errors.New("expected only one OtherName")
	}

	return otherNames[0], nil
}
