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
		}

		return nil
	})
}
