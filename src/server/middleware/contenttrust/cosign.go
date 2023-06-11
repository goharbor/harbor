// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package contenttrust

import (
	"context"
	"crypto"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/util"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"

	"github.com/sigstore/cosign/pkg/cosign"
	ociremote "github.com/sigstore/cosign/pkg/oci/remote"
	"github.com/sigstore/sigstore/pkg/cryptoutils"
	"github.com/sigstore/sigstore/pkg/signature"
)

var (
	cosignStrictVerificationEnabled = "COSIGN_STRICT_VERIFICATION_ENABLED"
	cosignCAPath                    = "/etc/core/cosign/ca.crt"
	cosignKeyPath                   = "/etc/core/cosign/key.pub"
)

// Cosign handle docker pull content trust check
func Cosign() func(http.Handler) http.Handler {
	return middleware.BeforeRequest(func(r *http.Request) error {
		ctx := r.Context()

		logger := log.G(ctx)

		none := lib.ArtifactInfo{}
		af := lib.GetArtifactInfo(ctx)
		if af == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}
		pro, err := project.Ctl.GetByName(ctx, af.ProjectName)
		if err != nil {
			return err
		}

		// If cosign policy enabled, it has to at least have one cosign signature.
		if pro.ContentTrustCosignEnabled() {
			art, err := artifact.Ctl.GetByReference(ctx, af.Repository, af.Reference, &artifact.Option{
				WithAccessory: true,
			})
			if err != nil {
				return err
			}

			ok, err := util.SkipPolicyChecking(r, pro.ProjectID, art.ID)
			if err != nil {
				return err
			}
			if ok {
				logger.Debugf("artifact %s@%s is pulling by the scanner/cosign, skip the checking", af.Repository, af.Digest)
				return nil
			}

			if len(art.Accessories) == 0 {
				pkgE := errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage("The image is not signed in Cosign.")
				return pkgE
			}

			var hasCosignSignature bool
			for _, acc := range art.Accessories {
				if acc.GetData().Type == model.TypeCosignSignature {
					hasCosignSignature = true
					break
				}
			}
			if !hasCosignSignature {
				pkgE := errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage("The image is not signed in Cosign.")
				return pkgE
			}

			// If cosign strict verification enabled, it has to verify the cosign signatures.
			if os.Getenv(cosignStrictVerificationEnabled) == "true" {
				u, _ := config.RegistryURL()
				url, err := url.Parse(u)
				if err != nil || url.Host == "" {
					return err
				}
				artPath := fmt.Sprintf("%s@%s", art.RepositoryName, art.Digest)
				err = verifyArtifactSignatures(ctx, url.Host, artPath)
				if err != nil {
					logger.Errorf("failed to verify image signature %s with error %v", artPath, err)
					pkgE := errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage("The image doesn't pass Cosign signature verification.")
					return pkgE
				} else {
					logger.Infof("succeeded to verify image singature %s", artPath)
				}
			}
		}

		return nil
	})
}

func verifyArtifactSignatures(ctx context.Context, registry, signedArtPath string) error {
	signedArtRef, err := name.ParseReference(signedArtPath, name.WithDefaultRegistry(registry))
	if err != nil {
		return err
	}

	co := &cosign.CheckOpts{}

	// Parse cosign CA root certs and pass it to cosign check options if exists.
	roots, err := readRootCertsFromFile(cosignCAPath)
	if err != nil {
		return errors.New(err).WithCode(errors.GeneralCode).WithMessage("fail to read root certs from file")
	}
	if roots != nil {
		co.RootCerts = roots
	}

	// Parse cosign public key and pass it to cosign check options if exists.
	pubKey, err := readPubKeyFromFile(cosignKeyPath)
	if err != nil {
		return errors.New(err).WithCode(errors.GeneralCode).WithMessage("fail to read public key from file %s", cosignKeyPath)
	}
	if pubKey != nil {
		co.SigVerifier, err = signature.LoadVerifier(pubKey, crypto.SHA256)
		if err != nil {
			return err
		}
	}

	if roots == nil && pubKey == nil {
		return errors.New(nil).WithCode(errors.GeneralCode).WithMessage("neither root certs nor public key exist for verification")
	}

	// Set basic auth in cosign check options.
	username, password := config.RegistryCredential()
	opts := []remote.Option{
		remote.WithAuth(&authn.Basic{
			Username: username,
			Password: password,
		}),
		remote.WithTransport(commonhttp.GetHTTPTransport()),
	}
	co.RegistryClientOpts = []ociremote.Option{
		ociremote.WithRemoteOptions(opts...),
	}

	// Verify signatures.
	_, _, err = cosign.VerifyImageSignatures(ctx, signedArtRef, co)
	return err
}

func readRootCertsFromFile(path string) (*x509.CertPool, error) {
	roots := x509.NewCertPool()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	ok := roots.AppendCertsFromPEM(data)
	if !ok {
		return nil, errors.New("fail to parse cert")
	}
	return roots, nil
}

func readPubKeyFromFile(path string) (crypto.PublicKey, error) {
	keyBytes, err := ioutil.ReadFile(cosignKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	pubKey, err := cryptoutils.UnmarshalPEMToPublicKey(keyBytes)
	if err != nil {
		return nil, err
	}
	return pubKey, nil
}
