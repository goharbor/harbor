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

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	libhttp "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/router"
)

// NewJobStatusHandler handles legacy callbacks using their job IDs.
// Legacy route IDs identify different resources and cannot all be used as task IDs.
func NewJobStatusHandler() http.Handler {
	return &jobStatusHandler{
		handler: task.HkHandler,
	}
}

// NewTaskJobStatusHandler handles callbacks addressed to a task ID.
// The numeric route ID identifies the persisted task, not the jobservice JobID.
func NewTaskJobStatusHandler() http.Handler {
	return &jobStatusHandler{handler: task.HkHandler, bindTaskID: true}
}

type jobStatusHookHandler interface {
	Handle(context.Context, *job.StatusChange) error
	HandleForTask(context.Context, int64, *job.StatusChange) error
}

type jobStatusHandler struct {
	handler    jobStatusHookHandler
	bindTaskID bool
}

func (j *jobStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	// Callbacks update task state and may invoke scheduler processors. Require an
	// internal service identity here even when core is reached without the proxy;
	// an ordinary signed-in user is not a callback producer.
	secCtx, ok := security.FromContext(r.Context())
	if !ok || !secCtx.IsSolutionUser() {
		libhttp.SendError(w, errors.UnauthorizedError(nil))
		return
	}

	sc := &job.StatusChange{}
	if err := json.NewDecoder(r.Body).Decode(sc); err != nil {
		libhttp.SendError(w, errors.BadRequestError(err).WithMessage("invalid callback payload"))
		return
	}
	// The task handler uses metadata for revisions and periodic-job lookup. Reject
	// incomplete payloads before that processing rather than dereferencing nil.
	if sc.JobID == "" || sc.Metadata == nil {
		libhttp.SendError(w, errors.BadRequestError(nil).WithMessage("job ID and metadata are required"))
		return
	}

	var err error
	// Pass both identifiers through: the task handler verifies their association
	// while retaining UpstreamJobID support for periodic callbacks.
	if j.bindTaskID {
		taskID, parseErr := strconv.ParseInt(router.Param(r.Context(), ":id"), 10, 64)
		if parseErr != nil || taskID <= 0 {
			libhttp.SendError(w, errors.BadRequestError(nil).WithMessage("invalid task ID"))
			return
		}
		err = j.handler.HandleForTask(r.Context(), taskID, sc)
	} else {
		err = j.handler.Handle(r.Context(), sc)
	}
	if err != nil {
		// When the status hook comes, the execution/task database record may not insert yet
		// because of that the transaction isn't committed
		// Do not ignore the NotFoundErr here to make jobservice resend the status hook
		// again to avoid the status lost
		/*
			// ignore the not found error to avoid the jobservice re-sending the hook
			if errors.IsNotFoundErr(err) {
				log.Warningf("got not found error: %v, ignore it to avoid subsequent retrying webhooks from jobservice", err)
				return
			}
		*/
		libhttp.SendError(w, err)
		return
	}
}
