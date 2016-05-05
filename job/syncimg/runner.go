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

import (
	"strings"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/schema1"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/vmware/harbor/job"
	"github.com/vmware/harbor/models"
	registry_util "github.com/vmware/harbor/utils/registry"
)

const (
	jobType           = "sync_img"
	stateCheck        = "check"
	statePullManifest = "pull_manifest"
	stateTransferBlob = "transfer_blob"
	statePushManifest = "push_manifest"
)

// Runner ...
type Runner struct {
	job.JobSM
	Logger job.Logger
}

// Run ...
func (r *Runner) Run(je models.JobEntry) error {
	r.init(je)
	r.Start(job.JobRunning)
	return nil
}

func (r *Runner) init(je models.JobEntry) {
	r.JobID = je.ID
	r.Logger = job.Logger{je.ID}

	r.InitJobSM()

	base := initBase(je)

	r.AddTransition(job.JobRunning, stateCheck, &checker{base: base})
	r.AddTransition(stateCheck, statePullManifest, &manifestPuller{base: base})
	r.AddTransition(statePullManifest, stateTransferBlob, &blobTransfer{base: base})
	r.AddTransition(statePullManifest, job.JobFinished, &job.StatusUpdater{job.DummyHandler{JobID: r.JobID}, job.JobFinished})
	r.AddTransition(stateTransferBlob, statePushManifest, &manifestPusher{base: base})
	r.AddTransition(statePushManifest, statePullManifest, &manifestPuller{base: base})
}

type base struct {
	project string // project_name
	name    string // project_name/repo_name
	tags    []string

	url       string // url of target registry
	username  string
	password  string
	srcClient *registry_util.Repository
	dstClient *registry_util.Repository

	manifest distribution.Manifest
	blobs    []string // blobs need to be transferred for tags[0]

	logger *job.Logger
}

func initBase(je models.JobEntry) *base {
	base := &base{}
	return base
}

func (b *base) Exit() error {
	return nil
}

type checker struct {
	*base
}

// check existence of project, if it does not exist, create it,
// if it exists, check whether the user has privileges to it.
func (c *checker) Enter() (string, error) {
	exist, err := c.projectExist()
	if err != nil {
		c.logger.Errorf("an error occurred while checking existence of project %s on %s with user %s : %v", c.project, c.url, c.username, err)
		return "", err
	}
	if !exist {
		if err := c.createProject(); err != nil {
			c.logger.Errorf("an error occurred while creating project %s on %s with user %s : %v", c.project, c.url, c.username, err)
			return "", err
		}
		c.logger.Infof("project %s is created on %s with user %s", c.project, c.url, c.username)
		return "", nil
	}

	c.logger.Infof("project %s already exists on %s", c.project, c.url)
	canWrite, err := c.canWrite()
	if err != nil {
		c.logger.Errorf("an error occurred while checking %s 's privileges to project %s on %s : %v", c.username, c.project, c.url, err)
		return "", err
	}

	if !canWrite {
		c.logger.Errorf("the user %s has no write privilege to project %s on %s", c.username, c.project, c.url)
		return "", err
	}
	c.logger.Infof("check completed for project %s on %s with user %s ", c.project, c.url, c.username)

	return statePullManifest, nil
}

func (c *checker) projectExist() (bool, error) {
	exist := false
	return exist, nil
}

func (c *checker) createProject() error {
	return nil
}

func (c *checker) canWrite() (bool, error) {
	canWrite := false
	return canWrite, nil
}

type manifestPuller struct {
	*base
}

func (m *manifestPuller) Enter() (string, error) {
	if m.tags == nil || len(m.tags) == 0 {
		m.logger.Infof("no tag needs to be replicated, trying to stop")
		return job.JobFinished, nil
	}

	name := m.name
	tag := m.tags[0]

	acceptMediaTypes := []string{schema1.MediaTypeManifest, schema2.MediaTypeManifest}
	digest, mediaType, payload, err := m.srcClient.PullManifest(tag, acceptMediaTypes)
	if err != nil {
		m.logger.Errorf("an error occurred while pulling manifest of %s:%s: %v", name, tag, err)
		return "", err
	}
	m.logger.Infof("manifest pulled successfully: %s:%s %s", m.name, tag, digest)

	if strings.Contains(mediaType, "application/json") {
		mediaType = schema1.MediaTypeManifest
	}

	manifest, _, err := registry_util.UnMarshal(mediaType, payload)
	if err != nil {
		m.logger.Errorf("an error occurred while parsing manifest of %s:%s: %v", name, tag, err)
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

	m.logger.Infof("all blobs of %s:%s : %v", name, tag, blobs)

	for _, blob := range blobs {
		exist, err := m.dstClient.BlobExist(blob)
		if err != nil {
			m.logger.Errorf("an error occurred while checking existence of blob %s of %s:%s %v", blob, name, tag, err)
			return "", err
		}
		if !exist {
			m.blobs = append(m.blobs, blob)
		}
	}
	m.logger.Infof("blobs need to be transferred of %s:%s : %v", name, tag, m.blobs)

	m.blobs = blobs
	m.tags = m.tags[1:]

	return stateTransferBlob, nil
}

type blobTransfer struct {
	*base
}

func (b *blobTransfer) Enter() (string, error) {
	name := b.name
	tag := b.tags[0]
	for _, blob := range b.blobs {
		size, data, err := b.srcClient.PullBlob(blob)
		if err != nil {
			b.logger.Errorf("an error occurred while pulling blob %s of %s:%s : %v", blob, name, tag, err)
			return "", err
		}
		if err = b.dstClient.PushBlob(blob, size, data); err != nil {
			b.logger.Errorf("an error occurred while pushing blob %s of %s:%s to %s : %v", blob, name, tag, b.url, err)
			return "", err
		}
		b.logger.Infof("blob %s of %s:%s tranferred completed", blob, name, tag)
	}

	return statePushManifest, nil
}

type manifestPusher struct {
	*base
}

func (m *manifestPusher) Enter() (string, error) {
	name := m.name
	tag := m.tags[0]
	_, exist, err := m.srcClient.ManifestExist(tag)
	if err != nil {
		m.logger.Infof("an error occurred while checking the existence of manifest of %s:%s : %v", name, tag, err)
		return "", err
	}
	if !exist {
		m.logger.Infof("manifest of %s:%s does not exist on source registry, cancel manifest pushing", name, tag)
	} else {
		mediaType, data, err := m.manifest.Payload()
		if err != nil {
			m.logger.Errorf("an error occurred while getting payload of manifest for %s:%s : %v", name, tag, err)
			return "", err
		}
		if _, err = m.dstClient.PushManifest(tag, mediaType, data); err != nil {
			m.logger.Errorf("an error occurred while pushing manifest of %s:%s to %s : %v", name, tag, m.url, err)
			return "", err
		}
		m.logger.Infof("manifest of %s:%s has been pushed to %s", name, tag, m.url)
	}

	m.tags = m.tags[1:]

	return statePullManifest, nil
}
