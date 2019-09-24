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

package hook

import (
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/manager"
)

// Global status hook handler
var (
	Hdlr = NewHandler()
)

// Handler defines the method that a status hook handler should implement
type Handler interface {
	Handle(id int64, sc *job.StatusChange) error
}

// NewHandler returns the default implementation of Handler
func NewHandler() Handler {
	return &handler{
		mgr: manager.New(),
	}
}

type handler struct {
	mgr manager.Manager
}

func (h *handler) Handle(id int64, sc *job.StatusChange) error {
	if sc == nil {
		return nil
	}
	// handle check in data
	if len(sc.CheckIn) > 0 {
		return h.mgr.AppendCheckInData(id, sc.CheckIn)
	}
	// handle status update
	statusCode := job.Status(sc.Status).Code()
	var statusRevision int64
	if sc.Metadata != nil {
		statusRevision = sc.Metadata.Revision
	}
	return h.mgr.UpdateStatus(id, sc.Status, statusCode, statusRevision)
}
