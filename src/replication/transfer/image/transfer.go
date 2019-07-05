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

package image

import (
	"errors"
	"fmt"
	"strings"

	"github.com/docker/distribution/manifest/manifestlist"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
	trans "github.com/goharbor/harbor/src/replication/transfer"
)

func init() {
	if err := trans.RegisterFactory(model.ResourceTypeImage, factory); err != nil {
		log.Errorf("failed to register transfer factory: %v", err)
	}
}

type repository struct {
	repository string
	tags       []string
}

func factory(logger trans.Logger, stopFunc trans.StopFunc) (trans.Transfer, error) {
	return &transfer{
		logger:    logger,
		isStopped: stopFunc,
	}, nil
}

type transfer struct {
	logger    trans.Logger
	isStopped trans.StopFunc
	src       adapter.ImageRegistry
	dst       adapter.ImageRegistry
}

func (t *transfer) Transfer(src *model.Resource, dst *model.Resource) error {
	// initialize
	if err := t.initialize(src, dst); err != nil {
		return err
	}

	// delete the repository on destination registry
	if dst.Deleted {
		return t.delete(&repository{
			repository: dst.Metadata.GetResourceName(),
			tags:       dst.Metadata.Vtags,
		})
	}

	srcRepo := &repository{
		repository: src.Metadata.GetResourceName(),
		tags:       src.Metadata.Vtags,
	}
	dstRepo := &repository{
		repository: dst.Metadata.GetResourceName(),
		tags:       dst.Metadata.Vtags,
	}
	// copy the repository from source registry to the destination
	return t.copy(srcRepo, dstRepo, dst.Override)
}

func (t *transfer) initialize(src *model.Resource, dst *model.Resource) error {
	if t.shouldStop() {
		return nil
	}
	// create client for source registry
	srcReg, err := createRegistry(src.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for source registry: %v", err)
		return err
	}
	t.src = srcReg
	t.logger.Infof("client for source registry [type: %s, URL: %s, insecure: %v] created",
		src.Registry.Type, src.Registry.URL, src.Registry.Insecure)

	// create client for destination registry
	dstReg, err := createRegistry(dst.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for destination registry: %v", err)
		return err
	}
	t.dst = dstReg
	t.logger.Infof("client for destination registry [type: %s, URL: %s, insecure: %v] created",
		dst.Registry.Type, dst.Registry.URL, dst.Registry.Insecure)

	return nil
}

func createRegistry(reg *model.Registry) (adapter.ImageRegistry, error) {
	factory, err := adapter.GetFactory(reg.Type)
	if err != nil {
		return nil, err
	}
	ad, err := factory(reg)
	if err != nil {
		return nil, err
	}
	registry, ok := ad.(adapter.ImageRegistry)
	if !ok {
		return nil, errors.New("the adapter doesn't implement the \"ImageRegistry\" interface")
	}
	return registry, nil
}

func (t *transfer) shouldStop() bool {
	isStopped := t.isStopped()
	if isStopped {
		t.logger.Info("the job is stopped")
	}
	return isStopped
}

func (t *transfer) copy(src *repository, dst *repository, override bool) error {
	srcRepo := src.repository
	dstRepo := dst.repository
	t.logger.Infof("copying %s:[%s](source registry) to %s:[%s](destination registry)...",
		srcRepo, strings.Join(src.tags, ","), dstRepo, strings.Join(dst.tags, ","))
	var err error
	for i := range src.tags {
		if e := t.copyImage(srcRepo, src.tags[i], dstRepo, dst.tags[i], override); e != nil {
			t.logger.Errorf(e.Error())
			err = e
		}
	}
	if err != nil {
		return err
	}

	t.logger.Infof("copy %s:[%s](source registry) to %s:[%s](destination registry) completed",
		srcRepo, strings.Join(src.tags, ","), dstRepo, strings.Join(dst.tags, ","))
	return nil
}

