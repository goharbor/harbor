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

	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
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

	chunkSize := 50 // default chunk size is 50
	for result := range project.ListAll(ctx, chunkSize, nil, project.Metadata(false)) {
		if result.Error != nil {
			log.Errorf("refresh quota for all projects got error: %v", result.Error)
			continue
		}

		p := result.Data
		referenceID := ReferenceID(p.ProjectID)

		_, err := Ctl.GetByRef(ctx, ProjectReference, referenceID)
		if errors.IsNotFoundErr(err) {
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
