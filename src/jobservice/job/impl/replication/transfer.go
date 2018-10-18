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

package replication

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/http/modifier"
	httpauth "github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/utils"
	reg "github.com/goharbor/harbor/src/common/utils/registry"
	"github.com/goharbor/harbor/src/common/utils/registry/auth"
	"github.com/goharbor/harbor/src/jobservice/env"
	job_utils "github.com/goharbor/harbor/src/jobservice/job/impl/utils"
	"github.com/goharbor/harbor/src/jobservice/logger"
)

var (
	errCanceled = errors.New("the job is canceled")
)

// Transfer images from source registry to the destination one
type Transfer struct {
	ctx         env.JobContext
	repository  *repository
	srcRegistry *registry
	dstRegistry *registry
	logger      logger.Interface
	retry       bool
}

// ShouldRetry : retry if the error is network error
func (t *Transfer) ShouldRetry() bool {
	return t.retry
}

// MaxFails ...
func (t *Transfer) MaxFails() uint {
	return 3
}

// Validate ....
func (t *Transfer) Validate(params map[string]interface{}) error {
	return nil
}

// Run ...
func (t *Transfer) Run(ctx env.JobContext, params map[string]interface{}) error {
	err := t.run(ctx, params)
	t.retry = retry(err)
	return err
}

func (t *Transfer) run(ctx env.JobContext, params map[string]interface{}) error {
	// initialize
	if err := t.init(ctx, params); err != nil {
		return err
	}
	// try to create project on destination registry
	if err := t.createProject(); err != nil {
		return err
	}
	// replicate the images
	for _, tag := range t.repository.tags {
		digest, manifest, err := t.pullManifest(tag)
		if err != nil {
			return err
		}
		if err := t.transferLayers(tag, manifest.References()); err != nil {
			return err
		}
		if err := t.pushManifest(tag, digest, manifest); err != nil {
			return err
		}
	}

	return nil
}

func (t *Transfer) init(ctx env.JobContext, params map[string]interface{}) error {
	t.logger = ctx.GetLogger()
	t.ctx = ctx

	if canceled(t.ctx) {
		t.logger.Warning(errCanceled.Error())
		return errCanceled
	}

	// init images that need to be replicated
	t.repository = &repository{
		name: params["repository"].(string),
	}
	if tags, ok := params["tags"]; ok {
		tgs := tags.([]interface{})
		for _, tg := range tgs {
			t.repository.tags = append(t.repository.tags, tg.(string))
		}
	}

	var err error
	// init source registry client
	srcURL := params["src_registry_url"].(string)
	srcInsecure := params["src_registry_insecure"].(bool)
	srcCred := httpauth.NewSecretAuthorizer(secret())
	srcTokenServiceURL := ""
	if stsu, ok := params["src_token_service_url"]; ok {
		srcTokenServiceURL = stsu.(string)
	}

	if len(srcTokenServiceURL) > 0 {
		t.srcRegistry, err = initRegistry(srcURL, srcInsecure, srcCred, t.repository.name, srcTokenServiceURL)
	} else {
		t.srcRegistry, err = initRegistry(srcURL, srcInsecure, srcCred, t.repository.name)
	}
	if err != nil {
		t.logger.Errorf("failed to create client for source registry: %v", err)
		return err
	}

	// init destination registry client
	dstURL := params["dst_registry_url"].(string)
	dstInsecure := params["dst_registry_insecure"].(bool)
	dstCred := auth.NewBasicAuthCredential(
		params["dst_registry_username"].(string),
		params["dst_registry_password"].(string))
	t.dstRegistry, err = initRegistry(dstURL, dstInsecure, dstCred, t.repository.name)
	if err != nil {
		t.logger.Errorf("failed to create client for destination registry: %v", err)
		return err
	}

	// get the tag list first if it is null
	if len(t.repository.tags) == 0 {
		tags, err := t.srcRegistry.ListTag()
		if err != nil {
			t.logger.Errorf("an error occurred while listing tags for the source repository: %v", err)
			return err
		}

		if len(tags) == 0 {
			err = fmt.Errorf("empty tag list for repository %s", t.repository.name)
			t.logger.Error(err)
			return err
		}
		t.repository.tags = tags
	}

	t.logger.Infof("initialization completed: repository: %s, tags: %v, source registry: URL-%s insecure-%v, destination registry: URL-%s insecure-%v",
		t.repository.name, t.repository.tags, t.srcRegistry.url, t.srcRegistry.insecure, t.dstRegistry.url, t.dstRegistry.insecure)

	return nil
}

func initRegistry(url string, insecure bool, credential modifier.Modifier,
	repository string, tokenServiceURL ...string) (*registry, error) {
	registry := &registry{
		url:      url,
		insecure: insecure,
	}

	// use the same transport for clients connecting to docker registry and Harbor UI
	transport := reg.GetHTTPTransport(insecure)

	authorizer := auth.NewStandardTokenAuthorizer(&http.Client{
		Transport: transport,
	}, credential, tokenServiceURL...)
	uam := &job_utils.UserAgentModifier{
		UserAgent: "harbor-registry-client",
	}
	repositoryClient, err := reg.NewRepository(repository, url,
		&http.Client{
			Transport: reg.NewTransport(transport, authorizer, uam),
		})
	if err != nil {
		return nil, err
	}
	registry.Repository = *repositoryClient

	registry.client = common_http.NewClient(
		&http.Client{
			Transport: transport,
		}, credential)
	return registry, nil
}

