//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxy

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

const (
	// wait more time than manifest (maxManifestWait) because manifest list depends on manifest ready
	maxManifestListWait = 20
	maxManifestWait     = 10
	sleepIntervalSec    = 20
)

var (
	// Ctl is a global proxy controller instance
	ctl  Controller
	once sync.Once
)

// Controller defines the operations related with pull through proxy
type Controller interface {
	// UseLocalBlob check if the blob should use local copy
	UseLocalBlob(ctx context.Context, art lib.ArtifactInfo) bool
	// UseLocalManifest check manifest should use local copy
	UseLocalManifest(ctx context.Context, art lib.ArtifactInfo, remote RemoteInterface) (bool, error)
	// ProxyBlob proxy the blob request to the remote server, p is the proxy project
	// art is the ArtifactInfo which includes the digest of the blob
	ProxyBlob(ctx context.Context, p *models.Project, art lib.ArtifactInfo) (int64, io.ReadCloser, error)
	// ProxyManifest proxy the manifest request to the remote server, p is the proxy project,
	// art is the ArtifactInfo which includes the tag or digest of the manifest
	ProxyManifest(ctx context.Context, p *models.Project, art lib.ArtifactInfo, remote RemoteInterface) (distribution.Manifest, error)
}
type controller struct {
	blobCtl     blob.Controller
	artifactCtl artifact.Controller
	local       localInterface
}

// ControllerInstance -- Get the proxy controller instance
func ControllerInstance() Controller {
	// Lazy load the controller
	// Because LocalHelper is not ready unless core startup completely
	once.Do(func() {
		ctl = &controller{
			blobCtl:     blob.Ctl,
			artifactCtl: artifact.Ctl,
			local:       newLocalHelper(),
		}
	})

	return ctl
}

func (c *controller) UseLocalBlob(ctx context.Context, art lib.ArtifactInfo) bool {
	if len(art.Digest) == 0 {
		return false
	}
	exist, err := c.local.BlobExist(ctx, art)
	if err != nil {
		return false
	}
	return exist
}

func (c *controller) UseLocalManifest(ctx context.Context, art lib.ArtifactInfo, remote RemoteInterface) (bool, error) {
	a, err := c.local.GetManifest(ctx, art)
	if err != nil {
		return false, err
	}
	if a == nil {
		return false, nil
	}
	// Pull by digest
	if len(art.Digest) > 0 {
		return true, nil
	}
	// Pull by tag
	remoteRepo := getRemoteRepo(art)
	exist, dig, err := remote.ManifestExist(remoteRepo, art.Tag) // HEAD
	if err != nil {
		return false, err
	}
	if !exist {
		go func() {
			c.local.DeleteManifest(remoteRepo, art.Tag)
		}()
		return false, errors.NotFoundError(fmt.Errorf("repo %v, tag %v not found", art.Repository, art.Tag))
	}
	return dig == a.Digest, nil // digest matches
}

func (c *controller) ProxyManifest(ctx context.Context, p *models.Project, art lib.ArtifactInfo, remote RemoteInterface) (distribution.Manifest, error) {
	var man distribution.Manifest
	remoteRepo := getRemoteRepo(art)
	ref := getReference(art)
	man, dig, err := remote.Manifest(remoteRepo, ref)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			go func() {
				c.local.DeleteManifest(remoteRepo, art.Tag)
			}()
		}
		return man, err
	}
	ct, _, err := man.Payload()
	if err != nil {
		return man, err
	}

	// Push manifest in background
	go func(operator string) {
		bCtx := orm.Context()
		a, err := c.local.GetManifest(bCtx, art)
		if err != nil {
			log.Errorf("failed to get manifest, error %v", err)
		}
		// Push manifest to local when pull with digest, or artifact not found, or digest mismatch
		if len(art.Tag) == 0 || a == nil || a.Digest != dig {
			// pull with digest
			c.waitAndPushManifest(ctx, remoteRepo, man, art, ct, remote)
		}

		// Query artifact after push
		if a == nil {
			a, err = c.local.GetManifest(ctx, art)
			if err != nil {
				log.Errorf("failed to get manifest, error %v", err)
			}
		}
		if a != nil {
			SendPullEvent(a, art.Tag, operator)
		}
	}(operator.FromContext(ctx))

	return man, nil
}

func (c *controller) ProxyBlob(ctx context.Context, p *models.Project, art lib.ArtifactInfo) (int64, io.ReadCloser, error) {
	remoteRepo := getRemoteRepo(art)
	log.Debugf("The blob doesn't exist, proxy the request to the target server, url:%v", remoteRepo)
	rHelper, err := NewRemoteHelper(p.RegistryID)
	if err != nil {
		return 0, nil, err
	}

	size, bReader, err := rHelper.BlobReader(remoteRepo, art.Digest)
	if err != nil {
		log.Errorf("failed to pull blob, error %v", err)
		return 0, nil, err
	}
	desc := distribution.Descriptor{Size: size, Digest: digest.Digest(art.Digest)}
	go func() {
		err := c.putBlobToLocal(remoteRepo, art.Repository, desc, rHelper)
		if err != nil {
			log.Errorf("error while putting blob to local repo, %v", err)
		}
	}()
	return size, bReader, nil
}

func (c *controller) putBlobToLocal(remoteRepo string, localRepo string, desc distribution.Descriptor, r RemoteInterface) error {
	log.Debugf("Put blob to local registry!, sourceRepo:%v, localRepo:%v, digest: %v", remoteRepo, localRepo, desc.Digest)
	_, bReader, err := r.BlobReader(remoteRepo, string(desc.Digest))
	if err != nil {
		log.Errorf("failed to create blob reader, error %v", err)
		return err
	}
	defer bReader.Close()
	err = c.local.PushBlob(localRepo, desc, bReader)
	return err
}

func (c *controller) waitAndPushManifest(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, contType string, r RemoteInterface) {
	if contType == manifestlist.MediaTypeManifestList || contType == v1.MediaTypeImageIndex {
		err := c.local.PushManifestList(ctx, art.Repository, getReference(art), man)
		if err != nil {
			log.Errorf("error when push manifest list to local :%v", err)
		}
		return
	}
	var waitBlobs []distribution.Descriptor
	for n := 0; n < maxManifestWait; n++ {
		time.Sleep(sleepIntervalSec * time.Second)
		waitBlobs = c.local.CheckDependencies(ctx, art.Repository, man)
		if len(waitBlobs) == 0 {
			break
		}
		log.Debugf("Current n=%v artifact: %v:%v", n, art.Repository, art.Tag)
	}
	if len(waitBlobs) > 0 {
		// docker client will skip to pull layers exist in local
		// these blobs are not exist in the proxy server
		// it will cause the manifest dependency check always fail
		// need to push these blobs before push manifest to avoid failure
		log.Debug("Waiting blobs not empty, push it to local repo directly")
		for _, desc := range waitBlobs {
			err := c.putBlobToLocal(remoteRepo, art.Repository, desc, r)
			if err != nil {
				log.Errorf("Failed to push blob to local repo, error: %v", err)
				return
			}
		}
	}
	err := c.local.PushManifest(art.Repository, getReference(art), man)
	if err != nil {
		log.Errorf("failed to push manifest, tag: %v, error %v", art.Tag, err)
	}
}

// getRemoteRepo get the remote repository name, used in proxy cache
func getRemoteRepo(art lib.ArtifactInfo) string {
	return strings.TrimPrefix(art.Repository, art.ProjectName+"/")
}

func getReference(art lib.ArtifactInfo) string {
	if len(art.Tag) > 0 {
		return art.Tag
	}
	return art.Digest
}
