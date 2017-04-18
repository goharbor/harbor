// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/common/utils/registry"
	"github.com/vmware/harbor/src/common/utils/registry/auth"
	"github.com/vmware/harbor/src/jobservice/config"
)

const (
	// StateInitialize ...
	StateInitialize = "initialize"
	// StateCheck ...
	StateCheck = "check"
	// StatePullManifest ...
	StatePullManifest = "pull_manifest"
	// StateTransferBlob ...
	StateTransferBlob = "transfer_blob"
	// StatePushManifest ...
	StatePushManifest = "push_manifest"
)

var (
	// ErrConflict represents http 409 error
	ErrConflict = errors.New("conflict")
)

// BaseHandler holds informations shared by other state handlers
type BaseHandler struct {
	project    string // project_name
	repository string // prject_name/repo_name
	tags       []string

	srcURL    string // url of source registry
	srcSecret string

	dstURL string // url of target registry
	dstUsr string // username ...
	dstPwd string // password ...

	insecure bool // whether skip secure check when using https

	srcClient *registry.Repository
	dstClient *registry.Repository

	manifest distribution.Manifest // manifest of tags[0]
	digest   string                //digest of tags[0]'s manifest
	blobs    []string              // blobs need to be transferred for tags[0]

	blobsExistence map[string]bool //key: digest of blob, value: existence

	logger *log.Logger
}

// InitBaseHandler initializes a BaseHandler.
func InitBaseHandler(repository, srcURL, srcSecret,
	dstURL, dstUsr, dstPwd string, insecure bool, tags []string, logger *log.Logger) *BaseHandler {

	base := &BaseHandler{
		repository:     repository,
		tags:           tags,
		srcURL:         srcURL,
		srcSecret:      srcSecret,
		dstURL:         dstURL,
		dstUsr:         dstUsr,
		dstPwd:         dstPwd,
		insecure:       insecure,
		blobsExistence: make(map[string]bool, 10),
		logger:         logger,
	}

	base.project = getProjectName(base.repository)

	return base
}

// Exit ...
func (b *BaseHandler) Exit() error {
	return nil
}

func getProjectName(repository string) string {
	repository = strings.TrimSpace(repository)
	repository = strings.TrimRight(repository, "/")
	return repository[:strings.LastIndex(repository, "/")]
}

// Initializer creates clients for source and destination registry,
// lists tags of the repository if parameter tags is nil.
type Initializer struct {
	*BaseHandler
}

// Enter ...
func (i *Initializer) Enter() (string, error) {
	i.logger.Infof("initializing: repository: %s, tags: %v, source URL: %s, destination URL: %s, insecure: %v, destination user: %s",
		i.repository, i.tags, i.srcURL, i.dstURL, i.insecure, i.dstUsr)

	state, err := i.enter()
	if err != nil && retry(err) {
		i.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err
}

func (i *Initializer) enter() (string, error) {
	c := &http.Cookie{Name: models.UISecretCookie, Value: i.srcSecret}
	srcCred := auth.NewCookieCredential(c)
	srcClient, err := newRepositoryClient(i.srcURL, i.insecure, srcCred,
		config.InternalTokenServiceEndpoint(), i.repository, "repository", i.repository, "pull", "push", "*")
	if err != nil {
		i.logger.Errorf("an error occurred while creating source repository client: %v", err)
		return "", err
	}
	i.srcClient = srcClient

	dstCred := auth.NewBasicAuthCredential(i.dstUsr, i.dstPwd)
	dstClient, err := newRepositoryClient(i.dstURL, i.insecure, dstCred,
		"", i.repository, "repository", i.repository, "pull", "push", "*")
	if err != nil {
		i.logger.Errorf("an error occurred while creating destination repository client: %v", err)
		return "", err
	}
	i.dstClient = dstClient

	if len(i.tags) == 0 {
		tags, err := i.srcClient.ListTag()
		if err != nil {
			i.logger.Errorf("an error occurred while listing tags for source repository: %v", err)
			return "", err
		}
		i.tags = tags
	}

	i.logger.Infof("initialization completed: project: %s, repository: %s, tags: %v, source URL: %s, destination URL: %s, insecure: %v, destination user: %s",
		i.project, i.repository, i.tags, i.srcURL, i.dstURL, i.insecure, i.dstUsr)

	return StateCheck, nil
}

// Checker checks the existence of project and the user's privlege to the project
type Checker struct {
	*BaseHandler
}

// Enter check existence of project, if it does not exist, create it,
// if it exists, check whether the user has write privilege to it.
func (c *Checker) Enter() (string, error) {
	state, err := c.enter()
	if err != nil && retry(err) {
		c.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err
}

func (c *Checker) enter() (string, error) {
	project, err := dao.GetProjectByName(c.project)
	if err != nil {
		c.logger.Errorf("an error occurred while getting project %s in DB: %v", c.project, err)
		return "", err
	}

	err = c.createProject(project.Public)
	if err == nil {
		c.logger.Infof("project %s is created on %s with user %s", c.project, c.dstURL, c.dstUsr)
		return StatePullManifest, nil
	}

	// other job may be also doing the same thing when the current job
	// is creating project, so when the response code is 409, continue
	// to do next step
	if err == ErrConflict {
		c.logger.Warningf("the status code is 409 when creating project %s on %s with user %s, try to do next step", c.project, c.dstURL, c.dstUsr)
		return StatePullManifest, nil
	}

	c.logger.Errorf("an error occurred while creating project %s on %s with user %s : %v", c.project, c.dstURL, c.dstUsr, err)

	return "", err
}

func (c *Checker) createProject(public int) error {
	project := struct {
		ProjectName string `json:"project_name"`
		Public      int    `json:"public"`
	}{
		ProjectName: c.project,
		Public:      public,
	}

	data, err := json.Marshal(project)
	if err != nil {
		return err
	}

	url := strings.TrimRight(c.dstURL, "/") + "/api/projects/"
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.dstUsr, c.dstPwd)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.insecure,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// version 0.1.1's reponse code is 200
	if resp.StatusCode == http.StatusCreated ||
		resp.StatusCode == http.StatusOK {
		return nil
	}

	if resp.StatusCode == http.StatusConflict {
		return ErrConflict
	}

	message, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.logger.Errorf("an error occurred while reading message from response: %v", err)
	}

	return fmt.Errorf("failed to create project %s on %s with user %s: %d %s",
		c.project, c.dstURL, c.dstUsr, resp.StatusCode, string(message))
}

