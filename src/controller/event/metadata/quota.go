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
	"time"

	event2 "github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/notifier/event"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
)

// QuotaMetaData defines quota related event data
type QuotaMetaData struct {
	Project  *proModels.Project
	RepoName string
	Tag      string
	Digest   string
	// used to define the event topic
	Level int
	// the msg contains the limitation and current usage of quota
	Msg      string
	OccurAt  time.Time
	Operator string
}

// Resolve quota exceed into common image event
func (q *QuotaMetaData) Resolve(evt *event.Event) error {
	var topic string
	switch q.Level {
	case 1:
		topic = event2.TopicQuotaExceed
	case 2:
		topic = event2.TopicQuotaWarning
	default:
		return errors.New("not supported quota status")
	}

	data := &event2.QuotaEvent{
		EventType: topic,
		Project:   q.Project,
		OccurAt:   q.OccurAt,
		RepoName:  q.RepoName,
		Msg:       q.Msg,
		Operator:  q.Operator,
	}
	if q.Tag != "" || q.Digest != "" {
		data.Resource = &event2.ImgResource{
			Tag:    q.Tag,
			Digest: q.Digest,
		}
	}

	evt.Topic = topic
	evt.Data = data
	return nil
}
