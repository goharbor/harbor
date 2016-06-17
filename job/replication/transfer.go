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

package replication

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
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

var (
	// ErrConflict represents http 409 error
	ErrConflict = errors.New("conflict")
)

// BaseHandler holds informations shared by other state handlers
type BaseHandler struct {
	project    string // project_name
	repository string // prject_name/repo_name
	tags       []string

	srcURL string // url of source registry

	dstURL string // url of target registry
	dstUsr string // username ...
	dstPwd string // password ...

	srcClient *registry.Repository
	dstClient *registry.Repository

	manifest distribution.Manifest // manifest of tags[0]
	blobs    []string              // blobs need to be transferred for tags[0]

	blobsExistence map[string]bool //key: digest of blob, value: existence

	logger *log.Logger
}

// InitBaseHandler initializes a BaseHandler: creating clients for source and destination registry,
// listing tags of the repository if parameter tags is nil.
func InitBaseHandler(repository, srcURL, srcSecret,
	dstURL, dstUsr, dstPwd string, tags []string, logger *log.Logger) (*BaseHandler, error) {

	logger.Infof("initializing: repository: %s, tags: %v, source URL: %s, destination URL: %s, destination user: %s",
		repository, tags, srcURL, dstURL, dstUsr)

	base := &BaseHandler{
		repository:     repository,
		tags:           tags,
		srcURL:         srcURL,
		dstURL:         dstURL,
		dstUsr:         dstUsr,
		dstPwd:         dstPwd,
		blobsExistence: make(map[string]bool, 10),
		logger:         logger,
	}

	base.project = getProjectName(base.repository)

	c := &http.Cookie{Name: models.UISecretCookie, Value: srcSecret}
	srcCred := auth.NewCookieCredential(c)
	//	srcCred := auth.NewBasicAuthCredential("admin", "Harbor12345")
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

	base.logger.Infof("initialization completed: project: %s, repository: %s, tags: %v, source URL: %s, destination URL: %s, destination user: %s",
		base.project, base.repository, base.tags, base.srcURL, base.dstURL, base.dstUsr)

	return base, nil
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

// Checker checks the existence of project and the user's privlege to the project
type Checker struct {
	*BaseHandler
}

// Enter check existence of project, if it does not exist, create it,
// if it exists, check whether the user has write privilege to it.
func (c *Checker) Enter() (string, error) {
enter:
	exist, canWrite, err := c.projectExist()
	if err != nil {
		c.logger.Errorf("an error occurred while checking existence of project %s on %s with user %s : %v", c.project, c.dstURL, c.dstUsr, err)
		return "", err
	}
	if !exist {
		err := c.createProject()
		if err != nil {
			// other job may be also doing the same thing when the current job
			// is creating project, so when the response code is 409, re-check
			// the existence of project
			if err == ErrConflict {
				goto enter
			} else {
				c.logger.Errorf("an error occurred while creating project %s on %s with user %s : %v", c.project, c.dstURL, c.dstUsr, err)
				return "", err
			}
		}
		c.logger.Infof("project %s is created on %s with user %s", c.project, c.dstURL, c.dstUsr)
		return StatePullManifest, nil
	}

	c.logger.Infof("project %s already exists on %s", c.project, c.dstURL)

	if !canWrite {
		err = fmt.Errorf("the user %s is unauthorized to write to project %s on %s", c.dstUsr, c.project, c.dstURL)
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

		for _, project := range projects {
			if project.Name == c.project {
				exist = true
				canWrite = (project.Role == models.PROJECTADMIN ||
					project.Role == models.DEVELOPER)
				break
			}
		}

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
		if resp.StatusCode == http.StatusConflict {
			return ErrConflict
		}

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

// ManifestPuller pulls the manifest of a tag. And if no tag needs to be pulled,
// the next state that state machine should enter is "finished".
type ManifestPuller struct {
	*BaseHandler
}

// Enter pulls manifest of a tag and checks if all blobs exist in the destination registry
func (m *ManifestPuller) Enter() (string, error) {
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
	name := b.repository
	tag := b.tags[0]
	for _, blob := range b.blobs {
		b.logger.Infof("transferring blob %s of %s:%s to %s ...", blob, name, tag, b.dstURL)
		size, data, err := b.srcClient.PullBlob(blob)
		if err != nil {
			b.logger.Errorf("an error occurred while pulling blob %s of %s:%s from %s: %v", blob, name, tag, b.srcURL, err)
			return "", err
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
	m.manifest = nil
	m.blobs = nil

	return StatePullManifest, nil
}