// ManifestPuller pulls the manifest of a tag. And if no tag needs to be pulled,
// the next state that state machine should enter is "finished".
type ManifestPuller struct {
	*BaseHandler
}

// Enter pulls manifest of a tag and checks if all blobs exist in the destination registry
func (m *ManifestPuller) Enter() (string, error) {
	state, err := m.enter()
	if err != nil && retry(err) {
		m.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err

}

func (m *ManifestPuller) enter() (string, error) {
	if len(m.tags) == 0 {
		m.logger.Infof("no tag needs to be replicated, next state is \"finished\"")
		return models.JobFinished, nil
	}

	name := m.repository
	tag := m.tags[0]

	acceptMediaTypes := []string{schema1.MediaTypeManifest, schema2.MediaTypeManifest}
	digest, mediaType, payload, err := m.srcClient.PullManifest(tag, acceptMediaTypes)
	if err != nil {
		m.logger.Errorf("an error occurred while pulling manifest of %s:%s from %s: %v", name, tag, m.srcURL, err)
		return "", err
	}
	m.digest = digest
	m.logger.Infof("manifest of %s:%s pulled successfully from %s: %s", name, tag, m.srcURL, digest)

	if strings.Contains(mediaType, "application/json") {
		mediaType = schema1.MediaTypeManifest
	}

	manifest, _, err := registry.UnMarshal(mediaType, payload)
	if err != nil {
		m.logger.Errorf("an error occurred while parsing manifest of %s:%s from %s: %v", name, tag, m.srcURL, err)
		return "", err
	}

	m.manifest = manifest

	// all blobs(layers and config)
	var blobs []string

	for _, discriptor := range manifest.References() {
		blobs = append(blobs, discriptor.Digest.String())
	}

	m.logger.Infof("all blobs of %s:%s from %s: %v", name, tag, m.srcURL, blobs)

	for _, blob := range blobs {
		exist, ok := m.blobsExistence[blob]
		if !ok {
			exist, err = m.dstClient.BlobExist(blob)
			if err != nil {
				m.logger.Errorf("an error occurred while checking existence of blob %s of %s:%s on %s: %v", blob, name, tag, m.dstURL, err)
				return "", err
			}
			m.blobsExistence[blob] = exist
		}

		if !exist {
			m.blobs = append(m.blobs, blob)
		} else {
			m.logger.Infof("blob %s of %s:%s already exists in %s", blob, name, tag, m.dstURL)
		}
	}
	m.logger.Infof("blobs of %s:%s need to be transferred to %s: %v", name, tag, m.dstURL, m.blobs)

	return StateTransferBlob, nil
}

// BlobTransfer transfers blobs of a tag
type BlobTransfer struct {
	*BaseHandler
}

// Enter pulls blobs and then pushs them to destination registry.
func (b *BlobTransfer) Enter() (string, error) {
	state, err := b.enter()
	if err != nil && retry(err) {
		b.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err

}

func (b *BlobTransfer) enter() (string, error) {
	name := b.repository
	tag := b.tags[0]
	for _, blob := range b.blobs {
		b.logger.Infof("transferring blob %s of %s:%s to %s ...", blob, name, tag, b.dstURL)
		size, data, err := b.srcClient.PullBlob(blob)
		if err != nil {
			b.logger.Errorf("an error occurred while pulling blob %s of %s:%s from %s: %v", blob, name, tag, b.srcURL, err)
			return "", err
		}
		if data != nil {
			defer data.Close()
		}
		if err = b.dstClient.PushBlob(blob, size, data); err != nil {
			b.logger.Errorf("an error occurred while pushing blob %s of %s:%s to %s : %v", blob, name, tag, b.dstURL, err)
			return "", err
		}
		b.logger.Infof("blob %s of %s:%s transferred to %s completed", blob, name, tag, b.dstURL)
	}

	return StatePushManifest, nil
}

// ManifestPusher pushs the manifest to destination registry
type ManifestPusher struct {
	*BaseHandler
}

// Enter checks the existence of manifest in the source registry first, and if it
// exists, pushs it to destination registry. The checking operation is to avoid
// the situation that the tag is deleted during the blobs transfering
func (m *ManifestPusher) Enter() (string, error) {
	state, err := m.enter()
	if err != nil && retry(err) {
		m.logger.Info("waiting for retrying...")
		return models.JobRetrying, nil
	}

	return state, err

}

func (m *ManifestPusher) enter() (string, error) {
	name := m.repository
	tag := m.tags[0]
	_, exist, err := m.srcClient.ManifestExist(tag)
	if err != nil {
		m.logger.Infof("an error occurred while checking the existence of manifest of %s:%s on %s: %v", name, tag, m.srcURL, err)
		return "", err
	}
	if !exist {
		m.logger.Infof("manifest of %s:%s does not exist on source registry %s, cancel manifest pushing", name, tag, m.srcURL)
	} else {
		m.logger.Infof("manifest of %s:%s exists on source registry %s, continue manifest pushing", name, tag, m.srcURL)

		digest, manifestExist, err := m.dstClient.ManifestExist(tag)
		if manifestExist && digest == m.digest {
			m.logger.Infof("manifest of %s:%s exists on destination registry %s, skip manifest pushing", name, tag, m.dstURL)

			m.tags = m.tags[1:]
			m.manifest = nil
			m.digest = ""
			m.blobs = nil

			return StatePullManifest, nil
		}

		mediaType, data, err := m.manifest.Payload()
		if err != nil {
			m.logger.Errorf("an error occurred while getting payload of manifest for %s:%s : %v", name, tag, err)
			return "", err
		}

		if _, err = m.dstClient.PushManifest(tag, mediaType, data); err != nil {
			m.logger.Errorf("an error occurred while pushing manifest of %s:%s to %s : %v", name, tag, m.dstURL, err)
			return "", err
		}
		m.logger.Infof("manifest of %s:%s has been pushed to %s", name, tag, m.dstURL)
	}

	m.tags = m.tags[1:]
	m.manifest = nil
	m.digest = ""
	m.blobs = nil

	return StatePullManifest, nil
}

func newRepositoryClient(endpoint string, insecure bool, credential auth.Credential,
	tokenServiceEndpoint, repository, scopeType, scopeName string,
	scopeActions ...string) (*registry.Repository, error) {
	authorizer := auth.NewStandardTokenAuthorizer(credential, insecure,
		tokenServiceEndpoint, scopeType, scopeName, scopeActions...)

	store, err := auth.NewAuthorizerStore(endpoint, insecure, authorizer)
	if err != nil {
		return nil, err
	}

	uam := &userAgentModifier{
		userAgent: "harbor-registry-client",
	}

	client, err := registry.NewRepositoryWithModifiers(repository, endpoint, insecure, store, uam)
	if err != nil {
		return nil, err
	}
	return client, nil
}

type userAgentModifier struct {
	userAgent string
}

// Modify adds user-agent header to the request
func (u *userAgentModifier) Modify(req *http.Request) error {
	req.Header.Set(http.CanonicalHeaderKey("User-Agent"), u.userAgent)
	return nil
}
