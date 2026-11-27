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

package githubcr

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/pkg/reg/util"
)

// githubAPIURL is the GitHub REST API. The ghcr.io catalog API is global: it
// shows all public repositories on GitHub. Only the GitHub Packages API can
// list the packages of one owner.
const githubAPIURL = "https://api.github.com"

// githubPageSize is the number of packages on each page. 100 is the GitHub
// maximum.
const githubPageSize = 100

type githubPackage struct {
	Name string `json:"name"`
}

// listRepositories returns the repositories that match the repository name
// filter pattern. It lists the packages of each owner in pattern with the
// GitHub Packages API. Then it applies the filter.
func (a *adapter) listRepositories(pattern string) ([]string, error) {
	owners, err := ownersOf(pattern)
	if err != nil {
		return nil, err
	}
	var repositories []string
	for _, owner := range owners {
		packages, err := a.listPackages(owner)
		if err != nil {
			return nil, err
		}
		for _, pkg := range packages {
			repository := owner + "/" + pkg.Name
			ok, err := util.Match(pattern, repository)
			if err != nil {
				return nil, err
			}
			if ok {
				repositories = append(repositories, repository)
			}
		}
	}
	return repositories, nil
}

// ownersOf returns the GitHub owners (orgs or users) that pattern can match.
// The first path component of pattern is the owner. It must be a fixed name
// or one {a,b} group. A wildcard owner is not possible, because the adapter
// must not list all repositories on ghcr.io.
func ownersOf(pattern string) ([]string, error) {
	owner, _, _ := strings.Cut(pattern, "/")
	owners, ok := util.IsSpecificPathComponent(owner)
	if !ok {
		return nil, fmt.Errorf("the owner %q in the repository name filter %q is not valid", owner, pattern)
	}
	return owners, nil
}

// listPackages returns the container packages of owner. It tries the org
// endpoint first. If the org does not exist, it tries the user endpoint.
func (a *adapter) listPackages(owner string) ([]*githubPackage, error) {
	var packages []*githubPackage
	err := a.client.GetAndIteratePagination(buildPackagesURL(a.apiURL, "orgs", owner), &packages)
	var httpErr *common_http.Error
	if errors.As(err, &httpErr) && httpErr.Code == http.StatusNotFound {
		err = a.client.GetAndIteratePagination(buildPackagesURL(a.apiURL, "users", owner), &packages)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to list the container packages of %s: %w", owner, err)
	}
	return packages, nil
}

func buildPackagesURL(apiURL, ownerType, owner string) string {
	return fmt.Sprintf("%s/%s/%s/packages?package_type=container&per_page=%d",
		apiURL, ownerType, url.PathEscape(owner), githubPageSize)
}
