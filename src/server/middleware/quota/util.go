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

package quota

import (
	"fmt"
	"net/http"

	"github.com/goharbor/harbor/src/api/quota"
	"github.com/goharbor/harbor/src/server/middleware/util"
)

func projectReferenceObject(r *http.Request) (string, string, error) {
	projectName := util.ParseProjectName(r)

	if projectName == "" {
		return "", "", fmt.Errorf("request %s not match any project", r.URL.Path)
	}

	project, err := projectController.GetByName(r.Context(), projectName)
	if err != nil {
		return "", "", err
	}

	return quota.ProjectReference, quota.ReferenceID(project.ProjectID), nil
}
