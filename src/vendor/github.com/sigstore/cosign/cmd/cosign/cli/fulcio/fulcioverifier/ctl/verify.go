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

package ctl

import (
	"context"
	"crypto"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	ct "github.com/google/certificate-transparency-go"
	ctx509 "github.com/google/certificate-transparency-go/x509"
	"github.com/google/certificate-transparency-go/x509util"
	"github.com/sigstore/cosign/cmd/cosign/cli/fulcio/fulcioverifier/ctutil"

	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/tuf"
)

// This is the CT log public key target name
var ctPublicKeyStr = `ctfe.pub`

// Setting this env variable will over ride what is used to validate
// the SCT coming back from Fulcio.
const altCTLogPublicKeyLocation = "SIGSTORE_CT_LOG_PUBLIC_KEY_FILE"

// logIDMetadata holds information for mapping a key ID hash (log ID) to associated data.
type logIDMetadata struct {
	pubKey crypto.PublicKey
	status tuf.StatusKind
}

// ContainsSCT checks if the certificate contains embedded SCTs. cert can either be
// DER or PEM encoded.
func ContainsSCT(cert []byte) (bool, error) {
	embeddedSCTs, err := x509util.ParseSCTsFromCertificate(cert)
	if err != nil {
		return false, err
	}
	if len(embeddedSCTs) != 0 {
		return true, nil
	}
	return false, nil
}

// VerifySCT verifies SCTs against the Fulcio CT log public key.
//
// The SCT is a `Signed Certificate Timestamp`, which promises that
// the certificate issued by Fulcio was also added to the public CT log within
// some defined time period.
//
// VerifySCT can verify an SCT list embedded in the certificate, or a detached
// SCT provided by Fulcio.
//
// By default the public keys comes from TUF, but you can override this for test
// purposes by using an env variable `SIGSTORE_CT_LOG_PUBLIC_KEY_FILE`. If using
// an alternate, the file can be PEM, or DER format.
func VerifySCT(ctx context.Context, certPEM, chainPEM, rawSCT []byte) error {
	// fetch SCT verification key
	pubKeys := make(map[[sha256.Size]byte]logIDMetadata)
	rootEnv := os.Getenv(altCTLogPublicKeyLocation)
	if rootEnv == "" {
		tufClient, err := tuf.NewFromEnv(ctx)
		if err != nil {
			return err
		}

		targets, err := tufClient.GetTargetsByMeta(tuf.CTFE, []string{ctPublicKeyStr})
		if err != nil {
			return err
		}
		for _, t := range targets {
			pub, err := cryptoutils.UnmarshalPEMToPublicKey(t.Target)
			if err != nil {
				return err
			}
			keyID, err := ctutil.GetCTLogID(pub)
			if err != nil {
				return fmt.Errorf("error getting CTFE public key hash")
			}
			pubKeys[keyID] = logIDMetadata{pub, t.Status}
		}
	} else {
		fmt.Fprintf(os.Stderr, "**Warning** Using a non-standard public key for verifying SCT: %s\n", rootEnv)
		raw, err := os.ReadFile(rootEnv)
		if err != nil {
			return fmt.Errorf("error reading alternate public key file")
		}
		pubKey, err := cryptoutils.UnmarshalPEMToPublicKey(raw)
		if err != nil {
			return fmt.Errorf("error parsing alternate public key from the file")
		}
		keyID, err := ctutil.GetCTLogID(pubKey)
		if err != nil {
			return fmt.Errorf("error getting CTFE public key hash")
		}
		pubKeys[keyID] = logIDMetadata{pubKey, tuf.Active}
	}
	if len(pubKeys) == 0 {
		return errors.New("none of the CTFE keys have been found")
	}

	// parse certificate and chain
	cert, err := x509util.CertificateFromPEM(certPEM)
	if err != nil {
		return err
	}
	certChain, err := x509util.CertificatesFromPEM(chainPEM)
	if err != nil {
		return err
	}
	if len(certChain) == 0 {
		return errors.New("no certificate chain found")
	}

	// fetch embedded SCT if present
	embeddedSCTs, err := x509util.ParseSCTsFromCertificate(certPEM)
	if err != nil {
		return err
	}
	// SCT must be either embedded or in header
	if len(embeddedSCTs) == 0 && len(rawSCT) == 0 {
		return errors.New("no SCT found")
	}

	// check SCT embedded in certificate
	if len(embeddedSCTs) != 0 {
		for _, sct := range embeddedSCTs {
			pubKeyMetadata, ok := pubKeys[sct.LogID.KeyID]
			if !ok {
				return errors.New("ctfe public key not found for embedded SCT")
			}
			err := ctutil.VerifySCT(pubKeyMetadata.pubKey, []*ctx509.Certificate{cert, certChain[0]}, sct, true)
			if err != nil {
				return fmt.Errorf("error verifying embedded SCT")
			}
			if pubKeyMetadata.status != tuf.Active {
				fmt.Fprintf(os.Stderr, "**Info** Successfully verified embedded SCT using an expired verification key\n")
			}
		}
		return nil
	}

	// check SCT in response header
	var addChainResp ct.AddChainResponse
	if err := json.Unmarshal(rawSCT, &addChainResp); err != nil {
		return fmt.Errorf("unmarshal")
	}
	sct, err := addChainResp.ToSignedCertificateTimestamp()
	if err != nil {
		return err
	}
	pubKeyMetadata, ok := pubKeys[sct.LogID.KeyID]
	if !ok {
		return errors.New("ctfe public key not found")
	}
	err = ctutil.VerifySCT(pubKeyMetadata.pubKey, []*ctx509.Certificate{cert}, sct, false)
	if err != nil {
		return fmt.Errorf("error verifying SCT")
	}
	if pubKeyMetadata.status != tuf.Active {
		fmt.Fprintf(os.Stderr, "**Info** Successfully verified SCT using an expired verification key\n")
	}
	return nil
}

// VerifyEmbeddedSCT verifies an embedded SCT in a certificate.
func VerifyEmbeddedSCT(ctx context.Context, chain []*x509.Certificate) error {
	if len(chain) < 2 {
		return errors.New("certificate chain must contain at least a certificate and its issuer")
	}
	certPEM, err := cryptoutils.MarshalCertificateToPEM(chain[0])
	if err != nil {
		return err
	}
	chainPEM, err := cryptoutils.MarshalCertificatesToPEM(chain[1:])
	if err != nil {
		return err
	}
	return VerifySCT(ctx, certPEM, chainPEM, []byte{})
}
