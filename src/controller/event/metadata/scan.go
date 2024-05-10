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

package metadata

import (
	"fmt"
	"time"

	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"
)

// ScanImageMetaData defines meta data of image scanning event
type ScanImageMetaData struct {
	Artifact *v1.Artifact
	ScanType string
	Status   string
	Operator string
}

// Resolve image scanning metadata into common chart event
func (si *ScanImageMetaData) Resolve(evt *event.Event) error {
	var eventType string
	var topic string

	switch job.Status(si.Status) {
	case job.SuccessStatus:
		eventType = event2.TopicScanningCompleted
		topic = event2.TopicScanningCompleted
	case job.StoppedStatus:
		eventType = event2.TopicScanningStopped
		topic = event2.TopicScanningStopped
	case job.ErrorStatus:
		eventType = event2.TopicScanningFailed
		topic = event2.TopicScanningFailed
	default:
		return fmt.Errorf("not supported scan hook status %s", si.Status)
	}

	data := &event2.ScanImageEvent{
		EventType: eventType,
		Artifact:  si.Artifact,
		OccurAt:   time.Now(),
		Operator:  si.Operator,
		ScanType:  si.ScanType,
	}

	evt.Topic = topic
	evt.Data = data
	return nil
}