func (t *Transfer) createProject() error {
	if canceled(t.ctx) {
		t.logger.Warning(errCanceled.Error())
		return errCanceled
	}
	p, _ := utils.ParseRepository(t.repository.name)
	project, err := t.srcRegistry.GetProject(p)
	if err != nil {
		t.logger.Errorf("failed to get project %s from source registry: %v", p, err)
		return err
	}

	if err = t.dstRegistry.CreateProject(project); err != nil {
		// other jobs may be also doing the same thing when the current job
		// is creating project or the project has already exist, so when the
		// response code is 409, continue to do next step
		if e, ok := err.(*common_http.Error); ok && e.Code == http.StatusConflict {
			t.logger.Warningf("the status code is 409 when creating project %s on destination registry, try to do next step", p)
			return nil
		}

		t.logger.Errorf("an error occurred while creating project %s on destination registry: %v", p, err)
		return err
	}
	t.logger.Infof("project %s is created on destination registry", p)
	return nil
}

func (t *Transfer) pullManifest(tag string) (string, distribution.Manifest, error) {
	if canceled(t.ctx) {
		t.logger.Warning(errCanceled.Error())
		return "", nil, errCanceled
	}

	acceptMediaTypes := []string{schema1.MediaTypeManifest, schema2.MediaTypeManifest}
	digest, mediaType, payload, err := t.srcRegistry.PullManifest(tag, acceptMediaTypes)
	if err != nil {
		t.logger.Errorf("an error occurred while pulling manifest of %s:%s from source registry: %v",
			t.repository.name, tag, err)
		return "", nil, err
	}
	t.logger.Infof("manifest of %s:%s pulled successfully from source registry: %s",
		t.repository.name, tag, digest)

	if strings.Contains(mediaType, "application/json") {
		mediaType = schema1.MediaTypeManifest
	}

	manifest, _, err := reg.UnMarshal(mediaType, payload)
	if err != nil {
		t.logger.Errorf("an error occurred while parsing manifest: %v", err)
		return "", nil, err
	}

	return digest, manifest, nil
}

func (t *Transfer) transferLayers(tag string, blobs []distribution.Descriptor) error {
	repository := t.repository.name

	// all blobs(layers and config)
	for _, blob := range blobs {
		if canceled(t.ctx) {
			t.logger.Warning(errCanceled.Error())
			return errCanceled
		}

		digest := blob.Digest.String()
		exist, err := t.dstRegistry.BlobExist(digest)
		if err != nil {
			t.logger.Errorf("an error occurred while checking existence of blob %s of %s:%s on destination registry: %v",
				digest, repository, tag, err)
			return err
		}
		if exist {
			t.logger.Infof("blob %s of %s:%s already exists on the destination registry, skip",
				digest, repository, tag)
			continue
		}

		t.logger.Infof("transferring blob %s of %s:%s to the destination registry ...",
			digest, repository, tag)
		size, data, err := t.srcRegistry.PullBlob(digest)
		if err != nil {
			t.logger.Errorf("an error occurred while pulling blob %s of %s:%s from the source registry: %v",
				digest, repository, tag, err)
			return err
		}
		if data != nil {
			defer data.Close()
		}
		if err = t.dstRegistry.PushBlob(digest, size, data); err != nil {
			t.logger.Errorf("an error occurred while pushing blob %s of %s:%s to the distination registry: %v",
				digest, repository, tag, err)
			return err
		}
		t.logger.Infof("blob %s of %s:%s transferred to the destination registry completed",
			digest, repository, tag)
	}

	return nil
}

func (t *Transfer) pushManifest(tag, digest string, manifest distribution.Manifest) error {
	if canceled(t.ctx) {
		t.logger.Warning(errCanceled.Error())
		return errCanceled
	}

	repository := t.repository.name
	dgt, exist, err := t.dstRegistry.ManifestExist(tag)
	if err != nil {
		t.logger.Warningf("an error occurred while checking the existence of manifest of %s:%s on the destination registry: %v, try to push manifest",
			repository, tag, err)
	} else {
		if exist && dgt == digest {
			t.logger.Infof("manifest of %s:%s exists on the destination registry, skip manifest pushing",
				repository, tag)
			return nil
		}
	}

	mediaType, data, err := manifest.Payload()
	if err != nil {
		t.logger.Errorf("an error occurred while getting payload of manifest for %s:%s : %v",
			repository, tag, err)
		return err
	}

	if _, err = t.dstRegistry.PushManifest(tag, mediaType, data); err != nil {
		t.logger.Errorf("an error occurred while pushing manifest of %s:%s to the destination registry: %v",
			repository, tag, err)
		return err
	}
	t.logger.Infof("manifest of %s:%s has been pushed to the destination registry",
		repository, tag)

	return nil
}

func canceled(ctx env.JobContext) bool {
	_, canceled := ctx.OPCommand()
	return canceled
}

func retry(err error) bool {
	if err == nil {
		return false
	}
	_, ok := err.(net.Error)
	return ok
}

func secret() string {
	return os.Getenv("JOBSERVICE_SECRET")
}
