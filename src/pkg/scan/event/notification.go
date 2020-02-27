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

package event

import (
	"context"
	bo "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/scan"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/notifier"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/pkg/errors"
)

// scanCtlGetter for getting a scan controller reference to avoid package importing order issue.
type scanCtlGetter func() scan.Controller

// artCtlGetter for getting a artifact controller reference to avoid package importing order issue.
type artCtlGetter func() artifact.Controller

// onDelImageHandler is a handler to listen to the internal delete image event.
type onDelImageHandler struct {
	// scan controller
	scanCtl scanCtlGetter
	// artifact controller
	artCtl artCtlGetter
}

// NewOnDelImageHandler creates a new handler to handle on del event.
func NewOnDelImageHandler() notifier.NotificationHandler {
	return &onDelImageHandler{
		scanCtl: func() scan.Controller {
			return scan.DefaultController
		},
		artCtl: func() artifact.Controller {
			return artifact.Ctl
		},
	}
}

func (o *onDelImageHandler) Handle(value interface{}) error {
	if value == nil {
		return errors.New("delete image event handler: nil value ")
	}

	evt, ok := value.(*model.ImageEvent)
	if !ok {
		return errors.New("delete image event handler: malformed image event model")
	}

	log.Debugf("clear the scan reports as receiving event %s", evt.EventType)

	digests := make([]string, 0)
	query := &q.Query{
		Keywords: make(map[string]interface{}),
	}

	ctx := orm.NewContext(context.TODO(), bo.NewOrm())
	for _, res := range evt.Resource {
		// Check if it is safe to delete the reports.
		query.Keywords["digest"] = res.Digest
		l, err := o.artCtl().List(ctx, query, nil)

		if err != nil {
			// Just logged
			log.Error(errors.Wrap(err, "delete image event handler"))
			// Passed for safe consideration
			continue
		}

		if len(l) == 0 {
			digests = append(digests, res.Digest)
			log.Debugf("prepare to remove the scan report linked with artifact: %s", res.Digest)
		}
	}

	if err := o.scanCtl().DeleteReports(digests...); err != nil {
		return errors.Wrap(err, "delete image event handler")
	}

	return nil
}

func (o *onDelImageHandler) IsStateful() bool {
	return false
}
