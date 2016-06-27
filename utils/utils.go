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

	"github.com/vmware/harbor/models"
)

// Repository holds information about repository
type Repository struct {
	Name string
}

// GetProject parses the repository and return the name of project.
func (r *Repository) GetProject() string {
	if !strings.ContainsRune(r.Name, '/') {
		return ""
	}
	return r.Name[0:strings.LastIndex(r.Name, "/")]
}

// ProjectSorter holds an array of projects
type ProjectSorter struct {
	Projects []models.Project
}

// Len returns the length of array in ProjectSorter
func (ps *ProjectSorter) Len() int {
	return len(ps.Projects)
}

// Less defines the comparison rules of project
func (ps *ProjectSorter) Less(i, j int) bool {
	return ps.Projects[i].Name < ps.Projects[j].Name
}

// Swap swaps the position of i and j
func (ps *ProjectSorter) Swap(i, j int) {
	ps.Projects[i], ps.Projects[j] = ps.Projects[j], ps.Projects[i]
}

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
