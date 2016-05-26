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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
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

	logger *log.Logger
}

// NewDeleter returns a Deleter
func NewDeleter(repository string, tags []string, dstURL, dstUsr, dstPwd string, logger *log.Logger) *Deleter {
	return &Deleter{
		repository: repository,
		tags:       tags,
		dstURL:     dstURL,
		dstUsr:     dstUsr,
		dstPwd:     dstPwd,
		logger:     logger,
	}
}

// Exit ...
func (d *Deleter) Exit() error {
	return nil
}

// Enter deletes repository or tags
func (d *Deleter) Enter() (string, error) {
	url := strings.TrimRight(d.dstURL, "/") + "/api/repositories/"

	// delete repository
	if len(d.tags) == 0 {
		u := url + "?repo_name=" + d.repository
		if err := del(u, d.dstUsr, d.dstPwd); err != nil {
			d.logger.Errorf("an error occurred while deleting repository %s on %s with user %s: %v", d.repository, d.dstURL, d.dstUsr, err)
			return "", err
		}

		d.logger.Infof("repository %s on %s has been deleted", d.repository, d.dstURL)

		return models.JobFinished, nil
	}

	// delele tags
	for _, tag := range d.tags {
		u := url + "?repo_name=" + d.repository + "&tag=" + tag
		if err := del(u, d.dstUsr, d.dstPwd); err != nil {
			d.logger.Errorf("an error occurred while deleting repository %s:%s on %s with user %s: %v", d.repository, tag, d.dstURL, d.dstUsr, err)
			return "", err
		}

		d.logger.Infof("repository %s:%s on %s has been deleted", d.repository, tag, d.dstURL)
	}

	return models.JobFinished, nil
}

func del(url, username, password string) error {
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.SetBasicAuth(username, password)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		return nil
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return fmt.Errorf("%d %s", resp.StatusCode, string(b))
}
