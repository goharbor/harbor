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

package syncimg

/*
import (
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/job"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/registry"
	"github.com/vmware/harbor/utils/registry/auth"
)

const (
	//jobType           = "sync_img"
	stateCheck        = "check"
	statePullManifest = "pull_manifest"
	stateTransferBlob = "transfer_blob"
	statePushManifest = "push_manifest"
)

func addTransition(sm *job.JobSM) error {
	base, err := initBaseHandler(sm)
	if err != nil {
		return err
	}

	sm.AddTransition(models.JobRunning, stateCheck, &checker{baseHandler: base})
	sm.AddTransition(stateCheck, statePullManifest, &manifestPuller{baseHandler: base})
	sm.AddTransition(statePullManifest, stateTransferBlob, &blobTransfer{baseHandler: base})
	sm.AddTransition(statePullManifest, models.JobFinished, &job.StatusUpdater{job.DummyHandler{JobID: sm.JobID}, models.JobFinished})
	sm.AddTransition(stateTransferBlob, statePushManifest, &manifestPusher{baseHandler: base})
	sm.AddTransition(statePushManifest, statePullManifest, &manifestPuller{baseHandler: base})

	return nil
}

type baseHandler struct {
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

	logger *job.Logger
}

func initBaseHandler(sm *job.JobSM) (*baseHandler, error) {
	base := &baseHandler{
		repository: sm.Parms.Repository,
		srcURL:     sm.Parms.LocalRegURL,
		dstURL:     sm.Parms.TargetURL,
		dstUsr:     sm.Parms.TargetUsername,
		dstPwd:     sm.Parms.TargetPassword,
		logger:     &(sm.Logger),
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

func (b *baseHandler) Exit() error {
	return nil
}

func getProjectName(repository string) string {
	repository = strings.TrimSpace(repository)
	repository = strings.TrimRight(repository, "/")
	return repository[:strings.LastIndex(repository, "/")]
}

type checker struct {
	*baseHandler
}

// check existence of project, if it does not exist, create it,
// if it exists, check whether the user has write privilege to it.
func (c *checker) Enter() (string, error) {
	exist, err := c.projectExist()
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
		return statePullManifest, nil
	}

	c.logger.Infof("project %s already exists on %s", c.project, c.dstURL)
	canWrite, err := c.canWrite()
	if err != nil {
		c.logger.Errorf("an error occurred while checking %s 's privileges to project %s on %s : %v", c.dstUsr, c.project, c.dstURL, err)
		return "", err
	}

	if !canWrite {
		c.logger.Errorf("the user %s has no write privilege to project %s on %s", c.dstUsr, c.project, c.dstURL)
		return "", err
	}
	c.logger.Infof("the user %s has write privilege to project %s on %s", c.dstUsr, c.project, c.dstURL)

	return statePullManifest, nil
}

func (c *checker) projectExist() (bool, error) {
	exist := true
	return exist, nil
}

func (c *checker) createProject() error {
	return nil
}

func (c *checker) canWrite() (bool, error) {
	canWrite := true
	return canWrite, nil
}

type manifestPuller struct {
	*baseHandler
}

func (m *manifestPuller) Enter() (string, error) {
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

	return stateTransferBlob, nil
}

type blobTransfer struct {
	*baseHandler
}

func (b *blobTransfer) Enter() (string, error) {
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

	return statePushManifest, nil
}

type manifestPusher struct {
	*baseHandler
}

func (m *manifestPusher) Enter() (string, error) {
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

	return statePullManifest, nil
}
*/