func (t *transfer) copyImage(srcRepo, srcRef, dstRepo, dstRef string, override bool) error {
	t.logger.Infof("copying %s:%s(source registry) to %s:%s(destination registry)...",
		srcRepo, srcRef, dstRepo, dstRef)
	// pull the manifest from the source registry
	manifest, digest, err := t.pullManifest(srcRepo, srcRef)
	if err != nil {
		return err
	}

	// check the existence of the image on the destination registry
	exist, digest2, err := t.exist(dstRepo, dstRef)
	if err != nil {
		return err
	}
	if exist {
		// the same image already exists
		if digest == digest2 {
			t.logger.Infof("the image %s:%s already exists on the destination registry, skip",
				dstRepo, dstRef)
			return nil
		}
		// the same name image exists, but not allowed to override
		if !override {
			t.logger.Warningf("the same name image %s:%s exists on the destination registry, but the \"override\" is set to false, skip",
				dstRepo, dstRef)
			return nil
		}
		// the same name image exists, but allowed to override
		t.logger.Warningf("the same name image %s:%s exists on the destination registry and the \"override\" is set to true, continue...",
			dstRepo, dstRef)
	}

	// copy contents between the source and destination registries
	for _, content := range manifest.References() {
		if err = t.copyContent(content, srcRepo, dstRepo); err != nil {
			return err
		}
	}

	// push the manifest to the destination registry
	if err := t.pushManifest(manifest, dstRepo, dstRef); err != nil {
		return err
	}

	t.logger.Infof("copy %s:%s(source registry) to %s:%s(destination registry) completed",
		srcRepo, srcRef, dstRepo, dstRef)
	return nil
}

// copy the content from source registry to destination according to its media type
func (t *transfer) copyContent(content distribution.Descriptor, srcRepo, dstRepo string) error {
	digest := content.Digest.String()
	switch content.MediaType {
	// when the media type of pulled manifest is manifest list,
	// the contents it contains are a few manifests
	case schema2.MediaTypeManifest:
		// as using digest as the reference, so set the override to true directly
		return t.copyImage(srcRepo, digest, dstRepo, digest, true)
	// handle foreign layer
	case schema2.MediaTypeForeignLayer:
		t.logger.Infof("the layer %s is a foreign layer, skip", digest)
		return nil
	// copy layer or image config
	// the media type of the layer or config can be "application/octet-stream",
	// schema1.MediaTypeManifestLayer, schema2.MediaTypeLayer, schema2.MediaTypeImageConfig
	default:
		return t.copyBlob(srcRepo, dstRepo, digest)
	}
}

// copy the layer or image config from the source registry to destination
func (t *transfer) copyBlob(srcRepo, dstRepo, digest string) error {
	if t.shouldStop() {
		return nil
	}
	t.logger.Infof("copying the blob %s...", digest)
	exist, err := t.dst.BlobExist(dstRepo, digest)
	if err != nil {
		t.logger.Errorf("failed to check the existence of blob %s on the destination registry: %v", digest, err)
		return err
	}
	if exist {
		t.logger.Infof("the blob %s already exists on the destination registry, skip", digest)
		return nil
	}

	size, data, err := t.src.PullBlob(srcRepo, digest)
	if err != nil {
		t.logger.Errorf("failed to pulling the blob %s: %v", digest, err)
		return err
	}
	defer data.Close()
	if err = t.dst.PushBlob(dstRepo, digest, size, data); err != nil {
		t.logger.Errorf("failed to pushing the blob %s: %v", digest, err)
		return err
	}
	t.logger.Infof("copy the blob %s completed", digest)
	return nil
}

