/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package imgout

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/job/utils"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/registry"
	"github.com/vmware/harbor/utils/registry/auth"
)

const (
	// StateCheck ...
	StateCheck = "check"
	// StatePullManifest ...
	StatePullManifest = "pull_manifest"
	// StateTransferBlob ...
	StateTransferBlob = "transfer_blob"
	// StatePushManifest ...
	StatePushManifest = "push_manifest"
)

type BaseHandler struct {
	project    string // project_name
	repository string // prject_name/repo_name
	tags       []string

	srcURL       string // url of source registry
	srcSecretKey string // secretKey ...

	dstURL string // url of target registry
	dstUsr string // username ...
	dstPwd string // password ...

	srcClient *registry.Repository
	dstClient *registry.Repository

	manifest distribution.Manifest // manifest of tags[0]
	blobs    []string              // blobs need to be transferred for tags[0]

	logger *utils.Logger
}

func InitBaseHandler(repository, srcURL, srcSecretKey,
	dstURL, dstUsr, dstPwd string, tags []string, logger *utils.Logger) (*BaseHandler, error) {

	base := &BaseHandler{
		repository:   repository,
		tags:         tags,
		srcURL:       srcURL,
		srcSecretKey: srcSecretKey,
		dstURL:       dstURL,
		dstUsr:       dstUsr,
		dstPwd:       dstPwd,
		logger:       logger,
	}

	base.project = getProjectName(base.repository)

	//TODO using secret key
	srcCred := auth.NewBasicAuthCredential("admin", "Harbor12345")
	srcClient, err := registry.NewRepositoryWithCredential(base.repository, base.srcURL, srcCred)
	if err != nil {
		return nil, err
	}
	base.srcClient = srcClient

	dstCred := auth.NewBasicAuthCredential(base.dstUsr, base.dstPwd)
	dstClient, err := registry.NewRepositoryWithCredential(base.repository, base.dstURL, dstCred)
	if err != nil {
		return nil, err
	}
	base.dstClient = dstClient

	if len(base.tags) == 0 {
		tags, err := base.srcClient.ListTag()
		if err != nil {
			return nil, err
		}
		base.tags = tags
	}

	base.logger.Infof("initialization of base handler completed: project: %s, repository: %s, tags: %v, source URL: %s, destination URL: %s, destination user: %s",
		base.project, base.repository, base.tags, base.srcURL, base.dstURL, base.dstUsr)

	return base, nil
}

func (b *BaseHandler) Exit() error {
	return nil
}

func getProjectName(repository string) string {
	repository = strings.TrimSpace(repository)
	repository = strings.TrimRight(repository, "/")
	return repository[:strings.LastIndex(repository, "/")]
}

type Checker struct {
	*BaseHandler
}

// check existence of project, if it does not exist, create it,
// if it exists, check whether the user has write privilege to it.
func (c *Checker) Enter() (string, error) {
	exist, canWrite, err := c.projectExist()
	if err != nil {
		c.logger.Errorf("an error occurred while checking existence of project %s on %s with user %s : %v", c.project, c.dstURL, c.dstUsr, err)
		return "", err
	}
	if !exist {
		if err := c.createProject(); err != nil {
			c.logger.Errorf("an error occurred while creating project %s on %s with user %s : %v", c.project, c.dstURL, c.dstUsr, err)
			return "", err
		}
		c.logger.Infof("project %s is created on %s with user %s", c.project, c.dstURL, c.dstUsr)
		return StatePullManifest, nil
	}

	c.logger.Infof("project %s already exists on %s", c.project, c.dstURL)

	if !canWrite {
		err = fmt.Errorf("the user %s has no write privilege to project %s on %s", c.dstUsr, c.project, c.dstURL)
		c.logger.Errorf("%v", err)
		return "", err
	}
	c.logger.Infof("the user %s has write privilege to project %s on %s", c.dstUsr, c.project, c.dstURL)

	return StatePullManifest, nil
}

// check the existence of project, if it exists, returning whether the user has write privilege to it
func (c *Checker) projectExist() (exist, canWrite bool, err error) {
	url := strings.TrimRight(c.dstURL, "/") + "/api/projects/?project_name=" + c.project
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	req.SetBasicAuth(c.dstUsr, c.dstPwd)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusNotFound {
		return
	}

	if resp.StatusCode == http.StatusUnauthorized {
		exist = true
		return
	}

	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusOK {
		projects := make([]models.Project, 1)
		if err = json.Unmarshal(data, &projects); err != nil {
			return
		}

		if len(projects) == 0 {
			return
		}

		exist = true
		// TODO handle canWrite when new API is ready
		canWrite = true

		return
	}

	err = fmt.Errorf("an error occurred while checking existen of project %s on %s with user %s: %d %s",
		c.project, c.dstURL, c.dstUsr, resp.StatusCode, string(data))

	return
}

func (c *Checker) createProject() error {
	// TODO handle publicity of project
	project := struct {
		ProjectName string `json:"project_name"`
		Public      bool   `json:"public"`
	}{
		ProjectName: c.project,
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
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusCreated {
		defer resp.Body.Close()
		message, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.logger.Errorf("an error occurred while reading message from response: %v", err)
		}

		return fmt.Errorf("failed to create project %s on %s with user %s: %d %s",
			c.project, c.dstURL, c.dstUsr, resp.StatusCode, string(message))
	}

	return nil
}

type ManifestPuller struct {
	*BaseHandler
}

func (m *ManifestPuller) Enter() (string, error) {
	if len(m.tags) == 0 {
		m.logger.Infof("no tag needs to be replicated, entering finish state")
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

	// config is also need to be transferred if the schema of manifest is v2
	manifest2, ok := manifest.(*schema2.DeserializedManifest)
	if ok {
		blobs = append(blobs, manifest2.Target().Digest.String())
	}

	m.logger.Infof("all blobs of %s:%s from %s: %v", name, tag, m.srcURL, blobs)

	for _, blob := range blobs {
		exist, err := m.dstClient.BlobExist(blob)
		if err != nil {
			m.logger.Errorf("an error occurred while checking existence of blob %s of %s:%s on %s: %v", blob, name, tag, m.dstURL, err)
			return "", err
		}
		if !exist {
			m.blobs = append(m.blobs, blob)
		}
	}
	m.logger.Infof("blobs of %s:%s need to be transferred to %s: %v", name, tag, m.dstURL, m.blobs)

	m.blobs = blobs

	return StateTransferBlob, nil
}

type BlobTransfer struct {
	*BaseHandler
}

func (b *BlobTransfer) Enter() (string, error) {
	name := b.repository
	tag := b.tags[0]
	for _, blob := range b.blobs {
		size, data, err := b.srcClient.PullBlob(blob)
		if err != nil {
			b.logger.Errorf("an error occurred while pulling blob %s of %s:%s from %s: %v", blob, name, tag, b.srcURL, err)
			return "", err
		}
		if err = b.dstClient.PushBlob(blob, size, data); err != nil {
			b.logger.Errorf("an error occurred while pushing blob %s of %s:%s to %s : %v", blob, name, tag, b.dstURL, err)
			return "", err
		}
		b.logger.Infof("blob %s of %s:%s tranferred to %s completed", blob, name, tag, b.dstURL)
	}

	return StatePushManifest, nil
}

type ManifestPusher struct {
	*BaseHandler
}

func (m *ManifestPusher) Enter() (string, error) {
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

	return StatePullManifest, nil
}
