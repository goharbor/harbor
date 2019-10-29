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
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/notifier"
	"github.com/goharbor/harbor/src/core/notifier/model"
	"github.com/pkg/errors"
)

// Init the events for scan
func Init() {
	log.Debugf("Subscribe topic %s for cascade deletion of scan reports", model.DeleteImageTopic)

	err := notifier.Subscribe(model.DeleteImageTopic, NewOnDelImageHandler())
	if err != nil {
		log.Error(errors.Wrap(err, "register on delete image handler: init: scan"))
	}
}
