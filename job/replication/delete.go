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
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
	"github.com/vmware/harbor/utils/registry"
	"github.com/vmware/harbor/utils/registry/auth"
)

const (
	// StateDelete ...
	StateDelete = "delete"
)

// Deleter deletes repository or tags
type Deleter struct {
	repository string // prject_name/repo_name
	tags       []string

	dstURL string // url of target registry
	dstUsr string // username ...
	dstPwd string // username ...

	insecure bool

	dstClient *registry.Repository

	logger *log.Logger
}

// NewDeleter returns a Deleter
func NewDeleter(repository string, tags []string, dstURL, dstUsr, dstPwd string, insecure bool, logger *log.Logger) (*Deleter, error) {
	dstCred := auth.NewBasicAuthCredential(dstUsr, dstPwd)
	dstClient, err := newRepositoryClient(dstURL, insecure, dstCred,
		repository, "repository", repository, "pull", "push", "*")
	if err != nil {
		return nil, err
	}

	deleter := &Deleter{
		repository: repository,
		tags:       tags,
		dstURL:     dstURL,
		dstUsr:     dstUsr,
		dstPwd:     dstPwd,
		insecure:   insecure,
		dstClient:  dstClient,
		logger:     logger,
	}
	deleter.logger.Infof("initialization completed: repository: %s, tags: %v, destination URL: %s, destination user: %s",
		deleter.repository, deleter.tags, deleter.dstURL, deleter.dstUsr)
	return deleter, nil
}

// Exit ...
func (d *Deleter) Exit() error {
	return nil
}

// Enter deletes repository or tags
func (d *Deleter) Enter() (string, error) {

	if len(d.tags) == 0 {
		tags, err := d.dstClient.ListTag()
		if err != nil {
			d.logger.Errorf("an error occurred while listing tags of repository %s on %s with user %s: %v", d.repository, d.dstURL, d.dstUsr, err)
			return "", err
		}

		d.tags = append(d.tags, tags...)
	}

	d.logger.Infof("tags %v will be deleted", d.tags)

	for _, tag := range d.tags {

		if err := d.dstClient.DeleteTag(tag); err != nil {
			d.logger.Errorf("an error occurred while deleting repository %s:%s on %s with user %s: %v", d.repository, tag, d.dstURL, d.dstUsr, err)
			return "", err
		}

		d.logger.Infof("repository %s:%s on %s has been deleted", d.repository, tag, d.dstURL)
	}

	return models.JobFinished, nil
}
