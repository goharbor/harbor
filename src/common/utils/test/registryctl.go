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

package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"
)

// GCResult ...
type GCResult struct {
	Status    bool      `json:"status"`
	Msg       string    `json:"msg"`
	StartTime time.Time `json:"starttime"`
	EndTime   time.Time `json:"endtime"`
}

// NewRegistryCtl returns a mock registry server
func NewRegistryCtl(config map[string]interface{}) (*httptest.Server, error) {
	m := []*RequestHandlerMapping{}

	gcr := GCResult{true, "hello-world", time.Now(), time.Now()}
	b, err := json.Marshal(gcr)
	if err != nil {
		return nil, err
	}

	resp := &Response{
		StatusCode: http.StatusOK,
		Body:       b,
	}

	m = append(m, &RequestHandlerMapping{
		Method:  "GET",
		Pattern: "/api/health",
		Handler: Handler(&Response{
			StatusCode: http.StatusOK,
		}),
	})

	m = append(m, &RequestHandlerMapping{
		Method:  "POST",
		Pattern: "/api/registry/gc",
		Handler: Handler(resp),
	})

	return NewServer(m...), nil
}
