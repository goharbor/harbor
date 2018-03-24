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
	common_http "github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/http/modifier"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	reg "github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/jobservice_v2/env"
	job_utils "github.com/vmware/harbor/src/jobservice_v2/job/impl/utils"
)

var (
	errCanceled = errors.New("the job is canceled")
)

// Replicator replicates images from source registry to the destination one
type Replicator struct {
	ctx         env.JobContext
	repository  *repository
	srcRegistry *registry
	dstRegistry *registry
	logger      *log.Logger
	retry       bool
}

// ShouldRetry : retry if the error is network error
func (r *Replicator) ShouldRetry() bool {
	return r.retry
}

// MaxFails ...
func (r *Replicator) MaxFails() uint {
	return 3
}

// Validate ....
func (r *Replicator) Validate(params map[string]interface{}) error {
	return nil
}

// Run ...
func (r *Replicator) Run(ctx env.JobContext, params map[string]interface{}) error {
	err := r.run(ctx, params)
	r.retry = retry(err)
	return err
}

func (r *Replicator) run(ctx env.JobContext, params map[string]interface{}) error {
	// initialize
	if err := r.init(ctx, params); err != nil {
		return err
	}
	// try to create project on destination registry
	if err := r.createProject(); err != nil {
		return err
	}
	// replicate the images
	for _, tag := range r.repository.tags {
		digest, manifest, err := r.pullManifest(tag)
		if err != nil {
			return err
		}
		if err := r.transferLayers(tag, manifest.References()); err != nil {
			return err
		}
		if err := r.pushManifest(tag, digest, manifest); err != nil {
			return err
		}
	}

	return nil
}

