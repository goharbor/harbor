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

package opm

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/utils"
)

const (
	clientTimeout         = 10 * time.Second
	maxIdleConnections    = 20
	idleConnectionTimeout = 30 * time.Second
)

// DefaultHookClient is for default use.
var DefaultHookClient = NewHookClient()

// HookClient is used to post the related data to the interested parties.
type HookClient struct {
	client *http.Client
}

// NewHookClient return the ptr of the new HookClient
func NewHookClient() *HookClient {
	client := &http.Client{
		Timeout: clientTimeout,
		Transport: &http.Transport{
			MaxIdleConns:    maxIdleConnections,
			IdleConnTimeout: idleConnectionTimeout,
		},
	}

	return &HookClient{
		client: client,
	}
}

// ReportStatus reports the status change info to the subscribed party.
// The status includes 'checkin' info with format 'check_in:<message>'
func (hc *HookClient) ReportStatus(hookURL string, status models.JobStatusChange) error {
	if utils.IsEmptyStr(hookURL) {
		return errors.New("empty hook url") // do nothing
	}

	// Parse and validate URL
	url, err := url.Parse(hookURL)
	if err != nil {
		return err
	}

	// Marshal data
	data, err := json.Marshal(&status)
	if err != nil {
		return err
	}

	// New post request
	req, err := http.NewRequest(http.MethodPost, url.String(), strings.NewReader(string(data)))
	if err != nil {
		return err
	}

	res, err := hc.client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close() // close connection for reuse

	// Should be 200
	if res.StatusCode != http.StatusOK {
		if res.ContentLength > 0 {
			// read error content and return
			dt, err := ioutil.ReadAll(res.Body)
			if err != nil {
				return err
			}
			return errors.New(string(dt))
		}

		return fmt.Errorf("failed to report status change via hook, expect '200' but got '%d'", res.StatusCode)
	}

	return nil
}
