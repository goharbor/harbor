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

package utils

import (
	"net/url"
	"strings"
)

// FormatEndpoint formats endpoint
func FormatEndpoint(endpoint string) string {
	endpoint = strings.TrimSpace(endpoint)
	endpoint = strings.TrimRight(endpoint, "/")
	if !strings.HasPrefix(endpoint, "http://") &&
		!strings.HasPrefix(endpoint, "https://") {
		endpoint = "http://" + endpoint
	}

	return endpoint
}

// ParseEndpoint parses endpoint to a URL
func ParseEndpoint(endpoint string) (*url.URL, error) {
	endpoint = FormatEndpoint(endpoint)

	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// ParseRepository splits a repository into two parts: project and rest
func ParseRepository(repository string) (project, rest string) {
	repository = strings.TrimLeft(repository, "/")
	repository = strings.TrimRight(repository, "/")
	if !strings.ContainsRune(repository, '/') {
		rest = repository
		return
	}
	index := strings.LastIndex(repository, "/")
	project = repository[0:index]
	rest = repository[index+1:]
	return
}
