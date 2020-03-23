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

package scan

import (
	"context"
	bo "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/pkg/errors"
)

// DelArtHandler is a handler to listen to the internal delete image event.
type DelArtHandler struct {
}

// Handle ...
func (o *DelArtHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("delete image event handler: nil value ")
	}

	evt, ok := value.(*event.DeleteArtifactEvent)
	if !ok {
		return errors.New("delete image event handler: malformed image event model")
	}

	log.Debugf("clear the scan reports as receiving event %s", evt.EventType)

	digests := make([]string, 0)
	query := &q.Query{
		Keywords: make(map[string]interface{}),
	}

	ctx := orm.NewContext(context.TODO(), bo.NewOrm())
	// Check if it is safe to delete the reports.
	query.Keywords["digest"] = evt.Artifact.Digest
	l, err := artifact.Ctl.List(ctx, query, nil)

	if err != nil && len(l) != 0 {
		// Just logged
		log.Error(errors.Wrap(err, "delete image event handler"))
		// Passed for safe consideration
	} else {
		if len(l) == 0 {
			digests = append(digests, evt.Artifact.Digest)
			log.Debugf("prepare to remove the scan report linked with artifact: %s", evt.Artifact.Digest)

		}
	}

	if err := scan.DefaultController.DeleteReports(digests...); err != nil {
		return errors.Wrap(err, "delete image event handler")
	}

	return nil
}

// IsStateful ...
func (o *DelArtHandler) IsStateful() bool {
	return false
}
