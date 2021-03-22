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

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
)

// DelArtHandler is a handler to listen to the internal delete image event.
type DelArtHandler struct {
}

// Name ...
func (o *DelArtHandler) Name() string {
	return "DeleteArtifactWebhook"
}

// Handle ...
func (o *DelArtHandler) Handle(ctx context.Context, value interface{}) error {
	if value == nil {
		return errors.New("delete image event handler: nil value ")
	}

	evt, ok := value.(*event.DeleteArtifactEvent)
	if !ok {
		return errors.New("delete image event handler: malformed image event model")
	}

	log.Debugf("clear the scan reports as receiving event %s", evt.EventType)

	// Check if it is safe to delete the reports.
	count, err := artifact.Ctl.Count(ctx, q.New(q.KeyWords{"digest": evt.Artifact.Digest}))
	if err != nil {
		// Just logged
		log.Error(errors.Wrap(err, "delete image event handler"))
	} else if count == 0 {
		log.Debugf("prepare to remove the scan report linked with artifact: %s", evt.Artifact.Digest)

		if err := scan.DefaultController.DeleteReports(ctx, evt.Artifact.Digest); err != nil {
			return errors.Wrap(err, "delete image event handler")
		}
	}

	return nil
}

// IsStateful ...
func (o *DelArtHandler) IsStateful() bool {
	return false
}