func (r *Replicator) init(ctx env.JobContext, params map[string]interface{}) error {
	// TODO
	r.logger = log.DefaultLogger()
	r.ctx = ctx

	if canceled(r.ctx) {
		r.logger.Warning(errCanceled.Error())
		return errCanceled
	}

	// init images that need to be replicated
	r.repository = &repository{
		name: params["repository"].(string),
	}
	if tags, ok := params["tags"]; ok {
		tgs := tags.([]interface{})
		for _, tg := range tgs {
			r.repository.tags = append(r.repository.tags, tg.(string))
		}
	}

	var err error
	// init source registry client
	srcURL := params["src_registry_url"].(string)
	srcInsecure := params["src_registry_insecure"].(bool)
	srcCred := auth.NewCookieCredential(&http.Cookie{
		Name:  models.UISecretCookie,
		Value: os.Getenv("JOBSERVICE_SECRET"),
	})
	srcTokenServiceURL := ""
	if stsu, ok := params["src_token_service_url"]; ok {
		srcTokenServiceURL = stsu.(string)
	}

	if len(srcTokenServiceURL) > 0 {
		r.srcRegistry, err = initRegistry(srcURL, srcInsecure, srcCred, r.repository.name, srcTokenServiceURL)
	} else {
		r.srcRegistry, err = initRegistry(srcURL, srcInsecure, srcCred, r.repository.name)
	}
	if err != nil {
		r.logger.Errorf("failed to create client for source registry: %v", err)
		return err
	}

	// init destination registry client
	dstURL := params["dst_registry_url"].(string)
	dstInsecure := params["dst_registry_insecure"].(bool)
	dstCred := auth.NewBasicAuthCredential(
		params["dst_registry_username"].(string),
		params["dst_registry_password"].(string))
	r.dstRegistry, err = initRegistry(dstURL, dstInsecure, dstCred, r.repository.name)
	if err != nil {
		r.logger.Errorf("failed to create client for destination registry: %v", err)
		return err
	}

	// get the tag list first if it is null
	if len(r.repository.tags) == 0 {
		tags, err := r.srcRegistry.ListTag()
		if err != nil {
			r.logger.Errorf("an error occurred while listing tags for the source repository: %v", err)
			return err
		}

		if len(tags) == 0 {
			err = fmt.Errorf("empty tag list for repository %s", r.repository.name)
			r.logger.Error(err)
			return err
		}
		r.repository.tags = tags
	}

	r.logger.Infof("initialization completed: repository: %s, tags: %v, source registry: URL-%s insecure-%v, destination registry: URL-%s insecure-%v",
		r.repository.name, r.repository.tags, r.srcRegistry.url, r.srcRegistry.insecure, r.dstRegistry.url, r.dstRegistry.insecure)

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

func (r *Replicator) createProject() error {
	if canceled(r.ctx) {
		r.logger.Warning(errCanceled.Error())
		return errCanceled
	}
	p, _ := utils.ParseRepository(r.repository.name)
	project, err := r.srcRegistry.GetProject(p)
	if err != nil {
		r.logger.Errorf("failed to get project %s from source registry: %v", p, err)
		return err
	}

	if err = r.dstRegistry.CreateProject(project); err != nil {
		// other jobs may be also doing the same thing when the current job
		// is creating project or the project has already exist, so when the
		// response code is 409, continue to do next step
		if e, ok := err.(*common_http.Error); ok && e.Code == http.StatusConflict {
			r.logger.Warningf("the status code is 409 when creating project %s on destination registry, try to do next step", p)
			return nil
		}

		r.logger.Errorf("an error occurred while creating project %s on destination registry: %v", p, err)
		return err
	}
	r.logger.Infof("project %s is created on destination registry", p)
	return nil
}

func (r *Replicator) pullManifest(tag string) (string, distribution.Manifest, error) {
	if canceled(r.ctx) {
		r.logger.Warning(errCanceled.Error())
		return "", nil, errCanceled
	}

	acceptMediaTypes := []string{schema1.MediaTypeManifest, schema2.MediaTypeManifest}
	digest, mediaType, payload, err := r.srcRegistry.PullManifest(tag, acceptMediaTypes)
	if err != nil {
		r.logger.Errorf("an error occurred while pulling manifest of %s:%s from source registry: %v",
			r.repository.name, tag, err)
		return "", nil, err
	}
	r.logger.Infof("manifest of %s:%s pulled successfully from source registry: %s",
		r.repository.name, tag, digest)

	if strings.Contains(mediaType, "application/json") {
		mediaType = schema1.MediaTypeManifest
	}

	manifest, _, err := reg.UnMarshal(mediaType, payload)
	if err != nil {
		r.logger.Errorf("an error occurred while parsing manifest: %v", err)
		return "", nil, err
	}

	return digest, manifest, nil
}

func (r *Replicator) transferLayers(tag string, blobs []distribution.Descriptor) error {
	repository := r.repository.name

	// all blobs(layers and config)
	for _, blob := range blobs {
		if canceled(r.ctx) {
			r.logger.Warning(errCanceled.Error())
			return errCanceled
		}

		digest := blob.Digest.String()
		exist, err := r.dstRegistry.BlobExist(digest)
		if err != nil {
			r.logger.Errorf("an error occurred while checking existence of blob %s of %s:%s on destination registry: %v",
				digest, repository, tag, err)
			return err
		}
		if exist {
			r.logger.Infof("blob %s of %s:%s already exists on the destination registry, skip",
				digest, repository, tag)
			continue
		}

		r.logger.Infof("transferring blob %s of %s:%s to the destination registry ...",
			digest, repository, tag)
		size, data, err := r.srcRegistry.PullBlob(digest)
		if err != nil {
			r.logger.Errorf("an error occurred while pulling blob %s of %s:%s from the source registry: %v",
				digest, repository, tag, err)
			return err
		}
		if data != nil {
			defer data.Close()
		}
		if err = r.dstRegistry.PushBlob(digest, size, data); err != nil {
			r.logger.Errorf("an error occurred while pushing blob %s of %s:%s to the distination registry: %v",
				digest, repository, tag, err)
			return err
		}
		r.logger.Infof("blob %s of %s:%s transferred to the destination registry completed",
			digest, repository, tag)
	}

	return nil
}

func (r *Replicator) pushManifest(tag, digest string, manifest distribution.Manifest) error {
	if canceled(r.ctx) {
		r.logger.Warning(errCanceled.Error())
		return errCanceled
	}

	repository := r.repository.name
	_, exist, err := r.dstRegistry.ManifestExist(digest)
	if err != nil {
		r.logger.Warningf("an error occurred while checking the existence of manifest of %s:%s on the destination registry: %v, try to push manifest",
			repository, tag, err)
	} else {
		if exist {
			r.logger.Infof("manifest of %s:%s exists on the destination registry, skip manifest pushing",
				repository, tag)
			return nil
		}
	}

	mediaType, data, err := manifest.Payload()
	if err != nil {
		r.logger.Errorf("an error occurred while getting payload of manifest for %s:%s : %v",
			repository, tag, err)
		return err
	}

	if _, err = r.dstRegistry.PushManifest(tag, mediaType, data); err != nil {
		r.logger.Errorf("an error occurred while pushing manifest of %s:%s to the destination registry: %v",
			repository, tag, err)
		return err
	}
	r.logger.Infof("manifest of %s:%s has been pushed to the destination registry",
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
