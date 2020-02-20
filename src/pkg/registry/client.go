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

package registry

import (
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/core/config"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/replication/util"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"net/http"
)

// TODO we'll merge all registry related code into this package before releasing 2.0

var (
	// Cli is the global registry client instance, it targets to the backend docker registry
	Cli = func() Client {
		url, _ := config.RegistryURL()
		username, password := config.RegistryCredential()
		return NewClient(url, true, username, password)
	}()

	accepts = []string{
		v1.MediaTypeImageIndex,
		manifestlist.MediaTypeManifestList,
		v1.MediaTypeImageManifest,
		schema2.MediaTypeManifest,
		schema1.MediaTypeSignedManifest,
	}
)

// Client defines the methods that a registry client should implements
type Client interface {
	// Copy the artifact from source repository to the destination. The "override"
	// is used to specify whether the destination artifact will be overridden if
	// its name is same with source but digest isn't
	Copy(srcRepo, srcRef, dstRepo, dstRef string, override bool) (err error)
	// TODO defines other methods
}

// NewClient creates a registry client based on the provided information
// TODO support HTTPS
func NewClient(url string, insecure bool, username, password string) Client {
	transport := util.GetHTTPTransport(insecure)
	authorizer := auth.NewAuthorizer(auth.NewBasicAuthCredential(username, password),
		&http.Client{
			Transport: transport,
		})
	return &client{
		url: url,
		client: &http.Client{
			Transport: registry.NewTransport(transport, authorizer),
		},
	}
}

type client struct {
	url    string
	client *http.Client
}

// TODO extend this method to support copy artifacts between different registries when merging codes
// TODO this can be used in replication to replace the existing implementation
// TODO add unit test case
func (c *client) Copy(srcRepo, srcRef, dstRepo, dstRef string, override bool) error {
	src, err := registry.NewRepository(srcRepo, c.url, c.client)
	if err != nil {
		return err
	}
	dst, err := registry.NewRepository(dstRepo, c.url, c.client)
	if err != nil {
		return err
	}
	// pull the manifest from the source repository
	srcDgt, mediaType, payload, err := src.PullManifest(srcRef, accepts)
	if err != nil {
		return err
	}

	// check the existence of the artifact on the destination repository
	dstDgt, exist, err := dst.ManifestExist(dstRef)
	if err != nil {
		return err
	}
	if exist {
		// the same artifact already exists
		if srcDgt == dstDgt {
			return nil
		}
		// the same name artifact exists, but not allowed to override
		if !override {
			return ierror.New(nil).WithCode(ierror.PreconditionCode).
				WithMessage("the same name but different digest artifact exists, but the override is set to false")
		}
	}

	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		return err
	}
	for _, descriptor := range manifest.References() {
		digest := descriptor.Digest.String()
		switch descriptor.MediaType {
		// skip foreign layer
		case schema2.MediaTypeForeignLayer:
			continue
		// manifest or index
		case v1.MediaTypeImageIndex, manifestlist.MediaTypeManifestList,
			v1.MediaTypeImageManifest, schema2.MediaTypeManifest,
			schema1.MediaTypeSignedManifest:
			if err = c.Copy(srcRepo, digest, dstRepo, digest, false); err != nil {
				return err
			}
		// common layer
		default:
			exist, err := dst.BlobExist(digest)
			if err != nil {
				return err
			}
			// the layer already exist, skip
			if exist {
				continue
			}
			// when the copy happens inside the same registry, use mount
			if err = dst.MountBlob(digest, srcRepo); err != nil {
				return err
			}
			/*
				// copy happens between different registries
				size, data, err := src.PullBlob(digest)
				if err != nil {
					return err
				}
				defer data.Close()
				if err = dst.PushBlob(digest, size, data); err != nil {
					return err
				}
			*/
		}
	}

	// push manifest to the destination repository
	if _, err = dst.PushManifest(dstRef, mediaType, payload); err != nil {
		return err
	}

	return nil
}
