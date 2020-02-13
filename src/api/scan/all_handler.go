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
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/scan/all"
	"github.com/pkg/errors"
)

// HandleCheckIn handles the check in data of the scan all job
func HandleCheckIn(checkIn string) {
	if len(checkIn) == 0 {
		// Nothing to handle, directly return
		return
	}

	ck := &all.CheckInData{}
	if err := ck.FromJSON([]byte(checkIn)); err != nil {
		log.Error(errors.Wrap(err, "handle check in"))
	}

	// Start to scan the artifacts
	for _, art := range ck.Artifacts {
		if err := DefaultController.Scan(art, WithRequester(ck.Requester)); err != nil {
			// Just logged
			log.Error(errors.Wrap(err, "handle check in"))
		}
	}
}