func (t *transfer) pullManifest(repository, reference string) (
	distribution.Manifest, string, error) {
	if t.shouldStop() {
		return nil, "", nil
	}
	t.logger.Infof("pulling the manifest of image %s:%s ...", repository, reference)
	manifest, digest, err := t.src.PullManifest(repository, reference, []string{
		schema1.MediaTypeManifest,
		schema1.MediaTypeSignedManifest,
		schema2.MediaTypeManifest,
		manifestlist.MediaTypeManifestList,
	})
	if err != nil {
		t.logger.Errorf("failed to pull the manifest of image %s:%s: %v", repository, reference, err)
		return nil, "", err
	}
	t.logger.Infof("the manifest of image %s:%s pulled", repository, reference)

	// this is a solution to work around that harbor doesn't support manifest list
	return t.handleManifest(manifest, repository, digest)
}

// if the media type of the specified manifest is manifest list, just abstract one
// manifest from the list and return it
func (t *transfer) handleManifest(manifest distribution.Manifest, repository, digest string) (
	distribution.Manifest, string, error) {
	mediaType, _, err := manifest.Payload()
	if err != nil {
		t.logger.Errorf("failed to call the payload method for manifest of %s:%s: %v", repository, digest, err)
		return nil, "", err
	}
	// manifest
	if mediaType == schema1.MediaTypeManifest ||
		mediaType == schema1.MediaTypeSignedManifest ||
		mediaType == schema2.MediaTypeManifest {
		return manifest, digest, nil
	}
	// manifest list
	t.logger.Info("trying abstract a manifest from the manifest list...")
	manifestlist, ok := manifest.(*manifestlist.DeserializedManifestList)
	if !ok {
		err := fmt.Errorf("the object isn't a DeserializedManifestList")
		t.logger.Errorf(err.Error())
		return nil, "", err
	}
	digest = ""
	for _, reference := range manifestlist.Manifests {
		if strings.ToLower(reference.Platform.Architecture) == "amd64" &&
			strings.ToLower(reference.Platform.OS) == "linux" {
			digest = reference.Digest.String()
			t.logger.Infof("a manifest(architecture: amd64, os: linux) found, using this one: %s", digest)
			break
		}
	}
	if len(digest) == 0 {
		digest = manifest.References()[0].Digest.String()
		t.logger.Infof("no manifest(architecture: amd64, os: linux) found, using the first one: %s", digest)
	}
	return t.pullManifest(repository, digest)
}

func (t *transfer) exist(repository, tag string) (bool, string, error) {
	exist, digest, err := t.dst.ManifestExist(repository, tag)
	if err != nil {
		t.logger.Errorf("failed to check the existence of the manifest of image %s:%s on the destination registry: %v",
			repository, tag, err)
		return false, "", err
	}
	return exist, digest, nil
}

func (t *transfer) pushManifest(manifest distribution.Manifest, repository, tag string) error {
	if t.shouldStop() {
		return nil
	}
	t.logger.Infof("pushing the manifest of image %s:%s ...", repository, tag)
	mediaType, payload, err := manifest.Payload()
	if err != nil {
		t.logger.Errorf("failed to push manifest of image %s:%s: %v",
			repository, tag, err)
		return err
	}
	if err := t.dst.PushManifest(repository, tag, mediaType, payload); err != nil {
		t.logger.Errorf("failed to push manifest of image %s:%s: %v",
			repository, tag, err)
		return err
	}
	t.logger.Infof("the manifest of image %s:%s pushed",
		repository, tag)
	return nil
}

func (t *transfer) delete(repo *repository) error {
	if t.shouldStop() {
		return nil
	}

	repository := repo.repository
	for _, tag := range repo.tags {
		exist, _, err := t.dst.ManifestExist(repository, tag)
		if err != nil {
			t.logger.Errorf("failed to check the existence of the manifest of image %s:%s on the destination registry: %v",
				repository, tag, err)
			return err
		}
		if !exist {
			t.logger.Infof("the image %s:%s doesn't exist on the destination registry, skip",
				repository, tag)
			continue
		}
		if err := t.dst.DeleteManifest(repository, tag); err != nil {
			t.logger.Errorf("failed to delete the manifest of image %s:%s on the destination registry: %v",
				repository, tag, err)
			return err
		}
		t.logger.Infof("the manifest of image %s:%s is deleted", repository, tag)
	}
	return nil
}
