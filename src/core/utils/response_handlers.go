// Copyright 2018 Project Harbor Authors
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

package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/goharbor/harbor/src/common/api"
	"github.com/goharbor/harbor/src/lib/log"
)

// ResponseHandler provides utility to handle http response.
type ResponseHandler interface {
	Handle(*http.Response) error
}

// StatusRespHandler handles the response to check if the status is expected, if not returns an error.
type StatusRespHandler struct {
	status int
}

// Handle ...
func (s StatusRespHandler) Handle(resp *http.Response) error {
	defer resp.Body.Close()
	if resp.StatusCode != s.status {
		b, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("unexpected status code: %d, text: %s", resp.StatusCode, string(b))
	}
	return nil
}

// NewStatusRespHandler ...
func NewStatusRespHandler(sc int) ResponseHandler {
	return StatusRespHandler{
		status: sc,
	}
}

// JobLogRespHandler handles the response from jobservice to show the log of a job
type JobLogRespHandler struct {
	theAPI *api.BaseAPI
}

// Handle will consume the response of job service and put the content of the job log in the response of the API.
func (h JobLogRespHandler) Handle(resp *http.Response) error {
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		h.theAPI.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), resp.Header.Get(http.CanonicalHeaderKey("Content-Length")))
		h.theAPI.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")

		if _, err := io.Copy(h.theAPI.Ctx.ResponseWriter, resp.Body); err != nil {
			log.Errorf("failed to write log to response; %v", err)
			return err
		}
		return nil
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read response body: %v", err)
		return err
	}
	h.theAPI.RenderError(resp.StatusCode, fmt.Sprintf("message from jobservice: %s", string(b)))
	return nil
}

// NewJobLogRespHandler ...
func NewJobLogRespHandler(apiHandler *api.BaseAPI) ResponseHandler {
	return &JobLogRespHandler{
		theAPI: apiHandler,
	}
}
