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
	"context"
	"fmt"
	"strconv"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/controller/project"
	ierror "github.com/goharbor/harbor/src/lib/error"
)

const (
	// ProjectReference reference type for project
	ProjectReference = "project"
)

// ReferenceID returns reference id for the interface
func ReferenceID(i interface{}) string {
	switch s := i.(type) {
	case string:
		return s
	case int64:
		return strconv.FormatInt(s, 10)
	case fmt.Stringer:
		return s.String()
	case error:
		return s.Error()
	default:
		return fmt.Sprintf("%v", i)
	}
}

// RefreshForProjects refresh quotas of all projects
func RefreshForProjects(ctx context.Context) error {
	log := log.G(ctx)

	driver, err := Driver(ctx, ProjectReference)
	if err != nil {
		return err
	}

	projects := func(chunkSize int) <-chan *models.Project {
		ch := make(chan *models.Project, chunkSize)

		go func() {
			defer close(ch)

			params := &models.ProjectQueryParam{
				Pagination: &models.Pagination{Page: 1, Size: int64(chunkSize)},
			}

			for {
				results, err := project.Ctl.List(ctx, params, project.Metadata(false))
				if err != nil {
					log.Errorf("list projects failed, error: %v", err)
					return
				}

				for _, p := range results {
					ch <- p
				}

				if len(results) < chunkSize {
					break
				}

				params.Pagination.Page++
			}

		}()

		return ch
	}(50) // default chunk size is 50

	for p := range projects {
		referenceID := ReferenceID(p.ProjectID)

		_, err := Ctl.GetByRef(ctx, ProjectReference, referenceID)
		if ierror.IsNotFoundErr(err) {
			if _, err := Ctl.Create(ctx, ProjectReference, referenceID, driver.HardLimits(ctx)); err != nil {
				log.Warningf("initialize quota for project %s failed, error: %v", p.Name, err)
				continue
			}
		} else if err != nil {
			log.Warningf("get quota of the project %s failed, error: %v", p.Name, err)
			continue
		}

		if err := Ctl.Refresh(ctx, ProjectReference, referenceID, IgnoreLimitation(true)); err != nil {
			log.Warningf("refresh quota usage for project %s failed, error: %v", p.Name, err)
		}
	}

	return nil
}
