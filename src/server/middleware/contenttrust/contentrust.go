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
	"net/http"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/util"
)

// ContentTrust handle docker pull content trust check
func ContentTrust() func(http.Handler) http.Handler {
	return middleware.BeforeRequest(func(r *http.Request) error {
		ctx := r.Context()

		none := lib.ArtifactInfo{}
		af := lib.GetArtifactInfo(ctx)
		if af == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}
		pro, err := project.Ctl.GetByName(ctx, af.ProjectName)
		if err != nil {
			return err
		}

		// If signature policy enabled, it has to at least have one signature.
		if pro.ContentTrustCosignEnabled() {
			if err := signatureChecking(ctx, r, af, pro.ProjectID, model.TypeCosignSignature); err != nil {
				if errors.IsErr(err, errors.PROJECTPOLICYVIOLATION) {
					return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage("The image is not signed by cosign.")
				}
				return err
			}
		}
		if pro.ContentTrustEnabled() {
			if err := signatureChecking(ctx, r, af, pro.ProjectID, model.TypeNotationSignature); err != nil {
				if errors.IsErr(err, errors.PROJECTPOLICYVIOLATION) {
					return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage("The image is not signed by notation.")
				}
				return err
			}
		}
		return nil
	})
}

func signatureChecking(ctx context.Context, r *http.Request, af lib.ArtifactInfo, projectID int64, signatureType string) error {
	logger := log.G(ctx)

	art, err := artifact.Ctl.GetByReference(ctx, af.Repository, af.Reference, &artifact.Option{
		WithAccessory: true,
	})
	if err != nil {
		return err
	}

	ok, err := util.SkipPolicyChecking(r, projectID, art.ID)
	if err != nil {
		return err
	}
	if ok {
		logger.Debugf("skip the checking of pulling artifact %s@%s", af.Repository, af.Digest)
		return nil
	}

	if len(art.Accessories) == 0 {
		return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION)
	}

	var hasSignature bool
	for _, acc := range art.Accessories {
		if acc.GetData().Type == signatureType {
			hasSignature = true
			break
		}
	}
	if !hasSignature {
		return errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION)
	}

	return nil
}
