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

package cosign

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/docker/distribution/reference"
	digest "github.com/opencontainers/go-digest"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	// repositorySubexp is the name for sub regex that maps to subject artifact digest in the url
	subArtDigestSubexp = "digest"
	// repositorySubexp is the name for sub regex that maps to repository name in the url
	repositorySubexp = "repository"
	cosignRe         = regexp.MustCompile(fmt.Sprintf(`^/v2/(?P<%s>%s)/manifests/%s-(?P<%s>%s).sig$`, repositorySubexp, reference.NameRegexp.String(), digest.SHA256, subArtDigestSubexp, reference.IdentifierRegexp))
	// the media type of consign signature layer
	mediaTypeCosignLayer = "application/vnd.dev.cosign.simplesigning.v1+json"
)

// SignatureMiddleware middleware to record the linkeage of artifact and its accessory
/* PUT /v2/library/hello-world/manifests/sha256-1b26826f602946860c279fce658f31050cff2c596583af237d971f4629b57792.sig
{
	"schemaVersion":2,
	"config":{
		"mediaType":"application/vnd.oci.image.config.v1+json",
		"size":233,
		"digest":"sha256:d4e6059ece7bea95266fd7766353130d4bf3dc21048b8a9783c98b8412618c38"
	},
	"layers":[
		{
			"mediaType":"application/vnd.dev.cosign.simplesigning.v1+json",
			"size":250,
			"digest":"sha256:91a821a0e2412f1b99b07bfe176451bcc343568b761388718abbf38076048564",
			"annotations":{
				"dev.cosignproject.cosign/signature":"MEUCIQD/imXjZJlcV82eXu9y9FJGgbDwVPw7AaGFzqva8G+CgwIgYc4CRvEjwoAwkzGoX+aZxQWCASpv5G+EAWDKOJRLbTQ="
			}
		}
	]
}
*/
// cosign adopt oci-spec 1.1 will have request and manifest like below
// It will skip this middleware since not using cosignRe for subject artifact reference
// use Subject Middleware indtead
/*
PUT /v2/library/goharbor/harbor-db/manifests/sha256:aabea2bdd5a6fb79c13837b88c7b158f4aa57a621194ee21959d0b520eda412f
{
  "schemaVersion": 2,
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "config": {
    "mediaType": "application/vnd.dev.cosign.artifact.sig.v1+json",
    "size": 233,
    "digest": "sha256:c025e9532dbc880534be96dbbb86a6bf63a272faced7f07bb8b4ceb45ca938d1"
  },
  "layers": [
    {
      "mediaType": "application/vnd.dev.cosign.simplesigning.v1+json",
      "size": 257,
      "digest": "sha256:38d07d81bf1d026da6420295113115d999ad6da90073b5e67147f978626423e6",
      "annotations": {
        "dev.cosignproject.cosign/signature": "MEUCIDOQc6I4MSd4/s8Bc8S7LXHCOnm4MGimpQdeCInLzM0VAiEAhWWYxmwEmYrFJ8xYNE3ow7PS4zeGe1R4RUbXRIawKJ4=",
        "dev.sigstore.cosign/bundle": "{\"SignedEntryTimestamp\":\"MEUCIC5DSFQx3nZhPFquF4NAdfetjqLR6qAa9i04cEtAg7VjAiEAzG2DUxqH+MdFSPih/EL/Vvsn3L1xCJUlOmRZeUYZaG0=\",\"Payload\":{\"body\":\"eyJhcGlWZXJzaW9uIjoiMC4wLjEiLCJraW5kIjoiaGFzaGVkcmVrb3JkIiwic3BlYyI6eyJkYXRhIjp7Imhhc2giOnsiYWxnb3JpdGhtIjoic2hhMjU2IiwidmFsdWUiOiIzOGQwN2Q4MWJmMWQwMjZkYTY0MjAyOTUxMTMxMTVkOTk5YWQ2ZGE5MDA3M2I1ZTY3MTQ3Zjk3ODYyNjQyM2U2In19LCJzaWduYXR1cmUiOnsiY29udGVudCI6Ik1FVUNJRE9RYzZJNE1TZDQvczhCYzhTN0xYSENPbm00TUdpbXBRZGVDSW5Mek0wVkFpRUFoV1dZeG13RW1ZckZKOHhZTkUzb3c3UFM0emVHZTFSNFJVYlhSSWF3S0o0PSIsInB1YmxpY0tleSI6eyJjb250ZW50IjoiTFMwdExTMUNSVWRKVGlCUVZVSk1TVU1nUzBWWkxTMHRMUzBLVFVacmQwVjNXVWhMYjFwSmVtb3dRMEZSV1VsTGIxcEplbW93UkVGUlkwUlJaMEZGWVVoSk1DOTZiWEpIYW1VNE9FeFVTM0ZDU2tvNWJXZDNhWEprWkFwaVJrZGpNQzlRYWtWUUwxbFJNelJwZFZweWJGVnRhMGx3ZDBocFdVTmxSV3M0YWpoWE5rSnBaV3BxTHk5WmVVRnZZaXN5VTFCTGRqUkJQVDBLTFMwdExTMUZUa1FnVUZWQ1RFbERJRXRGV1MwdExTMHRDZz09In19fX0=\",\"integratedTime\":1712651102,\"logIndex\":84313668,\"logID\":\"c0d23d6ad406973f9559f3ba2d1ca01f84147d8ffc5b8445c224f98b9591801d\"}}"
      }
    }
  ],
  "subject": {
    "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
    "size": 2621,
    "digest": "sha256:e50f88df1b11f94627e35bed9f34214392363508a2b07146d0a94516da97e4c0"
  }
}

*/
func SignatureMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusCreated {
			return nil
		}

		ctx := r.Context()
		logger := log.G(ctx).WithFields(log.Fields{"middleware": "cosign"})

		none := lib.ArtifactInfo{}
		info := lib.GetArtifactInfo(ctx)
		if info == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}
		if info.Tag == "" {
			return nil
		}

		// Needs tag to match the cosign tag pattern.
		_, subjectArtDigest, ok := matchCosignSignaturePattern(r.URL.Path)
		if !ok {
			return nil
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		contentType := r.Header.Get("Content-Type")
		manifest, desc, err := distribution.UnmarshalManifest(contentType, body)
		if err != nil {
			logger.Errorf("unmarshal manifest failed, error: %v", err)
			return err
		}

		var hasSignature bool
		for _, descriptor := range manifest.References() {
			if descriptor.MediaType == mediaTypeCosignLayer {
				hasSignature = true
				break
			}
		}

		if hasSignature {
			subjectArt, err := artifact.Ctl.GetByReference(ctx, info.Repository, fmt.Sprintf("%s:%s", digest.SHA256, subjectArtDigest), nil)
			if err != nil {
				if !errors.IsNotFoundErr(err) {
					logger.Errorf("failed to get subject artifact: %s, error: %v", subjectArtDigest, err)
					return err
				}
				log.Debug("the subject of the signature doesn't exist.")
			}
			art, err := artifact.Ctl.GetByReference(ctx, info.Repository, desc.Digest.String(), nil)
			if err != nil {
				logger.Errorf("failed to get cosign signature artifact: %s, error: %v", desc.Digest.String(), err)
				return err
			}
			accData := model.AccessoryData{
				ArtifactID:        art.ID,
				SubArtifactRepo:   info.Repository,
				SubArtifactDigest: fmt.Sprintf("%s:%s", digest.SHA256, subjectArtDigest),
				Size:              art.Size,
				Digest:            art.Digest,
				Type:              model.TypeCosignSignature,
			}
			if subjectArt != nil {
				accData.SubArtifactID = subjectArt.ID
			}
			if err := orm.WithTransaction(func(ctx context.Context) error {
				_, err := accessory.Mgr.Create(ctx, accData)
				return err
			})(orm.SetTransactionOpNameToContext(ctx, "tx-create-cosign-accessory")); err != nil {
				if !errors.IsConflictErr(err) {
					logger.Errorf("failed to create cosign signature artifact: %s, error: %v", desc.Digest.String(), err)
					return err
				}
			}
		}

		return nil
	})
}

// matchCosignSignaturePattern checks whether the provided path matches the blob upload URL pattern,
// if does, returns the repository as well
func matchCosignSignaturePattern(path string) (repository, digest string, match bool) {
	strs := cosignRe.FindStringSubmatch(path)
	if len(strs) < 3 {
		return "", "", false
	}
	return strs[1], strs[2], true
}
