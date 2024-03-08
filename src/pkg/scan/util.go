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

package scan

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/static"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/controller/robot"
	v1sq "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

type Insecure bool

func (i Insecure) RemoteOptions() []remote.Option {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: bool(i)}
	return []remote.Option{remote.WithTransport(tr)}
}

type referrer struct {
	Insecure
}

// GenAccessoryArt composes the accessory oci object and push it back to harbor core as an accessory of the scanned artifact.
func GenAccessoryArt(sq v1sq.ScanRequest, accData []byte, accAnnotations map[string]string, mediaType string, robot robot.Robot) (string, error) {
	accArt, err := mutate.Append(empty.Image, mutate.Addendum{
		Layer: static.NewLayer(accData, ocispec.MediaTypeImageLayer),
		History: v1.History{
			Author:    "harbor",
			CreatedBy: "harbor",
			Created:   v1.Time{}, // static
		},
	})
	if err != nil {
		return "", err
	}

	dg, err := digest.Parse(sq.Artifact.Digest)
	if err != nil {
		return "", err
	}
	accSubArt := &v1.Descriptor{
		MediaType: types.MediaType(sq.Artifact.MimeType),
		Size:      sq.Artifact.Size,
		Digest: v1.Hash{
			Algorithm: dg.Algorithm().String(),
			Hex:       dg.Hex(),
		},
	}
	// TODO to leverage the artifactType of distribution spec v1.1 to specify the sbom type.
	// https://github.com/google/go-containerregistry/issues/1832
	accArt = mutate.MediaType(accArt, ocispec.MediaTypeImageManifest)
	accArt = mutate.ConfigMediaType(accArt, types.MediaType(mediaType))
	accArt = mutate.Subject(accArt, *accSubArt).(v1.Image)
	accArt = mutate.Annotations(accArt, accAnnotations).(v1.Image)

	digest, err := accArt.Digest()
	if err != nil {
		return "", err
	}
	accRef, err := name.ParseReference(fmt.Sprintf("%s/%s@%s", sq.Registry.URL, sq.Artifact.Repository, digest.String()))
	if err != nil {
		return "", err
	}
	opts := append(referrer{Insecure: true}.RemoteOptions(), remote.WithAuth(&authn.Basic{Username: robot.Name, Password: robot.Secret}))
	if err := remote.Write(accRef, accArt, opts...); err != nil {
		return "", err
	}
	return digest.String(), nil
}
